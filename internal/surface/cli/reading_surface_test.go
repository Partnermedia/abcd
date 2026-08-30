package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/reading"
)

// readingRepo lays out the minimum an `abcd reading assemble` run needs, so the
// surface test proves the WIRING end to end rather than mocking the core.
func readingRepo(t *testing.T) string {
	t.Helper()
	return readingRepoAt(t, t.TempDir())
}

// readingRepoAt builds the same fixture at a caller-chosen root, so a test about
// PATH RESOLUTION can place the repository deep enough that its two candidate
// bases both stay inside the test's own temporary directory. A test that wrote
// outside it would leave state behind and pass or fail on the leavings.
func readingRepoAt(t *testing.T, root string) string {
	t.Helper()
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(rel, body string) {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(".abcd/record-lint.json", `{"schema_version": 1, "rules": {"record_schema": {"enabled": true,
  "severity": "blocker", "record_stores": {"itd": ".abcd/development/intents", "spc": ".abcd/development/specs"}}}}`)
	write(".abcd/development/brief/01-product/06-framing.md", "# Framing\n\n## Construal\n\nA gap in the record.\n")
	write(".abcd/development/brief/02-constraints/03-invariants.md", "# Invariants\n\n1. One core.\n")
	write(".abcd/development/specs/open/spc-1-a-spec.md", "---\nid: spc-1\n---\n\n# A spec\n\nThe mechanics.\n")
	write(".abcd/development/intents/shipped/itd-1-an-intent.md",
		"---\nid: itd-1\nspec_id: spc-1\n---\n\n# An intent\n\n## Press Release\n\nThe promise.\n")
	write(".abcd/development/intents/disciplines/itd-2-a-discipline.md",
		"---\nid: itd-2\n---\n\n# A discipline\n\nA standing commitment.\n")
	write(".abcd/development/intents/drafts/itd-3-a-draft.md",
		"---\nid: itd-3\n---\n\n# A draft\n\n## Press Release\n\nA candidate.\n")
	write(".abcd/development/brief/glossary/core/construal.md", "# Construal\n\nWhat it is treated as.\n")
	write("docs/reference/thing.md", "# Thing\n\nReference prose.\n")
	write("README.md", "# The repository\n")
	write("go.mod", "module example\n\ngo 1.24\n")

	gitCmd(t, root, "init", "-q", "-b", "main")
	gitCmd(t, root, "add", "-A")
	gitCmd(t, root, "commit", "-q", "-m", "fixture")
	return root
}

// treeOf lists every non-git path under root, so a test can assert that a dry
// run wrote nothing.
func treeOf(t *testing.T, root string) string {
	t.Helper()
	var out []string
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(root, p)
		if rerr != nil {
			return rerr
		}
		if rel == ".git" || strings.HasPrefix(rel, ".git"+string(filepath.Separator)) || info.IsDir() {
			return nil
		}
		out = append(out, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	sort.Strings(out)
	return strings.Join(out, "\n")
}

// TestReadingVerbReachesBothPlanes holds "wired or it isn't done": the verb is
// registered in the command tree AND reachable from the plugin markdown surface.
func TestReadingVerbReachesBothPlanes(t *testing.T) {
	root := NewRootCommand()
	var readingCmd, assembleCmd bool
	for _, c := range root.Commands() {
		if c.Name() != "reading" {
			continue
		}
		readingCmd = true
		for _, sub := range c.Commands() {
			if sub.Name() == "assemble" {
				assembleCmd = true
			}
		}
	}
	if !readingCmd {
		t.Fatal("the command tree registers no `reading` verb")
	}
	if !assembleCmd {
		t.Fatal("`reading` registers no `assemble` sub-verb")
	}

	raw, err := os.ReadFile(filepath.Join(repoRootFromTest(t), "commands", "reading.md"))
	if err != nil {
		t.Fatalf("read the plugin surface: %v", err)
	}
	body := string(raw)
	for _, want := range []string{"assemble", "--position", "--target"} {
		if !strings.Contains(body, want) {
			t.Errorf("commands/reading.md does not mention %q", want)
		}
	}
}

// repoRootFromTest walks up to the directory holding go.mod.
func repoRootFromTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found above the test working directory")
		}
		dir = parent
	}
}

// TestAssembleRefusesFreeTextOperands is the invocation interface's whole claim:
// position and target state, and nothing else. A free-text operand has nowhere
// to go, so there is no channel a framing could travel through.
func TestAssembleRefusesFreeTextOperands(t *testing.T) {
	repo := readingRepo(t)
	t.Chdir(repo)

	out, err := runCLIErr(t, "reading", "assemble", "read this against the ledger",
		"--position", "widening", "--target", "HEAD", "--dry-run")
	if err == nil {
		t.Fatalf("a free-text operand was accepted:\n%s", out)
	}
	if code := exitCodeOf(err); code != 2 {
		t.Errorf("a free-text operand exited %d, want 2", code)
	}
}

// TestPositionTokenIsClosed holds the first operand at the front door.
func TestPositionTokenIsClosed(t *testing.T) {
	repo := readingRepo(t)
	t.Chdir(repo)

	for _, bad := range []string{"framing", "Widening", "", "widening extra"} {
		args := []string{"reading", "assemble", "--target", "HEAD", "--dry-run"}
		if bad != "" {
			args = append(args, "--position", bad)
		}
		if out, err := runCLIErr(t, args...); err == nil {
			t.Errorf("position %q was accepted:\n%s", bad, out)
		}
	}
	if _, err := runCLIErr(t, "reading", "assemble", "--position", "widening", "--target", "HEAD", "--dry-run"); err != nil {
		t.Errorf("the widening position was refused: %v", err)
	}
}

// TestTargetRefusesBranchAndTag holds the second operand at the front door, and
// with it the non-zero exit the evals bind to.
func TestTargetRefusesBranchAndTag(t *testing.T) {
	repo := readingRepo(t)
	t.Chdir(repo)

	for _, bad := range []string{"main", "v1.0.0", "HEAD~1"} {
		out, err := runCLIErr(t, "reading", "assemble", "--position", "widening", "--target", bad, "--dry-run")
		if err == nil {
			t.Errorf("target %q was accepted:\n%s", bad, out)
			continue
		}
		if code := exitCodeOf(err); code == 0 {
			t.Errorf("target %q exited 0", bad)
		}
	}
	if out, err := runCLIErr(t, "reading", "assemble", "--position", "widening", "--dry-run"); err == nil {
		t.Errorf("a missing --target was accepted:\n%s", out)
	}
}

// TestAssembleDryRunWritesNothing holds the render-without-writing idiom at the
// front door.
func TestAssembleDryRunWritesNothing(t *testing.T) {
	repo := readingRepo(t)
	t.Chdir(repo)
	before := treeOf(t, repo)

	out := runCLI(t, "reading", "assemble", "--position", "widening", "--target", "HEAD", "--dry-run", "--json")
	var res reading.AssembleResult
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("the --json envelope does not decode: %v\n%s", err, out)
	}
	if res.Written {
		t.Error("a dry run reported a write")
	}
	if res.ItemCount == 0 {
		t.Error("a dry run assembled nothing")
	}
	if !strings.HasPrefix(res.RunID, "rdg-") {
		t.Errorf("run id %q does not carry the readings family tag", res.RunID)
	}
	if after := treeOf(t, repo); after != before {
		t.Errorf("a dry run changed the tree:\nbefore\n%s\nafter\n%s", before, after)
	}
}

// TestAssembleWritesTwoArtefactsIntoANamedDirectory holds the interface the
// read-block and amnesia evals bind to: the assembled input and the manifest as
// two separate files in a directory the operator names.
func TestAssembleWritesTwoArtefactsIntoANamedDirectory(t *testing.T) {
	repo := readingRepo(t)
	t.Chdir(repo)
	outDir := filepath.Join(t.TempDir(), "run")

	raw := runCLI(t, "reading", "assemble", "--position", "entailment", "--target", "HEAD",
		"--out", outDir, "--dry-run", "--json")
	var res reading.AssembleResult
	if err := json.Unmarshal(raw, &res); err != nil {
		t.Fatalf("the --json envelope does not decode: %v\n%s", err, raw)
	}
	if !res.Written {
		t.Fatal("an assembly into a named output directory reported no write")
	}
	bundleRaw, err := os.ReadFile(filepath.Join(outDir, reading.BundleFileName))
	if err != nil {
		t.Fatalf("read the assembled input: %v", err)
	}
	var bundle reading.Bundle
	if err := json.Unmarshal(bundleRaw, &bundle); err != nil {
		t.Fatalf("the assembled input does not decode: %v", err)
	}
	if len(bundle.Items) != res.ItemCount {
		t.Errorf("the assembled input holds %d items, the result reports %d", len(bundle.Items), res.ItemCount)
	}
	manifestRaw, err := os.ReadFile(filepath.Join(outDir, reading.ManifestFileName))
	if err != nil {
		t.Fatalf("read the manifest: %v", err)
	}
	if _, err := reading.DecodeManifest(manifestRaw); err != nil {
		t.Fatalf("the written manifest does not decode strictly: %v", err)
	}
}

// TestBareReadingRenders holds the parent's read-only status render.
func TestBareReadingRenders(t *testing.T) {
	repo := readingRepo(t)
	t.Chdir(repo)

	text := string(runCLI(t, "reading"))
	if !strings.Contains(text, reading.AssemblerVersion) {
		t.Errorf("the bare render does not name the assembler version:\n%s", text)
	}
	for _, p := range reading.Positions() {
		if !strings.Contains(text, string(p)) {
			t.Errorf("the bare render does not name the %s position:\n%s", p, text)
		}
	}
	out := runCLI(t, "reading", "--json")
	var status reading.Status
	if err := json.Unmarshal(out, &status); err != nil {
		t.Fatalf("the --json status does not decode: %v\n%s", err, out)
	}
	if status.AssemblerVersion != reading.AssemblerVersion {
		t.Errorf("status names assembler version %q", status.AssemblerVersion)
	}
}

// TestRelativeOutResolvesAgainstTheWorkingDirectory: the core takes a relative
// output directory against the repository root, which is right for the default
// it computes itself and wrong for a path an operator typed. Run from a
// subdirectory, `--out ../../x` must mean what the shell means by it.
func TestRelativeOutResolvesAgainstTheWorkingDirectory(t *testing.T) {
	repo := readingRepoAt(t, filepath.Join(t.TempDir(), "a", "b", "repo"))
	sub := filepath.Join(repo, "docs")
	t.Chdir(sub)

	raw := runCLI(t, "reading", "assemble", "--position", "widening", "--target", "HEAD",
		"--out", "../../outside-run", "--json")
	var res reading.AssembleResult
	if err := json.Unmarshal(raw, &res); err != nil {
		t.Fatalf("the --json envelope does not decode: %v\n%s", err, raw)
	}

	fromCwd := filepath.Join(filepath.Dir(filepath.Dir(sub)), "outside-run")
	fromRepoRoot := filepath.Join(filepath.Dir(filepath.Dir(repo)), "outside-run")
	if fromCwd == fromRepoRoot {
		t.Fatal("the two bases coincide, so this test proves nothing")
	}
	if _, err := os.Stat(filepath.Join(fromCwd, reading.BundleFileName)); err != nil {
		t.Errorf("the artefacts did not land beside the working directory: %v", err)
	}
	if _, err := os.Stat(filepath.Join(fromRepoRoot, reading.BundleFileName)); err == nil {
		t.Error("the artefacts landed against the repository root instead")
	}
	if !filepath.IsAbs(res.OutDir) {
		t.Errorf("the result echoes %q rather than the resolved path", res.OutDir)
	}
}
