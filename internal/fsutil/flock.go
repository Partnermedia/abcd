package fsutil

import (
	"errors"
	"fmt"
	"syscall"
	"time"
)

// ErrLockContention and ErrLockPathUnsafe are the file-lock sentinels: the
// exclusive lock could not be acquired within the caller's timeout, and the lock
// path is a symlink or not a regular file.
var (
	ErrLockContention = errors.New("fsutil: lock contention")
	ErrLockPathUnsafe = errors.New("fsutil: lock path unsafe")
)

// WithFileLock acquires an exclusive advisory (flock) lock on lockPath, holds it
// across fn, and releases it when fn returns. The lock file is opened O_NOFOLLOW
// (a symlinked lock path is refused) and verified on the same descriptor to be a
// regular file, both reported as ErrLockPathUnsafe; the exclusive flock is polled
// for up to timeout, after which ErrLockContention is returned. fn's own error is
// returned unwrapped.
//
// It is the single inter-process load-modify-write lock primitive: capture's
// ledger allocator and ahoy's ~/.abcd history registry both route through it
// rather than keep divergent flock loops (the one-canonical-primitive invariant).
// The lock is advisory — every writer of the guarded state must take it — and its
// scope is one lock file; callers that must not deadlock keep their acquisitions
// unnested.
func WithFileLock(lockPath string, timeout time.Duration, fn func() error) error {
	fd, err := openLockFd(lockPath)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)

	if err := acquireFlock(fd, timeout); err != nil {
		return err
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)

	return fn()
}

// openLockFd opens lockPath with O_CREAT|O_RDWR|O_NOFOLLOW and verifies, on the
// same descriptor, that it is a regular file — refusing a symlinked or
// non-regular lock path with ErrLockPathUnsafe.
func openLockFd(lockPath string) (int, error) {
	fd, err := syscall.Open(lockPath, syscall.O_CREAT|syscall.O_RDWR|syscall.O_NOFOLLOW, 0o644)
	if err != nil {
		if err == syscall.ELOOP {
			return -1, fmt.Errorf("%w: lock path is a symlink: %s", ErrLockPathUnsafe, lockPath)
		}
		return -1, err
	}
	var st syscall.Stat_t
	if err := syscall.Fstat(fd, &st); err != nil {
		syscall.Close(fd)
		return -1, err
	}
	if st.Mode&syscall.S_IFMT != syscall.S_IFREG {
		syscall.Close(fd)
		return -1, fmt.Errorf("%w: lock path is not a regular file: %s", ErrLockPathUnsafe, lockPath)
	}
	return fd, nil
}

// acquireFlock polls for an exclusive flock until timeout elapses, returning
// ErrLockContention on timeout.
func acquireFlock(fd int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	backoff := 5 * time.Millisecond
	for {
		err := syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return nil
		}
		if err != syscall.EWOULDBLOCK && err != syscall.EAGAIN {
			return err
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return fmt.Errorf("%w: could not acquire lock within %s", ErrLockContention, timeout)
		}
		if backoff > remaining {
			backoff = remaining
		}
		time.Sleep(backoff)
		if backoff < 100*time.Millisecond {
			backoff *= 2
		}
	}
}
