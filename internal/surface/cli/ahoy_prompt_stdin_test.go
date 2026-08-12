package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/gittest"
)

// runCLIPipedStdin executes the command tree with stdin bound to a REAL pipe —
// an *os.File that is not a character device, exactly what a host agent's
// `printf 'y\n' | abcd ahoy install` hands the process. A strings.Reader would
// not exercise the same seam: the terminal test is made against the file's
// mode, so only a real pipe proves the non-TTY path (iss-167).
//
// stdout and stderr are captured apart, as the process holds them: the prompts
// and their echoed answers are stderr, so a --json run's stdout stays a clean
// machine-readable envelope even while the questions are being answered.
func runCLIPipedStdin(t *testing.T, stdin string, args ...string) ([]byte, error) {
	t.Helper()
	out, _, err := runCLIPipedStdinSplit(t, stdin, args...)
	return out, err
}

func runCLIPipedStdinSplit(t *testing.T, stdin string, args ...string) (stdout, stderr []byte, err error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close() })
	if _, err := w.WriteString(stdin); err != nil {
		t.Fatal(err)
	}
	// Close the write half so the reader sees EOF once the answers run out,
	// instead of blocking on a pipe that stays open.
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCommand()
	var so, se bytes.Buffer
	cmd.SetOut(&so)
	cmd.SetErr(&se)
	cmd.SetIn(r)
	cmd.SetArgs(args)
	execErr := cmd.Execute()
	return so.Bytes(), se.Bytes(), execErr
}

// TestAhoyInstallAcceptsPipedAnswersFromNonTTYStdin is the iss-167 regression:
// a piped `y` must drive the interactive path. The install runs with NO --yes
// and NO --adopt, so every approval — the unmanaged-repo adoption gate included
// — has to come off stdin. A run that treats a non-TTY stdin as a decline
// adopts nothing and writes nothing.
func TestAhoyInstallAcceptsPipedAnswersFromNonTTYStdin(t *testing.T) {
	repo := hermeticRepo(t)

	// One `y` per question: the adoption gate, then one per gap category.
	answers := strings.Repeat("y\n", 12)
	out, errOut, err := runCLIPipedStdinSplit(t, answers, "ahoy", "install",
		"--visibility", "private", "--docs-target", "both",
		"--oracle-backend", "host-delegated", "--scan-deep", "false", "--json")
	if err != nil {
		t.Fatalf("install exited non-zero: %v\n%s\n%s", err, out, errOut)
	}
	var res struct {
		Status             string   `json:"status"`
		Writes             []string `json:"writes"`
		DeclinedCategories []string `json:"declined_categories"`
	}
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("install output not JSON: %v\n%s", err, out)
	}
	if res.Status == "aborted" {
		t.Fatalf("piped `y` was read as a decline at the adoption gate: status=%q\n%s", res.Status, out)
	}
	if len(res.DeclinedCategories) > 0 {
		t.Fatalf("piped `y` was read as a decline for %v\n%s", res.DeclinedCategories, out)
	}
	if res.Status != "clean" {
		t.Fatalf("install status = %q, want clean\n%s", res.Status, out)
	}
	if len(res.Writes) == 0 {
		t.Fatalf("a fully approved install wrote nothing\n%s", out)
	}
	body, err := os.ReadFile(filepath.Join(repo, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("CLAUDE.md not written by the piped-answer install: %v", err)
	}
	if !strings.Contains(string(body), "<!-- BEGIN ABCD -->") {
		t.Fatalf("CLAUDE.md has no marker block:\n%s", body)
	}
	// Off a terminal nothing echoes the piped answer, so the prompter writes it:
	// the diagnostic stream is a transcript of what was asked and answered.
	if !strings.Contains(string(errOut), "Adopt this unmanaged repo into abcd? [y/N] y") {
		t.Fatalf("the piped answer left no transcript on stderr:\n%s", errOut)
	}
}

// TestAhoyInstallEmptyStdinStillDeclines guards the safe default the non-TTY
// read must not cost: with nothing on stdin, every question reads EOF and EOF
// declines, so an unattended run never adopts a repo it was not told to adopt.
func TestAhoyInstallEmptyStdinStillDeclines(t *testing.T) {
	hermeticRepo(t)

	out, err := runCLIPipedStdin(t, "", "ahoy", "install", "--json")
	if err != nil {
		t.Fatalf("install exited non-zero: %v\n%s", err, out)
	}
	var res struct {
		Status string   `json:"status"`
		Writes []string `json:"writes"`
	}
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("install output not JSON: %v\n%s", err, out)
	}
	if res.Status != "aborted" || len(res.Writes) > 0 {
		t.Fatalf("empty stdin must decline the adoption gate and write nothing, got %+v\n%s", res, out)
	}
}

// gitRepoWithIdentity turns the hermetic repo into a real git repo carrying a
// resolvable commit identity, so the advisory git-identity pin gap is present.
func gitRepoWithIdentity(t *testing.T, repo, name, email string) {
	t.Helper()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		cmd.Env = gittest.Env(t)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.name", name)
	run("config", "user.email", email)
}

// TestAhoyInstallYesDisclosesOptionalIdentityPin is the iss-166 ruling in the
// output: --yes does NOT cover the optional git-identity pin (it would capture
// whatever identity happens to be configured), so a --yes run that reports
// "already up to date" must say the optional pin was excluded and how to apply
// it. Silence there is what made the skip ambiguous.
func TestAhoyInstallYesDisclosesOptionalIdentityPin(t *testing.T) {
	repo := hermeticRepo(t)
	gitRepoWithIdentity(t, repo, "Alex Reppel", "alex@example.com")

	base := []string{"ahoy", "install", "--yes", "--adopt",
		"--visibility", "private", "--docs-target", "both",
		"--oracle-backend", "host-delegated", "--scan-deep", "false"}

	// First --yes run closes every required gap and leaves the optional pin.
	if out, err := runCLIErr(t, base...); err != nil {
		t.Fatalf("install exited non-zero: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(repo, ".abcd", "config", "identity.json")); err == nil {
		t.Fatal("precondition: --yes must not pin the current git identity")
	}

	// Second --yes run: "already up to date", and it must name the exclusion.
	out, err := runCLIErr(t, base...)
	if err != nil {
		t.Fatalf("re-install exited non-zero: %v\n%s", err, out)
	}
	text := string(out)
	if !strings.Contains(text, "already_up_to_date") {
		t.Fatalf("precondition: second --yes run should report already_up_to_date:\n%s", text)
	}
	if !strings.Contains(text, "git_identity.unpinned") {
		t.Fatalf("--yes silently skipped the optional identity pin; output names no exclusion:\n%s", text)
	}
	if !strings.Contains(text, "abcd ahoy install") {
		t.Fatalf("the exclusion notice must say how to apply the optional pin:\n%s", text)
	}

	// The JSON envelope carries the same fact, so a script sees it too.
	jsonOut, err := runCLIErr(t, append(base, "--json")...)
	if err != nil {
		t.Fatalf("re-install --json exited non-zero: %v\n%s", err, jsonOut)
	}
	var res struct {
		Status          string   `json:"status"`
		OptionalSkipped []string `json:"optional_skipped"`
	}
	if err := json.Unmarshal(jsonOut, &res); err != nil {
		t.Fatalf("install output not JSON: %v\n%s", err, jsonOut)
	}
	if len(res.OptionalSkipped) != 1 || res.OptionalSkipped[0] != "git_identity.unpinned" {
		t.Fatalf("optional_skipped = %v, want [git_identity.unpinned]\n%s", res.OptionalSkipped, jsonOut)
	}
}

// TestAhoyInstallYesHelpStatesTheExclusion is the other half of the iss-166
// ruling: the flag's own help says what --yes does not cover.
func TestAhoyInstallYesHelpStatesTheExclusion(t *testing.T) {
	out, err := runCLIErr(t, "ahoy", "install", "--help")
	if err != nil {
		t.Fatalf("--help exited non-zero: %v\n%s", err, out)
	}
	help := string(out)
	i := strings.Index(help, "--yes")
	if i < 0 {
		t.Fatalf("--yes missing from install help:\n%s", help)
	}
	line := help[i:]
	if j := strings.Index(line, "\n"); j >= 0 {
		line = line[:j]
	}
	if !strings.Contains(line, "identity") {
		t.Fatalf("--yes help does not state the optional identity-pin exclusion: %q", line)
	}
}

// TestAhoyInstallPipedAnswerAdoptsOptionalIdentityPin closes the loop between
// the two issues: the way to apply the optional pin without a terminal is the
// piped answer iss-167 makes work.
func TestAhoyInstallPipedAnswerAdoptsOptionalIdentityPin(t *testing.T) {
	repo := hermeticRepo(t)
	gitRepoWithIdentity(t, repo, "Alex Reppel", "alex@example.com")

	if out, err := runCLIErr(t, "ahoy", "install", "--yes", "--adopt",
		"--visibility", "private", "--docs-target", "both",
		"--oracle-backend", "host-delegated", "--scan-deep", "false"); err != nil {
		t.Fatalf("install exited non-zero: %v\n%s", err, out)
	}

	// One `y` per remaining category, not one in total: the approval walk
	// iterates a map, so which category is asked first is not fixed, and a
	// single answer would land on whichever question happened to come up.
	if out, err := runCLIPipedStdin(t, strings.Repeat("y\n", 8), "ahoy", "install"); err != nil {
		t.Fatalf("piped re-install exited non-zero: %v\n%s", err, out)
	}
	body, err := os.ReadFile(filepath.Join(repo, ".abcd", "config", "identity.json"))
	if err != nil {
		t.Fatalf("piped `y` did not adopt the optional identity pin: %v", err)
	}
	if !strings.Contains(string(body), "alex@example.com") {
		t.Fatalf("pin written from the wrong identity:\n%s", body)
	}
}
