package positioning

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// writeRepo materialises a throwaway repo tree from a path->content map and
// returns its root.
func writeRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const briefLike = `# Product

## Identity (canonical)

The single recorded home for how the project names itself.

- **Title:** abcd — Agent-Based Configuration for Development
- **Tagline:** A host-agnostic configuration layer for intent-driven development.
- **Pitch:** A single Go binary that carries the why from idea to shipped
  reality, usable as a plugin in compatible agent harnesses.

## Something else

- **Title:** not the identity block
`

func TestParseBlockReadsTheThreeBullets(t *testing.T) {
	root := writeRepo(t, map[string]string{"brief.md": briefLike})

	got, err := ParseBlock(root, BlockLocation{File: "brief.md", Heading: "Identity (canonical)"})
	if err != nil {
		t.Fatalf("ParseBlock: %v", err)
	}
	if want := "abcd — Agent-Based Configuration for Development"; got.Title != want {
		t.Errorf("Title = %q, want %q", got.Title, want)
	}
	if want := "A host-agnostic configuration layer for intent-driven development."; got.Tagline != want {
		t.Errorf("Tagline = %q, want %q", got.Tagline, want)
	}
	// The pitch bullet wraps across two lines; the parser joins the continuation.
	if want := "A single Go binary that carries the why from idea to shipped reality, usable as a plugin in compatible agent harnesses."; got.Pitch != want {
		t.Errorf("Pitch = %q, want %q", got.Pitch, want)
	}
	if got.File != "brief.md" {
		t.Errorf("File = %q", got.File)
	}
	if got.Line != 3 {
		t.Errorf("Line = %d, want 3 (the heading line)", got.Line)
	}
}

func TestParseBlockStopsAtTheNextHeading(t *testing.T) {
	root := writeRepo(t, map[string]string{"brief.md": briefLike})
	got, err := ParseBlock(root, BlockLocation{File: "brief.md", Heading: "Identity (canonical)"})
	if err != nil {
		t.Fatalf("ParseBlock: %v", err)
	}
	if got.Title == "not the identity block" {
		t.Fatal("parser bled past the section boundary into the next heading")
	}
}

func TestParseBlockPitchIsOptional(t *testing.T) {
	root := writeRepo(t, map[string]string{"brief.md": `## Identity

- **Title:** Widget
- **Tagline:** Widgets, but sharper.
`})
	got, err := ParseBlock(root, BlockLocation{File: "brief.md", Heading: "Identity"})
	if err != nil {
		t.Fatalf("ParseBlock: %v", err)
	}
	if got.Pitch != "" {
		t.Errorf("Pitch = %q, want empty", got.Pitch)
	}
}

func TestParseBlockHeadingMatchIsCaseInsensitiveAndLevelAgnostic(t *testing.T) {
	root := writeRepo(t, map[string]string{"brief.md": `#### IDENTITY (Canonical)

- **Title:** Widget
- **Tagline:** Widgets, but sharper.
`})
	if _, err := ParseBlock(root, BlockLocation{File: "brief.md", Heading: "identity (canonical)"}); err != nil {
		t.Fatalf("ParseBlock: %v", err)
	}
}

func TestParseBlockTypedErrors(t *testing.T) {
	cases := []struct {
		name    string
		files   map[string]string
		loc     BlockLocation
		wantErr error
	}{
		{
			name:    "file absent",
			files:   map[string]string{"other.md": "# x\n"},
			loc:     BlockLocation{File: "brief.md", Heading: "Identity"},
			wantErr: ErrBlockFileMissing,
		},
		{
			name:    "heading absent",
			files:   map[string]string{"brief.md": "# Product\n\nnothing here\n"},
			loc:     BlockLocation{File: "brief.md", Heading: "Identity"},
			wantErr: ErrHeadingMissing,
		},
		{
			name:    "title bullet missing",
			files:   map[string]string{"brief.md": "## Identity\n\n- **Tagline:** t\n"},
			loc:     BlockLocation{File: "brief.md", Heading: "Identity"},
			wantErr: ErrBulletMissing,
		},
		{
			name:    "tagline bullet missing",
			files:   map[string]string{"brief.md": "## Identity\n\n- **Title:** t\n"},
			loc:     BlockLocation{File: "brief.md", Heading: "Identity"},
			wantErr: ErrBulletMissing,
		},
		{
			name:    "title bullet present but empty",
			files:   map[string]string{"brief.md": "## Identity\n\n- **Title:**\n- **Tagline:** t\n"},
			loc:     BlockLocation{File: "brief.md", Heading: "Identity"},
			wantErr: ErrBulletMissing,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := writeRepo(t, tc.files)
			_, err := ParseBlock(root, tc.loc)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestParseBlockRefusesEscapingLocation(t *testing.T) {
	root := writeRepo(t, map[string]string{"brief.md": briefLike})
	for _, bad := range []string{"../outside.md", "/etc/passwd", ""} {
		if _, err := ParseBlock(root, BlockLocation{File: bad, Heading: "Identity"}); err == nil {
			t.Fatalf("ParseBlock(%q) = nil error, want a refusal", bad)
		}
	}
}

func TestBlockFieldLookup(t *testing.T) {
	b := Block{Title: "T", Tagline: "TL", Pitch: "P"}
	for field, want := range map[string]string{"title": "T", "tagline": "TL", "pitch": "P"} {
		got, ok := b.Field(field)
		if !ok || got != want {
			t.Errorf("Field(%q) = %q,%v want %q,true", field, got, ok, want)
		}
	}
	if _, ok := b.Field("nope"); ok {
		t.Error("Field(\"nope\") reported ok")
	}
}
