package repolint_test

import (
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/repolint"
)

// A broken per-repo scanner override must be REPORTED by privacy-hygiene, not
// absorbed (iss-203).
//
// The bug this pins was a guard that could not fail. The rule read
//
//	if sc, err := scanner.New(ctx.RepoRoot); err == nil { patterns = sc.NetworkPatterns() }
//
// on the stated assumption that a scanner which "cannot be built falls back to
// the built-in set". scanner.New returns a nil error on EVERY degradation path —
// an unreadable, unparseable or uncompilable pii.json each yield a usable
// scanner marked unavailable — so the guard was always true and the fallback it
// appeared to protect was dead code.
//
// What actually happened on a broken override: the merge failed, the scanner
// kept the built-in defaults, and any severity the repo had RAISED silently
// stopped applying. The scan then reported "conforms" over a weaker pattern set
// than the repo asked for, which is the didn't-scan-reported-clean shape this
// rule's own contract forbids — and the operator who raised the severity is the
// one person who would never find out it had lapsed.
//
// Both halves are asserted, because only the pair distinguishes a real report
// from a rule that flags everything: a broken override is a finding that cites
// the config and carries the reason, and a valid one is silent.
func TestPrivacyReportsAnUnusableScannerConfig(t *testing.T) {
	for _, c := range []struct {
		name string
		body string
	}{
		{"unparseable json", "{ this is not json"},
		{"uncompilable override regex", `{"patterns":{"bad":{"regex":"([unclosed","severity":"hard_fail"}}}`},
		// The case an adversarial review demonstrated escaping this rule: a pull
		// request that BLANKS a custom detector's regex removed it entirely, and
		// the scan then reported "conforms" at exit 0 over a file the detector had
		// been catching. An empty regex on a new pattern name detects nothing, so
		// it is a config fault rather than an omission.
		{"blanked override regex", `{"patterns":{"corp_host":{"regex":"","severity":"hard_fail"}}}`},
		// A NEW pattern name with no regex at all: same defect, written as an
		// omission rather than a blanking. It detects nothing, so it is a config
		// fault, and the merge must fail closed rather than skip it.
		{"new pattern with no regex", `{"patterns":{"mine":{"severity":"hard_fail"}}}`},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := newFixtureRepo(t).conforming().
				file(".abcd/config/pii.json", c.body).
				commit()
			res := b.run()

			f := findingFor(res, "privacy-hygiene")
			if f == nil {
				t.Fatalf("a %s override produced no privacy-hygiene finding; the scan ran with "+
					"silently weakened severities and said nothing", c.name)
			}
			if f.Severity != repolint.SeverityError {
				t.Errorf("severity = %q, want %q — a weakened privacy scan is not advisory",
					f.Severity, repolint.SeverityError)
			}
			if f.File != ".abcd/config/pii.json" {
				t.Errorf("citation = %q, want the config that caused it", f.File)
			}
			// The reason is what makes the finding actionable: "unusable" alone
			// sends the reader looking, the reason tells them what to repair.
			if strings.TrimSpace(f.Message) == "" || !strings.Contains(f.Message, ":") {
				t.Errorf("message = %q, want the scanner's reason appended", f.Message)
			}
		})
	}
}

// The negative half. A repo with a VALID override must produce no degradation
// finding — otherwise the check above would pass on a rule that simply always
// fires, and the corpus would prove nothing.
//
// The second case is the load-bearing one for the empty-regex refusal. Raising a
// BUNDLED pattern's severity legitimately carries no regex, because the regex is
// the built-in one — so the refusal must bind new names only. Without this case
// the fix could be over-broad and nothing would say so.
func TestPrivacyIsSilentOnAUsableScannerConfig(t *testing.T) {
	for _, c := range []struct{ name, body string }{
		{"skip list only", `{"skip_extensions":[".png"]}`},
		{"bundled pattern severity raise, no regex", `{"patterns":{"net_ipv4":{"severity":"hard_fail"}}}`},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := newFixtureRepo(t).conforming().
				file(".abcd/config/pii.json", c.body).
				commit()
			res := b.run()

			if f := findingFor(res, "privacy-hygiene"); f != nil {
				t.Fatalf("a valid override was reported as unusable: %+v", f)
			}
			if res.ExitCode != 0 {
				t.Errorf("exit = %d, want 0", res.ExitCode)
			}
		})
	}
}
