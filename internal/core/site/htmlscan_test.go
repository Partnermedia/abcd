package site

import (
	"strings"
	"testing"
)

// TestHTMLScannerReadsTheGeneratorsSubset asserts the shapes the composer
// actually emits parse, so a refusal below means the document is wrong rather
// than the reader being narrow.
func TestHTMLScannerReadsTheGeneratorsSubset(t *testing.T) {
	src := "<!doctype html>\n<html lang=\"en\">\n<head>\n" +
		`<meta charset="utf-8">` + "\n" +
		`<link rel="stylesheet" href="https://fonts.example/css2?family=A&family=B">` + "\n" +
		"<title>Probe &amp; co</title>\n" +
		`<script src="site.js" defer></script>` + "\n" +
		"</head>\n<body>\n" +
		`<p data-src="docs/x.md#h">a &lt;tag&gt; and an &#39;apostrophe&#39;</p>` +
		`<span class="svgasset loop"><svg viewBox="0 0 4 4"><path d="M0 0"/><text><tspan>a</tspan></text></svg></span>` +
		"</body>\n</html>\n"

	doc, err := parseHTML("index.html", src)
	if err != nil {
		t.Fatalf("parseHTML refused the generator's own output: %v", err)
	}
	p := findElement(doc, func(n *htmlNode) bool { return n.Name == "p" })
	if p == nil {
		t.Fatal("no <p> parsed")
	}
	if got := p.TextContent(); got != "a <tag> and an 'apostrophe'" {
		t.Errorf("text = %q, want the entities decoded", got)
	}
	if p.Attr("data-src") != "docs/x.md#h" {
		t.Errorf("data-src = %q", p.Attr("data-src"))
	}
	// A bare `&` inside a URL is punctuation in an attribute nobody reads for
	// provenance, and must not fail a page.
	link := findElement(doc, func(n *htmlNode) bool { return n.Name == "link" })
	if !strings.Contains(link.Attr("href"), "&family=B") {
		t.Errorf("href = %q, want the bare ampersand kept", link.Attr("href"))
	}
	if findElement(doc, func(n *htmlNode) bool { return n.Name == "path" }) == nil {
		t.Error("the self-closing SVG element did not parse")
	}
}

// TestHTMLScannerRefusesWhatItCannotVouchFor is the reader's whole posture: a
// provenance walk that skips markup it does not understand passes by not
// looking, so every shape outside the emitted subset stops the read by name.
func TestHTMLScannerRefusesWhatItCannotVouchFor(t *testing.T) {
	cases := []struct{ name, src, want string }{
		{"comment", "<!doctype html>\n<html><body><!-- hi --></body></html>", "comment"},
		{"unquoted attribute", `<!doctype html>` + "\n" + `<html><body><p class=lede>x</p></body></html>`, "double-quoted"},
		{"single-quoted attribute", `<!doctype html>` + "\n" + `<html><body><p class='lede'>x</p></body></html>`, "double-quoted"},
		{"mismatched end tag", "<!doctype html>\n<html><body><p>x</span></body></html>", "closes <p>"},
		{"unclosed element", "<!doctype html>\n<html><body><div>x", "never closed"},
		{"crossed elements", "<!doctype html>\n<html><body><div>x</body></html>", "closes <div>"},
		{"stray end tag", "<!doctype html>\n</p>", "not open"},
		{"unknown entity", "<!doctype html>\n<html><body><p>a &nbsp; b</p></body></html>", "&nbsp;"},
		{"processing instruction", "<!doctype html>\n<html><body><?php ?></body></html>", "processing instruction"},
		{"foreign doctype", "<!DOCTYPE svg>\n<html></html>", "not <!doctype html>"},
		{"repeated attribute", "<!doctype html>\n<html><body><p id=\"a\" id=\"b\">x</p></body></html>", "repeats"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseHTML("page.html", tc.src)
			if err == nil {
				t.Fatalf("parseHTML accepted %s; want a refusal naming %q", tc.name, tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want it to name %q", err, tc.want)
			}
		})
	}
}

// TestHTMLScannerFindsARawTextCloseTagInTheOriginalBytes is the bug a
// lower-cased haystack causes: case folding is not length-preserving, so an
// index taken from a folded copy lands in the wrong place in the original — and
// the parser then resumes part-way through the document, skipping whatever
// follows the raw-text element without reporting anything at all.
func TestHTMLScannerFindsARawTextCloseTagInTheOriginalBytes(t *testing.T) {
	cases := []struct{ name, body string }{
		// A rune that SHRINKS when lower-cased (U+212A KELVIN SIGN, 3 bytes to 1).
		{"shrinking rune", "var s = \"K\";"},
		// A rune that shrinks by two (U+0130, 2 bytes to 1 plus a combining mark
		// in Unicode folding).
		{"dotted capital I", "var s = \"İ\";"},
		// Bytes that are not valid UTF-8 GROW into replacement characters.
		{"invalid utf-8", strings.Repeat("\xe0", 17)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := "<!doctype html>\n<html lang=\"en\">\n<body>\n<script>" + tc.body +
				"</script><p>Trusted by dozens of teams.</p>\n</body>\n</html>\n"
			doc, err := parseHTML("index.html", src)
			if err != nil {
				t.Fatalf("parseHTML: %v", err)
			}
			uses, _, _, _, _ := visibleText("index.html", doc)
			found := false
			for _, u := range uses {
				if strings.Contains(u.Text, "Trusted by dozens of teams.") {
					found = true
				}
				if strings.Contains(u.Text, "script") || strings.Contains(u.Text, "t>") {
					t.Errorf("the walk read script markup as visible text: %q", u.Text)
				}
			}
			if !found {
				t.Fatalf("the paragraph after the script was never seen by the walk; visible text = %+v", uses)
			}
		})
	}
}

// TestIndexFoldASCIIKeepsItsIndexes is the helper's whole contract.
func TestIndexFoldASCIIKeepsItsIndexes(t *testing.T) {
	hay := "K" + "</SCRIPT>"
	i := indexFoldASCII(hay, "</script>")
	if i != len("K") {
		t.Fatalf("indexFoldASCII = %d, want %d (an index into the ORIGINAL bytes)", i, len("K"))
	}
	if hay[i:] != "</SCRIPT>" {
		t.Errorf("hay[%d:] = %q, want the close tag", i, hay[i:])
	}
	// ASCII folding only: the Kelvin sign is not a 'k' as far as a tag name goes.
	if got := indexFoldASCII("K", "k"); got != -1 {
		t.Errorf("indexFoldASCII folded a non-ASCII rune to ASCII (%d)", got)
	}
}

// TestStylesheetReaderCreditsOnlyWhatItCanResolve asserts the CSS reader
// declines what it cannot attribute, so an unreadable rule makes a check
// stricter rather than weaker.
func TestStylesheetReaderCreditsOnlyWhatItCanResolve(t *testing.T) {
	s := parseStylesheet(`
/* a comment {overflow-x:auto} */
.wrap{max-width:1120px;margin:0 auto}
img{max-width:100%;height:auto}
pre{overflow-x:auto}
.tablewrap{overflow-x:auto}
.tabs{overflow:hidden}
.deep .nested{overflow-x:auto}
@media (max-width:700px){.narrow{max-width:420px}}
`)
	if !s.OverflowElements["pre"] {
		t.Error("pre was not credited with its overflow rule")
	}
	if !s.OverflowClasses["tablewrap"] {
		t.Error(".tablewrap was not credited with its overflow rule")
	}
	if s.OverflowClasses["tabs"] {
		t.Error("overflow:hidden was credited as scrolling; it clips instead")
	}
	if s.OverflowClasses["nested"] {
		t.Error("a descendant selector was credited; this reader cannot tell whether an element is inside its ancestor")
	}
	if !s.ImageMaxWidth {
		t.Error("the img max-width rule was not seen")
	}
	if s.ContentColumnPx != 1120 {
		t.Errorf("ContentColumnPx = %d, want 1120 (the media-query prelude is not a declaration)", s.ContentColumnPx)
	}
}
