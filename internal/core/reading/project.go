package reading

import (
	"fmt"
	"html"
	"path"
	"regexp"
	"strings"

	"github.com/intentdriven/abcd/internal/core/frontmatter"
	"github.com/intentdriven/abcd/internal/core/site"
)

// Field projection: the heading-scoped extractor this package owns.
//
// Enumeration of the records comes from the record graph, which reports a
// record's id, store, bucket and path and no body at all. Projection is
// therefore a read of the file the graph named, not a second reading of the
// record's shape — which is the distinction that keeps one parser of the
// record in this binary.
//
// A field resolves as a heading section where the file carries a heading of
// that name, and otherwise as a frontmatter key. Nothing else resolves: a field
// the file does not carry contributes no item, which is what lets one
// projection describe an intent whose sections the record is still growing.

// trimBlankEdges joins a section body, dropping the blank lines at either end so
// a projected field is the text and not the whitespace around it.
func trimBlankEdges(lines []string) string {
	start, end := 0, len(lines)
	for start < end && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return strings.Join(lines[start:end], "\n")
}

// redactExcluded removes the exclusion floor's key-signalled and
// heading-signalled material from a document before anything is taken out of
// it.
//
// It runs on every admitted file, projected or whole. Positive inclusion at
// file granularity is not enough on its own: a brief chapter and a spec travel
// whole, and a frontmatter key stamped on every record by the command that mints
// it would ride along inside them. The floor names those keys and headings
// precisely so the signal is mechanical, and this is where the signal is read.
func redactExcluded(rel, doc string, exclusions []Exclusion) (string, error) {
	// Both signals are RECORD shapes, and only a markdown file can carry one. A
	// Go file has no frontmatter and no Audit Notes heading; what it can have is
	// a raw string literal opening a fence at the left margin, which the section
	// scan rightly refuses as an unterminated block — and refusing it there would
	// let one unrelated source file stop every assembly the repository can run.
	// The scope of the signal is the scope of the parse.
	if !strings.EqualFold(path.Ext(rel), ".md") {
		return doc, nil
	}
	keys := map[string]bool{}
	headings := map[string]bool{}
	for _, e := range exclusions {
		switch e.Signal {
		case "frontmatter key":
			keys[e.Detail] = true
		case "heading":
			headings[e.Detail] = true
		}
	}
	lines := strings.Split(doc, "\n")
	drop := make([]bool, len(lines))

	// The key signal. A key's continuation lines are indented, so dropping the
	// indented run below it removes a block value whole rather than leaving its
	// body behind as orphaned prose.
	for key, field := range frontmatter.Fields(lines) {
		if !keys[key] {
			continue
		}
		i := field.Line - 1
		if i < 0 || i >= len(lines) {
			continue
		}
		drop[i] = true
		// A block scalar's continuation lines are indented, and a blank line
		// INSIDE one is still part of it — stopping at the first blank leaves the
		// rest of the value sitting in the frontmatter. The run ends at the first
		// non-blank line that is not indented.
		for j := i + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == "" {
				drop[j] = true
				continue
			}
			if lines[j][0] != ' ' && lines[j][0] != '\t' {
				break
			}
			drop[j] = true
		}
	}

	// The heading signal, over the fence-aware section scan so a `#` inside a
	// code block is not mistaken for one.
	body, offset := site.StripFrontmatter(doc)
	sections, err := site.Sections(rel, body, offset)
	if err != nil {
		return "", fmt.Errorf("reading: reading the sections of %s: %w", rel, err)
	}
	for i, sec := range sections {
		if sec.Level == 0 || !headings[normaliseHeadingTitle(sec.Title)] {
			continue
		}
		start, end := sectionSpan(sections, i, len(lines))
		for j := start; j < end; j++ {
			if j >= 0 {
				drop[j] = true
			}
		}
	}

	kept := make([]string, 0, len(lines))
	for i, line := range lines {
		if !drop[i] {
			kept = append(kept, line)
		}
	}
	out := strings.Join(kept, "\n")
	if err := verifyRedaction(rel, doc, out, keys, headings); err != nil {
		return "", err
	}
	return out, nil
}

var (
	// excludedKeyLineRe matches an excluded key inside a frontmatter block.
	//
	// Two spellings the field reader does not report, and so does not redact,
	// are matched here instead. A QUOTED key is the same key. And a block whose
	// keys all carry leading indent is valid YAML that the reader — which wants
	// a key at column 0 — looks straight past, so `origin` would travel intact.
	// Matching the indent means a nested mapping's key is matched too; that is
	// the fail-closed direction, and an excluded name nested under another key
	// is not a shape any record in this corpus has.
	excludedKeyLineRe = regexp.MustCompile(`^\s*(?:"([A-Za-z_][A-Za-z0-9_-]*)"|'([A-Za-z_][A-Za-z0-9_-]*)'|([A-Za-z_][A-Za-z0-9_-]*))\s*:`)
	// atxCloseRe matches an ATX heading's optional closing sequence. `## X ##`
	// and `## X` are one heading, and the section scan reports the closing hashes
	// as part of the title, so they are normalised away before any comparison.
	atxCloseRe = regexp.MustCompile(`\s+#+\s*$`)
	// setextRuleRe matches the underline that turns the line above it into a
	// heading. The section scan does not model setext at all.
	setextRuleRe = regexp.MustCompile(`^\s{0,3}(=+|-+)\s*$`)
	// indentedATXRe matches an ATX heading carrying the one-to-three-space indent
	// CommonMark allows. The section scan anchors its pattern at column 0, so it
	// reads such a line as prose — while every renderer, and every human, reads
	// it as the heading it is. Four spaces would make it an indented code block,
	// which is why the bound is three.
	// The indent is SPACES, one to three. A tab makes an indented code block,
	// not a heading, so `\s` would refuse a line no renderer treats as one.
	indentedATXRe = regexp.MustCompile(`^[ ]{1,3}#{1,6}\s+(.*)$`)
	// rawHeadingOpenRe matches a raw HTML heading's OPENING tag, attributes and
	// self-closing slash included. Matching the opening tag alone, rather than a
	// closed pair, is what covers the unclosed, self-closing and multi-line forms
	// — and a document does not have to be well-formed for a reader to see a
	// heading in it.
	rawHeadingOpenRe = regexp.MustCompile(`(?is)<h[1-6](?:\s[^>]*)?/?>`)
	// rawHeadingEndRe bounds the text such a tag introduces: its own close, the
	// next heading tag, or a blank line.
	rawHeadingEndRe = regexp.MustCompile(`(?is)</h[1-6]\s*>|<h[1-6](?:\s[^>]*)?/?>|\n[ \t]*\n`)
	// unresolvableFrontmatterRe matches the YAML constructions whose keys this
	// package cannot resolve without becoming a YAML parser: a tag, an anchor,
	// and an explicit key whose name is a block scalar.
	unresolvableFrontmatterRe = regexp.MustCompile(`^\s*(!!|&\S|\?\s*[|>])`)
	// htmlCommentRe and htmlTagRe strip the markup a title can carry without
	// changing how it reads on the page.
	htmlCommentRe = regexp.MustCompile(`(?s)<!--.*?-->`)
	// The tag name is bounded so an AUTOLINK is left alone: `<https://x>` looks
	// like a tag until the colon, and stripping it turns a heading carrying a URL
	// into a different heading.
	htmlTagRe = regexp.MustCompile(`</?[A-Za-z][A-Za-z0-9-]*(?:\s[^>]*)?/?>`)
	// mdLinkRe unwraps `[text](target)` to the text a reader sees.
	mdLinkRe = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	// explicitYAMLKeyRe matches YAML's explicit-key form, `? origin`.
	explicitYAMLKeyRe = regexp.MustCompile(`^\s*\?\s+["']?([A-Za-z_][A-Za-z0-9_-]*)["']?\s*$`)
	// flowKeyRe matches a key inside a flow mapping, at top level or nested.
	flowKeyRe = regexp.MustCompile(`[{,]\s*["']?([A-Za-z_][A-Za-z0-9_-]*)["']?\s*:`)
	// doubleQuotedKeyRe captures a double-quoted key's raw spelling, escapes and
	// all, so escapedQuotedKey can judge it.
	doubleQuotedKeyRe = regexp.MustCompile(`^\s*"([^"]*)"\s*:`)
	// fenceOpenRe matches a fenced code block's delimiter, on the section scan's
	// own rule so the two agree about what is inside a fence.
	fenceOpenRe = regexp.MustCompile("^[ \t]*```")
)

// Two shapes this floor does NOT see, disclosed rather than claimed. A heading
// nested inside a blockquote or a list item is indented and prefixed, so neither
// the section scan nor the raw-line patterns read it as a heading. And a title
// reaching the excluded one through a homoglyph or an invisible format character
// slugs differently by construction, because the slug compares code points. Both
// are residue; neither is caught.
//
// namesExcludedHeading reports whether a heading title is one of the excluded
// ones, under the ONE equality this floor uses: a case fold, or the same
// rendering. It exists so the three refusal paths — the section scan, the
// indented ATX line, the setext underline — cannot drift apart on what "the same
// heading" means. They did: the render comparison was added on the first path
// only, which closed the class on one of three.
func namesExcludedHeading(title string, headings map[string]bool) (string, bool) {
	for want := range headings {
		if strings.EqualFold(title, want) || sameRendering(title, want) {
			return want, true
		}
	}
	return "", false
}

// sameRendering reports whether two heading titles come out as the same heading
// on the page. `## **Audit Notes**`, "## `Audit Notes`" and a title carrying a
// non-breaking space differ in bytes and are the same heading to every reader,
// so a byte comparison is the wrong test for what the floor is trying to name.
//
// The site's own anchor slug is the comparison: it drops emphasis and code
// marks, lower-cases, and collapses every other run of non-alphanumerics to a
// hyphen — which is exactly the equivalence "renders as the same heading" needs,
// and it is one function rather than a table of markup shapes to keep current.
func sameRendering(a, b string) bool {
	slug := site.Slug(renderedText(a))
	return slug != "" && slug == site.Slug(renderedText(b))
}

// renderedText reduces a heading title to the text a reader sees: HTML comments
// and tags removed, link wrappers unwrapped to their label, and character
// references decoded. The slug then compares what the page shows rather than
// what the source happens to spell.
//
// Decoding is html.UnescapeString, one pass over the whole string. A hand list
// of entities applied by ranging a map was not merely incomplete — it was
// NONDETERMINISTIC: `Audit&amp;nbsp;Notes` decoded to the excluded title or not
// depending on whether `&amp;` was applied before `&nbsp;` that time round, so a
// determinism instrument had a coin-flip refusal. One pass also covers the
// numeric and hex character references a short list could never enumerate.
func renderedText(title string) string {
	out := htmlCommentRe.ReplaceAllString(title, "")
	out = mdLinkRe.ReplaceAllString(out, "$1")
	out = htmlTagRe.ReplaceAllString(out, "")
	return strings.TrimSpace(html.UnescapeString(out))
}

// normaliseHeadingTitle reduces a heading to the text it names: surrounding
// whitespace and the optional ATX closing sequence removed.
func normaliseHeadingTitle(title string) string {
	return strings.TrimSpace(atxCloseRe.ReplaceAllString(strings.TrimSpace(title), ""))
}

// fenceMask reports, per line, whether that line sits inside a fenced code
// block. It answers a LINE-level question the section scan does not expose —
// that scan reports where the headings are, having already skipped the fences —
// and the two scans share one fence rule so they cannot disagree about it.
//
// It is what keeps this floor from firing on a document that merely SHOWS the
// record template: a fenced example carries frontmatter and headings that are
// examples, not fields, and refusing them would stop every assembly the
// repository can run.
func fenceMask(lines []string) []bool {
	mask := make([]bool, len(lines))
	inside := false
	for i, line := range lines {
		if fenceOpenRe.MatchString(line) {
			inside = !inside
			mask[i] = true // the delimiter itself belongs to the block
			continue
		}
		mask[i] = inside
	}
	return mask
}

// verifyRedaction is the key-and-heading half of the exclusion floor made
// fail-closed, the way the path half already is.
//
// Redaction is a positive act over what a parser reported, and three shapes slip
// past a parser that reports one value per key and matches a title exactly. A
// DUPLICATED key keeps its second copy, because Fields keeps the first
// occurrence and drops the rest silently. A frontmatter block closed with four
// dashes is cut from the body by StripFrontmatter, which closes on a `---`
// PREFIX, while Fields wants the delimiter exactly and so reads no fields at all.
// And a heading spelled in another case is not the title the redactor looked for.
//
// In each case the field travels and the manifest still asserts it was refused,
// which is the one thing the manifest exists not to do. A floor a file can
// quietly walk through is a disclosure, not a gate — so a file that still
// carries an excluded shape after redaction refuses the run and names the shape.
func verifyRedaction(rel, original, redacted string, keys, headings map[string]bool) error {
	lines := strings.Split(redacted, "\n")
	fenced := fenceMask(lines)

	if len(keys) > 0 {
		for _, dup := range frontmatter.Duplicates(strings.Split(original, "\n")) {
			if keys[dup.Key] {
				return fmt.Errorf("reading: %s declares the excluded key %q more than once (line %d); "+
					"only the first occurrence is redactable, so the rest would travel", rel, dup.Key, dup.Line)
			}
		}
		if line, shape, ok := unresolvableFrontmatterShape(lines, fenced); ok {
			return fmt.Errorf("reading: %s uses the YAML construction %q at line %d in its frontmatter, "+
				"whose keys this package cannot resolve without becoming a YAML parser; a record has no "+
				"reason to use one, so it is refused rather than guessed at", rel, shape, line)
		}
		if line, key, ok := excludedKeyInFirstBlock(lines, fenced, keys); ok {
			return fmt.Errorf("reading: %s still carries the excluded key %q at line %d after redaction; "+
				"the frontmatter block is not closed the way the field reader expects it", rel, key, line)
		}
	}
	if len(headings) == 0 {
		return nil
	}

	// The heading check runs over the SAME fence-aware scan the redactor spans
	// by. Reading raw lines instead made this floor fire on a fenced example of
	// the record template — a heading inside a code block is an example, not a
	// field, and the redactor rightly left it alone while this refused the run.
	// Every raw-line scan below starts at `offset`, the first BODY line. Running
	// them over the frontmatter finds constructions that are not what they look
	// like: a block scalar whose last line is the excluded title, sitting above
	// the closing `---`, was refused as an underlined heading — a true refusal
	// reached by a false reading, which teaches whoever hits it the wrong fix.
	body, offset := site.StripFrontmatter(redacted)
	sections, err := site.Sections(rel, body, offset)
	if err != nil {
		return fmt.Errorf("reading: re-reading the sections of %s: %w", rel, err)
	}
	for _, sec := range sections {
		if sec.Level == 0 {
			continue
		}
		if want, ok := namesExcludedHeading(normaliseHeadingTitle(sec.Title), headings); ok {
			return fmt.Errorf("reading: %s still carries the excluded heading %q at line %d after "+
				"redaction; the floor names %q, and a heading is excluded however it is spelled",
				rel, sec.Title, sec.Line, want)
		}
	}

	// An indented ATX heading is refused for the same reason, and in the same
	// place. The section scan does not see one, so the redactor has no span to
	// delete and the section travels whole; widening that scan is not this
	// package's call to make, because the site renderer's own output turns on it.
	// A refusal here costs an edit and names the line; the alternative is a leak
	// under a manifest asserting the opposite.
	for i := offset; i < len(lines); i++ {
		line := lines[i]
		if fenced[i] {
			continue
		}
		m := indentedATXRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		if want, ok := namesExcludedHeading(normaliseHeadingTitle(m[1]), headings); ok {
			return fmt.Errorf("reading: %s indents the excluded heading %q at line %d; the floor "+
				"names %q, and a heading is excluded however it is spelled", rel, strings.TrimSpace(line), i+1, want)
		}
	}

	// A raw HTML heading is refused on the same ground: the site's page reader
	// REFUSES a raw HTML block rather than admitting one, so a heading spelled
	// that way is a heading to every other reader of the file and nothing at all
	// to the markdown scan — and there is again no span for the redactor.
	//
	// The scan runs over the unfenced body JOINED, not line by line, because
	// `<h2>` and its text and its close need not share a line. The match offset
	// maps back to a line so the refusal still names one.
	if line, title, ok := rawHTMLHeading(lines, fenced, offset, headings); ok {
		if want, hit := namesExcludedHeading(title, headings); hit {
			return fmt.Errorf("reading: %s carries the excluded heading %q as raw HTML at line %d; "+
				"the floor names %q, and a heading is excluded however it is spelled",
				rel, title, line, want)
		}
	}

	// Setext headings are a refusal rather than a redaction. The section scan
	// does not model them, so there is no span to delete, and inventing a second
	// heading scanner here to compute one is the second parser this package
	// exists not to grow. A record underlining its Audit Notes is rare and a
	// refusal names it; a leak would not.
	for i := offset; i+1 < len(lines); i++ {
		if fenced[i] || fenced[i+1] || !setextRuleRe.MatchString(lines[i+1]) {
			continue
		}
		title := normaliseHeadingTitle(lines[i])
		if title == "" {
			continue
		}
		if want, ok := namesExcludedHeading(title, headings); ok {
			return fmt.Errorf("reading: %s underlines the excluded heading %q at line %d; the floor "+
				"names %q, and a heading is excluded however it is spelled", rel, lines[i], i+1, want)
		}
	}
	return nil
}

// submatches returns a match's non-empty capture groups, so a pattern spelling
// one name in several alternatives is read the same way as one that does not.
func submatches(m []string) []string {
	if len(m) < 2 {
		return nil
	}
	out := make([]string, 0, len(m)-1)
	for _, g := range m[1:] {
		if g != "" {
			out = append(out, g)
		}
	}
	return out
}

// escapedQuotedKey refuses a double-quoted frontmatter key containing a
// backslash. YAML decodes escapes inside double quotes, so "ori\u0067in" IS
// `origin` to the reader and is nothing at all to a pattern over the bytes.
// Rather than grow a YAML decoder here, the escape itself is the signal: a
// record has no reason to spell a key that way, and refusing is the fail-closed
// answer to a name this package cannot resolve.
func escapedQuotedKey(line string) (string, bool) {
	m := doubleQuotedKeyRe.FindStringSubmatch(line)
	if m == nil || !strings.Contains(m[1], `\`) {
		return "", false
	}
	return m[1], true
}

// rawHTMLHeading finds the first raw HTML heading in the unfenced body whose
// text names an excluded heading, and reports the line it sits on.
//
// The body is joined before scanning because a raw heading need not fit on one
// line: `<h2>`, its text and its close can sit on three. Joining costs the line
// number, so the match offset is mapped back to one by counting newlines ahead
// of it — the refusal has to name a line a human can go and look at.
//
// A fenced line is replaced by an empty line rather than dropped, so the offset
// arithmetic keeps working and an example inside a code block still cannot fire.
func rawHTMLHeading(lines []string, fenced []bool, offset int, headings map[string]bool) (int, string, bool) {
	scan := make([]string, len(lines))
	for i := range lines {
		if i >= offset && !fenced[i] {
			scan[i] = lines[i]
		}
	}
	joined := strings.Join(scan, "\n")

	for _, open := range rawHeadingOpenRe.FindAllStringIndex(joined, -1) {
		rest := joined[open[1]:]
		end := len(rest)
		if b := rawHeadingEndRe.FindStringIndex(rest); b != nil {
			end = b[0]
		}
		title := normaliseHeadingTitle(renderedText(rest[:end]))
		if title == "" {
			continue
		}
		if _, ok := namesExcludedHeading(title, headings); ok {
			return strings.Count(joined[:open[0]], "\n") + 1, title, true
		}
	}
	return 0, "", false
}

// unresolvableFrontmatterShape reports a YAML construction in the first block
// whose keys this package cannot resolve: a tag (`!!`), an anchor (`&`), or an
// explicit key whose name is a block scalar (`? |`). It is the same reasoning as
// the escaped quoted key — resolving these means a YAML parser, a record has no
// reason to use one, so the construction itself is the signal and the answer is
// a refusal rather than a guess.
func unresolvableFrontmatterShape(lines []string, fenced []bool) (int, string, bool) {
	open := -1
	for i, line := range lines {
		if fenced[i] {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if i == 0 {
			trimmed = strings.TrimSpace(frontmatter.TrimBOM(line))
		}
		if strings.HasPrefix(trimmed, "---") {
			open = i
			break
		}
	}
	if open < 0 {
		return 0, "", false
	}
	for i := open + 1; i < len(lines); i++ {
		if fenced[i] {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "---") {
			return 0, "", false
		}
		if m := unresolvableFrontmatterRe.FindStringSubmatch(lines[i]); m != nil {
			return i + 1, strings.TrimSpace(m[1]), true
		}
	}
	return 0, "", false
}

// excludedKeyInFirstBlock reports an excluded key sitting at column 0 inside the
// document's first delimiter-fenced region.
//
// Two loosenesses are deliberate, and each closes a shape the strict reading
// misses. The region is bounded by any line OPENING with three dashes, not by an
// exact `---`, because that is the rule the frontmatter stripper applies and the
// gap between the two rules is where a key survives. And the region is found
// wherever it starts rather than at line 0, because a preamble ahead of the
// block makes the field reader report nothing at all while the keys sit there
// in plain sight.
func excludedKeyInFirstBlock(lines []string, fenced []bool, keys map[string]bool) (int, string, bool) {
	open := -1
	for i, line := range lines {
		if fenced[i] {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if i == 0 {
			trimmed = strings.TrimSpace(frontmatter.TrimBOM(line))
		}
		if strings.HasPrefix(trimmed, "---") {
			open = i
			break
		}
	}
	if open < 0 {
		return 0, "", false
	}
	for i := open + 1; i < len(lines); i++ {
		if fenced[i] {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "---") {
			return 0, "", false
		}
		// Four spellings, because the field reader reports one of them. A plain
		// or quoted key at any indent; YAML's explicit-key form; a key inside a
		// flow mapping at top level or nested; and a double-quoted key whose name
		// is spelled with an escape, which no pattern over the raw bytes can see
		// — so a quoted key carrying a backslash is refused on the backslash
		// rather than on the name it hides.
		for _, m := range [][]string{
			excludedKeyLineRe.FindStringSubmatch(lines[i]),
			explicitYAMLKeyRe.FindStringSubmatch(lines[i]),
		} {
			for _, key := range submatches(m) {
				if keys[key] {
					return i + 1, key, true
				}
			}
		}
		for _, m := range flowKeyRe.FindAllStringSubmatch(lines[i], -1) {
			for _, key := range submatches(m) {
				if keys[key] {
					return i + 1, key, true
				}
			}
		}
		if key, ok := escapedQuotedKey(lines[i]); ok {
			return i + 1, key, true
		}
	}
	return 0, "", false
}

// sectionSpan is the half-open line range one heading OWNS: the heading itself
// through everything under it, ending at the next heading of the same level or
// shallower. Indices are 0-based over the whole document, frontmatter included.
//
// It is shared by the redactor and the projector because the two must agree
// about where a section ends. The section scan's own Body ends at the next
// heading of ANY level, which is right for rendering a page and wrong here: a
// `###` under a projected `##` would be dropped from the item while the redactor
// treated it as part of the section it belongs to, so a field would travel short
// and the manifest would hash the short version.
func sectionSpan(sections []site.Section, i, total int) (int, int) {
	start := sections[i].Line - 1
	for _, next := range sections[i+1:] {
		if next.Level > 0 && next.Level <= sections[i].Level {
			return start, next.Line - 1
		}
	}
	return start, total
}

// projectField extracts one named field from a record's text. Only a record is
// ever projected, and a record is markdown, so the same scope holds here.
func projectField(rel, doc, field string) (string, bool, error) {
	body, offset := site.StripFrontmatter(doc)
	sections, err := site.Sections(rel, body, offset)
	if err != nil {
		return "", false, fmt.Errorf("reading: projecting %s from %s: %w", field, rel, err)
	}
	lines := strings.Split(doc, "\n")
	for i, sec := range sections {
		if sec.Level < 2 || sec.Title != field {
			continue
		}
		start, end := sectionSpan(sections, i, len(lines))
		return trimBlankEdges(lines[min(start+1, len(lines)):min(end, len(lines))]), true, nil
	}
	fields := frontmatter.Fields(strings.Split(doc, "\n"))
	if f, ok := fields[field]; ok && !frontmatter.IsNull(f.Value) {
		return f.Value, true, nil
	}
	return "", false, nil
}
