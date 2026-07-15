package lifeboat

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestAggregateDisambiguationSurvivesASuffixCollision is the fix for the
// column-collapse bug: when a real repo name coincides with the "#N" suffix
// generated for another, every repo must still get its own column. Names
// "foo", "foo#2", "foo" must yield three distinct columns, not two.
func TestAggregateDisambiguationSurvivesASuffixCollision(t *testing.T) {
	a := covWith("foo", 1, []Tier{TierGit}, map[Section]Status{"graveyard": StatusGrounded})
	b := covWith("foo#2", 2, []Tier{TierGit}, map[Section]Status{"graveyard": StatusPartial})
	c := covWith("foo", 3, []Tier{TierGit}, map[Section]Status{"graveyard": StatusBlank})

	agg := Aggregate([]Coverage{a, b, c})
	if len(agg.Repos) != 3 {
		t.Fatalf("want 3 repo columns, got %d", len(agg.Repos))
	}
	names := map[string]bool{}
	for _, r := range agg.Repos {
		if names[r.Name] {
			t.Errorf("duplicate column name %q — a repo column was collapsed", r.Name)
		}
		names[r.Name] = true
	}
	row := findRow(t, agg, "graveyard")
	if len(row.Cells) != 3 {
		t.Errorf("graveyard row has %d cells, want 3 (one per repo)", len(row.Cells))
	}
}

// TestRenderStripsTerminalControlCharacters is the fix for control-char
// injection: evidence, questions, and the repo name are built from repository
// content a hostile repo controls, so the rendered text report must not carry
// raw ANSI/control bytes that could spoof or corrupt a terminal.
func TestRenderStripsTerminalControlCharacters(t *testing.T) {
	esc := "\x1b[31mFAKE\x1b[0m"
	cov := Coverage{
		SchemaVersion: SchemaVersion,
		Repo:          RepoInfo{Name: "evil" + esc, Commits: 1},
		TiersPresent:  []Tier{TierGit},
		Sections: []SectionCoverage{
			{Name: "graveyard", Status: StatusGrounded, Confidence: ConfidenceHigh, Tier: TierGit,
				Evidence: []string{"deleted " + esc + ".txt"}},
			{Name: "product/context", Status: StatusBlank,
				Searched: []string{"README " + esc}, Question: "why " + esc + "?"},
		},
		Summary: Summary{Grounded: 1, Blank: 1},
	}
	out := cov.Render()
	if strings.ContainsRune(out, '\x1b') {
		t.Error("rendered report carries a raw ESC byte from repo-controlled content")
	}
	for _, r := range out {
		if r < 0x20 && r != '\n' {
			t.Errorf("rendered report carries control byte %#x", r)
		}
	}

	// The aggregate render sanitizes the repo name at construction time.
	agg := Aggregate([]Coverage{cov})
	if strings.ContainsRune(agg.Render(), '\x1b') {
		t.Error("aggregate render carries a raw ESC byte from a repo name")
	}
}

// gitFixtureDeletionOnly builds a git-only repo that deletes a file but never
// reverts anything — the ambiguous case the graveyard must treat as partial,
// not grounded.
func gitFixtureDeletionOnly(t *testing.T) string {
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
	run("init", "-q")
	run("commit", "-q", "--allow-empty", "-m", "root")
	if err := os.WriteFile(filepath.Join(repo, "tmp.txt"), []byte("scratch\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "-A")
	run("commit", "-q", "-m", "add tmp")
	run("rm", "-q", "tmp.txt")
	run("commit", "-q", "-m", "remove tmp")
	return repo
}

// TestGraveyardDeletionOnlyIsPartial holds the honesty of the headline number:
// a mere file deletion (which nearly every repo has) is only partial evidence of
// a graveyard, while an explicit revert grounds it. Over-grounding here would
// make the graveyard near-universally "grounded" and muddy the experiment.
func TestGraveyardDeletionOnlyIsPartial(t *testing.T) {
	del := findSection(t, mustProbe(t, gitFixtureDeletionOnly(t)), "graveyard")
	if del.Status != StatusPartial {
		t.Errorf("deletion-only graveyard = %s, want partial", del.Status)
	}
	if len(del.Evidence) == 0 {
		t.Error("partial graveyard cites no evidence")
	}

	rev := findSection(t, mustProbe(t, gitFixtureWithRevert(t)), "graveyard")
	if rev.Status != StatusGrounded {
		t.Errorf("revert graveyard = %s, want grounded", rev.Status)
	}
}

func mustProbe(t *testing.T, repo string) Coverage {
	t.Helper()
	cov, err := Probe(repo)
	if err != nil {
		t.Fatal(err)
	}
	return cov
}
