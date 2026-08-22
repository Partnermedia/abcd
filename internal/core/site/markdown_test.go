package site

import (
	"errors"
	"strings"
	"testing"
)

// testRenderer is the renderer with the two site hooks stubbed, so these tests
// judge the markdown subset and nothing else.
func testRenderer() *Renderer {
	return &Renderer{
		UI: UI{Copy: "copy", Copied: "copied"},
		Image: func(src, alt string, _ Source) (string, error) {
			return `<img src="` + src + `" alt="` + alt + `">`, nil
		},
		Link: func(href string, _ Source) string { return href },
	}
}

func render(t *testing.T, md string) string {
	t.Helper()
	r := testRenderer()
	out, err := r.RenderBlocks("docs/page.md", Blocks(md, 1))
	if err != nil {
		t.Fatalf("render %q: %v", md, err)
	}
	return out
}

func TestRenderSubset(t *testing.T) {
	cases := []struct{ md, want string }{
		{"# A heading", "<h1>A heading</h1>"},
		{"### A deeper one", "<h3>A deeper one</h3>"},
		{"Plain prose.", "<p>Plain prose.</p>"},
		{"**strong** and *em* and `code`", "<p><strong>strong</strong> and <em>em</em> and <code>code</code></p>"},
		{"__strong__ and _em_", "<p><strong>strong</strong> and <em>em</em></p>"},
		{"snake_case_name stays whole", "<p>snake_case_name stays whole</p>"},
		{"[text](https://example.invalid/)", `<p><a href="https://example.invalid/">text</a></p>`},
		{"![alt](x.png)", `<p><img src="x.png" alt="alt"></p>`},
		{"Ampersands & angle brackets 3 < 4", "<p>Ampersands &amp; angle brackets 3 &lt; 4</p>"},
		{"> quoted", "<blockquote>\n<p>quoted</p>\n</blockquote>"},
		{"- one\n- two", "<ul>\n<li>one</li>\n<li>two</li>\n</ul>"},
		{"1. one\n2. two", "<ol>\n<li>one</li>\n<li>two</li>\n</ol>"},
		{"| a | b |\n|---|---|\n| 1 | 2 |",
			"<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>1</td>\n<td>2</td>\n</tr>\n</tbody>\n</table>"},
		{"Escaped \\*not emphasis\\*", "<p>Escaped *not emphasis*</p>"},
	}
	for _, c := range cases {
		if got := render(t, c.md); got != c.want {
			t.Errorf("render(%q)\n got: %s\nwant: %s", c.md, got, c.want)
		}
	}
}

func TestRenderFenceIsACopyableCommand(t *testing.T) {
	got := render(t, "```sh\necho hello\n```")
	for _, want := range []string{
		`<div class="cmd">`,
		`<code class="language-sh">echo hello`,
		`data-copy="echo hello"`,
		`data-copied="copied"`,
		`>copy</button>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("fence render missing %q:\n%s", want, got)
		}
	}
}

// TestRenderRefusesOutOfSubset is the loud-failure rule: a construct the site
// does not render stops the build and names file and line, rather than reaching
// a reader as raw markdown or as a silent hole.
func TestRenderRefusesOutOfSubset(t *testing.T) {
	cases := []struct {
		name string
		md   string
		line int
		says string
	}{
		{"raw HTML block", "Fine.\n\n<div class=\"x\">hidden</div>", 3, "raw HTML block"},
		{"inline HTML", "Some <b>bold</b> prose.", 1, "inline HTML"},
		{"autolink", "Read <https://example.invalid/> for more.", 1, "inline HTML or autolink"},
		{"reference link", "A [reference][1] link.", 1, "link"},
		{"link title", `A [titled](https://example.invalid/ "hi") link.`, 1, "link"},
		{"setext heading", "A heading\n=========", 2, "setext heading"},
		{"thematic break", "Before.\n\n---\n\nAfter.", 3, "thematic break"},
		{"indented code", "Prose.\n\n    indented code", 3, "indented code block"},
		{"nested list", "- one\n  - nested", 2, "nested list"},
		{"unclosed emphasis", "An *unclosed emphasis.", 1, "unclosed emphasis"},
		{"unclosed code span", "An `unclosed code span.", 1, "unclosed code span"},
		{"table without a delimiter", "| a | b |\n| 1 | 2 |", 1, "pipe table"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := testRenderer()
			_, err := r.RenderBlocks("docs/page.md", Blocks(c.md, 1))
			if err == nil {
				t.Fatalf("rendered %q without complaint", c.md)
			}
			var ue *UnsupportedError
			if !errors.As(err, &ue) {
				t.Fatalf("error is not an UnsupportedError: %v", err)
			}
			if ue.Path != "docs/page.md" {
				t.Errorf("error names %q, not the file", ue.Path)
			}
			if ue.Line != c.line {
				t.Errorf("error names line %d, want %d: %v", ue.Line, c.line, err)
			}
			if !strings.Contains(err.Error(), c.says) {
				t.Errorf("error does not say %q: %v", c.says, err)
			}
		})
	}
}

// TestRenderRefusesRemoteImage keeps adr-47's every-picture-is-committed rule
// enforceable at render time rather than at review time.
func TestRenderRefusesRemoteImage(t *testing.T) {
	pipe := newAssetPipe(t.TempDir())
	r := testRenderer()
	r.Image = func(src, alt string, at Source) (string, error) { return pipe.render("docs", src, alt, at) }
	_, err := r.RenderBlocks("docs/page.md", Blocks("![alt](https://example.invalid/x.png)", 1))
	if err == nil || !strings.Contains(err.Error(), "remote image") {
		t.Fatalf("a remote image must be refused, got %v", err)
	}
}

// --- the structure walk ---------------------------------------------------

func TestSlug(t *testing.T) {
	cases := map[string]string{
		"Identity (canonical)":        "identity-canonical",
		"Capturing issues & thoughts": "capturing-issues-thoughts",
		"`abcd site build`":           "abcd-site-build",
		"macOS":                       "macos",
		"**Bold** heading":            "bold-heading",
	}
	for in, want := range cases {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStripFrontmatter(t *testing.T) {
	// The cut lands on the closing delimiter's own newline, so the body opens
	// with the tail of line 3 and only two lines are fully consumed. Reporting
	// it that way is what keeps a later error message's line number the one a
	// reader finds in the file.
	src := "---\nid: adr-1\n---\n\n# Title\n"
	body, n := StripFrontmatter(src)
	if body != "\n\n# Title\n" {
		t.Errorf("body = %q", body)
	}
	if n != 2 {
		t.Errorf("consumed %d lines, want 2", n)
	}
	if secs := Sections(body, n); len(secs) != 1 || secs[0].Line != 5 {
		t.Errorf("the H1 is not reported at its source line 5: %+v", secs)
	}
	body, n = StripFrontmatter("# No frontmatter\n")
	if body != "# No frontmatter\n" || n != 0 {
		t.Errorf("a document without frontmatter was changed: %q, %d", body, n)
	}
}

func TestSectionsHonourFencesAndReportLines(t *testing.T) {
	md := "# Title\n\nIntro.\n\n## Second\n\n```sh\n# not a heading\n```\n\nAfter.\n"
	secs := Sections(md, 0)
	if len(secs) != 2 {
		t.Fatalf("sections: %d — a '#' inside a fence was read as a heading: %+v", len(secs), secs)
	}
	if secs[0].Title != "Title" || secs[0].Line != 1 || secs[0].BodyLine != 3 {
		t.Errorf("first section: %+v", secs[0])
	}
	if secs[1].Title != "Second" || secs[1].Line != 5 {
		t.Errorf("second section: %+v", secs[1])
	}
	blocks := Blocks(secs[1].Body, secs[1].BodyLine)
	if len(blocks) != 2 {
		t.Fatalf("blocks: %+v", blocks)
	}
	if blocks[0].Line != 7 || !strings.HasPrefix(blocks[0].Text, "```sh") {
		t.Errorf("fence block: %+v", blocks[0])
	}
	if blocks[1].Line != 11 || blocks[1].Text != "After." {
		t.Errorf("trailing block: %+v", blocks[1])
	}
}

func TestSiteHref(t *testing.T) {
	cases := []struct{ dir, in, want string }{
		{"docs/explanation", "roles.md", "/docs/explanation/roles/"},
		{"docs/explanation", "../how-to/install.md#cli", "/docs/how-to/install/#cli"},
		{"docs/explanation", "../README.md", "/docs/"},
		{"docs/explanation", "#plugin", "#plugin"},
		{"docs/explanation", "https://example.invalid/", "https://example.invalid/"},
		{"docs/explanation", "../../CONTRIBUTING.md", "../../CONTRIBUTING.md"},
	}
	for _, c := range cases {
		if got := siteHref(c.dir, c.in); got != c.want {
			t.Errorf("siteHref(%q, %q) = %q, want %q", c.dir, c.in, got, c.want)
		}
	}
}
