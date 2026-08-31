//go:build smoke || coldreading

package evals

// The oracle: the exclusion table this eval judges by, and the three assertions
// it applies to what the assembler wrote.
//
// The table below is TRANSCRIBED BY HAND from the record — itd-183's exclusion
// list, brief invariants 14 and 15, and adr-55 — and every row carries the
// source it was transcribed from. Transcription is the point. A table generated
// from the assembler's own include list would agree with the assembler by
// construction and could only ever confirm it, never falsify it; that is why
// nothing in this package imports the assembler's package, and why
// TestOracleImportsNothingFromTheAssembler checks rather than promises it.
//
// The disclosed residue: a hand-transcribed table can fall behind the exclusion
// list it mirrors. That staleness is bounded by the holed variant and by the
// anti-vacuity guard, never by the import check. Prose-borne warmth inside an
// included chapter carries no structural signal and stays the residue itd-183
// records.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// assembleVerb is the verb this eval binds to. The eval binds to the VERB, not
// to the package, for the independence reason above; the name lives here, the
// single place a rename touches.
var assembleVerb = []string{"reading", "assemble"}

// The two artefacts an assembly writes into the operator-named directory.
const (
	bundleFile   = "bundle.json"
	manifestFile = "manifest.json"
)

// excludedKey is one frontmatter key the record refuses, with the source it is
// transcribed from.
type excludedKey struct {
	Key    string
	Source string
}

// excludedKeys is the key half of the exclusion floor.
var excludedKeys = []excludedKey{
	{Key: "origin", Source: "itd-183 exclusion list: `origin`, detected by frontmatter key"},
	{Key: "production_mode", Source: "itd-183 exclusion list: production mode, detected by frontmatter key"},
}

// excludedHeading is one heading the record refuses.
type excludedHeading struct {
	Heading string
	Source  string
}

// excludedHeadings is the heading half of the exclusion floor.
var excludedHeadings = []excludedHeading{
	{Heading: "Audit Notes", Source: "itd-183 exclusion list: Audit Notes, detected by heading"},
	{Heading: "Scope Condition Dispositions", Source: "itd-183 exclusion list: scope-condition dispositions"},
	{Heading: "Open Questions", Source: "itd-183: open questions and settled dead ends are deliberation"},
	{
		Heading: "Why This Matters",
		Source: "itd-183 assembler rule 2: a shipped intent travels as press release, acceptance " +
			"criteria, scope conditions, mechanism and spec_id; every other heading stays behind",
	},
}

// excludedFamily is one path no assembled item may come from.
type excludedFamily struct {
	// Path is repo-relative; an item at it or beneath it is a violation.
	Path string
	// Positions are the positions the exclusion binds at. Empty binds at every
	// position.
	Positions []string
	Source    string
}

// excludedFamilies is the path half of the exclusion floor.
var excludedFamilies = []excludedFamily{
	{Path: ".abcd/development/brief/03-evidence", Source: "itd-183: the evidence chapter is deliberation"},
	{Path: ".abcd/development/decisions", Source: "itd-183 exclusion list: decisions/"},
	{Path: ".abcd/development/roadmap/rfcs", Source: "itd-183 exclusion list: roadmap/rfcs/"},
	{Path: ".abcd/development/intents/superseded", Source: "itd-183 exclusion list: intents/superseded/"},
	{Path: ".abcd/development/plans", Source: "itd-183 exclusion list: plans/"},
	{Path: ".abcd/development/research/notes", Source: "itd-183 exclusion list: research/notes/"},
	{
		Path: ".abcd/work/issues",
		Source: "itd-183 exclusion list: work/issues/ in every state, reading records and " +
			"dispositions included, admission and selection grounds, and the lapse log",
	},
	{Path: ".abcd/work/DECISIONS.md", Source: "itd-183 assembler rule 1: .abcd/ is excluded but for what the include list names"},
	{
		Path:   ".abcd/development/readings",
		Source: "itd-183: manifests are warm on the next run, so the instrument's own output is never its input",
	},
	{Path: ".abcd/.work.local", Source: "brief invariant 14: the local ledger side, which no reading consumes"},
	{Path: "agents", Source: "itd-183: the reading definitions are the instrument, not its input"},
	{Path: "evals", Source: "itd-183: the evals that guard this assembler are the instrument"},
	{Path: "internal/core/reading", Source: "itd-183: a reading never receives the include table that decides what it sees"},
	{
		Path:      ".abcd/development/intents/drafts",
		Positions: []string{posWidening, posComparative, posDetection},
		Source:    "itd-183's drafts asymmetry: a reading's object excludes what it exists to change",
	},
	{
		Path:      ".abcd/development/intents/planned",
		Positions: []string{posWidening, posComparative, posDetection},
		Source:    "itd-183's drafts asymmetry: a reading's object excludes what it exists to change",
	},
}

// bindsAt reports whether the exclusion binds at position p.
func (e excludedFamily) bindsAt(p string) bool {
	if len(e.Positions) == 0 {
		return true
	}
	for _, q := range e.Positions {
		if q == p {
			return true
		}
	}
	return false
}

// admittedRecordPath is one record path itd-183 names individually as included.
//
// This is the positive side of assembler rule 1, transcribed: no include may
// name a directory containing a record family, `.abcd/` is excluded wholesale,
// and the record paths a reading legitimately needs are named INDIVIDUALLY. So
// an assembled item under `.abcd/` that lies beneath none of these is a
// violation whether or not the exclusion table above happens to name it — which
// is what makes a record family invented after this table was written excluded
// by default rather than included by oversight.
type admittedRecordPath struct {
	Path      string
	Positions []string
	Source    string
}

// admittedRecordPaths is itd-183's include list, record paths only.
var admittedRecordPaths = []admittedRecordPath{
	{Path: ".abcd/development/brief/01-product", Source: "itd-183 include list; adr-55: the construal as it presently stands is committed record"},
	{Path: ".abcd/development/brief/02-constraints", Source: "itd-183 include list"},
	{Path: ".abcd/development/brief/glossary", Source: "itd-183 include list; adr-55: the glossary's committed terms"},
	{Path: ".abcd/development/intents/shipped", Source: "itd-183 include list, projected to the claim record"},
	{Path: ".abcd/development/intents/disciplines", Source: "itd-183 include list"},
	{Path: ".abcd/development/specs", Source: "itd-183 include list"},
	{
		Path:      ".abcd/development/intents/drafts",
		Positions: []string{posEntailment},
		Source:    "itd-183: the entailment reading includes the candidate set, because articulation precedes selection",
	},
	{
		Path:      ".abcd/development/intents/planned",
		Positions: []string{posEntailment},
		Source:    "itd-183: the entailment reading includes the candidate set, because articulation precedes selection",
	},
}

// admitsAt reports whether the record path is admitted at position p.
func (a admittedRecordPath) admitsAt(p string) bool {
	if len(a.Positions) == 0 {
		return true
	}
	for _, q := range a.Positions {
		if q == p {
			return true
		}
	}
	return false
}

// violation is one finding: what rule was broken, at which position, and by
// what. Assertions RETURN violations rather than failing the test, so the
// negative control can demand exactly the two it expects.
type violation struct {
	Position string
	Rule     string
	Class    string
	Detail   string
	Source   string
}

// String renders a violation so the failure names the leaked sentinel's class
// token and the position at which it leaked.
func (v violation) String() string {
	head := fmt.Sprintf("[%s] %s: %s", v.Position, v.Rule, v.Detail)
	if v.Class != "" {
		head = fmt.Sprintf("[%s] %s: %s leaked (%s)", v.Position, v.Rule, sentinelPrefix+v.Class, v.Detail)
	}
	if v.Source != "" {
		head += " — " + v.Source
	}
	return head
}

// The three rules, named once so a test can demand a rule by name.
const (
	ruleSentinel       = "sentinel absence"
	ruleExcludedKey    = "field absence (excluded key)"
	ruleExcludedHeader = "field absence (excluded heading)"
	ruleFamily         = "family absence"
	ruleUnnamedRecord  = "family absence (record path no include names)"
)

// bundleItem and manifestItem are the two artefacts' item shapes, read as this
// eval needs them rather than as the assembler defines them.
type bundleItem struct {
	ItemKey string `json:"item_key"`
	Kind    string `json:"kind"`
	Text    string `json:"text"`
}

type manifestItem struct {
	ItemKey string `json:"item_key"`
	Path    string `json:"path"`
	Field   string `json:"field"`
	SHA256  string `json:"sha256"`
}

// assembled is one run's output as this eval reads it.
type assembled struct {
	Position      string
	BundleRaw     []byte
	ManifestRaw   []byte
	Items         []bundleItem
	ManifestItems []manifestItem
}

// assemble runs the assembler over a materialised fixture at one position and
// returns what it wrote.
//
// The invocation is out of process, through the built binary, which is the
// first of the four mechanisms that hold the oracle independent: this package
// cannot read the include table even by accident, because it never links it.
// The output directory is outside the fixture repository, so a run leaves the
// fixture's own tiers untouched.
func assemble(t *testing.T, f fixture, position string) assembled {
	t.Helper()
	outDir := filepath.Join(t.TempDir(), "run-"+position)
	args := append(append([]string{}, assembleVerb...),
		"--position", position, "--target", "HEAD", "--out", outDir, "--dry-run")
	out, code := runIn(t, f.Root, []string{"HOME=" + f.Home}, args...)
	if code != 0 {
		t.Fatalf("`abcd %s` over the %s fixture exited %d\n%s",
			strings.Join(args, " "), f.Variant, code, out)
	}
	a := assembled{Position: position}
	a.BundleRaw = readArtefact(t, filepath.Join(outDir, bundleFile))
	a.ManifestRaw = readArtefact(t, filepath.Join(outDir, manifestFile))

	var bundle struct {
		Items []bundleItem `json:"items"`
	}
	if err := json.Unmarshal(a.BundleRaw, &bundle); err != nil {
		t.Fatalf("decoding the assembled input at %s: %v", position, err)
	}
	var manifest struct {
		Items []manifestItem `json:"items"`
	}
	if err := json.Unmarshal(a.ManifestRaw, &manifest); err != nil {
		t.Fatalf("decoding the manifest at %s: %v", position, err)
	}
	a.Items, a.ManifestItems = bundle.Items, manifest.Items
	requireCarriers(t, a)
	return a
}

// requireCarriers is the floor under every absence assertion: the assembly must
// be non-empty, and every plant-bearing file the include list names must be in
// it at this position.
//
// The weaker floor — any item at all — cannot see the failure that matters. An
// assembler that stopped enumerating one whole class of source would drop the
// files carrying WARM-FIELD, two of the three WARM-KEY plants and both draft
// plants, leave nine walk-row items behind, and turn every assertion in this
// package green because there was nothing left to leak. That is the corpus
// argument one step further on: an absence assertion cannot see a corpus that
// lost its plants, and it cannot see an ASSEMBLY that lost their carriers
// either.
func requireCarriers(t *testing.T, a assembled) {
	t.Helper()
	if len(a.Items) == 0 {
		t.Fatalf("the assembly at %s passed no items; an absence assertion over an empty "+
			"bundle asserts nothing", a.Position)
	}
	seen := make(map[string]bool, len(a.ManifestItems))
	for _, it := range a.ManifestItems {
		seen[it.Path] = true
	}
	for _, c := range carriers {
		if !c.reachesAt(a.Position) {
			continue
		}
		// The path first, because a missing path and an empty text are different
		// faults and the message should say which.
		if !seen[c.Path] {
			t.Fatalf("the assembly at %s does not carry %s (%s); %s",
				a.Position, c.Path, c.Why, c.stake())
		}
		// Then the bytes. The manifest names what an assembly SAYS it passed; only
		// the bundle says what it actually passed, and an absence assertion over an
		// item whose text is empty is an absence assertion over nothing.
		for _, marker := range c.Markers {
			if bytes.Contains(a.BundleRaw, []byte(marker)) {
				continue
			}
			t.Fatalf("the assembly at %s names %s in its manifest, but the bundle does not carry "+
				"that file's own text (%q is absent); %s", a.Position, c.Path, marker, c.stake())
		}
	}
}

// requireOracleTables refuses a table that has changed size behind the
// assertions that consume it.
//
// A greater-than-zero floor on a table whose size is known is not a floor: a
// two-row table can halve under it, and a table emptied one row at a time never
// trips it at all. Every one of these tables is a declaration, so its size is a
// declaration too, and each of them is load-bearing — emptying carriers restores
// the vacuity the carriers floor exists to close, emptying excludedKeys and
// excludedHeadings disarms the field-absence check entirely, and emptying
// excludedFamilies disarms half the family-absence check.
//
// The count is deliberately duplicated here rather than derived from len(): a
// derived count agrees with the table by construction and could only confirm it,
// which is the same mistake as an oracle that reads the assembler's own table.
func requireOracleTables(t *testing.T) {
	t.Helper()
	for _, tbl := range []struct {
		name string
		got  int
		want int
	}{
		{"sentinelClasses", len(sentinelClasses), 14},
		{"carriers", len(carriers), 10},
		{"holes", len(holes), 2},
		{"excludedKeys", len(excludedKeys), 2},
		{"excludedHeadings", len(excludedHeadings), 4},
		{"excludedFamilies", len(excludedFamilies), 15},
		{"admittedRecordPaths", len(admittedRecordPaths), 8},
		{"coverage", len(coverage), 46},
	} {
		if tbl.got != tbl.want {
			t.Fatalf("the %s table holds %d row(s), and this eval is written against %d; "+
				"a table that changes size behind the assertions consuming it is how an "+
				"absence eval goes quietly vacuous, so update the declared count deliberately",
				tbl.name, tbl.got, tbl.want)
		}
	}
}

// readArtefact reads one of the two artefacts, failing if it is absent — the
// assembler is required to write both as separate files.
func readArtefact(t *testing.T, p string) []byte {
	t.Helper()
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("the assembly wrote no %s: %v", filepath.Base(p), err)
	}
	return data
}

// checkSentinelAbsence is assertion 1: no planted token appears in the RAW
// serialisation of the assembled input.
//
// It is a content oracle, not a path oracle. Nothing in it mentions where a
// plant was, so a plant that moves is still caught — which is exactly the
// failure a path assertion misses.
func checkSentinelAbsence(a assembled) []violation {
	var out []violation
	raw := string(a.BundleRaw)
	for _, c := range sentinelClasses {
		if c.coldAt(a.Position) {
			continue
		}
		if strings.Contains(raw, c.Token()) {
			out = append(out, violation{
				Position: a.Position,
				Rule:     ruleSentinel,
				Class:    c.Name,
				Detail:   "found in the assembled input's raw serialisation",
				Source:   c.Why,
			})
		}
	}
	return out
}

// excludedKeyLine matches a frontmatter key at ANY indentation, quoted or bare.
// The indentation allowance is what spc-64 means by depth-agnostic: a warm field
// nested one level deeper is the same warm field, and a pattern anchored at
// column 0 would call it clean.
//
// The allowance is UNFALSIFIABLE through this eval and is recorded as such in
// the coverage matrix: the assembler fails closed on an indented excluded key
// that survives redaction, so a corpus carrying one makes the assembly exit
// non-zero rather than leak, and the branch is never reached. It is kept because
// spc-64 asks for it and because it costs nothing to be right if that refusal
// ever softens — not because it has been watched catch anything.
var excludedKeyLine = regexp.MustCompile(`^\s*["']?([A-Za-z_][A-Za-z0-9_-]*)["']?\s*:`)

// atxHeading matches an ATX heading, including the one-to-three-space indent
// CommonMark allows.
var atxHeading = regexp.MustCompile(`^[ ]{0,3}#{1,6}\s+(.*?)\s*#*\s*$`)

// checkFieldAbsence is assertion 2: a recursive walk of the parsed documents
// reports no key on the excluded-key list at any depth, and no excluded heading
// in any projected body.
//
// Two walks, because a key can be a key in two different documents. The first
// is over the artefacts' own JSON, at any nesting depth. The second is over each
// item's TEXT, which is where a record's frontmatter actually travels, and it
// too matches at any depth.
func checkFieldAbsence(a assembled) []violation {
	var out []violation
	keys := map[string]excludedKey{}
	for _, k := range excludedKeys {
		keys[strings.ToLower(k.Key)] = k
	}

	for _, artefact := range []struct {
		name string
		raw  []byte
	}{{"the assembled input", a.BundleRaw}, {"the manifest", a.ManifestRaw}} {
		var doc any
		if err := json.Unmarshal(artefact.raw, &doc); err != nil {
			continue
		}
		for _, k := range walkJSONKeys(doc) {
			if hit, ok := keys[strings.ToLower(k)]; ok {
				out = append(out, violation{
					Position: a.Position,
					Rule:     ruleExcludedKey,
					Detail:   fmt.Sprintf("%s carries the key %q", artefact.name, hit.Key),
					Source:   hit.Source,
				})
			}
		}
	}

	// A manifest item's `field` NAMES the projected field, so an excluded key
	// that got itself projected shows up there as a value rather than a key.
	for _, it := range a.ManifestItems {
		if hit, ok := keys[strings.ToLower(it.Field)]; ok {
			out = append(out, violation{
				Position: a.Position,
				Rule:     ruleExcludedKey,
				Detail:   fmt.Sprintf("the manifest projects the field %q as item %s", hit.Key, it.ItemKey),
				Source:   hit.Source,
			})
		}
	}

	for _, it := range a.Items {
		for _, line := range strings.Split(it.Text, "\n") {
			if m := excludedKeyLine.FindStringSubmatch(line); m != nil {
				if hit, ok := keys[strings.ToLower(m[1])]; ok {
					out = append(out, violation{
						Position: a.Position,
						Rule:     ruleExcludedKey,
						Detail:   fmt.Sprintf("item %s carries the key %q", it.ItemKey, hit.Key),
						Source:   hit.Source,
					})
				}
			}
			m := atxHeading.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			for _, h := range excludedHeadings {
				if strings.EqualFold(strings.TrimSpace(m[1]), h.Heading) {
					out = append(out, violation{
						Position: a.Position,
						Rule:     ruleExcludedHeader,
						Detail:   fmt.Sprintf("item %s carries the heading %q", it.ItemKey, h.Heading),
						Source:   h.Source,
					})
				}
			}
		}
	}
	return out
}

// walkJSONKeys returns every object key in a decoded document, at any depth.
func walkJSONKeys(doc any) []string {
	var out []string
	var walk func(any)
	walk = func(v any) {
		switch t := v.(type) {
		case map[string]any:
			for k, child := range t {
				out = append(out, k)
				walk(child)
			}
		case []any:
			for _, child := range t {
				walk(child)
			}
		}
	}
	walk(doc)
	return out
}

// checkFamilyAbsence is assertion 3: no item's recorded path lies inside an
// excluded record family, and no item under `.abcd/` lies outside the record
// paths the include list names individually.
//
// The second half is assembler rule 1 as a positive test, which is what catches
// a record family invented after this table was written.
func checkFamilyAbsence(a assembled) []violation {
	var out []violation
	for _, it := range a.ManifestItems {
		rel := path.Clean(it.Path)
		for _, fam := range excludedFamilies {
			if !fam.bindsAt(a.Position) {
				continue
			}
			if rel == fam.Path || strings.HasPrefix(rel, fam.Path+"/") {
				out = append(out, violation{
					Position: a.Position,
					Rule:     ruleFamily,
					Detail:   fmt.Sprintf("item %s comes from %s, inside %s", it.ItemKey, rel, fam.Path),
					Source:   fam.Source,
				})
			}
		}
		if !strings.HasPrefix(rel, ".abcd/") {
			continue
		}
		named := false
		for _, adm := range admittedRecordPaths {
			if !adm.admitsAt(a.Position) {
				continue
			}
			if rel == adm.Path || strings.HasPrefix(rel, adm.Path+"/") {
				named = true
				break
			}
		}
		if !named {
			out = append(out, violation{
				Position: a.Position,
				Rule:     ruleUnnamedRecord,
				Detail: fmt.Sprintf("item %s comes from %s, a record path no include names at this position",
					it.ItemKey, rel),
				Source: "itd-183 assembler rule 1: .abcd/ is excluded wholesale and the record paths a " +
					"reading needs are named individually, so a family added later is excluded by construction",
			})
		}
	}
	return out
}

// checkReadBlock is the three assertions together.
func checkReadBlock(a assembled) []violation {
	var out []violation
	out = append(out, checkSentinelAbsence(a)...)
	out = append(out, checkFieldAbsence(a)...)
	out = append(out, checkFamilyAbsence(a)...)
	return out
}

// reportViolations renders a violation list for a failure message.
func reportViolations(vs []violation) string {
	lines := make([]string, 0, len(vs))
	for _, v := range vs {
		lines = append(lines, "  "+v.String())
	}
	return strings.Join(lines, "\n")
}
