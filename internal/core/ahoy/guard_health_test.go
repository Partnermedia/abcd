package ahoy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// hooksJSONWithGuard is a manifest that also arms the execution-time guard.
const hooksJSONWithGuard = `{
  "hooks": {
    "UserPromptSubmit": [{"hooks": [{"type": "command", "command": "\"$CLAUDE_PLUGIN_ROOT/abcd\" hook prompt-router"}]}],
    "SessionStart":     [{"hooks": [{"type": "command", "command": "\"$CLAUDE_PLUGIN_ROOT/abcd\" hook prompt-router-reset"}]}],
    "PreCompact":       [{"hooks": [{"type": "command", "command": "\"$CLAUDE_PLUGIN_ROOT/abcd\" hook prompt-router-reset"}]}],
    "PreToolUse":       [{"matcher": "Bash", "hooks": [{"type": "command", "command": "\"$CLAUDE_PLUGIN_ROOT/abcd\" guard hook"}]}]
  }
}`

// managedRepoAt turns dir into a folder ahoy classifies as a managed repo, so the
// full detection pass runs over it.
func managedRepoAt(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := append(append(append([]byte(nil), markerBegin...), []byte("\nx\n")...), markerEnd...)
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), append(body, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestGuardHealthArmed is the good state: the manifest carries the PreToolUse
// entry, the plugin-root binary is present, and the registry loads. Everything a
// facilitator needs to believe the guard is actually on.
func TestGuardHealthArmed(t *testing.T) {
	_, pluginRoot := setupHermetic(t)
	if err := os.WriteFile(filepath.Join(pluginRoot, "hooks", "hooks.json"), []byte(hooksJSONWithGuard), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	managedRepoAt(t, dir)

	det, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	g := det.Guard
	if !g.HookInstalled || !g.BinaryReachable || !g.RegistryLoadable {
		t.Errorf("a fully armed guard must report healthy; got %+v", g)
	}
	if !g.Healthy() {
		t.Errorf("Healthy() must agree with the three checks; got %+v", g)
	}
	for _, gap := range det.Gaps {
		if strings.HasPrefix(gap.ID, "guard.") {
			t.Errorf("a healthy guard must emit no guard gap; got %q", gap.ID)
		}
	}
}

// TestGuardHealthHookNotInstalled is AC 1's visibility half: a manifest without
// the PreToolUse entry means commands run unguarded, and ahoy must say so rather
// than report a clean bill of health.
func TestGuardHealthHookNotInstalled(t *testing.T) {
	setupHermetic(t) // validHooksJSON: no PreToolUse entry
	dir := t.TempDir()
	managedRepoAt(t, dir)

	det, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if det.Guard.HookInstalled {
		t.Error("a manifest with no PreToolUse guard entry must report hook_installed=false")
	}
	if det.Guard.Healthy() {
		t.Error("a guard with no hook is not healthy")
	}
	if !hasGap(det.Gaps, "guard.hook_missing") {
		t.Errorf("a missing guard hook must surface as a gap; gaps = %v", gapIDs(det.Gaps))
	}
}

// TestGuardHealthBinaryUnreachable covers the fail-open case ahoy exists to make
// visible: the hook is installed but the binary it calls is gone, so every
// session runs unguarded behind a shim warning.
func TestGuardHealthBinaryUnreachable(t *testing.T) {
	_, pluginRoot := setupHermetic(t)
	if err := os.WriteFile(filepath.Join(pluginRoot, "hooks", "hooks.json"), []byte(hooksJSONWithGuard), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(pluginRoot, "abcd")); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	managedRepoAt(t, dir)

	det, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if det.Guard.BinaryReachable {
		t.Error("an absent plugin-root binary must report binary_reachable=false")
	}
	if !hasGap(det.Gaps, "guard.binary_unreachable") {
		t.Errorf("an unreachable guard binary must surface as a gap; gaps = %v", gapIDs(det.Gaps))
	}
}

// TestGuardHealthRegistryUnloadable is the third failure mode: the hook runs, the
// binary runs, and the registry the repo committed does not parse — so the guard
// refuses to answer and the session is unguarded. Never silent.
func TestGuardHealthRegistryUnloadable(t *testing.T) {
	_, pluginRoot := setupHermetic(t)
	if err := os.WriteFile(filepath.Join(pluginRoot, "hooks", "hooks.json"), []byte(hooksJSONWithGuard), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	managedRepoAt(t, dir)
	if err := os.WriteFile(filepath.Join(dir, ".abcd", "guard.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	det, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if det.Guard.RegistryLoadable {
		t.Error("a malformed .abcd/guard.json must report registry_loadable=false")
	}
	if det.Guard.Detail == "" {
		t.Error("an unloadable registry must carry a reason a human can act on")
	}
	if !hasGap(det.Gaps, "guard.registry_unloadable") {
		t.Errorf("an unloadable registry must surface as a gap; gaps = %v", gapIDs(det.Gaps))
	}
}

// TestGuardHealthDisabledIsReported is the committed escape hatch: a repo that
// deliberately turned the guard off is not broken, but the fact must still be on
// the status board — a disabled guard that looks armed is the failure mode.
func TestGuardHealthDisabledIsReported(t *testing.T) {
	_, pluginRoot := setupHermetic(t)
	if err := os.WriteFile(filepath.Join(pluginRoot, "hooks", "hooks.json"), []byte(hooksJSONWithGuard), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	managedRepoAt(t, dir)
	cfg := `{"schema_version":1,"disabled":true,"entries":{}}`
	if err := os.WriteFile(filepath.Join(dir, ".abcd", "guard.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	det, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !det.Guard.RegistryLoadable {
		t.Error("a deliberately disabled registry still loads; it is not a fault")
	}
	if !det.Guard.Disabled {
		t.Error("a committed kill switch must be reported by ahoy, not hidden")
	}
}
