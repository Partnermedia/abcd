package cli

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/Partnermedia/abcd/internal/gittest"
)

// gitRepoNoStore builds an isolated git repo with one commit and a hermetic HOME
// whose ~/.abcd does NOT exist — the "plugin enabled but never installed" state
// iss-95 is about. Returns the repo dir.
func gitRepoNoStore(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	repo := t.TempDir()
	env := append(gittest.Env(t),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@e",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@e",
	)
	for _, args := range [][]string{{"init", "-q"}, {"commit", "-q", "--allow-empty", "-m", "root"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	t.Setenv("HOME", t.TempDir()) // hermetic, empty: no ~/.abcd
	return repo
}

// startPayload is the SessionStart-hook JSON the harness writes to stdin.
func startPayload(session, cwd string) string {
	return `{"session_id":"` + session + `","cwd":"` + cwd + `","hook_event_name":"SessionStart"}`
}

// runSessionStart drives the verb through the real exit-code mapping, with stdin
// wired, and returns what the harness would see: stdout, stderr, and the process
// exit code. SessionStart shows a hook's stderr only on a non-zero exit, so the
// code is load-bearing here, not incidental.
func runSessionStart(stdin string, args ...string) (stdout, stderr string, code int) {
	root := NewRootCommand()
	root.SetArgs(args)
	var so, se bytes.Buffer
	root.SetOut(&so)
	root.SetErr(&se)
	root.SetIn(strings.NewReader(stdin))
	err := root.Execute()
	if err == nil {
		return so.String(), se.String(), 0
	}
	var coded interface{ ExitCode() int }
	if errors.As(err, &coded) {
		return so.String(), se.String(), coded.ExitCode()
	}
	return so.String(), se.String(), 1
}

// TestHookSessionStartWarnsWhenStoreMissing is iss-95: a session that begins in a
// repo where abcd is not installed must SAY SO — visibly — so the user isn't left
// with a silently non-accruing transcript corpus. SessionStart shows a hook's
// stderr as a visible notice only on a non-zero exit, and never blocks on it.
func TestHookSessionStartWarnsWhenStoreMissing(t *testing.T) {
	repo := gitRepoNoStore(t)
	_, stderr, code := runSessionStart(startPayload("s1", repo), "hook", "session-start")

	if code == 0 {
		t.Error("a missing store must exit non-zero so SessionStart renders the notice; got exit 0 (silent)")
	}
	if !strings.Contains(stderr, "ahoy install") {
		t.Errorf("the notice must tell the user how to fix it (ahoy install); stderr = %q", stderr)
	}
}

// noAmbientPluginRoot points the plugin-root resolution at an empty directory.
// The session-start notices now include one read off `$CLAUDE_PLUGIN_ROOT/
// .binary-meta` (itd-105), so a silence assertion that inherits the ambient
// environment asserts something about the machine it runs on: inside a session
// where the abcd plugin is installed and skewed, "must be silent" fails for a
// reason that has nothing to do with the case under test.
func noAmbientPluginRoot(t *testing.T) {
	t.Helper()
	empty := t.TempDir()
	t.Setenv("ABCD_PLUGIN_ROOT", empty)
	t.Setenv("CLAUDE_PLUGIN_ROOT", empty)
}

// TestHookSessionStartSilentWhenStoreReady is the common case: an installed repo
// must start with no notice at all.
func TestHookSessionStartSilentWhenStoreReady(t *testing.T) {
	repo, _ := sessionEndRepo(t) // creates the hermetic store's transcripts dir
	noAmbientPluginRoot(t)
	stdout, stderr, code := runSessionStart(startPayload("s2", repo), "hook", "session-start")

	if code != 0 {
		t.Errorf("a ready store must exit 0, got %d (stderr %q)", code, stderr)
	}
	if stderr != "" || stdout != "" {
		t.Errorf("a ready store must be silent; stdout=%q stderr=%q", stdout, stderr)
	}
}

// TestHookSessionStartSilentAndNonBlocking holds the never-disrupt contract for
// cases that are not a missing-store problem: a non-repo cwd (capture would skip
// for a different reason, and no install fixes it) and a malformed or empty
// payload must each stay silent and exit 0.
func TestHookSessionStartSilentAndNonBlocking(t *testing.T) {
	cases := []struct {
		name  string
		stdin func(t *testing.T) string
	}{
		{"cwd is not a git repo", func(t *testing.T) string {
			t.Setenv("HOME", t.TempDir())
			return startPayload("s", t.TempDir())
		}},
		{"malformed payload", func(t *testing.T) string {
			t.Setenv("HOME", t.TempDir())
			return "{not json"
		}},
		{"empty payload", func(t *testing.T) string {
			t.Setenv("HOME", t.TempDir())
			return ""
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			noAmbientPluginRoot(t)
			stdout, stderr, code := runSessionStart(tc.stdin(t), "hook", "session-start")
			if code != 0 {
				t.Errorf("must exit 0 (not a store problem), got %d", code)
			}
			if stdout != "" || stderr != "" {
				t.Errorf("must be silent; stdout=%q stderr=%q", stdout, stderr)
			}
		})
	}
}
