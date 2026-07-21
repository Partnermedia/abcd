package changelog

import (
	"testing"

	"github.com/REPPL/abcd-cli/internal/gittest"
)

// fixtureRepo is this package's name for the shared hermetic git fixture,
// gittest.Repo. The set-difference and the tag anchor are only trustworthy if
// they are exercised against real git objects — a stubbed git would prove
// nothing about the squash-merge caveat this phase must pin — so the tests build
// actual histories rather than faking the plumbing.
//
// It delegates rather than duplicating: the same fixture is needed by the
// guardrail's front-door test and by the ship verb's, and three private copies
// of a git fixture would drift apart. The lower-case methods keep every existing
// call site in this package unchanged.
type fixtureRepo struct {
	*gittest.Repo
	t    *testing.T
	root string
}

func newFixtureRepo(t *testing.T) *fixtureRepo {
	t.Helper()
	r := gittest.NewRepo(t)
	return &fixtureRepo{Repo: r, t: t, root: r.Root()}
}

func (r *fixtureRepo) git(args ...string) string     { return r.Git(args...) }
func (r *fixtureRepo) write(rel, content string)     { r.Write(rel, content) }
func (r *fixtureRepo) remove(rel string)             { r.Remove(rel) }
func (r *fixtureRepo) commit(msg string)             { r.Commit(msg) }
func (r *fixtureRepo) record(rel, id, impact string) { r.Record(rel, id, impact) }
