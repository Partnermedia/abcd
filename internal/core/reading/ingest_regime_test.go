package reading

// ingest_regime_test.go is the supply-regime gate: itd-185's ac-4 through ac-9
// and ac-12.
//
// Every refusal case here carries the adjacent LEGAL payload in the same test.
// ac-7 is that assertion promoted to a criterion of its own, and the same shape
// is owed to the rest: a refusal test that would pass against a verb refusing
// everything establishes nothing about the rule it names.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
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
	f.parkRun("rdg-2608310000000002", "detection", AssemblerVersion())
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
// The refusal IS recorded. The run's identity is proven before the definition is
// resolved — it names a parked run and cites that run's manifest by content hash
// — and from that point every list-level refusal leaves a record, which is what
// ac-10 and the plugin page both say. The earlier reading, that a run whose
// definition does not resolve has no regime the verb could honestly write down,
// argued for writing nothing at all; the honest form is a record whose `regime`
// is EMPTY and whose reason says why, because that is a fact an operator can
// find, and a refusal nobody recorded is not (iss-2608311518250688).
func TestADriftedDefinitionRefusesTheRunRatherThanChangingTheLicence(t *testing.T) {
	f := newIngestFixture(t, "detection")
	f.mustIngest(f.payload(1))

	// The detection definition now states the evaluative licence. Under it, the
	// evaluative reserved names would be enforced against a registrative
	// reading, and the registrative ones would not be enforced at all.
	f.writeDefinition("detection", RegimeEvaluative)

	f.parkRun("rdg-2608310000000007", "detection", AssemblerVersion())
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
	if res.RefusalPath == "" {
		t.Fatal("a run refused after its identity was proven left no refusal record; ac-10 and the " +
			"plugin page both say the event is durable from that point on")
	}
	rec := f.readRefusalRecord("rdg-2608310000000007")
	if rec.RunID != "rdg-2608310000000007" || rec.Position != "detection" {
		t.Errorf("the refusal record carries the wrong run metadata: %+v", rec)
	}
	if rec.ManifestSHA256 == "" || rec.TargetCommit == "" {
		t.Errorf("the refusal record does not carry the run's manifest reference: %+v", rec)
	}
	// The regime is the definition's, and the definition did not resolve, so
	// there is none to state. An empty field is the honest value; a substituted
	// one would be the verb asserting a licence it refused to read.
	if rec.Regime != "" {
		t.Errorf("the refusal record states regime %q for a run whose definition did not resolve", rec.Regime)
	}
	if !strings.Contains(rec.Reason, RegimeEvaluative) || !strings.Contains(rec.Reason, "refused rather than resolved") {
		t.Errorf("the refusal record does not carry the named reason: %q", rec.Reason)
	}
	f.nothingDurableInTheLedger("rdg-2608310000000007")
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

// TestRegistrativeProseFixProposalFlagged is ac-5: a body matching a registered
// fix-proposal signature raises a review flag naming the item and the signature
// id, and the item lands. The signature is OBSERVED rather than enforcing — it
// cannot tell a reading that proposes a fix from one reporting that the document
// proposes one — so what a hit buys is a reader's attention, not a refusal.
//
// The residue is disclosed on itd-185 and is not tested here, because a residue
// is what a test cannot assert.
func TestRegistrativeProseFixProposalFlagged(t *testing.T) {
	f := newIngestFixture(t, "detection")
	doc := f.payload(2)
	doc["items"].([]any)[1].(map[string]any)["why_a_tension"] =
		"the record says one thing and the tree another; the fix is to restate the constraint"

	res := f.mustIngest(doc)
	if len(res.Records) != 2 {
		t.Fatalf("a signature hit refused the item: %d record(s), refusals %v", len(res.Records), res.RefusedItems)
	}
	if len(res.RefusedItems) != 0 {
		t.Errorf("an observed signature refused: %v", res.RefusedItems)
	}
	fl := f.flagOf(res, 2)
	if fl.SignatureID != "RG-REG-FIXPROPOSAL" {
		t.Errorf("the flag cites signature %q, want the registered signature id", fl.SignatureID)
	}
	if !strings.Contains(fl.Detail, "item 2") {
		t.Errorf("the flag does not name the item: %q", fl.Detail)
	}
	if !strings.Contains(fl.Detail, "RG-REG-FIXPROPOSAL") {
		t.Errorf("the flag does not name the signature id: %q", fl.Detail)
	}

	// And it is durable: the flag is on the run record, not only in the render.
	run := f.readRunRecord(f.runID)
	if len(run.ReviewFlags) != 1 || run.ReviewFlags[0].SignatureID != "RG-REG-FIXPROPOSAL" {
		t.Errorf("the run record carries review flags %v", run.ReviewFlags)
	}
}

// TestAReadingQuotingItsMaterialIsNotRefused is the reason the four semantic
// signatures are observed rather than enforcing, held as a check.
//
// Every case below is prose a real cold reading legitimately writes: it REPORTS
// what the read document says. The registry cannot distinguish that from the
// reading proposing it itself, and under enforcement each of these refused an
// item and cost the reading a finding.
func TestAReadingQuotingItsMaterialIsNotRefused(t *testing.T) {
	for _, tc := range []struct{ name, position, field, text string }{
		{"a detection reporting a settled fix", "detection", "why_a_tension",
			"section 3 says the fix is already merged while section 8 says it is pending"},
		{"a comparative quoting a recommendation", "comparative", "characterisation",
			"the cited paper closes with the sentence, we recommend further study"},
		{"a comparative reporting an adoption clause", "comparative", "characterisation",
			"clause 6 says the MIT licence should be adopted before the release"},
		{"a comparative reporting a score", "comparative", "characterisation",
			"the suite scores below the threshold the charter names"},
		{"an entailment quoting a record's disposition key", "entailment", "what_implies_it",
			"the record's frontmatter carries disposition: held, which is what implies it"},
		{"an entailment reporting a settled question", "entailment", "what_implies_it",
			"section 4 asserts the licensing question is already settled by adr-43"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			f := newIngestFixture(t, Position(tc.position))
			doc := f.payload(1)
			doc["items"].([]any)[0].(map[string]any)[tc.field] = tc.text

			res := f.mustIngest(doc)
			if len(res.Records) != 1 {
				t.Fatalf("a reading quoting its material was refused: %v", res.RefusedItems)
			}
			if len(res.RefusedItems) != 0 {
				t.Errorf("a reading quoting its material was refused: %v", res.RefusedItems)
			}
		})
	}
}

// TestAnEnforcingSignatureStillRefuses keeps the registry's enforcing half
// executable. No shipped signature is in that mode, so without this the branch
// checkItem runs would be a claim nothing checks — and the degradation would
// have quietly removed the mechanism instead of the four entries' use of it.
func TestAnEnforcingSignatureStillRefuses(t *testing.T) {
	prior := Signatures
	Signatures = append(append([]Signature{}, prior...), Signature{
		ID: "RG-TEST-ENFORCED", Regime: RegimeRegistrative, Mode: SignatureEnforce,
		Licence: "a test-only signature, registered enforced",
		Pattern: regexp.MustCompile(`(?i)\bthe canary phrase\b`),
	})
	t.Cleanup(func() { Signatures = prior })

	f := newIngestFixture(t, "detection")
	doc := f.payload(2)
	doc["items"].([]any)[1].(map[string]any)["why_a_tension"] =
		"the record and the tree disagree, and the canary phrase is in the body"

	r := f.refusedItem(doc, 2, 2)
	if r.Rule != "RG-TEST-ENFORCED" {
		t.Errorf("the refusal cites rule %q, want the enforced signature id", r.Rule)
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

// TestExplicativeProseDispositionFlagged is ac-9: a claim body matching a
// registered disposition signature raises a review flag naming the item and the
// signature id, and the item lands.
func TestExplicativeProseDispositionFlagged(t *testing.T) {
	f := newIngestFixture(t, "entailment")
	doc := f.payload(2)
	doc["items"].([]any)[0].(map[string]any)["what_implies_it"] =
		"the passed material implies it, and the claim is already accepted by the record"

	res := f.mustIngest(doc)
	if len(res.Records) != 2 {
		t.Fatalf("a signature hit refused the item: %d record(s), refusals %v", len(res.Records), res.RefusedItems)
	}
	fl := f.flagOf(res, 1)
	if fl.SignatureID != "RG-EXPL-DISPOSITION" {
		t.Errorf("the flag cites signature %q, want the registered signature id", fl.SignatureID)
	}
	if !strings.Contains(fl.Detail, "item 1") {
		t.Errorf("the flag does not name the item: %q", fl.Detail)
	}
	if !strings.Contains(fl.Detail, "RG-EXPL-DISPOSITION") {
		t.Errorf("the flag does not name the signature id: %q", fl.Detail)
	}
	run := f.readRunRecord(f.runID)
	if len(run.ReviewFlags) != 1 || run.ReviewFlags[0].SignatureID != "RG-EXPL-DISPOSITION" {
		t.Errorf("the run record carries review flags %v", run.ReviewFlags)
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

// TestEverySignatureShipsInItsRecordedMode pins the registry ENTRY BY ENTRY.
//
// A property saying "every signature ships enforced" was true until the four
// semantic detectors were degraded to flag mode, and a property saying "every
// signature ships flagged" would be the same instrument pointed the other way:
// it would let a signature be promoted to enforcement without anything failing.
// The mode of each entry is a decision with a decision-log line behind it, so it
// is pinned by name, and a move in EITHER direction turns this red.
func TestEverySignatureShipsInItsRecordedMode(t *testing.T) {
	// The four semantic detectors, observed rather than enforcing since the
	// degradation of 2026-08-31: the registry cannot tell a reading that
	// proposes from one reporting somebody else proposing.
	want := map[string]SignatureMode{
		"RG-EVAL-ORDERING":       SignatureFlag,
		"RG-EVAL-RECOMMENDATION": SignatureFlag,
		"RG-REG-FIXPROPOSAL":     SignatureFlag,
		"RG-EXPL-DISPOSITION":    SignatureFlag,
	}
	if len(Signatures) != len(want) {
		t.Fatalf("the registry holds %d signature(s) and %d are pinned here; a signature ships with its "+
			"mode recorded, so a new one is pinned in the same change", len(Signatures), len(want))
	}
	seen := map[string]bool{}
	for _, s := range Signatures {
		mode, pinned := want[s.ID]
		if !pinned {
			t.Errorf("%s is registered and not pinned here", s.ID)
			continue
		}
		if s.Mode != mode {
			t.Errorf("%s ships in %q mode, and %q is what the record says; moving a signature between "+
				"enforce and flag is a code change plus a decision-log entry", s.ID, s.Mode, mode)
		}
		if seen[s.ID] {
			t.Errorf("%s is registered twice", s.ID)
		}
		seen[s.ID] = true
		if s.Pattern == nil {
			t.Errorf("%s registers no detector", s.ID)
		}
		if strings.TrimSpace(s.Licence) == "" {
			t.Errorf("%s states no licence, so a flag citing it could not say what was breached", s.ID)
		}
		if _, ok := regimeLicence[s.Regime]; !ok {
			t.Errorf("%s polices regime %q, which is not one of the four", s.ID, s.Regime)
		}
	}
	for id := range want {
		if !seen[id] {
			t.Errorf("the registry does not carry %s, which spc-63 names", id)
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
// space inside a keyword, matched nothing and the item landed unremarked —
// inside the gate this verb exists to be.
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
		// U+034F is a MARK, not a format rune, and U+FE00 is a variation
		// selector: guarding Cf alone was the same defect one category over.
		{"combining grapheme joiner", "detection", "why_a_tension",
			"the record and the tree disagree; the f\u034fix is to restate the constraint",
			"RG-REG-FIXPROPOSAL"},
		{"variation selector", "detection", "why_a_tension",
			"the record and the tree disagree; the fix\ufe00 is to restate the constraint",
			"RG-REG-FIXPROPOSAL"},
		// Compatibility forms: the registry's own phrasing in code points that
		// render the same. NFKC folds them.
		{"ligature", "detection", "why_a_tension",
			"the record and the tree disagree; the \ufb01x is to restate the constraint",
			"RG-REG-FIXPROPOSAL"},
		{"fullwidth letters", "comparative", "characterisation",
			"options of this shape behave unremarkably, and \uff57\uff45 recommend option B",
			"RG-EVAL-RECOMMENDATION"},
	}
	for _, tc := range evasions {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			f := newIngestFixture(t, Position(tc.position))
			doc := f.payload(2)
			doc["items"].([]any)[1].(map[string]any)[tc.field] = tc.text

			res := f.mustIngest(doc)
			if len(res.RefusedItems) != 0 {
				t.Fatalf("an observed signature refused: %v", res.RefusedItems)
			}
			if fl := f.flagOf(res, 2); fl.SignatureID != tc.signature {
				t.Errorf("the flag cites signature %q, want the registered signature %s",
					fl.SignatureID, tc.signature)
			}
		})
	}

	// The folding is for MATCHING only: the stored text keeps its own bytes.
	//
	// This battery is the other half of the fix and the more important one. Every
	// transformation above widens what refuses, and NFKC is more aggressive than
	// the rest — so each case here is ordinary prose that legitimately carries
	// one of the runes being folded, and every one must still LAND. An evasion
	// traded for a false refusal costs a reading its findings, which is the cost
	// itd-185's open question is about.
	t.Run("innocent prose still lands", func(t *testing.T) {
		innocent := []struct{ name, text string }{
			{"non-breaking space in a measurement", "the constraint names a budget of 30" + nbsp + "ms"},
			{"soft hyphen at a line break", "the constraint is docu\u00admented in the record"},
			{"narrow no-break space in a figure", "the budget is 30" + narrowNbsp + "%, and the path exceeds it"},
			{"zero-width non-joiner in a compound", "the record names a well\u200cknown constraint"},
			{"combining acute", "the re\u0301cord and the tree disagree about the constraint"},
			{"variation selector after a glyph", "the record marks it \u2714\ufe0f and the tree does not"},
			{"the bare word fix", "the constraint names a fix window, and the tree has none"},
			{"the bare word order", "the record states an order of operations the tree does not keep"},
			{"the word recommend in quoted material", "the passed material uses the word recommend"},
			{"ideographic text", "the record states \u5236\u7d04 and the tree does not"},
			{"an em dash and curly quotes", "the record says \u201cone thing\u201d \u2014 the tree says another"},
			{"a ligature in ordinary prose", "the record de\ufb01nes the constraint the tree ignores"},
		}
		for _, tc := range innocent {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				f := newIngestFixture(t, "detection")
				doc := f.payload(1)
				doc["items"].([]any)[0].(map[string]any)["why_a_tension"] = tc.text

				res := f.mustIngest(doc)
				if len(res.Records) != 1 {
					t.Fatalf("innocent prose was refused: %v", res.RefusedItems)
				}
				if len(res.RefusedItems) != 0 {
					t.Errorf("an evasion was traded for a false refusal: %v", res.RefusedItems)
				}
				if len(res.ReviewFlags) != 0 {
					t.Errorf("an evasion was traded for a false flag: %v", res.ReviewFlags)
				}
			})
		}
	})
}

// TestTheSignaturesReadTheProvenanceFieldToo closes a channel that was open at
// every regime.
//
// The detectors read the item's body, and `pattern` is an ENVELOPE field, so an
// item whose pattern carried the registry's own phrasing landed unremarked: a
// registrative pattern proposing a fix, an explicative one carrying a
// disposition, an evaluative one recommending a candidate. No byte was
// substituted and no phrasing was novel — the detector simply did not read that
// field, and it is the field every item at every regime must carry, so the
// channel was always open.
//
// By itd-185's own test that is a defect in the gate rather than the residue:
// the residue covers phrasing OUTSIDE the registry, and this was the registry's
// exact phrasing moved one field along.
func TestTheSignaturesReadTheProvenanceFieldToo(t *testing.T) {
	for _, tc := range []struct{ name, position, pattern, signature string }{
		{"a registrative pattern proposing a fix", "detection",
			"P-1. The fix is to rewrite the charter.", "RG-REG-FIXPROPOSAL"},
		{"an explicative pattern carrying a disposition", "entailment",
			"P-2. This claim is already accepted by the maintainer.", "RG-EXPL-DISPOSITION"},
		{"an evaluative pattern recommending a candidate", "comparative",
			"P-3. We recommend the second candidate.", "RG-EVAL-RECOMMENDATION"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			f := newIngestFixture(t, Position(tc.position))
			doc := f.payload(2)
			doc["items"].([]any)[1].(map[string]any)[PatternField] = tc.pattern

			res := f.mustIngest(doc)
			if fl := f.flagOf(res, 2); fl.SignatureID != tc.signature {
				t.Errorf("the flag cites signature %q, want %s", fl.SignatureID, tc.signature)
			}
			// And it is durable, so the channel is closed in the record and not
			// only in the render.
			run := f.readRunRecord(f.runID)
			if len(run.ReviewFlags) != 1 || run.ReviewFlags[0].SignatureID != tc.signature {
				t.Errorf("the run record carries review flags %v", run.ReviewFlags)
			}
		})
	}

	// The generative position runs the whole registry, and its pattern is read
	// under the same rule.
	t.Run("a generative pattern is read too", func(t *testing.T) {
		f := newIngestFixture(t, PositionWidening)
		doc := f.payload(1)
		doc["items"].([]any)[0].(map[string]any)[PatternField] = "P-4. We recommend this configuration."

		res := f.mustIngest(doc)
		if fl := f.flagOf(res, 1); fl.SignatureID != "RG-EVAL-RECOMMENDATION" {
			t.Errorf("the flag cites signature %q", fl.SignatureID)
		}
	})

	// An ordinary pattern still raises nothing: a channel closed by flagging
	// every provenance value would be a channel traded for noise.
	t.Run("an ordinary pattern raises nothing", func(t *testing.T) {
		f := newIngestFixture(t, "detection")
		res := f.mustIngest(f.payload(1))
		if len(res.ReviewFlags) != 0 {
			t.Errorf("an ordinary pattern was flagged: %v", res.ReviewFlags)
		}
	})
}

// TestTheElisionEntryNamesNoItemInAnyDurableRecord.
//
// A bounded refusal list ends in an entry that is not an item: "and N more
// item(s) refused". It carries no ordinal, and the zero value of an ordinal is
// 0 — so both durable records asserted an item 0 that does not exist. The
// terminal renderer had the branch that suppresses it and neither record writer
// did, which is the shape a claim takes when it lives in one surface instead of
// in the value.
func TestTheElisionEntryNamesNoItemInAnyDurableRecord(t *testing.T) {
	t.Run("the run record", func(t *testing.T) {
		f := newIngestFixture(t, "detection")
		// More refusals than the cap, and survivors, so the run lands and its
		// commit marker carries the bounded list.
		doc := f.payload(maxReportedRefusals + 5)
		items := doc["items"].([]any)
		for i := 0; i < maxReportedRefusals+3; i++ {
			items[i].(map[string]any)[PatternField] = ""
		}
		res := f.mustIngest(doc)
		if res.RefusedCount != maxReportedRefusals+3 {
			t.Fatalf("the run refused %d item(s), want %d", res.RefusedCount, maxReportedRefusals+3)
		}
		run := f.readRunRecord(f.runID)
		last := run.RefusedItems[len(run.RefusedItems)-1]
		if last.Rule != "refusals-elided" {
			t.Fatalf("the bounded list does not end in the elision entry: %+v", last)
		}
		if last.Ordinal != 0 {
			t.Fatalf("the elision entry carries ordinal %d", last.Ordinal)
		}
		raw, err := os.ReadFile(filepath.Join(f.root, filepath.FromSlash(
			ReadingsRecordDir+"/"+f.runID+"/"+RunFileName)))
		if err != nil {
			t.Fatal(err)
		}
		var doc2 struct {
			RefusedItems []map[string]any `json:"refused_items"`
		}
		if err := json.Unmarshal(raw, &doc2); err != nil {
			t.Fatal(err)
		}
		entry := doc2.RefusedItems[len(doc2.RefusedItems)-1]
		if _, named := entry["ordinal"]; named {
			t.Errorf("the run record's elision entry states an ordinal: %v; there is no item 0, and a "+
				"reader sent looking for one finds nothing", entry)
		}
	})

	t.Run("the refusal record", func(t *testing.T) {
		f := newIngestFixture(t, "detection")
		// Every item refused, so the run is refused at list level and the reason
		// carries the bounded list into refusal.json.
		doc := f.payload(maxReportedRefusals + 5)
		for _, it := range doc["items"].([]any) {
			it.(map[string]any)[PatternField] = ""
		}
		if _, err := f.ingest(doc); err == nil {
			t.Fatal("a run in which every item was refused was accepted")
		}
		rec := f.readRefusalRecord(f.runID)
		if !strings.Contains(rec.Reason, "more item(s) refused") {
			t.Fatalf("the reason does not carry the elision entry: %q", rec.Reason)
		}
		if strings.Contains(rec.Reason, "item 0") {
			t.Errorf("the refusal record names an item 0: %q", rec.Reason)
		}
	})
}

// TestAPatternThatRendersAsNothingIsRefused.
//
// U+2800 BRAILLE PATTERN BLANK renders as nothing in every common font, but it
// is a graphic character: not Cf, not Other_Default_Ignorable_Code_Point, not a
// Variation_Selector, and unicode.IsSpace is false for it. So it cleared the
// blankness test, and a record asserted a provenance that renders blank — the
// failure the U+200B fix closed, one category further out.
func TestAPatternThatRendersAsNothingIsRefused(t *testing.T) {
	for _, p := range Positions() {
		p := p
		t.Run(string(p), func(t *testing.T) {
			f := newIngestFixture(t, p)
			doc := f.payload(2)
			doc["items"].([]any)[1].(map[string]any)[PatternField] = "⠀⠀"

			r := f.refusedItem(doc, 2, 2)
			if r.Rule != "named-provenance" {
				t.Errorf("the refusal cites rule %q, want named-provenance", r.Rule)
			}
			if r.Field != PatternField {
				t.Errorf("the refusal names field %q", r.Field)
			}
		})
	}

	// A braille pattern with dots is ordinary text and still lands: the fold
	// closes the blank, not the block.
	t.Run("a braille pattern with dots lands", func(t *testing.T) {
		f := newIngestFixture(t, "detection")
		doc := f.payload(1)
		doc["items"].([]any)[0].(map[string]any)[PatternField] = "⠁⠃"
		if res := f.mustIngest(doc); len(res.Records) != 1 {
			t.Errorf("a braille pattern carrying dots was refused: %v", res.RefusedItems)
		}
	})
}
