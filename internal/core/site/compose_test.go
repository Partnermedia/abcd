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

// cadenceDates pins the collision branch in both directions: releases far
// apart each keep their date, and releases days apart do not print two dates
// over each other. The fixture's own two releases sit a third of the chart
// apart, so only this test can see the suppression.
func TestCadenceSuppressesOnlyACollidingDate(t *testing.T) {
	dates := func(rel []changelog.DatedRelease) (versions, printed int) {
		e := &explorer{c: &composer{ui: UI{Panels: Panels{Cadence: "Release cadence"},
			Tiles: Tiles{Releases: "releases"}}}}
		e.export.Releases = rel
		out := e.cadence()
		return strings.Count(out, `font-size="10"`), strings.Count(out, `font-size="8"`)
	}
	// Two months apart: both dates fit.
	spread := []changelog.DatedRelease{{Version: "0.2.0", Date: "2026-03-01"}, {Version: "0.1.0", Date: "2026-01-01"}}
	if v, d := dates(spread); v != 2 || d != 2 {
		t.Errorf("well-spaced releases printed %d versions and %d dates, want 2 and 2", v, d)
	}
	// A cluster inside a long span — the shape that collides, since the chart
	// normalises to the span and a week of releases on its own simply spreads
	// to full width. Three of these land on the same side of the axis with the
	// last two ~38 units apart, inside a date label's own width.
	tight := []changelog.DatedRelease{
		{Version: "0.5.0", Date: "2026-02-13"}, {Version: "0.4.0", Date: "2026-02-11"},
		{Version: "0.3.0", Date: "2026-02-10"}, {Version: "0.2.0", Date: "2026-01-21"},
		{Version: "0.1.0", Date: "2026-01-01"},
	}
	v, d := dates(tight)
	if v != 5 {
		t.Errorf("a crowded chart printed %d versions, want 5 — a version is never suppressed", v)
	}
	if d >= v {
		t.Errorf("a crowded chart printed %d dates against %d versions; a colliding date must be given up", d, v)
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
