package capture

import (
	"os"
	"strings"
	"testing"
)

// rewriteIssue re-reads an issue file, applies fn to its text and writes it
// back. It is the hand edit a contributor-supplied record arrives with: capture
// writes none of the constructs these tests inject, and the point is that none
// of them may make the ledger's own triage routes lie.
func rewriteIssue(t *testing.T, ir, issID string, fn func(string) string) {
	t.Helper()
	path, _, err := findIssue(ir, issID)
	if err != nil {
		t.Fatalf("findIssue(%s): %v", issID, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(fn(string(data))), 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestGroundsLandInTheBodyNotAFrontmatterComment is the writer/reader scope
// agreement (iss-2608301805069999). parseFrontmatterBlock skips a line whose
// trimmed form starts with `#`, so `# Grounds` is a legal YAML comment inside
// the frontmatter block — and an ATX heading pattern matches it as the section
// heading. With the append judged over the WHOLE FILE the bullet landed in that
// pseudo-section, the whole-file read-back agreed, the verb reported success,
// and issueFromFrontmatter — which reads the BODY — found no grounds at all.
func TestGroundsLandInTheBodyNotAFrontmatterComment(t *testing.T) {
	repo, ir, issID := promoteFixture(t, "the loader drops rules silently when the config is stale")
	rewriteIssue(t, ir, issID, func(s string) string {
		return strings.Replace(s, "---\n", "---\n# Grounds\n", 1)
	})

	if _, err := Resolve(ResolveRequest{
		RepoRoot: repo, IssuesRoot: ir, ID: issID,
		Resolution: "closed by the fix under review", Impact: "fix", Grounds: testGrounds,
	}); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	iss := readIssue(t, ir, issID)
	if len(iss.Grounds) != 1 {
		t.Fatalf("the resolved record carries %d grounds entr(ies) the reader can see, want 1: %+v",
			len(iss.Grounds), iss.Grounds)
	}
	if !strings.Contains(iss.Body, "## Grounds") {
		t.Fatalf("the entry did not land in the body:\n%s", iss.Body)
	}
}
