package site

import (
	"math"
	"testing"
)

// synthRecords builds a corpus large and varied enough to stress the packing:
// many dates, all four stores, and a wide spread of connectedness, so bubbles
// of very different sizes have to sit beside one another.
func synthRecords(n int) ([]LayoutNode, [][2]int, [][2]int) {
	types := []string{"adr", "intent", "spec", "issue"}
	nodes := make([]LayoutNode, n)
	for i := range nodes {
		day := 1 + i%28
		month := 1 + (i/28)%12
		nodes[i] = LayoutNode{
			Type: types[i%len(types)],
			Date: isoDate(2026, month, day),
			Num:  i + 1,
		}
	}
	var typed, mentions [][2]int
	// A hub, a chain, some pairs, and a tail of records with nothing at all.
	for i := 1; i < n/6; i++ {
		typed = append(typed, [2]int{0, i})
	}
	for i := n / 6; i+1 < n/3; i++ {
		typed = append(typed, [2]int{i, i + 1})
	}
	for i := n / 3; i+1 < n/2; i += 2 {
		typed = append(typed, [2]int{i, i + 1})
	}
	for i := n / 2; i+3 < n; i += 4 {
		mentions = append(mentions, [2]int{i, i + 3})
	}
	return nodes, typed, mentions
}

// TestCoilNeverOverlaps is the packing's own invariant, over a corpus big
// enough to make it work for it. A bubble sitting on another one means the
// placement walked into a case it does not handle, and the picture is wrong in
// a way no reader could be expected to spot.
func TestCoilNeverOverlaps(t *testing.T) {
	for _, n := range []int{1, 2, 7, 60, 400} {
		nodes, typed, mentions := synthRecords(n)
		a := ComputeArrangements(nodes, typed, mentions)
		if got := countOverlaps(a.Coil, a.Radius, a.CoilRadius); got != 0 {
			t.Errorf("n=%d: %d overlapping bubbles", n, got)
		}
		for i, p := range a.Coil {
			if math.Hypot(p.X, p.Y) > 1.0001 {
				t.Errorf("n=%d: coil position %d is outside the unit disk: %+v", n, i, p)
			}
			if math.IsNaN(p.X) || math.IsNaN(p.Y) {
				t.Fatalf("n=%d: coil position %d is not a number", n, i)
			}
		}
		for i, p := range a.Links {
			if math.IsNaN(p.X) || math.IsNaN(p.Y) {
				t.Fatalf("n=%d: link position %d is not a number", n, i)
			}
			if math.Hypot(p.X, p.Y) > 1.2 {
				t.Errorf("n=%d: link position %d is far outside the stage: %+v", n, i, p)
			}
		}
	}
}

// TestArrangementsAreAFunctionOfTheirInput is what lets the layout ride in a
// build artifact at all: the seed is fixed, so two runs place every bubble in
// the same place.
func TestArrangementsAreAFunctionOfTheirInput(t *testing.T) {
	nodes, typed, mentions := synthRecords(120)
	a := ComputeArrangements(nodes, typed, mentions)
	b := ComputeArrangements(nodes, typed, mentions)
	for i := range a.Coil {
		if a.Coil[i] != b.Coil[i] || a.Links[i] != b.Links[i] {
			t.Fatalf("node %d moved between two runs: %+v/%+v vs %+v/%+v",
				i, a.Coil[i], a.Links[i], b.Coil[i], b.Links[i])
		}
	}
	if a.CoilRadius != b.CoilRadius || a.Pairs != b.Pairs || a.Unlinked != b.Unlinked {
		t.Error("the arrangement's summary changed between two runs")
	}
}

// TestArrangementsShape pins the parts of the port a reader of build_data.py
// would check first: date order, month markers, and the three island shapes.
func TestArrangementsShape(t *testing.T) {
	nodes, typed, mentions := synthRecords(60)
	a := ComputeArrangements(nodes, typed, mentions)

	if len(a.Order) != len(nodes) {
		t.Fatalf("placement order covers %d of %d records", len(a.Order), len(nodes))
	}
	for i := 1; i < len(a.Order); i++ {
		prev, cur := nodes[a.Order[i-1]], nodes[a.Order[i]]
		if prev.Date > cur.Date {
			t.Fatalf("the coil is not in date order at %d: %s then %s", i, prev.Date, cur.Date)
		}
	}
	if len(a.Months) == 0 {
		t.Error("no month markers")
	}
	seen := map[string]bool{}
	for _, m := range a.Months {
		if seen[m.Month] {
			t.Errorf("month %s marked twice", m.Month)
		}
		seen[m.Month] = true
	}
	if len(a.Islands) == 0 {
		t.Error("no islands found in a corpus built with a hub and a chain")
	}
	for i := 1; i < len(a.Islands); i++ {
		if a.Islands[i-1] < a.Islands[i] {
			t.Errorf("islands are not largest-first: %v", a.Islands)
		}
	}
	if a.Unlinked == 0 {
		t.Error("no unlinked records found in a corpus with a tail of them")
	}
	// A record with no typed cross-reference is on the rim: outside every record
	// that has one, not in among them. The rim used to be two circles fixed in
	// advance, so a radius could be named; it is now wound outward from wherever
	// the linked records actually end, so what can be named is the relation.
	linked := map[int]bool{}
	for _, e := range typed {
		linked[e[0]], linked[e[1]] = true, true
	}
	core := 0.0
	for i := range nodes {
		if linked[i] {
			core = math.Max(core, math.Hypot(a.Links[i].X, a.Links[i].Y))
		}
	}
	if core == 0 {
		t.Fatal("no linked records in a corpus built with a hub and a chain")
	}
	for i := range nodes {
		if linked[i] {
			continue
		}
		if r := math.Hypot(a.Links[i].X, a.Links[i].Y); r < core {
			t.Errorf("record %d has no typed link but sits at radius %.3f, inside the linked records at %.3f", i, r, core)
		}
	}
}

// TestArrangementsPlaceDatelessRecordsLast pins what a record with no date
// does to a chronological picture. The empty string is the smallest string
// there is, so the naive answer seats an undatable record at the centre of the
// spiral — the position that means "the first thing that happened" — and
// publishes a span that begins nowhere.
func TestArrangementsPlaceDatelessRecordsLast(t *testing.T) {
	nodes, typed, mentions := synthRecords(40)
	// Two records git cannot place: one written but not yet committed is the
	// everyday way this happens.
	nodes[7].Date = ""
	nodes[23].Date = ""
	a := ComputeArrangements(nodes, typed, mentions)

	if got := nodes[a.Order[0]].Date; got == "" {
		t.Errorf("an undated record took the centre of the coil (order[0] = %d)", a.Order[0])
	}
	last := []int{a.Order[len(a.Order)-1], a.Order[len(a.Order)-2]}
	for _, i := range last {
		if nodes[i].Date != "" {
			t.Errorf("a dated record sorted after an undated one (index %d, date %q)", i, nodes[i].Date)
		}
	}
	if a.DateRange[0] == "" {
		t.Errorf("the published span begins at the empty string: %v", a.DateRange)
	}
	if a.DateRange[0] > a.DateRange[1] {
		t.Errorf("the span runs backwards: %v", a.DateRange)
	}
	for _, m := range a.Months {
		if m.Month == "" {
			t.Error("an empty month marker was published")
		}
	}
	if a.Overlaps != 0 {
		t.Errorf("undated records broke the packing: %d overlaps", a.Overlaps)
	}
}

// TestArrangementsEmptyCorpus keeps an unpopulated repository a state rather
// than a panic.
func TestArrangementsEmptyCorpus(t *testing.T) {
	a := ComputeArrangements(nil, nil, nil)
	if len(a.Coil) != 0 || a.Overlaps != 0 {
		t.Fatalf("empty corpus: %+v", a)
	}
}

// TestByLinksNeverOverlaps is the by-links arrangement's half of the same
// invariant TestCoilNeverOverlaps holds the coil to. The arrangement used to
// settle its islands with a spring layout that had no collision pass at all and
// to map each rim record's arc-width onto a circle smaller than the sum of those
// widths, so it published overlapping positions by construction and the chart
// never came to rest on it.
func TestByLinksNeverOverlaps(t *testing.T) {
	for _, n := range []int{1, 2, 7, 60, 400} {
		nodes, typed, mentions := synthRecords(n)
		a := ComputeArrangements(nodes, typed, mentions)
		if got := countOverlaps(a.Links, a.Radius, a.CoilRadius); got != 0 {
			t.Errorf("n=%d: the by-links arrangement publishes %d overlapping bubbles", n, got)
		}
	}
}

// TestOverlapGateMeasuresBothArrangements pins what the published count is a
// count OF. It was the coil's alone, so an overlapping by-links picture shipped
// green.
func TestOverlapGateMeasuresBothArrangements(t *testing.T) {
	nodes, typed, mentions := synthRecords(200)
	a := ComputeArrangements(nodes, typed, mentions)
	coil := countOverlaps(a.Coil, a.Radius, a.CoilRadius)
	links := countOverlaps(a.Links, a.Radius, a.CoilRadius)
	if a.Overlaps != coil+links {
		t.Errorf("published overlaps = %d, but coil %d + links %d = %d",
			a.Overlaps, coil, links, coil+links)
	}
}

// TestOverlapCountIsTakenInTheRenderersSpace is the trap that made an earlier
// measurement useless: positions are published normalised to the unit disk while
// radii are quoted in reference pixels, and comparing across the two spaces
// reported an overlap for very nearly every pair.
func TestOverlapCountIsTakenInTheRenderersSpace(t *testing.T) {
	// Two bubbles a quarter of the disk apart, each an eighth of it wide: they
	// clear one another at the scale the renderer draws them.
	points := []Point{{X: -0.25, Y: 0}, {X: 0.25, Y: 0}}
	radii := []float64{12, 12}
	if got := countOverlaps(points, radii, 100); got != 0 {
		t.Errorf("bubbles 50 units apart with radii 12 count as %d overlaps", got)
	}
	// The same pair drawn on a disk small enough that they do collide.
	if got := countOverlaps(points, radii, 40); got != 1 {
		t.Errorf("bubbles 20 units apart with radii 12 count as %d overlaps, want 1", got)
	}
}

// TestByLinksStaysOnTheStage guards the other half of the fix. The renderer
// fits the stage to coil_radius (site-src/record.js), so an arrangement that
// separates its bubbles by growing settles them off-screen: a measured earlier
// attempt pushed the outermost radius from 0.97 to 1.83, which is why it was
// reverted. Coming to rest has to happen inside the picture, not beside it.
// Over this corpus the settled arrangement reaches 1.01 at its widest, where a
// hub island and a chain island of near-equal size have to sit apart.
func TestByLinksStaysOnTheStage(t *testing.T) {
	for _, n := range []int{7, 60, 200, 400} {
		nodes, typed, mentions := synthRecords(n)
		a := ComputeArrangements(nodes, typed, mentions)
		for i, p := range a.Links {
			if r := math.Hypot(p.X, p.Y); r > 1.1 {
				t.Errorf("n=%d: link position %d settled at radius %.3f, off the stage", n, i, r)
			}
		}
	}
}

// TestOverlappingByLinksFlagsRed is the gate's own failure mode, proven rather
// than assumed. The by-links arrangement no longer publishes an overlap, so the
// only way to show the gate would catch one is to put one there.
func TestOverlappingByLinksFlagsRed(t *testing.T) {
	nodes, typed, mentions := synthRecords(60)
	a := ComputeArrangements(nodes, typed, mentions)
	if a.Overlaps != 0 {
		t.Fatalf("the fixture arrives with %d overlaps of its own", a.Overlaps)
	}
	// Two records in the same place: the fault the old rim published by the
	// hundred, because it mapped more arc-width onto a circle than it had.
	last := len(a.Links) - 1
	a.Links[last] = a.Links[last-1]
	if got := a.overlapCount(); got == 0 {
		t.Error("two bubbles at the same position in the by-links arrangement report no overlap")
	}
}
