package banlist

import (
	"errors"
	"strings"

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
// <primary>/.git, and in a standalone checkout the two are the same path.
//
// WHICH working tree is the primary one is then ASKED OF GIT, never computed from
// the common dir's path. The arithmetic ("the common dir's parent") is wrong
// wherever the git dir is not inside its own working tree — a bare repository, or
// one made with `--separate-git-dir` — because there the parent is merely the
// directory that happens to HOLD the git dir. Guarding that with a `.git` existence
// test does not close it: clone bare into a directory that sits inside somebody
// else's checkout (or point `--separate-git-dir` there) and the parent is a real
// working tree carrying a real `.git`, so the test passes and this resolver hands
// back an UNRELATED repository — whose private store the caller then reads and the
// committed guard then enforces. That is a match/no-match oracle over another
// repository's private patterns, disclosure of its keys, and a cross-repo denial of
// service from one malformed line over there.
//
// So: `git worktree list --porcelain` names the main working tree in its first
// record, and the candidate is then required to CONFIRM the relationship from its
// own side, with git rediscovering the repository from that directory
// (gitutil.Run scrubs GIT_DIR and friends, so discovery is not short-circuited by
// an inherited environment):
//
//   - its `--show-toplevel` must be the candidate itself, which no directory that
//     merely holds a git dir can satisfy — git answers "must be run in a work tree"
//     for a bare repo and for a `--separate-git-dir` git dir alike; and
//   - its `--git-common-dir` must be OUR common dir, which is what makes the
//     answer unspoofable by layout. A neighbour checkout discovers its own
//     `.git`, never the mirror planted inside it, so the two common dirs differ and
//     the candidate is refused.
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
	list, err := gitutil.Run(repoRoot, "worktree", "list", "--porcelain")
	if err != nil {
		return "", false
	}
	primary := mainWorktreePath(list)
	// A primary that resolves back to THIS working tree is not a second store to
	// inherit.
	if primary == "" || primary == repoRoot {
		return "", false
	}
	// The candidate's own answers, from the candidate's own directory.
	top, err := gitutil.Run(primary, "rev-parse", "--path-format=absolute", "--show-toplevel")
	if err != nil || top != primary {
		return "", false
	}
	common, err := gitutil.Run(primary, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil || common != commonDir {
		return "", false
	}
	return primary, true
}

// mainWorktreePath reads the working-tree path out of the FIRST record of
// `git worktree list --porcelain`, which git documents as the main working tree.
// Anything else — no output, a first line that is not a `worktree ` record — yields
// "", and the caller treats that as "git cannot answer", which is the fallback's
// designed failure mode.
//
// A `bare` marker on the second line needs no special case: the caller's
// `--show-toplevel` confirmation refuses a bare repository on its own, and the one
// authority on whether a directory is a working tree should be git, not a marker
// this function reinterprets.
func mainWorktreePath(porcelain string) string {
	first, _, _ := strings.Cut(porcelain, "\n")
	path, ok := strings.CutPrefix(strings.TrimSuffix(first, "\r"), "worktree ")
	if !ok {
		return ""
	}
	return path
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
