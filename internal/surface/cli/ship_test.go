package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/surface"
	"github.com/REPPL/abcd-cli/internal/gittest"
)

// shipFixture is a repository the ship verb can actually cut from: a tagged
// release whose tree carries the surface baseline, and whose baseline is the
// surface THIS binary walks — because the guardrail refuses to compare a
// snapshot it cannot prove came from the tree being released.
//
// It is built through the front door's own SurfaceSnapshot, so the fixture and
// the verb agree on the surface by construction rather than by a hand-written
// copy that would rot the moment a command is added.
func shipFixture(t *testing.T) *gittest.Repo {
	t.Helper()
	r := gittest.NewRepo(t)
	r.Write(".claude-plugin/plugin.json", `{"name":"abcd","description":"fixture"}`+"\n")
	r.Write(".claude-plugin/marketplace.json", `{"name":"abcd","plugins":[{"name":"abcd","source":"./"}]}`+"\n")
	// The empty [Unreleased] heading is the post-cutover state (outcome 7) and the
	// anchor the ingest step inserts beneath.
	r.Write("CHANGELOG.md", "# Changelog\n\n## [Unreleased]\n\n## [0.4.0] - 2026-07-01\n\n### Added\n\n- the base.\n")

	live, err := SurfaceSnapshot(r.Root())
	if err != nil {
		t.Fatalf("SurfaceSnapshot: %v", err)
	}
	data, err := surface.Encode(live)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	r.Write(SurfaceSnapshotPath, string(data))
	r.Commit("the released state")
	r.Git("tag", "v0.4.0")
	return r
}

// shipIn runs the CLI with the working directory pointed at the fixture, the way
// the verb is really invoked.
func shipIn(t *testing.T, r *gittest.Repo, args ...string) ([]byte, error) {
	t.Helper()
	t.Chdir(r.Root())
	return runCLIErr(t, args...)
}

// TestLaunchShipEmitsAReadyCut is the wired path: `abcd launch ship` with no
// payload flag runs the deterministic emit step and exits 0 on a cut that may
// proceed.
func TestLaunchShipEmitsAReadyCut(t *testing.T) {
	r := shipFixture(t)
	r.Write(".abcd/development/intents/shipped/itd-73-derived-versioning.md",
		"---\nid: itd-73\nimpact: additive\n---\n\n# A Version Is A Fact\n\nthe version is derived.\n")
	r.Commit("ship an intent")

	out, err := shipIn(t, r, "launch", "ship")
	if code := exitCodeOf(err); code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	for _, want := range []string{"v0.4.1", "additive", "itd-73", "A Version Is A Fact"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("render does not mention %q:\n%s", want, out)
		}
	}
}

// TestLaunchShipRefusalExits1 pins the exit contract: a refusal is a REPORT, not
// a crash — the whole cut renders, the exit code is 1, and the blocking record
// is named. Exit 2 is reserved for a structural fault.
func TestLaunchShipRefusalExits1(t *testing.T) {
	r := shipFixture(t)
	r.Write(".abcd/development/intents/shipped/itd-73-x.md", "---\nid: itd-73\nimpact: additive\n---\n# x\n")
	r.Write(".abcd/development/intents/planned/itd-94-gate.md",
		"---\nid: itd-94\nkind: standalone\nspec_id: spc-9\n---\n# gate\n")
	r.Write(".abcd/development/specs/closed/spc-9-gate.md",
		"---\nid: spc-9\nslug: gate\nintent: itd-94\n---\n# spc-9\n")
	r.Commit("a merged feature whose intent never left planned/")

	out, err := shipIn(t, r, "launch", "ship")
	if code := exitCodeOf(err); code != 1 {
		t.Fatalf("exit = %d, want 1\n%s", code, out)
	}
	for _, want := range []string{"REFUSED", "stale-intent", "itd-94"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("refusal render does not mention %q:\n%s", want, out)
		}
	}
}

// TestLaunchShipRefusesReleaseInFlight is outcome 1 wired end to end: the newest
// CHANGELOG heading is ahead of the newest tag, so a release sits between its
// merge and its tag and the verb refuses rather than deriving against a
// mismatched base.
func TestLaunchShipRefusesReleaseInFlight(t *testing.T) {
	r := shipFixture(t)
	r.Write("CHANGELOG.md", "# Changelog\n\n## [0.5.0] - 2026-07-20\n\n### Added\n\n- the ship PR merged.\n")
	r.Write(".abcd/development/intents/shipped/itd-73-x.md", "---\nid: itd-73\nimpact: additive\n---\n# x\n")
	r.Commit("the ship PR merged; auto-release has not tagged it yet")

	out, err := shipIn(t, r, "launch", "ship")
	if code := exitCodeOf(err); code != 1 {
		t.Fatalf("exit = %d, want 1\n%s", code, out)
	}
	for _, want := range []string{"release-in-flight", "v0.5.0"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("refusal render does not mention %q:\n%s", want, out)
		}
	}
}

// composedPayload writes a well-formed composer payload citing exactly ids.
func composedPayload(t *testing.T, dir, nextTag string, ids ...string) string {
	t.Helper()
	entries := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		entries = append(entries, map[string]any{
			"section": "Added",
			"records": []string{id},
			"text":    "**Something shipped.** " + id + " landed.",
		})
	}
	data, err := json.Marshal(map[string]any{
		"schema_version": 1,
		"prompt_version": "1.0.0",
		"next_tag":       nextTag,
		"entries":        entries,
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "changelog.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// shipReadyRepo is a fixture whose cut is ready with one required record.
func shipReadyRepo(t *testing.T) *gittest.Repo {
	t.Helper()
	r := shipFixture(t)
	r.Write(".abcd/development/intents/shipped/itd-73-derived-versioning.md",
		"---\nid: itd-73\nimpact: additive\n---\n\n# A Version Is A Fact\n\nthe version is derived.\n")
	r.Commit("ship an intent")
	return r
}

// TestLaunchShipIngestWritesTheHeading is the wired write path: a payload whose
// citations match the cut lands the dated section and exits 0.
func TestLaunchShipIngestWritesTheHeading(t *testing.T) {
	r := shipReadyRepo(t)
	payload := composedPayload(t, t.TempDir(), "v0.4.1", "itd-73")

	out, err := shipIn(t, r, "launch", "ship", "--changelog-json", payload)
	if code := exitCodeOf(err); code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	for _, want := range []string{"wrote:", "CHANGELOG.md", "## [0.4.1] - "} {
		if !strings.Contains(string(out), want) {
			t.Errorf("render does not mention %q:\n%s", want, out)
		}
	}
	data, err := os.ReadFile(filepath.Join(r.Root(), "CHANGELOG.md"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, "### Added\n\n- **Something shipped.** itd-73 landed. (itd-73)\n") {
		t.Errorf("the composed line did not land:\n%s", got)
	}
	if !strings.Contains(got, "## [Unreleased]\n\n## [0.4.1] - ") {
		t.Errorf("the section was not inserted beneath an empty [Unreleased]:\n%s", got)
	}
}

// TestLaunchShipIngestReadsStdin pins the `-` operand: an orchestrating command
// pipes the composer's output straight in rather than staging a temp file.
func TestLaunchShipIngestReadsStdin(t *testing.T) {
	r := shipReadyRepo(t)
	data, err := os.ReadFile(composedPayload(t, t.TempDir(), "v0.4.1", "itd-73"))
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(r.Root())
	out, err := runCLIStdinErr(t, string(data), "launch", "ship", "--changelog-json", "-")
	if code := exitCodeOf(err); code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(string(out), "## [0.4.1] - ") {
		t.Errorf("render does not report the written heading:\n%s", out)
	}
}

// TestLaunchShipIngestBijectionExits2 is the loud stage at the front door: a
// composed changelog that omits a shipped record is a structural fault, the
// whole document is refused, and CHANGELOG.md is byte-identical afterwards.
func TestLaunchShipIngestBijectionExits2(t *testing.T) {
	r := shipReadyRepo(t)
	r.Write(".abcd/work/issues/resolved/iss-51-crash.md", "---\nid: iss-51\nimpact: fix\n---\n# x\n")
	r.Commit("resolve an issue too")
	payload := composedPayload(t, t.TempDir(), "v0.4.1", "itd-73")

	before := cliTreeDigest(t, r.Root())
	out, err := shipIn(t, r, "launch", "ship", "--changelog-json", payload)
	if code := exitCodeOf(err); code != 2 {
		t.Fatalf("exit = %d, want 2\n%s", code, out)
	}
	for _, want := range []string{"MISSING", "iss-51", "nothing was written"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err.Error(), want)
		}
	}
	if after := cliTreeDigest(t, r.Root()); after != before {
		t.Error("a refused ingest changed the working tree")
	}
}

// TestLaunchShipIngestOnARefusedCutExits1 keeps the two failure modes apart: the
// CUT refusing is a report (exit 1), not a payload fault (exit 2) — and it must
// not write, whatever the payload says.
func TestLaunchShipIngestOnARefusedCutExits1(t *testing.T) {
	r := shipFixture(t)
	r.Commit("nothing shipped")
	payload := composedPayload(t, t.TempDir(), "v0.4.1", "itd-73")

	before := cliTreeDigest(t, r.Root())
	out, err := shipIn(t, r, "launch", "ship", "--changelog-json", payload)
	if code := exitCodeOf(err); code != 1 {
		t.Fatalf("exit = %d, want 1\n%s", code, out)
	}
	if !strings.Contains(string(out), "empty-cut") {
		t.Errorf("the refusal report is missing:\n%s", out)
	}
	if after := cliTreeDigest(t, r.Root()); after != before {
		t.Error("a refused cut changed the working tree")
	}
}

// TestChangelogPreviewWritesNothing is deliverable 4's contract, asserted the
// only way that means anything: hash the whole tree before and after.
func TestChangelogPreviewWritesNothing(t *testing.T) {
	r := shipFixture(t)
	r.Write(".abcd/development/intents/shipped/itd-73-x.md", "---\nid: itd-73\nimpact: additive\n---\n# x\n")
	r.Commit("ship an intent")

	before := cliTreeDigest(t, r.Root())
	out, err := shipIn(t, r, "changelog")
	if code := exitCodeOf(err); code != 0 {
		t.Fatalf("exit = %d, want 0 (a preview always reports)\n%s", code, out)
	}
	if after := cliTreeDigest(t, r.Root()); after != before {
		t.Errorf("the preview changed the working tree:\nbefore %s\nafter  %s", before, after)
	}
}

// TestChangelogPreviewRefusalStillExitsZero separates the preview from the gate:
// `abcd changelog` REPORTS a refused cut (like `launch --dry-run`), while
// `launch ship` exits non-zero on the same repository.
func TestChangelogPreviewRefusalStillExitsZero(t *testing.T) {
	r := shipFixture(t)
	r.Commit("nothing shipped at all")

	out, err := shipIn(t, r, "changelog")
	if code := exitCodeOf(err); code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(string(out), "empty-cut") {
		t.Errorf("preview does not name the refusal:\n%s", out)
	}

	if _, err := shipIn(t, r, "launch", "ship"); exitCodeOf(err) != 1 {
		t.Errorf("launch ship exit = %d on the same repo, want 1", exitCodeOf(err))
	}
}

// TestChangelogPreviewJSON pins the machine surface the next stages read.
func TestChangelogPreviewJSON(t *testing.T) {
	r := shipFixture(t)
	r.Write(".abcd/development/intents/shipped/itd-73-x.md",
		"---\nid: itd-73\nimpact: additive\n---\n\n# Title\n\nsummary.\n")
	r.Commit("ship an intent")

	out, err := shipIn(t, r, "changelog", "--json")
	if code := exitCodeOf(err); code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	var got struct {
		Ready   bool   `json:"ready"`
		NextTag string `json:"next_tag"`
		Added   []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"added"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("changelog --json is not JSON: %v\n%s", err, out)
	}
	if !got.Ready || got.NextTag != "v0.4.1" {
		t.Errorf("ready=%v next_tag=%q, want true v0.4.1", got.Ready, got.NextTag)
	}
	if len(got.Added) != 1 || got.Added[0].ID != "itd-73" || got.Added[0].Title != "Title" {
		t.Errorf("added = %+v, want the one record with its title", got.Added)
	}
}

// TestLaunchShipWritesNothingYet pins that the emit stage of the verb is as
// read-only as the preview. The write path arrives with the ingest step; until
// then a ship that ran and a ship that did not are indistinguishable on disk.
func TestLaunchShipWritesNothingYet(t *testing.T) {
	r := shipFixture(t)
	r.Write(".abcd/development/intents/shipped/itd-73-x.md", "---\nid: itd-73\nimpact: additive\n---\n# x\n")
	r.Commit("ship an intent")

	before := cliTreeDigest(t, r.Root())
	if _, err := shipIn(t, r, "launch", "ship"); exitCodeOf(err) != 0 {
		t.Fatalf("launch ship failed unexpectedly")
	}
	if after := cliTreeDigest(t, r.Root()); after != before {
		t.Errorf("the emit step changed the working tree:\nbefore %s\nafter  %s", before, after)
	}
}

// TestShipStructuralFaultExits2 pins the third exit code. A directory that is
// not an abcd repository cannot be read at all, which is a fault rather than a
// refusal — and the diagnostic must be path-scrubbed, because it names files
// under the caller's working directory.
func TestShipStructuralFaultExits2(t *testing.T) {
	for _, args := range [][]string{{"launch", "ship"}, {"changelog"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Chdir(t.TempDir())
			out, err := runCLIErr(t, args...)
			if code := exitCodeOf(err); code != 2 {
				t.Fatalf("exit = %d, want 2\n%s", code, out)
			}
			if strings.Contains(err.Error(), os.TempDir()) {
				t.Errorf("diagnostic leaks an absolute path: %q", err.Error())
			}
		})
	}
}

// TestLaunchStillRunsWithoutASubcommand is the regression guard on hanging
// `ship` off the launch command: a cobra parent that gains a subcommand can stop
// running its own RunE, which would silently turn `abcd launch` into a usage
// dump. It must still reach its own refusal.
func TestLaunchStillRunsWithoutASubcommand(t *testing.T) {
	r := shipFixture(t)
	out, err := shipIn(t, r, "launch")
	if err == nil {
		t.Fatalf("bare `abcd launch` must still refuse without --dry-run\n%s", out)
	}
	if !strings.Contains(err.Error(), "pass --dry-run") {
		t.Errorf("error = %q, want the launch command's own refusal", err.Error())
	}
}

// cliTreeDigest hashes every path and byte under root except .git, whose
// internals git rewrites for reasons unrelated to the code under test.
func cliTreeDigest(t *testing.T, root string) string {
	t.Helper()
	var lines []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return fs.SkipDir
			}
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		lines = append(lines, filepath.ToSlash(rel)+" "+hex.EncodeToString(sum[:]))
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:])
}
