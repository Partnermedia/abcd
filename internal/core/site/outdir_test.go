package site

// The output-directory gate (GHSA-fpf2-pg82-72rj, CWE-59).
//
// `site build` empties and rewrites the directory named by --out, and `site
// check` reads it. Both took the path at its word: a symlink at the leaf or at
// any ancestor was followed, so a committed link (git mode 120000) named `site`
// — the default --out — redirected the purge and the writes to wherever the
// link pointed; and the `.abcd-site-build` marker was recognised by NAME, so a
// committed marker made any directory later passed to --out "ours" and had its
// other entries removed. These tests are the attack paths, each asserting the
// build refuses AND that nothing outside the lexical destination was touched.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// symlinkOrSkip plants a symlink, skipping on a filesystem that cannot hold
// one: the subject is abcd's refusal, not the platform's symlink support.
func symlinkOrSkip(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
}

// genuineMarker is the marker a build of f writes — the forger's best case,
// byte-identical to the real thing.
func genuineMarker(t *testing.T, f *fixture) []byte {
	t.Helper()
	out := t.TempDir()
	buildFixture(t, f, out)
	data, err := os.ReadFile(filepath.Join(out, siteMarkerName))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// TestBuildRefusesASymlinkedOutputLeaf is attack path 1: a committed symlink
// named `site` pointing outside the checkout. The target is EMPTY, which is the
// case a name-only inspection reads as "nothing here, go ahead" and writes the
// whole tree through the link.
func TestBuildRefusesASymlinkedOutputLeaf(t *testing.T) {
	f := newFixture(t)
	outside := t.TempDir()
	link := filepath.Join(f.Root(), DefaultOutDir)
	symlinkOrSkip(t, outside, link)

	_, err := Build(Request{RepoRoot: f.Root(), OutDir: link, Stamp: fixtureStamp})
	if err == nil {
		t.Fatal("the build wrote through a symlinked output directory")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Errorf("the refusal does not say why: %v", err)
	}
	entries, rerr := os.ReadDir(outside)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if len(entries) != 0 {
		t.Errorf("the refused build wrote %d entries through the link", len(entries))
	}
}

// TestBuildRefusesASymlinkedAncestorOfAnAbsentLeaf is attack path 2: the leaf
// does not exist, so a leaf-only lstat sees nothing, and MkdirAll walks the
// committed ancestor link to create the tree outside the checkout.
func TestBuildRefusesASymlinkedAncestorOfAnAbsentLeaf(t *testing.T) {
	f := newFixture(t)
	outside := t.TempDir()
	symlinkOrSkip(t, outside, filepath.Join(f.Root(), "parent-link"))
	out := filepath.Join(f.Root(), "parent-link", DefaultOutDir)

	_, err := Build(Request{RepoRoot: f.Root(), OutDir: out, Stamp: fixtureStamp})
	if err == nil {
		t.Fatal("the build created its tree through a symlinked ancestor")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Errorf("the refusal does not say why: %v", err)
	}
	if _, serr := os.Lstat(filepath.Join(outside, DefaultOutDir)); serr == nil {
		t.Error("the refused build still created the output tree outside the checkout")
	}
}

// TestBuildRefusesToPurgeThroughACommittedMarker is attack path 3: a marker
// committed into a tracked directory, byte-identical to a genuine one, next to
// a file that is not the build's. A name-only check reads the directory as
// "ours" and removes the file; the purge must never delete what abcd did not
// write, and a directory git tracks anything in is never abcd's.
func TestBuildRefusesToPurgeThroughACommittedMarker(t *testing.T) {
	f := newFixture(t)
	marker := genuineMarker(t, f)
	f.writeBytes("public/"+siteMarkerName, marker)
	f.write("public/precious.txt", "someone else's work")
	f.commitAt("2026-02-12T09:00:00+00:00", "chore: plant a marker", "None")

	out := filepath.Join(f.Root(), "public")
	_, err := Build(Request{RepoRoot: f.Root(), OutDir: out, Stamp: fixtureStamp})
	if err == nil {
		t.Fatal("the build purged a tracked directory on the strength of a committed marker")
	}
	got, rerr := os.ReadFile(filepath.Join(out, "precious.txt"))
	if rerr != nil || string(got) != "someone else's work" {
		t.Errorf("the refused build removed the file that was not its own: %q, %v", got, rerr)
	}
}

// TestBuildRefusesTheRepositoryRootAsOutput is attack path 3 at its worst: the
// marker committed at the root and `--out .`. The purge would remove every
// other entry — `.git` included.
func TestBuildRefusesTheRepositoryRootAsOutput(t *testing.T) {
	f := newFixture(t)
	marker := genuineMarker(t, f)
	f.writeBytes(siteMarkerName, marker)
	f.commitAt("2026-02-12T09:00:00+00:00", "chore: plant a marker at the root", "None")

	_, err := Build(Request{RepoRoot: f.Root(), OutDir: f.Root(), Stamp: fixtureStamp})
	if err == nil {
		t.Fatal("the build emptied the repository it was rendering")
	}
	for _, rel := range []string{".git", ManifestRelPath} {
		if _, serr := os.Lstat(filepath.Join(f.Root(), rel)); serr != nil {
			t.Errorf("the refused build removed %s: %v", rel, serr)
		}
	}
}

// TestBuildRefusesAnotherRepositorysOutput: the marker names the repository
// that wrote it, so a directory another repository's build filled is foreign —
// not a security boundary (a forger can copy the name), but the rule that keeps
// two projects sharing one --out from purging each other's site.
func TestBuildRefusesAnotherRepositorysOutput(t *testing.T) {
	a := newFixture(t)
	b := newFixtureRootedAt(t, "2026-01-04T09:00:00+00:00")
	rootOf := func(f *fixture) string { return f.gitOut("rev-list", "--max-parents=0", "HEAD") }
	if rootOf(a) == rootOf(b) {
		t.Fatal("the two fixtures share a root commit, so this test would prove nothing")
	}
	out := t.TempDir()
	buildFixture(t, a, out)

	_, err := Build(Request{RepoRoot: b.Root(), OutDir: out, Stamp: fixtureStamp})
	if err == nil {
		t.Fatal("a build of one repository purged another repository's site")
	}
	if _, serr := os.Stat(filepath.Join(out, "index.html")); serr != nil {
		t.Errorf("the refused build removed the other repository's page: %v", serr)
	}
}

// TestCheckRefusesASymlinkedOutputDir: the gate reads the directory it is
// handed, and a committed `site` link would have it grade — and, absent an
// index.html, build into — whatever the link points at.
func TestCheckRefusesASymlinkedOutputDir(t *testing.T) {
	f := newFixture(t)
	built := t.TempDir()
	buildFixture(t, f, built)
	link := filepath.Join(f.Root(), DefaultOutDir)
	symlinkOrSkip(t, built, link)

	_, err := Check(CheckRequest{RepoRoot: f.Root(), OutDir: link})
	if err == nil {
		t.Fatal("the check read through a symlinked output directory")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Errorf("the refusal does not say why: %v", err)
	}
}
