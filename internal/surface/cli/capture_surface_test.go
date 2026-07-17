package cli

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The three tests below are iss-29's acceptance corpus for the
// "unrecognized-input-never-writes" detector: a mistyped mutating subcommand
// must error without writing, --json errors must be JSON-shaped, and a missing
// config must surface a clean, path-safe error rather than raw Go text.

var reIssueFile = regexp.MustCompile(`^iss-\d+-.*\.md$`)

// ledgerIssueCount walks a repo tree and counts written ledger issue files.
func ledgerIssueCount(t *testing.T, root string) int {
	t.Helper()
	n := 0
	err := filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && reIssueFile.MatchString(d.Name()) {
			n++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return n
}

// TestCaptureTypoSubcommandNeverWrites is the headline: `capture resovle iss-1
// note` (a typo for `resolve`) must be refused with a did-you-mean, and must
// not file a new issue. Before the fix it was swallowed as free text and wrote.
func TestCaptureTypoSubcommandNeverWrites(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	out, err := runCLIErr(t, "capture", "resovle", "iss-1", "clear the flake")
	if err == nil {
		t.Fatalf("expected an error for the mistyped subcommand, got success:\n%s", out)
	}
	if !strings.Contains(err.Error(), "resolve") {
		t.Fatalf("expected a did-you-mean pointing at %q, got: %v", "resolve", err)
	}
	if n := ledgerIssueCount(t, repo); n != 0 {
		t.Fatalf("a mistyped subcommand filed %d issue(s); it must write nothing", n)
	}
}

// TestCaptureFreeTextStillWrites guards the contract: a genuine free-text
// capture whose first word merely resembles a subcommand (but is followed by
// prose, not an iss-id) still files. The typo guard must be high-precision.
func TestCaptureFreeTextStillWrites(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	out := runCLI(t, "capture", "resolved a flaky parser test by widening the timeout", "--json")
	var r struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatalf("capture output not JSON: %v\n%s", err, out)
	}
	if r.ID != "iss-1" {
		t.Fatalf("free-text capture id = %q, want iss-1", r.ID)
	}
	if n := ledgerIssueCount(t, repo); n != 1 {
		t.Fatalf("free-text capture wrote %d issue(s), want 1", n)
	}
}

// TestJSONErrorShapeIsJSON is the --json error-shape contract: when the caller
// asked for --json, a command error is emitted as a JSON envelope, not raw Go
// text. `capture list --json` (no state flag) is a stable erroring case.
func TestJSONErrorShapeIsJSON(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"capture", "list", "--json"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("expected a non-zero exit for `capture list` with no state flag")
	}
	var env struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(stderr.Bytes(), &env); err != nil {
		t.Fatalf("--json error not JSON-shaped: %v\nstderr: %q", err, stderr.String())
	}
	if env.Error == "" {
		t.Fatalf("--json error envelope has an empty message:\n%s", stderr.String())
	}
}

// TestDocsLintMissingConfigCleanError proves the third instance: a missing
// docs-lint config yields a clean, repo-relative diagnostic — never a raw
// os.Open error leaking the absolute config path.
func TestDocsLintMissingConfigCleanError(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"docs", "lint"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("expected a non-zero exit for `docs lint` with no config")
	}
	msg := stderr.String()
	if strings.Contains(msg, repo) {
		t.Fatalf("docs lint error leaked the absolute repo path %q:\n%s", repo, msg)
	}
	if !strings.Contains(msg, filepath.Join(".abcd", "docs-lint.json")) {
		t.Fatalf("docs lint error should name the repo-relative config path:\n%s", msg)
	}

	// And under --json it is JSON-shaped, not raw text.
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"docs", "lint", "--json"}, &stdout, &stderr); code == 0 {
		t.Fatalf("expected a non-zero exit for `docs lint --json` with no config")
	}
	var env struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(stderr.Bytes(), &env); err != nil {
		t.Fatalf("--json docs lint error not JSON-shaped: %v\nstderr: %q", err, stderr.String())
	}
}

// TestCaptureListOpenRendersIssueFields is itd-4 AC5's characterization gate:
// given five open captures, `capture list --open --json` must return all five,
// each carrying the id, slug, severity, and a one-line summary (the captured
// body). It exercises the exact binary path the acceptance criterion names, so
// a regression that dropped any of those fields from the list surface — or that
// failed to enumerate every open issue — would turn this red.
func TestCaptureListOpenRendersIssueFields(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	captures := []struct {
		text string
		sev  string
	}{
		{"issue one about the parser flake", "major"},
		{"issue two about a cache miss", "minor"},
		{"issue three nitpick on spacing", "nitpick"},
		{"issue four critical crash on boot", "critical"},
		{"issue five stray observation", "minor"},
	}
	for _, c := range captures {
		runCLI(t, "capture", c.text, "--severity", c.sev, "--json")
	}

	out := runCLI(t, "capture", "list", "--open", "--json")
	var res struct {
		Issues []struct {
			ID       string `json:"id"`
			Slug     string `json:"slug"`
			Severity string `json:"severity"`
			Body     string `json:"body"`
		} `json:"issues"`
	}
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("list --open --json not JSON: %v\n%s", err, out)
	}
	if len(res.Issues) != len(captures) {
		t.Fatalf("list --open returned %d issues, want %d", len(res.Issues), len(captures))
	}
	// Map by the one-line summary (body) so the assertion is order-independent
	// (list emits derived-priority order, not capture order).
	bySummary := make(map[string]struct{ id, slug, sev string })
	for _, iss := range res.Issues {
		if iss.ID == "" || iss.Slug == "" || iss.Severity == "" || iss.Body == "" {
			t.Fatalf("issue missing an AC5 field: id=%q slug=%q severity=%q body=%q",
				iss.ID, iss.Slug, iss.Severity, iss.Body)
		}
		bySummary[iss.Body] = struct{ id, slug, sev string }{iss.ID, iss.Slug, iss.Severity}
	}
	for _, c := range captures {
		got, ok := bySummary[c.text]
		if !ok {
			t.Fatalf("no listed issue carried the one-line summary %q", c.text)
		}
		if got.sev != c.sev {
			t.Errorf("summary %q: severity = %q, want %q", c.text, got.sev, c.sev)
		}
	}
}

// TestDocsLintUnreadableConfigNoPathLeak covers a non-not-exist load failure
// (the config path is a directory → EISDIR): a *PathError's Error() embeds the
// absolute path, so the branch must strip it. Guards the security-review BLOCK.
func TestDocsLintUnreadableConfigNoPathLeak(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	// Make .abcd/docs-lint.json a directory so os.ReadFile fails with EISDIR.
	if err := os.MkdirAll(filepath.Join(repo, ".abcd", "docs-lint.json"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if code := Run([]string{"docs", "lint"}, &stdout, &stderr); code == 0 {
		t.Fatalf("expected a non-zero exit for an unreadable docs-lint config")
	}
	if msg := stderr.String(); strings.Contains(msg, repo) {
		t.Fatalf("docs lint error leaked the absolute repo path %q:\n%s", repo, msg)
	}

	// Same guarantee under --json.
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"docs", "lint", "--json"}, &stdout, &stderr); code == 0 {
		t.Fatalf("expected a non-zero exit under --json for an unreadable config")
	}
	if msg := stderr.String(); strings.Contains(msg, repo) {
		t.Fatalf("--json docs lint error leaked the absolute repo path %q:\n%s", repo, msg)
	}
}
