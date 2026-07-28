// Package fsutil holds the durable-write and path-safety primitives shared by
// the ~/.abcd and repo .abcd store writers. It is transport-agnostic: no
// stdout, no os.Exit, no CLI knowledge.
//
// It is the single home for the atomic temp-file+fsync+rename write and the
// "is this a real directory, not a symlink" check: the ahoy, capture, and
// memory store writers all route through WriteFileAtomic /
// WriteFileAtomicPreserveMode / IsRealDir rather than keep divergent copies
// (the one-canonical-primitive invariant, guarded by
// TestNoNonCanonicalAtomicWritePrimitives).
//
// Each read/write primitive comes in two forms. The plain form takes a path and
// guards the LEAF only; the …InRoot form takes an *os.Root and resolves every
// component inside it, which is what contains a path whose ANCESTOR is a symlink
// planted by untrusted content. Any path that arrives as data — a committed
// configuration value, a packed manifest entry — belongs in the …InRoot form.
package fsutil

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"syscall"
)

// ErrNotRegular and ErrTooBig are the guarded-read sentinels: a non-regular leaf
// (symlink/FIFO/device/directory) and a file over the caller's byte cap.
var (
	ErrNotRegular = errors.New("fsutil: not a regular file")
	ErrTooBig     = errors.New("fsutil: file exceeds size cap")
)

// ReadGuarded opens path once, read-only, with O_NOFOLLOW (refuse a symlinked
// leaf) and O_NONBLOCK (a FIFO/device leaf returns immediately instead of
// blocking the open forever), then validates on the SAME descriptor that it is a
// regular file within limit bytes before reading through a LimitReader — so no
// symlink swap between stat and read, no non-regular leaf, and no size overrun
// can reach the caller. It is the shared trust-boundary read primitive for any
// file inside a repo working tree that untrusted content could have replaced
// with a symlink to /dev/zero or an endless device. The raw open error is
// returned so callers can test os.IsNotExist / syscall.ELOOP; a non-regular or
// oversize file returns ErrNotRegular / ErrTooBig.
func ReadGuarded(path string, limit int64) ([]byte, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !fi.Mode().IsRegular() {
		return nil, ErrNotRegular
	}
	if fi.Size() > limit {
		return nil, ErrTooBig
	}
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		// Grew past the cap between fstat and read (a size TOCTOU).
		return nil, ErrTooBig
	}
	return data, nil
}

// ReadGuardedInRoot is ReadGuarded resolved inside an os.Root containment
// scope. rel is a slash-separated path relative to root; every component is
// resolved by the OS within root, so a symlinked ANCESTOR directory — the shape
// a hostile clone commits as git mode 120000, `etclink -> /etc` or
// `home -> ../../..` — cannot walk the read outside the root. ValidRelPath is
// lexical and cannot see a symlink, so containment must rest here, not on it.
//
// It keeps every guarantee ReadGuarded gives at the leaf: a symlinked leaf is
// refused (lstat, never followed), the descriptor is confirmed to be the very
// file that was vetted (os.SameFile closes the lstat→open swap), a non-regular
// leaf returns ErrNotRegular, O_NONBLOCK stops a FIFO or device blocking the
// open, and the caller's byte cap is enforced against both the fstat size and
// the bytes actually read.
//
// Errors are returned raw where the OS raised them, so a caller can still test
// os.IsNotExist; a path that escapes the root is NOT an os.IsNotExist error, so
// a caller skipping absent files fails closed on the escape rather than treating
// it as "the file is not there".
func ReadGuardedInRoot(root *os.Root, rel string, limit int64) ([]byte, error) {
	fi, err := root.Lstat(rel)
	if err != nil {
		return nil, err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return nil, ErrNotRegular
	}
	if !fi.Mode().IsRegular() {
		return nil, ErrNotRegular
	}
	f, err := root.OpenFile(rel, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, ErrNotRegular
	}
	if !os.SameFile(fi, st) {
		// Swapped between the lstat and the open. os.Root already stops the
		// swap escaping the root, but the descriptor is no longer the file that
		// was vetted, so it is refused rather than read.
		return nil, ErrNotRegular
	}
	if st.Size() > limit {
		return nil, ErrTooBig
	}
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		// Grew past the cap between fstat and read (a size TOCTOU).
		return nil, ErrTooBig
	}
	return data, nil
}

// WriteFileAtomic writes data to path durably: a temp file in the target
// directory is written, chmod'd to perm on its open descriptor, flushed,
// fsync'd, then renamed over the target, and finally the parent directory is
// fsync'd best-effort so the rename survives a crash. Parent directories are
// created as needed.
//
// The rename is the commit point: a reader sees either the old file or the
// complete new one, never a half-written file. os.Rename does not follow a
// symlink at the leaf — a pre-planted symlink at path is replaced by the real
// file, not written through. The mode is set with fchmod on the open descriptor
// (never chmod-by-name on the closed temp), so the enumerable temp name cannot
// be swapped for a symlink between close and chmod and have the mode applied to
// an attacker-chosen target.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".abcd-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}
	syncParent(dir)
	return nil
}

// WriteFileAtomicPreserveMode is WriteFileAtomic that keeps the target's
// existing permission bits when it already exists, defaulting to 0644 for a new
// file. It is the canonical form for the store writers that rewrite a file in
// place and must not silently reset its mode.
func WriteFileAtomicPreserveMode(path string, data []byte) error {
	perm := os.FileMode(0o644)
	fi, err := os.Stat(path)
	switch {
	case err == nil:
		perm = fi.Mode().Perm()
	case !notPresent(err):
		// A real stat fault (a transient I/O error, ELOOP, EACCES) is NOT
		// "absent" — defaulting to 0644 here would silently widen an existing
		// restrictive mode, contrary to the contract. Fail closed, like paths.go.
		return err
	}
	return WriteFileAtomic(path, data, perm)
}

// WriteFileAtomicInRoot is WriteFileAtomic resolved inside an os.Root
// containment scope: rel is a slash-separated path relative to root, and the
// missing parents, the temp file, the rename, and the parent fsync all happen
// through root. A symlinked ancestor committed by a hostile repo therefore
// cannot land the write outside the root — the counterpart of
// ReadGuardedInRoot, and the reason a write path validated only lexically is not
// contained.
//
// The commit point and the ordering are WriteFileAtomic's: write, fchmod on the
// open descriptor (never chmod-by-name), fsync, close, rename, parent fsync.
// The temp name is unpredictable so the enumerable-name swap WriteFileAtomic
// guards against has nothing to aim at.
func WriteFileAtomicInRoot(root *os.Root, rel string, data []byte, perm os.FileMode) error {
	dir := path.Dir(rel)
	if dir != "." {
		if err := root.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	tmp, tmpRel, err := createTempInRoot(root, dir)
	if err != nil {
		return err
	}
	abandon := func(err error) error {
		tmp.Close()
		_ = root.Remove(tmpRel)
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		return abandon(err)
	}
	if err := tmp.Chmod(perm); err != nil {
		return abandon(err)
	}
	if err := tmp.Sync(); err != nil {
		return abandon(err)
	}
	if err := tmp.Close(); err != nil {
		_ = root.Remove(tmpRel)
		return err
	}
	if err := root.Rename(tmpRel, rel); err != nil {
		_ = root.Remove(tmpRel)
		return err
	}
	syncDirInRoot(root, dir)
	return nil
}

// WriteFileAtomicPreserveModeInRoot is WriteFileAtomicInRoot that keeps the
// target's existing permission bits when it already exists, defaulting to 0644
// for a new file — the contained counterpart of
// WriteFileAtomicPreserveMode.
func WriteFileAtomicPreserveModeInRoot(root *os.Root, rel string, data []byte) error {
	perm := os.FileMode(0o644)
	fi, err := root.Lstat(rel)
	switch {
	case err == nil:
		if fi.Mode().IsRegular() {
			perm = fi.Mode().Perm()
		}
	case !notPresent(err):
		// A real stat fault (a transient I/O error, an escaping path, EACCES) is
		// NOT "absent" — defaulting to 0644 here would silently widen an existing
		// restrictive mode, and an escaping path must refuse rather than proceed.
		return err
	}
	return WriteFileAtomicInRoot(root, rel, data, perm)
}

// createTempInRoot creates an exclusively-owned temp file in dir inside root,
// returning it and its root-relative path. os.CreateTemp cannot be used here: it
// resolves a path outside any containment scope, which is precisely what the
// root exists to prevent.
func createTempInRoot(root *os.Root, dir string) (*os.File, string, error) {
	var buf [8]byte
	for range 100 {
		if _, err := rand.Read(buf[:]); err != nil {
			return nil, "", err
		}
		rel := ".abcd-tmp-" + hex.EncodeToString(buf[:])
		if dir != "." {
			rel = dir + "/" + rel
		}
		f, err := root.OpenFile(rel, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
		if err == nil {
			return f, rel, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, "", err
		}
	}
	return nil, "", errors.New("fsutil: could not create a temp file in the containment root")
}

// syncDirInRoot fsyncs dir inside root so a crash right after the rename cannot
// lose it. Some filesystems refuse a directory fsync; that is tolerated.
func syncDirInRoot(root *os.Root, dir string) {
	d, err := root.Open(dir)
	if err != nil {
		return
	}
	_ = d.Sync()
	_ = d.Close()
}

// syncParent fsyncs the directory so a crash right after the rename cannot lose
// it. Some filesystems refuse a directory fsync; that is tolerated.
func syncParent(dir string) {
	d, err := os.Open(dir)
	if err != nil {
		return
	}
	_ = d.Sync()
	_ = d.Close()
}

// IsRealDir reports whether path is a directory and NOT a symlink. It lstats
// (never following) so a symlink pointing at a directory reads as false — the
// owned-directory guard the store re-runs on every mutating call.
func IsRealDir(path string) bool {
	fi, err := os.Lstat(path)
	return err == nil && fi.IsDir() && fi.Mode()&os.ModeSymlink == 0
}
