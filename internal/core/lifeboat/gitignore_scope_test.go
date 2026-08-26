package lifeboat

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/gittest"
)

// The walk honours .gitignore by DEFAULT, and reading ignored files is a
// caller's explicit choice (iss-2608241828356533).
//
// disembark packs a lifeboat whose evidence cites source by path:line. The walk
// consulted git nowhere — it read every regular file except a fixed directory
// list — so a file a user had told git to ignore could be quoted into an
// artefact meant to be shared. This repository's own gitignored local tier,
// holding scratch and logs, is not in that list and was therefore in scope.
//
// The default is what a user already believes. The opt-in exists because
// disembark is offered over DEAD and ARCHIVED repositories, where uncommitted
// residue is often the most valuable thing left — refusing to read it would lose
// the case the verb exists for. Both directions are asserted, because either one
// alone would pass on a walk that had simply stopped working.
func TestWalkHonoursGitignoreByDefaultAndWidensOnRequest(t *testing.T) {
	r := gittest.NewRepo(t)
	r.Write(".gitignore", "secret-notes.md\nscratch/\n")
	r.Write("tracked.go", "package a // TODO: tracked\n")
	r.Write("secret-notes.md", "TODO: a private note\n")
	r.Write("scratch/log.txt", "TODO: local scratch\n")
	r.Commit("a repo with ignored files")

	t.Run("default excludes what git ignores", func(t *testing.T) {
		ctx, err := newSourceContext(r.Root())
		if err != nil {
			t.Fatal(err)
		}
		defer ctx.Close()

		got := strings.Join(mustWalk(t, ctx), "\n")
		if !strings.Contains(got, "tracked.go") {
			t.Errorf("the tracked file must still be read; walk = %q", got)
		}
		if strings.Contains(got, "secret-notes.md") {
			t.Error("an ignored file was read by default; a packed lifeboat could cite it by path:line")
		}
		if strings.Contains(got, "scratch/log.txt") {
			t.Error("an ignored DIRECTORY's contents were read by default")
		}
		if ctx.IgnoredAreIncluded() {
			t.Error("the context reports the wide scan when the narrow one ran")
		}
	})

	t.Run("the opt-in widens it", func(t *testing.T) {
		ctx, err := newSourceContext(r.Root())
		if err != nil {
			t.Fatal(err)
		}
		defer ctx.Close()
		IncludeIgnored()(ctx)

		got := strings.Join(mustWalk(t, ctx), "\n")
		for _, want := range []string{"tracked.go", "secret-notes.md", "scratch/log.txt"} {
			if !strings.Contains(got, want) {
				t.Errorf("the opt-in must reach %s — the salvage case is the reason it exists; walk = %q", want, got)
			}
		}
		if !ctx.IgnoredAreIncluded() {
			t.Error("the context must report the wide scan so an adapter can say which one ran")
		}
	})
}

// A tree git knows nothing about must not be narrowed. Answering "ignored" when
// the question cannot be asked would blank a whole section on a repository that
// was never a git checkout — and losing evidence quietly is the failure this
// adapter family exists to avoid.
func TestANonGitTreeIsNotNarrowed(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.md"), []byte("TODO: still a note\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, err := newSourceContext(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	if got := strings.Join(mustWalk(t, ctx), "\n"); !strings.Contains(got, "notes.md") {
		t.Errorf("a non-git tree was narrowed as though git had an opinion; walk = %q", got)
	}
}

func mustWalk(t *testing.T, ctx *SourceContext) []string {
	t.Helper()
	paths, _ := ctx.WalkFiles(".")
	return paths
}

// A listing git could not deliver COMPLETELY narrows nothing.
//
// The set of ignored paths is the complement of one `ls-files` listing, so a
// truncated listing is not a shorter answer — it is an inverted one. Every path
// past the cut is absent from the set and therefore reads as ignored, and a large
// repository would silently drop the tail of its own tree from the scan with
// nothing to show for it.
//
// That is why this reads git through RunCapped (complete-or-error) rather than
// the cached RunLimited path (which truncates and says nothing). The cap is a var
// so this hazard is testable at all: at 16 MiB it is unreachable in a test, and
// an untested refusal path is the one that rots.
func TestATruncatedIgnoreListingNarrowsNothing(t *testing.T) {
	r := gittest.NewRepo(t)
	r.Write(".gitignore", "ignored.md\n")
	r.Write("tracked-a.go", "package a // TODO: a\n")
	r.Write("tracked-b.go", "package a // TODO: b\n")
	r.Write("ignored.md", "TODO: ignored\n")
	r.Commit("a repo whose listing will not fit")

	restore := ignoredListCapBytes
	ignoredListCapBytes = 8 // far below the listing; RunCapped must refuse
	defer func() { ignoredListCapBytes = restore }()

	ctx, err := newSourceContext(r.Root())
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	got := strings.Join(mustWalk(t, ctx), "\n")
	for _, want := range []string{"tracked-a.go", "tracked-b.go", "ignored.md"} {
		if !strings.Contains(got, want) {
			t.Errorf("%s vanished from the walk because the listing did not fit; an "+
				"unknowable answer must narrow NOTHING, or a big repository loses its "+
				"own tree silently. walk = %q", want, got)
		}
	}
}
