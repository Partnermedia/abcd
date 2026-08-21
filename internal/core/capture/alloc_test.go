package capture

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestOrphanStillRemovableRejectsCommittedFile (B26) proves the pre-unlink guard
// refuses to remove a placeholder that a capture commit has replaced/filled in
// the sweep's TOCTOU window: a zero-byte inode classified as an orphan that
// becomes a non-empty committed file must no longer be removable.
func TestOrphanStillRemovableRejectsCommittedFile(t *testing.T) {
	dir := t.TempDir()
	cand := filepath.Join(dir, "iss-1-note.md")
	if err := os.WriteFile(cand, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	seen, err := os.Lstat(cand)
	if err != nil {
		t.Fatal(err)
	}
	// Baseline: a still-empty, still-same inode is removable.
	if !orphanStillRemovable(cand, seen) {
		t.Fatal("a genuine zero-byte orphan must still be removable")
	}
	// A concurrent commit replaces the placeholder with a full issue file via
	// atomic rename (temp file + rename -> new, non-empty inode).
	tmp := filepath.Join(dir, "iss-1-note.md.tmp")
	if err := os.WriteFile(tmp, []byte("---\nid: \"iss-1\"\n---\n\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmp, cand); err != nil {
		t.Fatal(err)
	}
	if orphanStillRemovable(cand, seen) {
		t.Fatal("the sweep must not remove a placeholder a commit has since filled")
	}
}

// TestCleanOrphanPlaceholdersStillSweepsAgedOrphan guards against the B26 guard
// regressing normal sweep behaviour: a genuinely aged zero-byte placeholder is
// still removed.
func TestCleanOrphanPlaceholdersStillSweepsAgedOrphan(t *testing.T) {
	ir := filepath.Join(t.TempDir(), "issues")
	if err := ensureLedgerDirs(ir); err != nil {
		t.Fatal(err)
	}
	orphan := filepath.Join(ir, "open", "iss-1-note.md")
	if err := os.WriteFile(orphan, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-2 * orphanAgeThreshold)
	if err := os.Chtimes(orphan, old, old); err != nil {
		t.Fatal(err)
	}
	if err := cleanOrphanPlaceholders(ir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(orphan); !os.IsNotExist(err) {
		t.Fatalf("aged zero-byte orphan should have been swept, Lstat err=%v", err)
	}
}
