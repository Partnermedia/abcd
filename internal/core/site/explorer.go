package site

// The record explorer — every page under `/record/`, plus `/contributors/` and
// `/references/`.
//
// Ported from the clickable prototype in the investigation cluster, which is the
// behavioural spec: the sub-navigation, the dashboard's panels, the chart's
// stage and controls, the genealogy and the two attribution pages are its
// markup, with the hash router's routes replaced by real directories.
//
// Every page here is GENERIC-SIDE (itd-140): its inputs are the record format,
// git history and `CHANGELOG.md` — plus, for `/references/`, the CSL
// bibliography and `ACKNOWLEDGEMENTS.md`. An absent optional input omits the
// page it feeds AND that page's navigation entry, and the build succeeds.
//
// Every visible word is a count, a date, an id, a title, a file name, a span of
// a record file carrying `data-src`, or an interface label from
// `site-src/ui.json` (adr-47 decision 2). The verbatim record rendering under
// `/record/**` is the one part of the site exempt from the banned-token gate
// (adr-47 decision 3): the record legitimately contains change-narration, and
// rewriting it to look better on the site is what that decision forbids.

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/Partnermedia/abcd/internal/fsutil"
)

// The explorer's routes. They are directories with an index.html, so the served
// URL is the route itself.
const (
	routeDashboard    = "record/"
	routeGraph        = "record/graph/"
	routeTimeline     = "record/timeline/"
	routeFoundations  = "record/foundations/"
	routeContributors = "contributors/"
	routeReferences   = "references/"
)

// lifecycleRank orders a store's buckets from unstarted to retired, so a
// lifecycle bar reads left to right as the work moves and takes the same colour
// for the same meaning in every store.
var lifecycleRank = map[string]int{
	"drafts": 0, "proposed": 0, "open": 0,
	"planned": 1,
	// A discipline is a rule that holds rather than a change that ships, so it
	// sits outside the ramp entirely.
	disciplinesLifecycle: 2,
	"shipped":            3, "accepted": 3, "closed": 3, "resolved": 3, "active": 3,
	"superseded": 4, "wontfix": 4,
}

// lifecycleRamp is the colour each rank takes.
var lifecycleRamp = []string{"var(--seq-1)", "var(--seq-2)", "var(--ink-3)", "var(--seq-3)", "var(--rule-2)"}

// explorer holds everything the explorer's pages are rendered from.
type explorer struct {
	c      *composer
	export RecordExport
	// byID and byPath index the export's nodes the two ways the pages ask for
	// one: from a link, and from a relative path in a record's own prose.
	byID   map[string]ExportNode
	byPath map[string]ExportNode
	// out and in are the typed links, phrased from each end.
	out map[string][]ExportEdge
	in  map[string][]ExportEdge
	// mentions are the undirected body references, per record.
	mentions map[string][]string
	// stubs are the outbound references whose target no file answers to. They
	// render as dashed stubs — the ruling is that a retired id is shown as
	// absent, never as a link to nothing and never as an invented position.
	stubs map[string][]ExportEdge
	// principles and disciplines are the foundations page's two card decks.
	principles  []ExportNode
	disciplines []ExportNode
	// eyebrow is the record root's own heading, with its provenance.
	eyebrow, eyebrowSrc string
	// bib is the bibliography, or nil where the repository keeps none.
	bib *Bibliography
}

// newExplorer indexes the export for the pages.
func newExplorer(c *composer, export RecordExport, bib *Bibliography, recordRoot string) *explorer {
	e := &explorer{
		c: c, export: export, bib: bib,
		byID:     make(map[string]ExportNode, len(export.Nodes)),
		byPath:   make(map[string]ExportNode, len(export.Nodes)),
		out:      map[string][]ExportEdge{},
		in:       map[string][]ExportEdge{},
		mentions: map[string][]string{},
		stubs:    map[string][]ExportEdge{},
	}
	for _, n := range export.Nodes {
		e.byID[n.ID] = n
		e.byPath[n.Path] = n
		switch {
		case n.Type == "principle":
			e.principles = append(e.principles, n)
		case n.Lifecycle == disciplinesLifecycle:
			e.disciplines = append(e.disciplines, n)
		}
	}
	for _, ed := range export.Edges {
		e.out[ed.From] = append(e.out[ed.From], ed)
		e.in[ed.To] = append(e.in[ed.To], ed)
	}
	for _, m := range export.Mentions {
		e.mentions[m.From] = append(e.mentions[m.From], m.To)
		e.mentions[m.To] = append(e.mentions[m.To], m.From)
	}
	for _, d := range c.graph.Dangling {
		if _, ok := e.byID[d.From]; !ok {
			continue
		}
		rel := d.Field
		if m, ok := relationOf[d.Field]; ok {
			rel = m.rel
		}
		e.stubs[d.From] = append(e.stubs[d.From], ExportEdge{From: d.From, To: d.To, Rel: rel})
	}
	if recordRoot != "" {
		rel := recordRoot + "/README.md"
		if data, err := fsutil.ReadGuarded(joinRepo(c.repoRoot, rel), maxPageBytes); err == nil {
			if h := firstHeading(rel, string(data), ""); h != "" {
				e.eyebrow, e.eyebrowSrc = h, srcAttr(rel, "")
			}
		}
	}
	return e
}

// hasFoundations reports whether the repository declares anything to found the
// work on. Without either store the page and its navigation entry are omitted.
func (e *explorer) hasFoundations() bool {
	return len(e.principles) > 0 || len(e.disciplines) > 0
}

// hasReferences reports whether the bibliography rendered.
func (e *explorer) hasReferences() bool { return e.bib != nil && len(e.bib.Entries) > 0 }

// Pages renders every explorer page, keyed by its output path.
func (e *explorer) Pages() (map[string]string, error) {
	pages := map[string]string{}
	add := func(route string, render func() (string, error)) error {
		html, err := render()
		if err != nil {
			return err
		}
		pages[route+"index.html"] = html
		return nil
	}
	if err := add(routeDashboard, e.dashboard); err != nil {
		return nil, err
	}
	if err := add(routeGraph, e.graphPage); err != nil {
		return nil, err
	}
	if err := add(routeTimeline, e.timelinePage); err != nil {
		return nil, err
	}
	if err := add(routeContributors, e.contributorsPage); err != nil {
		return nil, err
	}
	if e.hasFoundations() {
		if err := add(routeFoundations, e.foundationsPage); err != nil {
			return nil, err
		}
	}
	if e.hasReferences() {
		if err := add(routeReferences, e.referencesPage); err != nil {
			return nil, err
		}
	}
	for _, n := range e.export.Nodes {
		html, err := e.recordPage(n)
		if err != nil {
			return nil, err
		}
		pages[RecordRoute(n)+"index.html"] = html
	}
	return pages, nil
}

// RecordRoute is where one record's page is served from.
func RecordRoute(n ExportNode) string { return "record/" + n.Type + "/" + n.ID + "/" }

// --- the shell ------------------------------------------------------------

// shell wraps one explorer page in the shared header, sub-navigation and footer.
func (e *explorer) shell(route, title, script, gen, body string) string {
	var b strings.Builder
	b.WriteString(e.c.headWith(title, script))
	b.WriteString(e.c.headerFor("/record/"))
	b.WriteString(`<main id="app"><div class="page">`)
	b.WriteString(e.subnav(route))
	b.WriteString(`<div class="wrap record">`)
	b.WriteString(`<div class="rechead">`)
	if e.eyebrow != "" {
		b.WriteString(`<p class="eyebrow"` + e.eyebrowSrc + `>` + escapeText(e.eyebrow) + `</p>`)
	}
	b.WriteString(`<h1 class="pagetitle">` + escapeText(title) + `</h1>`)
	if gen != "" {
		b.WriteString(`<p class="gen">` + escapeText(gen) + `</p>`)
	}
	b.WriteString(`</div>`)
	b.WriteString(body)
	b.WriteString(`</div></div></main>`)
	b.WriteString(e.c.footer())
	b.WriteString("</body>\n</html>\n")
	return b.String()
}

// subnav renders the explorer's own tab strip. An entry whose page is omitted is
// omitted here too, so the strip can never point at a route the build did not
// write.
func (e *explorer) subnav(active string) string {
	type tab struct{ route, label string }
	tabs := []tab{
		{routeDashboard, e.c.ui.RecordNav.Dashboard},
		{routeGraph, e.c.ui.RecordNav.Graph},
		{routeTimeline, e.c.ui.RecordNav.Timeline},
	}
	if e.hasFoundations() {
		tabs = append(tabs, tab{routeFoundations, e.c.ui.RecordNav.Foundations})
	}
	tabs = append(tabs, tab{routeContributors, e.c.ui.RecordNav.Contributors})
	if e.hasReferences() {
		tabs = append(tabs, tab{routeReferences, e.c.ui.NavReferences})
	}
	var b strings.Builder
	b.WriteString(`<nav class="sub" aria-label="` + escapeAttr(e.c.ui.NavRecord) + `"><div class="wrap">`)
	for _, t := range tabs {
		cls, cur := "", ""
		if t.route == active {
			cls, cur = ` class="on"`, ` aria-current="page"`
		}
		b.WriteString(`<a href="/` + t.route + `"` + cls + cur + `>` + escapeText(t.label) + `</a>`)
	}
	b.WriteString(`</div></nav>`)
	return b.String()
}

// genLine is the derived line under every explorer heading.
//
// It carries what adr-47 decision 2 lets the generator add on its own and
// nothing else: numbers, dates and ids. Anything that would need a word to
// explain it belongs in a labelled tile instead.
func (e *explorer) genLine(parts ...string) string {
	var kept []string
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	if d := e.export.Build.GeneratedAt; d != "" {
		kept = append(kept, d)
	}
	if v := e.export.Build.Version; v != "" {
		kept = append(kept, "v"+v)
	}
	if cm := e.export.Build.Commit; cm != "" {
		kept = append(kept, cm)
	}
	return strings.Join(kept, " · ")
}

// --- shared pieces --------------------------------------------------------

// panel is one dashboard card: a heading, an optional right-aligned note, and a
// body.
func panel(span, heading, note, body string) string {
	cls := "panel"
	if span != "" {
		cls += " " + span
	}
	var b strings.Builder
	b.WriteString(`<div class="` + escapeAttr(cls) + `"><h3>` + escapeText(heading))
	if note != "" {
		b.WriteString(`<span>` + escapeText(note) + `</span>`)
	}
	b.WriteString(`</h3>` + body + `</div>`)
	return b.String()
}

// segment is one slice of a lifecycle bar.
type segment struct {
	Label string
	N     int
	Rank  int
}

// stateSegments is how one store grades its records: by the directory a record
// sits in where the store moves them, by the frontmatter status where the store
// is flat. A store that does neither has nothing to bar.
func (e *explorer) stateSegments(typ string) []segment {
	segs := segments(e.export.Counts.ByLifecycle[typ])
	if len(segs) == 1 && segs[0].Label == "" {
		segs = segments(e.export.Counts.ByStatus[typ])
	}
	kept := segs[:0]
	for _, s := range segs {
		if s.Label != "" {
			kept = append(kept, s)
		}
	}
	return kept
}

// segments turns one store's state counts into ordered slices.
func segments(byLifecycle map[string]int) []segment {
	out := make([]segment, 0, len(byLifecycle))
	for k, v := range byLifecycle {
		rank, ok := lifecycleRank[k]
		if !ok {
			rank = len(lifecycleRamp) - 1
		}
		out = append(out, segment{Label: k, N: v, Rank: rank})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Rank != out[j].Rank {
			return out[i].Rank < out[j].Rank
		}
		return out[i].Label < out[j].Label
	})
	return out
}

// segBar renders a lifecycle bar with its legend and its table twin. Every
// visual on the dashboard carries one: the bar is the picture, the table is the
// same numbers in a form a screen reader can read out.
func (e *explorer) segBar(caption string, segs []segment) string {
	total := 0
	for _, s := range segs {
		total += s.N
	}
	if total == 0 {
		return ""
	}
	var bar, leg, rows strings.Builder
	for _, s := range segs {
		colour := lifecycleRamp[s.Rank]
		pct := strconv.FormatFloat(float64(s.N)/float64(total)*100, 'f', 2, 64)
		bar.WriteString(`<i style="width:` + pct + `%;background:` + colour + `" title="` +
			escapeAttr(s.Label) + `: ` + strconv.Itoa(s.N) + `"></i>`)
		leg.WriteString(`<span><i style="background:` + colour + `"></i>` + escapeText(s.Label) +
			` <b class="tnum">` + strconv.Itoa(s.N) + `</b></span>`)
		rows.WriteString(`<tr><td>` + escapeText(s.Label) + `</td><td class="tnum">` + strconv.Itoa(s.N) + `</td></tr>`)
	}
	return `<div class="seg" role="presentation">` + bar.String() + `</div>` +
		`<div class="legend">` + leg.String() + `</div>` +
		e.tableTwin(caption, `<tbody>`+rows.String()+`</tbody>`)
}

// tableTwin folds a visual's own numbers away behind a disclosure, so every
// picture on the page has a form that can be read out and copied.
func (e *explorer) tableTwin(caption, body string) string {
	return `<details class="twin"><summary>` + escapeText(e.c.ui.Panels.TableView) +
		`</summary><div class="tablewrap"><table><caption class="sr">` + escapeText(caption) +
		`</caption>` + body + `</table></div></details>`
}

// --- the dashboard --------------------------------------------------------

// dashboard renders `/record/`: what the record holds, counted.
func (e *explorer) dashboard() (string, error) {
	ui := e.c.ui
	var b strings.Builder
	b.WriteString(`<div class="dash">`)

	// Stat tiles. A store the repository does not keep gets no tile rather than
	// a tile reading zero.
	if n := len(e.export.Releases); n > 0 {
		sub := []string{}
		r := e.export.Releases[0]
		sub = append(sub, "v"+r.Version, r.Date)
		b.WriteString(tile(strconv.Itoa(n), ui.Tiles.Releases, sub))
	}
	for _, typ := range e.storeOrder() {
		n := e.export.Counts.ByType[typ]
		if n == 0 {
			continue
		}
		var sub []string
		for _, s := range e.stateSegments(typ) {
			sub = append(sub, strconv.Itoa(s.N)+" "+s.Label)
		}
		if len(sub) > 2 {
			sub = sub[:2]
		}
		b.WriteString(tile(strconv.Itoa(n), ui.Tiles.ForType(typ), sub))
	}

	// State bars, one per store that grades its records at all.
	for _, typ := range e.storeOrder() {
		segs := e.stateSegments(typ)
		if len(segs) < 2 {
			continue
		}
		caption := ui.Tiles.ForType(typ)
		b.WriteString(panel("c6", caption, strconv.Itoa(e.export.Counts.ByType[typ]), e.segBar(caption, segs)))
	}

	if cad := e.cadence(); cad != "" {
		b.WriteString(cad)
	}
	b.WriteString(e.latestDecisions())
	b.WriteString(e.health())
	b.WriteString(`</div>`)

	return e.shell(routeDashboard, ui.RecordNav.Dashboard, "", e.genLine(), b.String()), nil
}

// tile is one stat tile.
func tile(n, label string, sub []string) string {
	var b strings.Builder
	b.WriteString(`<div class="panel c2"><div class="tile">`)
	b.WriteString(`<span class="n">` + escapeText(n) + `</span>`)
	b.WriteString(`<span class="l">` + escapeText(label) + `</span>`)
	for _, s := range sub {
		if s == "" {
			continue
		}
		b.WriteString(`<span class="s">` + escapeText(s) + `</span>`)
	}
	b.WriteString(`</div></div>`)
	return b.String()
}

// storeOrder is the order the stores are read in, decisions first — the same
// order the chart's same-day tie-break uses, so the two pages agree.
func (e *explorer) storeOrder() []string {
	types := make([]string, 0, len(e.export.Counts.ByType))
	for t := range e.export.Counts.ByType {
		types = append(types, t)
	}
	sort.Slice(types, func(i, j int) bool {
		if a, b := typeRank(types[i]), typeRank(types[j]); a != b {
			return a < b
		}
		return types[i] < types[j]
	})
	return types
}

// cadence renders the release strip: one tick per release, alternating above and
// below the axis, with the same versions and dates in its table twin.
func (e *explorer) cadence() string {
	rel := e.export.Releases
	if len(rel) < 2 {
		return ""
	}
	// The export lists releases newest first; the axis runs the other way.
	asc := make([]int, len(rel))
	for i := range rel {
		asc[i] = len(rel) - 1 - i
	}
	first, last := dayNumber(rel[len(rel)-1].Date), dayNumber(rel[0].Date)
	span := float64(last - first)
	if span <= 0 {
		span = 1
	}
	var ticks, rows strings.Builder
	for k, i := range asc {
		r := rel[i]
		x := 24 + float64(dayNumber(r.Date)-first)/span*552
		y1, y2, ty := 26, 40, 18
		if k%2 == 1 {
			y1, y2, ty = 40, 54, 66
		}
		xs := strconv.FormatFloat(x, 'f', 1, 64)
		ticks.WriteString(`<g class="tick"><title>v` + escapeText(r.Version) + ` — ` + escapeText(r.Date) + `</title>`)
		ticks.WriteString(`<line x1="` + xs + `" y1="` + strconv.Itoa(y1) + `" x2="` + xs + `" y2="` + strconv.Itoa(y2) +
			`" stroke="var(--s-adr)" stroke-width="2"/>`)
		ticks.WriteString(`<text x="` + xs + `" y="` + strconv.Itoa(ty) +
			`" font-size="10" text-anchor="middle" fill="var(--ink-2)">v` + escapeText(r.Version) + `</text></g>`)
		rows.WriteString(`<tr><td>v` + escapeText(r.Version) + `</td><td>` + escapeText(r.Date) + `</td></tr>`)
	}
	svg := `<svg viewBox="0 0 600 70" class="cadsvg" role="img" aria-label="` +
		escapeAttr(e.c.ui.Panels.Cadence+" · "+rel[len(rel)-1].Date+" – "+rel[0].Date) + `">` +
		`<line x1="8" y1="40" x2="592" y2="40" class="axis"/>` + ticks.String() + `</svg>`
	body := svg + e.tableTwin(e.c.ui.Panels.Cadence, `<tbody>`+rows.String()+`</tbody>`)
	note := strconv.Itoa(len(rel)) + " " + e.c.ui.Tiles.Releases
	return panel("c12 cadence", e.c.ui.Panels.Cadence, note, body)
}

// latestDecisions lists the newest ratified decisions by their own dates. It is
// ids, dates and titles: the record's words, never a summary written here.
func (e *explorer) latestDecisions() string {
	var adrs []ExportNode
	for _, n := range e.export.Nodes {
		if n.Type == "adr" {
			adrs = append(adrs, n)
		}
	}
	if len(adrs) == 0 {
		return ""
	}
	sort.SliceStable(adrs, func(i, j int) bool {
		if adrs[i].Date != adrs[j].Date {
			return adrs[i].Date > adrs[j].Date
		}
		return handleNum(adrs[i].ID) > handleNum(adrs[j].ID)
	})
	if len(adrs) > 7 {
		adrs = adrs[:7]
	}
	var b strings.Builder
	b.WriteString(`<ul class="list">`)
	for _, n := range adrs {
		b.WriteString(`<li><span class="id"><a href="/` + escapeAttr(RecordRoute(n)) + `">` + escapeText(n.ID) +
			`</a><span class="d">` + escapeText(n.Date) + `</span></span>` +
			`<span>` + escapeText(shortTitle(n)) + `</span></li>`)
	}
	b.WriteString(`</ul>`)
	return panel("c8", e.c.ui.Panels.Latest, e.c.ui.Tiles.ADR, b.String())
}

// health renders the record's own reference hygiene: every typed reference the
// tree cannot resolve, measured against the committed ratchet.
func (e *explorer) health() string {
	h := e.export.Health
	var b strings.Builder
	b.WriteString(`<div class="health">`)
	for _, u := range h.Unresolved {
		b.WriteString(`<div><span class="w">!</span> <a href="/` + escapeAttr(e.routeOf(u.From)) + `">` +
			escapeText(u.From) + `</a> → <b class="stub">` + escapeText(u.To) + `</b> <span class="muted">` +
			escapeText(relationWord(u.Rel)) + ` · ` + escapeText(e.c.ui.Record.NotInTree) + `</span></div>`)
	}
	b.WriteString(`<div class="hsum">` + escapeText(strconv.Itoa(len(h.Unresolved))) + ` / ` +
		escapeText(strconv.Itoa(h.BaselineCount)) + ` · ` +
		escapeText(strconv.Itoa(e.export.Layout.Isolated)) + `</div>`)
	b.WriteString(`</div>`)
	return panel("c4", e.c.ui.Panels.Health, strconv.Itoa(h.BaselineCount), b.String())
}

// routeOf is a record's page, or "" where no file answers to the id.
func (e *explorer) routeOf(id string) string {
	n, ok := e.byID[id]
	if !ok {
		return ""
	}
	return RecordRoute(n)
}

// shortTitle drops the handle a record repeats in its own H1 ("ADR-47: …"), so
// a list of ids does not print each one twice.
func shortTitle(n ExportNode) string {
	prefix := strings.ToUpper(n.ID) + ":"
	if strings.HasPrefix(strings.ToUpper(n.Title), prefix) {
		return strings.TrimSpace(n.Title[len(prefix):])
	}
	return n.Title
}

// dayNumber converts a YYYY-MM-DD date to a day count, for placing a mark on an
// axis. It is a plain civil-calendar computation: no clock, no zone, no library.
func dayNumber(date string) int {
	if len(date) < 10 {
		return 0
	}
	y, err1 := strconv.Atoi(date[0:4])
	m, err2 := strconv.Atoi(date[5:7])
	d, err3 := strconv.Atoi(date[8:10])
	if err1 != nil || err2 != nil || err3 != nil {
		return 0
	}
	// Howard Hinnant's days-from-civil: exact for every date, and it is the same
	// arithmetic on every machine, which is what the golden files rest on.
	if m <= 2 {
		y--
	}
	era := y / 400
	if y < 0 {
		era = (y - 399) / 400
	}
	yoe := y - era*400
	mp := (m + 9) % 12
	doy := (153*mp+2)/5 + d - 1
	doe := yoe*365 + yoe/4 - yoe/100 + doy
	return era*146097 + doe - 719468
}

// --- foundations ----------------------------------------------------------

// foundationsPage lists what the work is founded on: the principles that hold
// across the record, and the disciplines that state a rule rather than ship a
// change. It LISTS AND LINKS and never explains — the explanation is the record
// page each card opens, and the context belongs in the documentation.
func (e *explorer) foundationsPage() (string, error) {
	var b strings.Builder
	b.WriteString(`<div class="dash">`)
	deck := func(label string, nodes []ExportNode) {
		if len(nodes) == 0 {
			return
		}
		var cards strings.Builder
		cards.WriteString(`<div class="fcards">`)
		for _, n := range nodes {
			cards.WriteString(`<a class="fcard" href="/` + escapeAttr(RecordRoute(n)) + `">` +
				`<span class="t">` + escapeText(shortTitle(n)) + `</span>` +
				`<span class="id">` + escapeText(n.ID) + `</span></a>`)
		}
		cards.WriteString(`</div>`)
		b.WriteString(panel("c12", label, strconv.Itoa(len(nodes)), cards.String()))
	}
	deck(e.c.ui.Tiles.Principle, e.principles)
	deck(e.c.ui.Tiles.Discipline, e.disciplines)
	b.WriteString(`</div>`)
	return e.shell(routeFoundations, e.c.ui.RecordNav.Foundations, "", e.genLine(), b.String()), nil
}

// --- contributors ---------------------------------------------------------

// contributorsPage renders `/contributors/`: who authored the history, and what
// assisted.
//
// The two are different facts and the page keeps them apart. Humans are the
// authors of record. The `Assisted-by:` tallies are DISCLOSURE, presented next
// to the policy that requires them — and this page is the one place on the site
// where a model name may appear, under the declared attribution escape
// (adr-47 decision 3).
func (e *explorer) contributorsPage() (string, error) {
	a := e.export.Authorship
	ui := e.c.ui
	var b strings.Builder
	b.WriteString(`<div class="dash">`)

	if a.Commits > 0 {
		b.WriteString(tile(strconv.Itoa(a.Commits), ui.Tiles.Commits, []string{e.export.History.FirstCommit}))
	}
	if len(a.Humans) > 0 {
		b.WriteString(tile(strconv.Itoa(len(a.Humans)), ui.Contributors.Authors, nil))
	}
	if a.Commits > 0 {
		share := strconv.Itoa(a.Assisted*100/a.Commits) + "%"
		b.WriteString(tile(share, ui.Contributors.Assisted,
			[]string{strconv.Itoa(a.Assisted) + " / " + strconv.Itoa(a.Commits)}))
	}

	if len(a.Humans) > 0 || len(a.Bots) > 0 {
		var rows strings.Builder
		rows.WriteString(`<thead><tr><th>` + escapeText(ui.Contributors.Authors) + `</th><th class="tnum">` +
			escapeText(ui.Tiles.Commits) + `</th></tr></thead><tbody>`)
		for _, h := range a.Humans {
			rows.WriteString(`<tr><td>` + escapeText(h.Name) + `</td><td class="tnum">` + strconv.Itoa(h.Commits) + `</td></tr>`)
		}
		for _, t := range a.Bots {
			rows.WriteString(`<tr class="muted"><td>` + escapeText(t.Name) + ` <span class="rel">` +
				escapeText(ui.Contributors.Tools) + `</span></td><td class="tnum">` + strconv.Itoa(t.Commits) + `</td></tr>`)
		}
		rows.WriteString(`</tbody>`)
		policy, err := e.policyQuote()
		if err != nil {
			return "", err
		}
		body := `<div class="tablewrap"><table>` + rows.String() + `</table></div>` + policy
		b.WriteString(panel("c6", ui.Contributors.Authors, "", body))
	}

	if len(a.ByModel) > 0 {
		maxN := a.ByModel[0].Commits
		var bars, rows strings.Builder
		bars.WriteString(`<div class="bars">`)
		for _, m := range a.ByModel {
			pct := strconv.FormatFloat(float64(m.Commits)/float64(maxN)*100, 'f', 1, 64)
			bars.WriteString(`<span class="lab">` + escapeText(m.Model) + `</span>` +
				`<span class="bar"><i style="width:` + pct + `%"></i></span>` +
				`<span class="tnum small">` + strconv.Itoa(m.Commits) + `</span>`)
			rows.WriteString(`<tr><td>` + escapeText(m.Model) + `</td><td class="tnum">` + strconv.Itoa(m.Commits) + `</td></tr>`)
		}
		bars.WriteString(`</div>`)
		body := bars.String() + e.tableTwin(ui.Contributors.Trailers, `<tbody>`+rows.String()+`</tbody>`)
		b.WriteString(panel("c6", ui.Contributors.Trailers, strconv.Itoa(a.Assisted), body))
	}
	b.WriteString(`</div>`)

	return e.shell(routeContributors, ui.RecordNav.Contributors, "", e.genLine(), b.String()), nil
}

// policyQuote renders the attribution policy the manifest selects, verbatim,
// with the file it came from linked. The number above it means nothing without
// the rule beside it.
//
// A repository that DECLARES no policy simply has none, and the page renders
// without it. A repository that names one and cannot supply it REFUSES: the
// tallies would go out unaccompanied, which is the reading — assistance as
// authorship — the whole page exists to prevent.
func (e *explorer) policyQuote() (string, error) {
	p := e.c.manifest.RecordPages.Contributors.Policy
	if p.File == "" {
		return "", nil
	}
	bad := func(why string) error {
		return fmt.Errorf("site: record_pages.contributors.policy names %s § %s, and %s — the assistance tallies are not published without the rule beside them",
			p.File, p.Heading, why)
	}
	data, err := fsutil.ReadGuarded(joinRepo(e.c.repoRoot, p.File), maxPageBytes)
	if err != nil {
		if os.IsNotExist(err) {
			return "", bad("the repository does not carry it")
		}
		return "", err
	}
	body, consumed := StripFrontmatter(string(data))
	secs, err := Sections(p.File, body, consumed)
	if err != nil {
		return "", err
	}
	for _, s := range secs {
		if !strings.EqualFold(s.Title, p.Heading) {
			continue
		}
		blocks := Blocks(s.Body, s.BodyLine)
		if len(blocks) == 0 {
			return "", bad("that section is empty")
		}
		if p.Part == "first-bullet" {
			// The first BULLET, not the first block. A policy section opens with
			// its own preamble more often than not, and quoting that instead
			// published a dangling lead-in — "The rules:" and then nothing —
			// under the number it was supposed to explain.
			item := -1
			for i, blk := range blocks {
				if isUnorderedItem(strings.TrimLeft(blk.Text, " \t")) {
					item = i
					break
				}
			}
			if item < 0 {
				return "", bad("that section has no bullet to quote")
			}
			text, line := firstListItem(blocks[item])
			blocks = []Block{{Text: text, Line: line}}
		}
		r := &Renderer{UI: e.c.ui, Refs: LinkDefinitions(body),
			Image: func(src, alt string, at Source) (string, error) {
				return e.c.assets.render(path.Dir(p.File), src, alt, at)
			},
			Link: func(href string, at Source) string { return e.href(p.File, href) }}
		h, err := r.RenderBlocks(p.File, blocks)
		if err != nil {
			return "", err
		}
		out := `<div class="prose small policy"` + srcAttr(p.File, s.Anchor) + `>` + h
		if e.c.repo.Repository != "" {
			out += `<p class="small"><a href="` + escapeAttr(e.c.repo.Repository+"/blob/main/"+p.File) + `">` +
				escapeText(p.File) + `</a></p>`
		}
		return out + `</div>`, nil
	}
	return "", bad("that file has no such heading")
}

// --- link rewriting -------------------------------------------------------

// href maps a link as a record wrote it to the link the site serves.
//
// A relative path to another RECORD becomes that record's page here, which is
// what the explorer adds: before it, the only honest destination was the file on
// the forge. Everything else falls through to the landing page's own rule.
func (e *explorer) href(fromPath, target string) string {
	switch {
	case target == "",
		strings.HasPrefix(target, "http://"),
		strings.HasPrefix(target, "https://"),
		strings.HasPrefix(target, "mailto:"),
		strings.HasPrefix(target, "#"),
		strings.HasPrefix(target, "/"):
		return target
	}
	file, frag, _ := strings.Cut(target, "#")
	rel := path.Clean(path.Join(path.Dir(fromPath), file))
	if strings.HasSuffix(file, ".md") {
		if n, ok := e.byPath[rel]; ok {
			out := "/" + RecordRoute(n)
			if frag != "" {
				out += "#" + frag
			}
			return out
		}
		return siteHref(path.Dir(fromPath), target, e.c.repo.Repository)
	}
	// A relative target that is NOT markdown — a directory, a script, a
	// configuration file. `siteHref` leaves those exactly as the record wrote
	// them, which on the landing page is harmless (nothing links one) and on a
	// record's page is a link relative to `/record/<type>/<id>/` that resolves
	// to nothing. It resolves against the repository root instead, and points at
	// the forge's own view of whatever is there.
	if e.c.repo.Repository == "" || file == "" || !fsutil.ValidRelPath(rel) {
		return target
	}
	kind := "blob"
	if dir, err := os.Stat(joinRepo(e.c.repoRoot, rel)); err == nil && dir.IsDir() {
		kind = "tree"
	} else if err != nil {
		// The tree does not carry it. The record's own text stands: inventing a
		// forge URL for a path that is not there trades a broken relative link
		// for a confident 404.
		return target
	}
	out := e.c.repo.Repository + "/" + kind + "/main/" + rel
	if frag != "" {
		out += "#" + frag
	}
	return out
}

// forgeBlob is the record file on the forge, or "" without a known forge.
func (e *explorer) forgeBlob(rel string) string {
	if e.c.repo.Repository == "" {
		return ""
	}
	return e.c.repo.Repository + "/blob/main/" + rel
}

// forgeCommits is the record file's commit history on the forge — the link that
// makes an amendment traceable from the date rather than merely visible as one.
func (e *explorer) forgeCommits(rel string) string {
	if e.c.repo.Repository == "" {
		return ""
	}
	return e.c.repo.Repository + "/commits/main/" + rel
}

// nodesOfType is every record of one store, in the export's order.
func (e *explorer) nodesOfType(typ string) []ExportNode {
	var out []ExportNode
	for _, n := range e.export.Nodes {
		if n.Type == typ {
			out = append(out, n)
		}
	}
	return out
}
