package capture

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/intentdriven/abcd/internal/core/intent"
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

	res, err := Promote(PromoteRequest{Grounds: testGrounds, RepoRoot: repo, IssuesRoot: ir, ID: issID})
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
			_, err := Resolve(ResolveRequest{Grounds: testGrounds, RepoRoot: repo, IssuesRoot: ir, ID: id, Resolution: "done", Impact: "fix"})
			return err
		}, StateResolved},
		{"wontfix", func(repo, ir, id string) error {
			_, err := Wontfix(WontfixRequest{Grounds: "declined: we expect this to stay out of scope for the foreseeable cycle", RepoRoot: repo, IssuesRoot: ir, ID: id, Reason: "out of scope"})
			return err
		}, StateWontfix},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, ir, issID := promoteFixture(t, "an observation that grew up after being closed")
			if err := tc.move(repo, ir, issID); err != nil {
				t.Fatal(err)
			}
			res, err := Promote(PromoteRequest{Grounds: testGrounds, RepoRoot: repo, IssuesRoot: ir, ID: issID})
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
	first, err := Promote(PromoteRequest{Grounds: testGrounds, RepoRoot: repo, IssuesRoot: ir, ID: issID})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Promote(PromoteRequest{Grounds: testGrounds, RepoRoot: repo, IssuesRoot: ir, ID: issID})
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
		if _, err := Promote(PromoteRequest{Grounds: testGrounds, RepoRoot: repo, IssuesRoot: ir, ID: bad}); err == nil {
			t.Fatalf("Promote(%q) must fail", bad)
		}
		if n := draftCount(t, repo); n != 0 {
			t.Fatalf("Promote(%q) minted a draft on a structural fault", bad)
		}
	}
	// Link mode with an unknown intent: structural fault, nothing written.
	if _, err := Promote(PromoteRequest{Grounds: testGrounds, RepoRoot: repo, IssuesRoot: ir, ID: issID, LinkIntent: "itd-42"}); err == nil {
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

	_, err := Promote(PromoteRequest{Grounds: testGrounds, RepoRoot: repo, IssuesRoot: ir, ID: issID})
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
	res, err := Promote(PromoteRequest{Grounds: testGrounds, RepoRoot: repo, IssuesRoot: ir, ID: issID, LinkIntent: "itd-1"})
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

	_, err = Promote(PromoteRequest{Grounds: testGrounds, RepoRoot: repo, IssuesRoot: ir, ID: issID})
	if !errors.Is(err, ErrAllocatorContention) {
		t.Fatalf("promote must serialize on the ledger lock, got err=%v", err)
	}
}
