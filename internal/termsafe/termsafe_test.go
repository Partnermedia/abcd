package termsafe

import (
	"testing"
)

// esc builds a string from ordinary text with a single attack rune spliced in, so
// the test source itself stays pure ASCII (no invisible characters).
func withRune(prefix string, r rune, suffix string) string {
	return prefix + string(r) + suffix
}

// TestSanitizeNeutralisesDisplayAttacks proves each terminal-display attack class
// is masked while ordinary text and tab survive.
func TestSanitizeNeutralisesDisplayAttacks(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"esc-ansi", withRune("red", 0x1b, "[31mtext"), "red?[31mtext"},
		{"c1-csi", withRune("x", 0x9b, "31mspoof"), "x?31mspoof"},
		{"del", withRune("a", 0x7f, "b"), "a?b"},
		{"newline-forges-lines", withRune("clean", '\n', "abcd audit — 0 errors"), "clean?abcd audit — 0 errors"},
		{"bidi-rlo", withRune("user", 0x202e, "esc"), "user?esc"},
		{"bidi-lri", withRune("a", 0x2066, "b"), "a?b"},
		{"bidi-lrm", withRune("a", 0x200e, "b"), "a?b"},
		{"arabic-mark", withRune("a", 0x061c, "b"), "a?b"},
		{"zero-width-space", withRune("ad", 0x200b, "min"), "ad?min"},
		{"bom", withRune("", 0xfeff, "head"), "?head"},
		{"tab-to-space", "a\tb", "a b"},
		{"plain-unicode-kept", "Emilie ok — fine", "Emilie ok — fine"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Sanitize(c.in); got != c.want {
				t.Errorf("Sanitize(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// TestSanitizeLeavesNoControlOrBidi is a property check: no attack rune survives.
func TestSanitizeLeavesNoControlOrBidi(t *testing.T) {
	in := ""
	for _, r := range []rune{0x1b, 0x9b, 0x7f, 0x202e, 0x2066, 0x200b, 0xfeff, 0x061c, 'a', 'b'} {
		in += string(r)
	}
	for _, r := range Sanitize(in) {
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) ||
			(r >= 0x202A && r <= 0x202E) || (r >= 0x2066 && r <= 0x2069) ||
			r == 0x200E || r == 0x200F || r == 0x061C ||
			r == 0x200B || r == 0x200C || r == 0x200D || r == 0xFEFF {
			t.Fatalf("attack rune %U survived Sanitize", r)
		}
	}
}

// TestSanitizeAll maps a slice and preserves nil.
func TestSanitizeAll(t *testing.T) {
	if SanitizeAll(nil) != nil {
		t.Error("SanitizeAll(nil) must stay nil")
	}
	got := SanitizeAll([]string{withRune("a", 0x1b, "b"), "ok"})
	if len(got) != 2 || got[0] != "a?b" || got[1] != "ok" {
		t.Errorf("SanitizeAll = %v", got)
	}
}
