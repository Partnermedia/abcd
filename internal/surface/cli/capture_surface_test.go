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
	if !regexp.MustCompile(`^iss-[0-9]{16}$`).MatchString(r.ID) {
		t.Fatalf("free-text capture id = %q, want a native timestamp-numeric id", r.ID)
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

// TestDocsLintEngineFaultExitsTwo pins that `docs lint` returns exit 2 — "could
// not be evaluated" — when the engine cannot run (here, a config whose root
// escapes the repository), the same tri-state code `abcd lint` and record-lint
// return. Exit 1 is reserved for a blocker finding, so a CI gate keying on >=2
// must not read a lint that never ran as an ordinary findings-pass.
func TestDocsLintEngineFaultExitsTwo(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	cfgDir := filepath.Join(repo, ".abcd")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// A root that escapes the repository is a containment refusal the engine
	// raises before it can scan anything — an engine fault, not a finding.
	if err := os.WriteFile(filepath.Join(cfgDir, "docs-lint.json"),
		[]byte(`{"roots":["../outside"]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{"docs", "lint"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("docs lint engine fault exit = %d, want 2\nstderr: %s", code, stderr.String())
	}
	if strings.Contains(stderr.String(), repo) {
		t.Errorf("engine fault message leaked the absolute repo path:\n%s", stderr.String())
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
	// (list emits derived-priority order, not capture order). The stored body now
	// carries exactly one trailing newline (iss-175: the writer terminates every
	// record), which the raw body field reflects — so key on the trimmed summary.
	bySummary := make(map[string]struct{ id, slug, sev string })
	for _, iss := range res.Issues {
		if iss.ID == "" || iss.Slug == "" || iss.Severity == "" || iss.Body == "" {
			t.Fatalf("issue missing an AC5 field: id=%q slug=%q severity=%q body=%q",
				iss.ID, iss.Slug, iss.Severity, iss.Body)
		}
		bySummary[strings.TrimRight(iss.Body, "\n")] = struct{ id, slug, sev string }{iss.ID, iss.Slug, iss.Severity}
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

// TestCapturePromoteJSONContract is the spc-24 surface AC: `capture promote
// <iss-N> --json` reports the issue id, the minted intent id, and both
// repo-relative paths; the minted draft and the stamped issue exist on disk.
func TestCapturePromoteJSONContract(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	capOut := runCLI(t, "capture", "the parser eats trailing newlines", "--json")
	var minted struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(capOut, &minted); err != nil || minted.ID == "" {
		t.Fatalf("capture envelope unreadable: %v\n%s", err, capOut)
	}

	out := runCLI(t, "capture", "promote", minted.ID, "--json")
	var r struct {
		IssueID    string `json:"issue_id"`
		IssuePath  string `json:"issue_path"`
		IntentID   string `json:"intent_id"`
		IntentPath string `json:"intent_path"`
		Linked     bool   `json:"linked"`
	}
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatalf("promote output not JSON: %v\n%s", err, out)
	}
	if r.IssueID != minted.ID || r.IntentID != "itd-1" || r.Linked {
		t.Fatalf("unexpected promote result: %+v", r)
	}
	for _, p := range []string{r.IssuePath, r.IntentPath} {
		if p == "" || filepath.IsAbs(p) {
			t.Fatalf("promote paths must be repo-relative and non-empty: %+v", r)
		}
		if _, err := os.Stat(filepath.Join(repo, p)); err != nil {
			t.Fatalf("promote-reported path %s missing: %v", p, err)
		}
	}

	// Second promote refuses (exit non-zero) and names the existing intent.
	if _, err := runCLIErr(t, "capture", "promote", minted.ID); err == nil || !strings.Contains(err.Error(), "itd-1") {
		t.Fatalf("second promote must refuse naming itd-1, got: %v", err)
	}
}

// TestCaptureResolveProvenanceJSON is the spc-25 surface AC: resolve with
// provenance flags reports the written resolved_by members in --json, and an
// unknown id refuses without writing.
func TestCaptureResolveProvenanceJSON(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	capOut := runCLI(t, "capture", "a provenance-carrying issue", "--json")
	var minted struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(capOut, &minted); err != nil || minted.ID == "" {
		t.Fatalf("capture envelope unreadable: %v\n%s", err, capOut)
	}
	intentRel := ".abcd/development/intents/shipped/itd-4-the-fixer.md"
	if err := os.MkdirAll(filepath.Join(repo, filepath.Dir(intentRel)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, intentRel), []byte("---\nid: itd-4\n---\n\nx\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Unknown intent refuses; the issue stays open.
	if _, err := runCLIErr(t, "capture", "resolve", minted.ID, "nope", "--impact", "fix", "--intent", "itd-99"); err == nil {
		t.Fatalf("resolve with an unknown --intent must refuse")
	}

	out := runCLI(t, "capture", "resolve", minted.ID, "fixed by the fixer", "--impact", "fix",
		"--intent", "itd-4", "--commit", "abcdef0123", "--json")
	var r struct {
		ID         string `json:"id"`
		ToStatus   string `json:"to_status"`
		ResolvedBy *struct {
			Intent string `json:"intent"`
			Spec   string `json:"spec"`
			Commit string `json:"commit"`
		} `json:"resolved_by"`
	}
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatalf("resolve output not JSON: %v\n%s", err, out)
	}
	if r.ID != minted.ID || r.ToStatus != "resolved" {
		t.Fatalf("unexpected resolve result: %+v", r)
	}
	if r.ResolvedBy == nil || r.ResolvedBy.Intent != "itd-4" || r.ResolvedBy.Commit != "abcdef0123" || r.ResolvedBy.Spec != "" {
		t.Fatalf("resolved_by members wrong: %+v", r.ResolvedBy)
	}
}

// iss-2608221328552172: the same lone-token hole on the capture create path.
// `abcd capture nosuchthing` is not near any sub-verb, so the did-you-mean
// guard never fired and the token was filed as an issue at exit 0.

// TestCaptureLoneWordNeverWrites is the headline: a single bare word must be
// refused, and must file no issue.
func TestCaptureLoneWordNeverWrites(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	out, err := runCLIErr(t, "capture", "nosuchthing")
	if err == nil {
		t.Fatalf("expected an error for the lone unknown token, got success:\n%s", out)
	}
	if n := ledgerIssueCount(t, repo); n != 0 {
		t.Fatalf("a lone unknown token filed %d issue(s); it must write nothing", n)
	}
}

// TestCaptureLoneRecordIDNeverWrites: `abcd capture iss-1` (a forgotten
// sub-verb) must not become an issue titled with a record id.
func TestCaptureLoneRecordIDNeverWrites(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	out, err := runCLIErr(t, "capture", "iss-1")
	if err == nil {
		t.Fatalf("expected an error for the lone record id, got success:\n%s", out)
	}
	if n := ledgerIssueCount(t, repo); n != 0 {
		t.Fatalf("a lone record id filed %d issue(s); it must write nothing", n)
	}
}

// TestCaptureEmptyTextNeverWrites: `abcd capture ""` — a script whose variable
// expanded to nothing — is the degenerate lone token. Core already refused it
// ("slug normalises to empty", exit 1); the guard now names it for what it is,
// at the usage code. Either way nothing is written, which is the part that binds.
func TestCaptureEmptyTextNeverWrites(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	out, err := runCLIErr(t, "capture", "")
	if err == nil {
		t.Fatalf("expected an error for empty capture text, got success:\n%s", out)
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Fatalf("expected the error to name the text as empty, got: %v", err)
	}
	if n := ledgerIssueCount(t, repo); n != 0 {
		t.Fatalf("empty capture text filed %d issue(s); it must write nothing", n)
	}
}

// TestCaptureDerivedSlugNeverCarriesHomePathUsername is the regression for
// GH #485 (reported by @jogrun): `abcd capture <text>` derived the issue slug
// from the RAW text before redaction ran, so a /Users/<name>/… home path in the abcd-audit:allow
// capture text put the caller's username into the committed issue filename even
// though the body was redacted. deriveSlug kebab-cased the path first, so nothing
// left looked like a path and the ledger redactor never saw it — the body was
// clean but open/iss-N-users-alice-….md still carried the name. Redaction must
// run BEFORE the slug is derived, so a finding can never reach the filename.
func TestCaptureDerivedSlugNeverCarriesHomePathUsername(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	// A fake home path — deliberately NOT the caller's own HOME — so the
	// assertion turns on the ordering of redaction vs slug derivation, not on the
	// test host's identity. The generic /Users|/home matcher (home_path_other)
	// redacts it either way.
	const username = "alice"
	body := "the key at /Users/" + username + "/.ssh/id_rsa leaked into the log" // abcd-audit:allow

	out := runCLI(t, "capture", body, "--json")
	var r struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
		Path string `json:"path"`
	}
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatalf("capture output not JSON: %v\n%s", err, out)
	}
	if strings.Contains(r.Slug, username) {
		t.Fatalf("derived slug %q carries the home-path username %q", r.Slug, username)
	}
	if strings.Contains(r.Path, username) {
		t.Fatalf("reported path %q carries the home-path username %q", r.Path, username)
	}
	// And no committed issue filename on disk carries it either.
	if err := filepath.WalkDir(repo, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && reIssueFile.MatchString(d.Name()) && strings.Contains(d.Name(), username) {
			t.Fatalf("committed issue filename %q carries the home-path username %q", d.Name(), username)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk %s: %v", repo, err)
	}
}

// TestCaptureStatusBoardRendersSkipped (iss-2608261437041050): a ledger record
// the reader refuses is counted by none of the board's three totals, so a board
// that printed the totals alone would report a ledger smaller than the one on
// disk and say nothing about the record it dropped. `capture list` already
// renders the skipped roster; the bare status board must not undercount in
// silence.
func TestCaptureStatusBoardRendersSkipped(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	// One well-formed capture, so the ledger dirs exist and the board has a total.
	runCLI(t, "capture", "a well formed observation", "--slug", "fine", "--json")
	// A committed record missing a required property: the reader validates before
	// it reads, so this one is skipped rather than counted.
	stripped := filepath.Join(repo, ".abcd", "work", "issues", "open", "iss-900-stripped.md")
	if err := os.WriteFile(stripped, []byte(
		"---\nid: iss-900\nslug: stripped\nseverity: minor\ncategory: bug\nsource: user-observation\nfound_during: t\n---\n\nan issue\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	board := string(runCLI(t, "capture"))
	if !strings.Contains(board, "skipped") || !strings.Contains(board, "iss-900-stripped.md") {
		t.Fatalf("the status board must render the skipped roster:\n%s", board)
	}
}

// writeReadingFixture lays down one committed reading record by hand, so the
// disposition verb has an item to answer. Reading records are written by the
// ingest path, which is the cold-reading output contract's front door, not this
// surface's — a fixture is the honest stand-in here.
func writeReadingFixture(t *testing.T, repo, run, item string) {
	t.Helper()
	dir := filepath.Join(repo, ".abcd", "work", "issues", "readings", run)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	record := "---\n" +
		"schema_version: 1\n" +
		"id: \"" + item + "\"\n" +
		"run: \"" + run + "\"\n" +
		"manifest: \"sha256:beef\"\n" +
		"position: \"detection\"\n" +
		"regime: \"supplied\"\n" +
		"pattern: \"a stated constraint\"\n" +
		"tension: \"the two sides disagree\"\n" +
		"constraint_in_play: \"the stated invariant\"\n" +
		"why_a_tension: \"one of them must give\"\n" +
		"---\n\n"
	if err := os.WriteFile(filepath.Join(dir, item+".md"), []byte(record), 0o644); err != nil {
		t.Fatal(err)
	}
}

// The disposition verb is the front door of every refusal spc-58 specifies, so
// it has to BE a front door: reachable from the CLI, refusing what the core
// refuses, and writing nothing when it refuses.
func TestCaptureDispositionRefusesEmptyGroundsAndWritesNothing(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	writeReadingFixture(t, repo, "rdg-2608300000000001", "rdi-2608300000000002")

	out, err := runCLIErr(t, "capture", "disposition", "rdi-2608300000000002", "--state", "accepted")
	if err == nil {
		t.Fatalf("a disposition with no grounds must be refused, got success:\n%s", out)
	}
	if !strings.Contains(err.Error(), "disposition_grounds") {
		t.Fatalf("the refusal must name the rule it enforces; got %v", err)
	}
	dispositions := filepath.Join(repo, ".abcd", "work", "issues", "dispositions")
	if entries, derr := os.ReadDir(filepath.Join(dispositions, "rdi-2608300000000002")); derr == nil && len(entries) > 0 {
		t.Fatalf("a refused disposition wrote %d file(s); it must write nothing", len(entries))
	}
}

// The happy path: one answer, one record, under a directory keyed by the item.
func TestCaptureDispositionWritesTheKeyedRecord(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	writeReadingFixture(t, repo, "rdg-2608300000000001", "rdi-2608300000000002")

	out := runCLI(t, "capture", "disposition", "rdi-2608300000000002",
		"--state", "accepted", "--grounds", "the tension is real and worth acting on", "--json")
	var r struct {
		ID       string `json:"id"`
		Item     string `json:"item"`
		State    string `json:"state"`
		Position string `json:"position"`
		Path     string `json:"path"`
	}
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatalf("disposition output not JSON: %v\n%s", err, out)
	}
	if !regexp.MustCompile(`^dsp-[0-9]{16}$`).MatchString(r.ID) {
		t.Fatalf("disposition id = %q, want a native timestamp-numeric dsp id", r.ID)
	}
	if r.Position != "detection" {
		t.Fatalf("position = %q, want the position read off the keyed reading record", r.Position)
	}
	want := filepath.ToSlash(filepath.Join(".abcd/work/issues/dispositions", r.Item, r.ID+".md"))
	if filepath.ToSlash(r.Path) != want {
		t.Fatalf("path = %q, want %q", r.Path, want)
	}
	if _, err := os.Stat(filepath.Join(repo, filepath.FromSlash(r.Path))); err != nil {
		t.Fatalf("the disposition record must exist on disk: %v", err)
	}
}
