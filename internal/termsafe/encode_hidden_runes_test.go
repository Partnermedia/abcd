package termsafe

import (
	"testing"
	"unicode/utf8"
)

// TestEncodeHiddenRunesContract pins the lossless percent-encoding contract the
// two former call-site copies (internal/core/cite and internal/core/memory)
// depended on: a value carrying a hidden rune is recorded losslessly but can no
// longer smuggle a terminal escape or a Trojan-Source reorder into a JSON surface
// or a committed record, while a legitimate address is returned byte-for-byte.
func TestEncodeHiddenRunesContract(t *testing.T) {
	// A canonical URL carries none of the masked runes and is returned unchanged.
	const plain = "https://example.com/a/b?q=1&r=2#frag"
	if got := EncodeHiddenRunes(plain); got != plain {
		t.Errorf("EncodeHiddenRunes(%q) = %q, want it unchanged", plain, got)
	}

	// Every class Sanitize masks is percent-encoded, not dropped or substituted.
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"ESC (C0)", withRune("a", 0x1b, "b"), "a%1Bb"},
		{"DEL", withRune("a", 0x7f, "b"), "a%7Fb"},
		{"C1 CSI", withRune("a", 0x9b, "b"), "a%C2%9Bb"},
		{"RLO bidi override", withRune("a", 0x202e, "b"), "a%E2%80%AEb"},
		{"ZWSP zero-width", withRune("a", 0x200b, "b"), "a%E2%80%8Bb"},
		{"tab", withRune("a", 0x09, "b"), "a%09b"},
	}
	for _, c := range cases {
		if got := EncodeHiddenRunes(c.in); got != c.want {
			t.Errorf("%s: EncodeHiddenRunes(%q) = %q, want %q", c.name, c.in, got, c.want)
		}
	}

	// Valid UTF-8 that carries no hidden rune passes through untouched, even when
	// it is multi-byte (an accented host label, say).
	const unicodeOK = "https://xn--caf-dma.example/café"
	if got := EncodeHiddenRunes(unicodeOK); got != unicodeOK {
		t.Errorf("EncodeHiddenRunes(%q) = %q, want it unchanged", unicodeOK, got)
	}

	// An invalid UTF-8 byte — which strings.Map would silently rewrite to U+FFFD —
	// is percent-encoded RAW so the record stays lossless on non-UTF-8 input.
	invalid := "a\xffb"
	if utf8.ValidString(invalid) {
		t.Fatal("test fixture is supposed to be invalid UTF-8")
	}
	if got, want := EncodeHiddenRunes(invalid), "a%FFb"; got != want {
		t.Errorf("EncodeHiddenRunes(invalid) = %q, want %q", got, want)
	}
}
