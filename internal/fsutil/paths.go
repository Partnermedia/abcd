package fsutil

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// RepoRel renders target as a path relative to base — the repo root, or the
// working directory — so machine output never carries an absolute
// developer-identity path (iss-81). A target outside base yields a "../…" form,
// which is acceptable: the contract is only that the result is never an absolute
// /Users/<name>/… path. It falls back to the base name when a relative form
// cannot be computed (a different volume, a relative target, or an empty base)
// and to target unchanged when target is already relative. A non-path value that
// merely looks absolute-free (e.g. a URL) is returned untouched.
func RepoRel(base, target string) string {
	if target == "" {
		return target
	}
	if base != "" {
		if rel, err := filepath.Rel(base, target); err == nil {
			return rel
		}
	}
	if filepath.IsAbs(target) {
		return filepath.Base(target)
	}
	return target
}

// notPresent reports whether a stat/open error means the path cannot exist: it
// is absent (ErrNotExist), or a component of its prefix is not a directory
// (ENOTDIR, e.g. asking about a/b where a is a regular file). Both are "not
// present", not a filesystem fault, so a fail-closed caller must not abort on
// them.
func notPresent(err error) bool {
	return errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOTDIR)
}

// Exists reports whether path exists, following symlinks — so a link to a real
// file exists and a dangling link does not. A stat error other than not-exist is
// returned rather than swallowed, so a caller checking a convention fails closed
// on a permission error instead of silently reporting "absent".
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if notPresent(err) {
		return false, nil
	}
	return false, err
}

// IsDir reports whether path exists and is a directory. An absent path is false
// with no error; any other stat error is returned (fail closed).
//
// It follows symlinks: use IsRealDir where a symlinked directory must read as
// false (the owned-store guard).
func IsDir(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil {
		return fi.IsDir(), nil
	}
	if notPresent(err) {
		return false, nil
	}
	return false, err
}

// DirHasEntries reports whether path is a directory holding at least one entry,
// dotfiles included — a directory kept alive by a lone .gitkeep is not empty.
//
// An absent path is false with no error: "missing" and "empty" are distinct
// conditions, and pairing this with Exists lets a presence rule and a non-empty
// rule report independently rather than one masking the other. A path that
// exists but is not a directory is likewise false with no error.
func DirHasEntries(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if notPresent(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()

	names, err := f.Readdirnames(1)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return false, nil // an empty directory
		}
		// Readdirnames on a non-directory errors; that is "no entries", not a
		// broken filesystem, so it stays a soft false.
		if isDir, dirErr := IsDir(path); dirErr == nil && !isDir {
			return false, nil
		}
		return false, err
	}
	return len(names) > 0, nil
}

// ModuleRoot walks up from start until it finds the directory holding go.mod —
// the module root.
//
// It exists for the repo's code generators, which must write to the same file
// whether they are invoked by `go generate` (the working directory is the package
// being generated) or run directly from the repo root. Anchoring on go.mod rather
// than on .git means a generator also works inside a worktree or an export where
// the git directory is not where the walk expects it.
//
// A start outside any module is an error rather than a fallback to the working
// directory: a generator that silently wrote its artefact into an unrelated
// directory would be worse than one that refuses.
func ModuleRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("fsutil: go.mod not found at or above %s", start)
		}
		dir = parent
	}
}
