package lint

import "testing"

// TestIsNull covers lint's private null predicate, the second copy of
// frontmatter.IsNull that record-lint gates on. It must recognise the same four
// YAML nulls (null | Null | NULL | ~) plus the empty scalar as the frontmatter
// copy, or the two-gate agreement capture/validate.go depends on breaks: an
// `impact: NULL` record would pass one gate and be refused by the other.
// Regression for iss #290 — the uppercase spellings were previously missed.
func TestIsNull(t *testing.T) {
	for _, v := range []string{"", "null", "Null", "NULL", "~"} {
		if !isNull(v) {
			t.Errorf("isNull(%q) = false, want true", v)
		}
	}
	for _, v := range []string{"itd-9", "spc-1", "standalone", "None", "nil", "NUL", "nullish"} {
		if isNull(v) {
			t.Errorf("isNull(%q) = true, want false", v)
		}
	}
}

// TestIsAbsentValueUppercaseNull confirms isAbsentValue, which delegates to
// isNull, inherits the widened set: an uppercase YAML null is an absence, not a
// malformed value.
func TestIsAbsentValueUppercaseNull(t *testing.T) {
	for _, v := range []string{"NULL", "Null", "null", "~", "", "[]"} {
		if !isAbsentValue(v) {
			t.Errorf("isAbsentValue(%q) = false, want true", v)
		}
	}
	if isAbsentValue("adr-9") {
		t.Errorf("isAbsentValue(%q) = true, want false", "adr-9")
	}
}
