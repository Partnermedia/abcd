package termsafe

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestJSONLeavesC1AndBidiRaw pins the encoding/json behaviour Sanitize's doc
// note rests on: C0 and U+2028/9 are escaped, while DEL, C1, bidi overrides
// and zero-width runes pass through raw — so a JSON surface is NOT covered by
// the marshaller and needs its values handled at the producing boundary
// (iss-345). If a Go release starts escaping these, this test says so and the
// note can be revisited.
func TestJSONLeavesC1AndBidiRaw(t *testing.T) {
	escaped := []rune{0x0000, 0x001B, 0x2028}
	raw := []rune{0x007F, 0x009B, 0x202E, 0x200B}

	for _, r := range escaped {
		b, err := json.Marshal(string(r))
		if err != nil {
			t.Fatal(err)
		}
		if strings.ContainsRune(string(b), r) {
			t.Errorf("U+%04X expected escaped, marshalled raw: %q", r, b)
		}
	}
	for _, r := range raw {
		b, err := json.Marshal(string(r))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.ContainsRune(string(b), r) {
			t.Errorf("U+%04X expected raw (the hazard this pins), got %q — encoding/json changed; revisit Sanitize's JSON note", r, b)
		}
	}
}
