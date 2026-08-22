package site

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden files from this run's output")

// fixtureStamp is the injected build metadata. Every field is fixed so the
// golden files are a function of the fixture and nothing else — no clock, no
// commit hash, no version resolved at run time.
var fixtureStamp = BuildStamp{Version: "0.2.0", Commit: "abcdef1", GeneratedAt: "2026-02-11"}

// buildFixture renders the fixture repository into a fresh directory.
func buildFixture(t *testing.T, f *fixture, out string) Result {
	t.Helper()
	res, err := Build(Request{RepoRoot: f.Root(), OutDir: out, Stamp: fixtureStamp})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	return res
}

// golden compares one output file against its committed copy.
func golden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s: %v (run `go test ./internal/core/site/ -update` to create it)", path, err)
	}
	if string(got) == string(want) {
		return
	}
	t.Errorf("%s differs from the golden file; run `go test ./internal/core/site/ -update` and read the diff", path)
	for i, line := range diffLines(string(want), string(got)) {
		if i > 12 {
			t.Errorf("  … and more")
			break
		}
		t.Errorf("  %s", line)
	}
}

// diffLines names the first lines that differ, so a failure says what changed
// rather than dumping two files.
func diffLines(want, got string) []string {
	w := strings.Split(want, "\n")
	g := strings.Split(got, "\n")
	var out []string
	for i := 0; i < len(w) || i < len(g); i++ {
		var a, b string
		if i < len(w) {
			a = w[i]
		}
		if i < len(g) {
			b = g[i]
		}
		if a == b {
			continue
		}
		out = append(out, "line "+itoa(i+1)+" want: "+clip120(a))
		out = append(out, "line "+itoa(i+1)+"  got: "+clip120(b))
		if len(out) > 26 {
			break
		}
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func clip120(s string) string {
	if len(s) > 120 {
		return s[:120] + "…"
	}
	return s
}

// TestBuildGolden pins the whole render: the composed landing page and the
// record export, byte for byte, over a repository this test built.
func TestBuildGolden(t *testing.T) {
	f := newFixture(t)
	out := t.TempDir()
	res := buildFixture(t, f, out)

	for _, name := range []string{"index.html", "record.json", "_redirects", "_headers", "site.css", "site.js"} {
		if !containsString(res.Files, name) {
			t.Errorf("build did not write %s: %v", name, res.Files)
		}
	}
	for _, name := range []string{"assets/img/intro.png", "assets/img/logo.png",
		"assets/img/role-thinker.png", "assets/img/role-facilitator.png"} {
		if !containsString(res.Files, name) {
			t.Errorf("build did not copy %s: %v", name, res.Files)
		}
	}

	html, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "index.html", html)

	rec, err := os.ReadFile(filepath.Join(out, "record.json"))
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "record.json", rec)
}

// TestBuildLayoutDoesNotOverlap is the coil packing's own sanity check, asserted
// rather than reported: a bubble sitting on top of another means the placement
// walked into a case it does not handle, and the chart is wrong.
func TestBuildLayoutDoesNotOverlap(t *testing.T) {
	f := newFixture(t)
	res := buildFixture(t, f, t.TempDir())
	if res.Overlaps != 0 {
		t.Fatalf("coil packing overlaps: %d", res.Overlaps)
	}
}

// TestBuildIsDeterministic builds the same tree twice into two directories and
// diffs every byte. It is the property the published site rests on: production
// is rendered from a tag, and a build that is not a function of its input cannot
// be checked against the tree it claims to describe.
func TestBuildIsDeterministic(t *testing.T) {
	f := newFixture(t)
	a, b := t.TempDir(), t.TempDir()
	ra := buildFixture(t, f, a)
	rb := buildFixture(t, f, b)
	if len(ra.Files) != len(rb.Files) {
		t.Fatalf("two builds wrote different files: %v vs %v", ra.Files, rb.Files)
	}
	for _, name := range ra.Files {
		x, err := os.ReadFile(filepath.Join(a, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		y, err := os.ReadFile(filepath.Join(b, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		if string(x) != string(y) {
			t.Errorf("%s differs between two builds of the same tree", name)
		}
	}
}

// TestBuildRecordExportShape asserts the facts the export is FOR, so a change
// that silently drops one is caught by something other than the golden diff.
func TestBuildRecordExportShape(t *testing.T) {
	f := newFixture(t)
	out := t.TempDir()
	buildFixture(t, f, out)
	data, err := os.ReadFile(filepath.Join(out, "record.json"))
	if err != nil {
		t.Fatal(err)
	}
	var exp RecordExport
	if err := json.Unmarshal(data, &exp); err != nil {
		t.Fatal(err)
	}

	if exp.SchemaVersion != 1 {
		t.Errorf("schema_version: %d", exp.SchemaVersion)
	}
	if len(exp.Nodes) != 13 {
		t.Fatalf("nodes: %d, want 13 (3 adrs, 6 intents, 1 spec, 1 issue, 2 principles)", len(exp.Nodes))
	}
	// The principle store carries no frontmatter, so nothing in the lint scan
	// can see it; the file name is the handle and the H1 is the title. The
	// store's own README is its index, not one of its records.
	if got := exp.Counts.ByType["principle"]; got != 2 {
		t.Errorf("principles: %d, want 2 (the README is the store's index)", got)
	}
	byID := map[string]ExportNode{}
	for _, n := range exp.Nodes {
		byID[n.ID] = n
	}
	// The shipped intent was moved into shipped/ in its own commit; that move is
	// the only record of the day it shipped.
	if got := byID["itd-2"].Dates.Entered; got != "2026-02-10" {
		t.Errorf("itd-2 entered shipped/: %q, want 2026-02-10", got)
	}
	if got := byID["itd-2"].Dates.Created; got != "2026-01-05" {
		t.Errorf("itd-2 created: %q", got)
	}
	// A record with no frontmatter date is dated from git; one with a date keeps
	// its own.
	if got := byID["adr-1"].Date; got != "2026-01-02" {
		t.Errorf("adr-1 effective date: %q", got)
	}
	if got := byID["itd-2"].Date; got != "2026-01-05" {
		t.Errorf("itd-2 effective date: %q", got)
	}

	// The intent↔spec link is declared from both ends and must render once.
	implements := 0
	for _, e := range exp.Edges {
		if e.Rel == "implements" {
			implements++
			if e.From != "spc-1" || e.To != "itd-2" {
				t.Errorf("implements edge points the wrong way: %+v", e)
			}
		}
	}
	if implements != 1 {
		t.Errorf("implements edges: %d, want 1 (the mirrored pair collapses)", implements)
	}
	// So must the `related` pair adr-1 ↔ adr-2, which both files record.
	adrPair := 0
	for _, e := range exp.Edges {
		if e.Rel != "related" {
			continue
		}
		if (e.From == "adr-1" && e.To == "adr-2") || (e.From == "adr-2" && e.To == "adr-1") {
			adrPair++
		}
	}
	if adrPair != 1 {
		t.Errorf("adr-1 ↔ adr-2 related edges: %d, want 1 (both files record it)", adrPair)
	}

	// The dangling supersession is measured and listed, and it is the one the
	// committed baseline admits.
	if len(exp.Health.Unresolved) != 1 || exp.Health.Unresolved[0].To != "adr-9" {
		t.Errorf("unresolved: %+v", exp.Health.Unresolved)
	}
	if exp.Health.BaselineCount != 1 {
		t.Errorf("baseline count: %d", exp.Health.BaselineCount)
	}

	// A body mention is carried, and never where a typed link already is.
	if len(exp.Mentions) == 0 {
		t.Error("no body mentions recorded")
	}
	for _, m := range exp.Mentions {
		for _, e := range exp.Edges {
			if (m.From == e.From && m.To == e.To) || (m.From == e.To && m.To == e.From) {
				t.Errorf("mention duplicates a typed link: %+v", m)
			}
		}
	}

	if len(exp.Releases) != 2 || exp.Releases[0].Version != "0.2.0" || exp.Releases[0].Date != "2026-02-11" {
		t.Errorf("releases: %+v", exp.Releases)
	}
	// A record that entered the trunk only inside a merge commit is dated like
	// any other. `git log --name-status` prints no file lines for a merge, so
	// without --diff-merges this record is invisible to the walk entirely and
	// ships with three empty dates — which then takes the centre of the coil.
	adr3 := byID["adr-3"]
	if adr3.Dates.Created == "" || adr3.Dates.Entered == "" || adr3.Dates.Touched == "" {
		t.Errorf("adr-3 entered the trunk in a merge and the walk did not see it: %+v", adr3.Dates)
	}
	if adr3.Dates.Created != "2026-03-05" {
		t.Errorf("adr-3 created: %q, want 2026-03-05", adr3.Dates.Created)
	}
	for _, n := range exp.Nodes {
		if n.Date == "" {
			t.Errorf("%s carries no effective date; nothing can place it in time", n.ID)
		}
	}
	if exp.Layout.DateRange[0] == "" {
		t.Errorf("the published date span begins at the empty string: %v", exp.Layout.DateRange)
	}

	if exp.Counts.ByType["adr"] != 3 || exp.Counts.ByType["intent"] != 6 ||
		exp.Counts.ByType["spec"] != 1 || exp.Counts.ByType["issue"] != 1 {
		t.Errorf("counts: %+v", exp.Counts.ByType)
	}
	if exp.Counts.ByLifecycle["intent"]["shipped"] != 3 || exp.Counts.ByLifecycle["intent"]["drafts"] != 1 ||
		exp.Counts.ByLifecycle["intent"]["disciplines"] != 1 || exp.Counts.ByLifecycle["intent"]["superseded"] != 1 {
		t.Errorf("lifecycle counts: %+v", exp.Counts.ByLifecycle)
	}
	// A FLAT store grades its records in frontmatter rather than by moving them,
	// so its whole shape is in the status counts. A page reading only the
	// lifecycle would show the decisions as one undifferentiated block.
	if exp.Counts.ByStatus["adr"]["accepted"] != 3 {
		t.Errorf("status counts: %+v", exp.Counts.ByStatus)
	}

	// Attribution: one commit declared a model, one declared None, the rest
	// declared nothing at all — and the export says so rather than guessing.
	if exp.Authorship.Assisted != 1 {
		t.Errorf("assisted commits: %d, want 1", exp.Authorship.Assisted)
	}
	// The changelog commit, the merge fixture's branch and merge commits, and the
	// commit that ships the stub — each declaring that no tool touched it.
	if exp.Authorship.DeclaredNone != 4 {
		t.Errorf("declared-None commits: %d, want 4", exp.Authorship.DeclaredNone)
	}
	if len(exp.Authorship.Humans) != 1 || exp.Authorship.Humans[0].Name != "Fixture" {
		t.Errorf("humans: %+v", exp.Authorship.Humans)
	}
	if len(exp.Authorship.Bots) != 0 {
		t.Errorf("bots: %+v — the fixture's only author is a person", exp.Authorship.Bots)
	}
	if len(exp.Authorship.ByModel) != 2 {
		t.Errorf("by_model: %+v, want the declared model and the declared None", exp.Authorship.ByModel)
	}
}

// TestAuthorshipSeparatesToolsFromPeople pins the derived bots-and-tools row: a
// forge bot is recognised by the suffix the forge itself gives it, and a
// pre-policy commit authored by the tool is recognised because the repository's
// own trailers name that vendor as an assistant.
func TestAuthorshipSeparatesToolsFromPeople(t *testing.T) {
	f := newFixture(t)
	f.write("bot.txt", "a dependency bump\n")
	f.git("2026-03-03T09:00:00+00:00", "add", "-A")
	f.git("2026-03-03T09:00:00+00:00",
		"-c", "user.name=depbot[bot]", "-c", "user.email=depbot@example.invalid",
		"commit", "-m", "chore: bump a dependency")
	f.write("tool.txt", "a pre-policy commit\n")
	f.git("2026-03-04T09:00:00+00:00", "add", "-A")
	f.git("2026-03-04T09:00:00+00:00",
		"-c", "user.name=Assistant", "-c", "user.email=noreply@example.invalid",
		"commit", "-m", "chore: written before the trailer convention")

	a, err := LoadAuthorship(f.Root())
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, b := range a.Bots {
		names[b.Name] = true
	}
	if !names["depbot[bot]"] {
		t.Errorf("a forge bot is not in the bots row: %+v", a.Bots)
	}
	if !names["Assistant"] {
		t.Errorf("the pre-policy tool author is not in the bots row: %+v", a.Bots)
	}
	for _, h := range a.Humans {
		if names[h.Name] {
			t.Errorf("%q is in both rows", h.Name)
		}
	}
}

// TestBuildLandingCarriesProvenance asserts the property the single-source rule
// is checked through: every composed block names the file and heading it came
// from, and the only added words are ui.json's.
func TestBuildLandingCarriesProvenance(t *testing.T) {
	f := newFixture(t)
	out := t.TempDir()
	buildFixture(t, f, out)
	data, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)

	for _, want := range []string{
		`data-src=".abcd/development/brief/01-product/README.md#identity-canonical"`,
		`data-src="docs/explanation/rationale.md"`,
		`data-src="docs/explanation/roles.md#product-thinker"`,
		`data-src="docs/explanation/artefacts.md#artefacts"`,
		`data-src="docs/explanation/process.md#capturing-intents"`,
		`data-src="docs/how-to/install.md#macos"`,
		`data-src=".abcd/development/intents/shipped/itd-2-the-shipped-one.md#press-release"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("landing page carries no %s", want)
		}
	}

	// The feature block is derived: the shipped intent with a MET audit, not the
	// drafted one and not the one whose audit says otherwise.
	if !strings.Contains(html, "I stopped guessing") {
		t.Error("the feature block does not quote the shipped intent's press release")
	}
	if strings.Contains(html, "It has not been built") {
		t.Error("the feature block quotes a DRAFTED intent")
	}
	// itd-3 is shipped, audited MET, and newer than itd-2 — but its press
	// release is the mint placeholder. Quoting a template at a reader as this
	// page's one testimonial is worse than quoting nothing, so the derivation
	// falls through to the newest candidate that actually says something.
	// Both minted placeholders are skipped — the capture-seeded one (itd-3) and
	// the promotion-seeded one (itd-4), which is newer still.
	if strings.Contains(html, "Expand into the full press-release narrative") {
		t.Error("the feature block quotes an unwritten press release")
	}
	if strings.Contains(html, "Seeded by promotion from") {
		t.Error("the feature block quotes a promotion-seeded placeholder")
	}
	if !strings.Contains(html, "itd-2") {
		t.Error("the feature block did not fall through to the newest written press release")
	}
	// Exactly one acceptance criterion, as the manifest asks.
	if strings.Contains(html, "it does not appear in the quote") {
		t.Error("the feature block quotes more than the first acceptance criterion")
	}

	// The Beta badge is a rule on the release version, not a copy decision.
	if !strings.Contains(html, `<span class="beta">Beta</span>`) {
		t.Error("no Beta badge at a 0.x release")
	}
	// The SVG is inlined so its var(--token) colours follow the theme; the
	// raster is referenced from the copied file.
	if !strings.Contains(html, `<span class="svgasset artefact-brief"`) || !strings.Contains(html, "var(--ink, #000)") {
		t.Error("the SVG asset was not inlined")
	}
	if strings.Contains(html, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10" width="10"`) {
		t.Error("the inlined SVG kept its width attribute; the stylesheet must size it")
	}
	// Root-absolute: the same picture is referenced from `/` and from
	// `/record/<type>/<id>/`, and a relative src resolves at only one of them.
	if !strings.Contains(html, `<img src="/assets/img/intro.png"`) {
		t.Error("the raster was not referenced from the output tree at its served path")
	}
	// The install tab row: the manifest's left-hand section, then the CLI group.
	if !strings.Contains(html, `<button role="tab" id="tab-plugin" aria-selected="true"`) {
		t.Error("the Plugin tab is not the open one")
	}
	if !strings.Contains(html, `<div class="tabgroup"><span class="grp">CLI</span>`) {
		t.Error("the CLI group is not labelled from ui.json")
	}
	if !strings.Contains(html, `data-copied="copied"`) {
		t.Error("the copy button carries no ui.json label for its done state")
	}
}

// TestBuildWithoutChangelog is the graceful-absence rule (itd-140): a missing
// optional source omits what depends on it and the build still succeeds.
func TestBuildWithoutChangelog(t *testing.T) {
	f := newFixture(t)
	if err := os.Remove(filepath.Join(f.Root(), "CHANGELOG.md")); err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	// No GeneratedAt either: with no releases there is no date to fall back on,
	// which is the path that actually runs when a repository has no changelog.
	res, err := Build(Request{RepoRoot: f.Root(), OutDir: out,
		Stamp: BuildStamp{Commit: "abcdef1"}})
	if err != nil {
		t.Fatalf("a repository with no CHANGELOG must still build: %v", err)
	}
	if res.Version != "" {
		t.Errorf("version without a changelog: %q", res.Version)
	}
	data, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)
	if strings.Contains(html, `class="beta"`) {
		t.Error("the Beta badge rendered with no release to read a major version from")
	}
	if strings.Contains(html, `class="pill"`) {
		t.Error("the release pill rendered with no release")
	}
	// The copyright year comes from the build stamp; with nothing to date the
	// build from, the span is omitted rather than rendered without its year.
	if strings.Contains(html, "©") {
		t.Error("the footer rendered a copyright line with no year to put in it")
	}
	var exp RecordExport
	rec, err := os.ReadFile(filepath.Join(out, "record.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(rec, &exp); err != nil {
		t.Fatal(err)
	}
	if len(exp.Releases) != 0 {
		t.Errorf("releases without a changelog: %+v", exp.Releases)
	}
	if len(exp.Nodes) == 0 {
		t.Error("the record vanished with the changelog")
	}
}

// TestBuildWithoutIdentityBlock is the same rule at the hero: the three spans
// the Identity block supplies are omitted, the headline and lede that carry the
// page stay, and the build succeeds.
func TestBuildWithoutIdentityBlock(t *testing.T) {
	f := newFixture(t)
	f.write(".abcd/development/brief/01-product/README.md", "# Product\n\nNo identity block here.\n")
	out := t.TempDir()
	if _, err := Build(Request{RepoRoot: f.Root(), OutDir: out, Stamp: fixtureStamp}); err != nil {
		t.Fatalf("a repository with no Identity block must still build: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)
	if strings.Contains(html, `class="eyebrow" data-src`) || strings.Contains(html, `class="tagline"`) {
		t.Error("the hero rendered Identity spans with no Identity block to render them from")
	}
	if !strings.Contains(html, "<h1 data-src=\"docs/explanation/rationale.md\">Who this is for</h1>") {
		t.Error("the headline did not survive the missing Identity block")
	}
}

// TestBuildWithoutManifest refuses rather than rendering a page from nothing.
func TestBuildWithoutManifest(t *testing.T) {
	f := newFixture(t)
	if err := os.Remove(filepath.Join(f.Root(), ".abcd", "site.json")); err != nil {
		t.Fatal(err)
	}
	_, err := Build(Request{RepoRoot: f.Root(), OutDir: t.TempDir(), Stamp: fixtureStamp})
	if err == nil {
		t.Fatal("a repository declaring no composition must not silently build one")
	}
	if !strings.Contains(err.Error(), ManifestRelPath) {
		t.Errorf("the refusal does not name the manifest: %v", err)
	}
}

// TestBaselineComesFromTheManifest pins the thread the manifest key is supposed
// to be. Declaring `checks.unresolved_reference_baseline` and having the build
// read a hardcoded path anyway is the parsed-and-dropped failure in its most
// consequential form: the ratchet a repo thinks it armed is not the one being
// measured against.
func TestBaselineComesFromTheManifest(t *testing.T) {
	f := newFixture(t)
	manifest := filepath.Join(f.Root(), ".abcd", "site.json")
	original, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	repoint := func(t *testing.T, to string) {
		t.Helper()
		body := strings.Replace(string(original),
			`"unresolved_reference_baseline": ".abcd/site-baseline.json"`,
			`"unresolved_reference_baseline": "`+to+`"`, 1)
		if body == string(original) {
			t.Fatal("the fixture manifest does not declare a baseline")
		}
		if err := os.WriteFile(manifest, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.WriteFile(manifest, original, 0o644) })
	}

	t.Run("a named baseline that does not exist refuses", func(t *testing.T) {
		repoint(t, ".abcd/no-such-baseline.json")
		_, err := Build(Request{RepoRoot: f.Root(), OutDir: t.TempDir(), Stamp: fixtureStamp})
		if err == nil {
			t.Fatal("the build measured against a baseline the manifest names but the repo does not carry")
		}
		if !strings.Contains(err.Error(), "no-such-baseline.json") {
			t.Errorf("the refusal does not name the missing baseline: %v", err)
		}
	})

	// The board is documented as read-only and exit-0 whatever it finds. A
	// missing declared baseline is exactly the kind of thing somebody runs the
	// board to DISCOVER, so it has to be reportable state there — while staying
	// a refusal in `build`, which would otherwise publish a health count
	// measured against nothing.
	t.Run("the board reports a missing declared baseline rather than failing", func(t *testing.T) {
		repoint(t, ".abcd/no-such-baseline.json")
		st, err := Describe(f.Root(), "site")
		if err != nil {
			t.Fatalf("the status board failed on a missing baseline: %v", err)
		}
		if st.Baseline {
			t.Error("the board reports a baseline it could not read as present")
		}
		if st.BaselinePath != ".abcd/no-such-baseline.json" {
			t.Errorf("the board reports %q, not the declared path", st.BaselinePath)
		}
		if st.BaselineN != 0 {
			t.Errorf("the board counts %d entries from a baseline it could not read", st.BaselineN)
		}
	})

	t.Run("the board reports an unreadable baseline rather than failing", func(t *testing.T) {
		broken := ".abcd/broken-baseline.json"
		if err := os.WriteFile(filepath.Join(f.Root(), filepath.FromSlash(broken)), []byte("{not json"), 0o644); err != nil {
			t.Fatal(err)
		}
		repoint(t, broken)
		st, err := Describe(f.Root(), "site")
		if err != nil {
			t.Fatalf("the status board failed on an unreadable baseline: %v", err)
		}
		if st.Baseline || st.BaselineN != 0 {
			t.Errorf("the board treated an unreadable baseline as present: %+v", st)
		}
		if st.BaselinePath != broken {
			t.Errorf("the board reports %q, not the declared path", st.BaselinePath)
		}
	})

	t.Run("an alternate baseline is the one counted", func(t *testing.T) {
		alt := ".abcd/other-baseline.json"
		if err := os.WriteFile(filepath.Join(f.Root(), filepath.FromSlash(alt)), []byte(`{
  "schema_version": 1,
  "unresolved_references": [
    {"from": "adr-2", "to": "adr-9"},
    {"from": "itd-9", "to": "spc-9"},
    {"from": "itd-8", "to": "spc-8"}
  ]
}
`), 0o644); err != nil {
			t.Fatal(err)
		}
		repoint(t, alt)
		out := t.TempDir()
		res, err := Build(Request{RepoRoot: f.Root(), OutDir: out, Stamp: fixtureStamp})
		if err != nil {
			t.Fatal(err)
		}
		if res.Baseline != 3 {
			t.Errorf("baseline count: %d, want 3 — the manifest's baseline is the one measured against", res.Baseline)
		}
		st, err := Describe(f.Root(), "site")
		if err != nil {
			t.Fatal(err)
		}
		if st.BaselinePath != alt {
			t.Errorf("the status board reports %q, not the path it used (%q)", st.BaselinePath, alt)
		}
		if st.BaselineN != 3 {
			t.Errorf("the status board counts %d, want 3", st.BaselineN)
		}
	})
}

// TestDescribeIsReadOnly pins the bare verb: it reports and writes nothing.
func TestDescribeIsReadOnly(t *testing.T) {
	f := newFixture(t)
	st, err := Describe(f.Root(), "site")
	if err != nil {
		t.Fatal(err)
	}
	if !st.Manifest || !st.UIStrings || !st.Baseline {
		t.Errorf("status: %+v", st)
	}
	if st.Chapters != 4 {
		t.Errorf("chapters: %d", st.Chapters)
	}
	if st.OutExists {
		t.Error("status reports an output directory nothing built")
	}
	if _, err := os.Stat(filepath.Join(f.Root(), "site")); !os.IsNotExist(err) {
		t.Error("describing the site created the output directory")
	}
}

func containsString(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
