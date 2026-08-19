package spec

import (
	"testing"

	"github.com/Partnermedia/abcd/internal/gittest"
)

// TestCreateMintsPastACommittedBranch reproduces the parallel-branch id-minting
// race (iss-115, iss-120): two branches cut from one base each mint a spec, and
// without the refs-union scan branch B re-mints the id branch A has already
// committed. After the fix, branch B mints past A's committed id.
func TestCreateMintsPastACommittedBranch(t *testing.T) {
	r := gittest.NewRepo(t)
	r.Write("README.md", "# base\n")
	r.Commit("base")

	// Branch A (main): mint spc-1 and commit it.
	spA, _, err := Create(r.Root(), "itd-1", "alpha")
	if err != nil {
		t.Fatalf("branch A Create: %v", err)
	}
	if spA.ID != "spc-1" {
		t.Fatalf("branch A minted %s, want spc-1", spA.ID)
	}
	r.Commit("A: " + spA.ID)

	// Branch B, cut from the base BEFORE A's commit: its working tree carries no
	// spec, so a tree-only mint would hand out spc-1 again.
	r.Git("checkout", "-b", "branch-b", "HEAD~1")

	spB, _, err := Create(r.Root(), "itd-2", "beta")
	if err != nil {
		t.Fatalf("branch B Create: %v", err)
	}
	// The class bug: spB.ID == "spc-1" (collides with A). The fix: the ref scan
	// sees A's committed spc-1 on refs/heads/main and mints spc-2.
	if spB.ID != "spc-2" {
		t.Fatalf("branch B minted %s, want spc-2 (collided with branch A's committed id)", spB.ID)
	}
}
