package livery

import (
	"strings"
	"testing"

	"github.com/Partnermedia/abcd/internal/term"
)

// TestAnsiTableParity holds every ANSI table's key set to the palette: a key
// added to one pipeline cannot silently miss the others.
func TestAnsiTableParity(t *testing.T) {
	for name, keys := range map[string][]rune{
		"ansi256":    mapKeys(ansi256),
		"ansi16":     mapKeys(ansi16),
		"monoGlyphs": mapKeysS(monoGlyphs),
	} {
		if len(keys) != len(palette) {
			t.Errorf("%s: %d keys, palette has %d", name, len(keys), len(palette))
		}
		for _, k := range keys {
			if _, ok := palette[k]; !ok {
				t.Errorf("%s: key %q not in palette", name, k)
			}
		}
	}
}

func mapKeys(m map[rune]int) []rune {
	out := make([]rune, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func mapKeysS(m map[rune]string) []rune {
	out := make([]rune, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func asset(t *testing.T, name string) Asset {
	t.Helper()
	for _, a := range Assets() {
		if a.Name == name {
			return a
		}
	}
	t.Fatalf("asset %s missing", name)
	return Asset{}
}

// TestRenderHalfBlockGeometry: the five-row flag strip renders as exactly
// three text rows (render-time padding), each 23 half-block glyphs wide, and
// the canonical grid is left untouched.
func TestRenderHalfBlockGeometry(t *testing.T) {
	logo := asset(t, "logo-flags")
	before := strings.Join(logo.Grid, "|")
	for _, mode := range []term.ColorMode{term.TrueColor, term.Ansi256, term.Ansi16} {
		lines := RenderHalfBlock(logo, mode)
		if len(lines) != 3 {
			t.Fatalf("mode %v: %d lines, want 3", mode, len(lines))
		}
		for i, l := range lines {
			if got := strings.Count(l, "▀"); got != 23 {
				t.Errorf("mode %v line %d: %d half-blocks, want 23", mode, i, got)
			}
			if !strings.HasSuffix(l, "\x1b[0m") {
				t.Errorf("mode %v line %d: missing reset", mode, i)
			}
		}
	}
	// Pin the fg=top/bg=bottom pairing on an asymmetric cell: the last text
	// row pairs grid row 4 with the padding row, so its first cell is white
	// over panel — swapping the pairing inverts the sequence.
	last := RenderHalfBlock(logo, term.TrueColor)[2]
	if !strings.HasPrefix(last, "\x1b[38;2;232;232;232;48;2;28;31;38m▀") {
		t.Errorf("half-block pairing is not fg=top/bg=bottom: %q", last[:40])
	}
	if strings.Join(logo.Grid, "|") != before {
		t.Fatalf("RenderHalfBlock mutated the grid it was given")
	}
	if canonical := strings.Join(asset(t, "logo-flags").Grid, "|"); canonical != before {
		t.Fatalf("canonical grid changed")
	}
}

// TestRenderShade: mono art is 5 rows, contains no escape bytes, and is not
// blank (adr-49: degraded output is never blank).
func TestRenderShade(t *testing.T) {
	lines := RenderShade(asset(t, "logo-flags"))
	if len(lines) != 5 {
		t.Fatalf("%d lines, want 5", len(lines))
	}
	blank := true
	for _, l := range lines {
		if strings.Contains(l, "\x1b") {
			t.Fatalf("mono line contains an escape byte: %q", l)
		}
		if strings.TrimSpace(l) != "" {
			blank = false
		}
	}
	if blank {
		t.Fatalf("mono render is blank")
	}
}

func TestSGRModes(t *testing.T) {
	if got := sgrFG('o', term.TrueColor); got != "38;2;217;119;87" {
		t.Errorf("truecolor fg for 'o': %s", got)
	}
	if got := sgrBG('b', term.Ansi256); got != "48;5;68" {
		t.Errorf("256 bg for 'b': %s", got)
	}
	if got := sgrFG('w', term.Ansi16); got != "97" {
		t.Errorf("16 fg for 'w' (bright white): %s", got)
	}
	if got := sgrBG('k', term.Ansi16); got != "40" {
		t.Errorf("16 bg for 'k' (black): %s", got)
	}
}
