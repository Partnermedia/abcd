package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/lifeboat"
)

// probeRepo builds an isolated one-commit git repo for the disembark verbs to
// read. It is a Tier-0 instrument: git only, no README, no .abcd.
func probeRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	repo := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"commit", "-q", "--allow-empty", "-m", "root"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(),
			"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@e",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@e",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	return repo
}

// TestDisembarkProbeEmitsFullCoverageJSON proves the verb is wired to the core
// and returns a schema-stamped report naming every brief section.
func TestDisembarkProbeEmitsFullCoverageJSON(t *testing.T) {
	repo := probeRepo(t)
	out := runCLI(t, "disembark", "probe", repo, "--json")
	var cov lifeboat.Coverage
	if err := json.Unmarshal(out, &cov); err != nil {
		t.Fatalf("probe --json is not a coverage report: %v", err)
	}
	if cov.SchemaVersion != lifeboat.SchemaVersion {
		t.Errorf("schema_version = %d, want %d", cov.SchemaVersion, lifeboat.SchemaVersion)
	}
	if len(cov.Sections) != len(lifeboat.Table) {
		t.Errorf("report has %d sections, want %d", len(cov.Sections), len(lifeboat.Table))
	}
	if len(cov.TiersPresent) != 1 || cov.TiersPresent[0] != lifeboat.TierGit {
		t.Errorf("tiers_present = %v, want [git]", cov.TiersPresent)
	}
}

// TestDisembarkProbeRejectsNonDirectory holds the input contract: a path that is
// not a directory exits non-zero with a clean message, not a stack trace.
func TestDisembarkProbeRejectsNonDirectory(t *testing.T) {
	f := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLIErr(t, "disembark", "probe", f); err == nil {
		t.Error("probing a file must fail, got nil error")
	}
}

// TestDisembarkCoverageAggregatesReports proves the cross-repo readout: two
// probe reports on disk reduce to one section×repo table with both columns.
func TestDisembarkCoverageAggregatesReports(t *testing.T) {
	repo := probeRepo(t)
	dir := t.TempDir()
	a := filepath.Join(dir, "a.json")
	b := filepath.Join(dir, "b.json")
	for _, p := range []string{a, b} {
		out := runCLI(t, "disembark", "probe", repo, "--json")
		if err := os.WriteFile(p, out, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	out := runCLI(t, "disembark", "coverage", a, b, "--json")
	var agg lifeboat.AggregateReport
	if err := json.Unmarshal(out, &agg); err != nil {
		t.Fatalf("coverage --json is not an aggregate: %v", err)
	}
	if len(agg.Repos) != 2 {
		t.Errorf("aggregate has %d repos, want 2", len(agg.Repos))
	}
	if len(agg.Sections) != len(lifeboat.Table) {
		t.Errorf("aggregate has %d rows, want %d", len(agg.Sections), len(lifeboat.Table))
	}
}

// TestDisembarkCoverageRejectsUnreadable holds the error path: a missing report
// file exits non-zero, and the message references the path the user typed
// without leaking an absolute filesystem path.
func TestDisembarkCoverageRejectsUnreadable(t *testing.T) {
	out, err := runCLIErr(t, "disembark", "coverage", "no-such-file.json")
	if err == nil {
		t.Fatal("a missing report must fail, got nil error")
	}
	if strings.Contains(string(out), string(os.PathSeparator)+"no-such-file.json") {
		t.Errorf("error leaked an absolute path: %q", out)
	}
}
