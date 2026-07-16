package gitutil

import (
	"os/exec"
	"strings"
	"testing"
)

// TestRunLimitedErrorCarriesStderr: a failing git command's error must carry
// git's own (bounded) stderr, so a probe failure over a hostile or transiently
// unreadable repo explains itself instead of reporting a bare exit status.
func TestRunLimitedErrorCarriesStderr(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir() // not a repo: git fails and writes the reason to stderr
	_, err := RunLimited(dir, 1<<20, "log", "-1")
	if err == nil {
		t.Fatal("RunLimited in a non-repo succeeded, want error")
	}
	if !strings.Contains(err.Error(), "stderr:") || !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("error does not carry git's stderr: %v", err)
	}
}

// TestRunErrorCarriesStderr: same contract for the unlimited Run.
func TestRunErrorCarriesStderr(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	_, err := Run(dir, "log", "-1")
	if err == nil {
		t.Fatal("Run in a non-repo succeeded, want error")
	}
	if !strings.Contains(err.Error(), "stderr:") || !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("error does not carry git's stderr: %v", err)
	}
}
