//go:build smoke || coldreading

package evals

// The sentinel fixture corpus: what is planted, where, and how a variant of it
// is materialised into a repository the assembler can be run over.
//
// The corpus is inert on disk under testdata/cold-reading/ and materialised into
// a temporary directory per run, because the assembler reads a git working tree
// and refuses one whose included paths are uncommitted. Nothing here reads the
// assembler: the plants are declared, the counts are declared, and the eval is
// wrong about the assembler exactly when the assembler is wrong about the
// record.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/gittest"
)

// The four reading positions, spelled as the closed tokens the verb accepts.
const (
	posWidening    = "widening"
	posEntailment  = "entailment"
	posComparative = "comparative"
	posDetection   = "detection"
)

// everyPosition is the closed set, in charter order.
var everyPosition = []string{posWidening, posEntailment, posComparative, posDetection}

// fullyAsserted are the positions every assertion runs at: all four.
//
// The comparative position is included even though a comparative reading's
// in-cycle candidate set arrives by a channel this eval makes no claim about.
// The two are separable, and conflating them cost the eval a whole position:
// the artefact asserted here is the bundle `reading assemble --position
// comparative` wrote, and an assertion over those bytes claims nothing about
// any other channel. Leaving comparative out left six of the ten sentinel
// classes unasserted there, and left the oracle's own drafts-at-comparative
// exclusion a row that could never fire — so an assembler admitting the
// candidate set, or the local ledger tier, at that one position was green.
var fullyAsserted = everyPosition

// sentinelPrefix is the shape every planted token takes, so a leak names the
// warm location class that leaked rather than reading as ordinary prose.
const sentinelPrefix = "ABCD-EVAL-SENTINEL-"

// sentinelClass is one warm location class: its token, every home the corpus
// plants it in, and the positions (if any) at which it may legitimately reach a
// reading.
type sentinelClass struct {
	// Name completes the token: sentinelPrefix + Name.
	Name string
	// Homes are the planted files. A "repo:" home is repo-relative inside the
	// materialised fixture repository; a "home:" home is relative to the
	// materialised fixture HOME.
	Homes []string
	// Count is how many times the token appears across the materialised fixture.
	// It is declared rather than counted so a corpus that lost a plant fails.
	Count int
	// ColdAt are the positions at which this token is COLD — legitimately part of
	// the reading's object, so its presence in the bundle is not a leak. Empty
	// means the token is warm at every position.
	ColdAt []string
	// Why states the warm location class in the record's own terms.
	Why string
}

// Token is the class's planted token.
func (c sentinelClass) Token() string { return sentinelPrefix + c.Name }

// coldAt reports whether the class is cold at position p.
func (c sentinelClass) coldAt(p string) bool {
	for _, q := range c.ColdAt {
		if q == p {
			return true
		}
	}
	return false
}

// sentinelClasses is the plants table: one class per warm location class the
// record names, each carrying a distinct token.
//
// The counts are the anti-vacuity guard's oracle (TestEverySentinelIsPlanted).
// An absence assertion cannot see a corpus that lost its plants, so the corpus
// is asserted separately from the absence.
var sentinelClasses = []sentinelClass{
	{
		Name:  "LEDGER-FRAMING",
		Homes: []string{"repo:.abcd/.work.local/ledger/2026-08-30-declined-construals.md"},
		Count: 1,
		Why: "brief invariant 14: declined construals and every other framing trace live " +
			"on the local ledger side, and no reading consumes that side",
	},
	{
		Name: "TRANSCRIPT",
		Homes: []string{
			"repo:.abcd/.work.local/scratch/session-notes.md",
			"home:.abcd/history/ROOT_COMMIT_SHA/transcripts/ses-0001-a-stored-session.md",
		},
		Count: 2,
		Why: "brief invariant 15: the session-transcript store has an enumerated consumer " +
			"list, and the cold readings are denied it structurally",
	},
	{
		Name: "DECISION",
		Homes: []string{
			"repo:.abcd/work/DECISIONS.md",
			"repo:.abcd/development/decisions/adrs/0001-a-ruled-decision.md",
			"repo:.abcd/work/issues/open/iss-1-a-lapse-entry.md",
			"repo:.abcd/work/issues/resolved/iss-2-a-resolved-defect.md",
			"repo:.abcd/work/issues/wontfix/iss-3-a-declined-defect.md",
		},
		Count: 5,
		Why: "itd-183 exclusion list: decisions/, work/issues/ in every state, a wontfix " +
			"reason, and the lapse log",
	},
	{
		Name: "DELIBERATION",
		Homes: []string{
			"repo:.abcd/development/brief/03-evidence/01-open-questions.md",
			"repo:.abcd/development/intents/superseded/itd-5-a-retired-intent.md",
			"repo:.abcd/development/plans/2026-08-30-a-plan.md",
			"repo:.abcd/development/research/notes/2026-08-30-a-note.md",
			"repo:.abcd/development/roadmap/rfcs/0001-an-rfc.md",
		},
		Count: 5,
		Why: "itd-183 exclusion list: brief/03-evidence/, intents/superseded/, plans/, " +
			"research/notes/, roadmap/rfcs/ are deliberation, not committed construal",
	},
	{
		Name: "WARM-FIELD",
		Homes: []string{
			"repo:.abcd/development/brief/01-product/01-press-release.md",
			"repo:.abcd/development/intents/disciplines/itd-4-selection-criteria.md",
			"repo:.abcd/development/intents/shipped/itd-1-a-shipped-intent.md",
			"repo:.abcd/development/specs/open/spc-1-a-design-record.md",
		},
		Count: 7,
		Why: "itd-183: a heading the exclusion floor names, on a record type the include " +
			"list admits. Every one of the four excluded headings has a home on a record " +
			"type that travels WHOLE — Why This Matters in a brief chapter, Audit Notes in " +
			"a discipline, Open Questions and the scope-condition dispositions in a spec — " +
			"because on a PROJECTED record type the projection keeps the heading out " +
			"whatever the floor says, so deleting that heading's exclusion there leaks " +
			"nothing and the rule cannot be falsified. The three on the shipped intent are " +
			"the projected shape, kept so both shapes are exercised",
	},
	{
		Name: "UNPROJECTED-SECTION",
		Homes: []string{
			"repo:.abcd/development/intents/shipped/itd-1-a-shipped-intent.md",
			"repo:.abcd/development/intents/drafts/itd-2-a-draft-intent.md",
			"repo:.abcd/development/intents/planned/itd-3-a-planned-intent.md",
		},
		Count: 3,
		Why: "itd-183 assembler rule 2: the intent projection is POSITIVE at field " +
			"granularity, so a section it does not name stays behind — a section that is " +
			"neither projected nor on the exclusion floor is the only plant that can tell " +
			"a live projection from a dead one the redaction tidied up after",
	},
	{
		Name: "WARM-KEY",
		Homes: []string{
			"repo:.abcd/development/specs/open/spc-1-a-design-record.md",
			"repo:.abcd/development/intents/disciplines/itd-4-selection-criteria.md",
			"repo:.abcd/development/brief/01-product/01-press-release.md",
		},
		Count: 3,
		Why:   "itd-183 exclusion list: origin and production mode, detected by frontmatter key",
	},
	{
		Name: "EXHAUST",
		Homes: []string{
			"repo:.abcd/development/readings/rdg-2608300900000001/manifest.json",
			"repo:.abcd/work/issues/readings/rdi-1-a-prior-reading.md",
			"repo:.abcd/work/issues/dispositions/dsp-1-a-prior-disposition.md",
		},
		Count: 3,
		Why: "itd-183: manifests, reading records and dispositions are warm on the next " +
			"run, so the instrument's own output is never its input",
	},
	{
		Name: "GROUNDS",
		Homes: []string{
			"repo:.abcd/work/issues/admissions/adm-1-admission-grounds.md",
			"repo:.abcd/work/issues/admissions/adm-2-selection-grounds.md",
		},
		Count: 2,
		Why:   "itd-183 exclusion list: admission and selection grounds",
	},
	{
		Name:  "DEFINITION",
		Homes: []string{"repo:agents/cold-reading-widening.md"},
		Count: 1,
		Why:   "itd-183: the reading definitions are the instrument, not its input",
	},
	{
		Name:  "INSTRUMENT",
		Homes: []string{"repo:evals/read_block.go"},
		Count: 1,
		Why:   "itd-183: the evals that guard this assembler are the instrument",
	},
	{
		Name:  "ASSEMBLER-SOURCE",
		Homes: []string{"repo:internal/core/reading/include.go"},
		Count: 1,
		Why:   "itd-183, ruling (18): a reading never receives the include table that decides what it sees",
	},
	{
		Name:   "DRAFT-BODY",
		Homes:  []string{"repo:.abcd/development/intents/drafts/itd-2-a-draft-intent.md"},
		Count:  1,
		ColdAt: []string{posEntailment},
		Why: "itd-183's drafts asymmetry: articulation precedes selection, so entailment " +
			"reads the candidate set and the reading asked to widen it does not",
	},
	{
		Name:  "DRAFT-ORIGIN",
		Homes: []string{"repo:.abcd/development/intents/drafts/itd-2-a-draft-intent.md"},
		Count: 1,
		Why: "the drafts asymmetry moves the BODY, never the excluded key: origin is warm " +
			"at every position, the entailment reading included",
	},
}

// hole is one relocation the negative-control variant performs: a plant taken
// out of its canonical home and put into material the include table admits.
//
// The corpus is a baseline plus a declared set of holes rather than two full
// trees, so the two variants cannot drift apart in everything the hole does not
// touch — a second full tree would have to be edited twice on every change.
type hole struct {
	// Class is the sentinel class relocated, which is what the eval's failure
	// must name.
	Class string
	// From is the canonical home the plant is removed from, repo-relative.
	From string
	// To is the included file it lands in, repo-relative. The file's replacement
	// content lives under testdata/cold-reading/holed/ at the same path.
	To string
	// Why states why a correct assembler passes the relocated plant through.
	Why string
}

// holes is the negative control: two plants relocated into positively included
// material. A correct assembler passes both through, so the eval must report
// exactly these two violations — which is the permanent proof that the
// assertion can fail.
// It is derived from nothing: emptying it would make the negative control pass
// with no control in it, which is why TestReadBlockCatchesAHoledFirewall refuses
// an empty table outright.
var holes = []hole{
	{
		Class: "LEDGER-FRAMING",
		From:  ".abcd/.work.local/ledger/2026-08-30-declined-construals.md",
		To:    ".abcd/development/brief/01-product/06-framing.md",
		Why:   "the 01-product chapters are admitted wholesale at every position",
	},
	{
		Class: "TRANSCRIPT",
		From:  ".abcd/.work.local/scratch/session-notes.md",
		To:    "docs/reference/thing.md",
		Why:   "the shipped tree's delivered documentation is admitted wholesale",
	},
}

// carrier is one plant-bearing file the include table admits, and the positions
// it must actually reach the assembly at.
//
// The anti-vacuity guard below the corpus proves the plants are on disk and
// tracked. This proves their CONTENT is in the assembly, which is a different
// claim and the one every absence assertion here rests on. Two failures it has
// to see, and both have been watched: an assembler that stopped enumerating a
// whole class of source drops the files carrying WARM-FIELD, two of the three
// WARM-KEY plants and every candidate plant, and leaves a bundle that is still
// comfortably non-empty; and an assembler that emitted an EMPTY text for each
// item leaves the manifest describing exactly the same paths while the bundle
// carries nothing to leak. A floor of "any item" cannot see the first. A floor
// of "this path is in the manifest" cannot see the second.
//
// The list is transcribed from the same source as the include list above —
// itd-183's positive includes — never from the assembler.
type carrier struct {
	// Path is the repo-relative file, which must appear among the manifest's
	// item paths.
	Path string
	// Positions are the positions it must reach. Empty means every position.
	Positions []string
	// Markers are COLD strings drawn from the carrier's own travelling text: not
	// sentinels, so they belong in the bundle, and distinctive, so finding one in
	// the bundle's bytes means this file's content arrived rather than its name.
	//
	// Requiring a marker rather than the manifest path is the whole point. A
	// manifest names what an assembly SAYS it passed; an assembly that emitted an
	// empty text for every projected item would satisfy a path check exactly while
	// every absence assertion over those items asserted nothing at all.
	//
	// A marker pins ONLY the field it was drawn from, so a projected carrier draws
	// one from each field it pins, and the set of them is drawn from DISTINCT
	// fields. Every marker taken from the same section is the mistake that made
	// this floor land short twice: with all of them inside `## Press Release`,
	// narrowing the projection to that one field drops four of the five contracted
	// fields and every marker still arrives.
	Markers []string
	// Fields are the projected fields the manifest must record for this path.
	//
	// It exists because one of the five contracted fields cannot be pinned by a
	// marker AT ALL, and no number of markers could have found that. `spec_id`
	// projects to the bare string `spc-1`, which also travels inside the whole
	// spec file, so bytes.Contains is satisfied whether or not the projection
	// emitted it. The manifest's `field` column names what was projected, which is
	// the only place that distinction is visible — so the two checks are not
	// belt-and-braces, they reach different halves of the contract.
	Fields []string
	// Classes are the sentinel classes the file carries, so a missing carrier
	// names the assertions it silently disarmed. It may be empty for a carrier
	// whose job is to pin an include row rather than a plant.
	Classes []string
	// Why states what the file is doing in the assembly.
	Why string
}

// carriers pins the include list at content level: every row it names must
// arrive at each position, carrying its own text.
//
// Six rows are plant-bearing, so a lost carrier disarms a named assertion. Four
// carry no plant and exist only to make their include row falsifiable — the
// carrier floor is a PRESENCE assertion, not a leak assertion, so it reaches a
// row whose removal leaks nothing, which is a reach an absence assertion alone
// does not have.
var carriers = []carrier{
	{
		Path:    ".abcd/development/brief/01-product/01-press-release.md",
		Markers: []string{"The fixture product, stated as it presently stands."},
		Classes: []string{"WARM-KEY", "WARM-FIELD"},
		Why:     "a brief chapter admitted wholesale, carrying the production-mode key",
	},
	{
		Path:    ".abcd/development/brief/02-constraints/03-invariants.md",
		Markers: []string{"The core is transport agnostic."},
		Why:     "the constraints chapter, which carries no plant; the carrier is what makes its row falsifiable",
	},
	{
		Path:    ".abcd/development/brief/glossary/core/construal.md",
		Markers: []string{"What the situation is treated as."},
		Why:     "the glossary, which carries no plant; the carrier is what makes its row falsifiable",
	},
	{
		Path:    ".abcd/development/intents/disciplines/itd-4-selection-criteria.md",
		Markers: []string{"Six criteria, recorded as a standing commitment."},
		Classes: []string{"WARM-KEY", "WARM-FIELD"},
		Why:     "a discipline admitted whole, carrying the origin key",
	},
	{
		Path:    ".abcd/development/specs/open/spc-1-a-design-record.md",
		Markers: []string{"The mechanics the capability was built against."},
		Classes: []string{"WARM-KEY", "WARM-FIELD"},
		Why:     "a spec admitted whole, carrying the origin key and two excluded headings",
	},
	{
		// The only fixture record carrying all five contracted fields, so it is the
		// one that can pin the whole projection: four markers from four distinct
		// sections, and spec_id off the manifest because no marker can reach it.
		Path: ".abcd/development/intents/shipped/itd-1-a-shipped-intent.md",
		Markers: []string{
			"The promise, as it was made.",
			"- Given a fixture state, when the assembly runs, then the read-block holds.",
			"- Holds while the record is one repository.",
			"A positive include table over a hand-transcribed exclusion floor.",
		},
		Fields: []string{
			"Press Release", "Acceptance Criteria", "Scope Conditions", "Mechanism", "spec_id",
		},
		Classes: []string{"WARM-FIELD", "UNPROJECTED-SECTION"},
		Why:     "a shipped intent projected to its claim record, pinned at all five contracted fields",
	},
	{
		Path:      ".abcd/development/intents/drafts/itd-2-a-draft-intent.md",
		Positions: []string{posEntailment},
		Markers: []string{
			"The draft's press release, which the entailment reading reads.",
			"The draft's mechanism claim, which the entailment reading reads.",
		},
		Fields:  []string{"Press Release", "Mechanism"},
		Classes: []string{"DRAFT-ORIGIN", "UNPROJECTED-SECTION"},
		Why:     "the candidate set the entailment reading reads, carrying the origin key",
	},
	{
		Path:      ".abcd/development/intents/planned/itd-3-a-planned-intent.md",
		Positions: []string{posEntailment},
		Markers: []string{
			"A planned promise the entailment reading reads.",
			"The planned mechanism claim, which the entailment reading reads.",
		},
		Fields:  []string{"Press Release", "Mechanism", "spec_id"},
		Classes: []string{"UNPROJECTED-SECTION"},
		Why:     "the planned half of the candidate set, projected the same way",
	},
	{
		Path:    "main.go",
		Markers: []string{"func main() {}"},
		Why:     "the shipped tree's source, which carries no plant; the carrier is what makes its row falsifiable",
	},
	{
		Path:    "fence.go",
		Markers: []string{"This block opens a fence at the left margin and never closes it."},
		Why: "a source file carrying an unterminated fence at the left margin; it is what " +
			"makes the body redaction's markdown-only scope falsifiable",
	},
	{
		Path:    "go.mod",
		Markers: []string{"module example.invalid/coldreadingfixture"},
		Why:     "the shipped tree's configuration, which carries no plant; the carrier is what makes its row falsifiable",
	},
}

// stake says what a missing carrier costs: the assertions it would have
// disarmed, or — for a carrier that pins an include row rather than a plant —
// that its row can no longer be falsified.
func (c carrier) stake() string {
	if len(c.Classes) == 0 {
		return "no plant depends on it, so nothing would leak, and its include row would " +
			"become unfalsifiable without a test noticing"
	}
	tokens := make([]string, 0, len(c.Classes))
	for _, name := range c.Classes {
		tokens = append(tokens, sentinelPrefix+name)
	}
	return "the " + strings.Join(tokens, " and ") + " assertion(s) over it would pass with nothing to catch"
}

// reachesAt reports whether the carrier must reach the assembly at position p.
func (c carrier) reachesAt(p string) bool {
	if len(c.Positions) == 0 {
		return true
	}
	for _, q := range c.Positions {
		if q == p {
			return true
		}
	}
	return false
}

// The fixture corpus on disk.
const (
	corpusDir      = "testdata/cold-reading"
	baselineDir    = corpusDir + "/baseline"
	holedDir       = corpusDir + "/holed"
	fixtureHomeDir = corpusDir + "/home"
	// rootSHAPlaceholder is the directory name the fixture home carries where
	// the transcript store keys on the repository's root-commit sha. The sha
	// exists only once the fixture is committed, so materialisation substitutes
	// it.
	rootSHAPlaceholder = "ROOT_COMMIT_SHA"
)

// The two variants of the corpus.
const (
	variantBaseline = "baseline"
	variantHoled    = "holed"
)

// fixture is one materialised corpus: a committed repository, the HOME its run
// sees, and the root-commit sha the transcript store is keyed on.
type fixture struct {
	Root    string
	Home    string
	RootSHA string
	Variant string
}

// materialise copies the corpus into a temporary directory, applies the
// variant's holes, commits it under a fixture identity in a reserved example
// domain, and plants the transcript store in a temporary HOME keyed on the
// resulting root-commit sha.
func materialise(t *testing.T, variant string) fixture {
	t.Helper()
	requireOracleTables(t)
	base := t.TempDir()
	f := fixture{
		Root:    filepath.Join(base, "repo"),
		Home:    filepath.Join(base, "home"),
		Variant: variant,
	}
	copyTree(t, baselineDir, f.Root)
	if variant == variantHoled {
		for _, h := range holes {
			if err := os.Remove(filepath.Join(f.Root, filepath.FromSlash(h.From))); err != nil {
				t.Fatalf("holed variant: removing %s: %v", h.From, err)
			}
			copyFile(t,
				filepath.Join(holedDir, filepath.FromSlash(h.To)),
				filepath.Join(f.Root, filepath.FromSlash(h.To)))
		}
	}
	gitCommitFixture(t, f.Root)
	f.RootSHA = rootCommit(t, f.Root)
	copyTree(t, fixtureHomeDir, f.Home)
	renamePlaceholder(t, f.Home, f.RootSHA)
	return f
}

// gitInit and the one commit. The identity is invented and its domain is
// reserved for examples, so nothing about the machine the eval runs on reaches
// the fixture's history.
func gitCommitFixture(t *testing.T, root string) {
	t.Helper()
	ident := []string{"-c", "user.name=abcd cold-reading fixture", "-c", "user.email=fixture@example.invalid"}
	steps := [][]string{
		{"init", "-q", "-b", "main"},
		append(append([]string{}, ident...), "add", "-A"),
		append(append([]string{}, ident...), "commit", "-q", "-m", "the cold-reading fixture corpus"),
	}
	for _, args := range steps {
		gitFixture(t, root, args...)
	}
}

// gitFixture runs one git command in the fixture repository, failing the test
// on error.
//
// Every command runs under gittest.Env, the repository's shared hermetic-git
// environment (iss-28): the ambient GIT_DIR/GIT_WORK_TREE and the machine's
// global git configuration are stripped, so nothing about the machine running
// the eval can reach the fixture and no fixture command can be redirected onto
// the real repository. That helper reaches internal/gitutil and nothing else, so
// it costs the oracle none of its independence from the assembler.
func gitFixture(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	cmd.Env = gittest.Env(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s in the fixture: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// rootCommit returns the fixture's root-commit sha, which is what the
// session-transcript store is keyed on.
func rootCommit(t *testing.T, root string) string {
	t.Helper()
	sha := gitFixture(t, root, "rev-list", "--max-parents=0", "HEAD")
	if sha == "" {
		t.Fatal("the fixture repository reports no root commit")
	}
	return sha
}

// trackedFiles returns the fixture repository's tracked paths. The assembler
// intersects its walk with this set, so a plant git refused to track is a plant
// the assembler could never have leaked — a corpus that silently tests nothing.
func trackedFiles(t *testing.T, root string) map[string]bool {
	t.Helper()
	out := gitFixture(t, root, "ls-files", "-z")
	set := map[string]bool{}
	for _, p := range strings.Split(out, "\x00") {
		if p != "" {
			set[p] = true
		}
	}
	return set
}

// renamePlaceholder swaps the transcript store's placeholder directory for the
// fixture's own root-commit sha.
func renamePlaceholder(t *testing.T, home, sha string) {
	t.Helper()
	parent := filepath.Join(home, ".abcd", "history")
	from := filepath.Join(parent, rootSHAPlaceholder)
	if err := os.Rename(from, filepath.Join(parent, sha)); err != nil {
		t.Fatalf("keying the fixture transcript store on the root commit: %v", err)
	}
}

// copyTree copies a fixture directory wholesale. It is a walk rather than a
// shell copy so the eval behaves the same on every machine that runs it.
func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(src, path)
		if rerr != nil {
			return rerr
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		if merr := os.MkdirAll(filepath.Dir(target), 0o755); merr != nil {
			return merr
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("materialising %s: %v", src, err)
	}
}

// copyFile writes one fixture file over its destination, creating parents.
func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("reading the fixture file %s: %v", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatalf("creating the parent of %s: %v", dst, err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("writing %s: %v", dst, err)
	}
}
