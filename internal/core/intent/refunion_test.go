package intent

import (
	"testing"

	"github.com/REPPL/abcd-cli/internal/gittest"
)

// TestCreateFromTextMintsPastACommittedBranch reproduces the parallel-branch
// id-minting race (iss-80, fixed under iss-115 and iss-120): two branches cut
// from one base each seed a draft intent, and without the refs-union scan branch
// B re-mints the id branch A has already committed. After the fix, branch B
// mints past A's committed id.
func TestCreateFromTextMintsPastACommittedBranch(t *testing.T) {
	r := gittest.NewRepo(t)
	r.Write("README.md", "# base\n")
	r.Commit("base")

	// Branch A (main): mint itd-1 and commit it.
	itA, _, err := CreateFromText(r.Root(), "some seed text alpha", "")
	if err != nil {
		t.Fatalf("branch A CreateFromText: %v", err)
	}
	if itA.ID != "itd-1" {
		t.Fatalf("branch A minted %s, want itd-1", itA.ID)
	}
	r.Commit("A: " + itA.ID)

	// Branch B, cut from the base BEFORE A's commit: its working tree carries no
	// intent, so a tree-only mint would hand out itd-1 again.
	r.Git("checkout", "-b", "branch-b", "HEAD~1")

	itB, _, err := CreateFromText(r.Root(), "some seed text beta", "")
	if err != nil {
		t.Fatalf("branch B CreateFromText: %v", err)
	}
	// The class bug: itB.ID == "itd-1" (collides with A). The fix: the ref scan
	// sees A's committed itd-1 on refs/heads/main and mints itd-2.
	if itB.ID != "itd-2" {
		t.Fatalf("branch B minted %s, want itd-2 (collided with branch A's committed id)", itB.ID)
	}
}
