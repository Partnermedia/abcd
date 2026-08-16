package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Every binary-invoking hook self-provisions, because SessionStart is a single
// point of failure the field has already seen fail (iss-253, iss-254): a session
// where the SessionStart chain never fired — or ran without CLAUDE_PLUGIN_ROOT
// and took its silent exit — left the plugin root without a binary for hours,
// while every other hook failed as a raw "/bin/sh: .../abcd: No such file or
// directory" on every prompt and every tool call: noisy on each, actionable on
// none. So each hook that calls the binary now (1) exits quietly when no plugin
// root is set, (2) attempts one rate-limited, silent bootstrap salvage when the
// binary is absent, and (3) on continued absence says what is degraded and how
// to fix it in one plain line, with a non-blocking exit. The rate limit is a
// stamp file, so a machine where provisioning cannot succeed pays the download
// timeout at most once per window, not on every hook firing.

// binaryHook is one binary-invoking hook event under test: the event name, an
// optional matcher, and the verb the stub binary must record when the hook runs.
type binaryHook struct {
	event string
	verb  string
}

// binaryHooks enumerates every hook that invokes the binary outside the
// SessionStart chain (which has its own tests and remains the loud, primary
// provisioner).
var binaryHooks = []binaryHook{
	{"UserPromptSubmit", "prompt-router"},
	{"PreToolUse", "guard"},
	{"PreCompact", "prompt-router-reset"},
	{"SessionEnd", "session-end"},
}

// hookCommand returns the single command string for an event, failing on any
// other shape — the one-entry-one-command rule holds for every event, for the
// same parallel-execution reason SessionStart pins it.
func hookCommand(t *testing.T, event string) string {
	t.Helper()
	data, err := os.ReadFile(hooksManifest(t))
	if err != nil {
		t.Fatalf("reading the committed hooks manifest: %v", err)
	}
	var doc sessionStartHooks
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("hooks/hooks.json does not parse: %v", err)
	}
	entries := doc.Hooks[event]
	if len(entries) != 1 || len(entries[0].Hooks) != 1 {
		t.Fatalf("%s must declare exactly one entry group holding one command", event)
	}
	return entries[0].Hooks[0].Command
}

// hookRun executes an event's shipped command under `sh -c` against a plugin
// root, returning stdout, stderr, and the exit code separately. PATH is
// controlled — system utilities plus one caller-supplied directory — because
// the hooks' last resolution rung is a PATH lookup and the developer machine
// running these tests may itself carry a real `abcd` there.
func hookRun(t *testing.T, event, root, pathDir string) (string, string, int) {
	t.Helper()
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh unavailable")
	}
	pathEnv := "/usr/bin:/bin"
	if pathDir != "" {
		pathEnv = pathDir + ":" + pathEnv
	}
	cmd := exec.Command("sh", "-c", hookCommand(t, event))
	cmd.Stdin = strings.NewReader(`{"session_id":"s1"}`)
	cmd.Env = []string{
		"PATH=" + pathEnv,
		"HOME=" + t.TempDir(),
		"CLAUDE_PLUGIN_ROOT=" + root,
		"ABCD_CALLS=" + filepath.Join(root, "calls.log"),
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	code := 0
	if err := cmd.Run(); err != nil {
		e, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("running the %s command: %v (stderr %s)", event, err, stderr.String())
		}
		code = e.ExitCode()
	}
	return stdout.String(), stderr.String(), code
}

// provisioningBootstrap is a stub bootstrap.sh that records its invocation and
// installs a stub binary — the success case of a salvage.
const provisioningBootstrap = `#!/bin/sh
printf 'bootstrap\n' >> "$CLAUDE_PLUGIN_ROOT/boot.log"
cat > "$CLAUDE_PLUGIN_ROOT/abcd" <<'EOF'
#!/bin/sh
cat >/dev/null
printf '%s %s\n' "$1" "$2" >> "$ABCD_CALLS"
exit 0
EOF
chmod +x "$CLAUDE_PLUGIN_ROOT/abcd"
exit 0
`

// failingBootstrap records its invocation and provisions nothing.
const failingBootstrap = `#!/bin/sh
printf 'bootstrap\n' >> "$CLAUDE_PLUGIN_ROOT/boot.log"
exit 1
`

// hookRoot builds a plugin root with the given bootstrap stub and, optionally,
// a pre-existing stub binary.
func hookRoot(t *testing.T, bootstrap string, withBinary bool) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "hooks", "bootstrap.sh"), []byte(bootstrap), 0o755); err != nil {
		t.Fatal(err)
	}
	if withBinary {
		bin := "#!/bin/sh\ncat >/dev/null\nprintf '%s %s\\n' \"$1\" \"$2\" >> \"$ABCD_CALLS\"\nexit 0\n"
		if err := os.WriteFile(filepath.Join(root, "abcd"), []byte(bin), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// TestBinaryHooksProvisionWhenTheBinaryIsAbsent is the iss-253 field failure
// inverted: an empty plugin root plus a working bootstrap must yield a running
// hook, whichever hook fires first.
func TestBinaryHooksProvisionWhenTheBinaryIsAbsent(t *testing.T) {
	for _, h := range binaryHooks {
		t.Run(h.event, func(t *testing.T) {
			root := hookRoot(t, provisioningBootstrap, false)
			_, stderr, _ := hookRun(t, h.event, root, "")
			if !strings.Contains(callLog(t, filepath.Join(root, "calls.log")), h.verb) {
				t.Fatalf("%s did not run the provisioned binary with verb %q; stderr: %s", h.event, h.verb, stderr)
			}
			if strings.Contains(stderr, "No such file") {
				t.Fatalf("%s leaked a raw exec failure: %s", h.event, stderr)
			}
		})
	}
}

// TestBinaryHooksSayWhatIsDegradedWhenUnprovisionable pins the failure UX: one
// plain actionable line naming the install remedy, a non-blocking exit, and no
// raw shell exec error.
func TestBinaryHooksSayWhatIsDegradedWhenUnprovisionable(t *testing.T) {
	for _, h := range binaryHooks {
		t.Run(h.event, func(t *testing.T) {
			root := hookRoot(t, failingBootstrap, false)
			_, stderr, code := hookRun(t, h.event, root, "")
			if code == 0 || code == 127 {
				t.Fatalf("%s exit = %d; want a non-zero, non-exec-failure exit", h.event, code)
			}
			if code == 2 && h.event == "UserPromptSubmit" {
				t.Fatalf("UserPromptSubmit must not exit 2 (it would block the prompt)")
			}
			if strings.Contains(stderr, "No such file") {
				t.Fatalf("%s leaked a raw exec failure: %s", h.event, stderr)
			}
			line := firstLine(stderr)
			if !strings.Contains(line, "abcd") || !strings.Contains(stderr, "#install") {
				t.Fatalf("%s stderr is not one actionable abcd line naming the install remedy: %q", h.event, stderr)
			}
		})
	}
}

// TestBinaryHooksRateLimitTheSalvage: a recent attempt stamp suppresses the
// bootstrap call entirely, so an unprovisionable machine pays the download
// timeout once per window — not once per prompt and per tool call.
func TestBinaryHooksRateLimitTheSalvage(t *testing.T) {
	for _, h := range binaryHooks {
		t.Run(h.event, func(t *testing.T) {
			root := hookRoot(t, failingBootstrap, false)
			stamp := filepath.Join(root, ".bootstrap.attempt")
			if err := os.WriteFile(stamp, nil, 0o644); err != nil {
				t.Fatal(err)
			}
			now := time.Now()
			if err := os.Chtimes(stamp, now, now); err != nil {
				t.Fatal(err)
			}
			hookRun(t, h.event, root, "")
			if callLog(t, filepath.Join(root, "boot.log")) != "" {
				t.Fatalf("%s invoked the bootstrap despite a fresh attempt stamp", h.event)
			}
		})
	}
}

// TestBinaryHooksSteadyStateBypassesTheSalvage: with the binary present, the
// bootstrap is never consulted and the binary's own exit code passes through.
func TestBinaryHooksSteadyStateBypassesTheSalvage(t *testing.T) {
	for _, h := range binaryHooks {
		t.Run(h.event, func(t *testing.T) {
			root := hookRoot(t, failingBootstrap, true)
			_, stderr, code := hookRun(t, h.event, root, "")
			if code != 0 {
				t.Fatalf("%s exit = %d with a healthy binary; stderr: %s", h.event, code, stderr)
			}
			if callLog(t, filepath.Join(root, "boot.log")) != "" {
				t.Fatalf("%s consulted the bootstrap despite a present binary", h.event)
			}
			if !strings.Contains(callLog(t, filepath.Join(root, "calls.log")), h.verb) {
				t.Fatalf("%s did not run the binary with verb %q", h.event, h.verb)
			}
		})
	}
}

// TestBinaryHooksExitQuietlyWithoutAPluginRoot: no CLAUDE_PLUGIN_ROOT means not
// a plugin session; every hook stands down silently instead of failing on an
// unexpanded path.
func TestBinaryHooksExitQuietlyWithoutAPluginRoot(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh unavailable")
	}
	for _, h := range binaryHooks {
		t.Run(h.event, func(t *testing.T) {
			cmd := exec.Command("sh", "-c", hookCommand(t, h.event))
			cmd.Stdin = strings.NewReader("{}")
			cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + t.TempDir()}
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("%s without a plugin root must exit 0; got %v (stderr %s)", h.event, err, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("%s without a plugin root must be silent; stderr: %s", h.event, stderr.String())
			}
		})
	}
}

// TestGuardHookKeepsItsExitCodeFence: the guard's 0/1/2 contract survives the
// salvage wrapper — a blocking verdict still blocks, and an unexpected exit is
// still converted to the loud UNGUARDED warning.
func TestGuardHookKeepsItsExitCodeFence(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh unavailable")
	}
	for _, tc := range []struct{ binExit, want int }{{0, 0}, {1, 1}, {2, 2}, {7, 1}} {
		root := hookRoot(t, failingBootstrap, false)
		bin := "#!/bin/sh\ncat >/dev/null\nexit " + string(rune('0'+tc.binExit)) + "\n"
		if err := os.WriteFile(filepath.Join(root, "abcd"), []byte(bin), 0o755); err != nil {
			t.Fatal(err)
		}
		_, stderr, code := hookRun(t, "PreToolUse", root, "")
		if code != tc.want {
			t.Fatalf("guard exit %d passed through as %d, want %d (stderr %s)", tc.binExit, code, tc.want, stderr)
		}
		if tc.binExit == 7 && !strings.Contains(stderr, "UNGUARDED") {
			t.Fatalf("an unexpected guard exit must warn UNGUARDED; stderr: %s", stderr)
		}
	}
}

// TestBinaryHooksFallBackToAPathBinary: the hooks carry the command surface's
// resolution ladder — plugin root first, PATH second — so a machine where the
// one-line install has run is rescued even when the plugin root cannot be
// provisioned at all.
func TestBinaryHooksFallBackToAPathBinary(t *testing.T) {
	for _, h := range binaryHooks {
		t.Run(h.event, func(t *testing.T) {
			root := hookRoot(t, failingBootstrap, false)
			pathDir := t.TempDir()
			stub := "#!/bin/sh\ncat >/dev/null\nprintf '%s %s\\n' \"$1\" \"$2\" >> \"$ABCD_CALLS\"\nexit 0\n"
			if err := os.WriteFile(filepath.Join(pathDir, "abcd"), []byte(stub), 0o755); err != nil {
				t.Fatal(err)
			}
			_, stderr, code := hookRun(t, h.event, root, pathDir)
			if code != 0 {
				t.Fatalf("%s exit = %d with an abcd on PATH; stderr: %s", h.event, code, stderr)
			}
			if !strings.Contains(callLog(t, filepath.Join(root, "calls.log")), h.verb) {
				t.Fatalf("%s did not run the PATH binary with verb %q; stderr: %s", h.event, h.verb, stderr)
			}
		})
	}
}
