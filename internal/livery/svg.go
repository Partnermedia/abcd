package livery

import (
	"fmt"
	"strings"
)

// SVG rendering constants. The output must stay deterministic — fixed cell
// ordering, no timestamps, no floating point — because the drift gate
// byte-compares regenerated output against the committed files.
const (
	svgCell   = 20 // pixel cell edge, SVG units
	svgPad    = 30 // padding between the art and the panel edge
	svgCorner = 24 // panel corner radius

	// PanelColor is the dark panel behind the panel variants. Pixels keyed
	// 'k' are this color by design, so eyes read as holes in the panel.
	PanelColor = "#1c1f26"

	// TransparentEyeColor replaces 'k' on surfaces with no panel behind the
	// art — without it the face vanishes. Any consumer rendering the grids
	// onto its own background (itd-112's terminal work included) must make
	// the same substitution; Palette alone does not carry this rule.
	TransparentEyeColor = "#1f2328"
)

// RenderSVG renders an asset to a complete SVG document on its natural
// canvas (the grid's aspect ratio plus padding). The panel variant draws a
// rounded dark panel behind the art and is legible on any background; the
// transparent variant is for dark surfaces only, and its metadata says so.
func RenderSVG(a Asset, panel bool) []byte {
	return renderSVG(a, panel, false)
}

// RenderSVGSquare renders an asset on a square canvas — the avatar/icon
// format — with the art centered. The side is the larger of the natural
// dimensions, so no art is ever cropped or scaled.
func RenderSVGSquare(a Asset, panel bool) []byte {
	return renderSVG(a, panel, true)
}

func renderSVG(a Asset, panel, square bool) []byte {
	gw := 0
	if len(a.Grid) > 0 {
		gw = len(a.Grid[0])
	}
	gh := len(a.Grid)
	w := gw*svgCell + 2*svgPad
	h := gh*svgCell + 2*svgPad
	cw, ch := w, h
	if square {
		if cw > ch {
			ch = cw
		} else {
			cw = ch
		}
	}
	// Natural dimensions are always even (cell and padding are), so the
	// centering offsets stay integral and the output stays byte-stable.
	offX := (cw - w) / 2
	offY := (ch - h) / 2

	var b strings.Builder
	fmt.Fprintf(&b,
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">`+"\n",
		cw, ch, cw, ch)
	fmt.Fprintf(&b, "<title>abcd %s</title>\n", a.Name)
	if desc := assetDesc(a, panel, square); desc != "" {
		fmt.Fprintf(&b, "<desc>%s</desc>\n", desc)
	}
	if panel {
		fmt.Fprintf(&b, `<rect width="%d" height="%d" rx="%d" fill="%s"/>`+"\n",
			cw, ch, svgCorner, PanelColor)
	}
	// crispEdges applies to the pixel cells only: on the root it would
	// inherit to the rounded panel and staircase its corners.
	b.WriteString(`<g shape-rendering="crispEdges">` + "\n")
	for y, row := range a.Grid {
		for x, key := range row {
			if key == '.' {
				continue
			}
			fill := palette[key]
			if key == 'k' && !panel {
				fill = TransparentEyeColor
			}
			fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>`+"\n",
				offX+svgPad+x*svgCell, offY+svgPad+y*svgCell, svgCell, svgCell, fill)
		}
	}
	b.WriteString("</g>\n</svg>\n")
	return []byte(b.String())
}

func assetDesc(a Asset, panel, square bool) string {
	var parts []string
	if square {
		parts = append(parts, "Square canvas variant for avatar and icon use.")
	}
	if !panel {
		parts = append(parts, "Transparent variant: for dark surfaces only.")
	}
	if a.ApproximateGeometry {
		parts = append(parts, "Compact variant: flag geometry is approximate, not the ICS specification.")
	}
	return strings.Join(parts, " ")
}
