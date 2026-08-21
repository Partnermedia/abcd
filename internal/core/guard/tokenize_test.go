package guard

import (
	"errors"
	"strings"
	"testing"
)

// render flattens segments to a comparable form: chain index, then the tokens
// joined with "|" so an empty or embedded-space token is still visible.
func render(segs []segment) []string {
	out := make([]string, len(segs))
	for i, s := range segs {
		out[i] = string(rune('0'+s.chain)) + ":" + strings.Join(s.tokens, "|")
	}
	return out
}

func TestTokenizeSegments(t *testing.T) {
	tests := []struct {
		name string
		line string
		want []string
	}{
		{
			name: "single command",
			line: "rm -rf *",
			want: []string{"0:rm|-rf|*"},
		},
		{
			name: "cd chain across &&",
			line: "cd scratch && rm -rf *",
			want: []string{"0:cd|scratch", "0:rm|-rf|*"},
		},
		{
			name: "quoted hazard is one argument token",
			line: `abcd capture "agent ran cd scratch && rm -rf * — one failed cd from disaster"`,
			want: []string{"0:abcd|capture|agent ran cd scratch && rm -rf * — one failed cd from disaster"},
		},
		{
			name: "single quotes suppress operators",
			line: `echo 'a && b'`,
			want: []string{"0:echo|a && b"},
		},
		{
			name: "backslash escapes a space",
			line: `echo a\ b`,
			want: []string{`0:echo|a b`},
		},
		{
			name: "backslash escapes a quote inside double quotes",
			line: `echo "a\"b"`,
			want: []string{`0:echo|a"b`},
		},
		{
			name: "backslash escapes a separator",
			line: `echo a\&\&b`,
			want: []string{`0:echo|a&&b`},
		},
		{
			name: "all compound separators split command position",
			line: "a; b | c || d && e",
			want: []string{"0:a", "0:b", "0:c", "0:d", "0:e"},
		},
		{
			name: "background operator splits",
			line: "sleep 1 & rm -rf *",
			want: []string{"0:sleep|1", "0:rm|-rf|*"},
		},
		{
			name: "subshell parentheses split",
			line: "(cd x && rm -rf *)",
			want: []string{"0:cd|x", "0:rm|-rf|*"},
		},
		{
			name: "newline starts a new chain",
			line: "cd x\nrm -rf *",
			want: []string{"0:cd|x", "1:rm|-rf|*"},
		},
		{
			name: "comment at word start is stripped",
			line: "echo hi # rm -rf *",
			want: []string{"0:echo|hi"},
		},
		{
			name: "hash inside a word is not a comment",
			line: "curl https://example.test/#frag",
			want: []string{"0:curl|https://example.test/#frag"},
		},
		{
			name: "quoted hash is not a comment",
			line: `git commit -m "fix #123"`,
			want: []string{"0:git|commit|-m|fix #123"},
		},
		{
			name: "a newline after a list operator continues the same chain",
			line: "cd scratch &&\nrm -rf *",
			want: []string{"0:cd|scratch", "0:rm|-rf|*"},
		},
		{
			name: "a newline after a pipe continues the same chain",
			line: "cd scratch |\nrm -rf *",
			want: []string{"0:cd|scratch", "0:rm|-rf|*"},
		},
		{
			// The `> doc.md` redirection is dropped (a target is never in
			// command position), and the heredoc body stays data — the
			// `git push --force` inside it must never reach command position.
			name: "heredoc body is data, not commands",
			line: "cat > doc.md <<'EOF'\ngit push --force\nEOF",
			want: []string{"0:cat"},
		},
		{
			name: "heredoc body ends at its delimiter line",
			line: "cat <<EOF\nrm -rf *\nEOF\nls -la",
			want: []string{"0:cat", "1:ls|-la"},
		},
		{
			name: "dash heredoc delimiter may be indented",
			line: "cat <<-EOF\n\trm -rf *\n\tEOF\nls",
			want: []string{"0:cat", "1:ls"},
		},
		{
			name: "the rest of the heredoc's own line is still command text",
			line: "cat <<EOF && ls\nrm -rf *\nEOF",
			want: []string{"0:cat", "0:ls"},
		},
		{
			// The body starts only once the command line is complete, so the
			// continuation of a pipeline is still command text, not body text.
			name: "a heredoc queued on a continued line waits for the command to finish",
			line: "cat <<EOF |\ngrep x\nrm -rf *\nEOF",
			want: []string{"0:cat", "0:grep|x"},
		},
		{
			name: "a blank line does not break a list continuation",
			line: "cd scratch &&\n\nrm -rf *",
			want: []string{"0:cd|scratch", "0:rm|-rf|*"},
		},
		{
			name: "a comment line does not break a list continuation",
			line: "cd scratch &&\n# note\nrm -rf *",
			want: []string{"0:cd|scratch", "0:rm|-rf|*"},
		},
		{
			name: "a quoted heredoc delimiter may be exotic",
			line: "cat <<'---'\nrm -rf *\n---\nls",
			want: []string{"0:cat", "1:ls"},
		},
		{
			name: "an arithmetic shift is not a heredoc",
			line: "echo $((1<<20))\ncd scratch",
			want: []string{"0:echo|$", "0:1<<20", "1:cd|scratch"},
		},
		{
			name: "a herestring is an argument, not a heredoc",
			line: `grep foo <<< "rm -rf *"`,
			want: []string{"0:grep|foo|<<<|rm -rf *"},
		},
		{
			name: "empty line yields no segments",
			line: "   \t  ",
			want: nil,
		},
		{
			name: "adjacent quoting concatenates into one token",
			line: `git push '--force' origin`,
			want: []string{"0:git|push|--force|origin"},
		},
		{
			name: "empty quoted token is preserved",
			line: `grep "" file`,
			want: []string{"0:grep||file"},
		},
		{
			name: "a glued redirection terminates the word and drops the target",
			line: "git push --force>/dev/null",
			want: []string{"0:git|push|--force"},
		},
		{
			name: "a glued redirection with a spaced target keeps later words",
			line: "git push --force>out.txt origin main",
			want: []string{"0:git|push|--force|origin|main"},
		},
		{
			name: "a leading redirection does not displace the command",
			line: ">/dev/null git push --force origin main",
			want: []string{"0:git|push|--force|origin|main"},
		},
		{
			name: "an append redirection drops only its target",
			line: "echo a >> log b",
			want: []string{"0:echo|a|b"},
		},
		{
			name: "an fd-prefixed dup redirection is dropped whole",
			line: "go test ./... 2>&1",
			want: []string{"0:go|test|./..."},
		},
		{
			name: "process substitution keeps its prior handling",
			line: "cat <(echo hi)",
			want: []string{"0:cat|<", "0:echo|hi"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			segs, err := tokenize(tc.line)
			if err != nil {
				t.Fatalf("tokenize(%q): unexpected error: %v", tc.line, err)
			}
			got := render(segs)
			if len(got) != len(tc.want) {
				t.Fatalf("tokenize(%q) = %q, want %q", tc.line, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("tokenize(%q) segment %d = %q, want %q", tc.line, i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestTokenizeRejectsUnterminatedQuote(t *testing.T) {
	for _, line := range []string{`echo "unterminated`, `echo 'unterminated`, `echo trailing\`} {
		if _, err := tokenize(line); !errors.Is(err, ErrUnparsableCommand) {
			t.Fatalf("tokenize(%q) error = %v, want ErrUnparsableCommand", line, err)
		}
	}
}

// TestTokenizeRejectsUnterminatedHeredoc: a here-document whose delimiter line
// never appears is unparsable, not a silent swallow of the remaining input —
// the guard must fail open LOUDLY on it rather than allow whatever commands
// happened to follow with no signal at all.
func TestTokenizeRejectsUnterminatedHeredoc(t *testing.T) {
	const line = "cat <<EOF\nrm -rf *"
	if _, err := tokenize(line); !errors.Is(err, ErrUnparsableCommand) {
		t.Fatalf("tokenize(%q) error = %v, want ErrUnparsableCommand", line, err)
	}
}

// TestArithmeticShiftByIdentifierIsNotAHeredoc pins the identifier-operand
// form of the same boundary the literal-digit case above already covers: an
// unquoted `<<` whose "delimiter" word is immediately followed by a bare paren
// — e.g. `$((1<<shift))`, where readHeredocDelim reads "shift" and stops at
// the arithmetic expression's own closing paren — is never a real heredoc. A
// genuine delimiter word is always followed by its body and terminator line,
// never by `(` or `)` with no separator. Classifying this correctly at parse
// time (rather than merely erroring when no terminator line happens to exist)
// matters: an unquoted `<<` misread this way must still let every later line
// reach command position and be matched normally, not merely fail loud.
func TestArithmeticShiftByIdentifierIsNotAHeredoc(t *testing.T) {
	const line = "shift=8\necho $((1<<shift))\ngit push --force origin main"
	segs, err := tokenize(line)
	if err != nil {
		t.Fatalf("tokenize(%q): unexpected error: %v", line, err)
	}
	got := render(segs)
	swallowed := true
	for _, s := range got {
		if strings.Contains(s, "push") {
			swallowed = false
		}
	}
	if swallowed {
		t.Errorf("tokenize(%q) = %q: the `git push --force` line never reached command position", line, got)
	}

	d, err := Defaults().Check(line)
	if err != nil {
		t.Fatalf("Check(%q): unexpected error: %v", line, err)
	}
	if d.Verdict != VerdictBlock {
		t.Fatalf("Check(%q).Verdict = %q (entry %q), want %q via git-push-force", line, d.Verdict, d.EntryID, VerdictBlock)
	}

	// Control: the literal-operand form must still parse and block normally.
	const lit = "echo $((1<<20))\ngit push --force origin main"
	dl, err := Defaults().Check(lit)
	if err != nil {
		t.Fatalf("Check(%q): unexpected error: %v", lit, err)
	}
	if dl.Verdict != VerdictBlock {
		t.Fatalf("control: Check(%q).Verdict = %q, want %q", lit, dl.Verdict, VerdictBlock)
	}
}

// TestArithmeticShiftCoincidentalDelimiterStillBlocks is the adversarial case
// that TestArithmeticShiftByIdentifierIsNotAHeredoc alone would miss: an
// attacker who knows the tokenizer once looked for a line matching the
// misread "delimiter" could supply exactly that line (here, a bare `shift`
// with no arguments — a harmless no-op POSIX builtin), so the earlier,
// narrower fix (erroring only when no such line exists) would still let this
// swallow the guarded command with no error and no signal. Classifying the
// `<<` correctly up front — never treating it as a heredoc in the first place
// — closes this regardless of what later lines happen to contain.
func TestArithmeticShiftCoincidentalDelimiterStillBlocks(t *testing.T) {
	for _, line := range []string{
		"echo $((1<<shift))\ngit push --force origin main\nshift",
		"echo $((1<<k))\ngit push --force origin main\nk",
		"echo $(( $((1<<a)) ))\ngit push --force origin main\na",
	} {
		d, err := Defaults().Check(line)
		if err != nil {
			t.Fatalf("Check(%q): unexpected error: %v", line, err)
		}
		if d.Verdict != VerdictBlock {
			t.Fatalf("Check(%q).Verdict = %q, want %q — the guarded line must not be silently swallowed", line, d.Verdict, VerdictBlock)
		}
	}
}
