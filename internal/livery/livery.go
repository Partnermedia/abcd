// Package livery holds the canonical pixel-grid definitions of abcd's visual
// identity — the duckling mascot, the signal-flag logo, and the lifeboat mark —
// and the palette they share (itd-133/spc-36). It is the single source of
// truth for every rendered artifact: the committed SVG assets under
// docs/assets/img/livery/ are generated from these grids and a package test
// regenerates and byte-compares them, so a hand edit to either side fails the
// gate. Terminal rendering (itd-112) consumes the same grids and palette keys;
// nothing in this package writes to a terminal.
//
// The package is named for a vessel's livery — its distinctive colors and
// markings — because "identity" already means two other things in this
// codebase (the git-author gate and the positioning block).
package livery

//go:generate go run ./gen

// An Asset is one identity artwork as a pixel grid. Grid rows are strings of
// palette keys, one rune per cell; '.' is transparent. Every row of a grid has
// the same width.
type Asset struct {
	Name string   // stable file stem for generated artifacts, e.g. "duckling"
	Grid []string // pixel rows, top to bottom

	// ApproximateGeometry marks a variant whose flag geometry is a spatial
	// approximation rather than the true ICS specification. Only the
	// full-size logo carries the true-geometry claim; the compact logo
	// cannot (five stripes do not fit three rows) and must say so.
	ApproximateGeometry bool
}

// Assets returns the identity artworks in stable order. The slice and its
// grids are copies; callers cannot mutate the canonical definitions.
func Assets() []Asset {
	out := make([]Asset, len(assets))
	for i, a := range assets {
		out[i] = Asset{
			Name:                a.Name,
			Grid:                append([]string(nil), a.Grid...),
			ApproximateGeometry: a.ApproximateGeometry,
		}
	}
	return out
}

// Palette returns the palette as a copy: pixel key to hex color. The same
// keys map to ANSI colors when a terminal surface renders these grids, which
// is what keeps the two pipelines on one source. One rule does not live in
// the map: 'k' is panel-colored, and a surface with no panel behind the art
// must substitute TransparentEyeColor for it or the faces vanish.
func Palette() map[rune]string {
	out := make(map[rune]string, len(palette))
	for k, v := range palette {
		out[k] = v
	}
	return out
}
