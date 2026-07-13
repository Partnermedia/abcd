package gitutil_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/REPPL/abcd-cli/internal/gitutil"
)

// newRepo makes an isolated git repo with the given .gitignore body. Global and
// system config are neutralised so a developer's ~/.gitignore cannot change the
// result.
func newRepo(t *testing.T, gitignore string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	if gitignore != "" {
		if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignore), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestIsIgnored(t *testing.T) {
	repo := newRepo(t, ".abcd/.work.local/\n*.log\n")

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"ignored directory", ".abcd/.work.local/", true},
		{"ignored by glob", "debug.log", true},
		{"tracked path", ".abcd/work/DECISIONS.md", false},
		{"tracked source", "internal/core/audit.go", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gitutil.IsIgnored(repo, tt.path); got != tt.want {
				t.Errorf("IsIgnored(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// A negation pattern un-ignores a path: `!keep.log` must NOT read as ignored,
// even though `*.log` matches it. git reports the match with a leading `!`.
func TestIsIgnoredNegationIsNotIgnored(t *testing.T) {
	repo := newRepo(t, "*.log\n!keep.log\n")

	if gitutil.IsIgnored(repo, "keep.log") {
		t.Error("IsIgnored(keep.log) = true; a negation pattern must not count as ignored")
	}
	if !gitutil.IsIgnored(repo, "other.log") {
		t.Error("IsIgnored(other.log) = false, want true")
	}
}

func TestCheckIgnoredBatch(t *testing.T) {
	repo := newRepo(t, "ignored/\n")

	got := gitutil.CheckIgnored(repo, []string{"ignored/a.txt", "kept/b.txt"})
	if _, ok := got["ignored/a.txt"]; !ok {
		t.Error("CheckIgnored: ignored/a.txt missing from the ignored set")
	}
	if _, ok := got["kept/b.txt"]; ok {
		t.Error("CheckIgnored: kept/b.txt must not be in the ignored set")
	}
}

// Fails open: outside a git repo nothing is reported ignored, and no error is
// raised. An audit rule must not claim "not gitignored" is a violation just
// because git is absent.
func TestCheckIgnoredOutsideRepoFailsOpen(t *testing.T) {
	dir := t.TempDir() // not a git repo

	got := gitutil.CheckIgnored(dir, []string{"anything.txt"})
	if len(got) != 0 {
		t.Errorf("CheckIgnored outside a repo = %v, want empty", got)
	}
}

func TestCheckIgnoredEmptyCandidates(t *testing.T) {
	repo := newRepo(t, "*.log\n")

	got := gitutil.CheckIgnored(repo, nil)
	if len(got) != 0 {
		t.Errorf("CheckIgnored(nil) = %v, want empty", got)
	}
}
