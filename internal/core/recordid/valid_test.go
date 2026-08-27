package recordid

import "testing"

// TestValidIDPredicates pins the exact accepted set of the one well-formedness
// rule the loaders and the record-lint gate now share. The absent-value
// spellings (empty, YAML nulls) are listed explicitly because they are the shape
// a hand-edited record actually carries.
func TestValidIDPredicates(t *testing.T) {
	intentOK := []string{"itd-1", "itd-999", "itd-0007"}
	intentBad := []string{"", "null", "~", "TBD", "itd-", "itd-1-slug", " itd-1", "itd-1\n", "ITD-1", "spc-1"}
	for _, id := range intentOK {
		if !ValidIntentID(id) {
			t.Errorf("ValidIntentID(%q) = false, want true", id)
		}
	}
	for _, id := range intentBad {
		if ValidIntentID(id) {
			t.Errorf("ValidIntentID(%q) = true, want false", id)
		}
	}

	specOK := []string{"spc-1", "spc-999", "spc-0007"}
	specBad := []string{"", "null", "~", "TBD", "spc-", "spc-1-slug", " spc-1", "spc-1\n", "SPC-1", "itd-1"}
	for _, id := range specOK {
		if !ValidSpecID(id) {
			t.Errorf("ValidSpecID(%q) = false, want true", id)
		}
	}
	for _, id := range specBad {
		if ValidSpecID(id) {
			t.Errorf("ValidSpecID(%q) = true, want false", id)
		}
	}
}
