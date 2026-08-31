package reading

// ingest_regime_test.go is the supply-regime gate: itd-185's ac-4 through ac-9
// and ac-12.
//
// Every refusal case here carries the adjacent LEGAL payload in the same test.
// ac-7 is that assertion promoted to a criterion of its own, and the same shape
// is owed to the rest: a refusal test that would pass against a verb refusing
// everything establishes nothing about the rule it names.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/issueschema"
)

// TestRegimeComesFromTheDefinitionNotThePayload is ac-12 from the verb's side:
// the regime is READ, from the position's definition file, and a run whose
// definition the locator cannot resolve has no regime at all.
//
// Removing the file is the discriminator. A verb that answered from
// issueschema's position table would carry straight on — the table still binds
// detection to registrative whether or not any file exists — so this case
// separates a verb that reads the definition from one that only says it does.
func TestRegimeComesFromTheDefinitionNotThePayload(t *testing.T) {
	f := newIngestFixture(t, "detection")
	legal := f.payload(1)
	f.mustIngest(legal)

	if err := os.Remove(filepath.Join(f.root, filepath.FromSlash(DefinitionPath("detection")))); err != nil {
		t.Fatal(err)
	}
	f.parkRun("rdg-2608310000000002", "detection", AssemblerVersion)
	doc := f.payload(1)
	doc["run_id"] = "rdg-2608310000000002"
	doc["manifest_sha256"] = f.manifestHashOf("rdg-2608310000000002")

	_, err := f.ingest(doc)
	if err == nil {
		t.Fatal("a run whose definition is absent was accepted; the regime then came from somewhere " +
			"other than the definition")
	}
	if !strings.Contains(err.Error(), DefinitionPath("detection")) {
		t.Errorf("the refusal does not name the definition it could not resolve: %v", err)
	}
	f.nothingDurableInTheLedger("rdg-2608310000000002")
}

// TestADriftedDefinitionRefusesTheRunRatherThanChangingTheLicence is the
// adversarial half of ac-12, which the criterion does not state: ac-12 is the
// PAYLOAD lying about its regime, and this is the DEFINITION lying.
//
// A file stating a legal regime under the wrong position would hand this verb
// the wrong licence with no refusal anywhere in the path — the gate whose whole
// purpose is to catch a reading that exceeded its licence would then be
// enforcing a different one, silently. The refusal lives in the locator, which
// is the one thing that claims to resolve a position to its regime
// (iss-2608311145258479); this case holds the ingest path to it, so the fix
// cannot be undone in the locator without a reading-side test going red.
//
// No refusal record is written, and that is deliberate: a run's durable record
// states the regime it read under, and a run whose definition does not resolve
// has none the verb could honestly write down.
func TestADriftedDefinitionRefusesTheRunRatherThanChangingTheLicence(t *testing.T) {
	f := newIngestFixture(t, "detection")
	f.mustIngest(f.payload(1))

	// The detection definition now states the evaluative licence. Under it, the
	// evaluative reserved names would be enforced against a registrative
	// reading, and the registrative ones would not be enforced at all.
	f.writeDefinition("detection", RegimeEvaluative)

	f.parkRun("rdg-2608310000000007", "detection", AssemblerVersion)
	doc := f.payload(1)
	doc["run_id"] = "rdg-2608310000000007"
	doc["manifest_sha256"] = f.manifestHashOf("rdg-2608310000000007")

	res, err := f.ingest(doc)
	if err == nil {
		t.Fatal("a definition stating another position's regime was resolved rather than refused, so " +
			"the run read under a licence nothing checked")
	}
	for _, want := range []string{RegimeEvaluative, RegimeRegistrative, DefinitionPath("detection")} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q: %v", want, err)
		}
	}
	if res.RefusalPath != "" {
		t.Errorf("a run with no resolvable regime wrote a refusal record at %s, which would have to "+
			"state a regime the verb could not resolve", res.RefusalPath)
	}
	f.nothingDurable("rdg-2608310000000007")
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
	f.mustIngest(f.nextRun(f.payload(3)))
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

// TestTheRegimeGateIsNotEvadedByInvisibleRunes.
//
// Go's regexp is RE2, whose whitespace and word-boundary classes are ASCII-only,
// and the terminal sanitiser does not mask U+00A0. So a registered signature's
// OWN phrasing, with a non-breaking space between two words or a zero-width
// space inside a keyword, matched nothing and the item landed with no refusal
// and no flag — inside the gate this verb exists to be.
//
// This is not the residue itd-185 discloses. That residue covers a fix proposal
// or a disposition phrased OUTSIDE the registry's signatures. This was the
// registry's own phrasing with one invisible byte substituted.
//
// The last case is the one that keeps the fix honest: a non-breaking space in
// innocent prose must still be ACCEPTED, or an evasion has been traded for a
// false refusal.
func TestTheRegimeGateIsNotEvadedByInvisibleRunes(t *testing.T) {
	const (
		nbsp       = " "
		zeroWidth  = "​"
		narrowNbsp = " "
		ideoSpace  = "　"
	)
	evasions := []struct {
		name, position, field, text, signature string
	}{
		{"non-breaking space", "detection", "why_a_tension",
			"the record and the tree disagree; the" + nbsp + "fix" + nbsp + "is to restate the constraint",
			"RG-REG-FIXPROPOSAL"},
		{"zero-width space inside a keyword", "detection", "why_a_tension",
			"the record and the tree disagree; the fi" + zeroWidth + "x is to restate the constraint",
			"RG-REG-FIXPROPOSAL"},
		{"narrow no-break space", "entailment", "what_implies_it",
			"the passed material implies it, and the claim is" + narrowNbsp + "already" + narrowNbsp + "accepted",
			"RG-EXPL-DISPOSITION"},
		{"ideographic space", "comparative", "characterisation",
			"options of this shape behave unremarkably, and we" + ideoSpace + "recommend option B",
			"RG-EVAL-RECOMMENDATION"},
	}
	for _, tc := range evasions {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			f := newIngestFixture(t, Position(tc.position))
			doc := f.payload(2)
			doc["items"].([]any)[1].(map[string]any)[tc.field] = tc.text

			r := f.refusedItem(doc, 2, 2)
			if r.Rule != tc.signature {
				t.Errorf("the refusal cites rule %q, want the registered signature %s", r.Rule, tc.signature)
			}
		})
	}

	// The folding is for MATCHING only: the stored text keeps its own bytes.
	t.Run("innocent prose carrying a non-breaking space is accepted", func(t *testing.T) {
		f := newIngestFixture(t, "detection")
		doc := f.payload(1)
		doc["items"].([]any)[0].(map[string]any)["why_a_tension"] =
			"the constraint names a budget of 30" + nbsp + "ms, and the shipped path exceeds it"

		res := f.mustIngest(doc)
		if len(res.Records) != 1 {
			t.Fatalf("innocent prose was refused: %v", res.RefusedItems)
		}
		if len(res.RefusedItems) != 0 {
			t.Errorf("an evasion was traded for a false refusal: %v", res.RefusedItems)
		}
	})
}
