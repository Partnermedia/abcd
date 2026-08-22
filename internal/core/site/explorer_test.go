package site

// The record explorer's tests, over the same in-process fixture repository the
// landing page's golden files use.
//
// The fixture carries one of every shape the pages have to handle: a flat store
// graded in frontmatter and a store graded by directory, a discipline, a
// superseded record, an opted-in issue ledger, two frontmatter-free principles,
// a supersession whose target has left the tree, and a bibliography.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// explorerGoldens are the routes pinned byte for byte. One page of each kind,
// plus a record page from a store with frontmatter and one from the store
// without, because those two take different paths through the build.
var explorerGoldens = []struct{ route, golden string }{
	{"record/index.html", "record-dashboard.html"},
	{"record/graph/index.html", "record-graph.html"},
	{"record/timeline/index.html", "record-timeline.html"},
	{"record/foundations/index.html", "record-foundations.html"},
	{"contributors/index.html", "contributors.html"},
	{"references/index.html", "references.html"},
	{"record/adr/adr-1/index.html", "record-adr-1.html"},
	{"record/intent/itd-2/index.html", "record-itd-2.html"},
	{"record/principle/one-fixture-principle/index.html", "record-principle.html"},
}

// TestExplorerGolden pins every explorer page the fixture produces.
func TestExplorerGolden(t *testing.T) {
	f := newFixture(t)
	out := t.TempDir()
	res := buildFixture(t, f, out)

	for _, g := range explorerGoldens {
		if !containsString(res.Files, g.route) {
			t.Fatalf("build did not write %s: %v", g.route, res.Files)
		}
		data, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(g.route)))
		if err != nil {
			t.Fatal(err)
		}
		golden(t, g.golden, data)
	}
	// Every record in the export has a page, and nothing else does.
	pages := 0
	for _, f := range res.Files {
		if strings.HasSuffix(f, "index.html") {
			pages++
		}
	}
	if res.Pages+1 != pages {
		t.Errorf("the build reported %d explorer pages but wrote %d index files", res.Pages, pages)
	}
}

// TestExplorerCoversEveryRecord is the promise the explorer makes: a page for
// every record in the tree, reachable from the list twin.
func TestExplorerCoversEveryRecord(t *testing.T) {
	f := newFixture(t)
	out := t.TempDir()
	res := buildFixture(t, f, out)

	for _, id := range []string{"adr-1", "adr-2", "adr-3", "itd-1", "itd-2", "itd-3", "itd-4",
		"itd-5", "itd-6", "spc-1", "iss-1", "one-fixture-principle", "a-second-fixture-principle"} {
		found := false
		for _, file := range res.Files {
			if strings.Contains(file, "/"+id+"/index.html") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s has no page", id)
		}
	}

	graph := outFile(t, out, "record/graph/index.html")
	for _, want := range []string{
		`href="/record/adr/adr-1/"`,
		`href="/record/intent/itd-5/"`,
		`href="/record/principle/one-fixture-principle/"`,
		`href="/record/issue/iss-1/"`,
	} {
		if !strings.Contains(graph, want) {
			t.Errorf("the list twin does not reach %s", want)
		}
	}
	// The list twin carries the LINKS as well as the records, so a keyboard
	// visitor reaches everything the chart draws without the chart running.
	if !strings.Contains(graph, "implements spc-1") && !strings.Contains(graph, "implemented by spc-1") {
		t.Error("the list twin does not carry the intent↔spec link")
	}
}

// TestRecordPageRendersItsBodyAndLinks is the per-record page's contract: the
// frontmatter, the body verbatim, the links phrased from this record's side, and
// the two forge links.
func TestRecordPageRendersItsBodyAndLinks(t *testing.T) {
	f := newFixture(t)
	out := t.TempDir()
	buildFixture(t, f, out)

	page := outFile(t, out, "record/intent/itd-2/index.html")
	for _, want := range []string{
		// the record's own body, verbatim
		"A fixture user opens the page and sees the record.",
		// frontmatter, and the dates git carries for the file
		`<th scope="row">id</th><td>itd-2</td>`,
		`<th scope="row">lifecycle</th><td>shipped</td>`,
		`<th scope="row">entered</th><td>2026-02-10</td>`,
		// the spec link, read from the intent's side
		`implemented by`,
		`href="/record/spec/spc-1/"`,
		// the forge links: the file, and its commit history
		`https://example.invalid/fixture/repo/blob/main/.abcd/development/intents/shipped/itd-2-the-shipped-one.md`,
		`https://example.invalid/fixture/repo/commits/main/.abcd/development/intents/shipped/itd-2-the-shipped-one.md`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("itd-2's page is missing %q", want)
		}
	}

	// The spec's own page phrases the same link the other way round.
	spec := outFile(t, out, "record/spec/spc-1/index.html")
	if !strings.Contains(spec, `implements`) || !strings.Contains(spec, `href="/record/intent/itd-2/"`) {
		t.Error("the spec's page does not phrase its link from its own side")
	}
}

// TestRecordPageRendersARetiredTargetAsAStub is the interview's ruling: a
// reference whose target has left the tree renders as the absence it is —
// dashed, labelled and UNLINKED — never as a link to a page that does not exist.
func TestRecordPageRendersARetiredTargetAsAStub(t *testing.T) {
	f := newFixture(t)
	out := t.TempDir()
	buildFixture(t, f, out)

	page := outFile(t, out, "record/adr/adr-2/index.html")
	if !strings.Contains(page, `<span class="id stub">adr-9</span>`) {
		t.Error("adr-2's dangling supersession does not render as a dashed stub")
	}
	if strings.Contains(page, `href="/record/adr/adr-9/"`) {
		t.Error("a reference to a record that is not in the tree rendered as a link")
	}
	// The dashboard says the same thing, measured against the ratchet.
	dash := outFile(t, out, "record/index.html")
	if !strings.Contains(dash, `<b class="stub">adr-9</b>`) {
		t.Error("record health does not name the unresolved reference")
	}
}

// TestTimelineIsDeterministicStaticSVG pins the genealogy as what it claims to
// be: one SVG emitted at build time, with a lane per store, links into the
// chart, and dashed stubs where a target has left the tree.
func TestTimelineIsDeterministicStaticSVG(t *testing.T) {
	f := newFixture(t)
	out := t.TempDir()
	buildFixture(t, f, out)
	page := outFile(t, out, "record/timeline/index.html")

	for _, want := range []string{
		`<svg viewBox="0 0 1000 `,
		// every mark opens its record in the relationship chart
		`href="/record/graph/?focus=adr-1"`,
		// releases tick to their own release on the forge
		`https://example.invalid/fixture/repo/releases/tag/v0.2.0`,
		// a supersession target that has left the tree is a dashed ×-stub
		`class="tlstub"`,
		`× adr-9`,
		// the issue ledger is a per-day histogram, not one dot per issue
		`class="tlbar"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the genealogy is missing %q", want)
		}
	}
	// No arc may point at a record the tree does not carry.
	if strings.Contains(page, `Q `) && strings.Contains(page, `adr-9"`) {
		t.Error("an arc was drawn to a record that has no position")
	}
}

// TestDashboardVisualsHaveTableTwins is the assistive-technology rule: every
// picture on the dashboard is accompanied by the same numbers in a table.
func TestDashboardVisualsHaveTableTwins(t *testing.T) {
	f := newFixture(t)
	out := t.TempDir()
	buildFixture(t, f, out)
	page := outFile(t, out, "record/index.html")

	bars := strings.Count(page, `<div class="seg" role="presentation">`)
	svgs := strings.Count(page, `class="cadsvg"`)
	twins := strings.Count(page, `<details class="twin">`)
	if bars == 0 {
		t.Fatal("the dashboard drew no lifecycle bars")
	}
	if twins != bars+svgs {
		t.Errorf("%d visuals, %d table twins — every visual needs one", bars+svgs, twins)
	}
}

// TestFoundationsListsAndLinks is the ruling on that page: it lists each
// principle and discipline as a card that LINKS its record, and never explains
// one — the explanation is the page the card opens.
func TestFoundationsListsAndLinks(t *testing.T) {
	f := newFixture(t)
	out := t.TempDir()
	buildFixture(t, f, out)
	page := outFile(t, out, "record/foundations/index.html")

	for _, want := range []string{
		`href="/record/principle/one-fixture-principle/"`,
		`href="/record/principle/a-second-fixture-principle/"`,
		`href="/record/intent/itd-5/"`,
		`One fixture principle`,
		`A Fixture Discipline`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the foundations page is missing %q", want)
		}
	}
	// The store's own index is not one of its records.
	if strings.Contains(page, "record/principle/README/") {
		t.Error("the principle store's README rendered as a principle")
	}
	// Lists and links, never explains: no card carries the body of its record.
	if strings.Contains(page, "the lint scan cannot see") {
		t.Error("a foundations card explained its record instead of linking it")
	}
}

// TestContributorsSeparatesAuthorshipFromDisclosure pins the page's whole point:
// humans are the authors of record, the trailer tallies are disclosure, and the
// policy that requires them is quoted beside the number.
func TestContributorsSeparatesAuthorshipFromDisclosure(t *testing.T) {
	f := newFixture(t)
	out := t.TempDir()
	buildFixture(t, f, out)
	page := outFile(t, out, "contributors/index.html")

	for _, want := range []string{
		"Authors of record",
		"Fixture",
		"Assisted-by trailers",
		"assistant-model-1",
		// the policy span the manifest selects, with its provenance and its file
		`data-src="CONTRIBUTING.md#attribution"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the contributors page is missing %q", want)
		}
	}
	// Model names are confined to this page under the attribution escape.
	for _, other := range []string{"record/index.html", "record/graph/index.html", "index.html"} {
		if strings.Contains(outFile(t, out, other), "assistant-model-1") {
			t.Errorf("%s names a model outside the attribution escape", other)
		}
	}
}

// TestReferencesRenderFromCSL pins the in-repo formatter: the sources render in
// the CSL file's own order, with a DOI or a URL linked, beside the credited
// inspirations under the acknowledgement file's own heading.
func TestReferencesRenderFromCSL(t *testing.T) {
	f := newFixture(t)
	out := t.TempDir()
	buildFixture(t, f, out)
	page := outFile(t, out, "references/index.html")

	for _, want := range []string{
		"References &amp; sources",
		"Fenella Quill. 2019. On the making of fixtures.",
		`<a href="https://doi.org/10.5555/fixture.2019.4">doi:10.5555/fixture.2019.4</a>`,
		"Orrin Tamsin and Iris Vole. 2021.",
		`<a href="https://example.invalid/rendering-the-record">`,
		"Inspirations",
		"A first invented inspiration",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the references page is missing %q", want)
		}
	}
	// The numbering is the citation key, so the order is the file's order.
	quill := strings.Index(page, "Fenella Quill")
	tamsin := strings.Index(page, "Orrin Tamsin")
	if quill < 0 || tamsin < 0 || quill > tamsin {
		t.Error("the sources are not rendered in the bibliography's own order")
	}
}

// TestNumberingDisagreementFailsTheBuild is the check the references page rests
// on. The record cites its sources BY NUMBER, so a page that numbered them
// differently would silently point every citation at the wrong source.
func TestNumberingDisagreementFailsTheBuild(t *testing.T) {
	cases := []struct {
		name, ack, says string
	}{
		{
			name: "a source is missing from the numbered list",
			ack: `# Acknowledgements

## References & sources

1. Fenella Quill. 2019. On the making of fixtures.
`,
			says: "entry for entry",
		},
		{
			name: "the two lists are in different orders",
			ack: `# Acknowledgements

## References & sources

1. Orrin Tamsin and Iris Vole. 2021. Rendering the record.
2. Fenella Quill. 2019. On the making of fixtures.
`,
			says: "would point at the wrong source",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t)
			f.write("ACKNOWLEDGEMENTS.md", c.ack)
			_, err := Build(Request{RepoRoot: f.Root(), OutDir: t.TempDir(), Stamp: fixtureStamp})
			if err == nil {
				t.Fatal("a bibliography numbered differently from the acknowledgement list must fail the build")
			}
			if !strings.Contains(err.Error(), c.says) {
				t.Errorf("the refusal does not say %q: %v", c.says, err)
			}
			if !strings.Contains(err.Error(), "ACKNOWLEDGEMENTS.md") {
				t.Errorf("the refusal does not name the file: %v", err)
			}
		})
	}
}

// --- graceful absence -----------------------------------------------------
//
// itd-140's rule, three times over: a missing optional source omits the page it
// feeds AND that page's navigation entry, and the build succeeds.

// TestBuildWithoutBibliography omits `/references/` rather than shipping a
// broken or half-rendered page.
func TestBuildWithoutBibliography(t *testing.T) {
	f := newFixture(t)
	if err := os.Remove(filepath.Join(f.Root(),
		filepath.FromSlash(".abcd/development/research/references.csl.json"))); err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	res, err := Build(Request{RepoRoot: f.Root(), OutDir: out, Stamp: fixtureStamp})
	if err != nil {
		t.Fatalf("a repository with no bibliography must still build: %v", err)
	}
	if containsString(res.Files, "references/index.html") {
		t.Error("the references page rendered with no bibliography to render")
	}
	dash := outFile(t, out, "record/index.html")
	if strings.Contains(dash, `href="/references/"`) {
		t.Error("the navigation points at a references page the build did not write")
	}
	if !strings.Contains(dash, `href="/record/graph/"`) {
		t.Error("the rest of the navigation went with it")
	}
}

// TestBuildWithoutFoundations omits the foundations page when the repository
// declares neither principles nor disciplines.
func TestBuildWithoutFoundations(t *testing.T) {
	f := newFixture(t)
	for _, dir := range []string{
		".abcd/development/principles",
		".abcd/development/intents/disciplines",
	} {
		if err := os.RemoveAll(filepath.Join(f.Root(), filepath.FromSlash(dir))); err != nil {
			t.Fatal(err)
		}
	}
	out := t.TempDir()
	res, err := Build(Request{RepoRoot: f.Root(), OutDir: out, Stamp: fixtureStamp})
	if err != nil {
		t.Fatalf("a repository with no principles or disciplines must still build: %v", err)
	}
	if containsString(res.Files, "record/foundations/index.html") {
		t.Error("the foundations page rendered with nothing to found on")
	}
	dash := outFile(t, out, "record/index.html")
	if strings.Contains(dash, `href="/record/foundations/"`) {
		t.Error("the navigation points at a foundations page the build did not write")
	}
	if strings.Contains(dash, `principles`) {
		t.Error("the dashboard captioned a principle tile with no principles")
	}
}

// TestExplorerWithoutChangelog drops the release lane, the cadence panel and the
// release tiles, and keeps every page that does not depend on them.
func TestExplorerWithoutChangelog(t *testing.T) {
	f := newFixture(t)
	if err := os.Remove(filepath.Join(f.Root(), "CHANGELOG.md")); err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	res, err := Build(Request{RepoRoot: f.Root(), OutDir: out,
		Stamp: BuildStamp{Commit: "abcdef1", GeneratedAt: "2026-02-11"}})
	if err != nil {
		t.Fatalf("a repository with no changelog must still build its explorer: %v", err)
	}
	if !containsString(res.Files, "record/timeline/index.html") {
		t.Fatal("the genealogy vanished with the changelog")
	}
	dash := outFile(t, out, "record/index.html")
	if strings.Contains(dash, "Release cadence") {
		t.Error("the cadence panel rendered with no releases")
	}
	tl := outFile(t, out, "record/timeline/index.html")
	if strings.Contains(tl, "releases/tag/") {
		t.Error("the genealogy drew a release lane with no releases")
	}
	if !strings.Contains(tl, `href="/record/graph/?focus=adr-1"`) {
		t.Error("the rest of the genealogy went with the release lane")
	}
}

// TestIssueLedgerIsOptIn is adr-32's tiering at the site: the working-tier
// ledger is published only because this repository asks for it, and a
// repository that does not ask gets no ledger pages at all.
func TestIssueLedgerIsOptIn(t *testing.T) {
	f := newFixture(t)
	manifest := outFile(t, f.Root(), ".abcd/site.json")
	f.write(".abcd/site.json", strings.Replace(manifest,
		`"record": {"issue_ledger": true}`, `"record": {"issue_ledger": false}`, 1))
	out := t.TempDir()
	res, err := Build(Request{RepoRoot: f.Root(), OutDir: out, Stamp: fixtureStamp})
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range res.Files {
		if strings.Contains(file, "record/issue/") {
			t.Errorf("the ledger was published without the opt-in: %s", file)
		}
	}
	if strings.Contains(outFile(t, out, "record/index.html"), "iss-1") {
		t.Error("the dashboard names a ledger entry the repository did not opt in to publishing")
	}
}

// outFile reads one file out of a build tree.
func outFile(t *testing.T, out, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
