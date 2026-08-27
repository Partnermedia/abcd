package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCollectCitedURLsCapsAPageRead is the bound containment cannot supply. A
// committed page that is simply enormous is INSIDE the repository and resolves
// to itself, so every path check passes and an unguarded read pulls the whole
// thing into memory. The collector feeds a network fetch, so it reads through
// the guarded primitive that already exists for exactly this.
func TestCollectCitedURLsCapsAPageRead(t *testing.T) {
	root := citeRepo(t, map[string]string{"docs/a.md": "# A\n"})
	huge := make([]byte, citationPageSizeLimit+1)
	for i := range huge {
		huge[i] = 'x'
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "huge.md"), huge, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := CollectCitedURLs(Config{Roots: []string{"docs"}}, root); err == nil {
		t.Fatal("CollectCitedURLs read a page past the size cap")
	}
}

// TestCollectCitedURLsRefusesANonRegularPage covers the other half of the same
// guard: a page that is a device rather than a file. Here containment already
// refuses it (a device resolves outside the repo), so this pins the combination
// rather than the cap — neither check alone should be relied on.
func TestCollectCitedURLsRefusesANonRegularPage(t *testing.T) {
	if _, err := os.Stat("/dev/zero"); err != nil {
		t.Skip("no /dev/zero on this platform")
	}
	root := citeRepo(t, map[string]string{"docs/a.md": "# A\n"})
	if err := os.Symlink("/dev/zero", filepath.Join(root, "docs", "boom.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := CollectCitedURLs(Config{Roots: []string{"docs"}}, root)
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("CollectCitedURLs read a character device as a page")
		}
	case <-time.After(20 * time.Second):
		t.Fatal("CollectCitedURLs hung reading a character device")
	}
}

// TestCollectCitedURLsRefusesAnEscapingRoot closes an exfiltration path this
// branch opened. Reading outside the repo via `roots` was inert while the lint
// only reported findings; now the collector feeds a LIVE FETCHER and its results
// are persisted into `.abcd/citations-baseline.json`, a file the workflow expects
// the maintainer to commit and push. A `roots` of "../private" would fetch every
// URL found in a sibling directory and write them verbatim into that artifact.
func TestCollectCitedURLsRefusesAnEscapingRoot(t *testing.T) {
	root := citeRepo(t, map[string]string{"docs/a.md": "# A\n"})
	for _, bad := range []string{"../private", "/etc", "docs/../../elsewhere"} {
		if _, err := CollectCitedURLs(Config{Roots: []string{bad}}, root); err == nil {
			t.Errorf("CollectCitedURLs accepted a root outside the repository: %q", bad)
		}
	}
	for _, ok := range []string{"docs", "./docs"} {
		if _, err := CollectCitedURLs(Config{Roots: []string{ok}}, root); err != nil {
			t.Errorf("CollectCitedURLs refused a contained root %q: %v", ok, err)
		}
	}
}

// TestCollectCitedURLsRefusesASymlinkedFile closes the half the root-level check
// misses. Containing the configured ROOT does nothing about a symlinked page
// INSIDE it: WalkDir yields a symlinked .md as an ordinary file and os.ReadFile
// follows it, so `docs/leak.md -> ../private/notes.md` is read, its URLs are
// fetched by the live checker, and they are written verbatim into the committed
// baseline — the exact harm the root check was added to prevent.
func TestCollectCitedURLsRefusesASymlinkedFile(t *testing.T) {
	outside := t.TempDir()
	secret := filepath.Join(outside, "notes.md")
	if err := os.WriteFile(secret,
		[]byte("# S\n\nA claim.[^a]\n\n[^a]: S, https://secret.example.invalid/leak\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	root := citeRepo(t, map[string]string{"docs/a.md": "# A\n"})
	if err := os.Symlink(secret, filepath.Join(root, "docs", "leak.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	refs, err := CollectCitedURLs(Config{Roots: []string{"docs"}}, root)
	if err == nil {
		for _, r := range refs {
			if r.URL == "https://secret.example.invalid/leak" {
				t.Fatal("a symlinked page leaked an outside-repo URL into the cited set")
			}
		}
		t.Fatal("CollectCitedURLs read a page that resolves outside the repository")
	}
}

// TestCollectCitedURLsAllowsAnInRepoSymlink keeps the containment about WHERE a
// path lands, not about symlinks being suspicious — this repo's own
// CLAUDE.md -> AGENTS.md bridge must keep working.
func TestCollectCitedURLsAllowsAnInRepoSymlink(t *testing.T) {
	root := citeRepo(t, map[string]string{
		"docs/real.md": "# R\n\nA claim.[^a]\n\n[^a]: S, https://example.org/a\n",
	})
	if err := os.Symlink(filepath.Join(root, "docs", "real.md"), filepath.Join(root, "docs", "bridge.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	refs, err := CollectCitedURLs(Config{Roots: []string{"docs"}}, root)
	if err != nil {
		t.Fatalf("CollectCitedURLs refused an in-repo symlink: %v", err)
	}
	if len(refs) != 1 || refs[0].URL != "https://example.org/a" {
		t.Fatalf("refs = %+v, want the one cited URL", refs)
	}
}

// TestCollectCitedURLsRefusesASymlinkedRoot is the same containment against the
// filesystem rather than the string.
func TestCollectCitedURLsRefusesASymlinkedRoot(t *testing.T) {
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.md"),
		[]byte("# S\n\nA claim.[^a]\n\n[^a]: S, https://secret.example/token\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	root := citeRepo(t, map[string]string{"docs/a.md": "# A\n"})
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	refs, err := CollectCitedURLs(Config{Roots: []string{"escape"}}, root)
	if err == nil {
		t.Fatalf("CollectCitedURLs walked out of the repo through a symlink and found %+v", refs)
	}
}

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

// TestCollectCitedURLsMissingRootIsLoud mirrors Lint: a configured root that does
// not resolve is misconfiguration, not a clean tree. It used to be skipped, but a
// scope-collapsed root then collects zero cited URLs and the refresh's
// wholesale-failure guard writes an empty baseline, dropping every entry
// including human confirm receipts — so an unresolvable root must fail loud
// (GitHub #360).
func TestCollectCitedURLsMissingRootIsLoud(t *testing.T) {
	root := citeRepo(t, map[string]string{"docs/a.md": "# A\n"})
	_, err := CollectCitedURLs(Config{Roots: []string{"docs", "nope"}}, root)
	if err == nil {
		t.Fatal("CollectCitedURLs must fail loudly on a configured root that does not exist")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Errorf("the loud error should name the unresolved root: %v", err)
	}
}
