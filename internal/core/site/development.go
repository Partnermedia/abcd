package site

// The development page — `/record/development/`, sibling to Foundations.
//
// Foundations reads the stores that HOLD: principles, and the intents that
// state a rule rather than ship a change. This reads the stores that MOVE —
// decisions, intents, specs and issues — as the same deck of linked cards, so
// the two pages are one shape read twice. A card carries a record's id and its
// title and links that record's own page; it never explains, exactly as its
// sibling never does.
//
// # Why it is grouped and folded rather than listed
//
// Foundations can list a store flat because a repository keeps tens of
// principles. This repository keeps ~138 intents, ~40 specs and ~525 issues, and
// a flat deck of 525 cards is a scroll, not a page. The structure that follows
// from the record's own shape is therefore:
//
//   - one panel per store, its note the store's true total;
//   - inside it the lifecycle bar, whose legend names EVERY bucket with its true
//     count — so the page states the whole store as text before it shows one
//     card;
//   - then one deck per lifecycle bucket, in `lifecycleRank` order. A bucket the
//     work has not settled yet (drafts, proposed, open, planned) is a panel a
//     reader sees; a settled one (shipped, accepted, closed, resolved,
//     superseded, wontfix) is folded behind `panelDisclosure`, because it is the
//     larger half and the one a reader chooses to open rather than reads on the
//     way past.
//
// The bucket a record is graded by is its own — the directory the store moves it
// between, or the frontmatter status where the store is flat. That is the same
// reading `stateSegments` takes, and the ordering, the colour and the legend all
// come from `segments`, `lifecycleRank` and `segBar` rather than from a second
// copy of them here.
//
// # The cap, and why it cannot be silent
//
// A deck is capped at `developmentDeckCap` cards. A cap that hid what it cut
// would publish a partial store as a whole one, so a capped panel's note states
// BOTH figures — how many cards are shown and how many the bucket holds — and
// the bar's legend above it states every bucket's true count whether its deck is
// capped or not. Nothing on this page is a number a reader has to trust.
//
// Cards run newest first, by the record's effective date and then by its handle,
// so the cut is always the oldest tail of a bucket and is the same on every
// build.
//
// # What is not here
//
// Principles and disciplines are Foundations' half and are not repeated. The
// issue ledger is working-tier data (adr-32) and is in the export only where the
// repository opts in — so a repository that has not opted in simply has no issue
// nodes, and the store is omitted by the same rule that omits any empty one.
// Graceful absence throughout (itd-140): a store with no records gets no panel,
// and a record with no store to speak of gets no page at all.

import (
	"sort"
	"strconv"
	"strings"
)

// developmentDeckCap bounds one lifecycle deck. It is a reading limit, not a
// data limit: what it cuts is stated beside what it shows, and the whole store
// is reachable from the relationship chart and from every card's own page.
const developmentDeckCap = 48

// settledRank is the first `lifecycleRank` whose deck is folded away. Ranks
// below it are work in flight; ranks at or above it are work the store has
// settled — the larger half, and the one a reader opens on purpose.
const settledRank = 3

// developmentStore is every record of one store that MOVES.
//
// It is the complement of what Foundations reads: a principle is not a change to
// ship, and neither is an intent filed as a discipline.
func (e *explorer) developmentStore(typ string) []ExportNode {
	if typ == principleType {
		return nil
	}
	var out []ExportNode
	for _, n := range e.nodesOfType(typ) {
		if n.Lifecycle == disciplinesLifecycle {
			continue
		}
		out = append(out, n)
	}
	return out
}

// bucketOf is how one record's own store grades it: the lifecycle directory it
// sits in, or the frontmatter status where the store is flat. It is the
// per-record reading of the same fact `stateSegments` counts per store, so the
// two can never disagree about which bucket a record belongs to.
func bucketOf(n ExportNode) string {
	if n.Lifecycle != "" {
		return n.Lifecycle
	}
	return n.Status
}

// hasDevelopment reports whether the record holds anything that moves. Without a
// single such record the page and its navigation entry are omitted.
func (e *explorer) hasDevelopment() bool {
	for _, typ := range e.storeOrder() {
		if len(e.developmentStore(typ)) > 0 {
			return true
		}
	}
	return false
}

// developmentPage renders `/record/development/`.
func (e *explorer) developmentPage() (string, error) {
	var b strings.Builder
	b.WriteString(`<div class="dash">`)
	for _, typ := range e.storeOrder() {
		nodes := e.developmentStore(typ)
		if len(nodes) == 0 {
			continue
		}
		b.WriteString(panel("c12", e.c.ui.Tiles.ForType(typ), strconv.Itoa(len(nodes)),
			e.developmentStorePanel(nodes)))
	}
	b.WriteString(`</div>`)
	return e.shell(routeDevelopment, e.c.ui.RecordNav.Development, "", b.String()), nil
}

// developmentStorePanel is one store's body: its lifecycle bar, then one deck
// per bucket.
func (e *explorer) developmentStorePanel(nodes []ExportNode) string {
	byBucket := map[string][]ExportNode{}
	counts := map[string]int{}
	for _, n := range nodes {
		k := bucketOf(n)
		byBucket[k] = append(byBucket[k], n)
		counts[k]++
	}

	var b strings.Builder
	// The bar drops an unlabelled bucket exactly as `stateSegments` does — a
	// slice with no name is a colour a reader cannot read. Its records are still
	// listed below, and the panel's own note still counts them.
	named := map[string]int{}
	for k, v := range counts {
		if k != "" {
			named[k] = v
		}
	}
	b.WriteString(e.segBar(segments(named)))

	for _, seg := range segments(counts) {
		group := byBucket[seg.Label]
		sortDevelopmentNodes(group)
		shown := group
		note := strconv.Itoa(len(group))
		if len(shown) > developmentDeckCap {
			shown = shown[:developmentDeckCap]
			// Both figures, each its own number: what is drawn, and what the
			// bucket actually holds. A cap that stated only one of them would be
			// the silent truncation this page must not perform.
			note = strconv.Itoa(len(shown)) + " / " + strconv.Itoa(len(group))
		}
		deck := developmentDeck(shown)
		switch {
		case seg.Label == "":
			// A record its store grades by neither directory nor status. It is
			// listed under the store's own heading rather than under a bucket
			// name invented for it here.
			b.WriteString(deck)
		case seg.Rank >= settledRank:
			b.WriteString(panelDisclosure("", seg.Label, "", note, deck))
		default:
			b.WriteString(panel("", seg.Label, note, deck))
		}
	}
	return b.String()
}

// sortDevelopmentNodes puts a bucket newest first, so a capped deck cuts its
// oldest tail and cuts the same one on every build.
func sortDevelopmentNodes(nodes []ExportNode) {
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].Date != nodes[j].Date {
			return nodes[i].Date > nodes[j].Date
		}
		if a, b := handleNum(nodes[i].ID), handleNum(nodes[j].ID); a != b {
			return a > b
		}
		return nodes[i].ID < nodes[j].ID
	})
}

// developmentDeck is a deck of linked record cards — the same card Foundations
// deals: a title, an id, and a link to the record's own page.
func developmentDeck(nodes []ExportNode) string {
	var b strings.Builder
	b.WriteString(`<div class="fcards">`)
	for _, n := range nodes {
		b.WriteString(`<a class="fcard" href="/` + escapeAttr(RecordRoute(n)) + `">` +
			`<span class="t">` + escapeText(shortTitle(n)) + `</span>` +
			`<span class="id">` + escapeText(n.ID) + `</span></a>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}
