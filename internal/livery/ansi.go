package livery

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/intentdriven/abcd/internal/term"
)

// The terminal side of the livery: the same palette keys mapped onto the
// colour ladder's rungs, and the grid renderers for terminal art. The pinned
// 256- and 16-colour tables are hand-chosen equivalents of the hex palette —
// TestAnsiTableParity holds their key sets to the palette so a new pixel key
// cannot land in one pipeline and silently miss the other (the itd-133
// discipline, extended to ANSI by spc-41).

// ansi256 maps a palette key to its xterm-256 colour index.
var ansi256 = map[rune]int{
	'o': 173, 'k': 236, 'y': 221, 'w': 255, 'r': 167, 'b': 68, 'c': 31, 'm': 246,
}

// ansi16 maps a palette key to a 16-colour index (0-15).
var ansi16 = map[rune]int{
	'o': 1, 'k': 0, 'y': 11, 'w': 15, 'r': 9, 'b': 4, 'c': 6, 'm': 8,
}

// monoGlyphs maps a palette key to its shade-block pair for the mono rung.
// Panel-coloured pixels ('k') and transparent cells render as gaps.
var monoGlyphs = map[rune]string{
	'o': "▓▓", 'k': "  ", 'y': "▒▒", 'w': "░░", 'r': "▓▓", 'b': "██", 'c': "░░", 'm': "██",
}

// Reset is the SGR reset sequence.
const Reset = "\x1b[0m"

// FG returns the complete escape sequence selecting a palette key's
// foreground colour at the given ladder rung. Mono callers must not colour
// at all — they have no business calling this.
func FG(key rune, mode term.ColorMode) string {
	return "\x1b[" + sgrFG(key, mode) + "m"
}

// hexRGB decodes a #rrggbb palette value.
func hexRGB(hex string) (r, g, b int) {
	v, _ := strconv.ParseInt(strings.TrimPrefix(hex, "#"), 16, 32)
	return int(v >> 16 & 0xff), int(v >> 8 & 0xff), int(v & 0xff)
}

// sgrFG and sgrBG compose the colour parameters for one palette key at one
// ladder rung. Mono never reaches here.
func sgrFG(key rune, mode term.ColorMode) string {
	switch mode {
	case term.TrueColor:
		r, g, b := hexRGB(palette[key])
		return fmt.Sprintf("38;2;%d;%d;%d", r, g, b)
	case term.Ansi256:
		return fmt.Sprintf("38;5;%d", ansi256[key])
	default:
		n := ansi16[key]
		if n >= 8 {
			return strconv.Itoa(90 + n - 8)
		}
		return strconv.Itoa(30 + n)
	}
}

func sgrBG(key rune, mode term.ColorMode) string {
	switch mode {
	case term.TrueColor:
		r, g, b := hexRGB(palette[key])
		return fmt.Sprintf("48;2;%d;%d;%d", r, g, b)
	case term.Ansi256:
		return fmt.Sprintf("48;5;%d", ansi256[key])
	default:
		n := ansi16[key]
		if n >= 8 {
			return strconv.Itoa(100 + n - 8)
		}
		return strconv.Itoa(40 + n)
	}
}

// RenderHalfBlock renders an asset as half-block terminal art: two pixel rows
// per text row, every cell painted (a transparent cell takes the panel
// colour — the terminal analogue of the SVG panel, which also keeps white
// pixels legible on light themes). The grid is padded to an even row count
// locally; the canonical grid is never mutated. Mono is not a valid mode
// here — use RenderShade.
func RenderHalfBlock(a Asset, mode term.ColorMode) []string {
	grid := a.Grid
	if len(grid)%2 == 1 && len(grid) > 0 {
		grid = append(append([]string(nil), grid...), strings.Repeat(".", len(grid[0])))
	}
	solid := func(key rune) rune {
		if key == '.' {
			return 'k' // panel colour behind the art
		}
		return key
	}
	var out []string
	for y := 0; y+1 < len(grid); y += 2 {
		top, bottom := grid[y], grid[y+1]
		var b strings.Builder
		for x, tk := range top {
			bk := rune(bottom[x])
			b.WriteString("\x1b[" + sgrFG(solid(tk), mode) + ";" + sgrBG(solid(bk), mode) + "m▀")
		}
		b.WriteString("\x1b[0m")
		out = append(out, b.String())
	}
	return out
}

// RenderShade renders an asset as the mono shade-block form: one glyph pair
// per pixel, no colour, never blank. This is the art the mono rung shows.
func RenderShade(a Asset) []string {
	out := make([]string, len(a.Grid))
	for i, row := range a.Grid {
		var b strings.Builder
		for _, key := range row {
			if key == '.' {
				b.WriteString("  ")
				continue
			}
			b.WriteString(monoGlyphs[key])
		}
		out[i] = strings.TrimRight(b.String(), " ")
	}
	return out
}
