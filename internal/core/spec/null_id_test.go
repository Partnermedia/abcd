package spec

import (
	"strings"
	"testing"
)

// TestParseSpecNullIDReadsAsUnset (iss-2608270908332975) proves a YAML null id
// (`id: NULL`, `id: ~`, `id: null`) is diagnosed as UNSET, not as a malformed
// literal value. Without the frontmatter.IsNull gate, parseSpec fed the literal
// "NULL" to Validate, which quoted it back as a bad id ("malformed shape") —
// while record-lint's schema gate, which gates on frontmatter.IsNull first,
// called the same field unset. Two diagnoses for one field. The fix routes every
// null spelling to the empty (unset) value so the two gates agree.
func TestParseSpecNullIDReadsAsUnset(t *testing.T) {
	for _, spelling := range []string{"NULL", "null", "Null", "~"} {
		content := "---\nid: " + spelling + "\nslug: thing\nintent: itd-9\n---\n# x\n"
		_, err := parseSpec(".abcd/development/specs/open/spc-x.md", content, StatusOpen)
		if err == nil {
			t.Fatalf("id: %s must still be rejected (an unset id is invalid)", spelling)
		}
		if strings.Contains(err.Error(), spelling) {
			t.Fatalf("id: %s diagnosed as a malformed literal (error quotes %q); a null id must read as unset:\n%s", spelling, spelling, err.Error())
		}
		if !strings.Contains(err.Error(), `id ""`) {
			t.Fatalf("id: %s error does not name the id as unset (empty):\n%s", spelling, err.Error())
		}
	}
}

// TestParseSpecNullIntentReadsAsUnset holds the same contract for the
// load-bearing intent link.
func TestParseSpecNullIntentReadsAsUnset(t *testing.T) {
	content := "---\nid: spc-1\nslug: thing\nintent: NULL\n---\n# x\n"
	_, err := parseSpec(".abcd/development/specs/open/spc-1-thing.md", content, StatusOpen)
	if err == nil {
		t.Fatal("intent: NULL must still be rejected (an unset intent is invalid)")
	}
	if strings.Contains(err.Error(), "NULL") {
		t.Fatalf("intent: NULL diagnosed as a malformed literal; a null intent must read as unset:\n%s", err.Error())
	}
}
