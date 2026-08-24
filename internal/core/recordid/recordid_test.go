package recordid

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/intentdriven/abcd/internal/gittest"
)

// TestMaxAcrossRefsSeesOtherBranches proves the ref scan finds an id committed on
// a branch whose file is absent from the current working tree — the core of the
// parallel-branch collision fix.
func TestMaxAcrossRefsSeesOtherBranches(t *testing.T) {
	r := gittest.NewRepo(t)
	r.Write("README.md", "# base\n")
	r.Commit("base")

	// main carries spc-1 and spc-3; a gap at spc-2 must not lower the max.
	r.Write(".abcd/development/specs/open/spc-1-alpha.md", "---\nid: spc-1\n---\n")
	r.Write(".abcd/development/specs/closed/spc-3-gamma.md", "---\nid: spc-3\n---\n")
	r.Commit("specs on main")

	// A branch cut from base has none of them in its working tree.
	r.Git("checkout", "-b", "feature", "HEAD~1")

	scan := MaxAcrossRefs(r.Root(), "spc", []string{".abcd/development/specs"})
	if scan.Degraded {
		t.Fatalf("scan degraded unexpectedly: %s", scan.Reason)
	}
	if scan.Max != 3 {
		t.Fatalf("MaxAcrossRefs = %d, want 3 (highest id committed on refs/heads/main)", scan.Max)
	}
}

// TestMaxAcrossRefsLoudWhenGitUnreadable proves the loud-degrade path: a
// repository is present (.git exists) but git cannot be run (PATH emptied, so the
// binary is unfindable) — the dangerous silent-collision case. The scan must
// degrade LOUDLY rather than fall back quietly.
func TestMaxAcrossRefsLoudWhenGitUnreadable(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", "") // git becomes unfindable for exec.LookPath

	scan := MaxAcrossRefs(dir, "spc", []string{".abcd/development/specs"})
	if !scan.Degraded {
		t.Fatal("a present repository git cannot read must degrade loudly, not fall back silently")
	}
	if scan.Warning() == "" {
		t.Fatal("a degraded scan must produce a non-empty Warning() for the surface to render")
	}
}

// TestMaxAcrossRefsQuietOutsideRepo proves a non-repository directory is NOT a
// loud degrade: there are no refs to collide against, so the scan stays quiet and
// the caller mints from the working tree alone.
func TestMaxAcrossRefsQuietOutsideRepo(t *testing.T) {
	dir := t.TempDir()
	scan := MaxAcrossRefs(dir, "spc", []string{".abcd/development/specs"})
	if scan.Degraded {
		t.Fatalf("non-repo directory must not degrade loudly; got reason %q", scan.Reason)
	}
	if w := scan.Warning(); w != "" {
		t.Fatalf("non-repo directory must emit no warning; got %q", w)
	}
	if scan.Max != 0 {
		t.Fatalf("non-repo Max = %d, want 0", scan.Max)
	}
}
