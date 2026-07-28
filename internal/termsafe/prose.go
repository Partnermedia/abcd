package termsafe

// prose.go is the shared home for the untrusted-prose cleaner every host-delegated
// ingest boundary needs before a model-supplied string is written into a durable
// markdown or JSON record.
//
// It sits beside Sanitize because it IS Sanitize plus the two neutralisations a
// terminal sanitiser has no reason to make, but a FILE writer always does:
//
//   - line breaks become spaces, so one prose field can never forge a second line
//     (a changelog bullet, a markdown table row, a list item) in the record;
//   - anything that could OPEN raw HTML is broken apart — a tag, a closing tag, a
//     declaration, a processing instruction, or a comment — so prose can neither
//     open nor close an HTML construct and swallow the record around it.
//
// The HTML rule is not cosmetic. In CommonMark a `<` followed by a letter, `/`,
// `!`, or `?` at the start of a line begins an HTML BLOCK, and several of those
// block types run to the end of the document if their closing condition never
// arrives: one `<script>` in a claim makes every later section of a verdict record
// — the falsified claims, the grill hits, the adversary's findings — render as
// inert text inside an unclosed element, while a forged table above it renders
// normally. An artefact whose whole value is that a later session trusts it must
// not be able to hide its own evidence.
//
// The rule fires ANYWHERE in the string, not only at a line start, and that is
// deliberate rather than an over-reach: the same swallow was demonstrated from
// INSIDE a markdown table cell, where the opener is mid-line, and the cleaner
// cannot know what surrounds the field it is handed. The cost is a fidelity
// regression on legitimate angle-bracket prose — a changelog line's
// `<repo>` placeholder reads `< repo>` — which is accepted: a slightly uglier
// placeholder is cheaper than a record that can conceal its own contents.
//
// Two forms exist because two callers legitimately want different whitespace
// handling, and the difference is visible in the record: CleanProse trims,
// CleanProseLine collapses every run to one space. A field landing in a
// line-structured file (a bullet, a table cell) wants the line form.
//
// This is the canonical home: internal/core/lifeboat and internal/core/release
// route through it rather than keep divergent copies, and a new trust boundary
// (internal/core/ideate) reuses it instead of writing a third.

import (
	"regexp"
	"strings"
)

// htmlOpenerRe matches every way CommonMark lets a `<` begin raw HTML: a tag, a
// closing tag, a declaration or comment (`<!`), or a processing instruction
// (`<?`). A bare `<` with anything else after it — `a < b`, `<-`, a trailing one —
// opens nothing and is left alone, so ordinary prose reads normally.
var htmlOpenerRe = regexp.MustCompile(`<[A-Za-z!/?]`)

// CleanProse neutralises one untrusted prose field and caps it at capBytes,
// preserving interior whitespace runs. The cap is applied last and the result is
// re-trimmed, so a cut landing mid-word leaves no dangling space; a cut landing
// mid-rune drops the partial rune rather than emitting replacement bytes.
func CleanProse(s string, capBytes int) string {
	return cleanProse(s, capBytes, strings.TrimSpace)
}

// CleanProseLine is CleanProse for a field that must occupy exactly one line:
// every whitespace run collapses to a single space before the cap is applied. Use
// it wherever the prose lands in a file whose line structure is machine-read.
func CleanProseLine(s string, capBytes int) string {
	return cleanProse(s, capBytes, func(v string) string { return strings.Join(strings.Fields(v), " ") })
}

// cleanProse is the shared body: neutralise, sanitise, normalise whitespace the
// caller's way, then cap. Written once so the two forms can only differ in the
// one dimension they are meant to differ in.
// The HTML neutralisation runs AFTER Sanitize, and the order is load-bearing:
// Sanitize SUBSTITUTES a '?' for each masked rune rather than deleting it, so a
// `<` followed by an escape byte becomes `<?` — a processing instruction, which
// is an HTML block that runs until `?>`. Neutralising before Sanitize would
// therefore let a masked control character forge the very construct this closes.
func cleanProse(s string, capBytes int, normalise func(string) string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = Sanitize(s)
	// A space after the `<` is enough: CommonMark needs the name, `/`, `!`, or `?`
	// to follow immediately. This subsumes the comment-open case — `<!--` becomes
	// `< !--` — so the two rules are one.
	s = htmlOpenerRe.ReplaceAllStringFunc(s, func(m string) string { return "< " + m[1:] })
	s = strings.ReplaceAll(s, "-->", "-- >")
	s = normalise(s)
	if len(s) > capBytes {
		s = strings.ToValidUTF8(s[:capBytes], "")
		s = strings.TrimSpace(s)
	}
	return s
}
