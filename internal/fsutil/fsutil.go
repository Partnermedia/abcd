// Package fsutil holds the durable-write and path-safety primitives shared by
// the ~/.abcd store writers. It is transport-agnostic: no stdout, no os.Exit,
// no CLI knowledge.
//
// It exists so the six-step atomic write and the "is this a real directory,
// not a symlink" check live in ONE place. ahoy's marker.go/store.go carry
// their own unexported copies (writeFileAtomic, isRealDir) predating this
// package; those are the flagged consolidation target — a follow-up
// behaviour-preserving refactor should route them through here rather than
// keep a divergent copy.
package fsutil

import (
	"os"
	"path/filepath"
)

// WriteFileAtomic writes data to path durably: a temp file in the target
// directory is written, flushed, fsync'd, chmod'd to perm, then renamed over
// the target, and finally the parent directory is fsync'd best-effort so the
// rename survives a crash. Parent directories are created as needed.
//
// The rename is the commit point: a reader sees either the old file or the
// complete new one, never a half-written file. os.Rename does not follow a
// symlink at the leaf — a pre-planted symlink at path is replaced by the real
// file, not written through.
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
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
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
