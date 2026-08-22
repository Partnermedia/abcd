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
		// A bracket that opens no link is text, as every markdown reader agrees.
		{"the array [0] holds it", "<p>the array [0] holds it</p>"},
		{"a model named opus-5[1m]", "<p>a model named opus-5[1m]</p>"},
		// An escaped pipe is table content, not a column boundary. Splitting on
		// it silently shifts every cell after it into the wrong column.
		{"| a | b |\n|---|---|\n| x \\| y | z |",
			"<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>x | y</td>\n<td>z</td>\n</tr>\n</tbody>\n</table>"},
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
		{"script href", "A [link](javascript:alert(1)) here.", 1, "link scheme"},
		// A control character inside the scheme is ignored by browsers and is the
		// shape a bypass takes; whitespace is refused one step earlier, by the
		// link parser, so this case is what the scheme check itself adds.
		{"obfuscated script href", "A [link](java\x01script:alert(1)) here.", 1, "link scheme"},
		{"upper-case script href", "A [link](JavaScript:alert(1)) here.", 1, "link scheme"},
		{"data href", "A [link](data:text/html;base64,PHNjcmlwdD4=) here.", 1, "link scheme"},
		// Flattening any of these publishes a structure the source does not
		// have, and no reader can tell. The next slice renders every record
		// body and will widen the subset guided by real refusals — silent wrong
		// output is the one outcome that must not happen in the meantime.
		{"nested blockquote", "> outer\n> > inner", 2, "nested blockquote"},
		{"list inside a blockquote", "> intro\n>\n> - one\n> - two", 3, "list inside a blockquote"},
		{"nested link", "A [link with [another](b) inside](a).", 1, "nested link"},
		// A fence that opens with no blank line before it is not a block of its
		// own: the block walk hands the whole run to the paragraph renderer,
		// which would escape the backticks into the page as visible punctuation.
		{"fence with no blank line before it", "Run this:\n```sh\necho hi\n```", 2, "fenced code block"},
		{"fence after a list item", "- one\n```sh\necho hi\n```", 2, "fenced code block"},
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
	secs, err := Sections("docs/page.md", body, n)
	if err != nil {
		t.Fatal(err)
	}
	if len(secs) != 1 || secs[0].Line != 5 {
		t.Errorf("the H1 is not reported at its source line 5: %+v", secs)
	}
	body, n = StripFrontmatter("# No frontmatter\n")
	if body != "# No frontmatter\n" || n != 0 {
		t.Errorf("a document without frontmatter was changed: %q, %d", body, n)
	}
}

func TestSectionsHonourFencesAndReportLines(t *testing.T) {
	md := "# Title\n\nIntro.\n\n## Second\n\n```sh\n# not a heading\n```\n\nAfter.\n"
	secs, err := Sections("docs/page.md", md, 0)
	if err != nil {
		t.Fatal(err)
	}
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

// TestSectionsRefuseAnUnterminatedFence pins the quietest failure this walk
// has. A fence that never closes swallows the rest of the document: no heading
// after it is a heading any more, so every following section simply is not
// there, and the page renders a short document nobody notices is short.
func TestSectionsRefuseAnUnterminatedFence(t *testing.T) {
	md := "# Title\n\nIntro.\n\n## Second\n\n```sh\necho hello\n\n## Third\n\nMore.\n"
	_, err := Sections("docs/page.md", md, 0)
	if err == nil {
		t.Fatal("an unterminated fence swallowed the rest of the document without complaint")
	}
	var ue *UnsupportedError
	if !errors.As(err, &ue) {
		t.Fatalf("error is not an UnsupportedError: %v", err)
	}
	if ue.Path != "docs/page.md" {
		t.Errorf("the refusal does not name the file: %v", err)
	}
	// The line that OPENED the fence is the one to go and look at.
	if ue.Line != 7 {
		t.Errorf("the refusal names line %d, want 7 (where the fence opens): %v", ue.Line, err)
	}
	if !strings.Contains(err.Error(), "fence") {
		t.Errorf("the refusal does not say what is wrong: %v", err)
	}

	// A closed fence is still fine, and still hides its '#' lines.
	secs, err := Sections("docs/page.md", md+"```\n", 0)
	if err != nil {
		t.Fatalf("a closed fence was refused: %v", err)
	}
	if len(secs) != 2 {
		t.Errorf("sections: %d, want 2 — the '## Third' inside the fence is not a heading", len(secs))
	}
}

// TestColumnLabelForDecodesEntities pins the portrait match against the one
// character that breaks it. A header reading "Research & design" renders as
// `Research &amp; design`; comparing that against a section title fails on the
// entity alone, and the only symptom is a missing picture that nobody reads as
// a bug. The cell must come back EXACTLY as it sits in the table, so the caller
// can substitute it.
func TestColumnLabelForDecodesEntities(t *testing.T) {
	table := "<table>\n<thead>\n<tr>\n<th></th>\n<th>Research &amp; design</th>\n<th>Delivery</th>\n</tr>\n</thead>\n</table>"
	got, ok := columnLabelFor(table, "Research & design")
	if !ok {
		t.Fatal("an entity in the header lost the portrait")
	}
	if got != "Research &amp; design" {
		t.Errorf("columnLabelFor = %q; the cell must come back as it sits in the table", got)
	}
	if _, ok := columnLabelFor(table, "Nothing like it"); ok {
		t.Error("an unrelated section title matched a column")
	}
}

func TestSiteHref(t *testing.T) {
	const forge = "https://example.invalid/o/r"
	cases := []struct{ dir, in, want string }{
		{"docs/explanation", "roles.md", "/docs/explanation/roles/"},
		{"docs/explanation", "../how-to/install.md#cli", "/docs/how-to/install/#cli"},
		{"docs/explanation", "../README.md", "/docs/"},
		{"docs/explanation", "#plugin", "#plugin"},
		{"docs/explanation", "https://example.invalid/", "https://example.invalid/"},
		// Outside docs/ there is no page on this site yet, so the link goes to
		// the forge's view of the file rather than to a relative path that 404s.
		{"docs/explanation", "../../CONTRIBUTING.md", forge + "/blob/main/CONTRIBUTING.md"},
		{"docs/explanation", "../../.abcd/development/decisions/adrs/0047-x.md#decision",
			forge + "/blob/main/.abcd/development/decisions/adrs/0047-x.md#decision"},
	}
	for _, c := range cases {
		if got := siteHref(c.dir, c.in, forge); got != c.want {
			t.Errorf("siteHref(%q, %q) = %q, want %q", c.dir, c.in, got, c.want)
		}
	}
	// With no forge URL there is nothing to point at, and the record's own text
	// is left exactly as written.
	if got := siteHref("docs/explanation", "../../CONTRIBUTING.md", ""); got != "../../CONTRIBUTING.md" {
		t.Errorf("siteHref with no forge = %q, want the href unchanged", got)
	}
}
