package capture

import (
	"errors"
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
	repo := t.TempDir()
	ir := filepath.Join(repo, "issues")
	if err := ensureLedgerDirs(repo, ir); err != nil {
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

// TestLedgerAncestorSymlinkRefused pins the ledger's ANCESTORS to the same rule
// its leaves already obey. ensureLedgerDirs refused a symlinked issues/ or
// status directory, but provisioned the parents by following whatever was there
// — so a committed `.abcd` or `.abcd/work` symlink in a hostile clone put the
// whole store (the allocator lock included) outside the checkout while the
// result still reported a repo-relative path. Every verb resolves its roots
// first, so the guard there covers the readers too: a List over a redirected
// ledger is refused rather than serialized.
func TestLedgerAncestorSymlinkRefused(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(t *testing.T, repo, outside string)
	}{
		{".abcd/work is a symlink", func(t *testing.T, repo, outside string) {
			if err := os.Mkdir(filepath.Join(repo, ".abcd"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(outside, filepath.Join(repo, ".abcd", "work")); err != nil {
				t.Skipf("symlink unsupported: %v", err)
			}
		}},
		{".abcd is a symlink", func(t *testing.T, repo, outside string) {
			if err := os.Symlink(outside, filepath.Join(repo, ".abcd")); err != nil {
				t.Skipf("symlink unsupported: %v", err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, outside := t.TempDir(), t.TempDir()
			tc.setup(t, repo, outside)
			ir := filepath.Join(repo, LedgerRelPath)

			_, err := Capture(CaptureRequest{
				RepoRoot: repo, IssuesRoot: ir, Text: "b", Severity: SeverityMinor,
				Category: "bug", Source: "manual-test", Slug: "s", FoundDuring: "t",
			})
			if !errors.Is(err, ErrPathUnsafe) {
				t.Fatalf("Capture through a symlinked ancestor: want ErrPathUnsafe, got %v", err)
			}
			entries, rerr := os.ReadDir(outside)
			if rerr != nil {
				t.Fatal(rerr)
			}
			if len(entries) != 0 {
				var names []string
				for _, e := range entries {
					names = append(names, e.Name())
				}
				t.Fatalf("Capture wrote outside the checkout: %v", names)
			}
			if _, err := List(ListRequest{RepoRoot: repo, IssuesRoot: ir}); !errors.Is(err, ErrPathUnsafe) {
				t.Fatalf("List through a symlinked ancestor: want ErrPathUnsafe, got %v", err)
			}
		})
	}
}
