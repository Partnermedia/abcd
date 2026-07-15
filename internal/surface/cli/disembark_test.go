package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// TestDisembarkCoverageRejectsNonReport holds the schema contract: a JSON object
// that is not a probe report (schema_version defaults to 0 on any arbitrary
// object) must be rejected with exit 2, not silently aggregated as an all-blank
// phantom repo (B38).
func TestDisembarkCoverageRejectsNonReport(t *testing.T) {
	f := filepath.Join(t.TempDir(), "pkg.json")
	if err := os.WriteFile(f, []byte(`{"name":"not-a-report","version":"1.2.3"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"disembark", "coverage", f}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit = %d, want 2 (a non-report must fail)\nstdout:%s\nstderr:%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "not a coverage report") {
		t.Errorf("want a 'not a coverage report' diagnostic, got stderr:\n%s", stderr.String())
	}
}

// TestDisembarkPlanEmitsManifestJSON proves the plan verb is wired to the core
// and returns a manifest: a schema-stamped list of destination-relative paths
// and the pinned hash, over a repo the packer would later write.
func TestDisembarkPlanEmitsManifestJSON(t *testing.T) {
	repo := probeRepo(t)
	out := runCLI(t, "disembark", "plan", repo, "--json")
	var m lifeboat.PlanManifest
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("plan --json is not a manifest: %v", err)
	}
	if m.SchemaVersion != lifeboat.SchemaVersion {
		t.Errorf("schema_version = %d, want %d", m.SchemaVersion, lifeboat.SchemaVersion)
	}
	if m.FileCount != len(m.Files) || m.FileCount == 0 {
		t.Errorf("file_count = %d, len(files) = %d", m.FileCount, len(m.Files))
	}
	if len(m.ManifestSHA256) != 64 {
		t.Errorf("manifest_sha256 = %q, want a 64-hex-char digest", m.ManifestSHA256)
	}
	// The provenance marker is always part of the plan.
	found := false
	for _, f := range m.Files {
		if f.Path == lifeboat.ProvenanceName {
			found = true
		}
		if filepath.IsAbs(f.Path) || strings.Contains(f.Path, "..") {
			t.Errorf("manifest path is not destination-safe: %q", f.Path)
		}
	}
	if !found {
		t.Errorf("manifest omits %s", lifeboat.ProvenanceName)
	}
}

// TestDisembarkPlanWritesNothing is the dry-run contract at the surface: running
// the verb leaves the target repository byte-for-byte unchanged.
func TestDisembarkPlanWritesNothing(t *testing.T) {
	repo := probeRepo(t)
	before := dirFingerprint(t, repo)
	runCLI(t, "disembark", "plan", repo)
	if after := dirFingerprint(t, repo); after != before {
		t.Error("disembark plan mutated the target repository")
	}
}

// TestDisembarkPlanRejectsNonDirectory mirrors probe's input contract.
func TestDisembarkPlanRejectsNonDirectory(t *testing.T) {
	f := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLIErr(t, "disembark", "plan", f); err == nil {
		t.Error("planning a file must fail, got nil error")
	}
}

// TestDisembarkPackWritesLifeboat proves the pack verb is wired end to end: it
// writes a lifeboat at <dest>, leaves the source byte-identical, and returns a
// result carrying the manifest hash.
func TestDisembarkPackWritesLifeboat(t *testing.T) {
	repo := probeRepo(t)
	t.Setenv("HOME", t.TempDir())
	dest := filepath.Join(t.TempDir(), "lifeboat")
	before := dirFingerprint(t, repo)

	out := runCLI(t, "disembark", "pack", repo, dest, "--json")

	var res lifeboat.PackResult
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("pack --json is not a result: %v", err)
	}
	if len(res.ManifestSHA256) != 64 || res.FilesWritten == 0 {
		t.Errorf("unexpected pack result: %+v", res)
	}
	if _, err := os.Stat(filepath.Join(dest, lifeboat.ProvenanceName)); err != nil {
		t.Errorf("lifeboat has no %s: %v", lifeboat.ProvenanceName, err)
	}
	if after := dirFingerprint(t, repo); after != before {
		t.Error("disembark pack mutated the source repository")
	}
}

// TestDisembarkPackRefusesNonEmptyDest holds the destination safety gate at the
// surface: packing over a non-empty non-lifeboat directory exits non-zero.
func TestDisembarkPackRefusesNonEmptyDest(t *testing.T) {
	repo := probeRepo(t)
	t.Setenv("HOME", t.TempDir())
	dest := t.TempDir()
	if err := os.WriteFile(filepath.Join(dest, "mine.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLIErr(t, "disembark", "pack", repo, dest); err == nil {
		t.Error("packing over a non-empty non-lifeboat dir must fail")
	}
	if _, err := os.Stat(filepath.Join(dest, "mine.txt")); err != nil {
		t.Errorf("refused pack destroyed the pre-existing directory: %v", err)
	}
}

// TestDisembarkPackRejectsNonDirectorySource mirrors the other verbs' input
// contract on the source argument.
func TestDisembarkPackRejectsNonDirectorySource(t *testing.T) {
	f := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLIErr(t, "disembark", "pack", f, filepath.Join(t.TempDir(), "lb")); err == nil {
		t.Error("packing a file as the source must fail")
	}
}

// dirFingerprint is a cheap path+size fingerprint of a tree, excluding .git.
func dirFingerprint(t *testing.T, root string) string {
	t.Helper()
	var acc strings.Builder
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		if rel == ".git" {
			return filepath.SkipDir
		}
		fmt.Fprintf(&acc, "%s\x00%d\x00%s\n", rel, info.Size(), info.Mode())
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return acc.String()
}
