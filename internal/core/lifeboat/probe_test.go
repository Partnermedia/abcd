package lifeboat

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// gitFixture builds an isolated git repo with the given commits applied in
// order. Each commit is (relpath, content); an empty relpath makes an
// empty commit. It returns the repo dir. The repo has no README, no .abcd — it
// is a pure Tier-0 instrument.
func gitFixture(t *testing.T, commits []fixtureCommit) string {
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
	for _, c := range commits {
		if c.path != "" {
			p := filepath.Join(repo, c.path)
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(c.content), 0o644); err != nil {
				t.Fatal(err)
			}
			run("add", "-A")
		}
		run("commit", "-q", "--allow-empty", "-m", c.message)
	}
	return repo
}

type fixtureCommit struct {
	path    string
	content string
	message string
}

// TestProbeReportsEverySection is the report's completeness invariant: a probe
// names every brief section in the mapping, exactly once, and its summary
// accounts for all of them. A section silently dropped from the report would be
// a gap the coverage experiment could not see.
func TestProbeReportsEverySection(t *testing.T) {
	cov, err := Probe(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if cov.SchemaVersion != SchemaVersion {
		t.Errorf("schema_version = %d, want %d", cov.SchemaVersion, SchemaVersion)
	}
	if len(cov.Sections) != len(Table) {
		t.Fatalf("report has %d sections, mapping has %d", len(cov.Sections), len(Table))
	}
	seen := map[Section]int{}
	for _, s := range cov.Sections {
		seen[s.Name]++
		if !s.Status.Valid() {
			t.Errorf("section %s has invalid status %q", s.Name, s.Status)
		}
	}
	for _, m := range Table {
		if seen[m.Section] != 1 {
			t.Errorf("section %s appears %d times, want 1", m.Section, seen[m.Section])
		}
	}
	if got := cov.Summary.Grounded + cov.Summary.Partial + cov.Summary.Blank; got != len(cov.Sections) {
		t.Errorf("summary totals %d, want %d sections", got, len(cov.Sections))
	}
}

// TestProbeNeverBlankSectionCarriesAQuestion holds the "a blank is a result"
// contract: every blank section states the question a human must answer, so the
// report is a to-do list, not a shrug.
func TestProbeNeverBlankSectionCarriesAQuestion(t *testing.T) {
	cov, err := Probe(gitFixture(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range cov.Sections {
		if s.Status == StatusBlank && s.Question == "" {
			t.Errorf("blank section %s carries no question for a human", s.Name)
		}
	}
}

// TestProbeGitOnlyRepoPresentsOnlyGitTier proves tier detection: a repo with no
// README, no docs, and no .abcd is Tier-0 only, and a grounded section there can
// only have been grounded from git.
func TestProbeGitOnlyRepoPresentsOnlyGitTier(t *testing.T) {
	cov, err := Probe(gitFixture(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(cov.TiersPresent) != 1 || cov.TiersPresent[0] != TierGit {
		t.Fatalf("tiers_present = %v, want [git]", cov.TiersPresent)
	}
	for _, s := range cov.Sections {
		if s.Status != StatusBlank && s.Tier != TierGit {
			t.Errorf("section %s grounded at tier %s in a git-only repo", s.Name, s.Tier)
		}
	}
}

// TestProbeIsDeterministic guards the aggregate: the same repo probed twice must
// produce byte-identical JSON, or a cross-repo diff would show phantom churn.
func TestProbeIsDeterministic(t *testing.T) {
	repo := gitFixture(t, []fixtureCommit{
		{message: "a"}, {message: "b"}, {path: "x.txt", content: "x", message: "add x"},
	})
	a, err := Probe(repo)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Probe(repo)
	if err != nil {
		t.Fatal(err)
	}
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	if string(ja) != string(jb) {
		t.Errorf("probe is not deterministic:\n a=%s\n b=%s", ja, jb)
	}
}

// TestProbeNeverMutatesTheSource is the read-only safety property: probing a
// repo leaves its working tree byte-identical. A probe that wrote to the source
// would defeat the whole "point it at an archived project, touch nothing" premise.
func TestProbeNeverMutatesTheSource(t *testing.T) {
	repo := gitFixture(t, []fixtureCommit{
		{path: "README.md", content: "# demo\n", message: "add readme"},
		{message: "empty"},
	})
	before := treeHash(t, repo)
	if _, err := Probe(repo); err != nil {
		t.Fatal(err)
	}
	if after := treeHash(t, repo); after != before {
		t.Errorf("probe mutated the source tree (hash %s -> %s)", before, after)
	}
}

// treeHash is a cheap fingerprint of every file's path, size, and mode under
// root (excluding .git, whose internal bookkeeping is not the source of truth).
func treeHash(t *testing.T, root string) string {
	t.Helper()
	var acc string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if rel == ".git" {
			return filepath.SkipDir
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		acc += rel + "\x00" + info.Mode().String() + "\x00"
		if !d.IsDir() {
			acc += itoa(info.Size())
		}
		acc += "\n"
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return acc
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}
