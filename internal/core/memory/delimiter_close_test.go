package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFrontmatterCloseToleratesTrailingWhitespace pins the two ends of a memory
// page's frontmatter block to the SAME rule. The opening delimiter has always
// been matched after a whitespace trim; the closing one was compared byte-exact,
// so a single trailing space ("--- ") made the parser report "frontmatter not
// terminated" and every reader fall back to its empty/permissive default —
// silently dropping the page's `source:` provenance. The canonical primitive
// (internal/core/frontmatter.Fields) closes such a block by trimming spaces and
// tabs off the delimiter; memory's three parsers must agree with it.
func TestFrontmatterCloseToleratesTrailingWhitespace(t *testing.T) {
	for _, close := range []string{"--- ", "---\t", "---  \t "} {
		t.Run(strings.ReplaceAll(strings.ReplaceAll(close, " ", "_"), "\t", "TAB"), func(t *testing.T) {
			page := "---\ntopic: q3-revenue\nsource:\n  class: external_pdf\n  source_hash: " +
				strings.Repeat("a", 64) + "\n" + close + "\nDistilled from a licensed PDF.\n"

			fm, err := parseFrontmatter(page)
			if err != nil {
				t.Fatalf("parseFrontmatter rejected a close carrying trailing whitespace (%q): %v", close, err)
			}
			src, ok := fm["source"].(map[string]any)
			if !ok {
				t.Fatalf("provenance dropped: source = %#v", fm["source"])
			}
			if src["class"] != "external_pdf" {
				t.Fatalf("source.class = %#v, want external_pdf", src["class"])
			}

			region, body, err := splitFileFrontmatter(page)
			if err != nil {
				t.Fatalf("splitFileFrontmatter rejected the same close (%q): %v", close, err)
			}
			if !strings.Contains(region, "class: external_pdf") {
				t.Fatalf("region lost the source block: %q", region)
			}
			if strings.Contains(region, "Distilled from") {
				t.Fatalf("body leaked into the frontmatter region: %q", region)
			}
			if !strings.Contains(body, "Distilled from a licensed PDF.") {
				t.Fatalf("body = %q", body)
			}
		})
	}

	// frontmatterKeyLine must stop at the decorated close too — otherwise a body
	// line that happens to start with "source:" is reported as the frontmatter
	// key's position, mislocating every finding on the page.
	page := "---\nsource:\n  class: external_pdf\n--- \nsource: not the frontmatter key\n"
	if got := frontmatterKeyLine(page, "source"); got != 2 {
		t.Fatalf("frontmatterKeyLine = %d, want 2 (the key inside the block, not the decoy after the close)", got)
	}
}

// TestLintBlocksExternalPageWithDecoratedClose is the consequence test: the
// blocking ML001 licence gate must still fire on an unlicensed external_* page
// whose closing delimiter carries a trailing space. Before the fix the page was
// classed "not a typed memory page" and skipped by the crawl entirely, so the
// blocker never evaluated and lint exited 0.
func TestLintBlocksExternalPageWithDecoratedClose(t *testing.T) {
	repo := t.TempDir()
	mem := Dir(repo)
	if err := os.MkdirAll(mem, 0o755); err != nil {
		t.Fatal(err)
	}
	page := "---\nsource:\n  class: external_pdf\n  source_hash: " + strings.Repeat("a", 64) +
		"\ntopic_hash: " + strings.Repeat("b", 64) + "\n--- \n\n# A quoted claim\n"
	if err := os.WriteFile(filepath.Join(mem, "topic_auth_bad.md"), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
	lr, err := Lint(LintRequest{RepoRoot: repo, Now: fixedNow})
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	found := false
	for _, f := range lr.Findings {
		if f.Code == "ML001" && f.Severity == "blocker" {
			found = true
		}
	}
	if !found || lr.ExitCode != 1 {
		t.Fatalf("unlicensed external page with a trailing-space close passed lint: exit=%d findings=%+v", lr.ExitCode, lr.Findings)
	}
}
