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
//     open nor close an HTML construct and swallow the record around it;
//   - markdown link syntax is broken apart — an inline link `[t](d)`, an image,
//     and a reference link `[t][l]` — so a faithful quotation of code can never
//     land as a live link in a committed record (an autolink `<https://…>` is
//     already caught by the HTML rule, since its scheme starts with a letter).
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
// cannot know what surrounds the field it is handed. The one place it does not
// fire is inside a CommonMark code span, where no HTML is parsed and the
// neutralisation only corrupts: a documented `--reading-json <path>` was
// written as `< path>`, a shell input redirection (iss-2609011217083577). Span
// boundaries follow the spec — a run of N backticks opens a span closed by the
// next run of exactly N, a backslash-escaped backtick is literal, and a run
// with no closer is literal too — so an unbalanced backtick leaves everything
// after it prose, and the primitive fails closed. The remaining cost is a
// fidelity regression on angle-bracket prose OUTSIDE a span — a bare `<repo>`
// placeholder reads `< repo>` — which is accepted: a slightly uglier
// placeholder is cheaper than a record that can conceal its own contents.
//
// The link rule is not a spoofing defence but a gate one (iss-2608311504353427):
// an auditor quoting an element path of the form `items[0](itm-0001)` writes an
// inline link whose target resolves to nothing, and record-lint's links_resolve
// rule then refuses the whole tree on a record the ingest itself just wrote — a
// failure no care in the author can prevent, because the offending text is a
// faithful quotation of the code under review. The neutralisation is the one the
// comment delimiters get, a single space breaking the adjacency CommonMark
// requires, so `items[0] (itm-0001)` still reads as the code it quotes. Only the
// adjacency is touched: a bracket or a parenthesis on its own is left alone, and
// a shortcut reference `[label]` is not neutralised, because doing so would have
// to rewrite every bracket in every field and it opens no link unless a matching
// definition exists in the surrounding document.
//
// The link rule does NOT take the code-span exemption the HTML rule takes, and
// the asymmetry follows from what each one defends. The HTML rule defends the
// RENDER, and a renderer parses no raw HTML inside a code span, so neutralising
// there buys nothing and corrupts the quoted content. The link rule defends a
// GATE, and the gate does not read the same grammar: record-lint's `checkLinks`
// masks fenced blocks only, so an inline span is scanned like any other prose
// and a `](` left inside one still refuses the whole tree. Until the gate
// exempts spans, the cleaner cannot.
//
// Every rule inserts a form the same rule no longer matches, so the cleaner is
// idempotent: a field cleaned twice is the field cleaned once, and a downstream
// caller that re-cleans what it was handed does not drift it further.
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
	s = neutraliseOutsideCodeSpans(s)
	// The link rule fires everywhere, code spans included, and that asymmetry
	// with the HTML rule is the difference between the two defences. The HTML
	// rule protects the RENDER, and a renderer parses no HTML inside a span, so
	// firing there would corrupt content for nothing. The link rule protects a
	// GATE, and record-lint's checkLinks masks fenced blocks only — an inline
	// span is scanned like any other prose — so a `](` left inside a span still
	// refuses the tree. A space after the `]` is enough: CommonMark requires the
	// destination `(` or the label `[` to follow the link text immediately, and
	// the links_resolve pattern requires the same adjacency.
	s = strings.ReplaceAll(s, "](", "] (")
	s = strings.ReplaceAll(s, "][", "] [")
	s = normalise(s)
	// The cap is applied last, and a cut can land inside a code span: the closing
	// backticks go, the span stops being one, and its untouched content is prose
	// again with its openers live. So every cut is followed by another HTML
	// neutralisation pass, and the pair repeats until the result fits — the pass
	// is idempotent and each cut retains strictly fewer of the original bytes, so
	// it ends. The link rule needs no second pass: a cut removes trailing bytes
	// and trims the ends, neither of which can make two characters adjacent that
	// were not adjacent before. A cut landing mid-rune drops the partial rune
	// rather than emitting replacement bytes.
	for len(s) > capBytes {
		s = strings.TrimSpace(strings.ToValidUTF8(s[:capBytes], ""))
		s = neutraliseOutsideCodeSpans(s)
	}
	return s
}

// neutraliseHTML is the opener rule over one run of prose. A space after the
// `<` is enough: CommonMark needs the name, `/`, `!`, or `?` to follow
// immediately. This subsumes the comment-open case — `<!--` becomes `< !--` —
// so the two rules are one; the comment-close rule is its pair.
func neutraliseHTML(prose string) string {
	prose = htmlOpenerRe.ReplaceAllStringFunc(prose, func(m string) string { return "< " + m[1:] })
	return strings.ReplaceAll(prose, "-->", "-- >")
}

// neutraliseOutsideCodeSpans applies neutraliseHTML to every part of s that is
// not a CommonMark code span, and copies each span through byte-for-byte. A
// span is a run of N backticks followed, anywhere later, by the next run of
// exactly N; a run with no such closer is literal backticks and the text after
// it is prose. Backslash escapes are honoured outside a span (an escaped
// backtick opens nothing) and ignored inside one (the spec parses none there),
// so the boundary the renderer will draw is the boundary this draws.
func neutraliseOutsideCodeSpans(s string) string {
	var out strings.Builder
	i := 0
	for i < len(s) {
		open := nextBacktickRun(s, i)
		if open < 0 {
			out.WriteString(neutraliseHTML(s[i:]))
			break
		}
		openEnd := open
		for openEnd < len(s) && s[openEnd] == '`' {
			openEnd++
		}
		n := openEnd - open
		closeAt := closingRun(s, openEnd, n)
		if closeAt < 0 {
			out.WriteString(neutraliseHTML(s[i:openEnd]))
			i = openEnd
			continue
		}
		out.WriteString(neutraliseHTML(s[i:open]))
		out.WriteString(s[open : closeAt+n])
		i = closeAt + n
	}
	return out.String()
}

// nextBacktickRun returns the index of the first backtick at or after from that
// is not backslash-escaped — the start of a run that may open a span — or -1.
// An odd number of preceding backslashes escapes the backtick; an even number
// is escaped backslashes followed by a live one.
func nextBacktickRun(s string, from int) int {
	for j := from; j < len(s); j++ {
		if s[j] != '`' {
			continue
		}
		backslashes := 0
		for k := j; k > 0 && s[k-1] == '\\'; k-- {
			backslashes++
		}
		if backslashes%2 == 0 {
			return j
		}
	}
	return -1
}

// closingRun returns the index of the first run of exactly n backticks at or
// after from, or -1. Runs of any other length are span content and skipped
// whole, so a longer run never matches by its prefix.
func closingRun(s string, from, n int) int {
	for j := from; j < len(s); {
		if s[j] != '`' {
			j++
			continue
		}
		k := j
		for k < len(s) && s[k] == '`' {
			k++
		}
		if k-j == n {
			return j
		}
		j = k
	}
	return -1
}
