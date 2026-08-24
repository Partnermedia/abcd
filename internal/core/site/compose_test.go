package site

import "testing"

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
		{"declared forge name wins", "https://github.com/intentdriven/abcd",
			map[string]string{"github.com": "GitHub"}, "GitHub"},
		{"undeclared forge falls back to the handle", "https://example.invalid/fixture/repo",
			map[string]string{"github.com": "GitHub"}, "fixture/repo"},
		{"no map at all falls back to the handle", "https://github.com/intentdriven/abcd",
			nil, "intentdriven/abcd"},
		{"blank declared name falls back to the handle", "https://github.com/intentdriven/abcd",
			map[string]string{"github.com": "  "}, "intentdriven/abcd"},
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
