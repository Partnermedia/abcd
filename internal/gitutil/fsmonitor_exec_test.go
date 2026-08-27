package gitutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// TestCheckIgnoredDoesNotExecuteRepoLocalFsmonitor is the GHSA-h2gm-w3hm-8xpq
// regression. A clone's own .git/config is fully trusted by git and cannot be
// disabled by environment, so a hostile repo-local core.fsmonitor is a command
// git will spawn to refresh the index. CheckIgnored reads the index (no
// --no-index), so unless it forces core.fsmonitor=false the probe EXECUTES that
// command against an untrusted tree.
//
// The test sets core.fsmonitor to a script that drops a sentinel file, drives
// CheckIgnored against that repo, and asserts the sentinel never appears. On
// unfixed code (probe built without the isolatedGit pins) the sentinel is
// created and the test fails.
func TestCheckIgnoredDoesNotExecuteRepoLocalFsmonitor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX sentinel script; core.fsmonitor exec path exercised on unix")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}

	// Hermetic HOME/XDG so a developer's ~/.gitconfig cannot alter the result;
	// gitEnv() (the production isolation) supplies the git child environment.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
	t.Setenv("GIT_TERMINAL_PROMPT", "0")

	repo := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		cmd.Env = gitEnv()
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}

	git("init", "-q")

	// A tracked file so check-ignore has an index to refresh (the code path that
	// consults core.fsmonitor).
	if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "tracked.txt")
	git("-c", "user.email=t@t", "-c", "user.name=t", "commit", "-q", "-m", "seed")

	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte("ignored/\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// The hostile payload: a repo-local core.fsmonitor command that, if git
	// spawns it, writes a sentinel. touch-only, exactly per the advisory PoC.
	sentinel := filepath.Join(repo, "fsmonitor-fired")
	script := filepath.Join(repo, "evil-fsmonitor.sh")
	body := "#!/bin/sh\ntouch " + sentinel + "\nexit 0\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	git("config", "core.fsmonitor", script)

	// Drive the live probe against the hostile tree.
	_ = CheckIgnored(repo, []string{"ignored/a.txt", "kept/b.txt"})

	if _, err := os.Stat(sentinel); err == nil {
		t.Fatalf("core.fsmonitor executed: CheckIgnored spawned the repo-local command and it wrote %s; the isolatedGit exec pins are missing", sentinel)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat sentinel: %v", err)
	}
}
