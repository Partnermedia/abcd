package ahoy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/vintage"
	"github.com/REPPL/abcd-cli/internal/gittest"
)

// noWrites asserts the install left none of its evidence artefacts — a refusal
// before the first apply step must not touch the repo.
func noWrites(t *testing.T, repo string) {
	t.Helper()
	for _, rel := range []string{".abcd/config.json", ".abcd/rules.json", "CLAUDE.md", "AGENTS.md"} {
		if _, err := os.Stat(filepath.Join(repo, rel)); err == nil {
			t.Errorf("refusal still wrote %s", rel)
		}
	}
}

func TestInstallRefusesUnknownVintage(t *testing.T) {
	setupHermetic(t)
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	orig := currentVintage
	t.Cleanup(func() { currentVintage = orig })
	currentVintage = func() vintage.Current { return vintage.Current{Known: false} }

	res, err := Install(repo, installOpts(), RefusingPrompter{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "refused" {
		t.Fatalf("status = %q, want refused", res.Status)
	}
	if len(res.Writes) != 0 {
		t.Fatalf("refusal reported writes: %v", res.Writes)
	}
	if len(res.Notes) == 0 {
		t.Fatal("a refusal must record its reason in Notes")
	}
	noWrites(t, repo)
}

func TestInstallRefusesStaleAgainstTip(t *testing.T) {
	setupHermetic(t)
	r := gittest.NewRepo(t)
	r.Commit("first")
	first := r.Git("rev-parse", "HEAD")
	r.Commit("second") // advance the tip past the binary's revision

	orig := currentVintage
	t.Cleanup(func() { currentVintage = orig })
	currentVintage = func() vintage.Current { return vintage.Current{Revision: first, Known: true} }

	res, err := Install(r.Root(), installOpts(), RefusingPrompter{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "refused" {
		t.Fatalf("status = %q, want refused", res.Status)
	}
	if len(res.Writes) != 0 {
		t.Fatalf("refusal reported writes: %v", res.Writes)
	}
	noWrites(t, r.Root())
}

func TestInstallOverrideProceedsThroughStaleBinary(t *testing.T) {
	setupHermetic(t)
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	orig := currentVintage
	t.Cleanup(func() { currentVintage = orig })
	currentVintage = func() vintage.Current { return vintage.Current{Known: false} }

	opts := installOpts()
	opts.AllowStaleBinary = true
	res, err := Install(repo, opts, RefusingPrompter{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status == "refused" {
		t.Fatalf("override should proceed, got refused (notes=%v)", res.Notes)
	}
	if _, err := os.Stat(filepath.Join(repo, ".abcd", "config.json")); err != nil {
		t.Errorf("override install wrote nothing: %v", err)
	}
}

func TestInstallProceedsThroughFreshBinary(t *testing.T) {
	setupHermetic(t)
	repo := t.TempDir()
	// A bare .git dir: adoptable, but git cannot answer for a tip, so the binary
	// is not a dogfood-stale one. A determinable vintage here must proceed.
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	orig := currentVintage
	t.Cleanup(func() { currentVintage = orig })
	currentVintage = func() vintage.Current {
		return vintage.Current{Revision: "1111111111111111111111111111111111111111", Known: true}
	}

	res, err := Install(repo, installOpts(), RefusingPrompter{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status == "refused" {
		t.Fatalf("a fresh, determinable binary must proceed, got refused (notes=%v)", res.Notes)
	}
	if _, err := os.Stat(filepath.Join(repo, ".abcd", "config.json")); err != nil {
		t.Errorf("fresh install wrote nothing: %v", err)
	}
}
