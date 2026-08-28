package lint

import (
	"os"
	"path/filepath"
	"testing"
)

// LoadConfigInRoot must resolve every path component inside the root, so a
// symlinked ancestor directory cannot redirect the read at a config outside
// the tree. This is the containment site's other reads have and the config
// read previously lacked (iss-2608270655498099).
func TestLoadConfigInRootRefusesSymlinkedAncestor(t *testing.T) {
	repo := t.TempDir()
	outside := t.TempDir()

	// A config the repo does not own, reachable only through a symlinked ancestor.
	if err := os.WriteFile(filepath.Join(outside, "docs-lint.json"), []byte(`{"schema_version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// repo/.abcd -> outside
	if err := os.Symlink(outside, filepath.Join(repo, ".abcd")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}

	root, err := os.OpenRoot(repo)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	if _, err := LoadConfigInRoot(root, ".abcd/docs-lint.json"); err == nil {
		t.Fatal("LoadConfigInRoot followed a symlinked ancestor out of the root; the read must be refused")
	}
}

// A genuine in-root config still loads and validates through the root path.
func TestLoadConfigInRootLoadsContainedConfig(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".abcd", "docs-lint.json"), []byte(`{"schema_version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(repo)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	if _, err := LoadConfigInRoot(root, ".abcd/docs-lint.json"); err != nil {
		t.Fatalf("a contained config must load: %v", err)
	}
}
