package lifeboat

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// gitTierFixture builds an isolated git repo that exercises the Tier-0 signals:
// a few descriptive commits, a real `git revert`, and a file added then
// `git rm`-deleted. It carries no manifest (go.mod etc.), so the dependency
// section is a genuine absent-material blank. Returns the repo dir.
func gitTierFixture(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	repo := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(),
			"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@e",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@e",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(repo, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	run("init", "-q")
	run("commit", "-q", "--allow-empty", "-m", "root")

	write("keep.txt", "keep\n")
	run("add", "-A")
	run("commit", "-q", "-m", "add the thing the project keeps")

	write("legacy.txt", "legacy\n")
	run("add", "-A")
	run("commit", "-q", "-m", "add legacy path")

	run("rm", "-q", "legacy.txt")
	run("commit", "-q", "-m", "drop legacy path")

	write("bad.txt", "bad\n")
	run("add", "-A")
	run("commit", "-q", "-m", "add an experimental feature")
	run("revert", "--no-edit", "HEAD")

	return repo
}

// sourceForSection returns the Tier-0 adapter that speaks for section, failing
// the test if none does.
func sourceForSection(t *testing.T, section Section) Source {
	t.Helper()
	for _, s := range gitSources() {
		if s.Section() == section {
			return s
		}
	}
	t.Fatalf("no git source for section %s", section)
	return nil
}

// TestGitGraveyardGroundsFromHistory is the flagship assertion: the graveyard
// adapter grounds at TierGit purely from git history — reverts and deletions —
// and cites the evidence it found.
func TestGitGraveyardGroundsFromHistory(t *testing.T) {
	ctx, err := newSourceContext(gitTierFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	src := sourceForSection(t, "graveyard")
	if src.Tier() != TierGit {
		t.Fatalf("graveyard tier = %s, want %s", src.Tier(), TierGit)
	}
	ev := src.Probe(ctx)
	if ev.Status != StatusGrounded {
		t.Fatalf("graveyard status = %s, want grounded", ev.Status)
	}
	if len(ev.Sources) == 0 {
		t.Fatal("grounded graveyard cites no evidence")
	}
	if ev.Confidence == "" {
		t.Error("grounded graveyard has no confidence")
	}
}

// TestGitWhatDidntGroundsFromReverts confirms a second adapter reads the same
// history and partially grounds "evidence/what-didnt" from the revert.
func TestGitWhatDidntGroundsFromReverts(t *testing.T) {
	ctx, err := newSourceContext(gitTierFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	ev := sourceForSection(t, "evidence/what-didnt").Probe(ctx)
	if ev.Status != StatusPartial {
		t.Fatalf("what-didnt status = %s, want partial", ev.Status)
	}
	if len(ev.Sources) == 0 {
		t.Fatal("partial what-didnt cites no evidence")
	}
}

// TestGitDependenciesBlankWhenNoManifest holds the "a blank is a result"
// contract for Tier-0: a repo with no manifest in history returns a blank that
// names what was searched and the question a human must answer.
func TestGitDependenciesBlankWhenNoManifest(t *testing.T) {
	ctx, err := newSourceContext(gitTierFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	ev := sourceForSection(t, "constraints/dependencies").Probe(ctx)
	if ev.Status != StatusBlank {
		t.Fatalf("dependencies status = %s, want blank (no manifest in fixture)", ev.Status)
	}
	if ev.Question == "" {
		t.Error("blank dependencies carries no question for a human")
	}
	if len(ev.Searched) == 0 {
		t.Error("blank dependencies names nothing it searched")
	}
}

// TestGitAdaptersAllReportTierGit guards the tier contract: every Tier-0 adapter
// reports TierGit, so the orchestrator's tier stamping cannot be misled.
func TestGitAdaptersAllReportTierGit(t *testing.T) {
	for _, s := range gitSources() {
		if s.Tier() != TierGit {
			t.Errorf("source for %s reports tier %s, want %s", s.Section(), s.Tier(), TierGit)
		}
	}
}

// TestGitProbeIsDeterministic asserts the Tier-0 adapters are byte-stable across
// runs against the same repo, as the deterministic-probe invariant requires.
func TestGitProbeIsDeterministic(t *testing.T) {
	repo := gitTierFixture(t)
	ctxA, err := newSourceContext(repo)
	if err != nil {
		t.Fatal(err)
	}
	defer ctxA.Close()
	ctxB, err := newSourceContext(repo)
	if err != nil {
		t.Fatal(err)
	}
	defer ctxB.Close()

	for _, s := range gitSources() {
		a := s.Probe(ctxA)
		b := s.Probe(ctxB)
		if len(a.Sources) != len(b.Sources) {
			t.Fatalf("%s: source count differs across runs", s.Section())
		}
		for i := range a.Sources {
			if a.Sources[i] != b.Sources[i] {
				t.Errorf("%s: source[%d] differs: %q vs %q", s.Section(), i, a.Sources[i], b.Sources[i])
			}
		}
	}
}
