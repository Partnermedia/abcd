package capture

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// ledger returns (repoRoot, issuesRoot) rooted in a temp dir, avoiding git
// discovery by supplying both explicitly (resolveRoots contract B).
func ledger(t *testing.T) (string, string) {
	t.Helper()
	repo := t.TempDir()
	return repo, filepath.Join(repo, LedgerRelPath)
}

func pinDate(t *testing.T, d string) {
	t.Helper()
	prev := nowDate
	nowDate = func() string { return d }
	t.Cleanup(func() { nowDate = prev })
}

func TestCaptureAppendAndReadBack(t *testing.T) {
	pinDate(t, "2026-07-07")
	tests := []struct {
		name string
		req  CaptureRequest
		want Issue
	}{
		{
			name: "minimal required fields",
			req: CaptureRequest{
				Text: "Something is off.\n", Severity: SeverityMinor,
				Category: "bug", Source: "manual-test", Slug: "Something Off!",
				FoundDuring: "manual smoke",
			},
			want: Issue{
				SchemaVersion: 1, ID: "iss-1", Slug: "something-off",
				Severity: SeverityMinor, Category: "bug", Source: "manual-test",
				FoundDuring: "manual smoke", Created: "2026-07-07",
				Status: StateOpen, Body: "Something is off.\n",
			},
		},
		{
			name: "optional found_at and related ids",
			req: CaptureRequest{
				Text: "b", Severity: SeverityMajor, Category: "drift",
				Source: "agent-finding", Slug: "drifted", FoundDuring: "fn-3 review",
				FoundAt: "internal/x.go", RelatedIntents: []string{"itd-4"},
				RelatedSpecs: []string{"fn-12"},
			},
			want: Issue{
				SchemaVersion: 1, ID: "iss-1", Slug: "drifted",
				Severity: SeverityMajor, Category: "drift", Source: "agent-finding",
				FoundDuring: "fn-3 review", FoundAt: "internal/x.go",
				Created: "2026-07-07", RelatedIntents: []string{"itd-4"},
				RelatedSpecs: []string{"fn-12"}, Status: StateOpen, Body: "b",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo, ir := ledger(t)
			tc.req.RepoRoot, tc.req.IssuesRoot = repo, ir
			res, err := Capture(tc.req)
			if err != nil {
				t.Fatalf("Capture: %v", err)
			}
			if res.ID != tc.want.ID || res.Status != StateOpen {
				t.Fatalf("result = %+v", res)
			}
			// Read back via List and compare the parsed issue (path aside).
			lr, err := List(ListRequest{RepoRoot: repo, IssuesRoot: ir, State: StateOpen})
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if len(lr.Issues) != 1 {
				t.Fatalf("want 1 issue, got %d (skipped=%v)", len(lr.Issues), lr.Skipped)
			}
			got := lr.Issues[0]
			want := tc.want
			want.Path = got.Path // path is env-specific
			if !issueEqual(got, want) {
				t.Fatalf("read-back mismatch:\n got %+v\nwant %+v", got, want)
			}
			if filepath.Base(got.Path) != tc.want.ID+"-"+tc.want.Slug+".md" {
				t.Errorf("filename = %s", filepath.Base(got.Path))
			}
		})
	}
}

func issueEqual(a, b Issue) bool {
	return a.SchemaVersion == b.SchemaVersion && a.ID == b.ID && a.Slug == b.Slug &&
		a.Severity == b.Severity && a.Category == b.Category && a.Source == b.Source &&
		a.FoundDuring == b.FoundDuring && a.FoundAt == b.FoundAt && a.Created == b.Created &&
		a.Updated == b.Updated && a.PromotedTo == b.PromotedTo && a.Resolution == b.Resolution &&
		a.WontfixReason == b.WontfixReason && a.Status == b.Status && a.Body == b.Body &&
		strings.Join(a.RelatedIntents, ",") == strings.Join(b.RelatedIntents, ",") &&
		strings.Join(a.RelatedSpecs, ",") == strings.Join(b.RelatedSpecs, ",") &&
		a.Path == b.Path
}

func TestCaptureAllocatesIncrementingIDs(t *testing.T) {
	pinDate(t, "2026-07-07")
	repo, ir := ledger(t)
	for i := 1; i <= 3; i++ {
		res, err := Capture(CaptureRequest{
			RepoRoot: repo, IssuesRoot: ir, Text: "x", Severity: SeverityNitpick,
			Category: "observation", Source: "manual-test", Slug: "note", FoundDuring: "loop",
		})
		if err != nil {
			t.Fatalf("capture %d: %v", i, err)
		}
		if want := "iss-" + strconv.Itoa(i); res.ID != want {
			t.Fatalf("id = %s want %s", res.ID, want)
		}
	}
}

func TestCaptureForceIDAndDuplicate(t *testing.T) {
	pinDate(t, "2026-07-07")
	repo, ir := ledger(t)
	base := CaptureRequest{
		RepoRoot: repo, IssuesRoot: ir, Text: "x", Severity: SeverityMinor,
		Category: "bug", Source: "manual-test", Slug: "forced", FoundDuring: "migration",
	}
	base.ForceID = "iss-42"
	res, err := Capture(base)
	if err != nil {
		t.Fatalf("forceID capture: %v", err)
	}
	if res.ID != "iss-42" {
		t.Fatalf("id = %s want iss-42", res.ID)
	}
	// Re-forcing the same id must be a duplicate error.
	if _, err := Capture(base); !errors.Is(err, ErrDuplicateIssueID) {
		t.Fatalf("want ErrDuplicateIssueID, got %v", err)
	}
}

func TestCaptureRejectsEmptyFoundDuring(t *testing.T) {
	repo, ir := ledger(t)
	_, err := Capture(CaptureRequest{
		RepoRoot: repo, IssuesRoot: ir, Text: "x", Severity: SeverityMinor,
		Category: "bug", Source: "manual-test", Slug: "s", FoundDuring: "  ",
	})
	if err == nil || !strings.Contains(err.Error(), "found_during") {
		t.Fatalf("want found_during error, got %v", err)
	}
	// No placeholder must be left behind.
	entries, _ := os.ReadDir(filepath.Join(ir, "open"))
	if len(entries) != 0 {
		t.Fatalf("expected empty open/, found %d entries", len(entries))
	}
}

func TestCaptureRejectsBadEnumAndSweepsPlaceholder(t *testing.T) {
	pinDate(t, "2026-07-07")
	repo, ir := ledger(t)
	_, err := Capture(CaptureRequest{
		RepoRoot: repo, IssuesRoot: ir, Text: "x", Severity: "bogus",
		Category: "bug", Source: "manual-test", Slug: "s", FoundDuring: "ctx",
	})
	if !errors.Is(err, ErrMalformedFrontmatter) {
		t.Fatalf("want ErrMalformedFrontmatter, got %v", err)
	}
	entries, _ := os.ReadDir(filepath.Join(ir, "open"))
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			t.Fatalf("placeholder not swept: %s", e.Name())
		}
	}
}

func TestResolveTransition(t *testing.T) {
	pinDate(t, "2026-07-07")
	repo, ir := ledger(t)
	res, err := Capture(CaptureRequest{
		RepoRoot: repo, IssuesRoot: ir, Text: "body", Severity: SeverityMajor,
		Category: "bug", Source: "manual-test", Slug: "fixme", FoundDuring: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	tr, err := Resolve(ResolveRequest{RepoRoot: repo, IssuesRoot: ir, ID: res.ID, Resolution: "patched in fn-9"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if tr.FromStatus != StateOpen || tr.ToStatus != StateResolved {
		t.Fatalf("transition = %+v", tr)
	}
	if _, err := os.Stat(res.Path); !os.IsNotExist(err) {
		t.Errorf("source still present at %s", res.Path)
	}
	lr, _ := List(ListRequest{RepoRoot: repo, IssuesRoot: ir, State: StateResolved})
	if len(lr.Issues) != 1 || lr.Issues[0].Resolution != "patched in fn-9" || lr.Issues[0].Updated != "2026-07-07" {
		t.Fatalf("resolved issue = %+v (skipped=%v)", lr.Issues, lr.Skipped)
	}
}

func TestResolveConflictAndUnknown(t *testing.T) {
	pinDate(t, "2026-07-07")
	repo, ir := ledger(t)
	res, _ := Capture(CaptureRequest{
		RepoRoot: repo, IssuesRoot: ir, Text: "b", Severity: SeverityMinor,
		Category: "bug", Source: "manual-test", Slug: "s", FoundDuring: "t",
	})
	if _, err := Resolve(ResolveRequest{RepoRoot: repo, IssuesRoot: ir, ID: res.ID, Resolution: "done"}); err != nil {
		t.Fatal(err)
	}
	// Already resolved -> conflict.
	if _, err := Resolve(ResolveRequest{RepoRoot: repo, IssuesRoot: ir, ID: res.ID, Resolution: "again"}); !errors.Is(err, ErrTransitionConflict) {
		t.Fatalf("want ErrTransitionConflict, got %v", err)
	}
	// Unknown id.
	if _, err := Wontfix(WontfixRequest{RepoRoot: repo, IssuesRoot: ir, ID: "iss-999", Reason: "nope"}); !errors.Is(err, ErrUnknownIssueID) {
		t.Fatalf("want ErrUnknownIssueID, got %v", err)
	}
}

func TestWontfixTransition(t *testing.T) {
	pinDate(t, "2026-07-07")
	repo, ir := ledger(t)
	res, _ := Capture(CaptureRequest{
		RepoRoot: repo, IssuesRoot: ir, Text: "b", Severity: SeverityMinor,
		Category: "process", Source: "user-observation", Slug: "meh", FoundDuring: "t",
	})
	if _, err := Wontfix(WontfixRequest{RepoRoot: repo, IssuesRoot: ir, ID: res.ID, Reason: "platform constraint"}); err != nil {
		t.Fatalf("Wontfix: %v", err)
	}
	lr, _ := List(ListRequest{RepoRoot: repo, IssuesRoot: ir, State: StateWontfix})
	if len(lr.Issues) != 1 || lr.Issues[0].WontfixReason != "platform constraint" {
		t.Fatalf("wontfix issue = %+v", lr.Issues)
	}
}

func TestListSortsNumericallyAndAll(t *testing.T) {
	pinDate(t, "2026-07-07")
	repo, ir := ledger(t)
	// Force ids out of lexical order: iss-2, iss-10, iss-1.
	for _, id := range []string{"iss-2", "iss-10", "iss-1"} {
		if _, err := Capture(CaptureRequest{
			RepoRoot: repo, IssuesRoot: ir, Text: "b", Severity: SeverityMinor,
			Category: "bug", Source: "manual-test", Slug: "s", FoundDuring: "t", ForceID: id,
		}); err != nil {
			t.Fatal(err)
		}
	}
	lr, err := List(ListRequest{RepoRoot: repo, IssuesRoot: ir, State: StateAll})
	if err != nil {
		t.Fatal(err)
	}
	got := []string{lr.Issues[0].ID, lr.Issues[1].ID, lr.Issues[2].ID}
	want := []string{"iss-1", "iss-2", "iss-10"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v want %v", got, want)
		}
	}
	// Empty State defaults to all.
	if lr2, _ := List(ListRequest{RepoRoot: repo, IssuesRoot: ir}); len(lr2.Issues) != 3 {
		t.Fatalf("empty-state list = %d issues", len(lr2.Issues))
	}
}

func TestListToleratesVirginLedgerAndStrayFiles(t *testing.T) {
	repo, ir := ledger(t)
	// Virgin ledger: no dirs created.
	lr, err := List(ListRequest{RepoRoot: repo, IssuesRoot: ir, State: StateAll})
	if err != nil || len(lr.Issues) != 0 {
		t.Fatalf("virgin list err=%v issues=%d", err, len(lr.Issues))
	}
	if _, err := os.Stat(ir); !os.IsNotExist(err) {
		t.Errorf("List must not create the ledger dir")
	}
	// Stray README is ignored; corrupt iss file is surfaced in Skipped.
	pinDate(t, "2026-07-07")
	if _, err := Capture(CaptureRequest{
		RepoRoot: repo, IssuesRoot: ir, Text: "b", Severity: SeverityMinor,
		Category: "bug", Source: "manual-test", Slug: "ok", FoundDuring: "t",
	}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ir, "open", "README.md"), []byte("stray"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ir, "open", "iss-99-corrupt.md"), []byte("not frontmatter"), 0o644); err != nil {
		t.Fatal(err)
	}
	lr2, _ := List(ListRequest{RepoRoot: repo, IssuesRoot: ir, State: StateOpen})
	if len(lr2.Issues) != 1 {
		t.Fatalf("want 1 valid issue, got %d", len(lr2.Issues))
	}
	if len(lr2.Skipped) != 1 || !strings.Contains(lr2.Skipped[0].Path, "iss-99-corrupt.md") {
		t.Fatalf("want 1 skipped corrupt file, got %+v", lr2.Skipped)
	}
}

func TestStatusCountsAndRecentOpen(t *testing.T) {
	pinDate(t, "2026-07-07")
	repo, ir := ledger(t)
	var ids []string
	for i := 0; i < 3; i++ {
		res, _ := Capture(CaptureRequest{
			RepoRoot: repo, IssuesRoot: ir, Text: "b", Severity: SeverityMinor,
			Category: "bug", Source: "manual-test", Slug: "s", FoundDuring: "t",
		})
		ids = append(ids, res.ID)
	}
	// Resolve the first one.
	if _, err := Resolve(ResolveRequest{RepoRoot: repo, IssuesRoot: ir, ID: ids[0], Resolution: "done"}); err != nil {
		t.Fatal(err)
	}
	st, err := Status(StatusRequest{RepoRoot: repo, IssuesRoot: ir})
	if err != nil {
		t.Fatal(err)
	}
	if st.OpenCount != 2 || st.ResolvedCount != 1 || st.WontfixCount != 0 {
		t.Fatalf("counts open=%d resolved=%d wontfix=%d", st.OpenCount, st.ResolvedCount, st.WontfixCount)
	}
	// Newest first: iss-3 before iss-2.
	if len(st.RecentOpen) != 2 || st.RecentOpen[0].ID != "iss-3" || st.RecentOpen[1].ID != "iss-2" {
		t.Fatalf("recent-open = %v", recentIDs(st.RecentOpen))
	}
}

func recentIDs(issues []Issue) []string {
	out := make([]string, len(issues))
	for i, is := range issues {
		out[i] = is.ID
	}
	return out
}

func TestStatusAndListAreReadOnly(t *testing.T) {
	repo, ir := ledger(t)
	if _, err := Status(StatusRequest{RepoRoot: repo, IssuesRoot: ir}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ir); !os.IsNotExist(err) {
		t.Fatalf("Status must not create the ledger dir")
	}
}

func TestPathUnsafeSymlinkedLedger(t *testing.T) {
	pinDate(t, "2026-07-07")
	repo := t.TempDir()
	real := filepath.Join(repo, "real-issues")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(repo, "linked-issues")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	_, err := Capture(CaptureRequest{
		RepoRoot: repo, IssuesRoot: link, Text: "b", Severity: SeverityMinor,
		Category: "bug", Source: "manual-test", Slug: "s", FoundDuring: "t",
	})
	if !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("want ErrPathUnsafe, got %v", err)
	}
}
