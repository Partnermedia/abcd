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

	it, err := CreateFromText(root, "I want users to feel the card respects their time")
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

	it, err := CreateFromText(root, "another product intent")
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
		if _, err := CreateFromText(root, in); err == nil {
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
	if _, err := CreateFromText(root, "seeded from a quoted-text capture"); err != nil {
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
