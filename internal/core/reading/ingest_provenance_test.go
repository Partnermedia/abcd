package reading

// ingest_provenance_test.go is itd-185's ac-11. Named provenance is the one
// condition enforced identically at all four regimes: the definitions instruct
// it, this contract enforces it, and nothing else checks it.

import (
	"strings"
	"testing"
)

// TestEmptyPatternNamedRefusesItemAtEveryRegime is ac-11, and it walks all four
// positions rather than sampling one: "without exception at any regime" is a
// claim about the whole set, and a test over one position would be the same
// evidence for a verb that checked provenance at one position only.
//
// Both the empty and the absent form are tried, because they are different
// payload bytes and only one of them is an obviously missing field.
func TestEmptyPatternNamedRefusesItemAtEveryRegime(t *testing.T) {
	for _, pos := range Positions() {
		pos := pos
		t.Run(string(pos), func(t *testing.T) {
			for _, form := range []string{"empty", "absent", "whitespace"} {
				t.Run(form, func(t *testing.T) {
					f := newIngestFixture(t, pos)
					// Two items, the FIRST illegal: the criterion refuses the
					// ITEM, so the run must land the other one.
					doc := f.payload(2)
					item := doc["items"].([]any)[0].(map[string]any)
					switch form {
					case "empty":
						item[PatternField] = ""
					case "absent":
						delete(item, PatternField)
					case "whitespace":
						item[PatternField] = "   \t "
					}

					// refusedItem establishes both halves at once: the item
					// was refused, and the adjacent legal item in the SAME
					// payload landed. Without the second half this case would
					// pass against a verb that refused every item at every
					// regime, which is the vacuity the provenance rule invites.
					r := f.refusedItem(doc, 1, 2)
					if r.Rule != "named-provenance" {
						t.Errorf("the refusal cites rule %q at the %s regime", r.Rule, f.regime)
					}
					if !strings.Contains(r.Detail, PatternField) {
						t.Errorf("the refusal does not name the provenance field: %q", r.Detail)
					}
				})
			}
		})
	}
}
