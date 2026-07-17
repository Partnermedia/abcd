// Package termsafe holds the one canonical sanitiser for a string built from
// untrusted content (commit subjects, refs, file paths, repo prose, error text
// echoing a malformed file) before it is written to a terminal or a human report.
// It is the single primitive every render path shares — the terminal analogue of
// fsutil's guarded read: a hostile or archived repository controls this text, and
// left raw it can spoof, corrupt, or visually rewrite the report.
package termsafe

import "strings"

// Sanitize replaces every terminal-display attack rune with a visible '?' (tab
// becomes a space). It neutralises:
//   - C0 controls (<0x20) and DEL (0x7f) — these carry ESC, so a raw ANSI escape
//     in a commit subject could recolour, move the cursor, or corrupt the report;
//     a newline is masked too, so an injected line break cannot forge extra lines;
//   - the C1 range (0x80–0x9F) — U+009B (CSI) acts like ESC[ on an 8-bit terminal,
//     so masking ESC (a C0 control) alone would leave an equivalent path open;
//   - bidirectional override/isolate controls (the "Trojan Source" class) and
//     zero-width characters, which reorder or hide text so the rendered line
//     differs from the bytes — the reader sees something the file does not say.
//
// JSON output needs no sanitising — encoding/json escapes control characters
// itself — so this is applied only on the human/terminal path.
func Sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == '\t':
			return ' '
		case r < 0x20 || r == 0x7f:
			return '?'
		case r >= 0x80 && r <= 0x9f:
			return '?'
		case isBidiControl(r) || isZeroWidth(r):
			return '?'
		}
		return r
	}, s)
}

// SanitizeAll sanitises every member of a slice, returning a new slice.
func SanitizeAll(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = Sanitize(s)
	}
	return out
}

// isBidiControl reports whether r is a Unicode bidirectional override/embedding/
// isolate or directional mark — the runes a "Trojan Source" attack uses to make a
// rendered line read differently from its bytes. Code points are written
// numerically so this source file carries none of the invisible characters it
// defends against.
func isBidiControl(r rune) bool {
	switch {
	case r >= 0x202A && r <= 0x202E: // LRE RLE PDF LRO RLO
		return true
	case r >= 0x2066 && r <= 0x2069: // LRI RLI FSI PDI
		return true
	case r == 0x200E || r == 0x200F: // LRM RLM
		return true
	case r == 0x061C: // Arabic letter mark
		return true
	}
	return false
}

// isZeroWidth reports whether r is a zero-width character that can hide or splice
// text invisibly (ZWSP/ZWNJ/ZWJ and the BOM/ZWNBSP).
func isZeroWidth(r rune) bool {
	switch r {
	case 0x200B, 0x200C, 0x200D, 0xFEFF:
		return true
	}
	return false
}
