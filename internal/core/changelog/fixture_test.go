package changelog

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/gittest"
)

// fixtureRepo is a hermetic throwaway git repository for the derivation tests.
// The set-difference and the tag anchor are only trustworthy if they are
// exercised against real git objects — a stubbed git would prove nothing about
// the squash-merge caveat this phase must pin — so the tests build actual
// histories here rather than faking the plumbing.
type fixtureRepo struct {
	t    *testing.T
	root string
	env  []string
}

// newFixtureRepo initialises an empty repo on a fixed branch name, with an
// identity pinned per command, under the shared hermetic git environment
// (gittest.Env) so an ambient GIT_DIR cannot redirect it at the real tree.
func newFixtureRepo(t *testing.T) *fixtureRepo {
	t.Helper()
	r := &fixtureRepo{t: t, root: t.TempDir(), env: gittest.Env(t)}
	init := exec.Command("git", "-C", r.root, "init", "--initial-branch=main")
	init.Env = r.env
	if out, err := init.CombinedOutput(); err != nil {
		t.Skipf("git init unavailable: %v (%s)", err, out)
	}
	return r
}

// git runs one git command in the fixture, failing the test on a non-zero exit.
func (r *fixtureRepo) git(args ...string) string {
	r.t.Helper()
	full := append([]string{
		"-C", r.root,
		"-c", "user.email=fixture@example.invalid",
		"-c", "user.name=Fixture",
		"-c", "commit.gpgsign=false",
	}, args...)
	cmd := exec.Command("git", full...)
	cmd.Env = r.env
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// write creates (or overwrites) a repo-relative file, creating parents.
func (r *fixtureRepo) write(rel, content string) {
	r.t.Helper()
	path := filepath.Join(r.root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

// remove deletes a repo-relative file from the working tree.
func (r *fixtureRepo) remove(rel string) {
	r.t.Helper()
	if err := os.Remove(filepath.Join(r.root, filepath.FromSlash(rel))); err != nil {
		r.t.Fatal(err)
	}
}

// commit stages everything and records a commit.
func (r *fixtureRepo) commit(msg string) {
	r.t.Helper()
	r.git("add", "-A")
	r.git("commit", "--allow-empty", "-m", msg)
}

// record writes a minimal record file with the given impact frontmatter.
func (r *fixtureRepo) record(rel, id, impact string) {
	r.t.Helper()
	r.write(rel, "---\nid: "+id+"\nimpact: "+impact+"\n---\n# "+id+"\n")
}
