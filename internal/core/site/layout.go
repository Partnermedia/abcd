package site

// The two precomputed chart arrangements, ported from `build_data.py` in
// `.abcd/development/research/abcdev-site/` — the executable spec.
//
// Both are computed HERE, at build time, and ride in record.json as plain
// numbers: the published pages make no API calls and run no layout engine
// (adr-38, adr-48). Determinism is therefore not a nicety but the contract —
// the same tree must produce the same picture, and CI proves it by building
// twice and diffing.
//
// The script leans on numpy and networkx; Go has neither, so the maths is
// written out. Matching Python's exact float output is explicitly NOT the goal
// (no two float pipelines agree to the last bit anyway); matching its SHAPE and
// being reproducible are, so this port is pinned by a golden test of its own
// output rather than by comparison with the script's.

import (
	"math"
	"math/rand"
	"sort"
)

// The coil's constants, all in the renderer's reference pixels.
const (
	// refScale converts the renderer's desktop bubble radius to the reference
	// scale the coil is packed at.
	refScale = 0.86
	// coilGap is the breathing room between neighbours on the coil.
	coilGap = 4.5
	// inwardDip is how far back toward the centre a bubble may fall relative to
	// its predecessor. A larger dip lets the sequence double back and stop
	// reading as a path; zero would force a strict spiral with visible gaps.
	inwardDip = 0.35
	// layoutSeed is the fixed seed for every random draw in this file. The
	// arrangement is a function of the record, and the seed is what keeps the
	// randomness from making it a function of the day as well.
	layoutSeed = 11
	// springIterations, springK and springWeight are the force layout's
	// parameters as the script passes them.
	springIterations = 300
	springThreshold  = 1e-4
	springKFactor    = 1.6
	springWeight     = 3.0
	// decompressPercentile is the radius the dense middle of an island is
	// normalised against before the square-root easing spreads it out.
	decompressPercentile = 97.0
	// islandScale is the largest island's radius; otherIslandScale the rest.
	islandScale      = 0.72
	otherIslandScale = 0.09
	// ringRadius is where pairs and the smaller islands sit; rimRow0 and rimRow1
	// are the two rows of records that carry no typed cross-reference at all.
	ringRadius = 0.80
	rimRow0    = 0.905
	rimRow1    = 0.97
	// pairOffset is half the gap between the two members of a pair.
	pairOffset = 0.012
	// mentionWeight is what a body mention contributes to a bubble's size,
	// relative to a typed link. A mention is evidence of relevance, not of a
	// declared relationship, and the size difference says so.
	mentionWeight = 0.35
)

// typeOrder is the tie-break between records that share a date: the order the
// record itself reads in, decisions first.
var typeOrder = map[string]int{
	"adr": 0, "principle": 1, "intent": 2, "spec": 3, "issue": 4, "rfc": 5, "phase": 6,
}

const typeOrderDefault = 9

// LayoutNode is what both arrangements need to know about one record.
type LayoutNode struct {
	// Type is the record's store name; it sets the bubble's size class and the
	// same-day tie-break.
	Type string
	// Date is the effective date the record is placed by.
	Date string
	// Num is the id number, the last tie-break.
	Num int
	// Touched is the day the record's file last changed. It places nothing — the
	// arrangements are chronological by when a record was WRITTEN — but it is
	// what the span a timeline draws has to reach.
	Touched string
}

// MonthMark is the first record of a month, in placement order.
type MonthMark struct {
	Month string `json:"month"`
	Node  int    `json:"node"`
}

// Point is one placed bubble centre, in the unit disk.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Arrangements holds both precomputed layouts plus the numbers that describe
// them, one entry per node in the caller's own order.
type Arrangements struct {
	// Coil places every record in date order, winding outward.
	Coil []Point `json:"coil"`
	// Links places connected work in islands and everything else on the rim.
	Links []Point `json:"links"`
	// Degree is each record's weighted connectedness; Radius its bubble size in
	// reference pixels.
	Degree []float64 `json:"degree"`
	Radius []float64 `json:"radius"`
	// Order is the coil's placement order.
	Order []int `json:"order"`
	// Months marks where each month begins along the coil.
	Months []MonthMark `json:"months"`
	// Overlaps is the packing's own sanity check: the number of bubble pairs
	// that intersect. It must be zero, and it is published so a reader can see
	// that it is rather than take the claim on trust.
	Overlaps int `json:"overlaps"`
	// CoilRadius is the packed radius before normalisation to the unit disk.
	CoilRadius float64 `json:"coil_radius"`
	// RefScale is the scale the radii are quoted at.
	RefScale float64 `json:"ref_scale"`
	// Isolated counts records with no link of any kind.
	Isolated int `json:"isolated"`
	// Islands are the sizes of the connected groups of three or more; Pairs and
	// Unlinked the counts of the two smaller shapes.
	Islands  []int `json:"islands"`
	Pairs    int   `json:"pairs"`
	Unlinked int   `json:"unlinked"`
	// DateRange is the first and last effective date placed — when the records
	// were WRITTEN.
	DateRange [2]string `json:"date_range"`
	// Span is DateRange extended to the last day any placed record's file was
	// touched, which is the axis a timeline draws: a decision written in May and
	// amended in August occupies both ends of it, and a range that stopped at the
	// writing dates would cut the chart short of its own newest fact.
	Span [2]string `json:"span"`
}

// ComputeArrangements places every record twice: once in date order along the
// coil, once by its typed links. typed and mentions are index pairs into nodes;
// mentions are excluded from the link arrangement (with them it is a hairball)
// but do count toward a bubble's size.
func ComputeArrangements(nodes []LayoutNode, typed, mentions [][2]int) Arrangements {
	n := len(nodes)
	a := Arrangements{
		Coil:     make([]Point, n),
		Links:    make([]Point, n),
		Degree:   make([]float64, n),
		Radius:   make([]float64, n),
		RefScale: refScale,
		Islands:  []int{},
		Months:   []MonthMark{},
		Order:    []int{},
	}
	if n == 0 {
		return a
	}

	for _, e := range typed {
		a.Degree[e[0]]++
		a.Degree[e[1]]++
	}
	for _, e := range mentions {
		a.Degree[e[0]] += mentionWeight
		a.Degree[e[1]] += mentionWeight
	}
	for i, nd := range nodes {
		base, step := 4.5, 3.2
		if nd.Type == "issue" {
			base, step = 3.0, 1.6
		}
		// The cap is what keeps a handful of hubs from swallowing the chart, and
		// the square root already compresses hard: at 4 it flattened everything
		// past sixteen links into one size, so the record's true hubs read the
		// same as a record with a middling few. At 7 they keep growing to about
		// fifty links, which is where this record's connectedness actually runs
		// out, and the hubs are visible as hubs.
		a.Radius[i] = (base + math.Min(math.Sqrt(a.Degree[i]), 7)*step) * refScale
		if a.Degree[i] == 0 {
			a.Isolated++
		}
	}

	a.Order = placementOrder(nodes)
	// A record can carry no date at all — one written but not yet committed, or
	// one git cannot place. An empty string is the smallest string there is, so
	// a plain minimum would publish a span starting at "" and would seat that
	// record at the centre of a chronological spiral, which is the position that
	// means "first". Dateless records are excluded from the span and placed last.
	for _, i := range a.Order {
		if nodes[i].Date == "" {
			continue
		}
		if a.DateRange[0] == "" || nodes[i].Date < a.DateRange[0] {
			a.DateRange[0] = nodes[i].Date
		}
		if nodes[i].Date > a.DateRange[1] {
			a.DateRange[1] = nodes[i].Date
		}
	}
	a.Span = a.DateRange
	for _, nd := range nodes {
		if nd.Touched > a.Span[1] {
			a.Span[1] = nd.Touched
		}
	}

	a.coil(nodes)
	a.byLinks(nodes, typed)
	a.round()
	return a
}

// round publishes the arrangement at the precision the chart draws it, as the
// script does. It is not cosmetic: the last bits of a transcendental are the one
// place two correct builds on two machines can differ, and this build's promise
// is that the same tree renders the same picture wherever it is rendered.
func (a *Arrangements) round() {
	r4 := func(v float64) float64 { return math.Round(v*1e4) / 1e4 }
	for i := range a.Coil {
		a.Coil[i] = Point{X: r4(a.Coil[i].X), Y: r4(a.Coil[i].Y)}
		a.Links[i] = Point{X: r4(a.Links[i].X), Y: r4(a.Links[i].Y)}
		a.Degree[i] = math.Round(a.Degree[i]*10) / 10
		a.Radius[i] = r4(a.Radius[i])
	}
	a.CoilRadius = math.Round(a.CoilRadius*10) / 10
}

// placementOrder sorts records by effective date, then by store, then by id.
func placementOrder(nodes []LayoutNode) []int {
	order := make([]int, len(nodes))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(x, y int) bool {
		a, b := nodes[order[x]], nodes[order[y]]
		// A record with no date is placed after every dated one rather than
		// before all of them: "we cannot date this" is not "this came first".
		if (a.Date == "") != (b.Date == "") {
			return b.Date == ""
		}
		if a.Date != b.Date {
			return a.Date < b.Date
		}
		if ta, tb := typeRank(a.Type), typeRank(b.Type); ta != tb {
			return ta < tb
		}
		return a.Num < b.Num
	})
	return order
}

func typeRank(t string) int {
	if r, ok := typeOrder[t]; ok {
		return r
	}
	return typeOrderDefault
}

// coil places records one by one in date order. The first sits at the centre;
// each next goes beside the previous one — a little further round, as close to
// the centre as the bubbles already placed allow — so the sequence winds outward
// like a snail shell. No simulation, so the picture is the same on every build.
func (a *Arrangements) coil(nodes []LayoutNode) {
	n := len(nodes)
	pos := make([]Point, n)
	pr := make([]float64, n)
	placed := make([]int, 0, n)
	phi := 0.0

	for k, i := range a.Order {
		r := a.Radius[i]
		if k == 0 {
			pos[i] = Point{}
			pr[i] = r
			placed = append(placed, i)
			continue
		}
		prev := placed[len(placed)-1]
		prho := math.Hypot(pos[prev].X, pos[prev].Y)
		need := pr[prev] + r + coilGap
		// The ray must clear the previous bubble: at the centre any direction
		// does, so the first step out takes a quarter turn.
		clear := math.Pi / 2
		if prho >= need {
			clear = math.Asin(need / prho)
		}
		phi += clear
		ux, uy := math.Cos(phi), math.Sin(phi)
		rho := math.Max(0, prho-inwardDip*need)

		// Every placed bubble forbids an interval of rho along the ray. Walk
		// them in order, jumping past each one the current rho falls inside,
		// until a pass changes nothing — a later interval can push rho back into
		// an earlier one, so a single pass is not enough.
		type interval struct{ lo, hi float64 }
		var forbidden []interval
		for _, j := range placed {
			rr := pr[j] + r + coilGap
			proj := pos[j].X*ux + pos[j].Y*uy
			perp2 := pos[j].X*pos[j].X + pos[j].Y*pos[j].Y - proj*proj
			if perp2 >= rr*rr {
				continue
			}
			half := math.Sqrt(math.Max(rr*rr-perp2, 0))
			forbidden = append(forbidden, interval{proj - half, proj + half})
		}
		sort.SliceStable(forbidden, func(x, y int) bool { return forbidden[x].lo < forbidden[y].lo })
		for changed := true; changed; {
			changed = false
			for _, iv := range forbidden {
				if iv.lo <= rho && rho < iv.hi {
					rho = iv.hi
					changed = true
				}
			}
		}

		pos[i] = Point{X: rho * ux, Y: rho * uy}
		pr[i] = r
		placed = append(placed, i)
	}

	rhoMax := 0.0
	for i := 0; i < n; i++ {
		if d := math.Hypot(pos[i].X, pos[i].Y) + pr[i]; d > rhoMax {
			rhoMax = d
		}
	}
	a.CoilRadius = rhoMax

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			d := math.Hypot(pos[i].X-pos[j].X, pos[i].Y-pos[j].Y)
			if d < pr[i]+pr[j]-0.01 {
				a.Overlaps++
			}
		}
	}

	seen := map[string]bool{}
	for _, i := range a.Order {
		m := nodes[i].Date
		if len(m) > 7 {
			m = m[:7]
		}
		if m == "" || seen[m] {
			continue
		}
		seen[m] = true
		a.Months = append(a.Months, MonthMark{Month: m, Node: i})
	}

	if rhoMax == 0 {
		rhoMax = 1
	}
	for i := 0; i < n; i++ {
		a.Coil[i] = Point{X: pos[i].X / rhoMax, Y: pos[i].Y / rhoMax}
	}
}

// byLinks arranges records by their typed links alone: connected work forms
// islands, pairs ring the main island, and records with no typed
// cross-reference sit on the rim in date order — so a bubble travels a short,
// readable path when the viewer switches arrangement.
func (a *Arrangements) byLinks(nodes []LayoutNode, typed [][2]int) {
	n := len(nodes)
	adj := make([][]int, n)
	deg := make([]int, n)
	seenEdge := map[[2]int]bool{}
	for _, e := range typed {
		s, t := e[0], e[1]
		if s == t {
			continue
		}
		key := [2]int{min(s, t), max(s, t)}
		if seenEdge[key] {
			continue
		}
		seenEdge[key] = true
		adj[s] = append(adj[s], t)
		adj[t] = append(adj[t], s)
		deg[s]++
		deg[t]++
	}

	comps := components(adj)
	var big [][]int
	var pairs [][]int
	var iso []int
	for _, c := range comps {
		switch {
		case len(c) >= 3:
			big = append(big, c)
		case len(c) == 2:
			pairs = append(pairs, c)
		default:
			iso = append(iso, c[0])
		}
	}
	for _, c := range big {
		a.Islands = append(a.Islands, len(c))
	}
	a.Pairs = len(pairs)
	a.Unlinked = len(iso)

	rng := rand.New(rand.NewSource(layoutSeed))
	for j, c := range big {
		cx, cy := 0.0, 0.0
		if j > 0 {
			ang := 2*math.Pi*float64(j-1)/math.Max(float64(len(big)-1), 1) - math.Pi/2 + 0.8
			cx, cy = math.Cos(ang)*ringRadius, math.Sin(ang)*ringRadius
		}
		scale := islandScale
		if j > 0 {
			scale = otherIslandScale
		}
		p := springLayout(c, adj, deg, rng)
		decompress(p)
		for q, i := range c {
			a.Links[i] = Point{X: cx + p[q].X*scale, Y: cy + p[q].Y*scale}
		}
	}

	for j, c := range pairs {
		ang := 2*math.Pi*float64(j)/math.Max(float64(len(pairs)), 1) - math.Pi/2 + 0.2
		cx, cy := math.Cos(ang)*ringRadius, math.Sin(ang)*ringRadius
		a.Links[c[0]] = Point{X: cx - math.Sin(ang)*pairOffset, Y: cy + math.Cos(ang)*pairOffset}
		a.Links[c[1]] = Point{X: cx + math.Sin(ang)*pairOffset, Y: cy - math.Cos(ang)*pairOffset}
	}

	sort.SliceStable(iso, func(x, y int) bool {
		p, q := nodes[iso[x]], nodes[iso[y]]
		if (p.Date == "") != (q.Date == "") {
			return q.Date == ""
		}
		if p.Date != q.Date {
			return p.Date < q.Date
		}
		if tp, tq := typeRank(p.Type), typeRank(q.Type); tp != tq {
			return tp < tq
		}
		return p.Num < q.Num
	})
	rows := [2][]int{}
	for j, i := range iso {
		rows[j%2] = append(rows[j%2], i)
	}
	for row, members := range rows {
		rr := rimRow0
		if row == 1 {
			rr = rimRow1
		}
		total := 0.0
		for _, i := range members {
			total += 2*a.Radius[i] + coilGap
		}
		if total == 0 {
			continue
		}
		run := 0.0
		for _, i := range members {
			w := 2*a.Radius[i] + coilGap
			ang := 2*math.Pi*(run+w/2)/total - math.Pi/2
			a.Links[i] = Point{X: math.Cos(ang) * rr, Y: math.Sin(ang) * rr}
			run += w
		}
	}
}

// components returns the connected components of adj, largest first and then by
// lowest member, so the biggest island is always the one placed in the middle.
func components(adj [][]int) [][]int {
	seen := make([]bool, len(adj))
	var out [][]int
	for i := range adj {
		if seen[i] {
			continue
		}
		var stack []int
		stack = append(stack, i)
		seen[i] = true
		var c []int
		for len(stack) > 0 {
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			c = append(c, v)
			for _, w := range adj[v] {
				if !seen[w] {
					seen[w] = true
					stack = append(stack, w)
				}
			}
		}
		sort.Ints(c)
		out = append(out, c)
	}
	sort.SliceStable(out, func(x, y int) bool {
		if len(out[x]) != len(out[y]) {
			return len(out[x]) > len(out[y])
		}
		return out[x][0] < out[y][0]
	})
	return out
}

// springLayout is the Fruchterman-Reingold force layout over one island, with
// the script's degree weighting: an edge between two well-connected records
// pulls less hard than one between two lonely ones, so a hub does not drag the
// whole island into itself.
func springLayout(members []int, adj [][]int, deg []int, rng *rand.Rand) []Point {
	n := len(members)
	idx := make(map[int]int, n)
	for q, i := range members {
		idx[i] = q
	}
	w := make([][]float64, n)
	for q := range w {
		w[q] = make([]float64, n)
	}
	for q, i := range members {
		for _, j := range adj[i] {
			r, ok := idx[j]
			if !ok {
				continue
			}
			w[q][r] = 1.0 / math.Sqrt(float64(max(deg[i], 1))*float64(max(deg[j], 1))) * springWeight
		}
	}

	pos := make([]Point, n)
	for q := range pos {
		pos[q] = Point{X: rng.NormFloat64() * 0.3, Y: rng.NormFloat64() * 0.3}
	}

	k := springKFactor / math.Sqrt(float64(n))
	// The initial step is a tenth of the layout's own extent, decayed linearly
	// to nothing over the run: big rearrangements early, fine settling late.
	minX, maxX := pos[0].X, pos[0].X
	minY, maxY := pos[0].Y, pos[0].Y
	for _, p := range pos {
		minX, maxX = math.Min(minX, p.X), math.Max(maxX, p.X)
		minY, maxY = math.Min(minY, p.Y), math.Max(maxY, p.Y)
	}
	t := math.Max(maxX-minX, maxY-minY) * 0.1
	dt := t / float64(springIterations+1)

	disp := make([]Point, n)
	for it := 0; it < springIterations; it++ {
		for q := range disp {
			disp[q] = Point{}
		}
		for q := 0; q < n; q++ {
			for r := 0; r < n; r++ {
				if q == r {
					continue
				}
				dx, dy := pos[q].X-pos[r].X, pos[q].Y-pos[r].Y
				d := math.Hypot(dx, dy)
				if d < 0.01 {
					d = 0.01
				}
				// Repulsion from every node, attraction along weighted edges.
				f := k*k/(d*d) - w[q][r]*d/k
				disp[q].X += dx * f
				disp[q].Y += dy * f
			}
		}
		moved := 0.0
		for q := 0; q < n; q++ {
			l := math.Hypot(disp[q].X, disp[q].Y)
			if l < 0.01 {
				l = 0.1
			}
			ddx, ddy := disp[q].X*t/l, disp[q].Y*t/l
			pos[q].X += ddx
			pos[q].Y += ddy
			moved += ddx*ddx + ddy*ddy
		}
		t -= dt
		if math.Sqrt(moved)/float64(n) < springThreshold {
			break
		}
	}
	rescale(pos)
	return pos
}

// rescale centres a layout and scales it so its widest axis spans the unit
// interval, matching the script's `scale=1.0`.
func rescale(pos []Point) {
	if len(pos) == 0 {
		return
	}
	var mx, my float64
	for _, p := range pos {
		mx += p.X
		my += p.Y
	}
	mx /= float64(len(pos))
	my /= float64(len(pos))
	lim := 0.0
	for q := range pos {
		pos[q].X -= mx
		pos[q].Y -= my
		lim = math.Max(lim, math.Max(math.Abs(pos[q].X), math.Abs(pos[q].Y)))
	}
	if lim <= 0 {
		return
	}
	for q := range pos {
		pos[q].X /= lim
		pos[q].Y /= lim
	}
}

// decompress spreads an island's dense middle: radii are normalised against the
// 97th percentile (so a lone outlier does not shrink everything else), clamped,
// and eased with a square root, which pushes the crowded centre outward while
// leaving the rim where it is.
func decompress(pos []Point) {
	if len(pos) == 0 {
		return
	}
	var mx, my float64
	for _, p := range pos {
		mx += p.X
		my += p.Y
	}
	mx /= float64(len(pos))
	my /= float64(len(pos))

	rad := make([]float64, len(pos))
	ang := make([]float64, len(pos))
	for q, p := range pos {
		dx, dy := p.X-mx, p.Y-my
		rad[q] = math.Hypot(dx, dy)
		ang[q] = math.Atan2(dy, dx)
	}
	ref := percentile(rad, decompressPercentile)
	if ref <= 0 {
		ref = 1
	}
	for q := range pos {
		r := math.Sqrt(math.Min(rad[q]/ref, 1.0))
		pos[q] = Point{X: math.Cos(ang[q]) * r, Y: math.Sin(ang[q]) * r}
	}
}

// percentile is the linear-interpolation percentile of a sample.
func percentile(xs []float64, p float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	if len(s) == 1 {
		return s[0]
	}
	pos := p / 100 * float64(len(s)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	if lo == hi {
		return s[lo]
	}
	return s[lo] + (s[hi]-s[lo])*(pos-float64(lo))
}
