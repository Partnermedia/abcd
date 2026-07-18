package lifeboat

import (
	"os"
	"strings"
	"testing"
)

// TestMain disables git's asynchronous background operations for every test in
// this package. The tests build real git repositories under t.TempDir(), and
// commands like `git commit` / `git fast-import` over substantial history can
// trip git's auto-gc, which by default DETACHES a background `git gc` process
// (gc.autoDetach). That detached process keeps writing into .git after the
// foreground command has returned, and races Go's t.TempDir() cleanup — which
// then fails with "unlinkat .../.git: directory not empty" (observed as a
// Linux-only flake in TestArchDeletedPaths). Background maintenance and the
// fsmonitor daemon are the same class of async .git writer.
//
// The config is injected via GIT_CONFIG_COUNT (git's environment-based config
// mechanism) rather than a config file, so it layers on top of each helper's
// existing GIT_CONFIG_GLOBAL=/dev/null + GIT_CONFIG_NOSYSTEM=1 and is inherited
// by every helper that derives its child env from os.Environ() — present and
// future — without touching each call site. gc.auto=0 disables auto-gc entirely
// (so nothing detaches), maintenance.auto=false disables background
// maintenance, and core.fsmonitor=false ensures no fsmonitor daemon is spawned.
func TestMain(m *testing.M) {
	setenv := map[string]string{
		"GIT_CONFIG_COUNT":   "3",
		"GIT_CONFIG_KEY_0":   "gc.auto",
		"GIT_CONFIG_VALUE_0": "0",
		"GIT_CONFIG_KEY_1":   "maintenance.auto",
		"GIT_CONFIG_VALUE_1": "false",
		"GIT_CONFIG_KEY_2":   "core.fsmonitor",
		"GIT_CONFIG_VALUE_2": "false",
	}
	for k, v := range setenv {
		if err := os.Setenv(k, v); err != nil {
			panic("lifeboat test setup: " + err.Error())
		}
	}
	os.Exit(m.Run())
}

// TestGitAsyncDisabledInTestEnv is the regression guard for TestMain: it proves
// the async-disabling config actually reaches a git command run through the
// package's standard test-repo helper (not just the parent process env). If a
// future change stops threading os.Environ() into the child env, or drops the
// TestMain injection, this fails loudly rather than letting the .git-cleanup
// flake creep back.
func TestGitAsyncDisabledInTestEnv(t *testing.T) {
	r := gvNewRepo(t)
	for key, want := range map[string]string{
		"gc.auto":          "0",
		"maintenance.auto": "false",
		"core.fsmonitor":   "false",
	} {
		got := strings.TrimSpace(r.git("config", "--get", key))
		if got != want {
			t.Errorf("git %s = %q in the test env; want %q (async git op not disabled — .git cleanup can flake)", key, got, want)
		}
	}
}
