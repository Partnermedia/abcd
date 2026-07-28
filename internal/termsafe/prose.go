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
//   - `<!--` and `-->` are broken apart, so prose can neither open nor close an
//     HTML comment and swallow — or expose — the record around it.
//
// Two forms exist because two callers legitimately want different whitespace
// handling, and the difference is visible in the record: CleanProse trims,
// CleanProseLine collapses every run to one space. A field landing in a
// line-structured file (a bullet, a table cell) wants the line form.
//
// This is the canonical home: internal/core/lifeboat and internal/core/release
// route through it rather than keep divergent copies, and a new trust boundary
// (internal/core/ideate) reuses it instead of writing a third.

import "strings"

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
func cleanProse(s string, capBytes int, normalise func(string) string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "<!--", "< !--")
	s = strings.ReplaceAll(s, "-->", "-- >")
	s = Sanitize(s)
	s = normalise(s)
	if len(s) > capBytes {
		s = strings.ToValidUTF8(s[:capBytes], "")
		s = strings.TrimSpace(s)
	}
	return s
}
