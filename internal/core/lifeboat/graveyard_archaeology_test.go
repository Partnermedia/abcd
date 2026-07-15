package lifeboat

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gvRepo is a self-built, isolated git repo for the Tier-0 archaeology tests. CI
// is shallow, so every test builds the history it asserts on rather than leaning
// on this project's own log. Commit dates are set explicitly where ordering by
// divergence age matters, so the assertion is deterministic.
type gvRepo struct {
	t   *testing.T
	dir string
	env []string
}

func gvNewRepo(t *testing.T) *gvRepo {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	r := &gvRepo{t: t, dir: t.TempDir()}
	r.env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@e",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@e",
	)
	r.git("init", "-q")
	// Force the initial branch to main regardless of the host git default, so the
	// default-branch resolution is deterministic across git versions.
	r.git("symbolic-ref", "HEAD", "refs/heads/main")
	return r
}

func (r *gvRepo) run(env []string, args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}

func (r *gvRepo) git(args ...string) string { return r.run(r.env, args...) }

func (r *gvRepo) gitAt(date string, args ...string) string {
	env := append(append([]string{}, r.env...),
		"GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date)
	return r.run(env, args...)
}

func (r *gvRepo) write(name, content string) {
	r.t.Helper()
	full := filepath.Join(r.dir, name)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

func (r *gvRepo) rm(name string) {
	r.t.Helper()
	if err := os.Remove(filepath.Join(r.dir, name)); err != nil {
		r.t.Fatal(err)
	}
}

func (r *gvRepo) commit(msg string)      { r.git("commit", "-q", "--allow-empty", "-m", msg) }
func (r *gvRepo) commitAt(d, msg string) { r.gitAt(d, "commit", "-q", "--allow-empty", "-m", msg) }
func (r *gvRepo) addCommit(msg string)   { r.git("add", "-A"); r.git("commit", "-q", "-m", msg) }
func (r *gvRepo) addCommitAt(d, msg string) {
	r.git("add", "-A")
	r.gitAt(d, "commit", "-q", "-m", msg)
}

// gvArch builds the archaeology dig over dir through a fresh SourceContext.
func gvArch(t *testing.T, dir string) Archaeology {
	t.Helper()
	ctx, err := newSourceContext(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(ctx.Close)
	return buildArchaeology(ctx)
}

func bySignal(a Archaeology, s Signal) []Finding {
	var out []Finding
	for _, f := range a.Findings {
		if f.Signal == s {
			out = append(out, f)
		}
	}
	return out
}

// TestArchRevertFinding is the flagship signal: a reverted commit yields exactly
// one rev-<12hex> finding, and the whole dig is byte-identical across two
// buildArchaeology calls (the re-plan id-stability invariant).
func TestArchRevertFinding(t *testing.T) {
	r := gvNewRepo(t)
	r.commit("root")
	r.write("bad.txt", "bad\n")
	r.addCommit("add an experiment")
	r.git("revert", "--no-edit", "HEAD")

	a := gvArch(t, r.dir)
	rev := bySignal(a, SignalRevert)
	if len(rev) != 1 {
		t.Fatalf("revert findings = %d, want 1 (%+v)", len(rev), a.Findings)
	}
	id := rev[0].ID
	if !strings.HasPrefix(id, "rev-") {
		t.Errorf("revert id = %q, want rev- prefix", id)
	}
	if got := len(strings.TrimPrefix(id, "rev-")); got != 12 {
		t.Errorf("revert id hex len = %d, want 12 (%q)", got, id)
	}
	if len(rev[0].Evidence) == 0 || !strings.Contains(rev[0].Evidence[0], "Revert") {
		t.Errorf("revert evidence does not quote the subject: %+v", rev[0].Evidence)
	}

	// Byte-identical re-plan: two digs of an unchanged repo marshal identically.
	b := gvArch(t, r.dir)
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	if string(ja) != string(jb) {
		t.Errorf("archaeology not stable across calls:\n a=%s\n b=%s", ja, jb)
	}
}

// TestArchUnmergedBranchesOrderedByDivergence: two unmerged branches are ranked
// by divergence age (older merge-base first), and the default branch never
// appears as its own finding.
func TestArchUnmergedBranchesOrderedByDivergence(t *testing.T) {
	r := gvNewRepo(t)
	r.commitAt("2021-01-01T00:00:00", "root")
	r.write("a.txt", "a\n")
	r.addCommitAt("2021-02-01T00:00:00", "c1") // old divergence point
	r.git("branch", "old-branch")
	r.write("b.txt", "b\n")
	r.addCommitAt("2021-03-01T00:00:00", "c2") // new divergence point
	r.git("branch", "new-branch")
	r.write("c.txt", "c\n")
	r.addCommitAt("2021-04-01T00:00:00", "c3 on main")

	// Put a commit on each branch so both are ahead of main (unmerged).
	r.git("checkout", "-q", "old-branch")
	r.write("oa.txt", "oa\n")
	r.addCommitAt("2021-05-01T00:00:00", "old work")
	r.git("checkout", "-q", "new-branch")
	r.write("nb.txt", "nb\n")
	r.addCommitAt("2021-06-01T00:00:00", "new work")
	r.git("checkout", "-q", "main")

	a := gvArch(t, r.dir)
	br := bySignal(a, SignalUnmergedBranch)
	if len(br) != 2 {
		t.Fatalf("unmerged branch findings = %d, want 2 (%+v)", len(br), a.Findings)
	}
	if br[0].ID != "branch-old-branch" || br[1].ID != "branch-new-branch" {
		t.Errorf("branch order = [%s, %s], want [branch-old-branch, branch-new-branch]", br[0].ID, br[1].ID)
	}
	for _, f := range br {
		if f.ID == "branch-main" {
			t.Error("the default branch appeared as its own unmerged finding")
		}
	}
	// Evidence cites the ahead count and the merge-base.
	joined := strings.Join(br[0].Evidence, " ")
	if !strings.Contains(joined, "ahead of main") || !strings.Contains(joined, "merge-base") {
		t.Errorf("branch evidence missing ahead/merge-base: %+v", br[0].Evidence)
	}
}

// TestDefaultBranchOriginHead: origin/HEAD, when set, wins over every local
// candidate.
func TestDefaultBranchOriginHead(t *testing.T) {
	r := gvNewRepo(t)
	r.commit("root")
	r.git("update-ref", "refs/remotes/origin/trunk", "HEAD")
	r.git("symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/trunk")

	ctx, err := newSourceContext(r.dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()
	if got := defaultBranch(ctx); got != "trunk" {
		t.Errorf("defaultBranch = %q, want trunk (origin/HEAD)", got)
	}
}

// TestDefaultBranchFallsToMain: no origin/HEAD, but a main branch exists.
func TestDefaultBranchFallsToMain(t *testing.T) {
	r := gvNewRepo(t)
	r.commit("root")

	ctx, err := newSourceContext(r.dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()
	if got := defaultBranch(ctx); got != "main" {
		t.Errorf("defaultBranch = %q, want main", got)
	}
}

// TestDefaultBranchFallsToHeadBranch: no origin/HEAD and none of
// {main,master,trunk,develop} — the branch HEAD points at is used.
func TestDefaultBranchFallsToHeadBranch(t *testing.T) {
	r := gvNewRepo(t)
	r.commit("root")
	r.git("branch", "-m", "feature") // rename main -> feature; no candidate exists

	ctx, err := newSourceContext(r.dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()
	if got := defaultBranch(ctx); got != "feature" {
		t.Errorf("defaultBranch = %q, want feature (HEAD branch)", got)
	}
}

// TestDefaultBranchDetachedIsEmpty: a detached HEAD with no candidate branch
// resolves to "" and yields zero branch findings without crashing.
func TestDefaultBranchDetachedIsEmpty(t *testing.T) {
	r := gvNewRepo(t)
	r.commit("root")
	r.git("branch", "-m", "feature")
	r.git("checkout", "-q", "--detach", "HEAD")

	ctx, err := newSourceContext(r.dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()
	if got := defaultBranch(ctx); got != "" {
		t.Errorf("defaultBranch = %q, want empty (detached, no candidate)", got)
	}
	if br := bySignal(buildArchaeology(ctx), SignalUnmergedBranch); len(br) != 0 {
		t.Errorf("detached HEAD produced %d branch findings, want 0", len(br))
	}
}

// TestArchDeletedPaths: a path deleted after substantial history is reported; a
// path deleted with too little history is not; a path deleted then re-added
// (present at HEAD) is not.
func TestArchDeletedPaths(t *testing.T) {
	r := gvNewRepo(t)
	r.commit("root")

	// longlived.txt: touched by >= substantialHistoryCommits commits, then deleted.
	r.write("longlived.txt", "v0\n")
	r.addCommit("add longlived")
	for i := 1; i < substantialHistoryCommits; i++ {
		r.write("longlived.txt", fmt.Sprintf("v%d\n", i))
		r.addCommit(fmt.Sprintf("touch longlived %d", i))
	}
	r.rm("longlived.txt")
	r.addCommit("drop longlived")

	// shortlived.txt: created then deleted — only two touches.
	r.write("shortlived.txt", "x\n")
	r.addCommit("add shortlived")
	r.rm("shortlived.txt")
	r.addCommit("drop shortlived")

	// readded.txt: substantial history, deleted, then re-added — present at HEAD.
	r.write("readded.txt", "r0\n")
	r.addCommit("add readded")
	for i := 1; i < substantialHistoryCommits; i++ {
		r.write("readded.txt", fmt.Sprintf("r%d\n", i))
		r.addCommit(fmt.Sprintf("touch readded %d", i))
	}
	r.rm("readded.txt")
	r.addCommit("drop readded")
	r.write("readded.txt", "back\n")
	r.addCommit("bring readded back")

	del := bySignal(gvArch(t, r.dir), SignalDeletedPath)
	got := map[string]bool{}
	for _, f := range del {
		got[f.ID] = true
	}
	if !got["del-longlived.txt"] {
		t.Errorf("substantial deleted path not reported; findings: %+v", del)
	}
	if got["del-shortlived.txt"] {
		t.Error("a path with too little history was reported as abandoned")
	}
	if got["del-readded.txt"] {
		t.Error("a path re-added at HEAD was reported as deleted")
	}
}

// TestArchRemovedDependencies: a dep present in the first manifest revision but
// gone at HEAD is reported; a manifest deleted wholesale reports its tokens; a
// manifest whose deps only grew reports nothing.
func TestArchRemovedDependencies(t *testing.T) {
	r := gvNewRepo(t)
	r.commit("root")

	// package.json: three deps at first, one at HEAD (two removed).
	r.write("package.json", `{
  "name": "demo",
  "dependencies": {
    "left-pad": "1.0.0",
    "request": "2.0.0",
    "lodash": "4.0.0"
  }
}
`)
	r.addCommit("add package.json")
	r.write("package.json", `{
  "name": "demo",
  "dependencies": {
    "lodash": "4.0.0"
  }
}
`)
	r.addCommit("shrink package.json deps")

	// go.mod: adopted then deleted wholesale.
	r.write("go.mod", `module example.com/demo

go 1.21

require (
	github.com/foo/bar v1.2.3
)
`)
	r.addCommit("add go.mod")
	r.rm("go.mod")
	r.addCommit("drop go.mod")

	// requirements.txt: deps only grow — nothing removed.
	r.write("requirements.txt", "alpha==1.0\n")
	r.addCommit("add requirements")
	r.write("requirements.txt", "alpha==1.0\nbeta==2.0\n")
	r.addCommit("grow requirements")

	dep := bySignal(gvArch(t, r.dir), SignalRemovedDependency)
	byID := map[string]Finding{}
	for _, f := range dep {
		byID[f.ID] = f
	}

	pkg, ok := byID["dep-package.json"]
	if !ok {
		t.Fatalf("no removed-dependency finding for package.json; findings: %+v", dep)
	}
	ev := strings.Join(pkg.Evidence, "\n")
	if !strings.Contains(ev, "removed: left-pad") || !strings.Contains(ev, "removed: request") {
		t.Errorf("package.json finding does not list removed deps: %+v", pkg.Evidence)
	}
	if strings.Contains(ev, "removed: lodash") {
		t.Errorf("a retained dep was reported removed: %+v", pkg.Evidence)
	}

	gomod, ok := byID["dep-go.mod"]
	if !ok {
		t.Fatalf("no removed-dependency finding for a wholesale-deleted go.mod; findings: %+v", dep)
	}
	if !strings.Contains(strings.Join(gomod.Evidence, "\n"), "github.com/foo/bar") {
		t.Errorf("wholesale go.mod deletion does not cite its dependency: %+v", gomod.Evidence)
	}

	if _, ok := byID["dep-requirements.txt"]; ok {
		t.Error("a manifest whose deps only grew was reported as having removals")
	}
}

// TestArchWholesaleRewrite: a single commit changing >= min files and >= the
// tree fraction is a rewrite.
func TestArchWholesaleRewrite(t *testing.T) {
	r := gvNewRepo(t)
	r.commit("root")
	for i := 0; i < 30; i++ {
		r.write(fmt.Sprintf("f%02d.txt", i), "x\n")
	}
	r.addCommit("switch to modular core")

	rw := bySignal(gvArch(t, r.dir), SignalWholesaleRewrite)
	if len(rw) != 1 {
		t.Fatalf("rewrite findings = %d, want 1 (%+v)", len(rw), rw)
	}
	if !strings.HasPrefix(rw[0].ID, "rewrite-") {
		t.Errorf("rewrite id = %q, want rewrite- prefix", rw[0].ID)
	}
	if !strings.Contains(strings.Join(rw[0].Evidence, "\n"), "of 30 tracked") {
		t.Errorf("rewrite evidence missing tree denominator: %+v", rw[0].Evidence)
	}
}

// TestArchWholesaleRewriteBelowFractionNone: a commit changing >= min files but
// under the tree fraction is not a rewrite.
func TestArchWholesaleRewriteBelowFractionNone(t *testing.T) {
	r := gvNewRepo(t)
	r.commit("root")
	// Build a 60-file tree in commits of 20 (each under the min-files floor).
	n := 0
	for batch := 0; batch < 3; batch++ {
		for i := 0; i < 20; i++ {
			r.write(fmt.Sprintf("f%03d.txt", n), "v0\n")
			n++
		}
		r.addCommit(fmt.Sprintf("batch %d", batch))
	}
	// Modify 29 of the 60 files in one commit: >= min files (25) but < 0.5*60.
	for i := 0; i < 29; i++ {
		r.write(fmt.Sprintf("f%03d.txt", i), "v1\n")
	}
	r.addCommit("touch 29 files")

	if rw := bySignal(gvArch(t, r.dir), SignalWholesaleRewrite); len(rw) != 0 {
		t.Errorf("below-fraction commit reported as rewrite: %+v", rw)
	}
}

// TestArchWholesaleRewriteTinyRepoNone: a commit touching most of a 3-file tree
// is below the min-files floor, so it is not a rewrite.
func TestArchWholesaleRewriteTinyRepoNone(t *testing.T) {
	r := gvNewRepo(t)
	r.commit("root")
	for i := 0; i < 3; i++ {
		r.write(fmt.Sprintf("f%d.txt", i), "x\n")
	}
	r.addCommit("add three files")

	if rw := bySignal(gvArch(t, r.dir), SignalWholesaleRewrite); len(rw) != 0 {
		t.Errorf("tiny-repo commit reported as rewrite: %+v", rw)
	}
}

// TestBoundProbeList is the pure bound guarding every per-entry git fan-out: the
// cap is respected, order is preserved, and a list already within the bound is
// returned untouched.
func TestBoundProbeList(t *testing.T) {
	long := []string{"a", "b", "c", "d", "e"}
	if got := boundProbeList(long, 3); len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("boundProbeList(len5, 3) = %v, want [a b c]", got)
	}
	short := []string{"x", "y"}
	if got := boundProbeList(short, 3); len(got) != 2 || got[0] != "x" || got[1] != "y" {
		t.Errorf("boundProbeList(len2, 3) = %v, want the list untouched", got)
	}
	if got := boundProbeList(long, 0); len(got) != 0 {
		t.Errorf("boundProbeList(_, 0) = %v, want empty", got)
	}
	var empty []string
	if got := boundProbeList(empty, 5); len(got) != 0 {
		t.Errorf("boundProbeList(nil, 5) = %v, want empty", got)
	}
}

// TestGvDeletedPathsConstantGitInvocations proves the deleted-path signal issues
// O(1) git execs, not O(deleted paths): a hostile history with many deleted paths
// must not fan out one `git log -- <p>` and one `cat-file -e` per path at plan
// time. It asserts the SourceContext git cache holds a small constant number of
// distinct invocations regardless of how many paths were deleted.
func TestGvDeletedPathsConstantGitInvocations(t *testing.T) {
	r := gvNewRepo(t)
	r.commit("root")
	// 12 distinct paths, each created then deleted (2 commits each) — well above
	// any small constant, so a per-path implementation floods the cache.
	const nPaths = 12
	for i := 0; i < nPaths; i++ {
		name := fmt.Sprintf("gone%02d.txt", i)
		r.write(name, "x\n")
		r.addCommit("add " + name)
		r.rm(name)
		r.addCommit("drop " + name)
	}

	ctx, err := newSourceContext(r.dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(ctx.Close)
	_ = gvDeletedPaths(ctx)

	// Independent of nPaths: one deleted-paths listing, one touch-count walk, one
	// tracked-file listing. A small bound catches the O(paths) regression (which
	// would leave >= nPaths+1 entries) without pinning the exact count.
	if got := len(ctx.gitCache); got > 5 {
		t.Errorf("gvDeletedPaths issued %d distinct git invocations for %d deleted paths; want O(1) (<= 5)", got, nPaths)
	}
}

// TestArchDepGoModReformatNoFalseRemoval: a go.mod whose single dependency is
// reformatted from block form to single-line form (what `go mod edit -fmt`/tidy
// does when one dependency remains) must NOT fabricate a removed-dependency
// finding — the dep is unchanged, only the directive layout moved.
func TestArchDepGoModReformatNoFalseRemoval(t *testing.T) {
	r := gvNewRepo(t)
	r.commit("root")
	r.write("go.mod", "module example.com/demo\n\ngo 1.21\n\nrequire (\n\tgithub.com/foo/bar v1.2.3\n)\n")
	r.addCommit("add go.mod (block form)")
	r.write("go.mod", "module example.com/demo\n\ngo 1.21\n\nrequire github.com/foo/bar v1.2.3\n")
	r.addCommit("reformat go.mod to single-line require")

	for _, f := range bySignal(gvArch(t, r.dir), SignalRemovedDependency) {
		if f.ID == "dep-go.mod" {
			t.Errorf("block->single-line reformat fabricated a removed-dependency finding: %+v", f.Evidence)
		}
	}
}

// TestArchWholesaleRewriteRootCommitNone: the parentless root commit diffs against
// the empty tree, so importing a whole tree in the first commit LOOKS like a
// wholesale rewrite. It must not be flagged — there is nothing it replaced.
func TestArchWholesaleRewriteRootCommitNone(t *testing.T) {
	r := gvNewRepo(t)
	// The very first commit imports 30 files: >= min-files and 100% of the tree.
	for i := 0; i < 30; i++ {
		r.write(fmt.Sprintf("f%02d.txt", i), "x\n")
	}
	r.addCommit("initial import")

	if rw := bySignal(gvArch(t, r.dir), SignalWholesaleRewrite); len(rw) != 0 {
		t.Errorf("parentless root import reported as a wholesale rewrite: %+v", rw)
	}
}

// TestArchEmptyMarshalsFindingsArray: a non-git tree and an empty git repo both
// yield {schema_version:1, findings:[]} — a valid empty file, never null.
func TestArchEmptyMarshalsFindingsArray(t *testing.T) {
	// Non-git plain directory.
	nonGit := t.TempDir()
	assertEmptyArch(t, "non-git", gvArch(t, nonGit))

	// Empty git repo (no commits).
	r := gvNewRepo(t)
	assertEmptyArch(t, "empty-git", gvArch(t, r.dir))
}

func assertEmptyArch(t *testing.T, label string, a Archaeology) {
	t.Helper()
	if a.SchemaVersion != GraveyardSchemaVersion {
		t.Errorf("%s: schema_version = %d, want %d", label, a.SchemaVersion, GraveyardSchemaVersion)
	}
	if a.Findings == nil {
		t.Errorf("%s: Findings is nil (would marshal as null)", label)
	}
	j, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("%s: marshal: %v", label, err)
	}
	if !strings.Contains(string(j), `"findings":[]`) {
		t.Errorf("%s: empty dig did not marshal findings:[]: %s", label, j)
	}
}

// TestArchRevertCapTruncates: more than maxGraveyardFindingsPerSignal reverts are
// truncated to the cap, and the dig still marshals to valid JSON.
func TestArchRevertCapTruncates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cap test in -short mode (builds many commits)")
	}
	r := gvNewRepo(t)
	r.commit("root")
	for i := 0; i < maxGraveyardFindingsPerSignal+2; i++ {
		r.commit(fmt.Sprintf("Revert \"experiment %d\"", i))
	}

	a := gvArch(t, r.dir)
	rev := bySignal(a, SignalRevert)
	if len(rev) != maxGraveyardFindingsPerSignal {
		t.Errorf("revert findings = %d, want the cap %d", len(rev), maxGraveyardFindingsPerSignal)
	}
	j, err := json.Marshal(a)
	if err != nil || !json.Valid(j) {
		t.Errorf("truncated dig is not valid JSON: err=%v", err)
	}
}

// TestArchSanitisesRevertEvidence: a revert subject carrying a terminal control
// character is neutralised in the evidence, so an archived/hostile repo cannot
// smuggle an escape into a human render.
func TestArchSanitisesRevertEvidence(t *testing.T) {
	r := gvNewRepo(t)
	r.commit("root")
	// A control character (ESC) embedded in a revert subject.
	r.commit("Revert \"danger \x1b[31m here\"")

	rev := bySignal(gvArch(t, r.dir), SignalRevert)
	if len(rev) != 1 {
		t.Fatalf("revert findings = %d, want 1", len(rev))
	}
	ev := strings.Join(rev[0].Evidence, "")
	if strings.ContainsRune(ev, 0x1b) {
		t.Errorf("raw ESC survived into evidence: %q", ev)
	}
	if !strings.ContainsRune(ev, '?') {
		t.Errorf("control character was not replaced with the sanitiser caret: %q", ev)
	}
}
