package livery

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

// assetsDir is the committed home of the generated SVGs, relative to this
// package directory (the test working directory).
const assetsDir = "../../docs/assets/img/livery"

func TestGridsWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, a := range Assets() {
		if a.Name == "" {
			t.Fatalf("asset with empty name")
		}
		if seen[a.Name] {
			t.Fatalf("duplicate asset name %q", a.Name)
		}
		seen[a.Name] = true
		if len(a.Grid) == 0 {
			t.Fatalf("%s: empty grid", a.Name)
		}
		width := len(a.Grid[0])
		for i, row := range a.Grid {
			if len(row) != width {
				t.Errorf("%s: row %d has width %d, want %d", a.Name, i, len(row), width)
			}
			// The renderer and the width checks index by byte; a multi-byte
			// key would silently misplace every later cell in its row, so
			// the single-byte assumption is asserted, not implied.
			if len(row) != utf8.RuneCountInString(row) {
				t.Errorf("%s: row %d contains a multi-byte cell key: %q", a.Name, i, row)
			}
			for _, key := range row {
				if key == '.' {
					continue
				}
				if _, ok := palette[key]; !ok {
					t.Errorf("%s: row %d has unknown palette key %q", a.Name, i, key)
				}
			}
		}
	}
}

func TestAccessorsReturnCopies(t *testing.T) {
	a := Assets()
	a[0].Grid[0] = "XXXX"
	if Assets()[0].Grid[0] == "XXXX" {
		t.Fatalf("Assets() exposes the canonical grid to mutation")
	}
	p := Palette()
	p['o'] = "#000000"
	if Palette()['o'] == "#000000" {
		t.Fatalf("Palette() exposes the canonical palette to mutation")
	}
}

// flag extracts one 5x5 flag from a logo grid: five columns starting at col,
// five rows starting at rowStart.
func flag(t *testing.T, grid []string, rowStart, col int) []string {
	t.Helper()
	if rowStart+5 > len(grid) {
		t.Fatalf("flag at row %d overruns grid of %d rows", rowStart, len(grid))
	}
	out := make([]string, 5)
	for i, row := range grid[rowStart : rowStart+5] {
		if col+5 > len(row) {
			t.Fatalf("flag at col %d overruns row %d", col, rowStart+i)
		}
		out[i] = row[col : col+5]
	}
	return out
}

// assertSwallowtail checks the fly edge of a swallowtail flag: full rows at
// top and bottom, the notch deepest at the middle row, vertically symmetric.
func assertSwallowtail(t *testing.T, name string, rows []string) {
	t.Helper()
	n := len(rows)
	for i := 0; i < n/2; i++ {
		if rows[i] != rows[n-1-i] {
			t.Errorf("%s: not vertically symmetric: row %d %q vs row %d %q", name, i, rows[i], n-1-i, rows[n-1-i])
		}
	}
	notch := func(row string) int { return len(row) - len(strings.TrimRight(row, ".")) }
	for i := 1; i <= n/2; i++ {
		if notch(rows[i]) <= notch(rows[i-1]) {
			t.Errorf("%s: notch does not deepen toward the middle: row %d depth %d, row %d depth %d",
				name, i-1, notch(rows[i-1]), i, notch(rows[i]))
		}
	}
	if notch(rows[0]) != 0 {
		t.Errorf("%s: top row is notched: %q", name, rows[0])
	}
	// The notch is a cut into the fly, never the whole flag: every row keeps
	// a majority of its width. Without this bound a blanked middle row reads
	// as a very deep notch and passes.
	for i, row := range rows {
		if notch(row) > len(row)/2 {
			t.Errorf("%s: row %d notch depth %d swallows the flag: %q", name, i, notch(row), row)
		}
	}
}

// uniform returns the single palette key a row consists of, or 0.
func uniform(row string) rune {
	var k rune
	for _, r := range row {
		if k == 0 {
			k = r
		}
		if r != k {
			return 0
		}
	}
	return k
}

// TestFlagGeometry holds the full-size logo to the ICS specification: alfa
// vertically halved white/blue with a swallowtail, bravo all-red with a
// swallowtail, charlie's five horizontal stripes blue-white-red-white-blue,
// delta's three bands yellow-blue-yellow. The compact logo is exempt by
// design and must say so instead.
func TestFlagGeometry(t *testing.T) {
	byName := map[string]Asset{}
	for _, a := range Assets() {
		byName[a.Name] = a
	}
	full, okFull := byName["logo-flags"]
	icon, okIcon := byName["logo-flags-icon"]
	compact, okCompact := byName["logo-flags-compact"]
	if !okFull || !okIcon || !okCompact {
		t.Fatalf("logo assets missing")
	}
	if full.ApproximateGeometry || icon.ApproximateGeometry {
		t.Errorf("full-size and icon logos must carry the true-geometry claim")
	}
	if !compact.ApproximateGeometry {
		t.Errorf("compact logo must be declared approximate")
	}

	// The strip lays the flags out in one row; the icon two-by-two. Both
	// carry the same true geometry.
	layouts := []struct {
		name                        string
		grid                        []string
		alfa, bravo, charlie, delta [2]int // rowStart, col
	}{
		{"logo-flags", full.Grid, [2]int{0, 0}, [2]int{0, 6}, [2]int{0, 12}, [2]int{0, 18}},
		{"logo-flags-icon", icon.Grid, [2]int{0, 0}, [2]int{0, 6}, [2]int{6, 0}, [2]int{6, 6}},
	}
	for _, l := range layouts {
		alfa := flag(t, l.grid, l.alfa[0], l.alfa[1])
		for i, row := range alfa {
			if row[:3] != "www" {
				t.Errorf("%s alfa: row %d hoist half not white: %q", l.name, i, row)
			}
			if fly := strings.TrimRight(row[3:], "."); strings.Trim(fly, "b") != "" {
				t.Errorf("%s alfa: row %d fly half not blue: %q", l.name, i, row)
			}
		}
		assertSwallowtail(t, l.name+" alfa", alfa)

		bravo := flag(t, l.grid, l.bravo[0], l.bravo[1])
		for i, row := range bravo {
			if row[:3] != "rrr" {
				t.Errorf("%s bravo: row %d hoist not red: %q", l.name, i, row)
			}
			if body := strings.TrimRight(row, "."); strings.Trim(body, "r") != "" {
				t.Errorf("%s bravo: row %d not all red: %q", l.name, i, row)
			}
		}
		assertSwallowtail(t, l.name+" bravo", bravo)

		charlie := flag(t, l.grid, l.charlie[0], l.charlie[1])
		for i, want := range []rune{'b', 'w', 'r', 'w', 'b'} {
			if got := uniform(charlie[i]); got != want {
				t.Errorf("%s charlie: stripe %d is %q, want %q", l.name, i, got, want)
			}
		}

		delta := flag(t, l.grid, l.delta[0], l.delta[1])
		for i, want := range []rune{'y', 'b', 'b', 'b', 'y'} {
			if got := uniform(delta[i]); got != want {
				t.Errorf("%s delta: band %d is %q, want %q", l.name, i, got, want)
			}
		}
	}
}

// TestSVGAssetsInSync is the drift gate: the committed SVGs under
// docs/assets/img/livery/ must be byte-identical to what the grids render,
// and the directory must hold nothing else. Regenerate with
// `go generate ./internal/livery/...`.
func TestSVGAssetsInSync(t *testing.T) {
	want := map[string][]byte{}
	for _, a := range Assets() {
		want[a.Name+".svg"] = RenderSVG(a, true)
		want[a.Name+"-transparent.svg"] = RenderSVG(a, false)
		want[a.Name+"-square.svg"] = RenderSVGSquare(a, true)
		want[a.Name+"-square-transparent.svg"] = RenderSVGSquare(a, false)
	}

	entries, err := os.ReadDir(assetsDir)
	if err != nil {
		t.Fatalf("reading committed assets: %v (run `go generate ./internal/livery/...`)", err)
	}
	got := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() {
			t.Errorf("unexpected directory %s in %s", e.Name(), assetsDir)
			continue
		}
		got[e.Name()] = true
		wantBytes, ok := want[e.Name()]
		if !ok {
			t.Errorf("stale file %s in %s: no grid generates it — delete it (go generate never prunes)", e.Name(), assetsDir)
			continue
		}
		gotBytes, err := os.ReadFile(filepath.Join(assetsDir, e.Name()))
		if err != nil {
			t.Fatalf("reading %s: %v", e.Name(), err)
		}
		if string(gotBytes) != string(wantBytes) {
			t.Errorf("%s has drifted from its grid: run `go generate ./internal/livery/...`", e.Name())
		}
	}
	for name := range want {
		if !got[name] {
			t.Errorf("missing committed asset %s: run `go generate ./internal/livery/...`", name)
		}
	}
}

// TestRenderSVGDeterministic guards the property the drift gate rests on.
func TestRenderSVGDeterministic(t *testing.T) {
	renderers := map[string]func(Asset, bool) []byte{
		"natural": RenderSVG,
		"square":  RenderSVGSquare,
	}
	for _, a := range Assets() {
		for shape, render := range renderers {
			for _, panel := range []bool{true, false} {
				one := render(a, panel)
				two := render(a, panel)
				if string(one) != string(two) {
					t.Fatalf("%s (%s, panel=%v): non-deterministic render", a.Name, shape, panel)
				}
				if !strings.HasPrefix(string(one), "<svg ") {
					t.Fatalf("%s: output does not start with <svg", a.Name)
				}
			}
		}
	}
	var mini Asset
	for _, a := range Assets() {
		if a.Name == "duckling-mini" {
			mini = a
		}
	}
	if s := string(RenderSVG(mini, false)); !strings.Contains(s, "dark surfaces only") {
		t.Errorf("transparent variant lacks its dark-surface-only label: %s", firstLine(s))
	}
}

// TestSquareCanvas holds the avatar variants to their format: the canvas is
// square, no smaller than the natural canvas, and labelled as square.
func TestSquareCanvas(t *testing.T) {
	dims := func(svg []byte) (w, h string) {
		s := string(svg)
		w = attr(t, s, "width")
		h = attr(t, s, "height")
		return
	}
	for _, a := range Assets() {
		sq := RenderSVGSquare(a, true)
		w, h := dims(sq)
		if w != h {
			t.Errorf("%s-square: canvas %sx%s is not square", a.Name, w, h)
		}
		if !strings.Contains(string(sq), "Square canvas variant") {
			t.Errorf("%s-square: missing square-variant label", a.Name)
		}
		nw, nh := dims(RenderSVG(a, true))
		side, nwi, nhi := atoi(t, w), atoi(t, nw), atoi(t, nh)
		if side < nwi || side < nhi {
			t.Errorf("%s-square: side %d smaller than natural canvas %dx%d — art would be cropped", a.Name, side, nwi, nhi)
		}
		if side != nwi && side != nhi {
			t.Errorf("%s-square: side %d matches neither natural dimension %dx%d", a.Name, side, nwi, nhi)
		}
	}
}

func atoi(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(s)
	if err != nil {
		t.Fatalf("non-numeric dimension %q", s)
	}
	return n
}

// attr extracts the first occurrence of an SVG attribute value.
func attr(t *testing.T, s, name string) string {
	t.Helper()
	marker := name + `="`
	i := strings.Index(s, marker)
	if i < 0 {
		t.Fatalf("attribute %s not found", name)
	}
	rest := s[i+len(marker):]
	j := strings.IndexByte(rest, '"')
	if j < 0 {
		t.Fatalf("attribute %s not terminated", name)
	}
	return rest[:j]
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
