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
