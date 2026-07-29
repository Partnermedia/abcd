package audit_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/audit"
)

// Flaggable specimens are assembled at runtime (see the scanner's network
// corpus for the reasoning): a literal would leave a non-reserved identifier in
// this repo's own tree. Reserved documentation values are written plainly.

func quad(a, b, c, d int) string { return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d) }

func joinDots(labels ...string) string { return strings.Join(labels, ".") }

// privacy-hygiene must flag a committed network identifier that sits outside the
// reserved documentation ranges — the gap the 2026-07-29 field incident exposed,
// where a tailnet address and two device names passed the audit silently.
func TestAC_PrivacyNetworkIdentifierOutsideReservedRanges(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"cgnat tailnet address", "peer reachable at " + quad(100, 64, 3, 9) + "\n"},
		{"private lan address", "gateway is " + quad(192, 168, 1, 1) + "\n"},
		{"lan hostname", "ssh into " + joinDots("printer", "local") + "\n"},
		{"device hostname", "synced from " + strings.Join([]string{"zeta", "laptop"}, "-") + "\n"},
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
			if f.Severity != audit.SeverityError {
				t.Errorf("severity = %q, want error", f.Severity)
			}
			if f.File != "reference/notes.md" || f.Line != 1 {
				t.Errorf("citation = %s:%d, want reference/notes.md:1", f.File, f.Line)
			}
			if res.ExitCode != 2 {
				t.Errorf("exit = %d, want 2", res.ExitCode)
			}
		})
	}
}

// The inversion's negative half at the audit surface: reserved documentation
// values, standard protocol values, and persona-derived device names are the
// only identifiers a committed file may carry, and none of them is a finding.
func TestAC_PrivacyReservedIdentifiersAreClean(t *testing.T) {
	body := strings.Join([]string{
		"bind 127.0.0.1 and 0.0.0.0",
		"peers 192.0.2.1, 198.51.100.7, 203.0.113.42",
		"v6 peer 2001:db8::1 and loopback ::1",
		"hw 00:00:5E:00:53:00",
		"see example.com, api.example, host.test",
		"synced from alice-laptop and bob-desktop",
	}, "\n") + "\n"
	b := newFixtureRepo(t).conforming().
		file("reference/reserved.md", body).
		commit()
	res := b.run()

	if f := findingFor(res, "privacy-hygiene"); f != nil {
		t.Fatalf("reserved documentation identifiers flagged: %+v", f)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit = %d, want 0", res.ExitCode)
	}
}

// The waiver escape covers the network class too, so a deliberately
// illustrative line can be kept without weakening the pattern set.
func TestAC_PrivacyNetworkWaiverSuppresses(t *testing.T) {
	body := "range example " + quad(100, 64, 0, 0) + "/10 is CGNAT  abcd-audit:allow\n"
	b := newFixtureRepo(t).conforming().
		file("reference/ranges.md", body).
		commit()
	res := b.run()

	if f := findingFor(res, "privacy-hygiene"); f != nil {
		t.Fatalf("waiver escape did not suppress the network finding: %+v", f)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit = %d, want 0 (waived)", res.ExitCode)
	}
}

// iss-153: /Users/Shared and /Users/Guest are macOS system directories, not
// usernames. Product code that legitimately names them must not need a waiver.
func TestAC_PrivacySharedAndGuestAreNotUsernames(t *testing.T) {
	body := "the installer writes to /Users/Shared/abcd and never /Users/Guest\n"
	b := newFixtureRepo(t).conforming().
		file("reference/install.md", body).
		commit()
	res := b.run()

	if f := findingFor(res, "privacy-hygiene"); f != nil {
		t.Fatalf("/Users/Shared or /Users/Guest flagged as a username: %+v", f)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit = %d, want 0", res.ExitCode)
	}
}

// The exemption is narrow: a real username under /Users is still a leak, and a
// segment that merely starts with a system-directory name is not exempt.
func TestAC_PrivacyRealUsernameStillFlaggedAlongsideExemption(t *testing.T) {
	body := "/Users/Shared/abcd is fine but /Users/sharedstuff/notes.md is a leak\n" // abcd-audit:allow — the specimen IS the exemption under test
	b := newFixtureRepo(t).conforming().
		file("reference/paths.md", body).
		commit()
	res := b.run()

	if f := findingFor(res, "privacy-hygiene"); f == nil {
		t.Fatal("a real username under /Users was not flagged")
	}
}
