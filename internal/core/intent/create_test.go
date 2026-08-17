package intent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/frontmatter"
	"github.com/REPPL/abcd-cli/internal/core/lint"
)

// TestCreateFromTextSeedsDraft is the itd-46 AC1 core: a quoted-text create files
// a new drafts/itd-N-<slug>.md seeded from the text, with the canonical draft
// frontmatter set, and the seeded body carries the text.
func TestCreateFromTextSeedsDraft(t *testing.T) {
	root := t.TempDir()

	it, _, err := CreateFromText(root, "I want users to feel the card respects their time", "")
	if err != nil {
		t.Fatalf("CreateFromText: %v", err)
	}
	if it.ID != "itd-1" {
		t.Fatalf("first minted id = %q, want itd-1", it.ID)
	}
	if it.Bucket != BucketDrafts {
		t.Fatalf("bucket = %q, want drafts", it.Bucket)
	}
	if !slugRe.MatchString(it.Slug) {
		t.Fatalf("slug %q is not kebab-case", it.Slug)
	}
	if err := Validate(it); err != nil {
		t.Fatalf("created intent fails Validate: %v", err)
	}
	abs := filepath.Join(root, it.Path)
	data, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("created file unreadable: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "I want users to feel the card respects their time") {
		t.Fatalf("seeded body missing the quoted text:\n%s", body)
	}
	// Canonical draft frontmatter: spec_id null, kind null/standalone/bundle-member.
	fields := frontmatter.Fields(strings.Split(body, "\n"))
	if fields["id"].Value != "itd-1" {
		t.Fatalf("frontmatter id = %q, want itd-1", fields["id"].Value)
	}
	if !frontmatter.IsNull(fields["spec_id"].Value) {
		t.Fatalf("drafts spec_id must be null, got %q", fields["spec_id"].Value)
	}
}

// TestCreateFromTextAllocatesNextID proves the allocator mints max+1 across every
// bucket, not always itd-1.
func TestCreateFromTextAllocatesNextID(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, draftsDir+"/itd-5-alpha.md", draftWithAC("itd-5", "alpha"))
	writeFile(t, root, plannedDir+"/itd-9-beta.md",
		"---\nid: itd-9\nslug: beta\nspec_id: spc-1\nkind: standalone\n---\n# beta\n")

	it, _, err := CreateFromText(root, "another product intent", "")
	if err != nil {
		t.Fatalf("CreateFromText: %v", err)
	}
	if it.ID != "itd-10" {
		t.Fatalf("minted id = %q, want itd-10 (max 9 + 1)", it.ID)
	}
}

// TestCreateFromTextRefusesEmpty proves empty/whitespace text is refused and
// nothing is written (unrecognized/empty input never writes).
func TestCreateFromTextRefusesEmpty(t *testing.T) {
	root := t.TempDir()
	for _, in := range []string{"", "   ", "\t\n"} {
		if _, _, err := CreateFromText(root, in, ""); err == nil {
			t.Fatalf("CreateFromText(%q) must be refused", in)
		}
	}
	// No drafts file appeared.
	if entries, _ := os.ReadDir(filepath.Join(root, draftsDir)); len(entries) != 0 {
		t.Fatalf("empty-text create wrote %d files, want 0", len(entries))
	}
}

// TestCreateFromTextPassesRecordLint runs the real intent_lifecycle record-lint
// over a freshly seeded draft — the "abcd audit stays green" guarantee.
func TestCreateFromTextPassesRecordLint(t *testing.T) {
	root := t.TempDir()
	if _, _, err := CreateFromText(root, "seeded from a quoted-text capture", ""); err != nil {
		t.Fatalf("CreateFromText: %v", err)
	}
	cfg := lint.Config{
		Roots: []string{".abcd/development"},
		Rules: map[string]lint.RuleConfig{
			"intent_lifecycle": {Enabled: true, Severity: "blocker", IntentsDir: "intents"},
		},
	}
	findings, err := lint.Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	for _, fnd := range findings {
		if fnd.RuleID == "intent_lifecycle" {
			t.Fatalf("seeded draft violates intent_lifecycle: %s:%d %s", fnd.File, fnd.Line, fnd.Message)
		}
	}
}

// TestCreateDraftPromotedFromRoundTrip: a draft minted through the promote
// path carries the promoted_from back-edge in its frontmatter, and the intent
// reader parses it back (two-sided edge, spc-24).
func TestCreateDraftPromotedFromRoundTrip(t *testing.T) {
	root := t.TempDir()

	it, _, err := CreateDraft(root, DraftOptions{
		Slug:         "an-issue-that-grew-up",
		Title:        "An issue that grew up",
		SeedBody:     "Graduated from `iss-7`: an issue that grew up. Read that issue record for the source observation.",
		PromotedFrom: "iss-7",
	})
	if err != nil {
		t.Fatalf("CreateDraft: %v", err)
	}
	if it.PromotedFrom != "iss-7" {
		t.Fatalf("created intent PromotedFrom = %q, want iss-7", it.PromotedFrom)
	}
	c, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got, ok := c.Lookup(it.ID)
	if !ok {
		t.Fatalf("minted draft %s not found by Load", it.ID)
	}
	if got.PromotedFrom != "iss-7" {
		t.Fatalf("parsed-back PromotedFrom = %q, want iss-7", got.PromotedFrom)
	}
	// Absent on every existing record: a draft minted from text has none.
	plain, _, err := CreateFromText(root, "a plain quoted-text draft", "")
	if err != nil {
		t.Fatal(err)
	}
	c, err = Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := c.Lookup(plain.ID); got.PromotedFrom != "" {
		t.Fatalf("text-created draft must carry no promoted_from, got %q", got.PromotedFrom)
	}
}

// TestCreateDraftValidatesInputs: a promote-path mint refuses a malformed slug
// or promoted_from before any path is built.
func TestCreateDraftValidatesInputs(t *testing.T) {
	root := t.TempDir()
	if _, _, err := CreateDraft(root, DraftOptions{Slug: "../evil", Title: "x", SeedBody: "y"}); err == nil {
		t.Fatalf("CreateDraft must refuse a non-kebab slug")
	}
	if _, _, err := CreateDraft(root, DraftOptions{Slug: "ok-slug", Title: "x", SeedBody: "y", PromotedFrom: "itd-3"}); err == nil {
		t.Fatalf("CreateDraft must refuse a promoted_from that is not an iss-N id")
	}
	if entries, err := os.ReadDir(filepath.Join(root, IntentsRelDir, BucketDrafts)); err == nil && len(entries) > 0 {
		t.Fatalf("a refused CreateDraft wrote %d file(s)", len(entries))
	}
}
