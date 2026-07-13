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
