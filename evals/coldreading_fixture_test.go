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
			"repo:.abcd/development/intents/shipped/itd-1-a-shipped-intent.md",
		},
		Count: 3,
		Why: "itd-183 assembler rule 2: a shipped intent travels as its claim record, so " +
			"its Audit Notes, its scope-condition dispositions and every other heading stay behind",
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
// tracked. This proves they are IN THE ASSEMBLY, which is a different claim and
// the one every absence assertion here rests on: an assembler that stopped
// enumerating a whole class of source would drop the files carrying WARM-FIELD,
// two of the three WARM-KEY plants and both draft plants, leave a bundle that is
// still comfortably non-empty, and turn every assertion green for the reason
// they exist to catch. A floor of "any item" cannot see that; a floor naming the
// carriers can.
//
// The list is transcribed from the same source as the include list above —
// itd-183's positive includes — never from the assembler.
type carrier struct {
	// Path is the repo-relative file, which must appear among the manifest's
	// item paths.
	Path string
	// Positions are the positions it must reach. Empty means every position.
	Positions []string
	// Class is the sentinel class the file carries, so a missing carrier names
	// the assertion it silently disarmed.
	Class string
	// Why states what the file is doing in the assembly.
	Why string
}

// carriers is the plant-bearing half of the include list.
var carriers = []carrier{
	{
		Path:  ".abcd/development/brief/01-product/01-press-release.md",
		Class: "WARM-KEY",
		Why:   "a brief chapter admitted wholesale, carrying the production-mode key",
	},
	{
		Path:  ".abcd/development/intents/disciplines/itd-4-selection-criteria.md",
		Class: "WARM-KEY",
		Why:   "a discipline admitted whole, carrying the origin key",
	},
	{
		Path:  ".abcd/development/specs/open/spc-1-a-design-record.md",
		Class: "WARM-KEY",
		Why:   "a spec admitted whole, carrying the origin key",
	},
	{
		Path:  ".abcd/development/intents/shipped/itd-1-a-shipped-intent.md",
		Class: "WARM-FIELD",
		Why:   "a shipped intent projected to its claim record, carrying the three headings that stay behind",
	},
	{
		Path:      ".abcd/development/intents/drafts/itd-2-a-draft-intent.md",
		Positions: []string{posEntailment},
		Class:     "DRAFT-ORIGIN",
		Why:       "the candidate set the entailment reading reads, carrying the origin key",
	},
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
