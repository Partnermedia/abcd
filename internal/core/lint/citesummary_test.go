package lint

import (
	"os"
	"path/filepath"
	"testing"
)

// summaryRepo plants a repo citing the given URLs with a baseline whose entries
// carry the given last-checked dates.
func summaryRepo(t *testing.T, entries map[string]BaselineEntry, cited ...string) string {
	t.Helper()
	files := map[string]string{}
	for i, u := range cited {
		files["docs/p"+string(rune('a'+i))+".md"] = "# P\n\nA claim.[^a]\n\n[^a]: S, " + u + "\n"
	}
	root := citeRepo(t, files)
	if entries != nil {
		path := filepath.Join(root, filepath.FromSlash(DefaultBaselinePath))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := SaveBaseline(path, Baseline{SchemaVersion: 1, Entries: entries}); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func summaryConfig() Config {
	return Config{Roots: []string{"docs"}, Rules: map[string]RuleConfig{
		ruleCitationBaseline: {Enabled: true, Severity: severityBlocker, WarnAfterDays: 180, BlockAfterDays: 365},
	}}
}

func alive(lastChecked, verification string) BaselineEntry {
	return BaselineEntry{FinalURL: "x", LastChecked: lastChecked, Outcome: OutcomeAlive,
		Verification: verification, VerifiedOn: lastChecked}
}

// TestCitationAgeSummaryGradesEveryBand is the one primitive the ahoy status line
// and the release nag both read. Grading it in one place is what keeps the number
// a maintainer sees at `ahoy` identical to the one the release gate acts on.
func TestCitationAgeSummaryGradesEveryBand(t *testing.T) {
	root := summaryRepo(t, map[string]BaselineEntry{
		// fixedNow is 2026-07-27.
		"https://e.org/fresh":       alive("2026-07-01", VerificationAutomatic), // 26 days
		"https://e.org/stale":       alive("2025-12-01", VerificationAutomatic), // 238 days: past warn
		"https://e.org/approaching": alive("2025-08-05", VerificationManual),    // 356 days: inside the nag window
		"https://e.org/overdue":     alive("2025-01-01", VerificationAutomatic), // 572 days: past block
		"https://e.org/broken": {FinalURL: "x", LastChecked: "2026-07-01", Outcome: OutcomeBroken,
			Verification: VerificationAutomatic, VerifiedOn: "2026-07-01"},
	},
		"https://e.org/fresh", "https://e.org/stale", "https://e.org/approaching",
		"https://e.org/overdue", "https://e.org/broken", "https://e.org/unrecorded",
	)

	got, err := CitationAgeSummaryAt(summaryConfig(), root, fixedNow)
	if err != nil {
		t.Fatalf("CitationAgeSummaryAt: %v", err)
	}
	if !got.Present {
		t.Fatal("Present = false, want true")
	}
	if got.Cited != 6 {
		t.Errorf("Cited = %d, want 6", got.Cited)
	}
	if got.Recorded != 5 {
		t.Errorf("Recorded = %d, want 5", got.Recorded)
	}
	if got.Missing != 1 {
		t.Errorf("Missing = %d, want 1", got.Missing)
	}
	if got.Broken != 1 {
		t.Errorf("Broken = %d, want 1", got.Broken)
	}
	if got.Manual != 1 {
		t.Errorf("Manual = %d, want 1 (a manual receipt is counted, and ages the same)", got.Manual)
	}
	// stale counts everything past the warn line, overdue everything past the
	// block line; approaching is the nag band just short of it.
	if got.Stale != 3 {
		t.Errorf("Stale = %d, want 3 (stale + approaching + overdue are all past warn)", got.Stale)
	}
	if got.Overdue != 1 {
		t.Errorf("Overdue = %d, want 1", got.Overdue)
	}
	if got.Approaching != 1 {
		t.Errorf("Approaching = %d, want 1", got.Approaching)
	}
	if len(got.OverdueURLs) != 1 || got.OverdueURLs[0] != "https://e.org/overdue" {
		t.Errorf("OverdueURLs = %v, want the overdue citation", got.OverdueURLs)
	}
	if len(got.ApproachingURLs) != 1 || got.ApproachingURLs[0] != "https://e.org/approaching" {
		t.Errorf("ApproachingURLs = %v, want the approaching citation", got.ApproachingURLs)
	}
	if got.OldestChecked != "2025-01-01" {
		t.Errorf("OldestChecked = %q, want 2025-01-01", got.OldestChecked)
	}
}

// TestCitationAgeSummaryWithNoBaseline reports the un-refreshed repo honestly
// rather than erroring, so a status line can say "no baseline yet".
func TestCitationAgeSummaryWithNoBaseline(t *testing.T) {
	root := summaryRepo(t, nil, "https://e.org/a")
	got, err := CitationAgeSummaryAt(summaryConfig(), root, fixedNow)
	if err != nil {
		t.Fatalf("CitationAgeSummaryAt: %v", err)
	}
	if got.Present {
		t.Error("Present = true, want false")
	}
	if got.Cited != 1 || got.Missing != 1 {
		t.Errorf("Cited=%d Missing=%d, want 1/1", got.Cited, got.Missing)
	}
}

// TestArmCitationOverdue is the release-gate promotion, built on the shape
// ArmReceiptGate established: the CALLER is the trust root, so a downgraded
// in-tree config cannot defang the gate at release time.
func TestArmCitationOverdue(t *testing.T) {
	cfg := Config{Rules: map[string]RuleConfig{
		ruleCitationBaseline: {Enabled: true, Severity: severityBlocker, OverdueSeverity: severityWarn},
	}}
	armed := ArmCitationOverdue(cfg)

	if got := armed.Rules[ruleCitationBaseline].OverdueSeverity; got != severityBlocker {
		t.Fatalf("overdue severity = %q, want %q", got, severityBlocker)
	}
	// The input config is not mutated: a caller may lint twice, once each way.
	if got := cfg.Rules[ruleCitationBaseline].OverdueSeverity; got != severityWarn {
		t.Fatalf("arming mutated the caller's config: overdue severity = %q", got)
	}
	// Arming does not enable a rule the repo turned off — the release gate
	// promotes a severity, it does not adopt a rule on the repo's behalf.
	off := ArmCitationOverdue(Config{Rules: map[string]RuleConfig{
		ruleCitationBaseline: {Enabled: false},
	}})
	if off.Rules[ruleCitationBaseline].Enabled {
		t.Fatal("arming enabled a rule the repo disabled")
	}
}

// TestArmCitationOverdueBlocksAtTheReleaseGate is the end-to-end proof: the same
// tree that only warns for a commit blocks once armed.
func TestArmCitationOverdueBlocksAtTheReleaseGate(t *testing.T) {
	root := summaryRepo(t, map[string]BaselineEntry{
		"https://e.org/old": alive("2025-01-01", VerificationAutomatic),
	}, "https://e.org/old")

	commit, err := LintAt(summaryConfig(), root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	for _, f := range findingsFor(commit, ruleCitationBaselineOverdu) {
		if f.Severity != severityWarn {
			t.Fatalf("commit gate severity = %q, want %q — commits are never calendar-blocked", f.Severity, severityWarn)
		}
	}

	release, err := LintAt(ArmCitationOverdue(summaryConfig()), root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	armed := findingsFor(release, ruleCitationBaselineOverdu)
	if len(armed) != 1 || armed[0].Severity != severityBlocker {
		t.Fatalf("release gate findings = %+v, want one blocker", armed)
	}
}
