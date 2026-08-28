package banlist

import (
	"errors"
	"path/filepath"

	"github.com/intentdriven/abcd/internal/gitutil"
)

// InheritedReport is the private layer a LINKED git worktree inherits from its
// primary checkout: which checkout it came from, and that checkout's private layer
// exactly as ListPrivate reads it there.
//
// It exists because the local-ephemeral tier is PER-WORKTREE and gitignored, so a
// checkout made with `git worktree add` starts with no private store at all. The
// committed guard resolves the primary checkout's store and enforces it in the
// worktree (itd-150); a status surface that did not report the same layer would
// call a worktree unprotected while every commit made in it is being checked, which
// is the one thing a board about a guard must never do.
type InheritedReport struct {
	// PrimaryRoot is the primary working tree's absolute root — the remedy's
	// location. An inherited entry is not in the store this checkout carries, so its
	// key alone reads as a phantom without it.
	PrimaryRoot string `json:"primary_root"`
	// Private is the primary checkout's private layer, read there.
	Private PrivateReport `json:"private"`
}

// PrimaryWorktreeRoot reports the PRIMARY checkout's working-tree root when
// repoRoot is inside a LINKED git worktree, and ok=false everywhere else — a
// standalone checkout, a directory that is not a repository, or a git that cannot
// answer.
//
// git tells the two shapes apart: in a linked worktree `--git-dir` resolves to
// <primary>/.git/worktrees/<name> while `--git-common-dir` resolves to
// <primary>/.git, and in a standalone checkout the two are the same path. The
// primary working tree is the common dir's parent.
//
// It is the Go half of the resolution the committed pre-commit guard makes, and it
// is deliberately the same three-way answer: resolution failure is ok=false, never
// an error. A read surface that refused to render because it could not find a
// SECOND store would be less useful than one that renders the first.
func PrimaryWorktreeRoot(repoRoot string) (string, bool) {
	gitDir, err := gitutil.Run(repoRoot, "rev-parse", "--path-format=absolute", "--git-dir")
	if err != nil || gitDir == "" {
		return "", false
	}
	commonDir, err := gitutil.Run(repoRoot, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil || commonDir == "" || gitDir == commonDir {
		return "", false
	}
	primary := filepath.Dir(commonDir)
	// A common dir with no parent left to take names no directory, and a primary
	// that resolves back to THIS working tree is not a second store to inherit.
	if primary == "" || primary == commonDir || primary == repoRoot {
		return "", false
	}
	return primary, true
}

// InheritedPrivate reports the private layer repoRoot inherits from its primary
// checkout, or nil when there is nothing to inherit — repoRoot is not a linked
// worktree, or the primary checkout has no store of its own.
//
// An unreadable or malformed primary store IS an error: the guard refuses every
// commit in that state, so a surface that quietly rendered "nothing inherited"
// would report a repo as healthy while nothing in it can be committed.
func InheritedPrivate(repoRoot string) (*InheritedReport, error) {
	primary, ok := PrimaryWorktreeRoot(repoRoot)
	if !ok {
		return nil, nil
	}
	rep, err := ListPrivate(primary)
	if err != nil {
		if errors.Is(err, ErrNoStore) {
			return nil, nil
		}
		return nil, err
	}
	if !rep.Present {
		// The primary checkout has not opted in either, so there is no second layer to
		// report — and reporting an absent one would put a second "INACTIVE" line on
		// every status render in every worktree, which teaches a reader to skip them.
		return nil, nil
	}
	return &InheritedReport{PrimaryRoot: primary, Private: rep}, nil
}
