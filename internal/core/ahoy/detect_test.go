package ahoy

import (
	"os"
	"path/filepath"
	"testing"
)

// validHooksJSON is a structurally-sound plugin hook manifest.
const validHooksJSON = `{
  "hooks": {
    "UserPromptSubmit": [{"hooks": [{"type": "command", "command": "\"$CLAUDE_PLUGIN_ROOT/abcd\" hook prompt-router"}]}],
    "SessionStart":     [{"hooks": [{"type": "command", "command": "\"$CLAUDE_PLUGIN_ROOT/abcd\" hook prompt-router-reset"}]}],
    "PreCompact":       [{"hooks": [{"type": "command", "command": "\"$CLAUDE_PLUGIN_ROOT/abcd\" hook prompt-router-reset"}]}]
  }
}`

// setupHermetic redirects HOME, the plugin root, and the PATH symlink target to
// temp locations so a test never touches the real machine.
func setupHermetic(t *testing.T) (home, pluginRoot string) {
	t.Helper()
	home = t.TempDir()
	pluginRoot = t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginRoot, "hooks", "hooks.json"), []byte(validHooksJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginRoot, "abcd"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	binTargetPath := filepath.Join(t.TempDir(), "bin", "abcd")
	t.Setenv("HOME", home)
	t.Setenv("ABCD_PLUGIN_ROOT", pluginRoot)
	t.Setenv("CLAUDE_PLUGIN_ROOT", "")
	t.Setenv("ABCD_BIN_TARGET", binTargetPath)
	return home, pluginRoot
}

func TestClassifyUnmanagedFolder(t *testing.T) {
	setupHermetic(t)
	dir := t.TempDir()
	det, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if det.FolderKind != UnmanagedFolder {
		t.Errorf("kind = %q, want %q", det.FolderKind, UnmanagedFolder)
	}
}

func TestClassifyUnmanagedRepo(t *testing.T) {
	setupHermetic(t)
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	det, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if det.FolderKind != UnmanagedRepo {
		t.Errorf("kind = %q, want %q", det.FolderKind, UnmanagedRepo)
	}
}

func TestClassifyManagedRepoByAbcdDir(t *testing.T) {
	setupHermetic(t)
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	det, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if det.FolderKind != ManagedRepo {
		t.Errorf("kind = %q, want %q", det.FolderKind, ManagedRepo)
	}
}

func TestClassifyManagedRepoByMarker(t *testing.T) {
	setupHermetic(t)
	dir := t.TempDir()
	// A CLAUDE.md carrying a BEGIN fence is a strong managed signal even with
	// no .git and no .abcd dir.
	body := "# Project\n\n<!-- BEGIN ABCD -->\nx\n<!-- END ABCD -->\n"
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	det, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if det.FolderKind != ManagedRepo {
		t.Errorf("kind = %q, want %q", det.FolderKind, ManagedRepo)
	}
}

func TestSymlinkedMarkerNotAManagedSignal(t *testing.T) {
	setupHermetic(t)
	dir := t.TempDir()
	// Plant a real file elsewhere, symlink CLAUDE.md to it. A symlinked marker
	// doc must NOT promote the folder to managed.
	real := filepath.Join(t.TempDir(), "real.md")
	if err := os.WriteFile(real, []byte("<!-- BEGIN ABCD -->\nx\n<!-- END ABCD -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, filepath.Join(dir, "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}
	det, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if det.FolderKind != UnmanagedFolder {
		t.Errorf("kind = %q, want %q (symlinked marker must not count)", det.FolderKind, UnmanagedFolder)
	}
}

func TestDetectUnmanagedFolderShortCircuits(t *testing.T) {
	setupHermetic(t)
	dir := t.TempDir()
	det, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	// An unmanaged folder runs no other detection checks — no gaps at all.
	if len(det.Gaps) != 0 {
		t.Errorf("unmanaged folder produced gaps: %+v", det.Gaps)
	}
}

func TestDetectHookManifestGapOnBrokenPlugin(t *testing.T) {
	home := t.TempDir()
	pluginRoot := t.TempDir()
	// hooks dir exists (so the root validates) but hooks.json is absent.
	if err := os.MkdirAll(filepath.Join(pluginRoot, "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("ABCD_PLUGIN_ROOT", pluginRoot)
	t.Setenv("CLAUDE_PLUGIN_ROOT", "")
	t.Setenv("ABCD_BIN_TARGET", filepath.Join(t.TempDir(), "abcd"))

	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	det, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !hasGap(det.Gaps, "hooks.manifest_missing") {
		t.Errorf("expected hooks.manifest_missing gap; got %+v", det.Gaps)
	}
	// It is a non-resolvable diagnostic, never actionable.
	for _, g := range det.Gaps {
		if g.ID == "hooks.manifest_missing" && (g.Resolvable || g.Required) {
			t.Errorf("hooks.manifest_missing must be non-resolvable/advisory: %+v", g)
		}
	}
}

func hasGap(gaps []Gap, id string) bool {
	for _, g := range gaps {
		if g.ID == id {
			return true
		}
	}
	return false
}

// TestShippedHookManifestVerifies pins the real hooks/hooks.json against the
// verifier, so the manifest and requiredHookCommand can never drift apart.
func TestShippedHookManifestVerifies(t *testing.T) {
	if reason := verifyHookManifest("../../.."); reason != "" {
		t.Fatalf("shipped hooks/hooks.json fails verification: %s", reason)
	}
}
