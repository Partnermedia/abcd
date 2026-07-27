package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// preToolUse builds the host's PreToolUse payload for a Bash tool call. The
// adapter reads exactly two things out of it — the tool name and
// tool_input.command — so the fixture spells both out rather than hiding them
// behind a helper's defaults.
func preToolUse(t *testing.T, tool, command, cwd string) string {
	t.Helper()
	payload := map[string]any{
		"session_id":      "s1",
		"cwd":             cwd,
		"hook_event_name": "PreToolUse",
		"tool_name":       tool,
		"tool_input":      map[string]any{"command": command, "description": "x"},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestGuardHookBlocksWithHostExitCode is AC 2 on the guard plane: a blocker match
// exits 2 — the host's blocking status — and puts the successor and the why on
// stderr, which is the channel the host feeds back to the agent. The block is the
// lesson, so both must be present.
func TestGuardHookBlocksWithHostExitCode(t *testing.T) {
	dir := guardRepo(t)
	_, stderr, code := runGuard(preToolUse(t, "Bash", "cd scratch && rm -rf *", dir), "guard", "hook")

	if code != 2 {
		t.Errorf("a blocker must map to the host's blocking exit status 2; got %d (stderr %q)", code, stderr)
	}
	if !strings.Contains(stderr, "absolute path") || !strings.Contains(stderr, "the delete still runs") {
		t.Errorf("the block message must carry the successor and the why; stderr = %q", stderr)
	}
}

// TestGuardHookAllowsSilently is the common case, and the one that must stay
// cheap and quiet: an allowed command produces no output at all.
func TestGuardHookAllowsSilently(t *testing.T) {
	dir := guardRepo(t)
	stdout, stderr, code := runGuard(preToolUse(t, "Bash", "git status --porcelain", dir), "guard", "hook")

	if code != 0 {
		t.Errorf("an allow must exit 0; got %d (stderr %q)", code, stderr)
	}
	if stdout != "" || stderr != "" {
		t.Errorf("an allow must be silent; stdout=%q stderr=%q", stdout, stderr)
	}
}

// TestGuardHookWarnAllowsAndSurfaces is AC 2's warn half on the guard plane: the
// command runs (exit 0, never the blocking status) and the warning is still said
// out loud.
func TestGuardHookWarnAllowsAndSurfaces(t *testing.T) {
	dir := guardRepo(t)
	_, stderr, code := runGuard(preToolUse(t, "Bash", "git clean -fd", dir), "guard", "hook")

	if code != 0 {
		t.Errorf("a warn must never block: want exit 0, got %d", code)
	}
	if !strings.Contains(stderr, "git-clean") {
		t.Errorf("the warning must be surfaced and name its entry; stderr = %q", stderr)
	}
}

// TestGuardHookFailsOpenLoud is AC 1 at the adapter: every input the adapter
// cannot turn into a decision allows the command AND says so unmissably. A silent
// allow would leave a session unguarded with nobody told; a block would brick it.
//
// "Unmissable" is why the exit code is 1 and not 0. A pre-tool-use hook that
// exits 0 has its stderr discarded — the warning would exist and nobody would
// ever see it. A non-zero, non-blocking status is the one channel that both lets
// the command run and puts the warning in front of a human.
func TestGuardHookFailsOpenLoud(t *testing.T) {
	cases := []struct {
		name  string
		stdin func(t *testing.T, dir string) string
	}{
		{"unparsable JSON", func(t *testing.T, _ string) string { return "{not json" }},
		{"empty payload", func(t *testing.T, _ string) string { return "" }},
		{"non-Bash tool call", func(t *testing.T, dir string) string {
			return preToolUse(t, "Write", "cd scratch && rm -rf *", dir)
		}},
		{"Bash call with no command", func(t *testing.T, dir string) string {
			return preToolUse(t, "Bash", "", dir)
		}},
		{"command the tokenizer cannot split", func(t *testing.T, dir string) string {
			return preToolUse(t, "Bash", `rm -rf "unterminated`, dir)
		}},
		{"malformed per-repo registry", func(t *testing.T, dir string) string {
			if err := os.WriteFile(filepath.Join(dir, ".abcd", "guard.json"), []byte("{not json"), 0o644); err != nil {
				t.Fatal(err)
			}
			return preToolUse(t, "Bash", "cd scratch && rm -rf *", dir)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := guardRepo(t)
			_, stderr, code := runGuard(tc.stdin(t, dir), "guard", "hook")

			if code == 2 {
				t.Errorf("must fail OPEN: the blocking status must never come from a non-decision")
			}
			if code == 0 {
				t.Errorf("must fail LOUD: exit 0 discards the hook's stderr, so the warning would never be seen")
			}
			if !strings.Contains(stderr, "abcd guard") {
				t.Errorf("must fail LOUD: stderr must carry an abcd guard warning; stderr = %q", stderr)
			}
		})
	}
}

// TestGuardHookHonoursTheCommittedKillSwitch is AC 2's escape hatch: the only way
// out is a committed, reviewable per-repo config, and when it is set the guard
// allows without pretending to be broken.
func TestGuardHookHonoursTheCommittedKillSwitch(t *testing.T) {
	dir := guardRepo(t)
	cfg := `{"schema_version":1,"disabled":true,"entries":{}}`
	if err := os.WriteFile(filepath.Join(dir, ".abcd", "guard.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, code := runGuard(preToolUse(t, "Bash", "cd scratch && rm -rf *", dir), "guard", "hook")

	if code != 0 {
		t.Errorf("a committed kill switch must allow; got exit %d (stderr %q)", code, stderr)
	}
	if stdout != "" || stderr != "" {
		t.Errorf("a deliberate disable is not a fault and must stay quiet; stdout=%q stderr=%q", stdout, stderr)
	}
}
