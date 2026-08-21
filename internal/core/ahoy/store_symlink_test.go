package ahoy

import (
	"crypto/sha256"
	"encoding/hex"
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

// TestResolvePluginRootThroughOwnedCopyRecord pins the SAME iss-170 contract by
// BEHAVIOUR rather than by the symlink mechanism: spc-35 replaced the pinned
// PATH symlink with an abcd-owned REGULAR-FILE copy, which has no link back into
// the plugin root for the executable-ancestor walk to follow home. The
// home-scoped provenance record carries the plugin root the copy was
// provisioned from, and resolvePluginRoot reads it as a candidate — otherwise
// every plugin-root verb no-ops from a terminal, where neither env candidate is
// set (iss-2608210934566230).
func TestResolvePluginRootThroughOwnedCopyRecord(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	pluginRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	// The owned copy sits in a directory with no plugin layout anywhere above it.
	binDir := t.TempDir()
	copyBin := filepath.Join(binDir, "abcd")
	body := []byte("#!/bin/sh\nexit 0\n")
	if err := os.WriteFile(copyBin, body, 0o755); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	if err := writePathEntry(copyBin, hex.EncodeToString(sum[:]), pluginRoot); err != nil {
		t.Fatal(err)
	}

	// Blank every env candidate so ONLY the record route can resolve the root.
	t.Setenv("ABCD_PLUGIN_ROOT", "")
	t.Setenv("CLAUDE_PLUGIN_ROOT", "")
	t.Setenv("CLAUDE_PLUGIN_DATA", "")

	saved := osExecutable
	t.Cleanup(func() { osExecutable = saved })
	osExecutable = func() (string, error) { return copyBin, nil }

	got, ok := resolvePluginRoot()
	if !ok {
		t.Fatalf("resolvePluginRoot() reported no plugin root; want %q via the owned-copy record", pluginRoot)
	}
	// The record stores the root verbatim (as the env candidates are returned
	// verbatim), so compare canonically rather than assuming a resolved form.
	if resolvePath(got) != resolvePath(pluginRoot) {
		t.Errorf("resolvePluginRoot() = %q, want %q", got, pluginRoot)
	}

	// A recorded root the harness has since garbage-collected is skipped, not
	// trusted: pluginRootValid still gates the candidate.
	if err := os.RemoveAll(pluginRoot); err != nil {
		t.Fatal(err)
	}
	if _, ok := resolvePluginRoot(); ok {
		t.Error("a garbage-collected recorded plugin root must not resolve")
	}
}
