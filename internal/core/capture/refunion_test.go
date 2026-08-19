package capture

import (
	"path/filepath"
	"testing"

	"github.com/Partnermedia/abcd/internal/gittest"
)

// TestCaptureMintsPastACommittedBranch reproduces the parallel-branch id-minting
// race (iss-80, fixed under iss-115 and iss-120): two branches cut from one base
// each capture an issue, and without the refs-union scan branch B re-mints the
// id branch A has already committed. After the fix, branch B mints past A's
// committed id.
func TestCaptureMintsPastACommittedBranch(t *testing.T) {
	r := gittest.NewRepo(t)
	r.Write("README.md", "# base\n")
	r.Commit("base")

	ledger := filepath.Join(r.Root(), LedgerRelPath)

	// Branch A (main): mint iss-1 and commit it.
	resA, err := Capture(CaptureRequest{
		RepoRoot:    r.Root(),
		IssuesRoot:  ledger,
		Text:        "alpha observation",
		Severity:    SeverityMinor,
		Category:    "bug",
		Source:      "user-observation",
		FoundDuring: "t",
		Slug:        "alpha",
	})
	if err != nil {
		t.Fatalf("branch A Capture: %v", err)
	}
	if resA.ID != "iss-1" {
		t.Fatalf("branch A minted %s, want iss-1", resA.ID)
	}
	r.Commit("A: " + resA.ID)

	// Branch B, cut from the base BEFORE A's commit: its working tree carries no
	// issue file, so a tree-only mint would hand out iss-1 again.
	r.Git("checkout", "-b", "branch-b", "HEAD~1")

	resB, err := Capture(CaptureRequest{
		RepoRoot:    r.Root(),
		IssuesRoot:  ledger,
		Text:        "beta observation",
		Severity:    SeverityMinor,
		Category:    "bug",
		Source:      "user-observation",
		FoundDuring: "t",
		Slug:        "beta",
	})
	if err != nil {
		t.Fatalf("branch B Capture: %v", err)
	}
	// The class bug: resB.ID == "iss-1" (collides with A). The fix: the ref scan
	// sees A's committed iss-1 on refs/heads/main and mints iss-2.
	if resB.ID != "iss-2" {
		t.Fatalf("branch B minted %s, want iss-2 (collided with branch A's committed id)", resB.ID)
	}
}
