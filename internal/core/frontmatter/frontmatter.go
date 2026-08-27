// Package frontmatter is abcd's shared markdown-frontmatter line scanner. It is
// deliberately a line scanner, not a YAML parser: it reads only the top-level
// keys of the leading `---`…`---` block, first key wins, and pulls in zero
// dependencies. It exists so its consumers (internal/core/spec,
// internal/core/intent, and record-lint's top-level frontmatter checks) share ONE
// copy of this primitive rather than each keeping a private replica.
//
// It is transport-agnostic: no stdout, no os.Exit, no filesystem access — the
// caller supplies the file's lines and decides what the fields mean.
package frontmatter

import (
	"regexp"
	"strings"
)

// keyRe matches a top-level frontmatter key (column 0, no indentation). The
// optional whitespace before the colon matches how the strict ledger parser
// (internal/core/capture) reads a key — it splits on the first colon and trims
// the key — so `id : value` is the key it plainly is, to BOTH readers. Without
// this the lenient scanner read `id :` as no key at all while capture read the
// id fine, and a gate armed on that key (record_schema's filename↔id blocker)
// was silently disarmed by a stray space (GitHub #357).
var keyRe = regexp.MustCompile(`^([A-Za-z0-9_]+)[ \t]*:(.*)$`)

// Field is a frontmatter key's value and its 1-based source line.
type Field struct {
	Value string
	Line  int
}

// Fields returns the top-level keys of the leading frontmatter block (the block
// between the first two `---` lines). Nested keys and list items are ignored,
// and the first occurrence of a key wins. An input whose first line is not `---`
// (or is empty) yields no fields. An unclosed block — a leading `---` with no
// closing `---` — is treated as no frontmatter (an empty map), so body prose is
// never harvested as fields.
func Fields(lines []string) map[string]Field {
	fields := map[string]Field{}
	// A delimiter line may carry trailing whitespace ("--- "); trim spaces/tabs/CR
	// before comparing, so a trailing-space closing delimiter is still seen as the
	// close and body lines after it do not leak in as fields.
	if len(lines) == 0 || strings.TrimRight(TrimBOM(lines[0]), " \t\r") != "---" {
		return fields
	}
	closed := false
	for i := 1; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		if strings.TrimRight(line, " \t") == "---" {
			closed = true
			break
		}
		m := keyRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		key := m[1]
		if _, exists := fields[key]; !exists {
			fields[key] = Field{Value: strings.TrimSpace(m[2]), Line: i + 1}
		}
	}
	// Contract: the block is delimited by the first TWO `---` lines. Without a
	// closing delimiter there is no block, so a document whose leading `---` is
	// really a thematic break (or whose close was fat-fingered, e.g. "----") does
	// not leak column-0 body lines in as phantom fields.
	if !closed {
		return map[string]Field{}
	}
	return fields
}

// Dup is a duplicated top-level key and the 1-based line of its SECOND (the
// offending) occurrence.
type Dup struct {
	Key  string
	Line int
}

// Duplicates returns the top-level keys that appear more than once in the leading
// frontmatter block, one Dup per extra occurrence, in source order.
//
// Fields keeps the FIRST occurrence and drops the rest silently — correct for a
// reader that wants one value, but it means a second `impact:` line hides the
// value a gate is armed to reject, and the strict ledger parser
// (internal/core/capture) refuses such a file outright (last-wins there would
// diverge from the first-occurrence rewrite setScalarField performs). A gate
// built on Fields calls Duplicates so it can refuse exactly what its consumer
// refuses (GitHub #357). The block-boundary contract is Fields': no leading `---`
// or no closing `---` yields nothing, so body prose is never scanned.
func Duplicates(lines []string) []Dup {
	if len(lines) == 0 || strings.TrimRight(TrimBOM(lines[0]), " \t\r") != "---" {
		return nil
	}
	seen := map[string]bool{}
	var dups []Dup
	closed := false
	for i := 1; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		if strings.TrimRight(line, " \t") == "---" {
			closed = true
			break
		}
		m := keyRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		key := m[1]
		if seen[key] {
			dups = append(dups, Dup{Key: key, Line: i + 1})
			continue
		}
		seen[key] = true
	}
	if !closed {
		return nil
	}
	return dups
}

// utf8BOM is U+FEFF, the byte-order mark some editors prepend to a file. It is
// not Unicode White_Space, so strings.TrimSpace leaves it in place.
const utf8BOM = "\ufeff"

// TrimBOM removes a single leading UTF-8 byte-order mark from s. The mark only
// ever appears as the file's very first bytes, so callers apply this to the
// first line before testing it for the opening `---`: a BOM sits invisibly
// ahead of the delimiter and, untrimmed, makes a well-formed record read as
// having no frontmatter — every frontmatter-keyed gate then passes it silently.
func TrimBOM(s string) string {
	return strings.TrimPrefix(s, utf8BOM)
}

// IsNull reports whether a frontmatter scalar is a YAML null.
//
// The set is the YAML 1.2 core schema's, exactly: the empty value, "~", and the
// three spellings "null", "Null" and "NULL". It is deliberately NOT a
// case-insensitive compare — YAML does not accept "nUlL", and an EqualFold here
// would make abcd read records no YAML parser would agree with, which is worse
// than the miss it fixes (iss-287, reported as GitHub #290).
//
// This is the ONE null predicate. internal/core/lint held a private copy that
// recognised only the lower-case spelling, so `impact: NULL` read as null in one
// gate and as a malformed impact in the other — the split-verdict shape that
// makes a record pass a lint and then fail the command that acts on it. Callers
// come here rather than re-deriving it.
//
// The value must be the RAW scalar, before unquoting. Quoting is what separates
// a null from a string in YAML: bare null is a null, and "null" with its quotes
// is the three-character string. A caller that unquotes first destroys that
// distinction and cannot get it back.
func IsNull(v string) bool {
	switch v {
	case "", "~", "null", "Null", "NULL":
		return true
	}
	return false
}
