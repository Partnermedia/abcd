package cite

import (
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/REPPL/abcd-cli/internal/core/lint"
)

// now is the clock every refresh test runs against, so a dated record is a
// function of the fixture rather than of the day the suite runs.
var now = time.Date(2026, 7, 27, 9, 0, 0, 0, time.UTC)

const today = "2026-07-27"

// stubChecker answers from a table and records every URL it was asked about, so
// a test can assert not only the outcome but that a URL was — or was not —
// fetched at all.
type stubChecker struct {
	answers map[string]CheckOutcome
	mu      sync.Mutex
	asked   []string
}

func (s *stubChecker) Check(rawURL string) CheckOutcome {
	s.mu.Lock()
	s.asked = append(s.asked, rawURL)
	s.mu.Unlock()
	if out, ok := s.answers[rawURL]; ok {
		out.URL = rawURL
		return out
	}
	return CheckOutcome{URL: rawURL, Status: StatusOK, FinalURL: rawURL}
}

func (s *stubChecker) askedFor(rawURL string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.asked {
		if a == rawURL {
			return true
		}
	}
	return false
}

// citeRepo plants a fixture repo whose docs cite the given URLs, one per page.
func citeRepo(t *testing.T, urls ...string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	for i, u := range urls {
		body := "# Page\n\nA claim.[^a]\n\n[^a]: A source, " + u + "\n"
		name := filepath.Join(root, "docs", "p"+string(rune('a'+i))+".md")
		if err := os.WriteFile(name, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// testConfig is a docs-lint config whose citation rule points at the default
// baseline path and carries spc-17's thresholds.
func testConfig() lint.Config {
	return lint.Config{
		Roots: []string{"docs"},
		Rules: map[string]lint.RuleConfig{
			"citation_baseline": {Enabled: true, Severity: "blocker", WarnAfterDays: 180, BlockAfterDays: 365},
		},
	}
}

// loadBaseline reads back what a refresh wrote, THROUGH the gate's own loader —
// so every test asserts not just the content but that the gate can read it.
func loadBaseline(t *testing.T, root string) lint.Baseline {
	t.Helper()
	b, err := lint.LoadBaseline(filepath.Join(root, filepath.FromSlash(lint.DefaultBaselinePath)))
	if err != nil {
		t.Fatalf("LoadBaseline: %v", err)
	}
	return b
}

func writeBaseline(t *testing.T, root string, entries map[string]lint.BaselineEntry) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(lint.DefaultBaselinePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := lint.SaveBaseline(path, lint.Baseline{SchemaVersion: 1, Entries: entries}); err != nil {
		t.Fatal(err)
	}
}

// TestRefreshWritesAliveEntries is the happy path: a live citation becomes an
// automatic entry dated today, recording where it actually resolved.
func TestRefreshWritesAliveEntries(t *testing.T) {
	root := citeRepo(t, "https://example.org/a")
	checker := &stubChecker{answers: map[string]CheckOutcome{
		"https://example.org/a": {Status: StatusOK, FinalURL: "https://example.org/a-moved"},
	}}

	res, err := Refresh(RefreshRequest{RepoRoot: root, Config: testConfig(), Checker: checker, Now: now})
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if res.Cited != 1 || res.Fetched != 1 {
		t.Fatalf("cited=%d fetched=%d, want 1/1", res.Cited, res.Fetched)
	}

	got := loadBaseline(t, root).Entries["https://example.org/a"]
	want := lint.BaselineEntry{
		URL:          "https://example.org/a",
		FinalURL:     "https://example.org/a-moved",
		LastChecked:  today,
		Outcome:      lint.OutcomeAlive,
		Verification: lint.VerificationAutomatic,
		VerifiedOn:   today,
	}
	if got != want {
		t.Fatalf("entry = %+v, want %+v", got, want)
	}
}

// TestRefreshRecordsBroken pins that a dead link is recorded as dead, so the
// offline gate blocks on it without anyone re-fetching.
func TestRefreshRecordsBroken(t *testing.T) {
	root := citeRepo(t, "https://example.org/gone")
	checker := &stubChecker{answers: map[string]CheckOutcome{
		"https://example.org/gone": {Status: StatusBroken, Detail: "HTTP 404 Not Found"},
	}}

	if _, err := Refresh(RefreshRequest{RepoRoot: root, Config: testConfig(), Checker: checker, Now: now}); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	got := loadBaseline(t, root).Entries["https://example.org/gone"]
	if got.Outcome != lint.OutcomeBroken {
		t.Fatalf("outcome = %q, want %q", got.Outcome, lint.OutcomeBroken)
	}
	// A broken entry still needs a final_url for the schema; the cited address
	// is the honest answer when nothing else resolved.
	if got.FinalURL != "https://example.org/gone" {
		t.Errorf("final_url = %q, want the cited address", got.FinalURL)
	}
}

// TestRefreshRoutesBlockedToTheQueueWithoutInventingAnEntry is the honesty
// clause. A 403 says the fetcher may not look — it is NOT evidence the source is
// gone, so nothing is written. The gate then reports the URL as unreceipted,
// which is exactly true, and the manual queue names it for a human.
func TestRefreshRoutesBlockedToTheQueueWithoutInventingAnEntry(t *testing.T) {
	root := citeRepo(t, "https://paywalled.example.org/x")
	checker := &stubChecker{answers: map[string]CheckOutcome{
		"https://paywalled.example.org/x": {Status: StatusBlocked, Detail: "HTTP 403 Forbidden"},
	}}

	res, err := Refresh(RefreshRequest{RepoRoot: root, Config: testConfig(), Checker: checker, Now: now})
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if _, ok := loadBaseline(t, root).Entries["https://paywalled.example.org/x"]; ok {
		t.Fatal("a blocked fetch invented a baseline entry; only a human may record this one")
	}
	if len(res.Queue) != 1 || res.Queue[0].URL != "https://paywalled.example.org/x" {
		t.Fatalf("queue = %+v, want the blocked URL", res.Queue)
	}
	if len(res.Queue[0].Sites) == 0 {
		t.Error("a queue item must name where the URL is cited")
	}
}

// TestRefreshPreservesAFreshManualReceipt is the never-downgrade rule: a human
// confirmation is not re-litigated by a fetcher, and the URL is not even
// requested — so a source that only ever answers 403 does not re-enter the queue
// on every run.
func TestRefreshPreservesAFreshManualReceipt(t *testing.T) {
	root := citeRepo(t, "https://paywalled.example.org/x")
	manual := lint.BaselineEntry{
		FinalURL: "https://paywalled.example.org/x", LastChecked: "2026-06-01",
		Outcome: lint.OutcomeAlive, Verification: lint.VerificationManual, VerifiedOn: "2026-06-01",
	}
	writeBaseline(t, root, map[string]lint.BaselineEntry{"https://paywalled.example.org/x": manual})

	checker := &stubChecker{}
	res, err := Refresh(RefreshRequest{RepoRoot: root, Config: testConfig(), Checker: checker, Now: now})
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if checker.askedFor("https://paywalled.example.org/x") {
		t.Error("a fresh manual receipt was re-fetched; it must be left alone until it ages out")
	}
	if res.Preserved != 1 || res.Fetched != 0 {
		t.Fatalf("preserved=%d fetched=%d, want 1/0", res.Preserved, res.Fetched)
	}
	got := loadBaseline(t, root).Entries["https://paywalled.example.org/x"]
	got.URL = "" // the writer restates the key; the fixture did not
	if got != manual {
		t.Fatalf("entry = %+v, want it preserved verbatim %+v", got, manual)
	}
}

// TestRefreshRecheckesAStaleManualReceipt is AC 3's "same clock": once a human
// verification passes the staleness threshold it stops being authoritative, and
// the refresh tries again. A fetch that now succeeds re-dates the entry as
// automatic — the manual receipt has already aged out, so nothing is downgraded.
func TestRefreshRechecksAStaleManualReceipt(t *testing.T) {
	root := citeRepo(t, "https://example.org/x")
	writeBaseline(t, root, map[string]lint.BaselineEntry{"https://example.org/x": {
		FinalURL: "https://example.org/x", LastChecked: "2025-01-01",
		Outcome: lint.OutcomeAlive, Verification: lint.VerificationManual, VerifiedOn: "2025-01-01",
	}})

	checker := &stubChecker{}
	res, err := Refresh(RefreshRequest{RepoRoot: root, Config: testConfig(), Checker: checker, Now: now})
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if !checker.askedFor("https://example.org/x") {
		t.Fatal("a stale manual receipt was not re-checked")
	}
	if res.Preserved != 0 {
		t.Errorf("preserved = %d, want 0", res.Preserved)
	}
	got := loadBaseline(t, root).Entries["https://example.org/x"]
	if got.Verification != lint.VerificationAutomatic || got.LastChecked != today {
		t.Fatalf("entry = %+v, want a fresh automatic receipt", got)
	}
}

// TestRefreshKeepsAStaleManualReceiptWhenTheFetchIsStillBlocked closes the other
// half: re-checking a stale manual entry must not DELETE it when the fetcher is
// still refused. The stale receipt stays (so the gate keeps warning, honestly)
// and the URL re-enters the queue for a human.
func TestRefreshKeepsAStaleManualReceiptWhenTheFetchIsStillBlocked(t *testing.T) {
	root := citeRepo(t, "https://paywalled.example.org/x")
	stale := lint.BaselineEntry{
		FinalURL: "https://paywalled.example.org/x", LastChecked: "2025-01-01",
		Outcome: lint.OutcomeAlive, Verification: lint.VerificationManual, VerifiedOn: "2025-01-01",
	}
	writeBaseline(t, root, map[string]lint.BaselineEntry{"https://paywalled.example.org/x": stale})

	checker := &stubChecker{answers: map[string]CheckOutcome{
		"https://paywalled.example.org/x": {Status: StatusBlocked, Detail: "HTTP 403 Forbidden"},
	}}
	res, err := Refresh(RefreshRequest{RepoRoot: root, Config: testConfig(), Checker: checker, Now: now})
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	got := loadBaseline(t, root).Entries["https://paywalled.example.org/x"]
	got.URL = ""
	if got != stale {
		t.Fatalf("entry = %+v, want the stale receipt kept %+v", got, stale)
	}
	if len(res.Queue) != 1 {
		t.Fatalf("queue = %+v, want the still-blocked URL", res.Queue)
	}
}

// TestRefreshDropsUncitedEntries keeps the record honest in the other direction:
// the baseline answers for what the docs cite TODAY, so a receipt for a citation
// that has been edited out is removed rather than left to age forever.
func TestRefreshDropsUncitedEntries(t *testing.T) {
	root := citeRepo(t, "https://example.org/kept")
	writeBaseline(t, root, map[string]lint.BaselineEntry{
		"https://example.org/kept": {FinalURL: "https://example.org/kept", LastChecked: "2026-07-01",
			Outcome: lint.OutcomeAlive, Verification: lint.VerificationAutomatic, VerifiedOn: "2026-07-01"},
		"https://example.org/removed": {FinalURL: "https://example.org/removed", LastChecked: "2026-07-01",
			Outcome: lint.OutcomeAlive, Verification: lint.VerificationManual, VerifiedOn: "2026-07-01"},
	})

	res, err := Refresh(RefreshRequest{RepoRoot: root, Config: testConfig(), Checker: &stubChecker{}, Now: now})
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	entries := loadBaseline(t, root).Entries
	if _, ok := entries["https://example.org/removed"]; ok {
		t.Error("an entry for a URL the docs no longer cite survived the refresh")
	}
	if len(res.Dropped) != 1 || res.Dropped[0] != "https://example.org/removed" {
		t.Errorf("dropped = %v, want the uncited URL", res.Dropped)
	}
}

// TestRefreshIsDeterministic pins that two runs over the same inputs produce the
// same bytes — the baseline is committed and reviewed in diffs, so a refresh
// that reordered entries would generate churn nobody can read.
func TestRefreshIsDeterministic(t *testing.T) {
	root := citeRepo(t, "https://example.org/a", "https://example.org/b", "https://example.org/c")
	path := filepath.Join(root, filepath.FromSlash(lint.DefaultBaselinePath))

	run := func() string {
		if _, err := Refresh(RefreshRequest{RepoRoot: root, Config: testConfig(), Checker: &stubChecker{}, Now: now}); err != nil {
			t.Fatalf("Refresh: %v", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	if first, second := run(), run(); first != second {
		t.Fatal("two refreshes over identical inputs produced different bytes")
	}
}

// TestRefreshOutcomesAreSorted keeps the printed transcript stable, which is what
// lets an operator diff two runs by eye.
func TestRefreshOutcomesAreSorted(t *testing.T) {
	root := citeRepo(t, "https://example.org/c", "https://example.org/a", "https://example.org/b")
	res, err := Refresh(RefreshRequest{RepoRoot: root, Config: testConfig(), Checker: &stubChecker{}, Now: now})
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	urls := make([]string, 0, len(res.Outcomes))
	for _, o := range res.Outcomes {
		urls = append(urls, o.URL)
	}
	if !sort.StringsAreSorted(urls) {
		t.Fatalf("outcomes are not sorted: %v", urls)
	}
}

// TestRefreshRefusesAMalformedBaseline is the fail-closed clause: a corrupt
// committed record must stop the run, never be silently replaced by a fresh one
// that discards whatever manual receipts it held.
func TestRefreshRefusesAMalformedBaseline(t *testing.T) {
	root := citeRepo(t, "https://example.org/a")
	path := filepath.Join(root, filepath.FromSlash(lint.DefaultBaselinePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"schema_version": 1, "entries": {"x": {"method": "curl"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Refresh(RefreshRequest{RepoRoot: root, Config: testConfig(), Checker: &stubChecker{}, Now: now}); err == nil {
		t.Fatal("Refresh overwrote a baseline it could not read")
	}
}

// TestRefreshRefusesToCommitATotalFailure is the record-preservation clause.
//
// Check collapses a DNS failure, a refused connection and a timeout into the same
// StatusBroken as a genuine 404, so a run made with the VPN down or behind a
// captive portal would otherwise rewrite EVERY entry as broken — including stale
// human receipts, which are re-checked and so are not protected by the blocked
// branch — and commit it before the operator saw the transcript. The gate would
// then block every commit in the repo until someone reverted the file by hand.
func TestRefreshRefusesToCommitATotalFailure(t *testing.T) {
	root := citeRepo(t, "https://example.org/a", "https://example.org/b")
	good := map[string]lint.BaselineEntry{
		"https://example.org/a": {FinalURL: "https://example.org/a", LastChecked: "2026-07-01",
			Outcome: lint.OutcomeAlive, Verification: lint.VerificationAutomatic, VerifiedOn: "2026-07-01"},
		"https://example.org/b": {FinalURL: "https://example.org/b", LastChecked: "2025-01-01",
			Outcome: lint.OutcomeAlive, Verification: lint.VerificationManual, VerifiedOn: "2025-01-01"},
	}
	writeBaseline(t, root, good)

	offline := &stubChecker{answers: map[string]CheckOutcome{
		"https://example.org/a": {Status: StatusBroken, Detail: "dial tcp: no such host"},
		"https://example.org/b": {Status: StatusBroken, Detail: "dial tcp: no such host"},
	}}
	if _, err := Refresh(RefreshRequest{RepoRoot: root, Config: testConfig(), Checker: offline, Now: now}); err == nil {
		t.Fatal("Refresh committed a run in which nothing succeeded")
	}

	entries := loadBaseline(t, root).Entries
	for url, want := range good {
		got := entries[url]
		got.URL = ""
		if got != want {
			t.Fatalf("%s = %+v, want the prior record untouched %+v", url, got, want)
		}
	}
}

// TestRefreshStillRecordsAnIsolatedBrokenLink keeps that guard narrow: a dead
// link among live ones is exactly what the baseline exists to record.
func TestRefreshStillRecordsAnIsolatedBrokenLink(t *testing.T) {
	root := citeRepo(t, "https://example.org/a", "https://example.org/gone")
	writeBaseline(t, root, map[string]lint.BaselineEntry{
		"https://example.org/a": {FinalURL: "https://example.org/a", LastChecked: "2026-07-01",
			Outcome: lint.OutcomeAlive, Verification: lint.VerificationAutomatic, VerifiedOn: "2026-07-01"},
	})
	checker := &stubChecker{answers: map[string]CheckOutcome{
		"https://example.org/gone": {Status: StatusBroken, Detail: "HTTP 404 Not Found"},
	}}
	if _, err := Refresh(RefreshRequest{RepoRoot: root, Config: testConfig(), Checker: checker, Now: now}); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if got := loadBaseline(t, root).Entries["https://example.org/gone"]; got.Outcome != lint.OutcomeBroken {
		t.Fatalf("entry = %+v, want outcome broken", got)
	}
}

// TestRefreshCommitsATotalFailureOnAFirstRun keeps the guard from deadlocking a
// repo whose citations really are all dead: with no prior record to protect,
// there is nothing to lose by writing one.
func TestRefreshCommitsATotalFailureOnAFirstRun(t *testing.T) {
	root := citeRepo(t, "https://example.org/gone")
	checker := &stubChecker{answers: map[string]CheckOutcome{
		"https://example.org/gone": {Status: StatusBroken, Detail: "HTTP 404 Not Found"},
	}}
	if _, err := Refresh(RefreshRequest{RepoRoot: root, Config: testConfig(), Checker: checker, Now: now}); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if got := loadBaseline(t, root).Entries["https://example.org/gone"]; got.Outcome != lint.OutcomeBroken {
		t.Fatalf("entry = %+v, want outcome broken", got)
	}
}

// TestRefreshRefusesAnUnknownCheckerStatus closes the seam's silent-drop hole.
// Status and Checker are exported for the adapter spc-17 AC 5 calls for; one
// returning anything else would otherwise omit the URL from BOTH the baseline and
// the queue, with no error anywhere.
func TestRefreshRefusesAnUnknownCheckerStatus(t *testing.T) {
	root := citeRepo(t, "https://example.org/a")
	rogue := &stubChecker{answers: map[string]CheckOutcome{
		"https://example.org/a": {Status: Status("probably-fine")},
	}}
	if _, err := Refresh(RefreshRequest{RepoRoot: root, Config: testConfig(), Checker: rogue, Now: now}); err == nil {
		t.Fatal("Refresh accepted an unrecognised checker status")
	}
}
