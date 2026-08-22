package site

// The Markdown subset renderer.
//
// The site renders repository prose, and the repository writes a small, stable
// dialect: ATX headings, paragraphs, bold/italic/code spans, links, images,
// fenced code, pipe tables, blockquotes and lists. That dialect is what this
// renders, and everything else is a BUILD ERROR naming file and line.
//
// Refusing is the point. A renderer that passes an unknown construct through
// unrendered publishes raw markdown to readers; a renderer that silently drops
// it publishes a hole. Both are worse than a build that stops and says
// "docs/explanation/roles.md:31 — reference link", because the fix for that is
// one edit and the fix for a silent hole is noticing it exists.

import (
	"fmt"
	"regexp"
	"strings"
)

// UnsupportedError is a markdown construct outside the rendered subset.
type UnsupportedError struct {
	Path      string
	Line      int
	Construct string
	Detail    string
}

func (e *UnsupportedError) Error() string {
	d := ""
	if e.Detail != "" {
		d = ": " + e.Detail
	}
	return fmt.Sprintf("%s:%d: unsupported markdown construct: %s%s — the site renders a fixed subset and never passes text through unrendered",
		e.Path, e.Line, e.Construct, d)
}

var (
	tableDelimRe   = regexp.MustCompile(`^\s*\|?\s*:?-{1,}:?\s*(\|\s*:?-{1,}:?\s*)*\|?\s*$`)
	orderedItemRe  = regexp.MustCompile(`^(\d+)[.)]\s+(.*)$`)
	setextRe       = regexp.MustCompile(`^\s{0,3}(=+|-{2,})\s*$`)
	thematicRe     = regexp.MustCompile(`^\s{0,3}((\*\s*){3,}|(-\s*){3,}|(_\s*){3,})$`)
	rawHTMLStartRe = regexp.MustCompile(`^\s*<[A-Za-z!/]`)
)

// Renderer turns the markdown subset into the site's HTML. Its two hooks are
// the places the site differs from a generic renderer: an image becomes a
// committed asset (inlined SVG or copied raster) rather than a bare <img>, and
// a repo-relative link becomes a site route.
type Renderer struct {
	// UI is the closed allowlist of interface strings; the renderer adds the
	// copy-button label and nothing else.
	UI UI
	// Image renders one image reference. src is as written in the markdown,
	// relative to the page; alt is its alt text.
	Image func(src, alt string, at Source) (string, error)
	// Link rewrites one href. It never fails: an href it does not recognise is
	// left exactly as the record wrote it.
	Link func(href string, at Source) string
}

// Source is a position in a repository file.
type Source struct {
	Path string
	Line int
}

// RenderBlocks renders a block sequence in order.
func (r *Renderer) RenderBlocks(path string, blocks []Block) (string, error) {
	var b strings.Builder
	for _, blk := range blocks {
		h, err := r.RenderBlock(path, blk)
		if err != nil {
			return "", err
		}
		b.WriteString(h)
	}
	return b.String(), nil
}

// RenderBlock renders one top-level block.
func (r *Renderer) RenderBlock(path string, blk Block) (string, error) {
	at := Source{Path: path, Line: blk.Line}
	lines := strings.Split(blk.Text, "\n")
	first := lines[0]

	// A fence must open its own block. Without a blank line before it the block
	// walk never sees it start, so the whole run — prose, backticks and code —
	// arrives here as one paragraph, and every backtick would be escaped into the
	// page as visible punctuation with the code inlined into the sentence.
	if !strings.HasPrefix(first, "```") {
		for i, ln := range lines {
			if strings.HasPrefix(ln, "```") {
				return "", &UnsupportedError{at.Path, at.Line + i, "fenced code block without a blank line before it",
					"a fence opens its own block; without the blank line the code renders as part of the paragraph above"}
			}
		}
	}

	switch {
	case strings.HasPrefix(first, "```"):
		return r.fence(at, lines)
	case headingRe.MatchString(first):
		return r.heading(at, first)
	case strings.HasPrefix(strings.TrimSpace(first), ">"):
		return r.blockquote(at, lines)
	case strings.HasPrefix(strings.TrimSpace(first), "|"):
		return r.table(at, lines)
	case isUnorderedItem(first) || orderedItemRe.MatchString(first):
		return r.list(at, lines)
	}
	return r.paragraph(at, lines)
}

// fence renders a fenced code block as a copyable command block: the button's
// label is a ui.json string, and its payload is the block's own raw text.
func (r *Renderer) fence(at Source, lines []string) (string, error) {
	info := strings.TrimSpace(strings.TrimPrefix(lines[0], "```"))
	if strings.ContainsAny(info, " \t") {
		return "", &UnsupportedError{at.Path, at.Line, "fenced-code info string", "only a bare language word is rendered, got " + quote(info)}
	}
	body := lines[1:]
	if n := len(body); n > 0 && strings.HasPrefix(strings.TrimSpace(body[n-1]), "```") {
		body = body[:n-1]
	}
	raw := strings.Join(body, "\n")
	cls := ""
	if info != "" {
		cls = ` class="language-` + escapeAttr(info) + `"`
	}
	// The button's two labels both travel in the markup: the page's script adds
	// no words of its own, so every string a reader sees comes from ui.json.
	return `<div class="cmd"><pre><code` + cls + `>` + escapeText(raw) + "\n" +
		`</code></pre><button class="copy" data-copy="` + escapeAttr(raw) +
		`" data-copied="` + escapeAttr(r.UI.Copied) + `">` + escapeText(r.UI.Copy) + `</button></div>`, nil
}

// heading renders an ATX heading.
func (r *Renderer) heading(at Source, line string) (string, error) {
	m := headingRe.FindStringSubmatch(line)
	level := len(m[1])
	inner, err := r.inline(at, strings.TrimSpace(m[2]))
	if err != nil {
		return "", err
	}
	tag := fmt.Sprintf("h%d", level)
	return "<" + tag + ">" + inner + "</" + tag + ">", nil
}

// blockquote renders a `>` block, splitting it into paragraphs on blank lines
// exactly as the source does.
func (r *Renderer) blockquote(at Source, lines []string) (string, error) {
	var stripped []string
	for i, ln := range lines {
		t := strings.TrimSpace(ln)
		if !strings.HasPrefix(t, ">") {
			return "", &UnsupportedError{at.Path, at.Line + i, "lazy blockquote continuation", "every line of a quoted block must start with '>'"}
		}
		inner := strings.TrimPrefix(strings.TrimPrefix(t, ">"), " ")
		// A quoted block carries paragraphs. Anything with structure of its own
		// inside one would be FLATTENED into those paragraphs — the nesting and
		// the bullets would simply cease to exist on the page, with nothing to
		// tell a reader they ever did. That is the one outcome worse than a
		// build error.
		if strings.HasPrefix(strings.TrimSpace(inner), ">") {
			return "", &UnsupportedError{at.Path, at.Line + i, "nested blockquote",
				"a quoted block carries paragraphs; a quote inside a quote would flatten into them"}
		}
		if isUnorderedItem(inner) || orderedItemRe.MatchString(inner) {
			return "", &UnsupportedError{at.Path, at.Line + i, "list inside a blockquote",
				"a quoted block carries paragraphs; the bullets would flatten into one"}
		}
		stripped = append(stripped, inner)
	}
	var b strings.Builder
	b.WriteString("<blockquote>\n")
	para := []string{}
	paraLine := at.Line
	flush := func() error {
		if len(para) == 0 {
			return nil
		}
		inner, err := r.inline(Source{at.Path, paraLine}, strings.Join(para, "\n"))
		if err != nil {
			return err
		}
		b.WriteString("<p>" + inner + "</p>\n")
		para = nil
		return nil
	}
	for i, ln := range stripped {
		if strings.TrimSpace(ln) == "" {
			if err := flush(); err != nil {
				return "", err
			}
			continue
		}
		if len(para) == 0 {
			paraLine = at.Line + i
		}
		para = append(para, ln)
	}
	if err := flush(); err != nil {
		return "", err
	}
	b.WriteString("</blockquote>")
	return b.String(), nil
}

// tableCells splits one table row into its cells, honouring `\|` as content.
// An escaped pipe is a pipe a cell contains, not a column boundary, and
// splitting on it shifts every cell after it one column to the left — a table
// that is quietly wrong, which is worse than one that fails to build. The
// escape itself is left in place for the inline pass, which already knows `\|`
// means `|`.
func tableCells(ln string) []string {
	t := strings.TrimSpace(ln)
	t = strings.TrimPrefix(t, "|")
	t = strings.TrimSuffix(t, "|")
	var out []string
	var cur strings.Builder
	for i := 0; i < len(t); i++ {
		switch {
		case t[i] == '\\' && i+1 < len(t):
			cur.WriteByte(t[i])
			cur.WriteByte(t[i+1])
			i++
		case t[i] == '|':
			out = append(out, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteByte(t[i])
		}
	}
	return append(out, strings.TrimSpace(cur.String()))
}

// table renders a pipe table. Column alignment is not rendered — the site's
// stylesheet aligns every column the same way — so an alignment row is accepted
// and its colons ignored.
func (r *Renderer) table(at Source, lines []string) (string, error) {
	if len(lines) < 2 || !tableDelimRe.MatchString(lines[1]) {
		return "", &UnsupportedError{at.Path, at.Line, "pipe table", "a table needs a header row and a |---|---| delimiter row"}
	}
	var b strings.Builder
	b.WriteString("<table>\n<thead>\n<tr>\n")
	for _, c := range tableCells(lines[0]) {
		inner, err := r.inline(Source{at.Path, at.Line}, c)
		if err != nil {
			return "", err
		}
		b.WriteString("<th>" + inner + "</th>\n")
	}
	b.WriteString("</tr>\n</thead>\n<tbody>\n")
	for i, ln := range lines[2:] {
		b.WriteString("<tr>\n")
		for _, c := range tableCells(ln) {
			inner, err := r.inline(Source{at.Path, at.Line + 2 + i}, c)
			if err != nil {
				return "", err
			}
			b.WriteString("<td>" + inner + "</td>\n")
		}
		b.WriteString("</tr>\n")
	}
	b.WriteString("</tbody>\n</table>")
	return b.String(), nil
}

// list renders a flat unordered or ordered list. A nested list is refused
// rather than flattened: the record does not write one, and flattening it would
// publish a different structure from the one on GitHub.
func (r *Renderer) list(at Source, lines []string) (string, error) {
	ordered := orderedItemRe.MatchString(lines[0])
	var items []string
	var itemLines []int
	for i, ln := range lines {
		switch {
		case strings.HasPrefix(ln, " ") || strings.HasPrefix(ln, "\t"):
			if len(items) == 0 {
				return "", &UnsupportedError{at.Path, at.Line + i, "indented list content", "the rendered subset carries flat lists only"}
			}
			t := strings.TrimSpace(ln)
			if isUnorderedItem(t) || orderedItemRe.MatchString(t) {
				return "", &UnsupportedError{at.Path, at.Line + i, "nested list", "the rendered subset carries flat lists only"}
			}
			items[len(items)-1] += "\n" + t
		case isUnorderedItem(ln):
			if ordered {
				return "", &UnsupportedError{at.Path, at.Line + i, "mixed list", "a list is ordered or unordered, not both"}
			}
			items = append(items, strings.TrimSpace(ln[2:]))
			itemLines = append(itemLines, at.Line+i)
		case orderedItemRe.MatchString(ln):
			if !ordered {
				return "", &UnsupportedError{at.Path, at.Line + i, "mixed list", "a list is ordered or unordered, not both"}
			}
			m := orderedItemRe.FindStringSubmatch(ln)
			items = append(items, strings.TrimSpace(m[2]))
			itemLines = append(itemLines, at.Line+i)
		default:
			if len(items) == 0 {
				return "", &UnsupportedError{at.Path, at.Line + i, "list", "a list item starts with '- ' or '1. '"}
			}
			items[len(items)-1] += "\n" + strings.TrimSpace(ln)
		}
	}
	tag := "ul"
	if ordered {
		tag = "ol"
	}
	var b strings.Builder
	b.WriteString("<" + tag + ">\n")
	for i, it := range items {
		inner, err := r.inline(Source{at.Path, itemLines[i]}, it)
		if err != nil {
			return "", err
		}
		b.WriteString("<li>" + inner + "</li>\n")
	}
	b.WriteString("</" + tag + ">")
	return b.String(), nil
}

// paragraph renders a run of prose, refusing the block constructs that are not
// in the subset before it starts — a setext heading, a thematic break, an
// indented code block, or a raw HTML block, each of which would otherwise be
// escaped into the page as visible punctuation.
func (r *Renderer) paragraph(at Source, lines []string) (string, error) {
	for i, ln := range lines {
		switch {
		case i > 0 && setextRe.MatchString(ln):
			return "", &UnsupportedError{at.Path, at.Line + i, "setext heading or thematic break", "write headings as '# ' and drop rules"}
		case thematicRe.MatchString(ln):
			return "", &UnsupportedError{at.Path, at.Line + i, "thematic break", "the rendered subset has no horizontal rule"}
		case rawHTMLStartRe.MatchString(ln):
			return "", &UnsupportedError{at.Path, at.Line + i, "raw HTML block", "every picture is a committed asset and every element is generated"}
		case i == 0 && (strings.HasPrefix(ln, "    ") || strings.HasPrefix(ln, "\t")):
			return "", &UnsupportedError{at.Path, at.Line, "indented code block", "fence code with ``` so it renders as a copyable command"}
		}
	}
	inner, err := r.inline(at, strings.Join(lines, "\n"))
	if err != nil {
		return "", err
	}
	// An image on its own line is a block-level figure, exactly as the source
	// markdown reads it; the layouts detect one by this shape.
	return "<p>" + inner + "</p>", nil
}

// isUnorderedItem reports whether a line opens an unordered list item.
func isUnorderedItem(ln string) bool {
	return len(ln) > 2 && (ln[0] == '-' || ln[0] == '*' || ln[0] == '+') && ln[1] == ' '
}

// inline renders the span-level subset: code spans, images, links, strong,
// emphasis, and backslash escapes. Everything else is text, and anything that
// looks like a construct the subset does not carry is refused.
func (r *Renderer) inline(at Source, s string) (string, error) {
	var b strings.Builder
	for i := 0; i < len(s); {
		switch c := s[i]; {
		case c == '\\' && i+1 < len(s) && isEscapable(s[i+1]):
			b.WriteString(escapeText(string(s[i+1])))
			i += 2
		case c == '`':
			n := runLen(s, i, '`')
			end := strings.Index(s[i+n:], s[i:i+n])
			if end < 0 {
				return "", &UnsupportedError{at.Path, at.Line, "unclosed code span", quote(clip(s[i:]))}
			}
			code := s[i+n : i+n+end]
			if n > 1 {
				code = strings.Trim(code, " ")
			}
			b.WriteString("<code>" + escapeText(code) + "</code>")
			i += n + end + n
		case c == '!' && i+1 < len(s) && s[i+1] == '[':
			text, href, next, kind := parseLink(s, i+1)
			switch kind {
			case linkLiteral:
				b.WriteString("!")
				i++
				continue
			case linkRefused:
				return "", &UnsupportedError{at.Path, at.Line, "image", "only ![alt](src) is rendered, got " + quote(clip(s[i:]))}
			}
			html, err := r.Image(href, text, at)
			if err != nil {
				return "", err
			}
			b.WriteString(html)
			i = next
		case c == '[':
			text, href, next, kind := parseLink(s, i)
			switch kind {
			case linkLiteral:
				// A bracket that opens no link is a bracket. Refusing it would
				// blocker-fail the build over prose like "the array [0]", which
				// every markdown reader renders literally.
				b.WriteString("[")
				i++
				continue
			case linkRefused:
				return "", &UnsupportedError{at.Path, at.Line, "link",
					"only [text](href) is rendered; reference links, footnotes and link titles are not in the subset"}
			}
			if scheme, bad := executableScheme(href); bad {
				return "", &UnsupportedError{at.Path, at.Line, "link scheme",
					quote(scheme) + " runs code in the reader's browser; the site links to pages and files"}
			}
			inner, err := r.inline(at, text)
			if err != nil {
				return "", err
			}
			// HTML has no nested anchor: a browser closes the outer one at the
			// inner, so the markup a reader gets is not the markup written here
			// and the outer link's tail stops being clickable. Markdown itself
			// forbids the construct; this says so instead of emitting it.
			if strings.Contains(inner, "<a ") {
				return "", &UnsupportedError{at.Path, at.Line, "nested link",
					"HTML has no anchor inside an anchor; the browser would silently close the outer one"}
			}
			b.WriteString(`<a href="` + escapeAttr(r.Link(href, at)) + `">` + inner + "</a>")
			i = next
		case c == '*' || c == '_':
			n := runLen(s, i, c)
			if n > 2 {
				return "", &UnsupportedError{at.Path, at.Line, "emphasis run", "only ** for strong and * for emphasis are rendered"}
			}
			if c == '_' && !underscoreDelimits(s, i, n) {
				b.WriteString("_")
				i++
				continue
			}
			delim := s[i : i+n]
			end := strings.Index(s[i+n:], delim)
			if end < 0 {
				if c == '_' {
					b.WriteString(strings.Repeat("_", n))
					i += n
					continue
				}
				return "", &UnsupportedError{at.Path, at.Line, "unclosed emphasis", quote(clip(s[i:]))}
			}
			inner, err := r.inline(at, s[i+n:i+n+end])
			if err != nil {
				return "", err
			}
			tag := "em"
			if n == 2 {
				tag = "strong"
			}
			b.WriteString("<" + tag + ">" + inner + "</" + tag + ">")
			i += n + end + n
		case c == '<':
			if i+1 < len(s) && (isAlpha(s[i+1]) || s[i+1] == '/' || s[i+1] == '!') {
				return "", &UnsupportedError{at.Path, at.Line, "inline HTML or autolink", "wrap it in a code span or write it as [text](href): " + quote(clip(s[i:]))}
			}
			b.WriteString("&lt;")
			i++
		default:
			j := i
			for j < len(s) && !strings.ContainsRune("\\`![*_<", rune(s[j])) {
				j++
			}
			if j == i {
				j = i + 1
			}
			b.WriteString(escapeText(s[i:j]))
			i = j
		}
	}
	return b.String(), nil
}

// The three things a '[' can turn out to be.
const (
	// linkInline is `[text](href)`: rendered.
	linkInline = iota
	// linkRefused is a construct that MEANS a link but is not in the subset — a
	// reference link, a footnote, a link title. Rendering it literally would
	// publish the markup; dropping it would lose the destination.
	linkRefused
	// linkLiteral is a bracket that opens no link at all. It is text.
	linkLiteral
)

// parseLink classifies a '[' at i and, for an inline link, returns its text,
// its href, and the index just past the closing paren.
func parseLink(s string, i int) (text, href string, next, kind int) {
	depth := 0
	j := i
	for ; j < len(s); j++ {
		switch s[j] {
		case '[':
			depth++
		case ']':
			depth--
		}
		if depth == 0 {
			break
		}
	}
	if j >= len(s) || j+1 >= len(s) {
		return "", "", 0, linkLiteral
	}
	switch s[j+1] {
	case '(':
		// An inline link, or an attempt at one.
	case '[':
		// `[text][ref]` — a reference link, whose destination is defined
		// elsewhere in a form this renderer does not read.
		return "", "", 0, linkRefused
	default:
		return "", "", 0, linkLiteral
	}
	text = s[i+1 : j]
	k := j + 2
	pd := 1
	for ; k < len(s); k++ {
		if s[k] == '(' {
			pd++
		} else if s[k] == ')' {
			pd--
			if pd == 0 {
				break
			}
		}
	}
	if k >= len(s) {
		return "", "", 0, linkRefused
	}
	href = strings.TrimSpace(s[j+2 : k])
	if href == "" || strings.ContainsAny(href, " \t") {
		// A link title (`[t](u "title")`) is not in the subset: it renders as an
		// attribute nothing on the site reads, and hiding prose in one would
		// smuggle unsourced text onto the page.
		return "", "", 0, linkRefused
	}
	return text, href, k + 1, linkInline
}

// executableScheme reports whether an href names a scheme that executes rather
// than navigates. Escaping an attribute is not enough on its own: a perfectly
// well-formed `javascript:` href needs no quote to break out of, and the site
// renders text from a repository whose files an outside contributor can edit.
// The comparison folds case and strips the whitespace and control characters a
// browser ignores inside a scheme, because those are exactly what a bypass is
// written with.
func executableScheme(href string) (string, bool) {
	var b strings.Builder
	for i := 0; i < len(href); i++ {
		c := href[i]
		if c == ':' {
			scheme := strings.ToLower(b.String())
			switch scheme {
			case "javascript", "vbscript", "data":
				return scheme + ":", true
			}
			return "", false
		}
		if c == '/' || c == '?' || c == '#' {
			return "", false
		}
		if c <= ' ' || c == 0x7f {
			continue
		}
		b.WriteByte(c)
	}
	return "", false
}

// underscoreDelimits reports whether an underscore run at i opens or closes
// emphasis rather than sitting inside a word — `snake_case` is a word, not
// emphasis, and Markdown's own readers agree.
func underscoreDelimits(s string, i, n int) bool {
	before := byte(' ')
	if i > 0 {
		before = s[i-1]
	}
	after := byte(' ')
	if i+n < len(s) {
		after = s[i+n]
	}
	return !isWord(before) || !isWord(after)
}

func isWord(c byte) bool {
	return isAlpha(c) || (c >= '0' && c <= '9')
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isEscapable(c byte) bool {
	return strings.IndexByte("\\`*_{}[]()#+-.!|<>&\"'~", c) >= 0
}

// runLen counts the run of c starting at i.
func runLen(s string, i int, c byte) int {
	n := 0
	for i+n < len(s) && s[i+n] == c {
		n++
	}
	return n
}

// escapeText escapes a text node.
func escapeText(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// escapeAttr escapes an attribute value.
func escapeAttr(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;").Replace(s)
}

func quote(s string) string { return `"` + s + `"` }

func clip(s string) string {
	const max = 40
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}
