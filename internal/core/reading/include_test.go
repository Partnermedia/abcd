package reading

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot walks up from the test's working directory to the directory holding
// go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
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

// recordFamilyRoots are the directories that hold a record family. Assembler
// rule 1 is stated against this list: no include may name a directory that
// CONTAINS one of them.
var recordFamilyRoots = []string{
	".abcd/development/decisions",
	".abcd/development/intents",
	".abcd/development/specs",
	".abcd/development/readings",
	".abcd/work/issues",
}

// TestReadingsCharterCarriesTheRenderedIncludeTable is the anti-drift detector:
// the charter and the Go table are one contract, and the code is its source of
// truth. Editing either alone fails here.
func TestReadingsCharterCarriesTheRenderedIncludeTable(t *testing.T) {
	path := filepath.Join(repoRoot(t), CharterPath)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", CharterPath, err)
	}
	doc := string(raw)

	begin := strings.Index(doc, MarkerBegin)
	end := strings.Index(doc, MarkerEnd)
	if begin < 0 || end < 0 {
		t.Fatalf("%s does not carry the generated-table markers %q / %q; the charter "+
			"is the include table's rendering, so it must render it", CharterPath, MarkerBegin, MarkerEnd)
	}
	if end < begin {
		t.Fatalf("%s: end marker precedes begin marker", CharterPath)
	}
	got := strings.TrimSpace(doc[begin+len(MarkerBegin) : end])
	want := strings.TrimSpace(Render())
	if got != want {
		t.Errorf("%s has drifted from reading.Table.\n\n--- charter has ---\n%s\n\n--- Render() wants ---\n%s",
			CharterPath, got, want)
	}
}

// TestEveryFramingRowCitesAdr55 holds adr-55's own consequence: the framing
// section's construal statement is admissible only because that ADR says so, so
// every row that admits it must name the rule admitting it.
func TestEveryFramingRowCitesAdr55(t *testing.T) {
	const framing = ".abcd/development/brief/01-product/06-framing.md"
	admitting := 0
	for _, row := range Table {
		if !row.Reaches(framing) {
			continue
		}
		admitting++
		if !strings.Contains(row.Rule, "adr-55") {
			t.Errorf("row %q admits the framing section but its rule %q does not cite adr-55",
				row.Source, row.Rule)
		}
	}
	if admitting == 0 {
		t.Fatalf("no row admits %s; the construal statement is committed record and must be passed", framing)
	}
}

// TestNoIncludeNamesARecordFamilyDirectory is assembler rule 1 as an executable
// property of the table: a row may name a family's own bucket individually, but
// no row may reach into a family from above.
func TestNoIncludeNamesARecordFamilyDirectory(t *testing.T) {
	for _, family := range recordFamilyRoots {
		for _, row := range Table {
			src := strings.TrimSuffix(row.Source, "/")
			if src == family || strings.HasPrefix(src+"/", family+"/") {
				continue // named individually, which rule 1 permits
			}
			for _, probe := range []string{family + "/probe.md", family + "/bucket/probe.md"} {
				if row.Reaches(probe) {
					t.Errorf("row %q reaches %s: an include may not name a directory containing a record family",
						row.Source, probe)
				}
			}
		}
	}
}

// TestTheAssemblersOwnOutputIsNeverItsInput is ruling (18): the instrument's own
// prior outputs, its definitions, the evals that guard it and the include table
// that decides what it sees are all unreachable at every position.
func TestTheAssemblersOwnOutputIsNeverItsInput(t *testing.T) {
	own := []string{
		".abcd/development/readings/rdg-2608301200000001/manifest.json",
		".abcd/development/readings/rdg-2608301200000001/run.json",
		".abcd/development/readings/README.md",
		".abcd/work/issues/readings/rdg-2608301200000001/rdi-1.md",
		"agents/cold-reading-widening.md",
		"evals/cold_reading_readblock_test.go",
		"internal/core/reading/include.go",
		"internal/core/reading/testdata/fixture.md",
	}
	for _, p := range Positions() {
		for _, rel := range own {
			if Admits(p, rel) {
				t.Errorf("position %s admits %s; the assembler's own output, definition, "+
					"eval and include table must never become its input", p, rel)
			}
		}
	}
}

// TestWideningExcludesDraftsAndPlannedEntailmentIncludesThem holds the one
// asymmetry the record insists is stated rather than remembered: a widening
// reading must not see the candidate set it is asked to widen, and articulation
// precedes selection.
func TestWideningExcludesDraftsAndPlannedEntailmentIncludesThem(t *testing.T) {
	candidates := []string{
		".abcd/development/intents/drafts/itd-900-a-draft.md",
		".abcd/development/intents/planned/itd-901-a-planned.md",
	}
	for _, rel := range candidates {
		if Admits(PositionWidening, rel) {
			t.Errorf("the widening position admits %s; it is the candidate set the reading is asked to widen", rel)
		}
		if Admits(PositionDetection, rel) {
			t.Errorf("the detection position admits %s; only entailment sees the candidate set", rel)
		}
		if !Admits(PositionEntailment, rel) {
			t.Errorf("the entailment position excludes %s; articulation precedes selection", rel)
		}
	}
}

// TestSupersededIntentsAreNeverAdmitted holds the exclusion floor's directory
// half at the one place the intent store makes it easy to get wrong.
func TestSupersededIntentsAreNeverAdmitted(t *testing.T) {
	const superseded = ".abcd/development/intents/superseded/itd-47-a-retired.md"
	for _, p := range Positions() {
		if Admits(p, superseded) {
			t.Errorf("position %s admits %s; superseded records are deliberation", p, superseded)
		}
	}
}

// TestBriefEvidenceChapterIsNeverAdmitted holds the chapter-level include bound:
// 03-evidence is deliberation, and a future brief-homed warm record must not
// walk in as "brief text".
func TestBriefEvidenceChapterIsNeverAdmitted(t *testing.T) {
	const evidence = ".abcd/development/brief/03-evidence/01-open-questions.md"
	for _, p := range Positions() {
		if Admits(p, evidence) {
			t.Errorf("position %s admits %s; the evidence chapter is deliberation", p, evidence)
		}
	}
}

// includeTableDigest is the sha256 of Render() at AssemblerVersion. Changing the
// table without moving the version fails the gate below.
const includeTableDigest = "7f19b6953a098544b519447db82e0da930b41f6546cf6436143e2d9871181d24"

// TestAssemblerVersionCoversTheIncludeTable pins the rendered table to the
// assembler version carried into every manifest: a manifest that names a version
// must describe the table that version rendered.
func TestAssemblerVersionCoversTheIncludeTable(t *testing.T) {
	sum := sha256.Sum256([]byte(Render()))
	got := hex.EncodeToString(sum[:])
	if got != includeTableDigest {
		t.Errorf("the include table has changed but AssemblerVersion is still %s.\n"+
			"Bump AssemblerVersion and set includeTableDigest to %q.", AssemblerVersion, got)
	}
}

// TestEveryRowNamesAKnownKindAndPosition holds the two closed vocabularies.
func TestEveryRowNamesAKnownKindAndPosition(t *testing.T) {
	if len(Table) == 0 {
		t.Fatal("the include table is empty; a reading with no input is not a blind reading")
	}
	kinds := map[Kind]bool{}
	for _, k := range Kinds() {
		kinds[k] = true
	}
	for _, row := range Table {
		if !kinds[row.Kind] {
			t.Errorf("row %q carries kind %q, which is not in the closed vocabulary", row.Source, row.Kind)
		}
		if len(row.Positions) == 0 {
			t.Errorf("row %q is admitted at no position", row.Source)
		}
		for _, p := range row.Positions {
			if _, err := ParsePosition(string(p)); err != nil {
				t.Errorf("row %q names position %q: %v", row.Source, p, err)
			}
		}
		if row.Rule == "" {
			t.Errorf("row %q states no admitting rule", row.Source)
		}
		if len(row.Match) == 0 {
			t.Errorf("row %q matches every file; inclusion is positive at every grain", row.Source)
		}
	}
}

// TestPositionTokenIsClosed holds the invocation interface's first operand.
func TestPositionTokenIsClosed(t *testing.T) {
	for _, p := range Positions() {
		if _, err := ParsePosition(string(p)); err != nil {
			t.Errorf("ParsePosition(%q): %v", p, err)
		}
	}
	for _, bad := range []string{"", "Widening", "widening ", "framing", "../widening", "widening; drop table"} {
		if _, err := ParsePosition(bad); err == nil {
			t.Errorf("ParsePosition(%q) accepted an unknown token", bad)
		}
	}
}

// TestExclusionFloorNamesEveryRecordedExclusion holds itd-183's exclusion list
// against the floor the manifest asserts, so an exclusion the record states
// cannot quietly stop being declared.
func TestExclusionFloorNamesEveryRecordedExclusion(t *testing.T) {
	want := []string{
		"origin",
		"production_mode",
		"Audit Notes",
		".abcd/development/brief/03-evidence",
		".abcd/development/decisions",
		".abcd/development/roadmap/rfcs",
		".abcd/development/intents/superseded",
		".abcd/work/issues",
		".abcd/development/plans",
		".abcd/development/research/notes",
		".abcd/development/readings",
	}
	joined := ""
	for _, e := range ExclusionsFor(PositionWidening) {
		joined += e.Rule + "\x00" + e.Signal + "\x00" + e.Detail + "\n"
	}
	for _, w := range want {
		if !strings.Contains(joined, "\x00"+w+"\n") {
			t.Errorf("the exclusion floor does not name %q", w)
		}
	}
}
