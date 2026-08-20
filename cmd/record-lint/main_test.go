package main

import (
	"os"
	"path/filepath"
	"testing"
)

// resolveRoot must discover the actual repository, not a tree an inherited
// GIT_WORK_TREE points at (iss-311). Before the env scrub, `git rev-parse
// --show-toplevel` honoured the injected work tree and record-lint (including
// --release-gate) could lint the wrong repository.
func TestResolveRootIgnoresInjectedWorkTree(t *testing.T) {
	decoy := t.TempDir()
	t.Setenv("GIT_WORK_TREE", decoy)
	t.Setenv("GIT_DIR", filepath.Join(decoy, ".git"))

	got := resolveRoot()
	if got == decoy {
		t.Fatalf("resolveRoot honoured the injected GIT_WORK_TREE (%q)", got)
	}
	// The real root carries this module's go.mod; the decoy does not.
	if _, err := os.Stat(filepath.Join(got, "go.mod")); err != nil {
		t.Fatalf("resolveRoot returned %q, which has no go.mod: %v", got, err)
	}
}
