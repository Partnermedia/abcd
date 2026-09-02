package rules

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/intentdriven/abcd/internal/gittest"
	"github.com/intentdriven/abcd/internal/gitutil"
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

// mustDir creates dir (and its parents) or fails the test.
func mustDir(t *testing.T, dir string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// refuseOwnership stages, deterministically, the state abcd's own isolated git
// environment cannot be rescued from: a real repository git will not answer
// for. GIT_TEST_ASSUME_DIFFERENT_OWNER makes git's ownership check fire, and
// the normal escape — `safe.directory` in the developer's global config — is
// discarded by the GIT_CONFIG_GLOBAL=/dev/null, GIT_CONFIG_NOSYSTEM=1 that
// gitutil.Run runs every command under. It is the same shape as a checkout
// owned by another uid, a container bind mount, or a corrupt .git.
func refuseOwnership(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("GIT_TEST_ASSUME_DIFFERENT_OWNER", "1")
	if _, err := gitutil.Run(dir, "rev-parse", "--show-toplevel"); err == nil {
		t.Skip("this git ignores GIT_TEST_ASSUME_DIFFERENT_OWNER; the refusal could not be staged")
	}
}

// TestResolveRootFallsBackToTheGitMarkerWhenGitRefuses is the second half of the
// GHSA-vvqc-3mv2-5p49 bound. "Not a repository" and "a repository git will not
// answer for" are different states, and only the first is safe to resolve as
// cwd-with-no-walk: in the second, content here IS version-controlled and the
// repo's own .abcd is the governing configuration. Collapsing the two dropped
// it silently — from a subdirectory the root became the subdirectory, so the
// repo's rules.json, its kill switch and its guard.json all stopped applying
// and nothing said so. The .git marker bounds the resolution when git will not.
func TestResolveRootFallsBackToTheGitMarkerWhenGitRefuses(t *testing.T) {
	outer := mustDir(t, t.TempDir())
	mustDir(t, filepath.Join(outer, ".abcd"))
	inner := filepath.Join(outer, "inner-repo")
	gitInitAt(t, inner)
	sub := mustDir(t, filepath.Join(inner, "internal", "deep"))
	wantInner := realPath(t, inner)

	refuseOwnership(t, sub)

	// No .abcd inside the tree: the marker root, never cwd and never the
	// ancestor's plant.
	if got := realPath(t, ResolveRoot(sub)); got != wantInner {
		t.Errorf("ResolveRoot(%q) with git refusing = %q, want the .git-marker root %q", sub, got, wantInner)
	}
	// The repo's own .abcd is what a subdirectory session must read.
	mustDir(t, filepath.Join(inner, ".abcd"))
	if got := realPath(t, ResolveRoot(sub)); got != wantInner {
		t.Errorf("ResolveRoot(%q) with git refusing = %q, want the repo's own .abcd at %q", sub, got, wantInner)
	}
	// Nearest-first still holds inside the tree.
	mid := mustDir(t, filepath.Join(inner, "internal"))
	mustDir(t, filepath.Join(mid, ".abcd"))
	if got, want := realPath(t, ResolveRoot(sub)), realPath(t, mid); got != want {
		t.Errorf("ResolveRoot(%q) with git refusing = %q, want the nearest .abcd inside the tree %q", sub, got, want)
	}
	// And the bound is still a bound: a genuinely non-repo directory does not
	// walk to the .abcd planted above it.
	plain := mustDir(t, filepath.Join(outer, "plain"))
	if got := ResolveRoot(plain); got != plain {
		t.Errorf("ResolveRoot(%q) = %q, want cwd (a non-repo directory must not walk)", plain, got)
	}
}

// TestResolveRootFallsBackToTheGitMarkerWhenGitIsAbsent is the same bound with
// the other unanswerable cause: git off the PATH the host hook runs under. A
// hook launched from a GUI session or a stripped container has no git at all,
// and every repository on that machine would otherwise resolve its rules and
// its guard from whatever directory the session happened to start in.
func TestResolveRootFallsBackToTheGitMarkerWhenGitIsAbsent(t *testing.T) {
	outer := mustDir(t, t.TempDir())
	mustDir(t, filepath.Join(outer, ".abcd"))
	inner := filepath.Join(outer, "inner-repo")
	gitInitAt(t, inner) // needs git, so it happens before the PATH is emptied
	mustDir(t, filepath.Join(inner, ".abcd"))
	sub := mustDir(t, filepath.Join(inner, "pkg"))
	wantInner := realPath(t, inner)

	t.Setenv("PATH", "/nonexistent")
	if out, err := gitutil.Run(sub, "rev-parse", "--show-toplevel"); err == nil {
		t.Fatalf("git answered %q with an emptied PATH; the fixture did not stage the failure", out)
	}

	if got := realPath(t, ResolveRoot(sub)); got != wantInner {
		t.Errorf("ResolveRoot(%q) with git absent = %q, want the .git-marker root %q", sub, got, wantInner)
	}
	plain := mustDir(t, filepath.Join(outer, "plain"))
	if got := ResolveRoot(plain); got != plain {
		t.Errorf("ResolveRoot(%q) with git absent = %q, want cwd (a non-repo directory must not walk)", plain, got)
	}
}

// plantMarker creates a `.git`-NAMED entry of the given shape at dir — the
// three shapes an unprivileged process can create anywhere it can write, none
// of which is a repository: an empty file (`: > .git`), an empty directory
// (`mkdir .git`), and a dangling symlink.
func plantMarker(t *testing.T, dir, shape string) {
	t.Helper()
	marker := filepath.Join(dir, ".git")
	switch shape {
	case "empty file":
		if err := os.WriteFile(marker, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	case "empty directory":
		mustDir(t, marker)
	case "dangling symlink":
		if err := os.Symlink(filepath.Join(dir, "nowhere"), marker); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("unknown marker shape %q", shape)
	}
}

// TestResolveRootRejectsAnImplausibleGitMarker is the bound on the git-refused
// fallback itself. gitutil.RepoShapedRoot is a NAME check — any `.git`-named
// entry, walked to the filesystem root — so on its own it hands the resolution
// to whoever can write a name into a shared ancestor: `: > /tmp/.git` beside a
// planted /tmp/.abcd would govern the rules and the guard of every session
// whose cwd is a plain directory under /tmp, on any host where git cannot
// answer for that directory (which it cannot, because it is not a repository).
//
// The fallback therefore accepts a marker only when it is a PLAUSIBLE
// repository — a .git directory carrying HEAD, or a .git file whose content
// begins "gitdir: " — which is what git itself reads before it will call a
// directory a repository. Anything else is the non-repo case: cwd, no walk,
// nothing above it consulted. This does not make the fallback a trust boundary
// against a writer who takes the trouble to lay out a real repository
// (iss-2609020259564193); it stops the one-command plant, and it stops a stray
// leftover `.git` from silently governing a session.
func TestResolveRootRejectsAnImplausibleGitMarker(t *testing.T) {
	for _, shape := range []string{"empty file", "empty directory", "dangling symlink"} {
		t.Run(shape, func(t *testing.T) {
			outer := mustDir(t, t.TempDir())
			mustDir(t, filepath.Join(outer, ".abcd"))
			plantMarker(t, outer, shape)
			plain := mustDir(t, filepath.Join(outer, "work"))

			if got, err := gitutil.Run(plain, "rev-parse", "--show-toplevel"); err == nil {
				t.Fatalf("git answered %q for a planted marker; the fixture did not stage the git-refused path", got)
			}
			if got := ResolveRoot(plain); got != plain {
				t.Errorf("ResolveRoot(%q) with a %s .git planted above = %q; a plant that is not a repository must not govern the session", plain, shape, got)
			}
		})
	}
}

// TestResolveRootAcceptsAPlausibleGitMarker pins the other side: the plausible
// shapes a real checkout has must still bound the resolution when git will not
// answer, or the fix would close the plant by reopening GHSA-vvqc-3mv2-5p49.
// A .git directory carrying HEAD is the ordinary checkout; a .git FILE reading
// "gitdir: …" is a linked worktree or a submodule, whose working-tree root is
// exactly the directory that file sits in.
func TestResolveRootAcceptsAPlausibleGitMarker(t *testing.T) {
	t.Run("directory with HEAD", func(t *testing.T) {
		outer := mustDir(t, t.TempDir())
		mustDir(t, filepath.Join(outer, ".abcd"))
		repo := mustDir(t, filepath.Join(outer, "checkout"))
		mustDir(t, filepath.Join(repo, ".git"))
		if err := os.WriteFile(filepath.Join(repo, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		sub := mustDir(t, filepath.Join(repo, "internal", "deep"))

		if got, want := realPath(t, ResolveRoot(sub)), realPath(t, repo); got != want {
			t.Errorf("ResolveRoot(%q) = %q, want the checkout root %q", sub, got, want)
		}
	})
	t.Run("gitdir file", func(t *testing.T) {
		outer := mustDir(t, t.TempDir())
		mustDir(t, filepath.Join(outer, ".abcd"))
		wt := mustDir(t, filepath.Join(outer, "linked-worktree"))
		if err := os.WriteFile(filepath.Join(wt, ".git"), []byte("gitdir: "+filepath.Join(outer, "main", ".git", "worktrees", "wt")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		sub := mustDir(t, filepath.Join(wt, "pkg"))

		if got, want := realPath(t, ResolveRoot(sub)), realPath(t, wt); got != want {
			t.Errorf("ResolveRoot(%q) = %q, want the linked worktree's root %q", sub, got, want)
		}
	})
}
