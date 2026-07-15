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
package fsutil

import (
	"os"
	"path/filepath"
)

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
