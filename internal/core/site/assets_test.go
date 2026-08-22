package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// assetFixture writes one file under a temporary repo root and returns the root.
func assetFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// TestAssetsRefuseAnythingOutsideTheAssetRoot closes the widest hole a
// generator like this has: an image reference is a path, and a path read from
// repository text will happily name a file that is not a picture.
//
// The build publishes what it reads. `![x](../../.git/config)` would therefore
// copy git's config — which on a CI runner carries the checkout token in an
// `http.…extraheader` line — into a directory served to the public, under a
// name nobody would look at twice. The same reference reaches `.env`, and the
// gitignored `.abcd/.work.local/` tier, none of which git would ever show as
// changed, because the output directory is not tracked.
func TestAssetsRefuseAnythingOutsideTheAssetRoot(t *testing.T) {
	root := assetFixture(t, map[string]string{
		".git/config":                     "[core]\n\textraheader = AUTHORIZATION: basic SECRET\n",
		".env":                            "TOKEN=not-a-real-token\n",
		".abcd/.work.local/scratch/n.txt": "local notes\n",
		"docs/assets/img/ok.svg":          `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"></svg>`,
		"docs/explanation/notes.md":       "prose\n",
	})
	pipe := newAssetPipe(root)
	at := Source{Path: "docs/explanation/page.md", Line: 3}

	for _, src := range []string{
		"../../.git/config",
		"../../.env",
		"../../.abcd/.work.local/scratch/n.txt",
		"../explanation/notes.md",
	} {
		if _, err := pipe.render("docs/explanation", src, "", at); err == nil {
			t.Errorf("%s was published; only committed pictures under the asset root may be", src)
		} else if !strings.Contains(err.Error(), at.Path) {
			t.Errorf("%s: the refusal does not name the page: %v", src, err)
		}
	}

	// The legitimate case still works.
	if _, err := pipe.render("docs/explanation", "../assets/img/ok.svg", "", at); err != nil {
		t.Errorf("a committed asset was refused: %v", err)
	}
}

// TestAssetsRefuseExecutableSVG is the other half of the same boundary. An SVG
// is INLINED — that is the whole point, so its var(--token) colours follow the
// reader's theme — which means its bytes become markup in the published page.
// A `<script>` element or an `on*` attribute inside one is therefore executable
// code on the site's own origin, reached with no click, from a file that is not
// code and that no reviewer reads as code.
func TestAssetsRefuseExecutableSVG(t *testing.T) {
	cases := []struct{ name, svg, says string }{
		{"script element",
			`<svg xmlns="http://www.w3.org/2000/svg"><script>fetch('https://evil.invalid/')</script></svg>`,
			"script"},
		{"event handler",
			`<svg xmlns="http://www.w3.org/2000/svg"><rect onload="alert(1)" width="1" height="1"/></svg>`,
			"onload"},
		{"mixed-case event handler",
			`<svg xmlns="http://www.w3.org/2000/svg"><rect OnLoad="alert(1)" width="1" height="1"/></svg>`,
			"onload"},
		{"foreign object",
			`<svg xmlns="http://www.w3.org/2000/svg"><foreignObject><body xmlns="http://www.w3.org/1999/xhtml"/></foreignObject></svg>`,
			"foreignObject"},
		{"executable href",
			`<svg xmlns="http://www.w3.org/2000/svg"><a href="javascript:alert(1)"><rect width="1" height="1"/></a></svg>`,
			""},
		{"external use",
			`<svg xmlns="http://www.w3.org/2000/svg"><use href="https://evil.invalid/x.svg#a"/></svg>`,
			"href"},
		{"style element",
			`<svg xmlns="http://www.w3.org/2000/svg"><style>@import url(https://evil.invalid/x.css);</style></svg>`,
			"style"},
		{"external image panel",
			`<svg xmlns="http://www.w3.org/2000/svg"><image href="https://evil.invalid/x.png"/></svg>`,
			"image"},
		// The <image> carve-out exists for the committed drawings' raster
		// panels; nesting a document inside one would reopen the surface.
		{"svg inside an image panel",
			`<svg xmlns="http://www.w3.org/2000/svg"><image href="data:image/svg+xml;base64,PHN2Zz48c2NyaXB0Lz48L3N2Zz4="/></svg>`,
			"image"},
		{"doctype entity",
			`<!DOCTYPE svg [<!ENTITY x SYSTEM "file:///etc/passwd">]><svg xmlns="http://www.w3.org/2000/svg"><desc>&x;</desc></svg>`,
			""},
	}
	at := Source{Path: "docs/explanation/page.md", Line: 5}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := assetFixture(t, map[string]string{"docs/assets/img/x.svg": c.svg})
			pipe := newAssetPipe(root)
			out, err := pipe.render("docs/explanation", "../assets/img/x.svg", "", at)
			if err == nil {
				t.Fatalf("inlined an executable SVG into the page:\n%s", out)
			}
			if !strings.Contains(err.Error(), "docs/assets/img/x.svg") {
				t.Errorf("the refusal does not name the asset: %v", err)
			}
			if c.says != "" && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(c.says)) {
				t.Errorf("the refusal does not say %q: %v", c.says, err)
			}
		})
	}
}

// TestAssetsAcceptDrawings keeps the refusal from swallowing the real corpus:
// the shapes, gradients, clip paths and text the committed drawings are made of
// stay renderable, and a same-document reference is not an external one.
func TestAssetsAcceptDrawings(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 10 10" width="10" height="10">
<title>A drawing</title><desc>Described</desc>
<defs><clipPath id="c"><rect width="10" height="10"/></clipPath>
<marker id="m" markerWidth="4" markerHeight="4" refX="2" refY="2" orient="auto"><path d="M0 0 L4 2 L0 4 z"/></marker>
<linearGradient id="g"><stop offset="0" stop-color="var(--accent, #06c)"/></linearGradient></defs>
<g clip-path="url(#c)"><circle cx="5" cy="5" r="4" fill="var(--ink, #000)"/>
<path d="M0 0 L10 10" stroke="url(#g)" marker-end="url(#m)"/>
<text x="1" y="9" font-family="sans-serif" font-size="2">hi<tspan dy="1">there</tspan></text>
<use xlink:href="#c"/></g></svg>`
	root := assetFixture(t, map[string]string{"docs/assets/img/x.svg": svg})
	pipe := newAssetPipe(root)
	out, err := pipe.render("docs/explanation", "../assets/img/x.svg", "", Source{Path: "p.md", Line: 1})
	if err != nil {
		t.Fatalf("a drawing was refused: %v", err)
	}
	if !strings.Contains(out, "var(--ink, #000)") {
		t.Error("the token colours did not survive inlining")
	}
	if strings.Contains(out, `width="10"`) {
		t.Error("the stylesheet must size the drawing; its own width survived")
	}
}

// TestCommittedSVGAssetsAreDrawings runs the same refusal over every SVG this
// repository actually carries, so an asset that could never be published is
// caught when it lands rather than when a page first references it.
func TestCommittedSVGAssetsAreDrawings(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..")
	dir := filepath.Join(repoRoot, "docs", "assets", "img")
	var found int
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".svg") {
			return err
		}
		found++
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := checkInlinableSVG(string(data), path); err != nil {
			t.Errorf("%v", err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if found == 0 {
		t.Fatal("no SVG assets found; the scan proved nothing")
	}
}
