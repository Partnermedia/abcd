package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestVersionJSON proves the CLI -> core -> JSON round-trip the Phase 0 exit
// criterion requires.
func TestVersionJSON(t *testing.T) {
	out := runCLI(t, "version", "--json")

	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out)
	}
	if got["name"] != "abcd" {
		t.Fatalf("name = %v, want abcd", got["name"])
	}
	if got["version"] == "" || got["version"] == nil {
		t.Fatalf("version missing: %v", got)
	}
}

func TestVersionText(t *testing.T) {
	out := runCLI(t, "version")
	if !strings.HasPrefix(string(out), "abcd ") {
		t.Fatalf("text output = %q, want it to start with \"abcd \"", out)
	}
}

func TestBareStatusJSON(t *testing.T) {
	out := runCLI(t, "--json")
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("bare status output is not JSON: %v\n%s", err, out)
	}
	if _, ok := got["dir"]; !ok {
		t.Fatalf("status JSON missing dir: %v", got)
	}
}

// validHooksJSON is a structurally-sound plugin hook manifest for the hermetic
// plugin root, so the install path's hook-manifest verification passes.
const validHooksJSON = `{
  "hooks": {
    "UserPromptSubmit": [{"hooks": [{"type": "command", "command": "$CLAUDE_PLUGIN_ROOT/hooks/prompt_router_hook"}]}],
    "SessionStart":     [{"hooks": [{"type": "command", "command": "$CLAUDE_PLUGIN_ROOT/hooks/prompt_router_reset"}]}],
    "PreCompact":       [{"hooks": [{"type": "command", "command": "$CLAUDE_PLUGIN_ROOT/hooks/prompt_router_reset"}]}]
  }
}`

// hermeticRepo redirects HOME, the plugin root and the PATH symlink target to
// temp locations, chdirs into a fresh adoptable repo, and returns its path.
func hermeticRepo(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	pluginRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginRoot, "hooks", "hooks.json"), []byte(validHooksJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginRoot, "abcd"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("ABCD_PLUGIN_ROOT", pluginRoot)
	t.Setenv("CLAUDE_PLUGIN_ROOT", "")
	t.Setenv("ABCD_BIN_TARGET", filepath.Join(t.TempDir(), "bin", "abcd"))

	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)
	return repo
}

// TestAhoyInstallWiredAndIdempotent proves `abcd ahoy install` reaches the core
// engine from the CLI front door (the Phase 1 install milestone) and that a
// re-run is an exact no-op.
func TestAhoyInstallWiredAndIdempotent(t *testing.T) {
	repo := hermeticRepo(t)

	out := runCLI(t, "ahoy", "install", "--yes", "--adopt",
		"--visibility", "private", "--docs-target", "both",
		"--oracle-backend", "host-delegated", "--scan-deep", "false", "--json")
	var res struct {
		Status string   `json:"status"`
		Writes []string `json:"writes"`
	}
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("install output not JSON: %v\n%s", err, out)
	}
	if res.Status != "clean" {
		t.Fatalf("install status = %q, want clean\n%s", res.Status, out)
	}
	// The marker block reached disk via the CLI path.
	body, err := os.ReadFile(filepath.Join(repo, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("CLAUDE.md not written: %v", err)
	}
	if !strings.Contains(string(body), "<!-- BEGIN ABCD -->") {
		t.Fatalf("CLAUDE.md has no marker block:\n%s", body)
	}

	// Second run is an exact no-op.
	out2 := runCLI(t, "ahoy", "install", "--yes", "--adopt",
		"--visibility", "private", "--docs-target", "both",
		"--oracle-backend", "host-delegated", "--scan-deep", "false", "--json")
	var res2 struct {
		Status string   `json:"status"`
		Writes []string `json:"writes"`
	}
	if err := json.Unmarshal(out2, &res2); err != nil {
		t.Fatalf("re-install output not JSON: %v\n%s", err, out2)
	}
	if res2.Status != "already_up_to_date" {
		t.Fatalf("re-install status = %q, want already_up_to_date", res2.Status)
	}
	if len(res2.Writes) != 0 {
		t.Fatalf("re-install wrote files: %v", res2.Writes)
	}
}

func runCLI(t *testing.T, args ...string) []byte {
	t.Helper()
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute %v: %v\n%s", args, err, out.String())
	}
	return out.Bytes()
}
