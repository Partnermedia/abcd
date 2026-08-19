package capture

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/Partnermedia/abcd/internal/fsutil"
)

const lockFilename = ".iss-alloc.lock"

// placeholderRetryBudget bounds the O_EXCL bump-retry loop.
const placeholderRetryBudget = 8

// orphanAgeThreshold is how old a zero-byte placeholder must be before the
// sweep removes it.
const orphanAgeThreshold = 60 * time.Second

// lockTimeout is the default flock acquisition budget. A var (not const) so a
// test can shorten it to exercise contention without a multi-second wait.
var lockTimeout = 5 * time.Second

// beforeOrphanRemoveHook, when non-nil, fires immediately before the sweep
// unlinks a classified orphan placeholder. It is a test-only seam (nil in
// production, zero overhead) used to force the iss-102 commit-in-the-unlink-window
// interleaving deterministically.
var beforeOrphanRemoveHook func(cand string)

var rePlaceholderName = regexp.MustCompile(`^iss-[0-9]+(-[a-z0-9]+(-[a-z0-9]+)*)?\.md$`)
var reMaxIssN = regexp.MustCompile(`^iss-([0-9]+)(?:-[a-z0-9-]+)?\.md$`)

// ensureLedgerDirs provisions issuesRoot and the three status sub-directories,
// refusing symlinked leaves. Idempotent.
func ensureLedgerDirs(issuesRoot string) error {
	if !filepath.IsAbs(issuesRoot) {
		return fmt.Errorf("issuesRoot must be absolute, got %q", issuesRoot)
	}
	if err := os.MkdirAll(filepath.Dir(issuesRoot), 0o755); err != nil {
		return err
	}
	if err := safeMkdirLeaf(issuesRoot); err != nil {
		return err
	}
	for _, sub := range []string{"open", "resolved", "wontfix"} {
		if err := safeMkdirLeaf(filepath.Join(issuesRoot, sub)); err != nil {
			return err
		}
	}
	return nil
}

// safeMkdirLeaf creates target if absent, then insists (via Lstat) that the
// result is a real directory and not a symlink.
func safeMkdirLeaf(target string) error {
	fi, err := os.Lstat(target)
	if os.IsNotExist(err) {
		if mkErr := os.Mkdir(target, 0o755); mkErr != nil && !os.IsExist(mkErr) {
			return fmt.Errorf("%w: mkdir failed for %s: %v", ErrPathUnsafe, target, mkErr)
		}
		fi, err = os.Lstat(target)
		if err != nil {
			return fmt.Errorf("%w: leaf disappeared after mkdir: %s", ErrPathUnsafe, target)
		}
	} else if err != nil {
		return fmt.Errorf("%w: lstat failed for %s: %v", ErrPathUnsafe, target, err)
	}
	if fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir() {
		return fmt.Errorf("%w: not a real directory: %s", ErrPathUnsafe, target)
	}
	return nil
}

// withLedgerLock runs fn while holding the exclusive allocator flock, so every
// ledger mutation — id allocation, status transitions, AND the capture commit
// write — serializes on one lock. It creates the ledger dirs, then routes through
// the shared fsutil.WithFileLock primitive (the one-canonical inter-process
// load-modify-write lock), mapping its sentinels back to the capture-facing
// ErrAllocatorContention / ErrPathUnsafe the ledger callers already test.
func withLedgerLock(issuesRoot string, fn func() error) error {
	if err := ensureLedgerDirs(issuesRoot); err != nil {
		return err
	}
	lockPath := filepath.Join(issuesRoot, lockFilename)
	err := fsutil.WithFileLock(lockPath, lockTimeout, fn)
	switch {
	case errors.Is(err, fsutil.ErrLockContention):
		return fmt.Errorf("%w: could not acquire allocator lock within %s", ErrAllocatorContention, lockTimeout)
	case errors.Is(err, fsutil.ErrLockPathUnsafe):
		return fmt.Errorf("%w: allocator lock path is unsafe: %s", ErrPathUnsafe, lockPath)
	}
	return err
}

// reservePath reserves an iss-N id and creates a zero-byte placeholder under
// open/, mirroring reserve_issue_path: flock -> scan max N -> O_EXCL create
// with bump-retry. When forceID is non-empty it demands that exact id.
//
// refFloor is the highest iss-N observed across other git refs
// (recordid.MaxAcrossRefs), computed by the caller before the lock: the reserved
// id starts at max(local max N, refFloor)+1, so a branch that has already
// committed a higher id is not re-minted (iss-115, iss-120). It is a floor, not a
// substitute for the local scan — the local scan alone sees uncommitted mints in
// this worktree, and the O_EXCL bump-retry still resolves any residual clash.
func reservePath(issuesRoot, slug, forceID string, refFloor int) (string, string, error) {
	// Validate a caller-supplied ForceID against the iss-N shape BEFORE it is used
	// to build a path or create a placeholder — a traversal id (../../evil) must
	// never touch the filesystem outside the ledger, even transiently.
	if forceID != "" && !reIssID.MatchString(forceID) {
		return "", "", fmt.Errorf("%w: ForceID %q must match ^iss-[0-9]+$", ErrPathUnsafe, forceID)
	}
	var resID, resTarget string
	err := withLedgerLock(issuesRoot, func() error {
		if forceID != "" {
			if issPresent(issuesRoot, forceID) {
				return fmt.Errorf("%w: %s already exists in the ledger", ErrDuplicateIssueID, forceID)
			}
			target := filepath.Join(issuesRoot, "open", forceID+"-"+slug+".md")
			fd, cErr := createPlaceholder(target)
			if cErr != nil {
				if os.IsExist(cErr) {
					return fmt.Errorf("%w: %s appeared between scan and create", ErrDuplicateIssueID, forceID)
				}
				return cErr
			}
			syscall.Close(fd)
			resID, resTarget = forceID, target
			return nil
		}

		maxN := maxIssN(issuesRoot)
		if refFloor > maxN {
			maxN = refFloor
		}
		// Guard the id arithmetic below (maxN+1+attempt) against int overflow: a
		// hand-crafted MaxInt-adjacent filename would otherwise wrap to a negative
		// "iss--N" that fails reIssID, creating a bogus placeholder that only fails
		// downstream. Refuse clearly instead.
		if maxN > math.MaxInt-placeholderRetryBudget {
			return fmt.Errorf("%w: iss-N counter near the integer ceiling (highest observed %d); refusing to allocate", ErrAllocatorContention, maxN)
		}
		for attempt := 0; attempt < placeholderRetryBudget; attempt++ {
			issID := fmt.Sprintf("iss-%d", maxN+1+attempt)
			target := filepath.Join(issuesRoot, "open", issID+"-"+slug+".md")
			fd, cErr := createPlaceholder(target)
			if cErr != nil {
				if os.IsExist(cErr) {
					continue
				}
				return cErr
			}
			syscall.Close(fd)
			resID, resTarget = issID, target
			return nil
		}
		return fmt.Errorf("%w: could not allocate iss-N after %d retries", ErrAllocatorContention, placeholderRetryBudget)
	})
	if err != nil {
		return "", "", err
	}
	return resID, resTarget, nil
}

// createPlaceholder does an O_EXCL|O_NOFOLLOW create of the placeholder file.
func createPlaceholder(target string) (int, error) {
	fd, err := syscall.Open(target, syscall.O_CREAT|syscall.O_EXCL|syscall.O_WRONLY|syscall.O_NOFOLLOW, 0o644)
	if err != nil {
		if err == syscall.ELOOP {
			return -1, fmt.Errorf("%w: placeholder path is a symlink: %s", ErrPathUnsafe, target)
		}
		if err == syscall.EEXIST {
			if fi, lerr := os.Lstat(target); lerr == nil && fi.Mode()&os.ModeSymlink != 0 {
				return -1, fmt.Errorf("%w: placeholder path is a symlink: %s", ErrPathUnsafe, target)
			}
			return -1, os.ErrExist
		}
		return -1, err
	}
	return fd, nil
}

// maxIssN scans all three status dirs for the highest iss-N (0 if none).
func maxIssN(issuesRoot string) int {
	maxN := 0
	for _, sub := range []string{"open", "resolved", "wontfix"} {
		entries, err := os.ReadDir(filepath.Join(issuesRoot, sub))
		if err != nil {
			continue
		}
		for _, e := range entries {
			m := reMaxIssN.FindStringSubmatch(e.Name())
			if m == nil {
				continue
			}
			// An over-int digit run (or otherwise unparseable N) is not a usable
			// maximum — skip it rather than silently folding it to 0, which would
			// mask a filename that should have driven allocation higher.
			n, err := strconv.Atoi(m[1])
			if err != nil {
				continue
			}
			if n > maxN {
				maxN = n
			}
		}
	}
	return maxN
}

// issPresent reports whether issID exists in any status dir.
func issPresent(issuesRoot, issID string) bool {
	prefix := issID + "-"
	exact := issID + ".md"
	for _, sub := range []string{"open", "resolved", "wontfix"} {
		entries, err := os.ReadDir(filepath.Join(issuesRoot, sub))
		if err != nil {
			continue
		}
		for _, e := range entries {
			n := e.Name()
			if n == exact || (len(n) > len(prefix) && n[:len(prefix)] == prefix && filepath.Ext(n) == ".md") {
				return true
			}
		}
	}
	return false
}

// cancelReservation removes a zero-byte placeholder idempotently. It refuses a
// symlinked or non-empty target (real content is the caller's transactional
// responsibility).
func cancelReservation(path string) error {
	fi, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%w: refusing to cancel a symlinked placeholder: %s", ErrPathUnsafe, path)
	}
	if !fi.Mode().IsRegular() {
		return fmt.Errorf("%w: placeholder is not a regular file: %s", ErrPathUnsafe, path)
	}
	if fi.Size() != 0 {
		return fmt.Errorf("refusing to cancel non-empty placeholder (%d bytes): %s", fi.Size(), path)
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// cleanOrphanPlaceholders sweeps zero-byte iss-N placeholders older than the
// threshold from open/. Tolerates a virgin ledger. Refuses symlinked roots.
func cleanOrphanPlaceholders(issuesRoot string) error {
	fi, err := os.Lstat(issuesRoot)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%w: issuesRoot is a symlink: %s", ErrPathUnsafe, issuesRoot)
	}
	openDir := filepath.Join(issuesRoot, "open")
	ofi, err := os.Lstat(openDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if ofi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%w: issuesRoot/open is a symlink: %s", ErrPathUnsafe, openDir)
	}
	entries, err := os.ReadDir(openDir)
	if err != nil {
		return nil
	}
	now := time.Now()
	for _, e := range entries {
		if !rePlaceholderName.MatchString(e.Name()) {
			continue
		}
		cand := filepath.Join(openDir, e.Name())
		cfi, err := os.Lstat(cand)
		if err != nil {
			continue
		}
		if cfi.Mode()&os.ModeSymlink != 0 || !cfi.Mode().IsRegular() {
			continue
		}
		if cfi.Size() != 0 {
			continue
		}
		if now.Sub(cfi.ModTime()) <= orphanAgeThreshold {
			continue
		}
		if !orphanStillRemovable(cand, cfi) {
			continue
		}
		// Test-only seam (nil in production): fires in the residual window between
		// the pre-unlink re-check and the unlink, to force the documented iss-102
		// interleaving where a stalled commit's fill lands here.
		if beforeOrphanRemoveHook != nil {
			beforeOrphanRemoveHook(cand)
		}
		os.Remove(cand)
	}
	return nil
}

// orphanStillRemovable re-verifies, in the tightest possible window immediately
// before the unlink, that cand is still the same zero-byte inode the sweep
// classified as an orphan. A capture commits by atomically renaming a full issue
// file over its reserved placeholder, and that write happens OUTSIDE this
// (lockless) sweep — so between the caller's Lstat and its os.Remove a stalled
// capture can land its committed file at this exact path. Re-checking here means
// the sweep never deletes a placeholder a commit has since replaced or filled,
// closing the race for any commit that lands before this check. The residual
// micro-window between this stat and the unlink can only be fully eliminated by
// serialising the commit write on the ledger lock (see commitCapture in
// workflow.go), which is outside this sweep's reach.
func orphanStillRemovable(cand string, seen os.FileInfo) bool {
	recheck, err := os.Lstat(cand)
	if err != nil {
		return false
	}
	if recheck.Mode()&os.ModeSymlink != 0 || !recheck.Mode().IsRegular() {
		return false
	}
	if recheck.Size() != 0 || !os.SameFile(seen, recheck) {
		return false
	}
	return true
}

// findIssue locates issID across the three status dirs, mirroring find_issue.
func findIssue(issuesRoot, issID string) (string, State, error) {
	if !reIssID.MatchString(issID) {
		return "", "", fmt.Errorf("invalid iss-N identifier: %q", issID)
	}
	prefix := issID + "-"
	exact := issID + ".md"
	type match struct {
		path   string
		status State
	}
	var matches []match
	for _, sub := range statusDirs {
		dir := filepath.Join(issuesRoot, statusDirName[sub])
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			n := e.Name()
			if filepath.Ext(n) != ".md" {
				continue
			}
			if n == exact {
				matches = append(matches, match{filepath.Join(dir, n), sub})
				continue
			}
			m := reFilenameID.FindStringSubmatch(n)
			if len(n) > len(prefix) && n[:len(prefix)] == prefix && m != nil && m[1] == issID {
				matches = append(matches, match{filepath.Join(dir, n), sub})
			}
		}
	}
	if len(matches) == 0 {
		return "", "", fmt.Errorf("%w: %s not found in any status directory", ErrUnknownIssueID, issID)
	}
	if len(matches) > 1 {
		return "", "", fmt.Errorf("%w: %s present in multiple files", ErrDuplicateIssueID, issID)
	}
	return matches[0].path, matches[0].status, nil
}
