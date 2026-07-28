package ahoy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// citationRepo plants a repo with a docs-lint config, one cited page, and
// optionally a baseline.
func citationRepo(t *testing.T, ruleEnabled bool, baseline string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	enabled := "false"
	if ruleEnabled {
		enabled = "true"
	}
	cfg := `{"roots": ["docs"], "banned_tokens": [], "rules": {
	  "citation_baseline": {"enabled": ` + enabled + `, "severity": "blocker",
	    "baseline": ".abcd/citations-baseline.json", "warn_after_days": 180, "block_after_days": 365}}}`
	if err := os.WriteFile(filepath.Join(root, ".abcd", "docs-lint.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	page := "# P\n\nA claim.[^a]\n\n[^a]: S, https://example.org/a\n"
	if err := os.WriteFile(filepath.Join(root, "docs", "p.md"), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
	if baseline != "" {
		if err := os.WriteFile(filepath.Join(root, ".abcd", "citations-baseline.json"), []byte(baseline), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// TestCitationSignalReportsCoverageAndAge is the ahoy status line spc-17 asks
// for: the baseline's age and coverage, visible without running the gate.
func TestCitationSignalReportsCoverageAndAge(t *testing.T) {
	root := citationRepo(t, true, `{"schema_version": 1, "entries": {
	  "https://example.org/a": {"final_url": "https://example.org/a", "last_checked": "2020-01-01",
	    "outcome": "alive", "verification": "automatic", "verified_on": "2020-01-01"}}}`)

	got := detectCitations(root)
	if got == "" {
		t.Fatal("no citations signal for a repo with the rule armed")
	}
	if !strings.Contains(got, "1 cited") {
		t.Errorf("signal = %q, want the cited count", got)
	}
	if !strings.Contains(got, "overdue") {
		t.Errorf("signal = %q, want the long-stale entry called out as overdue", got)
	}
}

// TestCitationSignalNamesAnUnrefreshedRepo makes the un-refreshed state visible
// rather than silent — a maintainer must be able to see that the record the gate
// enforces does not exist yet.
func TestCitationSignalNamesAnUnrefreshedRepo(t *testing.T) {
	got := detectCitations(citationRepo(t, true, ""))
	if !strings.Contains(got, "no baseline") {
		t.Errorf("signal = %q, want it to name the missing baseline", got)
	}
}

// TestCitationSignalIsSilentWhenTheRuleIsOff keeps the line out of every repo
// that has not adopted the gate. A status board that reports on a rule nobody
// armed is noise.
func TestCitationSignalIsSilentWhenTheRuleIsOff(t *testing.T) {
	if got := detectCitations(citationRepo(t, false, "")); got != "" {
		t.Errorf("signal = %q, want none when citation_baseline is disabled", got)
	}
}

// TestCitationSignalIsSilentWithoutAConfig keeps Detect total over an ordinary
// folder: no docs-lint config, no line, no error.
func TestCitationSignalIsSilentWithoutAConfig(t *testing.T) {
	if got := detectCitations(t.TempDir()); got != "" {
		t.Errorf("signal = %q, want none without a docs-lint config", got)
	}
}

// TestCitationSignalSurvivesAMalformedBaseline pins the fail-soft contract for a
// read-only status board: a corrupt record is REPORTED, never a hard error, so
// `ahoy` still renders everything else it knows.
func TestCitationSignalSurvivesAMalformedBaseline(t *testing.T) {
	root := citationRepo(t, true, `{"schema_version": 1, "entries": {"x": {"method": "curl"}}}`)
	got := detectCitations(root)
	if !strings.Contains(got, "unreadable") {
		t.Errorf("signal = %q, want it to report the unreadable baseline", got)
	}
}

// TestDetectSurfacesTheCitationSignal wires the check into the envelope both
// front doors read.
func TestDetectSurfacesTheCitationSignal(t *testing.T) {
	setupHermetic(t)
	root := citationRepo(t, true, "")
	det, err := Detect(root)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if s, _ := det.Signals["citations"].(string); s == "" {
		t.Fatalf("signals = %+v, want a citations entry", det.Signals)
	}
}
