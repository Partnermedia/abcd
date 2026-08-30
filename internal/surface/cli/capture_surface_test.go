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

	out := runCLI(t, "capture", "promote", minted.ID, "--grounds", cliGrounds, "--json")
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
	if _, err := runCLIErr(t, "capture", "promote", minted.ID, "--grounds", cliGrounds); err == nil || !strings.Contains(err.Error(), "itd-1") {
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
	if _, err := runCLIErr(t, "capture", "resolve", minted.ID, "nope", "--impact", "fix", "--grounds", cliGrounds, "--intent", "itd-99"); err == nil {
		t.Fatalf("resolve with an unknown --intent must refuse")
	}

	out := runCLI(t, "capture", "resolve", minted.ID, "fixed by the fixer", "--impact", "fix",
		"--grounds", cliGrounds, "--intent", "itd-4", "--commit", "abcdef0123", "--json")
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

// TestCaptureLapsedAtWritesTheGivenInstant pins the flag half of spc-60: the
// instant handed to --lapsed-at is the instant committed to the record. The
// record id is minted from the wall clock, so a surface that dropped, rounded or
// re-derived the value would leave a lapse entry stamped with its own write-up
// time and nothing to say so.
func TestCaptureLapsedAtWritesTheGivenInstant(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	const lapsedAt = "2026-08-28T09:15:00Z"
	out := runCLI(t, "capture", "the discipline gave way here",
		"--category", "lapse", "--lapsed-at", lapsedAt, "--json")
	var r struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatalf("capture output not JSON: %v\n%s", err, out)
	}
	body, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(r.Path)))
	if err != nil {
		t.Fatalf("read the captured record: %v", err)
	}
	if want := "lapsed_at: \"" + lapsedAt + "\""; !strings.Contains(string(body), want) {
		t.Fatalf("the captured record does not carry %s:\n%s", want, body)
	}
}

// TestCaptureLapsedAtHasNoDefault is the flag's whole point, made checkable: a
// lapse capture with --lapsed-at omitted is refused, names the flag, and writes
// nothing. Every other provenance flag on capture falls back to a default; the
// only fallback available here is the wall clock at write-up, which is the one
// value itd-182's criterion rules out — so the surface must refuse rather than
// invent, and it must refuse BEFORE the ledger gains a record.
func TestCaptureLapsedAtHasNoDefault(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	out, err := runCLIErr(t, "capture", "the discipline gave way here", "--category", "lapse")
	if err == nil {
		t.Fatalf("a lapse capture with no --lapsed-at succeeded:\n%s", out)
	}
	if !strings.Contains(err.Error(), "--lapsed-at") {
		t.Fatalf("the refusal does not name the flag the caller must supply: %v", err)
	}
	if n := ledgerIssueCount(t, repo); n != 0 {
		t.Fatalf("the refused lapse capture wrote %d record(s); it must write nothing", n)
	}
}

// cliGrounds is a conjecture-shaped operand for the capture surface tests.
const cliGrounds = "pursued: we expect the recorded reasoning to outlive the session that had it"

// TestCapturePromoteMissingGroundsExit2 is itd-179's refusal at the surface: the
// flag is mandatory in effect and its absence is a USAGE error, refused at exit 2
// with nothing written — the same shape `--category lapse` without `--lapsed-at`
// already has. Exit 1 is reserved for a gate's own verdict.
func TestCapturePromoteMissingGroundsExit2(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	capOut := runCLI(t, "capture", "an observation that may turn out to be a capability", "--json")
	var minted struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(capOut, &minted); err != nil || minted.ID == "" {
		t.Fatalf("capture envelope unreadable: %v\n%s", err, capOut)
	}

	for _, args := range [][]string{
		{"capture", "promote", minted.ID},
		{"capture", "resolve", minted.ID, "fixed", "--impact", "fix"},
	} {
		_, err := runCLIErr(t, args...)
		if exitCodeOf(err) != 2 {
			t.Fatalf("%v exit = %d (%v), want 2", args, exitCodeOf(err), err)
		}
		if !strings.Contains(err.Error(), "--grounds") {
			t.Fatalf("%v error must name the flag, got %q", args, err.Error())
		}
	}
	// Nothing was written: no draft minted, and the issue is still open.
	if entries, _ := os.ReadDir(filepath.Join(repo, cliDrafts)); len(entries) != 0 {
		t.Fatalf("a refused promote minted %d draft(s), want 0", len(entries))
	}
	out := runCLI(t, "capture", "list", "--open", "--json")
	if !strings.Contains(string(out), minted.ID) {
		t.Fatalf("a refused triage moved the issue out of open/:\n%s", out)
	}
}

// TestCaptureGroundsReachTheRecord: the wired flag actually records, on all three
// triage routes, and wontfix derives its `declined:` grounds from the reason it
// already takes.
//
// It asserts the BULLET in the record body, which is where grounds live: a
// frontmatter scalar is set, and setting is what let one route overwrite
// another's conjecture (iss-2608301657354776). The end-to-end path is what this
// test is for — a writer that appended somewhere the record's own section reader
// cannot see would still satisfy the core tests.
func TestCaptureGroundsReachTheRecord(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	mint := func(text string) string {
		t.Helper()
		var m struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(runCLI(t, "capture", text, "--json"), &m); err != nil || m.ID == "" {
			t.Fatalf("capture envelope unreadable: %v", err)
		}
		return m.ID
	}
	readRecord := func(id string) string {
		t.Helper()
		var listed struct {
			Issues []struct {
				ID   string `json:"id"`
				Path string `json:"path"`
			} `json:"issues"`
		}
		if err := json.Unmarshal(runCLI(t, "capture", "list", "--all", "--json"), &listed); err != nil {
			t.Fatal(err)
		}
		for _, iss := range listed.Issues {
			if iss.ID == id {
				data, err := os.ReadFile(filepath.Join(repo, iss.Path))
				if err != nil {
					t.Fatal(err)
				}
				return string(data)
			}
		}
		t.Fatalf("%s not found in the ledger", id)
		return ""
	}

	promoted := mint("an observation worth graduating into an intent")
	runCLI(t, "capture", "promote", promoted, "--grounds", cliGrounds)
	if !strings.Contains(readRecord(promoted), "\n## Grounds\n\n- "+cliGrounds+"\n") {
		t.Fatalf("promote did not record the grounds as a body bullet:\n%s", readRecord(promoted))
	}

	resolved := mint("an observation that will be fixed outright")
	runCLI(t, "capture", "resolve", resolved, "fixed", "--impact", "fix", "--grounds", cliGrounds)
	if !strings.Contains(readRecord(resolved), "\n## Grounds\n\n- "+cliGrounds+"\n") {
		t.Fatalf("resolve did not record the grounds as a body bullet:\n%s", readRecord(resolved))
	}

	declined := mint("an observation that will not be acted on")
	runCLI(t, "capture", "wontfix", declined, "out of scope for this cycle")
	if !strings.Contains(readRecord(declined), "\n## Grounds\n\n- declined: out of scope for this cycle\n") {
		t.Fatalf("wontfix did not derive its declined grounds as a body bullet:\n%s", readRecord(declined))
	}
}

// TestCaptureMalformedGroundsExit2 closes the uneven half of
// iss-2608300930057882: a MISSING --grounds exited 2 while a MALFORMED one
// exited 1, so a caller distinguishing usage errors from real failures learned
// the wrong thing from the same flag. Every grounds refusal is a usage error.
func TestCaptureMalformedGroundsExit2(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	capOut := runCLI(t, "capture", "an observation that may turn out to be a capability", "--json")
	var minted struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(capOut, &minted); err != nil || minted.ID == "" {
		t.Fatalf("capture envelope unreadable: %v\n%s", err, capOut)
	}

	for _, bad := range []string{
		"planned: out of vocabulary entirely",
		"no token at all in this operand",
		"pursued: yes",
	} {
		for _, args := range [][]string{
			{"capture", "promote", minted.ID, "--grounds", bad},
			{"capture", "resolve", minted.ID, "fixed", "--impact", "fix", "--grounds", bad},
			{"capture", "wontfix", minted.ID, "no", "--grounds", bad},
		} {
			_, err := runCLIErr(t, args...)
			if exitCodeOf(err) != 2 {
				t.Fatalf("%v with %q: exit = %d (%v), want 2", args[:2], bad, exitCodeOf(err), err)
			}
		}
	}
}

// TestGroundsFlagUsageRendersAStringPlaceholder: cobra's UnquoteUsage takes the
// first backquoted word of a flag's usage string as the flag's value
// placeholder and strips it from the prose, so backticks in the wontfix
// `--grounds` help printed `--grounds declined` and lost the word
// (iss-2608301212428844). Every grounds flag names a string.
func TestGroundsFlagUsageRendersAStringPlaceholder(t *testing.T) {
	for _, verb := range []string{"promote", "resolve", "wontfix"} {
		out, err := runCLIErr(t, "capture", verb, "--help")
		if err != nil {
			t.Fatalf("capture %s --help: %v\n%s", verb, err, out)
		}
		if !strings.Contains(string(out), "--grounds string") {
			t.Fatalf("capture %s --help renders no `--grounds string` placeholder:\n%s", verb, out)
		}
	}
}
