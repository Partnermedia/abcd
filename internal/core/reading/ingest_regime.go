package reading

// ingest_regime.go is the supply-regime gate (itd-185, spc-63): the half of the
// output contract that checks what the reading was LICENSED to produce, not only
// what it saw.
//
// The regime's source of truth is the DEFINITION, resolved from the run's
// position. Two enforcement layers sit behind that comparison.
//
// RESERVED NAMES are structural and absolute: a field is present or it is not.
// Strict decoding would already refuse an unknown field, but a bare "unknown
// field" is a poor account of a licence breach, so each regime declares the
// names that name one and the refusal states the licence.
//
// SEMANTIC SIGNATURES are bounded by a registry: prose that ranks, settles or
// proposes without the field. itd-185 discloses that residue — a fix proposal or
// a disposition phrased outside the registry is not caught. Every signature
// ships in enforce mode; degrading one is a code change plus a decision-log
// entry, which is what makes a weakening from enforced to observed a recorded
// act rather than a runtime toggle.

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/intentdriven/abcd/internal/core/capture"
	"github.com/intentdriven/abcd/internal/core/issueschema"
)

// The four supply regimes, named where the gate branches on them so a literal
// never drifts out of the table it came from.
const (
	RegimeGenerative   = "generative"
	RegimeExplicative  = "explicative"
	RegimeEvaluative   = "evaluative"
	RegimeRegistrative = "registrative"
)

// ReservedNames is the per-regime reserved-name table. A payload naming one of
// these is refused with the field named and the licence stated.
//
// `generative` has no row and needs none: its body schema is two fields, so any
// other key is refused as an unknown field anyway, and the generative licence is
// the widest — the constraint on it falls at admission, not here.
var ReservedNames = map[string][]string{
	RegimeEvaluative:   {"order", "rank", "recommended", "score"},
	RegimeRegistrative: {"fix", "remedy", "resolution"},
	RegimeExplicative:  {"disposition", "status"},
}

// regimeLicence states, per regime, what the reserved names breach. It is the
// sentence a refusal quotes, so an operator reads what the rule protects rather
// than only which field tripped it.
var regimeLicence = map[string]string{
	RegimeGenerative: "a generative reading proposes configurations and what would admit them",
	RegimeExplicative: "an explicative reading surfaces a claim; dispositioning one is the " +
		"researcher's act, and it is a separate record",
	RegimeEvaluative: "an evaluative reading characterises candidates against criteria; ordering, " +
		"scoring or recommending among them is not its licence",
	RegimeRegistrative: "a registrative reading names a tension and the constraint in play; " +
		"proposing the resolution is not its licence",
}

// SignatureMode is whether a signature refuses or records.
type SignatureMode string

const (
	// SignatureEnforce refuses the item. Every shipped signature is in this mode.
	SignatureEnforce SignatureMode = "enforce"
	// SignatureFlag records the hit on the run record and lands the item. It is
	// the reserved degradation path; no shipped signature uses it.
	SignatureFlag SignatureMode = "flag"
)

// Signature is one named detector over an item's body text.
type Signature struct {
	// ID is the name a refusal cites, so a reader can find the rule that fired.
	ID string
	// Regime is the supply regime this signature polices.
	Regime string
	// Mode is a literal in Go with no configuration seam. Degrading a signature
	// on observed noise is an edit here plus a decision-log entry, which is what
	// makes the weakening from enforced to observed a recorded act.
	Mode SignatureMode
	// Licence is what a hit breaches.
	Licence string
	// Pattern is the detector. Every one is case-insensitive and anchored on a
	// verb phrase rather than on a bare word: a signature that fired on the word
	// "recommend" appearing anywhere would refuse a reading for quoting the
	// material it was handed.
	Pattern *regexp.Regexp
}

// Signatures is the registry. It is deliberately small and conservative: whether
// these lint cleanly in practice is itd-185's recorded open question, and a
// noisy signature costs a reading its findings.
var Signatures = []Signature{
	{
		ID: "RG-EVAL-ORDERING", Regime: RegimeEvaluative, Mode: SignatureEnforce,
		Licence: regimeLicence[RegimeEvaluative],
		Pattern: regexp.MustCompile(`(?i)\b(?:ranks?|ranked|rates?|rated|scores?|scored)\s+` +
			`(?:it\s+|them\s+|this\s+)?(?:first|second|third|last|highest|lowest|above|below)\b` +
			`|\bin\s+order\s+of\s+(?:merit|preference|strength|quality)\b` +
			`|\b(?:the\s+)?(?:strongest|weakest|best|worst)\s+(?:candidate|option|choice)\b`),
	},
	{
		ID: "RG-EVAL-RECOMMENDATION", Regime: RegimeEvaluative, Mode: SignatureEnforce,
		Licence: regimeLicence[RegimeEvaluative],
		Pattern: regexp.MustCompile(`(?i)\b(?:we|i)\s+recommend\b` +
			`|\brecommend(?:ation)?\s+(?:is|that)\b` +
			`|\bshould\s+be\s+(?:chosen|selected|adopted|preferred|picked)\b`),
	},
	{
		ID: "RG-REG-FIXPROPOSAL", Regime: RegimeRegistrative, Mode: SignatureEnforce,
		Licence: regimeLicence[RegimeRegistrative],
		Pattern: regexp.MustCompile(`(?i)\bthe\s+(?:fix|remedy|resolution)\s+is\b` +
			`|\bto\s+fix\s+(?:this|it|that)\b` +
			`|\bpropos(?:e|es|ed|ing)\s+(?:a\s+|the\s+)?(?:fix|remedy|resolution)\b` +
			`|\bshould\s+be\s+(?:changed|replaced|rewritten|removed|deleted)\s+to\b`),
	},
	{
		ID: "RG-EXPL-DISPOSITION", Regime: RegimeExplicative, Mode: SignatureEnforce,
		Licence: regimeLicence[RegimeExplicative],
		Pattern: regexp.MustCompile(`(?i)\b(?:this\s+claim\s+is|the\s+claim\s+is|it\s+is)\s+` +
			`(?:already\s+)?(?:accepted|rejected|declined|settled|resolved)\b` +
			`|\balready\s+(?:accepted|rejected|declined|settled|resolved)\b` +
			`|\bdisposition\s*[:=]`),
	},
}

// checkRegime is ac-12: the payload's self-declared regime is compared against
// the definition's, never trusted. A disagreement refuses the RUN, not an item —
// a run read under the wrong licence is wrong in whole, and which items it would
// have been wrong about is not a question worth asking.
func checkRegime(out Output, def Definition) error {
	if out.Regime == def.Regime {
		return nil
	}
	return fmt.Errorf("the output declares the %s regime and %s states %s; the regime is the "+
		"definition's property, resolved from the run's position, and a self-declared regime that "+
		"disagrees with it refuses the run", echo(out.Regime), def.Path, def.Regime)
}

// checkInstrument closes the instrument identity against the two things that can
// disagree with it (ruling (12)): the definition file, whose bytes this verb
// hashes itself, and the manifest of the run the output names.
//
// Presence of all three parts is already established by the envelope check. What
// happens here is the part presence cannot do: a claim is compared with the
// artefact it is a claim about, so "two runs claiming the same instrument are
// provably the same" is a proof rather than a convention.
func checkInstrument(out Output, def Definition, m Manifest) error {
	if out.Instrument.DefinitionSHA256 != def.SHA256 {
		return fmt.Errorf("the instrument claims definition_sha256 %s and %s hashes to %s; the "+
			"definition's content hash is half of the instrument's identity, and it is recomputed here",
			echo(out.Instrument.DefinitionSHA256), def.Path, def.SHA256)
	}
	if out.Instrument.AssemblerVersion != m.AssemblerVersion {
		return fmt.Errorf("the instrument claims assembler_version %s and the manifest of run %s "+
			"carries %s", echo(out.Instrument.AssemblerVersion), m.RunID, m.AssemblerVersion)
	}
	return nil
}

// validateItems judges every item and turns the survivors into ledger items.
//
// An item-level violation refuses THAT item and lands the rest; only a payload
// in which nothing survives becomes a list-level refusal, because a run with no
// items is not a run with an empty item set — it is a run whose every finding
// was refused, and recording it as the former would lose that.
func validateItems(out Output, def Definition) ([]capture.ReadingItem, []ItemRefusal, []ReviewFlag, error) {
	bodyFields := issueschema.ReadingBodyFields[string(def.Position)]
	allowed := map[string]bool{PatternField: true}
	for _, f := range bodyFields {
		allowed[f] = true
	}

	var items []capture.ReadingItem
	refusals := []ItemRefusal{}
	flags := []ReviewFlag{}

	for i, raw := range out.Items {
		ordinal := i + 1
		fields, err := decodeItemFields(raw)
		if err != nil {
			refusals = append(refusals, ItemRefusal{Ordinal: ordinal, Rule: "item-shape",
				Detail: fmt.Sprintf("item %d: %s", ordinal, echo(err.Error()))})
			continue
		}
		if r := checkItem(ordinal, fields, def, allowed, bodyFields); r != nil {
			refusals = append(refusals, *r)
			continue
		}
		flags = append(flags, itemReviewFlags(ordinal, fields, def, bodyFields)...)

		body := make(map[string]string, len(bodyFields))
		for _, f := range bodyFields {
			body[f] = fields[f]
		}
		items = append(items, capture.ReadingItem{Pattern: fields[PatternField], Body: body})
	}

	if len(items) == 0 {
		return nil, refusals, flags, fmt.Errorf("every one of the %d item(s) was refused, so the run "+
			"carries nothing to record: %s", len(out.Items), renderRefusals(refusals))
	}
	return items, refusals, flags, nil
}

// checkItem judges one item, returning the first rule it breaks.
//
// The order is deliberate. RESERVED NAMES come first, because strict decoding
// would refuse the same field as a bare unknown one and a licence breach
// deserves a better account than that. PROVENANCE comes next, before the body,
// because it is the one condition every regime shares. Then the body's own key
// set, and only then the signatures — a body that is not yet well formed is not
// prose worth running a detector over.
func checkItem(ordinal int, fields map[string]string, def Definition,
	allowed map[string]bool, bodyFields []string) *ItemRefusal {

	if named := present(fields, ReservedNames[def.Regime]); len(named) > 0 {
		return &ItemRefusal{Ordinal: ordinal, Rule: "reserved-name", Field: strings.Join(named, ", "),
			Detail: fmt.Sprintf("item %d carries the reserved %s field %s: %s",
				ordinal, def.Regime, renderFields(named), regimeLicence[def.Regime])}
	}

	if strings.TrimSpace(fields[PatternField]) == "" {
		return &ItemRefusal{Ordinal: ordinal, Rule: "named-provenance", Field: PatternField,
			Detail: fmt.Sprintf("item %d names no %q: every item at every regime carries the pattern it "+
				"was read under, without exception, and the definitions instruct it", ordinal, PatternField)}
	}

	var unknown []string
	for k := range fields {
		if !allowed[k] {
			unknown = append(unknown, k)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return &ItemRefusal{Ordinal: ordinal, Rule: "unknown-field", Field: strings.Join(unknown, ", "),
			Detail: fmt.Sprintf("item %d carries %s, which the %s body does not declare (%s); the item "+
				"identity is the verb's to mint and the envelope is the verb's to compose, so neither has "+
				"a field here", ordinal, renderFields(unknown), def.Regime, renderFields(bodyFields))}
	}

	var missing []string
	for _, f := range bodyFields {
		if strings.TrimSpace(fields[f]) == "" {
			missing = append(missing, f)
		}
	}
	if len(missing) > 0 {
		return &ItemRefusal{Ordinal: ordinal, Rule: "missing-body-field", Field: strings.Join(missing, ", "),
			Detail: fmt.Sprintf("item %d states no %s; the %s body is %s",
				ordinal, renderFields(missing), def.Regime, renderFields(bodyFields))}
	}

	for _, s := range Signatures {
		if s.Regime != def.Regime || s.Mode != SignatureEnforce {
			continue
		}
		if s.Pattern.MatchString(bodyText(fields, bodyFields)) {
			return &ItemRefusal{Ordinal: ordinal, Rule: s.ID,
				Detail: fmt.Sprintf("item %d matches the registered signature %s: %s",
					ordinal, s.ID, s.Licence)}
		}
	}
	return nil
}

// itemReviewFlags is the generative path, and the only place a signature hit
// does not refuse.
//
// The generative licence is the widest — a widening reading proposes, which is
// what it is for — so the constraint on what it produces falls at ADMISSION
// rather than here. The prohibitions the widening definition states are still
// worth seeing, so every signature runs and each hit is recorded on the run
// record as a flag. The item lands.
func itemReviewFlags(ordinal int, fields map[string]string, def Definition, bodyFields []string) []ReviewFlag {
	if def.Regime != RegimeGenerative {
		return nil
	}
	var out []ReviewFlag
	text := bodyText(fields, bodyFields)
	for _, s := range Signatures {
		if s.Pattern.MatchString(text) {
			out = append(out, ReviewFlag{Ordinal: ordinal, SignatureID: s.ID,
				Detail: fmt.Sprintf("item %d matches %s; the generative licence does not refuse it, and "+
					"the constraint on a widening reading falls at admission", ordinal, s.ID)})
		}
	}
	return out
}

// bodyText is what a signature reads: the item's body values, joined. The
// pattern is not among them — it names the reading's own basis, not a finding —
// and no key name is either, because a detector over key names would be the
// reserved-name table written twice.
func bodyText(fields map[string]string, bodyFields []string) string {
	parts := make([]string, 0, len(bodyFields))
	for _, f := range bodyFields {
		parts = append(parts, fields[f])
	}
	return strings.Join(parts, "\n")
}

// present returns the members of names the item carries, in the table's order.
func present(fields map[string]string, names []string) []string {
	var out []string
	for _, n := range names {
		if _, ok := fields[n]; ok {
			out = append(out, n)
		}
	}
	return out
}

// renderFields quotes a field list for a message.
func renderFields(names []string) string {
	quoted := make([]string, 0, len(names))
	for _, n := range names {
		quoted = append(quoted, fmt.Sprintf("%q", echo(n)))
	}
	return strings.Join(quoted, ", ")
}

// decodeItemFields decodes one item's values as strings. A non-string value is
// refused naming its field: an item is a flat map of text, as every definition's
// item shape instructs.
func decodeItemFields(raw map[string]json.RawMessage) (map[string]string, error) {
	out := make(map[string]string, len(raw))
	for _, k := range sortedKeys(raw) {
		var s string
		if err := json.Unmarshal(raw[k], &s); err != nil {
			return nil, fmt.Errorf("the field %q is not text; an item is a flat map of text fields", echo(k))
		}
		out[k] = s
	}
	return out, nil
}

// sortedKeys renders a map's keys in a stable order, so two runs over one
// payload refuse in the same order and a refusal list is comparable.
func sortedKeys(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// renderRefusals renders the refusal list for a list-level message.
func renderRefusals(refusals []ItemRefusal) string {
	parts := make([]string, 0, len(refusals))
	for _, r := range refusals {
		parts = append(parts, fmt.Sprintf("item %d (%s): %s", r.Ordinal, r.Rule, r.Detail))
	}
	return strings.Join(parts, "; ")
}
