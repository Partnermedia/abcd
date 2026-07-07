package history

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/adapter/scanner"
)

const testRootSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

// setupStore points HOME at a temp dir (so both the store root and the probed
// identity home path resolve there), creates the transcripts dir the way abcd
// install would, and returns (repoRoot, home).
func setupStore(t *testing.T) (string, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	tdir := filepath.Join(home, ".abcd", "history", testRootSHA, "transcripts")
	if err := os.MkdirAll(tdir, 0o755); err != nil {
		t.Fatal(err)
	}
	return t.TempDir(), home
}

// TestCaptureRedactsSecretsAndHomePaths is the load-bearing guarantee: a stored
// transcript must NEVER contain a live secret or the caller's absolute home
// path. It plants a ghp_-style PAT and the caller's own home path, captures,
// then asserts the on-disk record carries neither. (In production HOME is a
// /Users/<user> path, so this is exactly the "absolute /Users path redacted"
// case; TestRedactionEngineStripsUsersHomePath proves the literal /Users form
// through the same scanner engine Capture uses.)
func TestCaptureRedactsSecretsAndHomePaths(t *testing.T) {
	repoRoot, home := setupStore(t)

	pat := "ghp_" + strings.Repeat("a", 40)
	selfPath := home + "/private/journal.md"

	transcript := strings.Join([]string{
		"user: deploy with token " + pat,
		"assistant: reading " + selfPath,
		"assistant: done",
	}, "\n")

	res, err := Capture(repoRoot, testRootSHA, "sess-abc123", []byte(transcript), "native")
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	if !res.Wrote {
		t.Fatalf("expected Wrote=true on first capture")
	}

	onDisk, err := os.ReadFile(res.Record.Path)
	if err != nil {
		t.Fatalf("read record: %v", err)
	}

	// The literal sensitive values must be gone from the whole record file.
	mustBeRedacted := []struct {
		name  string
		value string
	}{
		{"github PAT", pat},
		{"self home path", home},
	}
	for _, c := range mustBeRedacted {
		if bytes.Contains(onDisk, []byte(c.value)) {
			t.Errorf("%s leaked into the stored record: %q found on disk", c.name, c.value)
		}
	}

	// The redaction actually happened and was audited.
	if res.Record.Secrets < 1 {
		t.Errorf("expected >=1 secret redaction counted, got %d", res.Record.Secrets)
	}
	if res.Record.HomePaths < 1 {
		t.Errorf("expected >=1 home-path redaction counted, got %d", res.Record.HomePaths)
	}

	// The body is still present and readable, just sanitised.
	if !bytes.Contains(onDisk, []byte("deploy with token")) {
		t.Errorf("non-secret body content was lost")
	}
	if !bytes.Contains(onDisk, []byte("~/private/journal.md")) {
		t.Errorf("self home path should collapse to ~/…, not vanish; got:\n%s", onDisk)
	}
}

// TestCaptureRedactsHomePathFollowedByPunctuation is the Finding-A regression:
// a home path followed by a non-boundary punctuation char (#, &, =, @) must NOT
// survive in a stored record. Before the fix the scanner's trailing-boundary
// heuristic DROPPED such spans and the store's stage-two re-scan (same detector)
// missed them too, leaving the absolute home path and the local username
// verbatim on disk. It uses a distinctive home basename so the username assertion
// cannot alias a timestamp/schema digit.
func TestCaptureRedactsHomePathFollowedByPunctuation(t *testing.T) {
	base := t.TempDir()
	user := "zzhomeuser42"
	home := filepath.Join(base, user)
	t.Setenv("HOME", home)
	tdir := filepath.Join(home, ".abcd", "history", testRootSHA, "transcripts")
	if err := os.MkdirAll(tdir, 0o755); err != nil {
		t.Fatal(err)
	}
	repoRoot := t.TempDir()

	cases := []struct {
		name string
		line string
	}{
		{"hash-after-self-home", "cd " + home + "#draft"},
		{"amp-after-self-home", "run --root=" + home + "&flag=1"},
		{"at-after-self-home", "scp " + home + "@backup"},
		{"eq-after-users-user", "see /Users/" + user + "=cfg"},
	}
	var lines []string
	for _, c := range cases {
		lines = append(lines, c.line)
	}
	transcript := strings.Join(lines, "\n") + "\n"

	res, err := Capture(repoRoot, testRootSHA, "sess-punct", []byte(transcript), "native")
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	if !res.Wrote {
		t.Fatalf("expected Wrote=true on first capture")
	}
	onDisk, err := os.ReadFile(res.Record.Path)
	if err != nil {
		t.Fatalf("read record: %v", err)
	}

	if bytes.Contains(onDisk, []byte(home)) {
		t.Errorf("absolute self home path survived on disk:\n%s", onDisk)
	}
	if bytes.Contains(onDisk, []byte("/Users/"+user)) {
		t.Errorf("/Users/<user> home path survived on disk:\n%s", onDisk)
	}
	if bytes.Contains(onDisk, []byte(user)) {
		t.Errorf("local username %q survived on disk:\n%s", user, onDisk)
	}
	// The counters must reflect that redaction actually occurred.
	if res.Record.HomePaths < 1 {
		t.Errorf("expected >=1 home-path redaction counted, got %d", res.Record.HomePaths)
	}
	// Non-secret body content is preserved.
	if !bytes.Contains(onDisk, []byte("draft")) || !bytes.Contains(onDisk, []byte("cfg")) {
		t.Errorf("non-secret body content was lost:\n%s", onDisk)
	}
}

// TestSurvivingCallerHomeBackstop proves the deterministic, scanner-independent
// backstop (Finding A part 2): it recognises a "/Users/<user>" or "/home/<user>"
// segment for the caller's own username regardless of the trailing character,
// while rejecting a longer, different username. This is the fail-closed guard
// that stands even if the scanner heuristic ever regresses.
func TestSurvivingCallerHomeBackstop(t *testing.T) {
	home := "/base/zzhomeuser42"
	cases := []struct {
		name string
		text string
		want bool
	}{
		{"clean-redacted", "wrote ~/notes and /Users/[redacted-user]/x", false},
		{"literal-home-survives", "path " + home + "/x", true},
		{"users-user-hash", "root=/Users/zzhomeuser42#frag", true},
		{"home-user-amp", "root=/home/zzhomeuser42&x", true},
		{"users-user-eq", "cfg=/Users/zzhomeuser42=v", true},
		{"different-longer-user", "path /Users/zzhomeuser42x/y", false},
		{"different-user", "path /Users/someoneelse/y", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := len(survivingCallerHome(c.text, home)) > 0
			if got != c.want {
				t.Errorf("survivingCallerHome(%q) = %v, want %v", c.text, got, c.want)
			}
		})
	}
}

// TestRedactionEngineStripsUsersHomePath proves the shared scanner engine that
// Capture drives (scanner.ScanText + scanner.Redact) strips a literal
// /Users/<user> absolute home path — the redaction Capture applies to every
// transcript before it lands on disk.
func TestRedactionEngineStripsUsersHomePath(t *testing.T) {
	id := scanner.Identity{HomePath: "/Users/alice", HomeUser: "alice"}
	line := "assistant: wrote /Users/alice/.aws/credentials"

	findings := scanner.ScanText(line, id, scanner.DefaultPatterns(), scanner.DefaultIdentitySeverities(), "transcript")
	redacted, n := scanner.Redact(line, findings)

	if n < 1 {
		t.Fatalf("expected at least one rewrite, got %d (findings: %+v)", n, findings)
	}
	if strings.Contains(redacted, "/Users/alice") {
		t.Errorf("literal /Users home path survived redaction: %q", redacted)
	}
	if !strings.Contains(redacted, "~/.aws/credentials") {
		t.Errorf("home path should collapse to ~/…; got %q", redacted)
	}
}

// TestCaptureIdempotentOnSourceSHA proves a re-capture of identical source is a
// no-op: no second file, mtime preserved, Wrote=false, same path returned.
func TestCaptureIdempotentOnSourceSHA(t *testing.T) {
	repoRoot, _ := setupStore(t)
	raw := []byte("user: hello\nassistant: hi\n")

	first, err := Capture(repoRoot, testRootSHA, "sess-idem", raw, "native")
	if err != nil {
		t.Fatalf("first capture: %v", err)
	}
	if !first.Wrote {
		t.Fatalf("first capture should write")
	}
	fi1, err := os.Stat(first.Record.Path)
	if err != nil {
		t.Fatal(err)
	}

	second, err := Capture(repoRoot, testRootSHA, "sess-idem", raw, "native")
	if err != nil {
		t.Fatalf("second capture: %v", err)
	}
	if second.Wrote {
		t.Errorf("identical source should be an idempotent no-op")
	}
	if second.Record.Path != first.Record.Path {
		t.Errorf("idempotent capture should return the existing record path")
	}
	fi2, err := os.Stat(first.Record.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !fi1.ModTime().Equal(fi2.ModTime()) {
		t.Errorf("mtime changed on idempotent no-op: %v -> %v", fi1.ModTime(), fi2.ModTime())
	}

	recs, err := List(testRootSHA)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Errorf("expected exactly 1 record after idempotent re-capture, got %d", len(recs))
	}
}

// TestListAndRead exercises the read side: metadata newest-first, and a body
// fetch that returns the sanitised content without frontmatter.
func TestListAndRead(t *testing.T) {
	repoRoot, _ := setupStore(t)

	if _, err := Capture(repoRoot, testRootSHA, "sess-one", []byte("first session\n"), "native"); err != nil {
		t.Fatal(err)
	}
	if _, err := Capture(repoRoot, testRootSHA, "sess-two", []byte("second session\n"), "native"); err != nil {
		t.Fatal(err)
	}

	recs, err := List(testRootSHA)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 2 {
		t.Fatalf("expected 2 records, got %d", len(recs))
	}
	// Newest first: sess-two was captured last.
	if recs[0].SessionID != "sess-two" {
		t.Errorf("expected newest (sess-two) first, got %q", recs[0].SessionID)
	}

	rec, body, err := Read(testRootSHA, "sess-one")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if rec.SessionID != "sess-one" {
		t.Errorf("Read returned wrong record: %q", rec.SessionID)
	}
	if strings.Contains(string(body), "---") {
		t.Errorf("Read body should not include frontmatter fence")
	}
	if !strings.Contains(string(body), "first session") {
		t.Errorf("Read body missing content: %q", body)
	}
}

// TestListAbsentCorpusIsCleanEmpty distinguishes an un-populated store (no
// records, no error) from a malformed one, mirroring history_store's LoadResult.
func TestListAbsentCorpusIsCleanEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// No transcripts dir created.
	recs, err := List(testRootSHA)
	if err != nil {
		t.Fatalf("absent corpus should be clean-empty, got error: %v", err)
	}
	if len(recs) != 0 {
		t.Errorf("expected 0 records for absent corpus, got %d", len(recs))
	}
}

// TestCapturePreconditionMissingDir refuses to capture when install never
// created the transcripts dir (never bootstraps it itself).
func TestCapturePreconditionMissingDir(t *testing.T) {
	repoRoot := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	_, err := Capture(repoRoot, testRootSHA, "sess-x", []byte("hi\n"), "native")
	if err == nil {
		t.Fatalf("expected a precondition error when transcripts dir is absent")
	}
	var spe *StorePathError
	if !errors.As(err, &spe) {
		t.Errorf("expected *StorePathError, got %T: %v", err, err)
	}
}

// TestCaptureRejectsBadInput validates the external-input boundary.
func TestCaptureRejectsBadInput(t *testing.T) {
	repoRoot, _ := setupStore(t)
	cases := []struct {
		name      string
		rootSHA   string
		sessionID string
		kind      string
	}{
		{"short sha", "abc", "sess", "native"},
		{"uppercase sha", strings.ToUpper(testRootSHA), "sess", "native"},
		{"empty session", testRootSHA, "", "native"},
		{"path-traversal session", testRootSHA, "../evil", "native"},
		{"bad kind", testRootSHA, "sess", "bogus"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Capture(repoRoot, c.rootSHA, c.sessionID, []byte("x\n"), c.kind); err == nil {
				t.Errorf("expected rejection for %s", c.name)
			}
		})
	}
}
