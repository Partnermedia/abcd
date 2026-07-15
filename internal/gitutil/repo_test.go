package gitutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/REPPL/abcd-cli/internal/gitutil"
)

// commitAll stages and commits everything in repo, with an isolated identity so
// no developer git config is needed.
func commitAll(t *testing.T, repo string) {
	t.Helper()
	for _, args := range [][]string{
		{"add", "-A"},
		{"-c", "user.email=t@example.com", "-c", "user.name=t", "commit", "-q", "-m", "fixture"},
	} {
		if out, err := runGit(t, repo, args...); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
}

func TestInRepo(t *testing.T) {
	repo := newRepo(t, "")
	if !gitutil.InRepo(repo) {
		t.Error("InRepo(git repo) = false, want true")
	}

	plain := t.TempDir() // not a git repo
	if gitutil.InRepo(plain) {
		t.Error("InRepo(non-repo) = true, want false")
	}
}

func TestTrackedFiles(t *testing.T) {
	repo := newRepo(t, "")
	// Two committed files, one untracked. TrackedFiles reports only committed.
	write := func(rel, body string) {
		p := filepath.Join(repo, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("a.md", "x")
	write("sub/b.go", "y")
	commitAll(t, repo)
	write("untracked.txt", "z") // added after the commit, never staged

	got, err := gitutil.TrackedFiles(repo)
	if err != nil {
		t.Fatal(err)
	}
	set := map[string]bool{}
	for _, f := range got {
		set[f] = true
	}
	if !set["a.md"] || !set["sub/b.go"] {
		t.Errorf("TrackedFiles = %v, want a.md and sub/b.go", got)
	}
	if set["untracked.txt"] {
		t.Errorf("TrackedFiles included an untracked file: %v", got)
	}
}

// Outside a repo, TrackedFiles returns no files and no error (fail open) — a
// privacy scan degrades to "nothing to scan", not a crash.
func TestTrackedFilesOutsideRepo(t *testing.T) {
	got, err := gitutil.TrackedFiles(t.TempDir())
	if err != nil {
		t.Fatalf("TrackedFiles outside a repo errored: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("TrackedFiles outside a repo = %v, want empty", got)
	}
}

// Inside a repo, a failure other than not-a-repo (here a corrupt index) must be
// returned as an error, not swallowed as "nothing tracked". Otherwise a
// scanning rule (privacy hygiene) would read zero files and report the repo
// clean. InRepo still says true — it does not read the index — so the earlier
// blanket (nil, nil) hid a real read failure.
func TestTrackedFilesCorruptIndexErrors(t *testing.T) {
	repo := newRepo(t, "")
	if err := os.WriteFile(filepath.Join(repo, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	commitAll(t, repo)
	if err := os.WriteFile(filepath.Join(repo, ".git", "index"), []byte("garbage"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !gitutil.InRepo(repo) {
		t.Skip("InRepo unexpectedly false with a corrupt index")
	}
	if _, err := gitutil.TrackedFiles(repo); err == nil {
		t.Error("TrackedFiles with a corrupt index returned nil error; want the failure surfaced")
	}
}

// An inherited GIT_DIR/GIT_WORK_TREE must not redirect queries to a different
// repository: isolation strips them so the answer is always about root, not the
// repo those vars point at.
func TestIsolationIgnoresInheritedGitDir(t *testing.T) {
	writeFile := func(dir, rel, body string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	target := newRepo(t, "")
	writeFile(target, "target.md", "x")
	commitAll(t, target)

	other := newRepo(t, "")
	writeFile(other, "secret.md", "y")
	commitAll(t, other)

	// Point the environment at the OTHER repo, then ask about target.
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)

	got, err := gitutil.TrackedFiles(target)
	if err != nil {
		t.Fatalf("TrackedFiles: %v", err)
	}
	set := map[string]bool{}
	for _, f := range got {
		set[f] = true
	}
	if set["secret.md"] {
		t.Errorf("TrackedFiles followed inherited GIT_DIR into another repo: %v", got)
	}
	if !set["target.md"] {
		t.Errorf("TrackedFiles = %v, want target.md", got)
	}

	// A plain non-repo directory must not read as a repo just because GIT_DIR
	// is set in the environment.
	if gitutil.InRepo(t.TempDir()) {
		t.Error("InRepo(non-repo with inherited GIT_DIR) = true, want false")
	}
}
