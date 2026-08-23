package site

// The development page's tests: the deck renders every store that moves, omits
// the ones the repository does not keep, links only records the build wrote, and
// never cuts a deck without saying what it cut.

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestDevelopmentPageDealsEveryMovingStore is the page's own shape: one panel
// per store that moves, each bucket named by the record's own lifecycle, and
// nothing on it that belongs to Foundations.
func TestDevelopmentPageDealsEveryMovingStore(t *testing.T) {
	f := newFixture(t)
	out := t.TempDir()
	res := buildFixture(t, f, out)

	if !containsString(res.Files, "record/development/index.html") {
		t.Fatalf("the build wrote no development page: %v", res.Files)
	}
	page := outFile(t, out, "record/development/index.html")

	// Every store the fixture keeps that MOVES gets its panel, captioned with
	// that store's own interface label.
	for _, label := range []string{"decisions", "intents", "specs", "issues"} {
		if !strings.Contains(page, ">"+label+"<") {
			t.Errorf("the page has no panel for the %s store", label)
		}
	}
	// Every record of those stores is on the page.
	for _, id := range []string{"adr-1", "adr-2", "itd-1", "itd-2", "itd-6", "spc-1", "iss-1"} {
		if !strings.Contains(page, ">"+id+"<") {
			t.Errorf("the page omits %s", id)
		}
	}
	// The lifecycle a record's own store grades it by names its deck — the
	// directory it sits in, or the frontmatter status where the store is flat.
	for _, bucket := range []string{"drafts", "shipped", "superseded", "open", "accepted"} {
		if !strings.Contains(page, ">"+bucket+"<") {
			t.Errorf("no deck is named for the %s bucket", bucket)
		}
	}
	// Foundations' half is not repeated here.
	if strings.Contains(page, ">principles<") {
		t.Error("the development page deals the principle store, which Foundations holds")
	}
	if strings.Contains(page, ">itd-5<") {
		t.Error("the development page deals a discipline, which Foundations holds")
	}
	// A settled bucket is folded rather than listed; an unsettled one is not.
	if !strings.Contains(page, `<details class="panel fold"><summary><h3>superseded`) {
		t.Error("the superseded deck is not folded away")
	}
	if strings.Contains(page, `<details class="panel fold"><summary><h3>drafts`) {
		t.Error("the drafts deck — work in flight — was folded away")
	}
	// The navigation reaches the page it wrote.
	if !strings.Contains(outFile(t, out, "record/index.html"), `href="/record/development/"`) {
		t.Error("the navigation does not reach the development page")
	}
}

// TestDevelopmentCardsLinkRealRecordRoutes: every card is a link to a page this
// build actually wrote. A card pointing at a route that is not there is a
// confident 404, which is worse than an omission.
func TestDevelopmentCardsLinkRealRecordRoutes(t *testing.T) {
	f := newFixture(t)
	out := t.TempDir()
	res := buildFixture(t, f, out)
	page := outFile(t, out, "record/development/index.html")

	written := map[string]bool{}
	for _, name := range res.Files {
		written[name] = true
	}
	cards := regexp.MustCompile(`<a class="fcard" href="/([^"]+)"`).FindAllStringSubmatch(page, -1)
	if len(cards) == 0 {
		t.Fatal("the development page dealt no cards at all")
	}
	for _, m := range cards {
		if !written[m[1]+"index.html"] {
			t.Errorf("a card links /%s, which the build did not write", m[1])
		}
	}
}

// TestDevelopmentOmitsTheLedgerWithoutTheOptIn is adr-32 at this page: the issue
// ledger is working-tier data, so a repository that has not opted in has no
// issue nodes in the export at all and the store is omitted like any empty one.
func TestDevelopmentOmitsTheLedgerWithoutTheOptIn(t *testing.T) {
	f := newFixture(t)
	manifest := outFile(t, f.Root(), ".abcd/site.json")
	f.write(".abcd/site.json", strings.Replace(manifest,
		`"record": {"issue_ledger": true}`, `"record": {"issue_ledger": false}`, 1))
	out := t.TempDir()
	res, err := Build(Request{RepoRoot: f.Root(), OutDir: out, Stamp: fixtureStamp})
	if err != nil {
		t.Fatal(err)
	}
	if !containsString(res.Files, "record/development/index.html") {
		t.Fatal("the development page went with the ledger")
	}
	page := outFile(t, out, "record/development/index.html")
	if strings.Contains(page, ">issues<") {
		t.Error("the page deals an issue panel the repository did not opt in to publishing")
	}
	if strings.Contains(page, "iss-1") {
		t.Error("the page names a ledger entry the repository did not opt in to publishing")
	}
	if !strings.Contains(page, ">intents<") {
		t.Error("the rest of the page went with the ledger")
	}
}

// TestBuildWithoutDevelopment is itd-140's graceful absence: a repository whose
// stores all HOLD gets no development page and no navigation entry pointing at
// one, and the build succeeds.
func TestBuildWithoutDevelopment(t *testing.T) {
	f := newFixture(t)
	for _, dir := range []string{
		".abcd/development/decisions/adrs",
		".abcd/development/specs",
		".abcd/work/issues",
		".abcd/development/intents/drafts",
		".abcd/development/intents/planned",
		".abcd/development/intents/shipped",
		".abcd/development/intents/superseded",
	} {
		if err := os.RemoveAll(filepath.Join(f.Root(), filepath.FromSlash(dir))); err != nil {
			t.Fatal(err)
		}
	}
	out := t.TempDir()
	res, err := Build(Request{RepoRoot: f.Root(), OutDir: out, Stamp: fixtureStamp})
	if err != nil {
		t.Fatalf("a repository whose stores all hold must still build: %v", err)
	}
	if containsString(res.Files, "record/development/index.html") {
		t.Error("the development page rendered with nothing that moves")
	}
	dash := outFile(t, out, "record/index.html")
	if strings.Contains(dash, `href="/record/development/"`) {
		t.Error("the navigation points at a development page the build did not write")
	}
	if !strings.Contains(dash, `href="/record/foundations/"`) {
		t.Error("Foundations went with it")
	}
}

// TestDevelopmentCapStatesTheTrueTotal: a deck longer than the cap is cut, and a
// cut that did not say so would publish part of a store as the whole of it. Both
// figures are stated on the panel, and the bar's legend above states the bucket's
// true count whether the deck was cut or not.
func TestDevelopmentCapStatesTheTrueTotal(t *testing.T) {
	const held = developmentDeckCap + 12
	var nodes []ExportNode
	for i := 1; i <= held; i++ {
		id := "iss-" + strconv.Itoa(i)
		nodes = append(nodes, ExportNode{ID: id, Type: "issue", Lifecycle: "open",
			Title: "issue " + strconv.Itoa(i), Path: ".abcd/work/issues/open/" + id + ".md",
			Date: "2026-01-01"})
	}
	body := (&explorer{}).developmentStorePanel(nodes)

	if n := strings.Count(body, `<a class="fcard"`); n != developmentDeckCap {
		t.Errorf("the deck drew %d cards, not the %d the cap allows", n, developmentDeckCap)
	}
	want := strconv.Itoa(developmentDeckCap) + " / " + strconv.Itoa(held)
	if !strings.Contains(body, ">"+want+"<") {
		t.Errorf("a cut deck does not state %q — it publishes part of a store as the whole of it", want)
	}
	// The legend states the bucket's true count as text, beside its own name.
	if !strings.Contains(body, `open <b class="tnum">`+strconv.Itoa(held)+`</b>`) {
		t.Errorf("the lifecycle legend does not state the bucket's true total of %d", held)
	}
	// An uncut deck states one figure, and it is the whole of the bucket.
	small := (&explorer{}).developmentStorePanel(nodes[:3])
	if !strings.Contains(small, ">3<") {
		t.Error("an uncut deck does not state its count")
	}
	if strings.Contains(small, " / ") {
		t.Error("an uncut deck claims to have been cut")
	}
}
