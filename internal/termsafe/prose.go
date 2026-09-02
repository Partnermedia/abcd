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
// One exemption exists, and it is opt-in (KeepCodeSpans): CommonMark never
// parses HTML inside a code span, so a placeholder a caller documents as
// `--scope <record-id>` is inert there, while the neutralised `< record-id>` is
// a shell redirect a reader would copy (iss-2609011217083577). The content of a
// CLOSED span is left to the HTML rule's exclusion; Sanitize and the caller's
// whitespace normalisation still apply, since a code span rendered into a
// terminal carries an escape byte exactly as prose does. Why opt-in: the span
// structure the cleaner sees is the field's own, and it is the RENDERED line's
// structure that decides what is inert. A caller that wraps the cleaned field in
// its own backticks ("`%s`") re-pairs every run in it, so the caller must prove
// the field stands alone on its line — nothing it adds before the field opens a
// backtick string, and nothing splits the field — before asking for the
// exemption. The default is the neutralise-everything form for exactly that
// reason.
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

// Option relaxes one neutralisation a caller can prove unnecessary for the line
// its field lands on. Options only ever narrow the rewrite; none adds a rule.
type Option int

// KeepCodeSpans leaves the content of a closed CommonMark code span (6.1: a run
// of N backticks closed by the next run of exactly N) out of the raw-HTML
// neutralisation. Text outside a span, and an unterminated opener's tail, keep
// every rule. The exemption stands down for the whole field when any shape in it
// could make a renderer see different span boundaries from the cleaner: a
// backslash before a backtick (an escape outside a span, so a live backtick to
// one parse and a literal to another) or a pipe (a table row re-splits the field
// around it). A caller asks for this only where the field stands alone on its
// line — see the package comment above.
const KeepCodeSpans Option = iota + 1

// CleanProse neutralises one untrusted prose field and caps it at capBytes,
// preserving interior whitespace runs. The cap is applied last and the result is
// re-trimmed, so a cut landing mid-word leaves no dangling space; a cut landing
// mid-rune drops the partial rune rather than emitting replacement bytes.
func CleanProse(s string, capBytes int, opts ...Option) string {
	return cleanProse(s, capBytes, strings.TrimSpace, opts)
}

// CleanProseLine is CleanProse for a field that must occupy exactly one line:
// every whitespace run collapses to a single space before the cap is applied. Use
// it wherever the prose lands in a file whose line structure is machine-read.
func CleanProseLine(s string, capBytes int, opts ...Option) string {
	return cleanProse(s, capBytes, func(v string) string { return strings.Join(strings.Fields(v), " ") }, opts)
}

// cleanProse is the shared body: neutralise, sanitise, normalise whitespace the
// caller's way, then cap. Written once so the two forms can only differ in the
// one dimension they are meant to differ in.
// The HTML neutralisation runs AFTER Sanitize, and the order is load-bearing:
// Sanitize SUBSTITUTES a '?' for each masked rune rather than deleting it, so a
// `<` followed by an escape byte becomes `<?` — a processing instruction, which
// is an HTML block that runs until `?>`. Neutralising before Sanitize would
// therefore let a masked control character forge the very construct this closes.
//
// The neutralise-normalise-cap sequence runs to a fixed point rather than once,
// and that too is load-bearing once code spans are exempt: the cap can cut inside
// a span, which turns its opener into an unterminated one and its content — left
// raw because the span was closed — into prose. Re-running the sequence on the
// cut result neutralises what the cut exposed and re-caps; each round either
// changes nothing (done) or pushes the cut earlier into the field, so it ends.
func cleanProse(s string, capBytes int, normalise func(string) string, opts []Option) string {
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = Sanitize(s)
	neutralise := neutraliseHTML
	if hasOption(opts, KeepCodeSpans) && codeSpansAreUnambiguous(s) {
		neutralise = neutraliseHTMLOutsideCodeSpans
	}
	for {
		prev := s
		s = normalise(neutralise(s))
		if len(s) > capBytes {
			s = strings.ToValidUTF8(s[:capBytes], "")
			s = strings.TrimSpace(s)
		}
		if s == prev {
			return s
		}
	}
}

func hasOption(opts []Option, want Option) bool {
	for _, o := range opts {
		if o == want {
			return true
		}
	}
	return false
}

// neutraliseHTML breaks every way a `<` can open raw HTML. A space after the `<`
// is enough: CommonMark needs the name, `/`, `!`, or `?` to follow immediately.
// This subsumes the comment-open case — `<!--` becomes `< !--` — so the two
// rules are one; the comment-close rule is the only other.
func neutraliseHTML(s string) string {
	s = htmlOpenerRe.ReplaceAllStringFunc(s, func(m string) string { return "< " + m[1:] })
	s = strings.ReplaceAll(s, "-->", "-- >")
	// A space after the `]` is enough for links too: CommonMark requires the
	// destination `(` or the label `[` to follow the link text immediately, and
	// the record-lint links_resolve pattern requires the same adjacency.
	s = strings.ReplaceAll(s, "](", "] (")
	return strings.ReplaceAll(s, "][", "] [")
}

// codeSpansAreUnambiguous reports whether the span boundaries the cleaner would
// parse are the ones every renderer of the field will see. Two shapes break that:
// a backslash directly before a backtick, which CommonMark reads as an escape
// outside a span but which a caller's own escaping (an ideate table cell doubles
// backslashes) turns back into a live backtick; and a pipe, which a table row
// splits the field on, re-pairing every run around it. Either one fails the whole
// field closed — it keeps the neutralise-everything form.
func codeSpansAreUnambiguous(s string) bool {
	return !strings.Contains(s, "\\`") && !strings.Contains(s, "|")
}

// neutraliseHTMLOutsideCodeSpans is neutraliseHTML applied to every stretch of s
// that lies outside a closed CommonMark code span, with each span copied through
// byte for byte. The walk is the spec's: a backtick string is a maximal run; a
// run of N opens a span closed by the NEXT run of exactly N; a run with no closer
// is literal, the walk resumes after it, and later runs may still open. A run
// length that has once failed to find a closer never searches again — there is
// no closer of that length further along either — which keeps a hostile field of
// unmatched runs linear.
func neutraliseHTMLOutsideCodeSpans(s string) string {
	var b strings.Builder
	noCloser := map[int]bool{}
	start, i := 0, 0
	for i < len(s) {
		if s[i] != '`' {
			i++
			continue
		}
		n := backtickRun(s, i)
		if noCloser[n] {
			i += n
			continue
		}
		end := closingRun(s, i+n, n)
		if end < 0 {
			noCloser[n] = true
			i += n
			continue
		}
		b.WriteString(neutraliseHTML(s[start:i]))
		b.WriteString(s[i : end+n])
		i = end + n
		start = i
	}
	b.WriteString(neutraliseHTML(s[start:]))
	return b.String()
}

// backtickRun returns the length of the maximal backtick run starting at i.
func backtickRun(s string, i int) int {
	n := 0
	for i+n < len(s) && s[i+n] == '`' {
		n++
	}
	return n
}

// closingRun returns the index of the first backtick run of exactly n at or
// after from, or -1 when there is none.
func closingRun(s string, from, n int) int {
	for j := from; j < len(s); {
		if s[j] != '`' {
			j++
			continue
		}
		run := backtickRun(s, j)
		if run == n {
			return j
		}
		j += run
	}
	return -1
}
