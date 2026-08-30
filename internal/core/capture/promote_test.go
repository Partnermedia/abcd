package capture

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/intentdriven/abcd/internal/core/frontmatter"
	"github.com/intentdriven/abcd/internal/core/intent"
	"github.com/intentdriven/abcd/internal/core/issueschema"
)

// promoteFixture captures one issue into a fresh ledger and returns the roots
// plus the minted iss id.
func promoteFixture(t *testing.T, text string) (repo, ir, issID string) {
	t.Helper()
	repo, ir = ledger(t)
	res, err := Capture(CaptureRequest{
		RepoRoot: repo, IssuesRoot: ir, Text: text, Severity: SeverityMinor,
		Category: "observation", Source: "user-observation", FoundDuring: "t",
		Slug: "a-promotable-observation",
	})
	if err != nil {
		t.Fatal(err)
	}
	return repo, ir, res.ID
}

// draftCount counts intent files under drafts/.
func draftCount(t *testing.T, repo string) int {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(repo, intent.IntentsRelDir, intent.BucketDrafts))
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			n++
		}
	}
	return n
}

// readIssue re-reads an issue by id and returns it.
func readIssue(t *testing.T, ir, issID string) Issue {
	t.Helper()
	path, status, err := findIssue(ir, issID)
	if err != nil {
		t.Fatalf("findIssue(%s): %v", issID, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fm, body, err := parseFrontmatterAndBody(string(data))
	if err != nil {
		t.Fatal(err)
	}
	return issueFromFrontmatter(fm, status, path, body)
}

// TestPromoteMintsDraftAndStampsIssue is the spc-24 headline: one invocation
// mints an intent draft from the issue (slug reused, by-id pointer body,
// promoted_from back-edge) and stamps the issue's promoted_to with the minted
// itd-N.
func TestPromoteMintsDraftAndStampsIssue(t *testing.T) {
	repo, ir, issID := promoteFixture(t, "the loader drops rules silently when the config is stale")

	res, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: issID})
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if res.IssueID != issID {
		t.Fatalf("IssueID = %q, want %q", res.IssueID, issID)
	}
	if res.Linked {
		t.Fatalf("mint mode must report Linked=false")
	}
	if res.IntentID != "itd-1" {
		t.Fatalf("IntentID = %q, want itd-1", res.IntentID)
	}
	if filepath.IsAbs(res.IntentPath) || filepath.IsAbs(res.IssuePath) {
		t.Fatalf("result paths must be repo-relative, got %q / %q", res.IntentPath, res.IssuePath)
	}

	// The minted draft: reused slug, placeholder press release, by-id pointer —
	// never a copy of the issue body — and the promoted_from back-edge.
	abs := filepath.Join(repo, res.IntentPath)
	data, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("minted draft unreadable: %v", err)
	}
	draft := string(data)
	if !strings.Contains(res.IntentPath, "a-promotable-observation") {
		t.Fatalf("draft slug not reused from the issue: %q", res.IntentPath)
	}
	if !strings.Contains(draft, "## Press Release") {
		t.Fatalf("draft missing the placeholder Press Release section:\n%s", draft)
	}
	if !strings.Contains(draft, "Graduated from `"+issID+"`") {
		t.Fatalf("draft missing the by-id pointer to %s:\n%s", issID, draft)
	}
	if !strings.Contains(draft, "promoted_from: "+issID) {
		t.Fatalf("draft frontmatter missing promoted_from %s:\n%s", issID, draft)
	}

	// The issue: stamped in place, still in open/.
	iss := readIssue(t, ir, issID)
	if iss.PromotedTo != res.IntentID {
		t.Fatalf("issue promoted_to = %q, want %q", iss.PromotedTo, res.IntentID)
	}
	if iss.Status != StateOpen {
		t.Fatalf("promotion must not move the issue; status = %q", iss.Status)
	}
}

// TestPromoteWorksInAnyStatusAndKeepsFolder: promotion is orthogonal to
// fix-status — a resolved or wontfixed issue graduates in place, no move.
func TestPromoteWorksInAnyStatusAndKeepsFolder(t *testing.T) {
	for _, tc := range []struct {
		name   string
		move   func(repo, ir, id string) error
		status State
	}{
		{"resolved", func(repo, ir, id string) error {
			_, err := Resolve(ResolveRequest{RepoRoot: repo, IssuesRoot: ir, ID: id, Resolution: "done", Impact: "fix"})
			return err
		}, StateResolved},
		{"wontfix", func(repo, ir, id string) error {
			_, err := Wontfix(WontfixRequest{RepoRoot: repo, IssuesRoot: ir, ID: id, Reason: "out of scope"})
			return err
		}, StateWontfix},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, ir, issID := promoteFixture(t, "an observation that grew up after being closed")
			if err := tc.move(repo, ir, issID); err != nil {
				t.Fatal(err)
			}
			res, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: issID})
			if err != nil {
				t.Fatalf("Promote in %s/: %v", tc.status, err)
			}
			iss := readIssue(t, ir, issID)
			if iss.Status != tc.status {
				t.Fatalf("promotion moved the issue: status = %q, want %q", iss.Status, tc.status)
			}
			if iss.PromotedTo != res.IntentID {
				t.Fatalf("promoted_to = %q, want %q", iss.PromotedTo, res.IntentID)
			}
		})
	}
}

// TestPromoteRefusesAlreadyPromoted: a second promote reports the existing
// itd-N and mints no duplicate draft.
func TestPromoteRefusesAlreadyPromoted(t *testing.T) {
	repo, ir, issID := promoteFixture(t, "promote me once")
	first, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: issID})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: issID})
	if err == nil {
		t.Fatalf("second promote must refuse")
	}
	if !strings.Contains(err.Error(), first.IntentID) {
		t.Fatalf("refusal must name the existing %s, got: %v", first.IntentID, err)
	}
	if n := draftCount(t, repo); n != 1 {
		t.Fatalf("second promote minted a duplicate draft: %d drafts", n)
	}
}

// TestPromoteUnknownOrMalformedIDWritesNothing: a structural fault leaves both
// stores byte-identical.
func TestPromoteUnknownOrMalformedIDWritesNothing(t *testing.T) {
	repo, ir, issID := promoteFixture(t, "an innocent bystander")
	for _, bad := range []string{"iss-999", "not-an-id", "../../evil"} {
		if _, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: bad}); err == nil {
			t.Fatalf("Promote(%q) must fail", bad)
		}
		if n := draftCount(t, repo); n != 0 {
			t.Fatalf("Promote(%q) minted a draft on a structural fault", bad)
		}
	}
	// Link mode with an unknown intent: structural fault, nothing written.
	if _, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: issID, LinkIntent: "itd-42"}); err == nil {
		t.Fatalf("link mode must refuse an unknown itd-N")
	}
	if iss := readIssue(t, ir, issID); iss.PromotedTo != "" {
		t.Fatalf("failed link mode stamped promoted_to = %q", iss.PromotedTo)
	}
}

// TestPromoteStampFailureReportsOrphanAndLinkRepairs simulates a stamp failure
// after the mint (unwritable ledger, injected via the stampWriteHook seam —
// a chmod'd status dir is a no-op for root): the error must carry the orphan
// draft's path and the --intent remedy, and a follow-up link-mode run must
// complete the stamp without minting a second draft.
func TestPromoteStampFailureReportsOrphanAndLinkRepairs(t *testing.T) {
	repo, ir, issID := promoteFixture(t, "the stamp will fail after the mint")

	stampWriteHook = func(string, []byte) error {
		return errors.New("simulated unwritable ledger")
	}
	defer func() { stampWriteHook = nil }()

	_, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: issID})
	if err == nil {
		t.Fatalf("stamp into an unwritable ledger must fail")
	}
	if !strings.Contains(err.Error(), "itd-1") {
		t.Fatalf("orphan report must name the minted draft, got: %v", err)
	}
	if !strings.Contains(err.Error(), "--intent itd-1") {
		t.Fatalf("orphan report must carry the --intent remedy, got: %v", err)
	}
	if n := draftCount(t, repo); n != 1 {
		t.Fatalf("mint-first contract: expected the orphan draft to persist, got %d", n)
	}

	// Repair: link the orphan draft. No second mint.
	stampWriteHook = nil
	res, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: issID, LinkIntent: "itd-1"})
	if err != nil {
		t.Fatalf("link-mode repair: %v", err)
	}
	if !res.Linked {
		t.Fatalf("repair must report Linked=true")
	}
	if n := draftCount(t, repo); n != 1 {
		t.Fatalf("link mode minted: %d drafts, want 1", n)
	}
	if iss := readIssue(t, ir, issID); iss.PromotedTo != "itd-1" {
		t.Fatalf("repair did not stamp promoted_to: %q", iss.PromotedTo)
	}
}

// TestPromoteSerializesOnLedgerLock: the stamp acquires the same allocator
// flock every ledger mutation takes.
func TestPromoteSerializesOnLedgerLock(t *testing.T) {
	repo, ir, issID := promoteFixture(t, "contention target")

	lockPath := filepath.Join(ir, lockFilename)
	fd, err := syscall.Open(lockPath, syscall.O_CREAT|syscall.O_RDWR|syscall.O_NOFOLLOW, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Close(fd)
	if err := syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatalf("could not take the ledger lock for the test: %v", err)
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)

	origTimeout := lockTimeout
	lockTimeout = 200 * time.Millisecond
	defer func() { lockTimeout = origTimeout }()

	_, err = Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: issID})
	if !errors.Is(err, ErrAllocatorContention) {
		t.Fatalf("promote must serialize on the ledger lock, got err=%v", err)
	}
}

// dispositionedReadingFixture ingests one detection item and answers it, so the
// promote path has an item that has actually been dispositioned.
func dispositionedReadingFixture(t *testing.T) (repo, ir, item string) {
	t.Helper()
	repo, ir, item = readingFixture(t, "detection")
	if _, err := Disposition(DispositionRequest{
		RepoRoot: repo, IssuesRoot: ir, Item: item,
		State: issueschema.DispositionAccepted, Grounds: "the tension is real and worth acting on",
	}); err != nil {
		t.Fatalf("Disposition: %v", err)
	}
	return repo, ir, item
}

// Item-to-intent without a disposition is the collapse this record family exists
// to prevent: it would make the action the answer, and leave nothing able to show
// that the finding was ever weighed. promote refuses, and names the collapse.
func TestPromoteRefusesUndispositionedReadingItem(t *testing.T) {
	repo, ir, item := readingFixture(t, "detection")
	before := draftCount(t, repo)

	_, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: item})
	if err == nil {
		t.Fatal("promoting an undispositioned reading item must be refused")
	}
	if !strings.Contains(err.Error(), item) || !strings.Contains(err.Error(), "disposition") {
		t.Fatalf("the refusal must name the item and the collapse it prevents; got %v", err)
	}
	if after := draftCount(t, repo); after != before {
		t.Fatalf("a refused promote minted a draft (%d -> %d); the probe runs before anything is minted", before, after)
	}
}

// Acceptance is one record; the action is a separate admission, joined by the
// item id stamped forward on promoted_to and back in the draft's promoted_from.
func TestPromoteStampsReadingItemPromotedTo(t *testing.T) {
	repo, ir, item := dispositionedReadingFixture(t)

	res, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: item})
	if err != nil {
		t.Fatalf("Promote(%s): %v", item, err)
	}
	if res.IntentID == "" {
		t.Fatal("promote must mint an intent draft")
	}

	path, err := findReadingItem(ir, item)
	if err != nil {
		t.Fatalf("findReadingItem: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fm, _, err := parseFrontmatterAndBody(string(content))
	if err != nil {
		t.Fatalf("parse reading record: %v", err)
	}
	if got := asString(fm["promoted_to"]); got != res.IntentID {
		t.Fatalf("reading record promoted_to = %q, want %q", got, res.IntentID)
	}

	draft, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(res.IntentPath)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(draft), "promoted_from: "+item) {
		t.Fatalf("the minted draft must carry the back edge promoted_from: %s\n%s", item, draft)
	}

	// A second promote is refused with the existing id, exactly as it is for an
	// issue: the join is one-to-one in both directions.
	if _, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: item}); err == nil {
		t.Fatal("a second promote of one reading item must be refused")
	}
}

// A rejected or declined item has been answered — and the answer was "no". The
// spec's rule is that acceptance is one record and the action it licenses is a
// separate admission; promoting an item whose standing answer refuses it would
// let the action contradict the record it is supposed to follow from, and the
// ledger would then hold both a refusal and the admission it refused. The
// undispositioned refusal exists to stop an action outrunning the answer; this is
// the same rule where the answer has arrived and says no.
func TestPromoteRefusesARefusedReadingItem(t *testing.T) {
	for _, tc := range []struct{ position, state string }{
		{"detection", issueschema.DispositionRejected},
		{"widening", issueschema.DispositionDeclined},
	} {
		t.Run(tc.state, func(t *testing.T) {
			repo, ir, item := readingFixture(t, tc.position)
			if _, err := Disposition(DispositionRequest{
				RepoRoot: repo, IssuesRoot: ir, Item: item,
				State: tc.state, Grounds: "the constraint already covers it",
			}); err != nil {
				t.Fatalf("Disposition: %v", err)
			}
			before := draftCount(t, repo)

			_, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: item})
			if err == nil {
				t.Fatalf("promoting a %s item must be refused", tc.state)
			}
			if !strings.Contains(err.Error(), tc.state) {
				t.Fatalf("the refusal must name the standing answer; got %v", err)
			}
			if after := draftCount(t, repo); after != before {
				t.Fatalf("a refused promote minted a draft (%d -> %d)", before, after)
			}
		})
	}
}

// A held item is directional, not refused: it is still open, and the answer that
// ends it is a superseding disposition. Promoting it would settle by action what
// the hold left open, so it is refused too — and the refusal names the exit
// condition, which is the thing that would have to happen first.
func TestPromoteRefusesAHeldReadingItem(t *testing.T) {
	repo, ir, item := readingFixture(t, "detection")
	if _, err := Disposition(DispositionRequest{
		RepoRoot: repo, IssuesRoot: ir, Item: item,
		State: issueschema.DispositionHeld, ExitCondition: "the closing run returns it again",
	}); err != nil {
		t.Fatalf("Disposition: %v", err)
	}

	_, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: item})
	if err == nil {
		t.Fatal("promoting a held item must be refused")
	}
	if !strings.Contains(err.Error(), "held") {
		t.Fatalf("the refusal must name the standing answer; got %v", err)
	}
}

// The standing answer is read in the pre-flight, but the pre-flight is not where
// the stamp lands. A disposition arriving between the two — a colleague
// superseding an acceptance with a rejection while the mint runs — would leave a
// standing `rejected` beside a `promoted_to`: a ledger holding both a refusal and
// the admission it refused, which is the state the refusal exists to prevent.
//
// So the state is recomputed inside the locked closure, where nothing can land
// after it. The hook below is that window, forced open.
func TestPromoteRechecksTheStandingStateUnderTheLock(t *testing.T) {
	repo, ir, item := dispositionedReadingFixture(t)
	standing, err := standingDispositions(filepath.Join(ir, issueschema.DispositionsDir, item))
	if err != nil || len(standing) != 1 {
		t.Fatalf("fixture: standing = %v, err = %v", standing, err)
	}

	// Between the pre-flight and the stamp, the acceptance is superseded by a
	// rejection.
	orig := beforeStampHook
	beforeStampHook = func() {
		beforeStampHook = nil
		if _, err := Disposition(DispositionRequest{
			RepoRoot: repo, IssuesRoot: ir, Item: item,
			State: issueschema.DispositionRejected, Grounds: "the constraint already covers it",
			Supersedes: standing[0],
		}); err != nil {
			t.Errorf("superseding disposition: %v", err)
		}
	}
	t.Cleanup(func() { beforeStampHook = orig })

	_, err = Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: item})
	if err == nil {
		t.Fatal("a promote whose standing answer became a rejection must be refused")
	}
	if !strings.Contains(err.Error(), issueschema.DispositionRejected) {
		t.Fatalf("the refusal must name the standing answer it read under the lock; got %v", err)
	}

	// And nothing was stamped: the record must not carry a promoted_to it was
	// refused for.
	path, err := findReadingItem(ir, item)
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "promoted_to") {
		t.Fatalf("the reading record was stamped despite the refusal:\n%s", content)
	}
}

// TestPromoteStampsExtractedFromRecord proves the one arrival path a command
// derives from what it did rather than from what it was told: promote is the one
// shipped verb that derives a record from another record, so the draft it mints
// says so. No flag carries the value, and the issue's own origin is untouched —
// promotion does not change where the issue came from.
func TestPromoteStampsExtractedFromRecord(t *testing.T) {
	repo, ir, issID := promoteFixture(t, "an observation worth graduating")
	res, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: issID})
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(repo, res.IntentPath))
	if err != nil {
		t.Fatal(err)
	}
	fields := frontmatter.Fields(strings.Split(string(data), "\n"))
	if got := fields["origin"].Value; got != "extracted-from-record" {
		t.Errorf("promoted draft origin = %q, want extracted-from-record", got)
	}
	if got := fields["production_mode"].Value; got != "hand-written" {
		t.Errorf("promoted draft production_mode = %q, want hand-written", got)
	}
	if got := readIssue(t, ir, issID); got.ID != issID {
		t.Fatalf("issue re-read as %+v", got)
	}
	fm := readLedgerFrontmatter(t, ir, issID)
	if fm["origin"] != "researcher-authored" {
		t.Errorf("the source issue's origin moved to %v; promotion must not rewrite it", fm["origin"])
	}
}
