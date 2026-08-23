package history

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// stagedNames lists the staging dir's entries for assertions.
func stagedNames(t *testing.T, home string) []string {
	t.Helper()
	sdir := filepath.Join(home, ".abcd", "history", testRootSHA, "staging")
	entries, err := os.ReadDir(sdir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out
}

// TestStageWritesRawWithoutRedacting pins the whole point of staging: the
// SessionEnd half must NOT run the scanner. A secret planted in the transcript
// is expected to survive into the staged file verbatim — that is what makes the
// write cheap, and it is why the directory is 0o700 and drained promptly.
func TestStageWritesRawWithoutRedacting(t *testing.T) {
	_, _ = setupStore(t)
	secret := "ghp_" + strings.Repeat("a", 36)
	raw := []byte("assistant: token is " + secret + "\n")

	res, err := Stage(testRootSHA, "sess-stage", raw)
	if err != nil {
		t.Fatalf("Stage failed: %v", err)
	}
	if !res.Wrote {
		t.Fatal("expected Wrote=true on a first stage")
	}
	body, err := os.ReadFile(res.Staged.Path)
	if err != nil {
		t.Fatalf("staged file unreadable: %v", err)
	}
	if !strings.Contains(string(body), secret) {
		t.Fatal("staged file was redacted; staging must write raw bytes or it is not cheap")
	}
	if res.Staged.Bytes != int64(len(raw)) {
		t.Errorf("Bytes = %d, want %d", res.Staged.Bytes, len(raw))
	}
}

// TestStagingDirIsOwnerOnly guards the one place abcd stores unredacted
// transcript text. A group- or world-readable staging dir would leak every
// secret the store exists to keep out.
func TestStagingDirIsOwnerOnly(t *testing.T) {
	_, home := setupStore(t)
	if _, err := Stage(testRootSHA, "sess-perm", []byte("hello\n")); err != nil {
		t.Fatalf("Stage failed: %v", err)
	}
	sdir := filepath.Join(home, ".abcd", "history", testRootSHA, "staging")
	fi, err := os.Stat(sdir)
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o700 {
		t.Errorf("staging dir mode = %o, want 700 (it holds unredacted transcripts)", perm)
	}
	names := stagedNames(t, home)
	if len(names) != 1 {
		t.Fatalf("expected exactly one staged file, got %v", names)
	}
	ffi, err := os.Stat(filepath.Join(sdir, names[0]))
	if err != nil {
		t.Fatal(err)
	}
	if perm := ffi.Mode().Perm(); perm != 0o600 {
		t.Errorf("staged file mode = %o, want 600", perm)
	}
}

// TestStageIsIdempotentPerSession proves a re-fired SessionEnd cannot stage the
// same session twice, which would drain into two records for one session.
func TestStageIsIdempotentPerSession(t *testing.T) {
	_, home := setupStore(t)
	if _, err := Stage(testRootSHA, "sess-dup", []byte("one\n")); err != nil {
		t.Fatal(err)
	}
	second, err := Stage(testRootSHA, "sess-dup", []byte("two\n"))
	if err != nil {
		t.Fatalf("second Stage failed: %v", err)
	}
	if second.Wrote {
		t.Error("expected Wrote=false when the session is already staged")
	}
	if names := stagedNames(t, home); len(names) != 1 {
		t.Errorf("expected 1 staged file after a duplicate stage, got %v", names)
	}
}

// TestDrainCapturesRedactedAndRemovesStaged is the round trip: what staging
// wrote raw must reach the store redacted, and the raw copy must be gone.
func TestDrainCapturesRedactedAndRemovesStaged(t *testing.T) {
	repoRoot, home := setupStore(t)
	secret := "ghp_" + strings.Repeat("b", 36)
	if _, err := Stage(testRootSHA, "sess-drain", []byte("assistant: "+secret+"\n")); err != nil {
		t.Fatal(err)
	}

	res, err := Drain(repoRoot, testRootSHA, 0)
	if err != nil {
		t.Fatalf("Drain failed: %v", err)
	}
	if len(res.Failed) != 0 {
		t.Fatalf("unexpected drain failures: %+v", res.Failed)
	}
	if len(res.Captured) != 1 {
		t.Fatalf("expected 1 captured record, got %d", len(res.Captured))
	}
	if names := stagedNames(t, home); len(names) != 0 {
		t.Errorf("staged copy survived a successful drain: %v", names)
	}
	recs, err := List(testRootSHA)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 || recs[0].SessionID != "sess-drain" {
		t.Fatalf("store does not hold the drained session: %+v", recs)
	}
	body, err := os.ReadFile(recs[0].Path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), secret) {
		t.Fatal("drained record contains the live secret; the store invariant is broken")
	}
}

// TestDrainBudgetLeavesRemainderLoudly pins that a bounded pass reports what it
// did not do. A silent cap would read as "everything is captured" while a
// backlog sat undrained — the exact silence this mechanism exists to end.
func TestDrainBudgetLeavesRemainderLoudly(t *testing.T) {
	repoRoot, home := setupStore(t)
	for _, id := range []string{"sess-a", "sess-b", "sess-c"} {
		if _, err := Stage(testRootSHA, id, []byte("body "+id+"\n")); err != nil {
			t.Fatal(err)
		}
		time.Sleep(2 * time.Millisecond) // distinct stamps so ordering is stable
	}

	res, err := Drain(repoRoot, testRootSHA, 2)
	if err != nil {
		t.Fatalf("Drain failed: %v", err)
	}
	if len(res.Captured) != 2 {
		t.Fatalf("expected 2 captured under a budget of 2, got %d", len(res.Captured))
	}
	if res.Remaining != 1 {
		t.Errorf("Remaining = %d, want 1 — a budgeted pass must report the remainder", res.Remaining)
	}
	if names := stagedNames(t, home); len(names) != 1 {
		t.Errorf("expected 1 staged file left, got %v", names)
	}
}

// TestDrainKeepsStagedOnCaptureFailure is the fail-closed guarantee. If
// redaction cannot produce a clean record, the staged copy is the only copy abcd
// holds; deleting it would turn a reported failure into permanent silent loss.
func TestDrainKeepsStagedOnCaptureFailure(t *testing.T) {
	repoRoot, home := setupStore(t)
	if _, err := Stage(testRootSHA, "sess-fail", []byte("hello\n")); err != nil {
		t.Fatal(err)
	}
	// Break the store's transcripts dir so Capture cannot write.
	tdir := filepath.Join(home, ".abcd", "history", testRootSHA, "transcripts")
	if err := os.RemoveAll(tdir); err != nil {
		t.Fatal(err)
	}

	res, err := Drain(repoRoot, testRootSHA, 0)
	if err != nil {
		t.Fatalf("Drain returned a hard error rather than a per-item failure: %v", err)
	}
	if len(res.Failed) != 1 {
		t.Fatalf("expected 1 recorded failure, got %+v", res.Failed)
	}
	if names := stagedNames(t, home); len(names) != 1 {
		t.Fatal("staged copy was removed despite a failed capture — the only copy would be gone")
	}
}

// TestListStagedIsTheEndedSignal pins the outcome axis the store never had. A
// staged entry means exactly one thing: this session ended and its capture has
// not completed. Absence of a record no longer has to be guessed at.
func TestListStagedIsTheEndedSignal(t *testing.T) {
	repoRoot, _ := setupStore(t)
	if got, err := ListStaged(testRootSHA); err != nil || len(got) != 0 {
		t.Fatalf("empty staging should list nothing: %v %v", got, err)
	}
	if _, err := Stage(testRootSHA, "sess-ended", []byte("body\n")); err != nil {
		t.Fatal(err)
	}
	got, err := ListStaged(testRootSHA)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].SessionID != "sess-ended" {
		t.Fatalf("ListStaged = %+v, want one entry for sess-ended", got)
	}
	if _, err := Drain(repoRoot, testRootSHA, 0); err != nil {
		t.Fatal(err)
	}
	if got, err := ListStaged(testRootSHA); err != nil || len(got) != 0 {
		t.Fatalf("a drained session must leave staging: %+v %v", got, err)
	}
}

// TestSessionIDFromStagedRoundTrips guards the filename parse against session
// ids containing dashes, which the host's uuid-shaped ids always do.
func TestSessionIDFromStagedRoundTrips(t *testing.T) {
	id := "7d884491-9773-4621-b438-77e2ef3d3ed5"
	name := stagedFilename(time.Now().UTC(), id)
	if got := sessionIDFromStaged(name); got != id {
		t.Errorf("sessionIDFromStaged(%q) = %q, want %q", name, got, id)
	}
}
