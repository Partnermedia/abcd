package fsutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestModuleRootWalksUpToGoMod covers the three cases a generator meets: invoked
// from the module root, invoked from a nested package directory (what `go
// generate` does), and invoked from outside any module.
func TestModuleRootWalksUpToGoMod(t *testing.T) {
	root := t.TempDir()
	// t.TempDir can hand back a symlinked path (/var vs /private/var on macOS);
	// resolve it so the comparison is about the walk, not about link spelling.
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}
	if err := os.WriteFile(filepath.Join(resolved, "go.mod"), []byte("module example.invalid\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	nested := filepath.Join(resolved, "internal", "surface", "cli")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	tests := []struct {
		name  string
		start string
	}{
		{"from the module root", resolved},
		{"from a nested package directory", nested},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ModuleRoot(tc.start)
			if err != nil {
				t.Fatalf("ModuleRoot(%s): %v", tc.start, err)
			}
			if got != resolved {
				t.Fatalf("ModuleRoot(%s) = %q, want %q", tc.start, got, resolved)
			}
		})
	}

	t.Run("outside any module", func(t *testing.T) {
		outside := t.TempDir()
		if _, err := ModuleRoot(outside); err == nil {
			t.Fatalf("ModuleRoot(%s) = nil error, want one", outside)
		} else if !strings.Contains(err.Error(), "go.mod") {
			t.Fatalf("error = %q, want it to name go.mod", err)
		}
	})
}
