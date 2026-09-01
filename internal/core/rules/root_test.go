package rules

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/intentdriven/abcd/internal/gittest"
)

func gitInitAt(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "-C", dir, "init", "--initial-branch=main")
	cmd.Env = gittest.Env(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("git init unavailable: %v (%s)", err, out)
	}
}

func realPath(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", p, err)
	}
	return r
}

// ResolveRoot is bounded at the git working tree (GHSA-vvqc-3mv2-5p49): the
// nearest .abcd directory INSIDE the tree wins, a tree with none resolves to
// its own toplevel, an ancestor's .abcd is never reached, and outside git there
// is no walk at all.
func TestResolveRootIsBoundedByTheWorkingTree(t *testing.T) {
	outer := t.TempDir()
	if err := os.MkdirAll(filepath.Join(outer, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	inner := filepath.Join(outer, "inner-repo")
	gitInitAt(t, inner)
	sub := filepath.Join(inner, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	wantInner := realPath(t, inner)

	if got := realPath(t, ResolveRoot(sub)); got != wantInner {
		t.Errorf("ResolveRoot(%q) = %q, escaped the working tree %q", sub, got, wantInner)
	}
	if err := os.MkdirAll(filepath.Join(inner, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := realPath(t, ResolveRoot(sub)); got != wantInner {
		t.Errorf("ResolveRoot(%q) = %q, want the tree's own .abcd at %q", sub, got, wantInner)
	}
	mid := filepath.Join(inner, "a")
	if err := os.MkdirAll(filepath.Join(mid, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got, want := realPath(t, ResolveRoot(sub)), realPath(t, mid); got != want {
		t.Errorf("ResolveRoot(%q) = %q, want the nearest .abcd inside the tree %q", sub, got, want)
	}

	plain := filepath.Join(outer, "plain")
	if err := os.MkdirAll(plain, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := ResolveRoot(plain); got != plain {
		t.Errorf("ResolveRoot(%q) = %q, want cwd (no walk outside git)", plain, got)
	}
}
