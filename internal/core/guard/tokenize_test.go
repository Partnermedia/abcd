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
			name: "heredoc body is data, not commands",
			line: "cat > doc.md <<'EOF'\ngit push --force\nEOF",
			want: []string{"0:cat|>|doc.md"},
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

// TestArithmeticShiftByIdentifierIsRejected pins the identifier-operand form of
// the same boundary the literal-digit case above already covers: an unquoted
// `<<` whose "delimiter" word is never terminated by a matching line — because
// it was never a real heredoc delimiter to begin with, e.g. `$((1<<shift))` —
// must not silently swallow the rest of the command as unchecked body text. It
// is unparsable, exactly like an unterminated quote, so tokenize errors rather
// than dropping the remaining lines from command position.
func TestArithmeticShiftByIdentifierIsRejected(t *testing.T) {
	const line = "shift=8\necho $((1<<shift))\ngit push --force origin main"
	if _, err := tokenize(line); !errors.Is(err, ErrUnparsableCommand) {
		t.Fatalf("tokenize(%q) error = %v, want ErrUnparsableCommand", line, err)
	}

	// The guard must not silently allow the command either: Check surfaces the
	// same error rather than returning VerdictAllow.
	if _, err := Defaults().Check(line); !errors.Is(err, ErrUnparsableCommand) {
		t.Fatalf("Check(%q) error = %v, want ErrUnparsableCommand", line, err)
	}

	// Control: the literal-operand form must still parse and block normally.
	const lit = "echo $((1<<20))\ngit push --force origin main"
	d, err := Defaults().Check(lit)
	if err != nil {
		t.Fatalf("Check(%q): unexpected error: %v", lit, err)
	}
	if d.Verdict != VerdictBlock {
		t.Fatalf("control: Check(%q).Verdict = %q, want %q", lit, d.Verdict, VerdictBlock)
	}
}
