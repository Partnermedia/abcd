// Package reading assembles the input a cold reading is handed.
//
// Blindness is a property of the input, not a promise the reader makes. The
// include table below is the whole of what a reading may see: what it does not
// name is absent, including a record family invented after the table was
// written, and including this instrument's own output (itd-183, spc-61).
//
// The package is cobra-free and stdout-free like every sibling under
// internal/core (adr-23): it takes a structured request and returns a
// structured result, and the front doors format it.
package reading

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// AssemblerVersionCore is the hand-set semver of the assembly contract: the
// include table, the projection, the bundle shape and the manifest shape
// together. It moves when the contract moves, which a digest cannot detect —
// a rewritten rule text moves the rendering without changing what the
// assembler promises, and a projection change alters the promise without
// touching the table (spc-61, ruling (12); spc-68).
const AssemblerVersionCore = "1.1.0"

// AssemblerVersion is the core semver with the rendered include table's digest
// as semver build metadata. The digest is computed, not declared, so a table
// change moves the stamped version whether or not anyone notices: a manifest
// cannot name a version that does not describe the table it was built from.
//
// This is structural where the previous gate was advisory. That gate compared
// the rendering's digest to a standalone literal and never read the version at
// all, so changing the table and restating the literal was green with the
// version unmoved (iss-2608311949385350) — an attestation asserting more than
// its examination establishes, which brief invariant 16 forbids.
func AssemblerVersion() string {
	sum := sha256.Sum256([]byte(Render()))
	return AssemblerVersionCore + "+" + hex.EncodeToString(sum[:])[:12]
}

// Position is the reading position an assembly is invoked at. The set is
// closed: an unknown token is refused by name, never defaulted.
type Position string

const (
	// PositionWidening reads for configurations the record has not considered.
	// It is the one position that must not see the candidate set it is asked to
	// widen, which is why its rows exclude the draft and planned intents.
	PositionWidening Position = "widening"
	// PositionEntailment reads for what the record already commits to. It sees
	// the draft and planned intents because articulation precedes selection.
	PositionEntailment Position = "entailment"
	// PositionComparative reads two or more configurations against the
	// selection-criteria discipline.
	PositionComparative Position = "comparative"
	// PositionDetection registers detections against the object.
	PositionDetection Position = "detection"
)

// Positions lists every position, in the order the charter renders them.
func Positions() []Position {
	return []Position{PositionWidening, PositionEntailment, PositionComparative, PositionDetection}
}

// ParsePosition resolves a token to a position, refusing anything else by name.
func ParsePosition(s string) (Position, error) {
	for _, p := range Positions() {
		if string(p) == s {
			return p, nil
		}
	}
	names := make([]string, 0, len(Positions()))
	for _, p := range Positions() {
		names = append(names, string(p))
	}
	return "", fmt.Errorf("unknown reading position %q; the set is closed: %s",
		s, strings.Join(names, ", "))
}

// Kind is a bundle item's material class. It names WHAT a passed item is and
// never WHERE it came from: the vocabulary is closed, and no member of it
// carries a location. The path mapping lives in the manifest alone, so an
// auditor can resolve an item to its source and the reading cannot
// (brief invariant 15).
type Kind string

const (
	KindBriefSection     Kind = "brief-section"
	KindGlossaryTerm     Kind = "glossary-term"
	KindIntentProjection Kind = "intent-projection"
	KindDiscipline       Kind = "discipline"
	KindSpec             Kind = "spec"
	KindSource           Kind = "source"
	KindTest             Kind = "test"
	KindDoc              Kind = "doc"
	KindConfig           Kind = "config"
)

// Kinds lists the closed material-class vocabulary.
func Kinds() []Kind {
	return []Kind{KindBriefSection, KindGlossaryTerm, KindIntentProjection,
		KindDiscipline, KindSpec, KindSource, KindTest, KindDoc, KindConfig}
}

// Row is one include-table row: which positions admit this source, what is
// matched inside it, which fields are projected out of each matched file, the
// material class the items carry, and the rule that admits the row.
//
// A row's Source is the ONLY directory it reaches. The structural deny is
// measured from the Source downward, which is exactly why naming a record
// family's leaf bucket individually is legal while naming a directory that
// CONTAINS a family is not (assembler rule 1, ruling (18)).
type Row struct {
	// Positions are the reading positions this row is admitted at.
	Positions []Position
	// Source is the repo-relative directory the row reaches, "." for the
	// repository root.
	Source string
	// Match selects files inside Source: an entry beginning with "." is a file
	// extension, any other entry is an exact basename. An empty Match admits
	// every file, which no row uses — inclusion is positive at every grain.
	Match []string
	// MatchSuffix selects files inside Source by basename suffix, matched
	// case-sensitively. It is a separate field rather than a third convention
	// inside Match so the form is named by where it sits rather than inferred
	// from a string's first character: no disambiguation rule against the two
	// Match forms is needed, and none exists to get wrong. The case rule is
	// deliberate and differs from Match's extension form, which folds case:
	// the Go toolchain recognises only a lowercase _test.go as a test file, so
	// folding here would label material a test that Go does not build as one.
	// A file matched by either field is admitted; the two are ORed (spc-68).
	MatchSuffix []string
	// Fields are the named fields projected out of each matched file, in the
	// order they are emitted. A field is resolved as a heading section where the
	// file carries that heading, otherwise as a frontmatter key. An empty Fields
	// passes the whole file as one item.
	Fields []string
	// Store and Bucket, when set, route the row's enumeration through the
	// record graph (internal/core/lint's LoadRecordGraph, the one parser of the
	// record's shape in this binary) rather than through a file walk. Bucket is
	// the lifecycle directory, empty for every bucket of the store.
	Store  string
	Bucket string
	// Kind is the material class every item from this row carries.
	Kind Kind
	// Rule is the rule that admits the row, quoted in the charter.
	Rule string
}

// AdmittedAt reports whether the row is admitted at p.
func (r Row) AdmittedAt(p Position) bool {
	for _, q := range r.Positions {
		if q == p {
			return true
		}
	}
	return false
}

// intentProjection is the field set a shipped, planned or draft intent travels
// as. It is positive at field granularity: an intent's Audit Notes, its Open
// Questions and its scope-condition dispositions carry no row here, so they
// cannot travel however the record's shape changes around them.
//
// The list is a CONTRACT, not a census, and the two differ per POSITION today.
// `Scope Conditions` and `Mechanism` are spc-55's headings: no shipped intent
// carries them, so the shipped rows yield three fields, while two drafts already
// do, so the entailment position — the only one that reads drafts — projects all
// five. A field the file does not carry simply contributes no item, which is
// what lets one projection describe a record whose sections the record is still
// growing, and what keeps this list from needing an edit on the day the rest of
// the corpus catches up.
var intentProjection = []string{
	"Press Release",
	"Acceptance Criteria",
	"Scope Conditions",
	"Mechanism",
	"spec_id",
}

// allPositions is the shared floor's position set.
var allPositions = []Position{PositionWidening, PositionEntailment, PositionComparative, PositionDetection}

// Table is the include table: the single source of truth for what a cold
// reading may see. It is rendered into the readings charter under a test
// asserting the two agree, on the idiom internal/core/lifeboat/mapping.go
// carries for the brief.
//
// Order is significant only for rendering and for tie-breaking a path two rows
// admit: the first row that reaches a path owns the projection applied to it.
var Table = []Row{
	{
		Positions: allPositions,
		Source:    ".abcd/development/brief/01-product",
		Match:     []string{".md"},
		Kind:      KindBriefSection,
		Rule: "adr-55: the construal as it presently stands is committed record, " +
			"admissible to every reader including a cold reading",
	},
	{
		Positions: allPositions,
		Source:    ".abcd/development/brief/02-constraints",
		Match:     []string{".md"},
		Kind:      KindBriefSection,
		Rule: "The constraints chapter states the platform, the dependency stance, " +
			"the invariants and the naming a reading reads against",
	},
	{
		Positions: allPositions,
		Source:    ".abcd/development/brief/glossary",
		Match:     []string{".md"},
		Kind:      KindGlossaryTerm,
		Rule: "adr-55: the glossary's committed terms are committed record; " +
			"superseded terms and the reasoning that settled them are not",
	},
	{
		Positions: allPositions,
		Source:    ".abcd/development/intents/disciplines",
		Match:     []string{".md"},
		Store:     "itd",
		Bucket:    "disciplines",
		Kind:      KindDiscipline,
		Rule: "A discipline is a standing commitment the record already holds, " +
			"named individually inside the intent family",
	},
	{
		Positions: allPositions,
		Source:    ".abcd/development/intents/shipped",
		Match:     []string{".md"},
		Store:     "itd",
		Bucket:    "shipped",
		Fields:    intentProjection,
		Kind:      KindIntentProjection,
		Rule: "Assembler rule 2: a shipped intent travels as its claim record, " +
			"so the Audit Notes and dispositions it also carries stay behind",
	},
	{
		Positions: allPositions,
		Source:    ".abcd/development/specs",
		Match:     []string{".md"},
		Store:     "spc",
		Kind:      KindSpec,
		Rule:      "The design record a capability was built against",
	},
	{
		Positions: []Position{PositionEntailment},
		Source:    ".abcd/development/intents/drafts",
		Match:     []string{".md"},
		Store:     "itd",
		Bucket:    "drafts",
		Fields:    intentProjection,
		Kind:      KindIntentProjection,
		Rule: "Assembler rule 2: articulation precedes selection, so entailment sees " +
			"the candidate set and the reading asked to widen it does not",
	},
	{
		Positions: []Position{PositionEntailment},
		Source:    ".abcd/development/intents/planned",
		Match:     []string{".md"},
		Store:     "itd",
		Bucket:    "planned",
		Fields:    intentProjection,
		Kind:      KindIntentProjection,
		Rule: "Assembler rule 2: articulation precedes selection, so entailment sees " +
			"the candidate set and the reading asked to widen it does not",
	},
	{
		// Ordered above the .go row deliberately: path.Ext("foo_test.go") is
		// ".go", so the source row would otherwise claim every test file, and
		// the first row that reaches a path owns it. Both rows admit — this
		// row narrows nothing and widens nothing, it only labels.
		Positions:   allPositions,
		Source:      ".",
		MatchSuffix: []string{"_test.go"},
		Kind:        KindTest,
		Rule: "Assembler rule 1: the shipped tree is source and tests, counted apart " +
			"because tests are the largest single class and admitted identically",
	},
	{
		Positions: allPositions,
		Source:    ".",
		Match:     []string{".go"},
		Kind:      KindSource,
		Rule: "Assembler rule 1: the shipped tree is source and tests, with the record, " +
			"the definitions, the evals and the assembler's own package denied structurally",
	},
	{
		Positions: allPositions,
		Source:    ".",
		Match:     []string{".md"},
		Kind:      KindDoc,
		Rule: "Assembler rule 1: the shipped tree is the delivered documentation and the " +
			"root prose, with the record denied structurally",
	},
	{
		Positions: allPositions,
		Source:    ".",
		Match:     []string{".json", ".yml", ".yaml", ".toml", ".mod", ".sum", "Makefile"},
		Kind:      KindConfig,
		Rule: "Assembler rule 1: the shipped tree is the delivered configuration and build " +
			"files, with the record denied structurally",
	},
}

// Exclusions is the exclusion floor: every field, heading and directory the
// assembler refuses, each with the signal by which a reader detects it. It is
// asserted into every manifest so a reader can check the exclusions rather than
// trust a disclosure.
//
// The floor is a DECLARATION a reader checks, and the assembler additionally
// refuses to emit an item under any path-shaped entry, so the two cannot part
// company (see assertExclusions).
var Exclusions = []Exclusion{
	{Rule: "field projection", Signal: "frontmatter key", Detail: "origin"},
	{Rule: "field projection", Signal: "frontmatter key", Detail: "production_mode"},
	{Rule: "field projection", Signal: "heading", Detail: "Audit Notes"},
	{Rule: "field projection", Signal: "heading", Detail: "Open Questions"},
	{Rule: "field projection", Signal: "heading", Detail: "Why This Matters"},
	{Rule: "a reading's object excludes what it exists to change", Signal: "heading", Detail: "Scope Condition Dispositions"},
	{Rule: "absent from the positive walk", Signal: "directory", Detail: ".abcd/development/brief/03-evidence"},
	{Rule: "absent from the positive walk", Signal: "directory", Detail: ".abcd/development/decisions"},
	{Rule: "absent from the positive walk", Signal: "directory", Detail: ".abcd/development/roadmap/rfcs"},
	{Rule: "absent from the positive walk", Signal: "directory", Detail: ".abcd/development/intents/superseded"},
	{Rule: "absent from the positive walk", Signal: "directory", Detail: ".abcd/development/plans"},
	{Rule: "absent from the positive walk", Signal: "directory", Detail: ".abcd/development/research/notes"},
	{Rule: "no include names a directory containing a record family", Signal: "directory", Detail: ".abcd/work/issues"},
	{Rule: "absent from the positive walk", Signal: "file", Detail: ".abcd/work/DECISIONS.md"},
	{Rule: "absent from the positive walk", Signal: "record type in a denied path", Detail: "the lapse log"},
	{Rule: "absent from the positive walk", Signal: "record type in a denied path", Detail: "admission and selection grounds"},
	{Rule: "the instrument's own output is never its input", Signal: "directory", Detail: ".abcd/development/readings"},
	{Rule: "the instrument's own output is never its input", Signal: "directory", Detail: "agents"},
	{Rule: "the instrument's own output is never its input", Signal: "directory", Detail: "evals"},
	{Rule: "the instrument's own output is never its input", Signal: "directory", Detail: "internal/core/reading"},
	{Rule: "the store sits outside the repository tree", Signal: "unreachable path", Detail: "the session-transcript store"},
	{
		Rule:      "a reading's object excludes what it exists to change",
		Signal:    "directory",
		Detail:    ".abcd/development/intents/drafts",
		Positions: []Position{PositionWidening, PositionComparative, PositionDetection},
	},
	{
		Rule:      "a reading's object excludes what it exists to change",
		Signal:    "directory",
		Detail:    ".abcd/development/intents/planned",
		Positions: []Position{PositionWidening, PositionComparative, PositionDetection},
	},
}

// Exclusion is one refused source and the signal that detects it.
type Exclusion struct {
	// Rule is the assembler rule the exclusion instances.
	Rule string `json:"rule"`
	// Signal is how a reader detects it: a frontmatter key, a heading, a
	// directory absent from the positive walk, or a store outside the tree.
	Signal string `json:"signal"`
	// Detail names the excluded thing.
	Detail string `json:"detail"`
	// Positions are the positions the exclusion binds at. An empty Positions
	// binds at every position.
	Positions []Position `json:"-"`
}

// ExclusionsFor returns the exclusions binding at p, in table order.
func ExclusionsFor(p Position) []Exclusion {
	out := []Exclusion{}
	for _, e := range Exclusions {
		if len(e.Positions) == 0 {
			out = append(out, e)
			continue
		}
		for _, q := range e.Positions {
			if q == p {
				out = append(out, e)
				break
			}
		}
	}
	return out
}

// Marker delimiters for the rendered table in the readings charter.
const (
	MarkerBegin = "<!-- BEGIN GENERATED: reading-include-table -->"
	MarkerEnd   = "<!-- END GENERATED: reading-include-table -->"
)

// CharterPath is the readings family's charter, which renders the table.
const CharterPath = ".abcd/development/readings/README.md"

// Render renders the include table and the exclusion floor as the markdown the
// charter carries between the markers.
func Render() string {
	var b strings.Builder
	b.WriteString("### Include table\n\n")
	b.WriteString("| Positions | Source | Matches | Suffixes | Fields | Kind | Admitting rule |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- |\n")
	for _, row := range Table {
		fmt.Fprintf(&b, "| %s | `%s` | %s | %s | %s | `%s` | %s |\n",
			sortedPositions(row.Positions),
			row.Source,
			matchList(row),
			suffixList(row.MatchSuffix),
			fieldList(row.Fields),
			row.Kind,
			row.Rule)
	}
	b.WriteString("\n### Exclusion floor\n\n")
	b.WriteString("| Detail | Signal | Rule | Positions |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	for _, e := range Exclusions {
		positions := "every position"
		if len(e.Positions) > 0 {
			positions = sortedPositions(e.Positions)
		}
		fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", e.Detail, e.Signal, e.Rule, positions)
	}
	return b.String()
}

// codeList renders a match list as backticked tokens.
// matchList renders a row's Matches column. An empty Match means "every file"
// only when the row selects by nothing else; on a row that selects by suffix it
// means the Match form contributes nothing, and rendering that as "every file"
// would state the opposite of what the row admits. The rendering IS the
// contract the assembler version digests, so the two cases are distinguished
// here rather than left to a reader (spc-68).
func matchList(row Row) string {
	if len(row.Match) == 0 && len(row.MatchSuffix) == 0 {
		return "every file"
	}
	if len(row.Match) == 0 {
		return "none"
	}
	return codeList(row.Match)
}

// suffixList renders a row's Suffixes column.
func suffixList(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, "`"+it+"`")
	}
	return strings.Join(out, ", ")
}

func codeList(items []string) string {
	if len(items) == 0 {
		return "every file"
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, "`"+it+"`")
	}
	return strings.Join(out, ", ")
}

// fieldList renders a projection's field list.
func fieldList(items []string) string {
	if len(items) == 0 {
		return "the whole file"
	}
	return codeList(items)
}

// sortedPositions renders a row's positions in canonical order.
func sortedPositions(ps []Position) string {
	order := map[Position]int{}
	for i, p := range Positions() {
		order[p] = i
	}
	cp := append([]Position{}, ps...)
	sort.Slice(cp, func(i, j int) bool { return order[cp[i]] < order[cp[j]] })
	names := make([]string, 0, len(cp))
	for _, p := range cp {
		names = append(names, string(p))
	}
	return strings.Join(names, ", ")
}
