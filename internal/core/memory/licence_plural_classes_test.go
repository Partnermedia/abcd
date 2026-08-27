package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// lintPage writes one page into a fresh store and returns the lint result.
func lintPage(t *testing.T, filename, page string) LintResult {
	t.Helper()
	repo := t.TempDir()
	mem := Dir(repo)
	if err := os.MkdirAll(mem, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mem, filename), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
	lr, err := Lint(LintRequest{RepoRoot: repo, Now: fixedNow})
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	return lr
}

func hasBlocker(lr LintResult, code string) bool {
	for _, f := range lr.Findings {
		if f.Code == code && f.Severity == "blocker" {
			return true
		}
	}
	return false
}

// TestLicenceGateReadsPluralSourceClasses pins ML001 to the same class
// derivation every sibling check uses. checkLicence read the page's class only
// from the scalar source.class (or from source.sources[].class), never from the
// plural source.classes list that derivedClasses, SourceClasses and the index
// rendering all honour — so an external_* page declared `classes: [external_pdf]`
// with no licence passed lint with zero blockers, while MS001 in the same run
// asserted the page did carry an external class.
func TestLicenceGateReadsPluralSourceClasses(t *testing.T) {
	hash := strings.Repeat("a", 64)
	unlicensed := "---\ntopic: q3\nsource:\n  classes: [external_pdf]\n  source_hash: " + hash + "\n---\nBody.\n"

	lr := lintPage(t, "note_finance_q3.md", unlicensed)
	if !hasBlocker(lr, "ML001") || lr.ExitCode != 1 {
		t.Fatalf("unlicensed external page declared through source.classes passed lint: exit=%d findings=%+v", lr.ExitCode, lr.Findings)
	}

	// Control: the same page with a licence is clean of ML001 — the finding is
	// the missing licence's doing, not the plural shape's.
	licensed := "---\ntopic: q3\nsource:\n  classes: [external_pdf]\n  licence: MIT\n  source_hash: " + hash + "\n---\nBody.\n"
	if lr := lintPage(t, "note_finance_q3.md", licensed); hasBlocker(lr, "ML001") {
		t.Fatalf("a licensed plural-shape page was blocked: %+v", lr.Findings)
	}

	// Control: a non-external plural class needs no licence.
	internal := "---\ntopic: q3\nsource:\n  classes: [session_memory]\n---\nBody.\n"
	if lr := lintPage(t, "note_finance_q3.md", internal); hasBlocker(lr, "ML001") {
		t.Fatalf("a session_memory page was held to the licence requirement: %+v", lr.Findings)
	}

	// Control: a well-formed multi-source page, whose licences live per entry,
	// must not pick up a spurious page-level ML001 from its classes list.
	multi := "---\nsource:\n  classes: [external_pdf]\n  sources:\n" +
		"    -\n      class: external_pdf\n      citation: { type: knowledge }\n      licence: MIT\n      source_hash: " + hash + "\n      ingested_at: 2026-07-06\n" +
		"---\nBody.\n"
	if lr := lintPage(t, "note_finance_q3.md", multi); hasBlocker(lr, "ML001") {
		t.Fatalf("a fully licensed multi-source page was blocked: %+v", lr.Findings)
	}
}
