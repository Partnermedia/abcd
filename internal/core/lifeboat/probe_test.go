package lifeboat

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
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

// writeTree writes each repo-relative path in files under dir, creating parent
// directories. It is the plain-directory counterpart of gitFixture: a tree with
// no git and no record, carrying exactly the files a test names.
func writeTree(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// TestWalkFilesCannotEscapeTheContainmentRoot is the containment property of the
// recursive walk: a probed tree is foreign and may point anywhere, so a
// symlinked file and a symlinked directory must both be skipped rather than
// followed, and no path the walk yields may resolve outside the repository root.
func TestWalkFilesCannotEscapeTheContainmentRoot(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	outside := filepath.Join(base, "outside")
	writeTree(t, repo, map[string]string{"src/inside.go": "package src\n\n// TODO: inside\n"})
	writeTree(t, outside, map[string]string{
		"secret.txt":    "// TODO: outside the root\n",
		"nested/out.go": "// FIXME: outside the root\n",
	})
	if err := os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(repo, "link-file.txt")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(repo, "link-dir")); err != nil {
		t.Fatal(err)
	}

	ctx, err := newSourceContext(repo)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	paths, truncated := ctx.WalkFiles(".")
	if truncated {
		t.Error("WalkFiles reported truncation on a two-file tree")
	}
	if got := strings.Join(paths, ","); got != "src/inside.go" {
		t.Fatalf("WalkFiles = %v, want only [src/inside.go]; a symlink was followed", paths)
	}
	realRoot, err := filepath.EvalSymlinks(repo)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		real, err := filepath.EvalSymlinks(filepath.Join(repo, filepath.FromSlash(p)))
		if err != nil {
			t.Fatal(err)
		}
		rel, err := filepath.Rel(realRoot, real)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			t.Errorf("WalkFiles yielded %q resolving to %q, outside the containment root %q", p, real, realRoot)
		}
	}
}

// TestWalkFilesSkipsDependencyAndGeneratedTrees pins the skip set: VCS
// internals, dependency trees, and generated code are never a team's own
// material and are the dominant cost of an unfiltered walk, so the walk never
// descends into them — at any depth, matched by directory name.
func TestWalkFilesSkipsDependencyAndGeneratedTrees(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"src/keep.go":               "package src\n\n// TODO: keep me\n",
		".git/hooks/pre-commit":     "# TODO: git internals\n",
		"node_modules/pkg/index.js": "// TODO: a dependency's marker\n",
		"vendor/dep/dep.go":         "// TODO: vendored\n",
		"generated/api.pb.go":       "// TODO: generated\n",
		"deep/vendor/nested.go":     "// TODO: vendored below the root\n",
	})
	ctx, err := newSourceContext(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	paths, _ := ctx.WalkFiles(".")
	if got := strings.Join(paths, ","); got != "src/keep.go" {
		t.Errorf("WalkFiles = %v, want only [src/keep.go]; skip set is %v", paths, walkSkipDirs)
	}
}

// TestWalkFilesStopsAtTheFileCap proves the walk is bounded: it stops at its
// file cap and says so, rather than running to exhaustion on a vast tree. The
// cap branch is exercised through the same code path at an affordable scale —
// the shipped cap stays a const, so concurrent adapters share no mutable state.
func TestWalkFilesStopsAtTheFileCap(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{}
	for i := 0; i < 10; i++ {
		files[fmt.Sprintf("f%02d.txt", i)] = "x\n"
	}
	writeTree(t, dir, files)
	ctx, err := newSourceContext(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	paths, truncated := ctx.WalkFiles(".")
	if truncated || len(paths) != len(files) {
		t.Errorf("WalkFiles = %d files, truncated=%v; want %d files untruncated under the %d cap",
			len(paths), truncated, len(files), maxWalkFiles)
	}

	capped, truncated := ctx.walkFilesLimited(".", 3)
	if !truncated {
		t.Error("walk did not report truncation with 10 files under a 3-file cap")
	}
	if len(capped) != 3 {
		t.Errorf("walk returned %d files under a 3-file cap, want 3", len(capped))
	}
}

// TestWalkFilesStopsAtTheDirectoryCap holds the bound a file cap alone cannot:
// a tree of directories holding no regular file at all never reaches the file
// cap, so the walk would run to exhaustion over a foreign tree that costs two
// syscalls per directory and yields nothing. Directories are counted against the
// same cap, and reaching it is reported like any other truncation.
func TestWalkFilesStopsAtTheDirectoryCap(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 10; i++ {
		if err := os.MkdirAll(filepath.Join(dir, fmt.Sprintf("d%02d", i)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	ctx, err := newSourceContext(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	paths, truncated := ctx.walkFilesLimited(".", 3)
	if !truncated {
		t.Errorf("walk of 10 empty directories under a 3-entry cap reported no truncation (paths %v)", paths)
	}
}

// TestWalkFilesStopsAtTheDepthCap holds the other unbounded dimension: every
// directory the walk opens is resolved from the containment root one component
// at a time, so the cost of a chain of directories grows with the square of its
// depth — a few thousand nested directories, cheap to create, cost minutes to
// walk. The walk descends maxWalkDepth levels and says it stopped there.
func TestWalkFilesStopsAtTheDepthCap(t *testing.T) {
	dir := t.TempDir()
	deep := make([]string, 0, maxWalkDepth+2)
	for i := 0; i < maxWalkDepth+2; i++ {
		deep = append(deep, "d")
	}
	buried := path.Join(append(deep, "buried.txt")...)
	writeTree(t, dir, map[string]string{
		"shallow.txt": "x\n",
		buried:        "x\n",
	})
	ctx, err := newSourceContext(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	paths, truncated := ctx.WalkFiles(".")
	if !truncated {
		t.Errorf("walk of a %d-deep chain under a depth cap of %d reported no truncation", len(deep), maxWalkDepth)
	}
	for _, p := range paths {
		if p == buried {
			t.Errorf("walk returned %q, %d levels below its depth cap of %d", p, len(deep)-maxWalkDepth, maxWalkDepth)
		}
	}
	// Pruning a chain is not abandoning the tree: what sits above the cap is
	// still walked and still returned.
	if len(paths) == 0 || paths[len(paths)-1] != "shallow.txt" {
		t.Errorf("walk = %v, want the shallow file still returned", paths)
	}
}

// TestProbeLeavesEveryFileByteIdentical is the read-only invariance property at
// full strength: a probe of a marker-bearing tree must leave every file's
// contents unchanged and add or remove nothing. The marker adapter reads every
// file in the tree, so byte-level proof — not a path/size fingerprint — is what
// makes "point it at an archived project, touch nothing" true.
func TestProbeLeavesEveryFileByteIdentical(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"README.md":     "# demo\n\nA demo project kept for the probe.\n",
		"go.mod":        "module example.com/demo\n\ngo 1.22\n",
		"src/a.go":      "package a\n\n// TODO: handle the retry case\n",
		"src/b.go":      "package a\n\n// FIXME: this leaks a connection\n",
		"docs/notes.md": "Notes about the demo.\n",
	})
	before := fileHashes(t, dir)
	if _, err := Probe(dir); err != nil {
		t.Fatal(err)
	}
	after := fileHashes(t, dir)

	for p, want := range before {
		got, ok := after[p]
		if !ok {
			t.Errorf("probe removed %s", p)
			continue
		}
		if got != want {
			t.Errorf("probe rewrote %s (sha256 %s -> %s)", p, want, got)
		}
	}
	for p := range after {
		if _, ok := before[p]; !ok {
			t.Errorf("probe created %s", p)
		}
	}
}

// fileHashes maps every regular file under root to the SHA-256 of its contents.
func fileHashes(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = fmt.Sprintf("%x", sha256.Sum256(data))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
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
