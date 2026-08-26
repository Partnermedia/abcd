package repolint

import (
	"errors"
	"io/fs"
	"syscall"
	"testing"
)

// The open-failure classification is the seam between "legitimate worktree
// state, skip silently" and "content nobody looked at, warn" (the engine
// contract: a check that cannot run must not be silently reported as
// passing). The polarity table is pinned here so a future edit cannot
// silently widen either side.
func TestOpenFailureWarrantsWarnPolarity(t *testing.T) {
	pe := func(errno syscall.Errno) error {
		return &fs.PathError{Op: "openat", Path: "x", Err: errno}
	}
	warns := []error{pe(syscall.EACCES), pe(syscall.EPERM), pe(syscall.EIO)}
	for _, err := range warns {
		if !openFailureWarrantsWarn(err) {
			t.Errorf("%v must warn: the file exists and was not scanned", err)
		}
	}
	silents := []error{
		nil,
		pe(syscall.ENOENT),  // deleted in the worktree, sparse checkout
		pe(syscall.ENOTDIR), // parent replaced by a file — path is absent
		pe(syscall.ELOOP),   // tracked symlink leaf under O_NOFOLLOW — skip by design
		errors.New("openat sub: path escapes from parent"), // os.Root containment refusal
	}
	for _, err := range silents {
		if openFailureWarrantsWarn(err) {
			t.Errorf("%v must stay silent: a legitimate worktree state, not unscanned content", err)
		}
	}
}
