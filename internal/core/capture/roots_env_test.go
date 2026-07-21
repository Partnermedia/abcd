package capture

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/REPPL/abcd-cli/internal/gittest"
)

// TestDiscoverRepoRootIgnoresInheritedWorkTree is the attack-input test for the
// discoverRepoRoot env scrub: an inherited GIT_WORK_TREE overrides cwd-based
// `rev-parse --show-toplevel` and would redirect repo-root discovery (and thus
// where capture reads/writes the issue ledger) at an attacker-chosen tree.
func TestDiscoverRepoRootIgnoresInheritedWorkTree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	repo := t.TempDir()
	gitInit := exec.Command("git", "-C", repo, "init")
	gitInit.Env = gittest.Env(t)
	if out, err := gitInit.CombinedOutput(); err != nil {
		t.Skipf("git init unavailable: %v (%s)", err, out)
	}
	other := t.TempDir()
	t.Setenv("GIT_WORK_TREE", other)

	got := discoverRepoRoot(repo)
	gotResolved, _ := filepath.EvalSymlinks(got)
	repoResolved, _ := filepath.EvalSymlinks(repo)
	if gotResolved != repoResolved {
		t.Errorf("discoverRepoRoot(%q) = %q under inherited GIT_WORK_TREE; want the real repo root %q (discovery was redirected)", repo, got, repoResolved)
	}
}
