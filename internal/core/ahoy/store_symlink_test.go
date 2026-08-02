package ahoy

import (
	"os"
	"path/filepath"
	"testing"
)

// TestResolvePluginRootThroughPathSymlink proves the executable-ancestor fallback
// survives the install layout `ahoy install` itself writes: a PATH symlink (e.g.
// /usr/local/bin/abcd) pointing at <plugin-root>/abcd. Without resolving the
// symlink first, the walk climbs the SYMLINK's ancestors (/usr/local/bin,
// /usr/local, /usr — the walk's own guard excludes /, so it never reaches
// that far) and never sees hooks/ — a permanent "plugin root not resolvable"
// gap in a normal shell, where neither env-var candidate is set (iss-170).
func TestResolvePluginRootThroughPathSymlink(t *testing.T) {
	pluginRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	pinnedBin := filepath.Join(pluginRoot, "abcd")
	if err := os.WriteFile(pinnedBin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	// The PATH symlink, in a directory with no plugin layout anywhere above it.
	pathLink := filepath.Join(t.TempDir(), "abcd")
	if err := os.Symlink(pinnedBin, pathLink); err != nil {
		t.Fatal(err)
	}

	// Blank both env candidates so ONLY the executable-ancestor fallback runs.
	t.Setenv("ABCD_PLUGIN_ROOT", "")
	t.Setenv("CLAUDE_PLUGIN_ROOT", "")

	saved := osExecutable
	t.Cleanup(func() { osExecutable = saved })
	osExecutable = func() (string, error) { return pathLink, nil }

	got, ok := resolvePluginRoot()
	if !ok {
		t.Fatalf("resolvePluginRoot() reported no plugin root; want %q via the PATH symlink at %q", pluginRoot, pathLink)
	}
	if want := resolvePath(pluginRoot); got != want {
		t.Errorf("resolvePluginRoot() = %q, want %q", got, want)
	}
}
