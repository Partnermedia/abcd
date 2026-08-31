package reading

// ingest_regime_test.go is the supply-regime gate: itd-185's ac-4 through ac-9
// and ac-12.
//
// Every refusal case here carries the adjacent LEGAL payload in the same test.
// ac-7 is that assertion promoted to a criterion of its own, and the same shape
// is owed to the rest: a refusal test that would pass against a verb refusing
// everything establishes nothing about the rule it names.

import (
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/issueschema"
)

// TestRegimeComesFromTheDefinitionNotThePayload is ac-12 from the other side: the
// verb follows the DEFINITION FILE, not the position table. Rewriting the file's
// regime makes the payload that agreed with the table refuse, and the payload
// that agrees with the file pass. A verb that answered from the table could not
// tell the two apart.
func TestRegimeComesFromTheDefinitionNotThePayload(t *testing.T) {
	f := newIngestFixture(t, "detection")
	f.mustIngest(f.payload(1))

	// The file now states a regime the position table does not.
	f.writeDefinition("detection", RegimeEvaluative)
	def, err := LoadDefinition(f.root, "detection")
	if err != nil {
		t.Fatalf("reload the rewritten definition: %v", err)
	}
	if def.Regime != RegimeEvaluative {
		t.Fatalf("the locator read regime %q from a file stating %q", def.Regime, RegimeEvaluative)
	}

	f.parkRun("rdg-2608310000000002", "detection", AssemblerVersion)
	doc := f.payload(1)
	doc["run_id"] = "rdg-2608310000000002"
	doc["manifest_sha256"] = f.manifestHashOf("rdg-2608310000000002")
	doc["instrument"].(map[string]any)["definition_sha256"] = def.SHA256
	doc["regime"] = issueschema.ReadingRegime("detection")

	res, err := f.ingest(doc)
	if err == nil {
		t.Fatal("the payload agreeing with the POSITION TABLE was accepted against a definition " +
			"stating a different regime; the definition is the source of truth")
	}
	if !strings.Contains(err.Error(), RegimeEvaluative) {
		t.Errorf("the refusal does not name the regime the definition states: %v", err)
	}
	// The refusal must be THIS verb's regime gate, not a downstream writer
	// noticing the same disagreement: a list-level refusal writes a refusal
	// record and leaves the ledger untouched, and a record writer's complaint
	// arrives after the stage has already been taken.
	if res.RefusalPath == "" {
		t.Error("the run was refused without a refusal record, so the refusal came from somewhere " +
			"other than the regime gate")
	}
	if !strings.Contains(err.Error(), DefinitionPath("detection")) {
		t.Errorf("the refusal does not cite the definition it read the regime from: %v", err)
	}
	f.nothingDurableInTheLedger("rdg-2608310000000002")
}

// TestSelfDeclaredRegimeMismatchRefusesRun is ac-12: a self-declared regime that
// disagrees with the definition refuses the RUN, at list level, not one item.
func TestSelfDeclaredRegimeMismatchRefusesRun(t *testing.T) {
	f := newIngestFixture(t, "detection")
	doc := f.payload(3)
	doc["regime"] = RegimeGenerative

	res, err := f.ingest(doc)
	if err == nil {
		t.Fatal("an output declaring a regime the definition does not state was accepted")
	}
	for _, want := range []string{RegimeGenerative, RegimeRegistrative} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q: %v", want, err)
		}
	}
	if len(res.RefusedItems) != 0 {
		t.Errorf("a list-level refusal reported %d item refusals; the run is wrong in whole",
			len(res.RefusedItems))
	}
	if got := f.ledgerRecords(f.runID); len(got) != 0 {
		t.Errorf("the refused run wrote %v", got)
	}
	f.mustIngest(f.payload(3))
}

// TestEvaluativeRankScoreRecommendedRefused is ac-6: each reserved name on the
// evaluative table refuses, and the refusal names the field.
func TestEvaluativeRankScoreRecommendedRefused(t *testing.T) {
	for _, field := range ReservedNames[RegimeEvaluative] {
		field := field
		t.Run(field, func(t *testing.T) {
			f := newIngestFixture(t, "comparative")
			doc := f.payload(2)
			doc["items"].([]any)[1].(map[string]any)[field] = "1"

			r := f.refusedItem(doc, 2, 2)
			if !strings.Contains(r.Field, field) {
				t.Errorf("the refusal names field %q, want %q", r.Field, field)
			}
			if !strings.Contains(r.Detail, field) {
				t.Errorf("the refusal does not name the field: %q", r.Detail)
			}
			if r.Rule != "reserved-name" {
				t.Errorf("the refusal cites rule %q, not the reserved-name table", r.Rule)
			}
		})
	}
}

// TestEvaluativeDocumentOrderIsNeverRefused is ac-7, and it is the reason the
// reserved table names a FIELD rather than a property of the list: items arrive
// in document order by mandate, and arrangement order is never inspected.
func TestEvaluativeDocumentOrderIsNeverRefused(t *testing.T) {
	f := newIngestFixture(t, "comparative")
	doc := f.payload(3)
	// A deliberate arrangement: the three items are ordered, and none carries a
	// reserved field. Nothing about the ORDER may be refused.
	for i, raw := range doc["items"].([]any) {
		raw.(map[string]any)["candidate_id"] = []string{"the first candidate", "the second candidate",
			"the third candidate"}[i]
	}
	res := f.mustIngest(doc)
	if len(res.Records) != 3 {
		t.Fatalf("an arranged evaluative output landed %d of 3 items", len(res.Records))
	}
	if len(res.RefusedItems) != 0 {
		t.Errorf("arrangement order was refused: %v", res.RefusedItems)
	}
}

// TestRegistrativeResolutionFieldRefused is ac-4: a registrative reserved name
// refuses, and the refusal names the item's ordinal, the field, and the licence.
func TestRegistrativeResolutionFieldRefused(t *testing.T) {
	for _, field := range ReservedNames[RegimeRegistrative] {
		field := field
		t.Run(field, func(t *testing.T) {
			f := newIngestFixture(t, "detection")
			doc := f.payload(3)
			// The SECOND of three, so the ordinal in the message is not the only
			// one a verb could have printed.
			doc["items"].([]any)[1].(map[string]any)[field] = "rewrite the constraint"

			r := f.refusedItem(doc, 2, 3)
			if r.Ordinal != 2 {
				t.Errorf("the refusal names ordinal %d, want 2", r.Ordinal)
			}
			if !strings.Contains(r.Field, field) {
				t.Errorf("the refusal names field %q, want %q", r.Field, field)
			}
			if !strings.Contains(r.Detail, "item 2") {
				t.Errorf("the refusal does not name the item's ordinal: %q", r.Detail)
			}
			if !strings.Contains(r.Detail, "not its licence") {
				t.Errorf("the refusal does not state the licence breached: %q", r.Detail)
			}
		})
	}
}

// TestRegistrativeProseFixProposalRefused is ac-5: a body matching a registered
// fix-proposal signature refuses, and the refusal names the item and the
// signature id. The residue is disclosed on itd-185 and is not tested here,
// because a residue is what a test cannot assert.
func TestRegistrativeProseFixProposalRefused(t *testing.T) {
	f := newIngestFixture(t, "detection")
	doc := f.payload(2)
	doc["items"].([]any)[1].(map[string]any)["why_a_tension"] =
		"the record says one thing and the tree another; the fix is to restate the constraint"

	r := f.refusedItem(doc, 2, 2)
	if r.Rule != "RG-REG-FIXPROPOSAL" {
		t.Errorf("the refusal cites rule %q, want the registered signature id", r.Rule)
	}
	if !strings.Contains(r.Detail, "item 2") {
		t.Errorf("the refusal does not name the item: %q", r.Detail)
	}
	if !strings.Contains(r.Detail, "RG-REG-FIXPROPOSAL") {
		t.Errorf("the refusal does not name the signature id: %q", r.Detail)
	}
}

// TestExplicativeDispositionRefused is the structural half of ac-8: a
// disposition-bearing FIELD on a surfaced claim refuses, and the refusal names
// the field.
func TestExplicativeDispositionRefused(t *testing.T) {
	for _, field := range ReservedNames[RegimeExplicative] {
		field := field
		t.Run(field, func(t *testing.T) {
			f := newIngestFixture(t, "entailment")
			doc := f.payload(2)
			doc["items"].([]any)[1].(map[string]any)[field] = "accepted"

			r := f.refusedItem(doc, 2, 2)
			if !strings.Contains(r.Field, field) {
				t.Errorf("the refusal names field %q, want %q", r.Field, field)
			}
			if r.Rule != "reserved-name" {
				t.Errorf("the refusal cites rule %q, not the reserved-name table", r.Rule)
			}
		})
	}
}

// TestExplicativeProseDispositionRefused is ac-9: a claim body matching a
// registered disposition signature refuses, naming the item and the signature.
func TestExplicativeProseDispositionRefused(t *testing.T) {
	f := newIngestFixture(t, "entailment")
	doc := f.payload(2)
	doc["items"].([]any)[0].(map[string]any)["what_implies_it"] =
		"the passed material implies it, and the claim is already accepted by the record"

	r := f.refusedItem(doc, 1, 2)
	if r.Rule != "RG-EXPL-DISPOSITION" {
		t.Errorf("the refusal cites rule %q, want the registered signature id", r.Rule)
	}
	if !strings.Contains(r.Detail, "item 1") {
		t.Errorf("the refusal does not name the item: %q", r.Detail)
	}
	if !strings.Contains(r.Detail, "RG-EXPL-DISPOSITION") {
		t.Errorf("the refusal does not name the signature id: %q", r.Detail)
	}
}

// TestGenerativeHasNoRegimeRefusalButFlagsRecommendation is the generative
// path, which carries no criterion of its own and is recorded on spc-63 rather
// than left to be discovered at the audit. The generative licence is the widest
// and the constraint on it falls at admission, so a signature hit raises a
// review flag on the run record and the item lands.
func TestGenerativeHasNoRegimeRefusalButFlagsRecommendation(t *testing.T) {
	f := newIngestFixture(t, PositionWidening)
	doc := f.payload(1)
	doc["items"].([]any)[0].(map[string]any)["what_admits_it"] =
		"the criterion the record states; we recommend this configuration"

	res := f.mustIngest(doc)
	if len(res.Records) != 1 {
		t.Fatalf("the generative item did not land: %d record(s)", len(res.Records))
	}
	if len(res.ReviewFlags) != 1 {
		t.Fatalf("the generative run raised %d review flag(s), want 1: %v", len(res.ReviewFlags), res.ReviewFlags)
	}
	if res.ReviewFlags[0].SignatureID != "RG-EVAL-RECOMMENDATION" {
		t.Errorf("the flag names signature %q", res.ReviewFlags[0].SignatureID)
	}

	// And the flag is durable: it is on the run record, not only in the render.
	run := f.readRunRecord(f.runID)
	if len(run.ReviewFlags) != 1 || run.ReviewFlags[0].SignatureID != "RG-EVAL-RECOMMENDATION" {
		t.Errorf("the run record carries review flags %v", run.ReviewFlags)
	}
}

// TestEverySignatureShipsEnforced is a property over the registry, not a case:
// every entry ships in enforce mode, every entry names a regime the vocabulary
// holds, and every entry states the licence it protects. Degrading one is an
// edit that turns this red, which is what makes the weakening from enforced to
// observed a recorded act rather than a runtime toggle.
func TestEverySignatureShipsEnforced(t *testing.T) {
	if len(Signatures) < 4 {
		t.Fatalf("the registry holds %d signature(s); spc-63 names four", len(Signatures))
	}
	seen := map[string]bool{}
	for _, s := range Signatures {
		if s.Mode != SignatureEnforce {
			t.Errorf("%s ships in %q mode; every signature ships enforced, and a degradation is a code "+
				"change plus a decision-log entry", s.ID, s.Mode)
		}
		if seen[s.ID] {
			t.Errorf("%s is registered twice", s.ID)
		}
		seen[s.ID] = true
		if s.Pattern == nil {
			t.Errorf("%s registers no detector", s.ID)
		}
		if strings.TrimSpace(s.Licence) == "" {
			t.Errorf("%s states no licence, so a refusal citing it could not say what was breached", s.ID)
		}
		if _, ok := regimeLicence[s.Regime]; !ok {
			t.Errorf("%s polices regime %q, which is not one of the four", s.ID, s.Regime)
		}
	}
	for _, want := range []string{"RG-EVAL-ORDERING", "RG-EVAL-RECOMMENDATION",
		"RG-REG-FIXPROPOSAL", "RG-EXPL-DISPOSITION"} {
		if !seen[want] {
			t.Errorf("the registry does not carry %s, which spc-63 names", want)
		}
	}
}

// TestReservedNamesNeverCollideWithABodyField holds the reserved tables honest:
// a reserved name that was also a position's body field would refuse every legal
// output at that position, and the reserved-name refusal would be unreachable.
func TestReservedNamesNeverCollideWithABodyField(t *testing.T) {
	for regime, names := range ReservedNames {
		for _, p := range Positions() {
			if issueschema.ReadingRegime(string(p)) != regime {
				continue
			}
			for _, field := range issueschema.ReadingBodyFields[string(p)] {
				for _, name := range names {
					if field == name {
						t.Errorf("%q is both a reserved name at the %s regime and a body field of the "+
							"%s position", name, regime, p)
					}
				}
			}
		}
	}
}
