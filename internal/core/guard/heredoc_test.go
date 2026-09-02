package guard

import (
	"errors"
	"testing"
)

// TestSubshellHeredocIsNotArithmetic pins the line between the two shapes a `<<`
// followed by a paren can be. `( cmd <<EOF )` is a subshell opening a genuine
// here-document, which every shell here runs with the following lines as its
// document (probed on bash 3.2: the subshell runs, the body is data, and the
// line after the terminator runs). `$(( x << y ))` is an arithmetic shift whose
// right-hand identifier readHeredocDelim reads as a delimiter word, stopping at
// the expression's own closing paren.
//
// Testing only for "a paren after the delimiter" read the subshell as
// arithmetic, so the document was tokenized as commands: one apostrophe in it
// became ErrUnparsableCommand, which the pre-tool-use hook maps to fail-OPEN, so
// a blocked command ran unguarded; and a document merely NAMING a hazard warned
// as if it ran one. The discriminator is WHICH paren: arithmetic closes with
// `))` on the same line and has no body.
func TestSubshellHeredocIsNotArithmetic(t *testing.T) {
	cases := []struct {
		name    string
		line    string
		verdict Verdict
		entry   string
	}{
		{"apostrophe body in a subshell heredoc", "( gh api -X DELETE repos/owner/repo <<EOF )\nit's fine\nEOF", VerdictBlock, "gh-api-repo-delete"},
		{"body of a subshell heredoc is data", "( cat <<EOF )\ngit clean -fd\nEOF", VerdictAllow, ""},
		{"spaced arithmetic still parses", "echo $(( x << y ))\ngit push --force origin main", VerdictBlock, "git-push-force"},
		{"tight arithmetic still parses", "echo $((x<<y))\ngit push --force origin main", VerdictBlock, "git-push-force"},
		{"spaced arithmetic alone", "echo $(( x << y ))", VerdictAllow, ""},
		{"arithmetic command group", "(( x << y ))\ngit push --force origin main", VerdictBlock, "git-push-force"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, err := Defaults().Check(tc.line)
			if err != nil {
				t.Fatalf("Check(%q): %v — bash runs this line, so the guard must answer, not error", tc.line, err)
			}
			if d.Verdict != tc.verdict {
				t.Fatalf("Check(%q).Verdict = %q via %q, want %q", tc.line, d.Verdict, d.EntryID, tc.verdict)
			}
			if tc.entry != "" && d.EntryID != tc.entry {
				t.Fatalf("Check(%q).EntryID = %q, want %q", tc.line, d.EntryID, tc.entry)
			}
		})
	}
	// Where the shape stays ambiguous the answer is the fail-closed
	// heredoc-unterminated verdict, never an error the hook fails open on.
	for _, line := range []string{
		"( cat <<EOF ) | tee out",
		"( cat <<EOF\t)",
		"x=$(cat <<EOF )\nit's",
		"x=$(cat <<EOF )",
	} {
		if _, err := Defaults().Check(line); err != nil {
			t.Errorf("Check(%q): %v — an ambiguous `<<` must take a verdict, never an error", line, err)
		}
	}
}

// TestHeredocBodyStartsDespiteTrailingListOperator pins where a here-document
// body begins: on the line after the redirection, always. bash collects the
// bodies at the end of the PHYSICAL line, so a trailing `&&`, `||` or `|` does
// not defer them — probed on bash 3.2, `cat <<EOF &&` / body / `EOF` / `echo ok`
// prints ok and never runs the body.
//
// Deferring the body until the command list completed put document text in
// command position twice over: an apostrophe in a document became
// ErrUnparsableCommand and the hook failed OPEN, and the scan for the terminator
// started past an early delimiter line, so a stray delimiter at the end
// swallowed the real commands between them as body — a SILENT allow.
func TestHeredocBodyStartsDespiteTrailingListOperator(t *testing.T) {
	const line = "cat <<EOF &&\ndon't\nEOF\necho ok"
	if _, err := Defaults().Check(line); err != nil {
		t.Fatalf("Check(%q): %v — bash runs this; the body starts on the next line regardless of the `&&`", line, err)
	}
	const hazard = "cat <<EOF &&\ndon't\nEOF\ngit push --force origin main"
	d, err := Defaults().Check(hazard)
	if err != nil {
		t.Fatalf("Check(%q): %v", hazard, err)
	}
	if d.Verdict != VerdictBlock || d.EntryID != "git-push-force" {
		t.Fatalf("Check(%q) = %q via %q, want block via git-push-force", hazard, d.Verdict, d.EntryID)
	}
	const swallowed = "cat <<EOF &&\nEOF\ngit push --force origin main\nEOF"
	d, err = Defaults().Check(swallowed)
	if err != nil {
		t.Fatalf("Check(%q): %v", swallowed, err)
	}
	if d.Verdict != VerdictBlock || d.EntryID != "git-push-force" {
		t.Fatalf("Check(%q) = %q via %q, want block via git-push-force", swallowed, d.Verdict, d.EntryID)
	}
	// What stays an error is what no shell runs: an unterminated quote in
	// COMMAND text. A quote inside a body is document text and is not one.
	if _, err := tokenize(`echo 'unterminated`); !errors.Is(err, ErrUnparsableCommand) {
		t.Fatalf("an unterminated quote in command text must stay ErrUnparsableCommand")
	}
}
