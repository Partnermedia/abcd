//go:build smoke || coldreading

package evals

// The coverage matrix: for every rule the assembler's contract carries, the
// mutation that removes it and the plant that dies when it does.
//
// It exists because the previous two review rounds found the same shape twice —
// an assertion with nothing underneath it — and the answer to that is not
// another floor. A floor protects an assertion from vacuity in the one dimension
// somebody thought of; a plant that dies when a rule is removed protects it in
// the dimension that matters, which is the rule's own. So this table is the
// eval's real claim, and everything above it is machinery for making the table
// true.
//
// Two things make it more than a comment (itd-195). Every row names sentinel
// classes that must exist and must be planted, and every declared class must be
// named by at least one row — so a plant no rule depends on, and a rule naming a
// plant that has gone, both fail here. And a row that no mutation can falsify
// carries its reason in Gap rather than being quietly omitted: a declared gap is
// worth more than a floor that turns out to be one dimension short.
//
// What it does NOT claim: that each Falsifier has been executed on every run.
// The ones marked in the record as watched red have been; the rest are the
// mutation each row is written against, stated so the next reader can run it.

import (
	"sort"
	"strings"
	"testing"
)

// How the eval notices a rule's removal.
const (
	// caughtLeak: a planted token reaches the bundle and assertion 1 names it.
	caughtLeak = "sentinel absence names the leaked class"
	// caughtFamily: an item's manifest path lands in a refused family, or under
	// no individually named include, and assertion 3 names it.
	caughtFamily = "family absence names the path"
	// caughtCarrier: a plant-bearing file, or its content, stops arriving and the
	// carrier floor names the assertions that would have gone vacuous.
	caughtCarrier = "the carrier floor names the disarmed assertion"
	// caughtControl: the negative control stops reporting one of its two classes.
	caughtControl = "the holed control reports the wrong class set"
	// caughtRefusal: the assembler refuses the run and the verb exits non-zero,
	// which the eval reports as a failed assembly rather than as a leak.
	caughtRefusal = "the assembly is refused and the verb exits non-zero"
)

// coverageRow is one rule of the assembler's contract.
type coverageRow struct {
	// Rule is the rule in the record's own terms.
	Rule string
	// Falsifier is the mutation to the assembler that removes the rule.
	Falsifier string
	// Caught is how this eval notices. Empty exactly when Gap is set.
	Caught string
	// Classes are the sentinel classes that die or leak. It may be empty for a
	// rule this eval catches by path rather than by plant.
	Classes []string
	// Gap is why no mutation falsifies this rule against this corpus. Set exactly
	// when Caught is empty. A row carrying one is declared unfalsifiable
	// coverage, not an open hole to be discovered later.
	Gap string
}

// coverage is the matrix. One row per rule, positive and negative alike.
var coverage = []coverageRow{
	// ---- the include table's positive rows ----
	{
		Rule:      "brief/01-product is admitted whole at every position",
		Falsifier: "delete the 01-product row from Table",
		Caught:    caughtCarrier,
		Classes:   []string{"WARM-KEY", "WARM-FIELD"},
	},
	{
		Rule:      "brief/02-constraints is admitted whole at every position",
		Falsifier: "delete the 02-constraints row from Table",
		Gap: "no plant sits in that chapter, so its removal loses input without disarming " +
			"an assertion; a leak eval has nothing to say about a row whose absence leaks nothing",
	},
	{
		Rule:      "brief/glossary is admitted whole at every position",
		Falsifier: "delete the glossary row from Table",
		Gap:       "no plant sits in the glossary, for the same reason as 02-constraints",
	},
	{
		Rule:      "intents/disciplines is admitted whole at every position",
		Falsifier: "delete the disciplines row from Table",
		Caught:    caughtCarrier,
		Classes:   []string{"WARM-KEY", "WARM-FIELD"},
	},
	{
		Rule:      "intents/shipped is admitted at every position",
		Falsifier: "delete the shipped row from Table",
		Caught:    caughtCarrier,
		Classes:   []string{"WARM-FIELD", "UNPROJECTED-SECTION"},
	},
	{
		Rule: "the shipped intent's projection is POSITIVE at field granularity: a section " +
			"the field list does not name stays behind",
		Falsifier: "delete `Fields: intentProjection` from the shipped row",
		Caught:    caughtLeak,
		Classes:   []string{"UNPROJECTED-SECTION"},
	},
	{
		Rule:      "specs are admitted whole at every position",
		Falsifier: "delete the specs row from Table",
		Caught:    caughtCarrier,
		Classes:   []string{"WARM-KEY", "WARM-FIELD"},
	},
	{
		Rule:      "intents/drafts are admitted at entailment",
		Falsifier: "delete the drafts row from Table",
		Caught:    caughtCarrier,
		Classes:   []string{"DRAFT-ORIGIN", "UNPROJECTED-SECTION"},
	},
	{
		Rule:      "the draft projection is positive at field granularity",
		Falsifier: "delete `Fields: intentProjection` from the drafts row",
		Caught:    caughtLeak,
		Classes:   []string{"UNPROJECTED-SECTION"},
	},
	{
		Rule:      "intents/planned are admitted at entailment",
		Falsifier: "delete the planned row from Table",
		Caught:    caughtCarrier,
		Classes:   []string{"UNPROJECTED-SECTION"},
	},
	{
		Rule:      "the planned projection is positive at field granularity",
		Falsifier: "delete `Fields: intentProjection` from the planned row",
		Caught:    caughtLeak,
		Classes:   []string{"UNPROJECTED-SECTION"},
	},
	{
		Rule:      "the shipped tree's delivered documentation and root prose are admitted",
		Falsifier: "delete the root `.md` row from Table",
		Caught:    caughtControl,
		Classes:   []string{"TRANSCRIPT"},
	},
	{
		Rule:      "the shipped tree's source is admitted",
		Falsifier: "delete the root `.go` row from Table",
		Gap: "the fixture's own Go file is cold; the two planted Go files sit under the " +
			"structural deny, where a different row of this matrix covers them",
	},
	{
		Rule:      "the shipped tree's configuration and build files are admitted",
		Falsifier: "delete the root config row from Table",
		Gap:       "no plant sits in the fixture's configuration or build files",
	},

	// ---- the per-position asymmetry ----
	{
		Rule: "a reading's object excludes what it exists to change: the widening, " +
			"comparative and detection readings do not see the drafts",
		Falsifier: "widen the drafts row's Positions to allPositions",
		Caught:    caughtLeak,
		Classes:   []string{"DRAFT-BODY"},
	},
	{
		Rule:      "the same asymmetry for the planned intents",
		Falsifier: "widen the planned row's Positions to allPositions",
		Caught:    caughtFamily,
	},

	// ---- the exclusion floor: keys ----
	{
		Rule:      "the `origin` frontmatter key never travels",
		Falsifier: "delete the origin row from Exclusions",
		Caught:    caughtLeak,
		Classes:   []string{"WARM-KEY", "DRAFT-ORIGIN"},
	},
	{
		Rule:      "the production-mode frontmatter key never travels",
		Falsifier: "delete the production_mode row from Exclusions",
		Caught:    caughtLeak,
		Classes:   []string{"WARM-KEY"},
	},

	// ---- the exclusion floor: headings ----
	{
		Rule:      "an Audit Notes heading never travels",
		Falsifier: "delete the Audit Notes row from Exclusions",
		Caught:    caughtLeak,
		Classes:   []string{"WARM-FIELD"},
	},
	{
		Rule:      "a scope-condition disposition never travels",
		Falsifier: "delete the Scope Condition Dispositions row from Exclusions",
		Caught:    caughtLeak,
		Classes:   []string{"WARM-FIELD"},
	},
	{
		Rule:      "an Open Questions heading never travels",
		Falsifier: "delete the Open Questions row from Exclusions",
		Caught:    caughtLeak,
		Classes:   []string{"WARM-FIELD"},
	},
	{
		Rule:      "a Why This Matters heading never travels",
		Falsifier: "delete the Why This Matters row from Exclusions",
		Caught:    caughtLeak,
		Classes:   []string{"WARM-FIELD"},
	},

	// ---- the exclusion floor: families ----
	//
	// Each of these needs a TWO-part falsifier, and that is a property of the
	// assembler rather than a shortcut here: exclusion by absence from the
	// positive walk and the fail-closed assertExclusions gate are two independent
	// mechanisms, so removing either alone changes no output.
	{
		Rule:      "the brief's evidence chapter is deliberation and never travels",
		Falsifier: "add an include row for brief/03-evidence and delete its Exclusions row",
		Caught:    caughtLeak,
		Classes:   []string{"DELIBERATION"},
	},
	{
		Rule:      "the decisions family never travels",
		Falsifier: "add an include row for development/decisions and delete its Exclusions row",
		Caught:    caughtLeak,
		Classes:   []string{"DELIBERATION"},
	},
	{
		Rule:      "roadmap/rfcs never travel",
		Falsifier: "add an include row for roadmap/rfcs and delete its Exclusions row",
		Caught:    caughtLeak,
		Classes:   []string{"DELIBERATION"},
	},
	{
		Rule:      "superseded intents never travel",
		Falsifier: "add an include row for intents/superseded and delete its Exclusions row",
		Caught:    caughtLeak,
		Classes:   []string{"DELIBERATION"},
	},
	{
		Rule:      "plans never travel",
		Falsifier: "add an include row for development/plans and delete its Exclusions row",
		Caught:    caughtLeak,
		Classes:   []string{"DELIBERATION"},
	},
	{
		Rule:      "research notes never travel",
		Falsifier: "add an include row for research/notes and delete its Exclusions row",
		Caught:    caughtLeak,
		Classes:   []string{"DELIBERATION"},
	},
	{
		Rule: "the issue ledger never travels in any state, reading records, dispositions, " +
			"admission and selection grounds and the lapse log included",
		Falsifier: "add an include row under work/issues and delete the work/issues Exclusions row",
		Caught:    caughtLeak,
		Classes:   []string{"DECISION", "EXHAUST", "GROUNDS"},
	},
	{
		Rule:      "the shared decision log never travels",
		Falsifier: "add an include row for work/DECISIONS.md and delete its Exclusions row",
		Caught:    caughtLeak,
		Classes:   []string{"DECISION"},
	},
	{
		Rule:      "the readings family — the instrument's own output — never travels",
		Falsifier: "add an include row for development/readings and delete its Exclusions row",
		Caught:    caughtRefusal,
		Classes:   []string{"EXHAUST"},
	},
	{
		Rule:      "the local ledger tier never travels: framing traces and declined construals",
		Falsifier: "add an include row for .work.local/ledger",
		Caught:    caughtLeak,
		Classes:   []string{"LEDGER-FRAMING"},
	},
	{
		Rule:      "the local scratch tier never travels: transcript-class text inside the repository",
		Falsifier: "add an include row for .work.local/scratch",
		Caught:    caughtLeak,
		Classes:   []string{"TRANSCRIPT"},
	},
	{
		Rule:      "the reading definitions are the instrument and never travel",
		Falsifier: "add an include row for agents and delete its Exclusions row",
		Caught:    caughtLeak,
		Classes:   []string{"DEFINITION"},
	},
	{
		Rule:      "the evals that guard this assembler never travel",
		Falsifier: "add an include row for evals and delete its Exclusions row",
		Caught:    caughtLeak,
		Classes:   []string{"INSTRUMENT"},
	},
	{
		Rule:      "a reading never receives the include table that decides what it sees",
		Falsifier: "empty denyPrefixes and delete the internal/core/reading Exclusions row",
		Caught:    caughtLeak,
		Classes:   []string{"ASSEMBLER-SOURCE"},
	},
	{
		Rule:      "the session-transcript store sits outside the repository and never travels",
		Falsifier: "none: the assembler holds no code path that walks HOME, so there is nothing to remove",
		Gap: "unfalsifiable by construction, and deliberately so. The plant in the fixture " +
			"HOME is what makes the class REACHABLE — it is keyed on the fixture's own " +
			"root-commit sha, exactly where the store would be — so the day a walk over " +
			"HOME is added, this row becomes falsifiable with no change to the corpus. " +
			"Until then the eval asserts an absence nothing could have produced",
		Classes: []string{"TRANSCRIPT"},
	},

	// ---- structural mechanisms ----
	{
		Rule:      "the `.abcd` namespace is denied structurally, from each row's own Source downward",
		Falsifier: "drop \".abcd\" from denySegments",
		Caught:    caughtRefusal,
	},
	{
		Rule:      "the record graph is what enumerates the record rows",
		Falsifier: "point storeNodeType at a node type the graph does not report",
		Caught:    caughtCarrier,
		Classes:   []string{"WARM-KEY", "WARM-FIELD", "UNPROJECTED-SECTION", "DRAFT-ORIGIN"},
	},
	{
		Rule:      "the bundle carries the text of what it passed, not merely a manifest naming it",
		Falsifier: "emit an empty Text for every BundleItem",
		Caught:    caughtCarrier,
		Classes:   []string{"WARM-KEY", "WARM-FIELD", "UNPROJECTED-SECTION", "DRAFT-ORIGIN"},
	},
	{
		Rule:      "assertExclusions refuses an item under a path-shaped exclusion the manifest asserts",
		Falsifier: "delete the assertExclusions call",
		Gap: "unfalsifiable in isolation: inclusion is positive, so removing the fail-closed " +
			"half admits nothing new and the output does not change. It is the second half " +
			"of every two-part family falsifier above, where its removal is exactly what " +
			"lets the leak through",
	},
	{
		Rule:      "refuseOwnArtefact refuses an admitted file carrying this assembler's own artefact tag",
		Falsifier: "delete the refuseOwnArtefact call",
		Gap: "unfalsifiable in isolation for the same reason: the fixture's prior manifest is " +
			"already denied by path, so nothing reaches the scan. It is what turns the " +
			"readings-family falsifier above into a refusal rather than a leak",
	},
	{
		Rule:      "the walk is intersected with the tracked set, so an untracked file never travels",
		Falsifier: "drop the trackedSet intersection",
		Gap: "the corpus commits every plant, so there is no untracked file to admit. The " +
			"complementary claim IS asserted: the anti-vacuity guard fails if a declared " +
			"plant is untracked, because the assembler would then never have seen it",
	},
	{
		Rule:      "an excluded frontmatter key is refused at any indentation, not only at column 0",
		Falsifier: "anchor excludedKeyLine at column 0",
		Gap: "unfalsifiable through this eval: the assembler FAILS CLOSED on an indented " +
			"excluded key that survives redaction, so a corpus carrying one makes the " +
			"assembly exit non-zero and the eval's own depth-agnostic branch is never " +
			"reached. The branch is kept because spc-64 asks for it and it costs nothing " +
			"to be right if that refusal ever softens, not because it has caught anything",
	},
}

// TestEveryAssemblerRuleHasAFalsifier keeps the coverage matrix honest against
// the rest of this package.
//
// It is the check that turns the matrix from a comment into a claim: a rule
// naming a plant that no longer exists fails here, a plant that no rule depends
// on fails here, and a row that is neither falsifiable nor declared unfalsifiable
// fails here. What it does not and cannot check is that the matrix names every
// rule the assembler HAS — that is a human reading of the include table against
// this list, and it is why the matrix carries each row's falsifier in full.
func TestEveryAssemblerRuleHasAFalsifier(t *testing.T) {
	requireOracleTables(t)

	declared := map[string]bool{}
	for _, c := range sentinelClasses {
		declared[c.Name] = true
	}
	named := map[string]bool{}
	gaps := 0

	for _, row := range coverage {
		switch {
		case row.Caught == "" && row.Gap == "":
			t.Errorf("the coverage row %q is neither falsifiable nor declared unfalsifiable; "+
				"a rule with no falsifier and no stated reason is the hole this matrix exists "+
				"to make visible", row.Rule)
		case row.Caught != "" && row.Gap != "":
			t.Errorf("the coverage row %q is both caught and declared a gap; it is one or the "+
				"other, and a row that hedges tells a reader nothing", row.Rule)
		}
		if row.Gap != "" {
			gaps++
		}
		if row.Falsifier == "" {
			t.Errorf("the coverage row %q names no mutation, so nobody can re-run it", row.Rule)
		}
		for _, name := range row.Classes {
			if !declared[name] {
				t.Errorf("the coverage row %q names the sentinel class %q, which no longer "+
					"exists in the plants table", row.Rule, name)
			}
			named[name] = true
		}
	}

	// A plant no rule depends on is a plant that proves nothing. It is either a
	// rule this matrix has not written down, or a plant that should go.
	var orphans []string
	for _, c := range sentinelClasses {
		if !named[c.Name] {
			orphans = append(orphans, c.Token())
		}
	}
	sort.Strings(orphans)
	if len(orphans) > 0 {
		t.Errorf("%d planted class(es) are named by no coverage row, so nothing says which "+
			"assembler rule they falsify: %s", len(orphans), strings.Join(orphans, ", "))
	}

	// The gap count is declared, so a row silently becoming unfalsifiable — the
	// exact way this eval would decay — has to be an explicit edit.
	const declaredGaps = 9
	if gaps != declaredGaps {
		t.Errorf("the matrix declares %d unfalsifiable row(s) and holds %d; a rule sliding "+
			"into or out of unfalsifiable coverage is the change this eval most needs said "+
			"out loud", declaredGaps, gaps)
	}
}
