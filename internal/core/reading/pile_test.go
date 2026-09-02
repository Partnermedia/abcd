package reading

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
)

// The baselines below are the item sets the assembler produced over the
// fixture repository BEFORE the per-position pile existed, recorded from the
// unmodified tree at origin/main. They are what "the default stays one shared
// assembly" means as a number: a change that moves one of them has moved what a
// position sees, whatever it says about itself.
//
// The digest is over the manifest's items — key, path, field, kind, bytes and
// hash — rather than over the manifest document, because the document also
// carries the run identifier, which differs between runs by construction, and
// the pile stamp, which is the field this change adds. The items are the pile.
var baselineItemDigest = map[Position]string{
	PositionWidening:   "703ef4f49ac8bc43030438f8f1f40f12b87976e9cb8d9b5806bf4faa66dc87da",
	PositionEntailment: "9fa2e66e3fecea910238e70fede5bbf333431efa2816d11f92f8ef32672d45ba",
	PositionDetection:  "703ef4f49ac8bc43030438f8f1f40f12b87976e9cb8d9b5806bf4faa66dc87da",
}

// itemDigest hashes a manifest's item set, which is the assembly's pile as an
// auditor can check it.
func itemDigest(m Manifest) string {
	var b strings.Builder
	for _, it := range m.Items {
		fmt.Fprintf(&b, "%s|%s|%s|%s|%d|%s\n",
			it.ItemKey, it.Path, it.Field, it.Kind, it.Bytes, it.SHA256)
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

// withOwnPile declares one position's own pile for the duration of a test and
// restores the map afterwards. The map is package state, so the restore is a
// whole-map swap rather than a delete: a test that left a key behind would
// change what every later test in the package assembles.
func withOwnPile(t *testing.T, p Position, pile PositionPile) {
	t.Helper()
	restore := PositionTables
	t.Cleanup(func() { PositionTables = restore })
	next := make(map[Position]PositionPile, len(restore)+1)
	for k, v := range restore {
		next[k] = v
	}
	next[p] = pile
	PositionTables = next
}

// docsOnlyPile is a pile narrow enough that its effect on an assembly is
// unmistakable: the root documentation rows alone, none of the record.
func docsOnlyPile(p Position) PositionPile {
	return PositionPile{
		Rule: "a test's own pile: the delivered documentation alone",
		Rows: []Row{{
			Positions: []Position{p},
			Source:    ".",
			Match:     []string{".md"},
			Kind:      KindDoc,
			Rule:      "a test's own pile: the delivered documentation alone",
		}},
	}
}

// TestNoPerPositionSectionLeavesTheSharedPileUnchanged is the ruling's default
// half: reading positions share one pile, and declaring nothing must assemble
// exactly what the assembler assembled before the mechanism existed.
func TestNoPerPositionSectionLeavesTheSharedPileUnchanged(t *testing.T) {
	if len(PositionTables) != 0 {
		t.Fatalf("PositionTables declares %d own pile(s); the default is one shared assembly",
			len(PositionTables))
	}
	root := fixtureRepo(t)
	sharedHash := PileHashOf(Table)
	for _, p := range AssemblingPositions() {
		rows, src := RowsFor(p)
		if src != PileShared {
			t.Errorf("position %s draws from the %s pile with none declared", p, src)
		}
		if len(rows) != len(Table) {
			t.Errorf("position %s draws %d rows; the shared table has %d", p, len(rows), len(Table))
		}
		res := assembleFixture(t, root, p)
		if got := itemDigest(res.Manifest); got != baselineItemDigest[p] {
			t.Errorf("the %s pile has moved: item digest %s, recorded baseline %s",
				p, got, baselineItemDigest[p])
		}
		if res.Manifest.Pile.Source != PileShared {
			t.Errorf("the %s manifest names the %q pile; nothing is declared, so it is shared",
				p, res.Manifest.Pile.Source)
		}
		if res.Manifest.Pile.Hash != sharedHash {
			t.Errorf("the %s manifest hashes its pile as %s; the shared table hashes to %s",
				p, res.Manifest.Pile.Hash, sharedHash)
		}
		if res.Pile != res.Manifest.Pile {
			t.Errorf("the %s result reports pile %+v and its manifest %+v", p, res.Pile, res.Manifest.Pile)
		}
	}
}

// TestAnOwnPileChangesOnlyItsOwnPosition is the ruling's second half. A pile
// handed to one position must move that position's assembly and no other's:
// the shared pile is what every undeclared position keeps.
func TestAnOwnPileChangesOnlyItsOwnPosition(t *testing.T) {
	root := fixtureRepo(t)
	withOwnPile(t, PositionDetection, docsOnlyPile(PositionDetection))

	own := assembleFixture(t, root, PositionDetection)
	if got := itemDigest(own.Manifest); got == baselineItemDigest[PositionDetection] {
		t.Error("the detection position assembled its shared item set although it was given its own pile")
	}
	if own.Manifest.Pile.Source != PileOwn {
		t.Errorf("the detection manifest names the %q pile; it was given its own", own.Manifest.Pile.Source)
	}
	wantHash := PileHashOf(PositionTables[PositionDetection].Rows)
	if own.Manifest.Pile.Hash != wantHash {
		t.Errorf("the detection manifest hashes its pile as %s, want %s", own.Manifest.Pile.Hash, wantHash)
	}
	if own.Manifest.Pile.Hash == PileHashOf(Table) {
		t.Error("an own pile hashes to the shared table's hash, so the two cannot be told apart")
	}
	for _, it := range own.Manifest.Items {
		if it.Kind != KindDoc {
			t.Errorf("the detection pile passed %s as %s; its rows admit documentation alone",
				it.Path, it.Kind)
		}
	}

	// Every other position keeps the shared pile, byte for byte.
	for _, p := range AssemblingPositions() {
		if p == PositionDetection {
			continue
		}
		res := assembleFixture(t, root, p)
		if got := itemDigest(res.Manifest); got != baselineItemDigest[p] {
			t.Errorf("giving detection its own pile moved the %s pile: %s, want %s",
				p, got, baselineItemDigest[p])
		}
		if res.Manifest.Pile.Source != PileShared {
			t.Errorf("the %s manifest names the %q pile after another position was given its own",
				p, res.Manifest.Pile.Source)
		}
	}
}

// TestAnOwnPileAssemblesTwiceToTheSameItems holds amnesia for a per-position
// pile: the property itd-187's eval holds for the shared assembly must hold for
// an own one, and the eval drives the built binary, which carries no declared
// pile to exercise.
func TestAnOwnPileAssemblesTwiceToTheSameItems(t *testing.T) {
	root := fixtureRepo(t)
	withOwnPile(t, PositionWidening, docsOnlyPile(PositionWidening))
	first := assembleFixture(t, root, PositionWidening)
	second := assembleFixture(t, root, PositionWidening)
	if first.RunID == second.RunID {
		t.Fatal("both assemblies carry one run identifier, so the comparison read one artefact twice")
	}
	if itemDigest(first.Manifest) != itemDigest(second.Manifest) {
		t.Error("two assemblies of one state at one commit produced different item sets from one own pile")
	}
	a, err := EncodeBundle(first.Bundle)
	if err != nil {
		t.Fatalf("encode the first bundle: %v", err)
	}
	b, err := EncodeBundle(second.Bundle)
	if err != nil {
		t.Fatalf("encode the second bundle: %v", err)
	}
	if string(a) != string(b) {
		t.Error("two assemblies from one own pile produced different assembled input")
	}
}

// TestAnOwnPileIsHeldToTheExclusionFloor is the read-block property for a
// per-position pile: the floor and the structural deny bind the same at every
// position, so a pile cannot be the way warm material walks in.
func TestAnOwnPileIsHeldToTheExclusionFloor(t *testing.T) {
	root := fixtureRepo(t)
	withOwnPile(t, PositionWidening, docsOnlyPile(PositionWidening))
	res := assembleFixture(t, root, PositionWidening)
	text := bundleText(res.Bundle)
	for _, sentinel := range []string{
		sentinelEvidence, sentinelDecision, sentinelIssue, sentinelAuditNotes,
		sentinelSuperseded, sentinelPlan, sentinelPriorRun, sentinelDraftBody,
		sentinelDefinition, sentinelLapse,
	} {
		if strings.Contains(text, sentinel) {
			t.Errorf("an own pile passed %s; the floor binds a pile exactly as it binds the shared table", sentinel)
		}
	}
	if len(res.Manifest.Exclusions) != len(ExclusionsFor(PositionWidening)) {
		t.Errorf("an own pile's manifest asserts %d exclusion(s), the position's floor has %d",
			len(res.Manifest.Exclusions), len(ExclusionsFor(PositionWidening)))
	}
}

// TestTheValidatorRefusesAnOwnPileTheShapesItRefusesInTheSharedTable is
// itd-194's validation applied to the per-position lists through the same
// function rather than through a copy of it. Each case is fed twice: once as a
// whole table and once as one position's own pile.
func TestTheValidatorRefusesAnOwnPileTheShapesItRefusesInTheSharedTable(t *testing.T) {
	good := Row{
		Positions: []Position{PositionWidening},
		Source:    ".",
		Match:     []string{".md"},
		Kind:      KindDoc,
		Rule:      "a well-formed row",
	}
	mutate := func(f func(*Row)) []Row {
		r := good
		f(&r)
		return []Row{r}
	}
	cases := []struct {
		name string
		rows []Row
		want string
	}{
		{"an empty pile", nil, "admits nothing"},
		{"an unknown kind", mutate(func(r *Row) { r.Kind = "hearsay" }), "closed vocabulary"},
		{"no position", mutate(func(r *Row) { r.Positions = nil }), "admitted at no position"},
		{"an unknown position", mutate(func(r *Row) {
			r.Positions = []Position{PositionWidening, "framing"}
		}), "unknown reading position"},
		{"no admitting rule", mutate(func(r *Row) { r.Rule = "" }), "states no admitting rule"},
		{"no positive match", mutate(func(r *Row) { r.Match, r.MatchSuffix = nil, nil }), "positive at every grain"},
		{"an unknown record store", mutate(func(r *Row) { r.Store = "rdg" }), "unknown record store"},
		{"a directory containing a record family", mutate(func(r *Row) {
			r.Source = ".abcd/development"
		}), "record family"},
		{"the assembler's own package", mutate(func(r *Row) {
			r.Source = "internal/core/reading"
			r.Match = []string{".go"}
			r.Kind = KindSource
		}), "structurally denied"},
		{"the instrument's own definitions", mutate(func(r *Row) {
			r.Source = "agents"
		}), "excluded directory"},
		{"the material the reading exists to change", mutate(func(r *Row) {
			r.Source = ".abcd/development/intents/drafts"
			r.Store, r.Bucket = "itd", "drafts"
		}), "excluded directory"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRows(PositionWidening, tc.rows)
			if err == nil {
				t.Fatalf("the validator accepted %s as a whole table", tc.name)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("the refusal does not say %q: %v", tc.want, err)
			}
			withOwnPile(t, PositionWidening, PositionPile{Rule: "a test's pile", Rows: tc.rows})
			perr := ValidatePile(PositionWidening)
			if perr == nil {
				t.Fatalf("the validator accepted %s as the widening position's own pile", tc.name)
			}
			if !strings.Contains(perr.Error(), tc.want) {
				t.Errorf("the pile refusal does not say %q: %v", tc.want, perr)
			}
		})
	}
}

// TestTheSharedTableAndEveryDeclaredPileValidate is the validator turned on the
// table this binary actually ships, at every position.
func TestTheSharedTableAndEveryDeclaredPileValidate(t *testing.T) {
	for _, p := range Positions() {
		if err := ValidatePile(p); err != nil {
			t.Errorf("the pile the %s position assembles from does not validate: %v", p, err)
		}
	}
}

// TestAnOwnPileMustAdmitAtItsOwnPosition holds the one rule a pile has that a
// row does not: a pile handed to a position is assembled at that position, so a
// row that does not admit there would be declared and never read.
func TestAnOwnPileMustAdmitAtItsOwnPosition(t *testing.T) {
	withOwnPile(t, PositionDetection, PositionPile{
		Rule: "a test's pile",
		Rows: []Row{{
			Positions: []Position{PositionWidening},
			Source:    ".",
			Match:     []string{".md"},
			Kind:      KindDoc,
			Rule:      "a row admitted somewhere else",
		}},
	})
	err := ValidatePile(PositionDetection)
	if err == nil {
		t.Fatal("a pile whose rows admit at another position was accepted")
	}
	if !strings.Contains(err.Error(), "detection") {
		t.Errorf("the refusal does not name the position the pile was given to: %v", err)
	}
}

// TestAssembleRefusesAnInvalidOwnPile keeps the validation at the door rather
// than at the far end of a walk: an assembly under a pile that does not validate
// is a run whose manifest would describe a table nobody sanctioned.
func TestAssembleRefusesAnInvalidOwnPile(t *testing.T) {
	root := fixtureRepo(t)
	withOwnPile(t, PositionWidening, PositionPile{
		Rule: "a test's pile",
		Rows: []Row{{
			Positions: []Position{PositionWidening},
			Source:    ".abcd/development/intents",
			Match:     []string{".md"},
			Kind:      KindIntentProjection,
			Rule:      "reaching into a family from above",
		}},
	})
	if _, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "HEAD", Scope: fixtureScopeName, DryRun: true,
	}); err == nil {
		t.Fatal("an assembly ran under a pile reaching into a record family from above")
	}
}

// TestAPerPositionSectionMovesTheStampedAssemblerVersion holds the pile inside
// the assembly contract. A manifest names a version, and a version that did not
// move when a position's pile was declared would name a table that does not
// describe what the run was assembled from.
func TestAPerPositionSectionMovesTheStampedAssemblerVersion(t *testing.T) {
	before := AssemblerVersion()
	beforeRender := Render()
	withOwnPile(t, PositionDetection, docsOnlyPile(PositionDetection))
	if after := AssemblerVersion(); after == before {
		t.Error("declaring a position's own pile left the stamped assembler version unmoved")
	}
	if Render() == beforeRender {
		t.Error("declaring a position's own pile left the rendered charter table unchanged")
	}
	if !strings.Contains(Render(), string(PositionDetection)) {
		t.Error("the rendered table does not name the position that was given its own pile")
	}
}

// TestTheManifestRefusesToDecodeWithoutAPileStamp is the fail-closed half of
// the stamp. Which pile a run was assembled from is the fact the closing-run
// comparison turns on, and a shape that can omit it cannot tell a shared
// assembly from a manifest that forgot to say.
func TestTheManifestRefusesToDecodeWithoutAPileStamp(t *testing.T) {
	root := fixtureRepo(t)
	res := assembleFixture(t, root, PositionWidening)
	raw, err := EncodeManifest(res.Manifest)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if !strings.Contains(string(raw), `"pile"`) {
		t.Fatal("the encoded manifest carries no pile stamp")
	}
	round, err := DecodeManifest(raw)
	if err != nil {
		t.Fatalf("decode a manifest this package wrote: %v", err)
	}
	if round.Pile != res.Manifest.Pile {
		t.Errorf("the pile stamp did not survive the round trip: %+v, want %+v", round.Pile, res.Manifest.Pile)
	}
	for _, bad := range []struct {
		name string
		from string
		to   string
	}{
		{"an empty source", `"source": "shared"`, `"source": ""`},
		{"an unknown source", `"source": "shared"`, `"source": "borrowed"`},
		{"an empty hash", `"hash": "` + res.Manifest.Pile.Hash + `"`, `"hash": ""`},
	} {
		t.Run(bad.name, func(t *testing.T) {
			mangled := strings.Replace(string(raw), bad.from, bad.to, 1)
			if mangled == string(raw) {
				t.Fatalf("the manifest does not carry %q, so the case tests nothing", bad.from)
			}
			if _, err := DecodeManifest([]byte(mangled)); err == nil {
				t.Errorf("a manifest with %s decoded clean", bad.name)
			}
		})
	}
}

// TestStatusReportsWhichPositionsHaveTheirOwnPile is the reporting half of the
// ruling: which pile a position draws from is visible without running an
// assembly.
func TestStatusReportsWhichPositionsHaveTheirOwnPile(t *testing.T) {
	shared, err := Describe("")
	if err != nil {
		t.Fatalf("describe: %v", err)
	}
	if len(shared.Piles) != len(Positions()) {
		t.Fatalf("the status reports %d pile(s) for %d position(s)", len(shared.Piles), len(Positions()))
	}
	for _, p := range shared.Piles {
		if p.Pile != PileShared {
			t.Errorf("position %s reports the %q pile with none declared", p.Position, p.Pile)
		}
		if p.Rows != len(Table) || p.Hash != PileHashOf(Table) {
			t.Errorf("position %s reports %d row(s) at %s; the shared table has %d at %s",
				p.Position, p.Rows, p.Hash, len(Table), PileHashOf(Table))
		}
	}

	withOwnPile(t, PositionComparative, docsOnlyPile(PositionComparative))
	own, err := Describe("")
	if err != nil {
		t.Fatalf("describe: %v", err)
	}
	found := false
	for _, p := range own.Piles {
		if p.Position != PositionComparative {
			if p.Pile != PileShared {
				t.Errorf("position %s reports the %q pile although comparative was the one declared",
					p.Position, p.Pile)
			}
			continue
		}
		found = true
		if p.Pile != PileOwn {
			t.Errorf("the comparative position reports the %q pile although it was given its own", p.Pile)
		}
		if p.Rows != 1 {
			t.Errorf("the comparative pile reports %d row(s), want 1", p.Rows)
		}
		if p.Rule == "" {
			t.Error("the comparative pile reports no rule; a pile states why the position is handed its own object")
		}
	}
	if !found {
		t.Error("the status does not report the position that was given its own pile")
	}
}
