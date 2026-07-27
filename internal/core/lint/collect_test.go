package lint

import (
	"testing"
)

// TestCollectCitedURLs pins the refresh verb's contract with the gate: the set
// of URLs the refresh fetches is EXACTLY the set the baseline rule enforces,
// because both come out of this one collector. A second scraper with its own
// idea of what a citation is would produce a baseline the gate rejects.
func TestCollectCitedURLs(t *testing.T) {
	root := citeRepo(t, map[string]string{
		"docs/a.md": "# A\n\nA claim.[^a] A body link to https://prose.example.org/x is not a citation.\n\n" +
			"[^a]: Source A, https://example.org/a.\n",
		"docs/b.md": "# B\n\nTwo claims.[^b][^c]\n\n" +
			"[^b]: Source B, https://example.org/b.\n" +
			"[^c]: Source A again, https://example.org/a.\n",
		"docs/fenced.md": "# Fenced\n\n```\n[^x]: https://fenced.example.org/x\n```\n",
		"README.md":      "# Root\n\nA claim.[^r]\n\n[^r]: Source R, https://example.org/r.\n",
	})
	cfg := Config{Roots: []string{"docs", "README.md"}}

	refs, err := CollectCitedURLs(cfg, root)
	if err != nil {
		t.Fatalf("CollectCitedURLs: %v", err)
	}

	var got []string
	for _, r := range refs {
		got = append(got, r.URL)
	}
	want := []string{
		"https://example.org/a",
		"https://example.org/b",
		"https://example.org/r",
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v (sorted, deduplicated)", got, want)
		}
	}

	// Every URL carries at least one citation site, so a refresh can tell the
	// maintainer WHERE a blocked link is cited without a second walk.
	for _, r := range refs {
		if len(r.Sites) == 0 {
			t.Errorf("%s has no citation site", r.URL)
		}
		for _, s := range r.Sites {
			if s.File == "" || s.Line <= 0 {
				t.Errorf("%s has a malformed site %+v", r.URL, s)
			}
		}
	}
	// The URL cited twice reports both sites, in file order.
	for _, r := range refs {
		if r.URL != "https://example.org/a" {
			continue
		}
		if len(r.Sites) != 2 || r.Sites[0].File != "docs/a.md" || r.Sites[1].File != "docs/b.md" {
			t.Errorf("duplicate citation sites = %+v, want docs/a.md then docs/b.md", r.Sites)
		}
	}
}

// TestCollectCitedURLsMissingRootIsNotAnError mirrors Lint: a configured root
// that does not exist is skipped, so a repo that has not written docs/ yet can
// still run the refresh.
func TestCollectCitedURLsMissingRootIsNotAnError(t *testing.T) {
	root := citeRepo(t, map[string]string{"docs/a.md": "# A\n"})
	refs, err := CollectCitedURLs(Config{Roots: []string{"docs", "nope"}}, root)
	if err != nil {
		t.Fatalf("CollectCitedURLs: %v", err)
	}
	if len(refs) != 0 {
		t.Fatalf("got %+v, want no citations", refs)
	}
}
