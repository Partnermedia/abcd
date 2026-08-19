package repolint_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Partnermedia/abcd/internal/core/repolint"
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
	// Severity follows the pattern set's documented split: addresses block,
	// hostname shapes are advisory.
	cases := []struct {
		name string
		body string
		sev  repolint.Severity
		exit int
	}{
		{"cgnat tailnet address", "peer reachable at " + quad(100, 64, 3, 9) + "\n", repolint.SeverityError, 2},
		{"private lan address", "gateway is " + quad(192, 168, 1, 1) + "\n", repolint.SeverityError, 2},
		{"lan hostname", "ssh into " + joinDots("printer", "local") + "\n", repolint.SeverityWarn, 1},
		{"device hostname", "synced from " + strings.Join([]string{"zeta", "laptop"}, "-") + "\n", repolint.SeverityWarn, 1},
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
			if f.Severity != c.sev {
				t.Errorf("severity = %q, want %q", f.Severity, c.sev)
			}
			if f.File != "reference/notes.md" || f.Line != 1 {
				t.Errorf("citation = %s:%d, want reference/notes.md:1", f.File, f.Line)
			}
			if res.ExitCode != c.exit {
				t.Errorf("exit = %d, want %d", res.ExitCode, c.exit)
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
// illustrative line can be kept without weakening the pattern set. The specimen
// must be one that genuinely flags — a masked CIDR base is exempt on its own and
// would make this test unfalsifiable — so the negative control asserts the same
// body without the escape does produce the finding.
func TestAC_PrivacyNetworkWaiverSuppresses(t *testing.T) {
	const body = "reachable at %s\n"
	specimen := quad(100, 64, 3, 9)

	waived := newFixtureRepo(t).conforming().
		file("reference/ranges.md", fmt.Sprintf(body, specimen+"  abcd-audit:allow")).
		commit().run()
	if f := findingFor(waived, "privacy-hygiene"); f != nil {
		t.Fatalf("waiver escape did not suppress the network finding: %+v", f)
	}
	if waived.ExitCode != 0 {
		t.Errorf("exit = %d, want 0 (waived)", waived.ExitCode)
	}

	// Negative control: the identical body without the escape must flag, or the
	// assertion above proves nothing.
	bare := newFixtureRepo(t).conforming().
		file("reference/ranges.md", fmt.Sprintf(body, specimen)).
		commit().run()
	if f := findingFor(bare, "privacy-hygiene"); f == nil {
		t.Fatal("negative control: the unwaived specimen did not flag, so the waiver test is vacuous")
	}
}

// C5: the audit surface honours the pattern set's documented severity split —
// addresses block, hostname shapes warn — rather than flattening everything to
// error.
func TestAC_PrivacyNetworkSeverityFollowsThePatternSet(t *testing.T) {
	addr := newFixtureRepo(t).conforming().
		file("reference/a.md", "peer "+quad(100, 64, 3, 9)+"\n").
		commit().run()
	f := findingFor(addr, "privacy-hygiene")
	if f == nil || f.Severity != repolint.SeverityError {
		t.Fatalf("address finding severity = %+v, want error", f)
	}
	if addr.ExitCode != 2 {
		t.Errorf("exit = %d, want 2 for a blocking finding", addr.ExitCode)
	}

	hostFinding := newFixtureRepo(t).conforming().
		file("reference/b.md", "ssh "+joinDots("printer", "local")+"\n").
		commit().run()
	h := findingFor(hostFinding, "privacy-hygiene")
	if h == nil || h.Severity != repolint.SeverityWarn {
		t.Fatalf("hostname finding severity = %+v, want warn", h)
	}
	if hostFinding.ExitCode != 1 {
		t.Errorf("exit = %d, want 1 for a warning-only finding", hostFinding.ExitCode)
	}
}

// F6: the pattern set says a repo that wants the hostname shapes to block can
// raise their severity in .abcd/config/pii.json. The audit rule must therefore
// read the same MERGED set the scanner builds for this repo — binding the
// built-in set made that documented override a no-op at the one surface a
// maintainer meets it.
func TestAC_PrivacyNetworkHonoursRepoSeverityOverride(t *testing.T) {
	const cfg = `{"patterns":{"net_lan_hostname":{"severity":"hard_fail"}}}` + "\n"
	res := newFixtureRepo(t).conforming().
		file(".abcd/config/pii.json", cfg).
		file("reference/notes.md", "ssh into "+joinDots("printer", "local")+"\n").
		commit().run()

	f := findingFor(res, "privacy-hygiene")
	if f == nil {
		t.Fatal("no privacy-hygiene finding for a LAN hostname")
	}
	if f.Severity != repolint.SeverityError {
		t.Errorf("severity = %q, want %q — the repo raised this pattern to hard_fail", f.Severity, repolint.SeverityError)
	}
	if res.ExitCode != 2 {
		t.Errorf("exit = %d, want 2 for a raised, blocking finding", res.ExitCode)
	}
}

// S3: an exempt system directory must not shield a username nested under it.
func TestAC_PrivacyNestedUsernameUnderSystemDirectory(t *testing.T) {
	cases := []struct {
		name string
		body string
		want bool
	}{
		{"nested username", "keys at /Users/Shared/" + strings.Join([]string{"j", "doe"}, "") + "/keys.txt\n", true},
		{"nested non-username segment", "data at /Users/Shared/abcd-data/x\n", true}, // abcd-audit:allow
		{"bare system directory", "the installer writes to /Users/Shared\n", false},
		{"system directory trailing slash", "the installer writes to /Users/Shared/ and stops\n", false},
		// Prose, not a path: an ellipsis after the system directory is not a
		// nested username, and a segment of pure dots never names a user.
		{"prose ellipsis", "privacy-hygiene flags /Users/Shared/... in committed files\n", false},
		{"relative marker alone", "the installer writes to /Users/Shared/.\n", false},
		// F3b: a segment of pure dots names no user, but it must not END the
		// search either — a dots-only or an empty segment between the system
		// directory and a name re-creates the very shield the exemption forbids.
		{"relative marker then name", "keys at /Users/Shared/./" + strings.Join([]string{"j", "doe"}, "") + "/keys.txt\n", true},
		{"parent marker then name", "keys at /Users/Shared/../" + strings.Join([]string{"j", "doe"}, "") + "/keys.txt\n", true},
		{"doubled separator then name", "keys at /Users/Shared//" + strings.Join([]string{"j", "doe"}, "") + "/keys.txt\n", true},
		// F3: absPathRe matches the Windows spelling too, so the same nested-name
		// semantics have to hold on a backslash separator — otherwise the system
		// directory shields a username there while flagging it on POSIX.
		{"windows nested username", `keys at C:\Users\Public\` + strings.Join([]string{"j", "doe"}, "") + `\keys.txt` + "\n", true},
		{"windows parent marker then name", `keys at C:\Users\Public\..\` + strings.Join([]string{"j", "doe"}, "") + "\n", true},
		{"windows bare system directory", `the installer writes to C:\Users\Public` + "\n", false},
		{"windows system directory trailing separator", `the installer writes to C:\Users\Public\ and stops` + "\n", false},
		// Parity, deliberately: a plain file under the system directory flags on
		// BOTH spellings, exactly as its POSIX twin /Users/Shared/abcd-data/x  abcd-audit:allow
		// above does. The exemption covers the system directory itself, never a
		// name-bearing segment beneath it.
		{"windows file under system directory", `report at C:\Users\Public\report.txt` + "\n", true}, // abcd-audit:allow
		{"posix file under system directory", "report at /Users/Shared/report.txt\n", true},          // abcd-audit:allow
		// G3: Windows accepts BOTH separators in one path, so a nested name written
		// after a forward slash is a name the system directory must not shield.
		// Picking a single separator for the walk exempted the whole mixed spelling.
		{"windows mixed separator then name", `keys at C:\Users\Shared/` + strings.Join([]string{"j", "doe"}, "") + `/keys.txt` + "\n", true},
		{"windows mixed separator then marker and name", `keys at C:\Users\Public/../` + strings.Join([]string{"j", "doe"}, "") + "\n", true},
		// The POSIX walk stays slash-only: a backslash after a POSIX path is a Go
		// string escape, not a segment, and reading it as one flagged the escape.
		{"posix path then a string escape", `the tier lives at /Users/Shared\n and stops` + "\n", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := newFixtureRepo(t).conforming().
				file("reference/paths.md", c.body).
				commit().run()
			got := findingFor(res, "privacy-hygiene") != nil
			if got != c.want {
				t.Fatalf("finding = %v, want %v for %q", got, c.want, c.body)
			}
		})
	}
}

// iss-153: /Users/Shared and /Users/Guest are macOS system directories, not
// usernames. Product code that legitimately names them must not need a waiver.
func TestAC_PrivacySharedAndGuestAreNotUsernames(t *testing.T) {
	body := "the installer writes under /Users/Shared and never /Users/Guest\n"
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

// The exemption is narrow in two independent ways, so each is asserted on its
// OWN line: a segment that merely BEGINS with a system-directory name is not
// exempt, and a name nested under the system directory is not shielded by it.
// Carried on one line, the first path satisfied the assertion by itself and the
// second proved nothing.
func TestAC_PrivacyRealUsernameStillFlaggedAlongsideExemption(t *testing.T) {
	cases := []struct{ name, body string }{
		{"segment merely beginning with a system directory name", "notes at /Users/sharedstuff/notes.md\n"}, // abcd-audit:allow — the specimen IS the case under test
		{"name nested under a system directory", "notes at /Users/Shared/abcd/notes.md\n"},                  // abcd-audit:allow — the specimen IS the case under test
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := newFixtureRepo(t).conforming().
				file("reference/paths.md", c.body).
				commit().run()

			f := findingFor(res, "privacy-hygiene")
			if f == nil {
				t.Fatalf("%q was not flagged", c.body)
			}
			if f.File != "reference/paths.md" || f.Line != 1 {
				t.Errorf("citation = %s:%d, want reference/paths.md:1", f.File, f.Line)
			}
		})
	}
}
