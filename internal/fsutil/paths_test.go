package fsutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/REPPL/abcd-cli/internal/fsutil"
)

func TestExists(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "present.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"regular file", file, true},
		{"directory", dir, true},
		{"absent", filepath.Join(dir, "nope.txt"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fsutil.Exists(tt.path)
			if err != nil {
				t.Fatalf("Exists(%q) returned error: %v", tt.path, err)
			}
			if got != tt.want {
				t.Errorf("Exists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// Exists follows symlinks: a link to a real file exists, a dangling link does not.
func TestExistsSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	good := filepath.Join(dir, "good.link")
	if err := os.Symlink(target, good); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
	dangling := filepath.Join(dir, "dangling.link")
	if err := os.Symlink(filepath.Join(dir, "absent"), dangling); err != nil {
		t.Fatal(err)
	}

	if got, err := fsutil.Exists(good); err != nil || !got {
		t.Errorf("Exists(link to file) = %v, %v; want true, nil", got, err)
	}
	if got, err := fsutil.Exists(dangling); err != nil || got {
		t.Errorf("Exists(dangling link) = %v, %v; want false, nil", got, err)
	}
}

// A path whose parent component is a regular file cannot exist; stat returns
// ENOTDIR, and Exists/IsDir/DirHasEntries must read that as "not present"
// (false, nil), not propagate it as a hard error — otherwise a caller that
// fails closed on errors aborts on an obviously-absent path.
func TestPathUnderAFileIsNotPresent(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "afile")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	under := filepath.Join(file, "child") // afile is not a directory

	if got, err := fsutil.Exists(under); err != nil || got {
		t.Errorf("Exists(under-a-file) = %v, %v; want false, nil", got, err)
	}
	if got, err := fsutil.IsDir(under); err != nil || got {
		t.Errorf("IsDir(under-a-file) = %v, %v; want false, nil", got, err)
	}
	if got, err := fsutil.DirHasEntries(under); err != nil || got {
		t.Errorf("DirHasEntries(under-a-file) = %v, %v; want false, nil", got, err)
	}
}

func TestIsDir(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if got, err := fsutil.IsDir(dir); err != nil || !got {
		t.Errorf("IsDir(dir) = %v, %v; want true, nil", got, err)
	}
	if got, err := fsutil.IsDir(file); err != nil || got {
		t.Errorf("IsDir(file) = %v, %v; want false, nil", got, err)
	}
	if got, err := fsutil.IsDir(filepath.Join(dir, "absent")); err != nil || got {
		t.Errorf("IsDir(absent) = %v, %v; want false, nil", got, err)
	}
}

func TestDirHasEntries(t *testing.T) {
	empty := t.TempDir()
	full := t.TempDir()
	if err := os.WriteFile(filepath.Join(full, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if got, err := fsutil.DirHasEntries(empty); err != nil || got {
		t.Errorf("DirHasEntries(empty) = %v, %v; want false, nil", got, err)
	}
	if got, err := fsutil.DirHasEntries(full); err != nil || !got {
		t.Errorf("DirHasEntries(full) = %v, %v; want true, nil", got, err)
	}
}

// An absent directory is not an error — it simply holds no entries. The caller
// distinguishes "missing" from "empty" with Exists, so a presence rule and a
// non-empty rule stay independent.
func TestDirHasEntriesAbsent(t *testing.T) {
	got, err := fsutil.DirHasEntries(filepath.Join(t.TempDir(), "absent"))
	if err != nil {
		t.Fatalf("DirHasEntries(absent) returned error: %v", err)
	}
	if got {
		t.Errorf("DirHasEntries(absent) = true, want false")
	}
}

// A dotfile is an entry. A directory holding only .gitkeep is not empty.
func TestDirHasEntriesDotfileCounts(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".gitkeep"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := fsutil.DirHasEntries(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Errorf("DirHasEntries(dir with only .gitkeep) = false, want true")
	}
}
