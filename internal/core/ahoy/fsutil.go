package ahoy

import (
	"os"
	"path/filepath"
)

// modeSymlink aliases os.ModeSymlink so detection reads naturally.
const modeSymlink = os.ModeSymlink

// fileExists reports whether path exists as a regular file.
func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.Mode().IsRegular()
}

func lstat(p string) (os.FileInfo, error) { return os.Lstat(p) }

func isNotExist(err error) bool { return os.IsNotExist(err) }

func readlink(p string) (string, error) { return os.Readlink(p) }

// resolvePath returns a canonical absolute form of p for symlink-target
// comparison, tolerating non-existent paths.
func resolvePath(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	if a, err := filepath.Abs(p); err == nil {
		return a
	}
	return p
}
