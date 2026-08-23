package site

import (
	"strings"
	"testing"

	"github.com/Partnermedia/abcd/internal/core/changelog"
)

// The header and footer forge links are labelled with the forge's declared
// interface name, not the owner/repo handle — a reader who has never heard of
// the repository still knows where the link goes. A forge no name is declared
// for keeps the handle: the generic fallback the fixture forge exercises.
func TestForgeLabel(t *testing.T) {
	cases := []struct {
		name  string
		repo  string
		names map[string]string
		want  string
	}{
		{"declared forge name wins", "https://github.com/Partnermedia/abcd",
			map[string]string{"github.com": "GitHub"}, "GitHub"},
		{"undeclared forge falls back to the handle", "https://example.invalid/fixture/repo",
			map[string]string{"github.com": "GitHub"}, "fixture/repo"},
		{"no map at all falls back to the handle", "https://github.com/Partnermedia/abcd",
			nil, "Partnermedia/abcd"},
		{"blank declared name falls back to the handle", "https://github.com/Partnermedia/abcd",
			map[string]string{"github.com": "  "}, "Partnermedia/abcd"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &composer{}
			c.repo.Repository = tc.repo
			c.ui.ForgeNames = tc.names
			if got := c.forgeLabel(); got != tc.want {
				t.Fatalf("forgeLabel(%q) = %q, want %q", tc.repo, got, tc.want)
			}
		})
	}
}

// profileURL derives a forge profile only from a noreply address — both
// GitHub forms — and derives nothing from a real mailbox, which the export
// rule protects. A bot's bracketed name never matches.
func TestProfileURL(t *testing.T) {
	cases := map[string]string{
		"77722411+REPPL@users.noreply.github.com": "https://github.com/REPPL",
		"REPPL@users.noreply.github.com":          "https://github.com/REPPL",
		"someone@example.com":                     "",
		"":                                        "",
		"49699333+dependabot[bot]@users.noreply.github.com": "",
	}
	for in, want := range cases {
		if got := profileURL(in); got != want {
			t.Fatalf("profileURL(%q) = %q, want %q", in, got, want)
		}
	}
}

// uiStrings walks MAPS as well as fields. The provenance gate holds every
// rendered word to this list, so a family of declared interface strings the
// walk cannot see is refused on the page that renders it — which is how the
// forge names were refused when the walk knew only structs and strings.
func TestUIStringsIncludeMapValues(t *testing.T) {
	ui := UI{ForgeNames: map[string]string{"github.com": "GitHub"}}
	var found bool
	for _, s := range uiStrings(ui) {
		if s == "GitHub" {
			found = true
		}
		if s == "github.com" {
			t.Error("a map KEY reached the allowlist; only the values are rendered")
		}
	}
	if !found {
		t.Error("a declared forge name is not in the interface-string allowlist")
	}
}

// A key declared blank is named, so it cannot pass for a repository that
// simply chose not to declare it.
func TestUIMissingNamesABlankMapValue(t *testing.T) {
	ui := UI{ForgeNames: map[string]string{"github.com": "  "}}
	var named bool
	for _, m := range ui.missing() {
		if m == "forge_names.github.com" {
			named = true
		}
	}
	if !named {
		t.Errorf("a blank forge name went unnamed: %v", ui.missing())
	}
	if len(UI{ForgeNames: map[string]string{}}.missing()) != len(UI{}.missing()) {
		t.Error("an empty map was treated as a missing declaration")
	}
}

// The cadence ridgeline draws one band per release over the commits made in
// that release's own window. It replaced a tick strip whose date labels
// collided when two releases fell days apart; a band carries its own label
// column, so crowding is no longer expressible.
func TestCadenceDrawsABandPerRelease(t *testing.T) {
	build := func(rel []changelog.DatedRelease, days map[string]int) string {
		e := &explorer{c: &composer{ui: UI{Panels: Panels{Cadence: "Release cadence"},
			Tiles: Tiles{Releases: "releases"}}}}
		e.export.Releases = rel
		e.export.History.Days = days
		return e.cadence()
	}
	rel := []changelog.DatedRelease{
		{Version: "0.3.0", Date: "2026-01-20"},
		{Version: "0.2.0", Date: "2026-01-10"},
		{Version: "0.1.0", Date: "2026-01-05"},
	}
	days := map[string]int{
		"2026-01-02": 3, "2026-01-04": 1, // before the first release
		"2026-01-07": 9, "2026-01-09": 2, // the second release's window
		"2026-01-15": 4, "2026-01-20": 1, // the third's
	}
	out := build(rel, days)

	// Every release is a row, labelled with its version and its date, and the
	// commits of its own window are totalled beside it.
	for _, want := range []string{">v0.1.0<", ">v0.2.0<", ">v0.3.0<",
		">2026-01-05<", ">2026-01-10<", ">2026-01-20<"} {
		if !strings.Contains(out, want) {
			t.Errorf("the ridgeline does not label %s", want)
		}
	}
	if n := strings.Count(out, `<g class="tick">`); n != len(rel) {
		t.Errorf("the ridgeline drew %d bands for %d releases", n, len(rel))
	}
	// One filled band per release whose window holds commits.
	if n := strings.Count(out, `<path d="M `); n != len(rel) {
		t.Errorf("the ridgeline drew %d shapes for %d releases", n, len(rel))
	}
	// The oldest release owns everything before it, so the four commits that
	// led to v0.1.0 are drawn rather than dropped.
	if !strings.Contains(out, ">4<") {
		t.Error("the first release's window does not carry the work that preceded it")
	}
	// The narrow-screen list stands in for the drawing and carries every row.
	if n := strings.Count(out, `<li><span class="id">v`); n != len(rel) {
		t.Errorf("the narrow-screen list drew %d rows for %d releases", n, len(rel))
	}

	// Without a commit history there is nothing to draw a shape from, and the
	// panel is omitted rather than rendered empty (itd-140: graceful absence).
	if got := build(rel, nil); got != "" {
		t.Errorf("a repository with no dated history still drew a ridgeline: %q", got)
	}
	// One release is not a cadence.
	if got := build(rel[:1], days); got != "" {
		t.Errorf("a single release drew a cadence: %q", got)
	}
}

// forgeHost strips exactly the scheme and path; a bare or schemeless value
// still yields its host.
func TestForgeHost(t *testing.T) {
	cases := map[string]string{
		"https://github.com/owner/repo": "github.com",
		"http://example.invalid/f/r":    "example.invalid",
		"github.com/owner/repo":         "github.com",
		"":                              "",
	}
	for in, want := range cases {
		if got := forgeHost(in); got != want {
			t.Fatalf("forgeHost(%q) = %q, want %q", in, got, want)
		}
	}
}
