package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRulesRootWalksUpToDotAbcd is the attack/behaviour test for the rules-loader
// root fix: run from a subdirectory, rulesRoot must return the nearest ancestor
// holding a .abcd directory, not cwd. Handing rules.Load a subdirectory silently
// ignored the per-repo overrides AND the kill switch, so a repo that had disabled
// a domain would still inject it from any nested directory.
func TestRulesRootWalksUpToDotAbcd(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(repo, "internal", "deep", "pkg")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	// Resolve symlinks so macOS /var -> /private/var does not defeat the compare.
	wantRepo, _ := filepath.EvalSymlinks(repo)
	got, _ := filepath.EvalSymlinks(rulesRoot(sub))
	if got != wantRepo {
		t.Errorf("rulesRoot(%q) = %q, want the .abcd-bearing ancestor %q", sub, got, wantRepo)
	}
	// From the repo root itself, rulesRoot returns it unchanged.
	gotRoot, _ := filepath.EvalSymlinks(rulesRoot(repo))
	if gotRoot != wantRepo {
		t.Errorf("rulesRoot(repo root) = %q, want %q", gotRoot, wantRepo)
	}
}
