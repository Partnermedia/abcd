package site

// Ancestor-symlink containment for the site READ path (gh #487).
//
// `site build` writes through an os.Root at the destination, but the READ side
// resolved repo-relative source paths with fsutil.ReadGuarded(joinRepo(...)),
// whose O_NOFOLLOW binds the LEAF only. A hostile clone that COMMITS a directory
// symlink (git mode 120000) as an ANCESTOR of a composed page — `docs/explanation
// -> /outside` — has that ancestor followed by the kernel during path
// resolution, so the build reads a file the repository does not own and composes
// its bytes into published HTML. This is the same class already contained for
// positioning (containment_test.go), memory and ahoy.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// siteContainmentCanary stands in for whatever a hostile repo is aiming at — an
// SSH key, a credentials file, a host token. If it ever reaches a written page,
// content from outside the repository escaped into the published site.
const siteContainmentCanary = "SUPER-SECRET-SITE-CONTAINMENT-CANARY"

// TestBuildRefusesSymlinkedAncestorRead plants a committed-shape directory
// symlink as an ancestor of the composed pages, pointing at a directory outside
// the repository, and asserts the build never reads through it.
func TestBuildRefusesSymlinkedAncestorRead(t *testing.T) {
	f := newFixture(t)

	// The outside directory the hostile symlink lands in — OUTSIDE the repo, the
	// way `../../.ssh` or an absolute `/etc` target is from a clone.
	outside := filepath.Join(t.TempDir(), "outside-explanation")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}

	// Copy the four composed explanation pages into the outside directory so a
	// build that follows the symlink still finds a full, valid page set and
	// SUCCEEDS — which is what makes the leak observable rather than a plain
	// missing-page failure. The hero page is poisoned with the canary in its
	// first paragraph, which the hero renders into the landing page's lede.
	realDir := filepath.Join(f.Root(), "docs", "explanation")
	entries, err := os.ReadDir(realDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(realDir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if e.Name() == "rationale.md" {
			// Prepend the canary as the first paragraph of the H1 section, so the
			// hero's lede carries it straight into index.html.
			data = []byte(strings.Replace(string(data),
				"# Who this is for\n",
				"# Who this is for\n\n"+siteContainmentCanary+" leaked from outside the repository.\n",
				1))
		}
		if err := os.WriteFile(filepath.Join(outside, e.Name()), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Replace the real directory with a symlink to the outside one — exactly what
	// a checkout of a mode-120000 directory symlink materialises in the working
	// tree the build reads.
	if err := os.RemoveAll(realDir); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, realDir); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(t.TempDir(), "site")
	res, buildErr := Build(Request{RepoRoot: f.Root(), OutDir: out, Stamp: fixtureStamp})

	// The security invariant: the out-of-repo canary must never reach a written
	// page. On unfixed code the build follows the symlinked ancestor, succeeds,
	// and index.html carries the canary — this is the watched failure.
	if breach := treeContainsCanary(t, out); breach != "" {
		t.Fatalf("containment breach: the out-of-repo canary reached %s — the symlinked ancestor was followed and its bytes were published", breach)
	}

	// And the build must REFUSE rather than silently render a site missing its
	// hero page: the contained read returns an error the composer surfaces.
	if buildErr == nil {
		t.Fatalf("expected the build to refuse the symlinked-ancestor read, but it succeeded: %+v", res)
	}
}

// mustOpenRoot opens dir as an os.Root for a test that constructs an assetPipe
// directly, and closes it at cleanup.
func mustOpenRoot(t *testing.T, dir string) *os.Root {
	t.Helper()
	r, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { r.Close() })
	return r
}

// treeContainsCanary walks a built site tree and returns the first file whose
// bytes contain the canary, or "" if none do (including an absent tree).
func treeContainsCanary(t *testing.T, dir string) string {
	t.Helper()
	hit := ""
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		if strings.Contains(string(data), siteContainmentCanary) {
			hit = path
			return filepath.SkipAll
		}
		return nil
	})
	return hit
}
