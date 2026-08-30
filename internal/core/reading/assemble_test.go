package reading

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestExcludedFieldsNeverReachTheBundle is itd-183's first criterion: given a
// repository state, the assembler's output contains no field on the exclusion
// list. Every warm class is planted with its own sentinel, so a leak names
// itself.
func TestExcludedFieldsNeverReachTheBundle(t *testing.T) {
	root := fixtureRepo(t)
	banned := map[string]string{
		sentinelEvidence:    "the brief's evidence chapter",
		sentinelDecision:    "a decision record",
		sentinelIssue:       "the issue ledger",
		sentinelAuditNotes:  "a shipped intent's Audit Notes",
		sentinelWhyItMatter: "a shipped intent's Why This Matters",
		sentinelOrigin:      "the origin frontmatter key",
		sentinelProdMode:    "the production_mode frontmatter key",
		sentinelSuperseded:  "a superseded intent",
		sentinelPlan:        "a plan record",
		sentinelPriorRun:    "the instrument's own prior manifest",
		sentinelDefinition:  "a reading definition and its eval",
		sentinelLapse:       "the lapse log",
	}
	for _, p := range Positions() {
		res := assembleFixture(t, root, p)
		text := bundleText(res.Bundle)
		for token, what := range banned {
			if strings.Contains(text, token) {
				t.Errorf("position %s passed %s (%s)", p, what, token)
			}
		}
	}
}

// TestDraftBodyIsColdAtEntailmentAndWarmElsewhere is the drafts asymmetry
// measured on the assembled input rather than on the table alone.
func TestDraftBodyIsColdAtEntailmentAndWarmElsewhere(t *testing.T) {
	root := fixtureRepo(t)
	if text := bundleText(assembleFixture(t, root, PositionEntailment).Bundle); !strings.Contains(text, sentinelDraftBody) {
		t.Error("the entailment position did not pass the draft body; articulation precedes selection")
	}
	for _, p := range []Position{PositionWidening, PositionComparative, PositionDetection} {
		if text := bundleText(assembleFixture(t, root, p).Bundle); strings.Contains(text, sentinelDraftBody) {
			t.Errorf("position %s passed the draft body; only entailment sees the candidate set", p)
		}
	}
}

// TestNewRecordFamilyIsAbsentWithoutTableChange is itd-183's second criterion:
// a record type added under .abcd/development/ after the table was written is
// absent, and the assembler is not edited to make it so.
func TestNewRecordFamilyIsAbsentWithoutTableChange(t *testing.T) {
	root := fixtureRepo(t)
	const invented = "SENTINEL-INVENTED-FAMILY"
	writeFile(t, root, ".abcd/development/inventions/inv-1-a-new-record.md",
		"---\nid: inv-1\n---\n\n# An invented record\n\n"+invented+"\n")
	gitCommitAll(t, root)

	for _, p := range Positions() {
		res := assembleFixture(t, root, p)
		if strings.Contains(bundleText(res.Bundle), invented) {
			t.Errorf("position %s passed a record family invented after the table was written", p)
		}
		for _, path := range itemPaths(res.Manifest) {
			if strings.HasPrefix(path, ".abcd/development/inventions/") {
				t.Errorf("position %s named %s in the manifest", p, path)
			}
		}
	}
}

// TestShippedIntentProjectsFiveFieldsOnly holds field granularity: a shipped
// intent travels as its claim record and nothing else.
func TestShippedIntentProjectsFiveFieldsOnly(t *testing.T) {
	root := fixtureRepo(t)
	res := assembleFixture(t, root, PositionWidening)
	const rel = ".abcd/development/intents/shipped/itd-1-a-shipped-intent.md"

	var got []string
	for _, it := range res.Manifest.Items {
		if it.Path == rel {
			got = append(got, it.Field)
		}
	}
	want := []string{"Press Release", "Acceptance Criteria", "Scope Conditions", "Mechanism", "spec_id"}
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("shipped intent projected %v, want %v", got, want)
	}
}

// TestBundleCarriesNoRepositoryPath is itd-183's fifth criterion. Item text is
// necessarily prose and source that may quote paths of its own, so the claim is
// made where it is a claim: the bundle's own structure carries no location.
func TestBundleCarriesNoRepositoryPath(t *testing.T) {
	root := fixtureRepo(t)
	res := assembleFixture(t, root, PositionWidening)

	skeleton := res.Bundle
	skeleton.Items = nil
	for _, it := range res.Bundle.Items {
		it.Text = ""
		skeleton.Items = append(skeleton.Items, it)
	}
	raw, err := EncodeBundle(skeleton)
	if err != nil {
		t.Fatalf("encode bundle skeleton: %v", err)
	}
	serialised := string(raw)
	for _, fragment := range []string{"/", "\\", ".abcd", "internal", "docs", root} {
		if strings.Contains(serialised, fragment) {
			t.Errorf("the bundle's structure carries %q; only the manifest maps an item to a path", fragment)
		}
	}
	for _, it := range res.Bundle.Items {
		if !itemKeyRe.MatchString(it.ItemKey) {
			t.Errorf("item key %q is not an ordinal of the form itm-NNNN", it.ItemKey)
		}
	}
}

// TestWalkIsLexicographicAndByteStable is the determinism precondition itd-187's
// eval checks independently: one state assembled twice is one bundle.
func TestWalkIsLexicographicAndByteStable(t *testing.T) {
	root := fixtureRepo(t)
	first := assembleFixture(t, root, PositionEntailment)
	second := assembleFixture(t, root, PositionEntailment)

	a, err := EncodeBundle(first.Bundle)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	b, err := EncodeBundle(second.Bundle)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if string(a) != string(b) {
		t.Error("two assemblies of one state produced different bundles")
	}

	paths := itemPaths(first.Manifest)
	sorted := append([]string{}, paths...)
	sort.Strings(sorted)
	if strings.Join(paths, "\n") != strings.Join(sorted, "\n") {
		t.Error("the walk is not lexicographic over repo-relative paths")
	}
}

// TestAssembleRefusesDirtyIncludedPath holds the re-runnability the manifest
// promises: a dirty tree cannot be described by a commit reference.
func TestAssembleRefusesDirtyIncludedPath(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, "README.md", "# The repository\n\nEdited, uncommitted.\n")

	_, err := Assemble(AssembleRequest{RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true})
	if err == nil {
		t.Fatal("assembly over a dirty included path succeeded; the manifest would promise a re-run it cannot deliver")
	}
	if !strings.Contains(err.Error(), "README.md") {
		t.Errorf("the refusal does not name the dirty path: %v", err)
	}
}

// TestAssembleIgnoresDirtinessOutsideTheIncludeTable keeps the refusal
// proportionate: an edit the reading can never see does not block a run.
func TestAssembleIgnoresDirtinessOutsideTheIncludeTable(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, ".abcd/work/DECISIONS.md", "# Decisions\n\nedited, uncommitted.\n")

	if _, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true,
	}); err != nil {
		t.Fatalf("a dirty excluded path blocked the assembly: %v", err)
	}
}

// TestTargetRefusesBranchAndTag holds the second operand's closed shape.
func TestTargetRefusesBranchAndTag(t *testing.T) {
	root := fixtureRepo(t)
	for _, bad := range []string{"main", "v0.6.8", "HEAD~1", "origin/main", "", "abc", strings.Repeat("a", 41)} {
		_, err := Assemble(AssembleRequest{RepoRoot: root, Position: PositionWidening, Target: bad, DryRun: true})
		if err == nil {
			t.Errorf("target %q was accepted; only HEAD or a 7-to-40-digit hex sha may be named", bad)
		}
	}
	head := headOf(t, root)
	for _, good := range []string{"HEAD", head, head[:7], head[:12]} {
		if _, err := Assemble(AssembleRequest{
			RepoRoot: root, Position: PositionWidening, Target: good, DryRun: true,
		}); err != nil {
			t.Errorf("target %q was refused: %v", good, err)
		}
	}
}

// TestTargetMustResolveToHead holds the other half: a well-shaped sha that is
// not the tree in front of the assembler is refused, not silently assembled.
func TestTargetMustResolveToHead(t *testing.T) {
	root := fixtureRepo(t)
	_, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "0123456789abcdef", DryRun: true,
	})
	if err == nil {
		t.Fatal("a target that is not HEAD was accepted; assembly reads the working tree")
	}
}

// TestAssembleRefusesAnUnknownPosition holds the first operand at the core.
func TestAssembleRefusesAnUnknownPosition(t *testing.T) {
	root := fixtureRepo(t)
	if _, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: Position("framing"), Target: "HEAD", DryRun: true,
	}); err == nil {
		t.Fatal("an unknown position was accepted")
	}
}

// TestAssembleDryRunWritesNothing holds the render-without-writing idiom: with
// no output directory named, a dry run leaves the tree exactly as it found it.
func TestAssembleDryRunWritesNothing(t *testing.T) {
	root := fixtureRepo(t)
	before := treeSnapshot(t, root)

	res, err := Assemble(AssembleRequest{RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true})
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	if res.Written {
		t.Error("a dry run without an output directory reported a write")
	}
	if after := treeSnapshot(t, root); after != before {
		t.Error("a dry run without an output directory changed the tree")
	}
}

// TestAssembleWritesTwoSeparateArtefacts holds the interface the evals bind to:
// the assembled input and the manifest are two files in a directory the
// operator names.
func TestAssembleWritesTwoSeparateArtefacts(t *testing.T) {
	root := fixtureRepo(t)
	before := treeSnapshot(t, root)
	out := filepath.Join(t.TempDir(), "run")

	res, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "HEAD", OutDir: out, DryRun: true,
	})
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	if !res.Written {
		t.Fatal("an assembly into a named output directory reported no write")
	}
	bundleRaw, err := os.ReadFile(filepath.Join(out, BundleFileName))
	if err != nil {
		t.Fatalf("read the assembled input: %v", err)
	}
	manifestRaw, err := os.ReadFile(filepath.Join(out, ManifestFileName))
	if err != nil {
		t.Fatalf("read the manifest: %v", err)
	}
	if strings.Contains(string(bundleRaw), "\"path\"") {
		t.Error("the assembled input carries a path field")
	}
	m, err := DecodeManifest(manifestRaw)
	if err != nil {
		t.Fatalf("the written manifest does not decode strictly: %v", err)
	}
	if m.RunID != res.RunID {
		t.Errorf("the written manifest names run %s, the result names %s", m.RunID, res.RunID)
	}
	if treeSnapshot(t, root) != before {
		t.Error("writing into a named output directory changed the repository")
	}
}

// TestAssembleDefaultsToTheLocalTier holds the default artefact home.
func TestAssembleDefaultsToTheLocalTier(t *testing.T) {
	root := fixtureRepo(t)
	res, err := Assemble(AssembleRequest{RepoRoot: root, Position: PositionWidening, Target: "HEAD"})
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	want := DefaultRunDir + "/" + res.RunID
	if res.OutDir != want {
		t.Errorf("out dir is %q, want %q", res.OutDir, want)
	}
	for _, name := range []string{BundleFileName, ManifestFileName} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(want), name)); err != nil {
			t.Errorf("stat %s: %v", name, err)
		}
	}
}

// treeSnapshot lists every tracked and untracked path with its size, so a test
// can assert an assembly wrote nothing.
func treeSnapshot(t *testing.T, root string) string {
	t.Helper()
	var lines []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			return rerr
		}
		if strings.HasPrefix(rel, ".git"+string(filepath.Separator)) || rel == ".git" {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		lines = append(lines, rel+"\x00"+itoa(info.Size()))
		return nil
	})
	if err != nil {
		t.Fatalf("walk fixture: %v", err)
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// TestSourceFileWithAnUnterminatedFenceDoesNotAbortTheAssembly holds the scope
// of the two record-shaped signals. A heading and a frontmatter key are things a
// RECORD carries; a Go file carries neither, and parsing one as markdown makes a
// raw string literal holding a fence able to stop every assembly the repository
// can run.
func TestSourceFileWithAnUnterminatedFenceDoesNotAbortTheAssembly(t *testing.T) {
	root := fixtureRepo(t)
	// A raw string literal whose content opens a fence at the left margin. The
	// site scan reads it as an unterminated fenced block and refuses.
	writeFile(t, root, "fence.go", "package main\n\nconst usage = `\n```sh\nabcd reading\n`\n")
	gitCommitAll(t, root)

	res, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true,
	})
	if err != nil {
		t.Fatalf("a fence inside a Go raw string aborted the assembly: %v", err)
	}
	found := false
	for _, m := range res.Manifest.Items {
		if m.Path == "fence.go" {
			found = true
		}
	}
	if !found {
		t.Error("the source file did not reach the bundle")
	}
}

// TestAssembleRefusesADeletedIncludedPath closes the hole a
// collected-paths-only dirty check leaves. A file deleted in the working tree
// but present at the target commit is absent from the assembly and absent from
// the status intersection, so the manifest would name a commit whose content it
// did not read — the exact promise the dirty gate exists to keep.
func TestAssembleRefusesADeletedIncludedPath(t *testing.T) {
	root := fixtureRepo(t)
	if err := os.Remove(filepath.Join(root, "README.md")); err != nil {
		t.Fatalf("remove: %v", err)
	}
	_, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true,
	})
	if err == nil {
		t.Fatal("an included path deleted in the working tree did not refuse the assembly")
	}
	if !strings.Contains(err.Error(), "README.md") {
		t.Errorf("the refusal does not name the deleted path: %v", err)
	}
}

// TestIgnoredFilesNeverEnterTheAssembly closes the gap between the filesystem
// and the commit. A gitignored file is not part of the target commit and `git
// status` says nothing about it, so a filesystem walk would pass content the
// dirty gate cannot see and the manifest would name a commit whose content it
// did not read. Build output, a virtual environment and a vendored tree all
// land in this class.
func TestIgnoredFilesNeverEnterTheAssembly(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, ".gitignore", "build/\nignored.json\n")
	gitCommitAll(t, root)

	const canary = "SENTINEL-IGNORED-BUILD-OUTPUT"
	writeFile(t, root, "ignored.json", "{\"note\": \""+canary+"\"}\n")
	writeFile(t, root, "build/generated.go", "package build\n\n// "+canary+"\n")

	res, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true,
	})
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	if strings.Contains(bundleText(res.Bundle), canary) {
		t.Error("an ignored file reached the bundle; it is not part of the commit the manifest names")
	}
	for _, m := range res.Manifest.Items {
		if m.Path == "ignored.json" || strings.HasPrefix(m.Path, "build/") {
			t.Errorf("the manifest names the ignored path %s", m.Path)
		}
	}
}

// TestUntrackedIncludedFileRefusesTheAssembly is the other side of the same
// boundary: a file that is NOT ignored and not yet committed is a genuine
// divergence from the target commit, so it refuses rather than being quietly
// dropped.
func TestUntrackedIncludedFileRefusesTheAssembly(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, "docs/newcomer.md", "# A newcomer\n\nUncommitted.\n")

	_, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true,
	})
	if err == nil {
		t.Fatal("an untracked included file did not refuse the assembly")
	}
	if !strings.Contains(err.Error(), "docs/newcomer.md") {
		t.Errorf("the refusal does not name the untracked path: %v", err)
	}
}

// TestAssembleRefusesAnUnconfiguredRecordScan holds the fail-closed stance at
// the one place a silent success is worse than a refusal. Record enumeration
// runs through the record-lint configuration; without it every record row
// contributes nothing, and the run reports a clean assembly of a reading that
// saw none of the record it exists to read against.
func TestAssembleRefusesAnUnconfiguredRecordScan(t *testing.T) {
	root := fixtureRepo(t)
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(LintConfigPath))); err != nil {
		t.Fatalf("remove: %v", err)
	}
	gitCommitAll(t, root)

	_, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true,
	})
	if err == nil {
		t.Fatal("an unconfigured record scan assembled silently")
	}
	if !strings.Contains(err.Error(), LintConfigPath) {
		t.Errorf("the refusal does not name the missing configuration: %v", err)
	}
}

// TestAssembleRefusesAStoreTheConfigurationDoesNotName is the same refusal one
// level in: a configuration present but silent about a store the include table
// names is the same silent nothing, arriving by a different route.
func TestAssembleRefusesAStoreTheConfigurationDoesNotName(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, LintConfigPath, `{"schema_version": 1, "rules": {"record_schema":
  {"enabled": true, "severity": "blocker", "record_stores": {"itd": ".abcd/development/intents"}}}}`)
	gitCommitAll(t, root)

	_, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true,
	})
	if err == nil {
		t.Fatal("an include row naming an unconfigured store assembled silently")
	}
	if !strings.Contains(err.Error(), "spc") {
		t.Errorf("the refusal does not name the unconfigured store: %v", err)
	}
}

// The three shapes that slip past a first-occurrence, exact-match redaction.
// Each is a warm field the manifest asserts was refused, still in the bundle.

// TestDuplicateExcludedKeyRefusesTheFile: Fields keeps the first occurrence of a
// key and drops the rest silently, so redacting what it reports leaves the
// second copy in place.
func TestDuplicateExcludedKeyRefusesTheFile(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, ".abcd/development/specs/open/spc-2-duplicated.md",
		"---\nid: spc-2\norigin: cold\norigin: "+sentinelOrigin+"\n---\n\n# A spec\n\nProse.\n")
	gitCommitAll(t, root)

	_, err := Assemble(AssembleRequest{RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true})
	if err == nil {
		t.Fatal("a duplicated excluded key did not refuse the file; the second copy survives redaction")
	}
	if !strings.Contains(err.Error(), "origin") {
		t.Errorf("the refusal does not name the duplicated key: %v", err)
	}
}

// TestExcludedKeySurvivingRedactionRefusesTheFile: StripFrontmatter closes on a
// `---` PREFIX, so a four-dash close is cut from the body while Fields, which
// wants the delimiter exactly, reads no fields at all and redacts nothing.
func TestExcludedKeySurvivingRedactionRefusesTheFile(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, ".abcd/development/specs/open/spc-3-four-dashes.md",
		"---\nid: spc-3\norigin: "+sentinelOrigin+"\n----\n\n# A spec\n\nProse.\n")
	gitCommitAll(t, root)

	_, err := Assemble(AssembleRequest{RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true})
	if err == nil {
		t.Fatal("an excluded key survived redaction and was passed")
	}
	if !strings.Contains(err.Error(), "origin") {
		t.Errorf("the refusal does not name the surviving key: %v", err)
	}
}

// TestCaseVariantExcludedHeadingRefusesTheFile: the heading redaction matches a
// title exactly, so another spelling of the same heading travels whole.
func TestCaseVariantExcludedHeadingRefusesTheFile(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, ".abcd/development/specs/open/spc-4-case-variant.md",
		"---\nid: spc-4\n---\n\n# A spec\n\n## audit notes\n\n"+sentinelAuditNotes+"\n")
	gitCommitAll(t, root)

	_, err := Assemble(AssembleRequest{RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true})
	if err == nil {
		t.Fatal("a case-variant excluded heading was passed whole")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "audit notes") {
		t.Errorf("the refusal does not name the surviving heading: %v", err)
	}
}

// TestStagedRenameOutOfTheIncludeSetRefuses: a rename's SOURCE path is the one
// that was in the target commit. Discarding it leaves an included file that is
// neither in the bundle nor refused, and the manifest names HEAD for a bundle
// that is not HEAD.
func TestStagedRenameOutOfTheIncludeSetRefuses(t *testing.T) {
	root := fixtureRepo(t)
	gitRun(t, root, "mv", "docs/reference/thing.md", ".abcd/development/plans/thing.md")

	_, err := Assemble(AssembleRequest{RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true})
	if err == nil {
		t.Fatal("an included file renamed out of the include set did not refuse the assembly")
	}
	if !strings.Contains(err.Error(), "docs/reference/thing.md") {
		t.Errorf("the refusal does not name the rename's source path: %v", err)
	}
}

// TestWorktreeRenameRefuses holds the same case in the second status column,
// which the source-consuming branch has to read as well as the first.
func TestWorktreeRenameRefuses(t *testing.T) {
	root := fixtureRepo(t)
	gitRun(t, root, "mv", "README.md", "READYOU.md")
	gitRun(t, root, "reset", "-q")

	_, err := Assemble(AssembleRequest{RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true})
	if err == nil {
		t.Fatal("a worktree rename of an included file did not refuse the assembly")
	}
	if !strings.Contains(err.Error(), "README.md") {
		t.Errorf("the refusal does not name the renamed path: %v", err)
	}
}

// TestRetargetedStoreDirectoryRefuses: a store key present but pointed at a
// directory that is not there enumerates nothing and reports a clean run.
func TestRetargetedStoreDirectoryRefuses(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, LintConfigPath, `{"schema_version": 1, "rules": {"record_schema":
  {"enabled": true, "severity": "blocker", "record_stores": {"itd": ".abcd/development/intents",
   "spc": ".abcd/development/specifications"}}}}`)
	gitCommitAll(t, root)

	_, err := Assemble(AssembleRequest{RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true})
	if err == nil {
		t.Fatal("a store pointed at an absent directory assembled silently")
	}
	if !strings.Contains(err.Error(), "specifications") {
		t.Errorf("the refusal does not name the absent store directory: %v", err)
	}
}

// TestUncommittedLintConfigRefuses: the record configuration decides what the
// record scan sees, and it sits under the deny, so no include row puts it in the
// dirty set. An uncommitted retarget would silently reshape the assembly.
func TestUncommittedLintConfigRefuses(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, LintConfigPath, `{"schema_version": 1, "rules": {"record_schema":
  {"enabled": true, "severity": "blocker", "record_stores": {"itd": ".abcd/development/intents",
   "spc": ".abcd/development/specs"}}}}`)

	_, err := Assemble(AssembleRequest{RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true})
	if err == nil {
		t.Fatal("an uncommitted record configuration did not refuse the assembly")
	}
	if !strings.Contains(err.Error(), LintConfigPath) {
		t.Errorf("the refusal does not name the configuration: %v", err)
	}
}

// TestUntrackedFileInANewDirectoryRefuses: git collapses an untracked directory
// to a single entry, so an admitted file inside it is never named.
func TestUntrackedFileInANewDirectoryRefuses(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, "docs/newdir/page.md", "# A page\n\nUncommitted, in a new directory.\n")

	_, err := Assemble(AssembleRequest{RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true})
	if err == nil {
		t.Fatal("an untracked admitted file in a new directory did not refuse the assembly")
	}
	if !strings.Contains(err.Error(), "docs/newdir/page.md") {
		t.Errorf("the refusal does not name the untracked file: %v", err)
	}
}

// TestProjectionKeepsASubsectionOfAProjectedField: the redactor ends a section
// at the next heading of level <= its own, and the projection must agree. Ending
// at the next heading of ANY level drops a `###` under a projected `##` — the
// item is silently short and the manifest hashes the short version.
func TestProjectionKeepsASubsectionOfAProjectedField(t *testing.T) {
	root := fixtureRepo(t)
	const nested = "SENTINEL-NESTED-CRITERION"
	writeFile(t, root, ".abcd/development/intents/shipped/itd-6-nested.md",
		"---\nid: itd-6\nspec_id: spc-1\n---\n\n# A shipped intent\n\n"+
			"## Acceptance Criteria\n\n- A top-level criterion.\n\n"+
			"### A grouped set\n\n- "+nested+"\n\n"+
			"## Audit Notes\n\n"+sentinelAuditNotes+"\n")
	gitCommitAll(t, root)

	res := assembleFixture(t, root, PositionWidening)
	text := bundleText(res.Bundle)
	if !strings.Contains(text, nested) {
		t.Error("a subsection of a projected field was dropped; projection must end where the redactor ends")
	}
	if strings.Contains(text, sentinelAuditNotes) {
		t.Error("the projection ran past its own section into an excluded one")
	}
}

// TestAssembleRefusesANonEmptyOutputDirectory: the two artefacts are one run's
// evidence, and writing them beside another run's leaves a directory whose
// manifest describes only half of what is in it.
func TestAssembleRefusesANonEmptyOutputDirectory(t *testing.T) {
	root := fixtureRepo(t)
	out := filepath.Join(t.TempDir(), "run")
	writeFile(t, out, "leftover.txt", "from an earlier run\n")

	_, err := Assemble(AssembleRequest{RepoRoot: root, Position: PositionWidening, Target: "HEAD", OutDir: out})
	if err == nil {
		t.Fatal("a non-empty output directory was accepted")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("the refusal does not say what is wrong with the directory: %v", err)
	}

	empty := filepath.Join(t.TempDir(), "fresh")
	if _, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "HEAD", OutDir: empty,
	}); err != nil {
		t.Errorf("an absent output directory was refused: %v", err)
	}
	if _, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "HEAD", OutDir: empty,
	}); err == nil {
		t.Error("a second run over the same directory was accepted; it is no longer empty")
	}
}
