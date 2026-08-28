package repolint_test

import (
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/repolint"
)

// The harness-leak class reaches `abcd lint`'s privacy rule over ANY committed
// file, not only over a freshly created pull-request body: the leak found in
// this repository was model-authored and landed in tracked prose, and the
// post-create strip cannot see a file that was committed rather than posted.
func TestAC_PrivacyHarnessLeakInCommittedFile(t *testing.T) {
	// Assembled at runtime for the reason the network specimens are: a literal
	// leak in this repo's own tree is the thing the rule exists to refuse.
	sessionURL := "https://agent-host.dev/code/" + strings.Join([]string{"session", "01Gy4Zo93PdMmggA8sfGyb"}, "_")
	footer := strings.Join([]string{"Generated", "with"}, " ") + " [Some Tool](https://tool.dev)"

	cases := []struct {
		name string
		body string
	}{
		{"live session url", "Run recorded at " + sessionURL + "\n"},
		{"tool attribution footer", footer + "\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := newFixtureRepo(t).conforming().
				file("reference/notes.md", c.body).
				commit()
			res := b.run()

			f := findingFor(res, "privacy-hygiene")
			if f == nil {
				t.Fatalf("no privacy-hygiene finding for %q", c.body)
			}
			if f.Severity != repolint.SeverityError {
				t.Errorf("severity = %q, want %q", f.Severity, repolint.SeverityError)
			}
			if f.File != "reference/notes.md" || f.Line != 1 {
				t.Errorf("citation = %s:%d, want reference/notes.md:1", f.File, f.Line)
			}
			if !strings.Contains(f.Fix, "Assisted-by") {
				t.Errorf("the finding does not name the sanctioned alternative; Fix = %q", f.Fix)
			}
		})
	}
}

// The negative half: prose ABOUT the ban, and a worked example pointing at a
// reserved documentation host, are how the convention gets written down. A rule
// that flags its own documentation is a rule people route around.
func TestAC_PrivacyHarnessLeakSparesProseAndExamples(t *testing.T) {
	body := strings.Join([]string{
		"We ban the \"" + strings.Join([]string{"Generated", "with"}, " ") + " [tool](url)\" footer in public text.",
		"- " + strings.Join([]string{"Generated", "with"}, " ") + " [Some Tool](https://example.invalid) is refused.",
		"A session link looks like https://example.invalid/code/session_01Gy4Zo93PdMmggA8sfGyb.",
	}, "\n") + "\n"

	b := newFixtureRepo(t).conforming().
		file("reference/attribution.md", body).
		commit()
	res := b.run()

	if f := findingFor(res, "privacy-hygiene"); f != nil {
		t.Fatalf("unexpected privacy-hygiene finding: %s:%d %s", f.File, f.Line, f.Message)
	}
}
