package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomicCreatesWithPerm(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sub", "f.txt") // parent dir does not exist yet
	if err := WriteFileAtomic(p, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("content = %q, want hello", got)
	}
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("perm = %o, want 600", fi.Mode().Perm())
	}
}

func TestWriteFileAtomicOverwrites(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	if err := WriteFileAtomic(p, []byte("first"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(p, []byte("second"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(p)
	if string(got) != "second" {
		t.Fatalf("content = %q, want second", got)
	}
	// No temp files linger after a successful write.
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("expected exactly one file, got %d: %v", len(entries), entries)
	}
}

// TestWriteFileAtomicReplacesSymlink proves the leaf symlink is REPLACED, not
// written through: a pre-planted symlink at path must not clobber its target.
func TestWriteFileAtomicReplacesSymlink(t *testing.T) {
	dir := t.TempDir()
	victim := filepath.Join(dir, "victim.txt")
	if err := os.WriteFile(victim, []byte("do-not-touch"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(victim, link); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(link, []byte("new"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// The symlink is now a real file with the new content...
	fi, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("path is still a symlink; the write followed it")
	}
	// ...and the victim was not written through.
	got, _ := os.ReadFile(victim)
	if string(got) != "do-not-touch" {
		t.Fatalf("symlink target was clobbered: %q", got)
	}
}

func TestWriteFileAtomicPreserveMode(t *testing.T) {
	dir := t.TempDir()

	// New file defaults to 0644.
	fresh := filepath.Join(dir, "fresh.txt")
	if err := WriteFileAtomicPreserveMode(fresh, []byte("x")); err != nil {
		t.Fatal(err)
	}
	if fi, _ := os.Stat(fresh); fi.Mode().Perm() != 0o644 {
		t.Fatalf("new file perm = %o, want 644", fi.Mode().Perm())
	}

	// Existing file keeps its mode across a rewrite.
	kept := filepath.Join(dir, "kept.txt")
	if err := WriteFileAtomic(kept, []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomicPreserveMode(kept, []byte("b")); err != nil {
		t.Fatal(err)
	}
	if fi, _ := os.Stat(kept); fi.Mode().Perm() != 0o600 {
		t.Fatalf("rewritten file perm = %o, want 600 (preserved)", fi.Mode().Perm())
	}
}

func TestIsRealDir(t *testing.T) {
	dir := t.TempDir()
	realDir := filepath.Join(dir, "d")
	if err := os.Mkdir(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "f")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	linkToDir := filepath.Join(dir, "ld")
	if err := os.Symlink(realDir, linkToDir); err != nil {
		t.Fatal(err)
	}

	if !IsRealDir(realDir) {
		t.Errorf("real dir reported as not-real")
	}
	if IsRealDir(file) {
		t.Errorf("file reported as real dir")
	}
	if IsRealDir(linkToDir) {
		t.Errorf("symlink-to-dir reported as real dir (must lstat, not follow)")
	}
	if IsRealDir(filepath.Join(dir, "missing")) {
		t.Errorf("missing path reported as real dir")
	}
}

// TestWriteFileAtomicAppliesPermViaDescriptor guards B16: the mode must be set on
// the open temp descriptor (fchmod), not chmod-by-name on the closed temp. The
// requested perm differs from CreateTemp's 0600 default, so a dropped/no-op chmod
// would surface here as the wrong final mode.
func TestWriteFileAtomicAppliesPermViaDescriptor(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	if err := WriteFileAtomic(p, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o644 {
		t.Fatalf("perm = %o, want 644 — mode not applied via the descriptor", fi.Mode().Perm())
	}
	// No temp file may be left behind.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if len(e.Name()) >= 10 && e.Name()[:10] == ".abcd-tmp-" {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}
}

// TestWriteFileAtomicPreserveModeFailsClosedOnStatFault guards B17: a real stat
// fault (here ELOOP from a self-referencing symlink at the target) is NOT
// absence, so PreserveMode must fail closed rather than silently default to 0644
// and write through — which would widen an existing restrictive mode.
func TestWriteFileAtomicPreserveModeFailsClosedOnStatFault(t *testing.T) {
	dir := t.TempDir()
	loop := filepath.Join(dir, "loop")
	if err := os.Symlink("loop", loop); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	// os.Stat(loop) now fails with ELOOP (a genuine fault, not not-exist).
	if err := WriteFileAtomicPreserveMode(loop, []byte("data")); err == nil {
		t.Fatal("want error on a real stat fault (ELOOP), got nil — mode would be silently reset to 0644")
	}
}
