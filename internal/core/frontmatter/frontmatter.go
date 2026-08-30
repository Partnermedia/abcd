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

// IsDelimiter reports whether a line is a frontmatter `---` delimiter.
//
// This is the ONE delimiter rule. It is deliberately tolerant of trailing
// whitespace, and deliberately intolerant of everything else: `----`,
// `--- yaml` and an indented `  ---` are not delimiters, because leading
// whitespace survives the trim.
//
// It does NOT trim a BOM, and must not. U+FEFF is a byte-order mark only at
// byte 0 of a file; anywhere else it is ZERO WIDTH NO-BREAK SPACE, an ordinary
// character, and a line spelled "\ufeff---" is a body line to every other
// reader. Trimming it here made Fields close a block early at such a line while
// intent's writer read on to the next bare `---` and inserted keys into the body
// — a lint-green record, a write that reports success, and a value invisible on
// reload (iss-2608270926036966). The BOM is a property of the FILE's first
// position, not of the character, so callers strip it from line 0 themselves
// (see TrimBOM) before asking this predicate anything.
//
// Tolerance here is about the delimiter LINE, not about the block's content —
// capture's strictness about what is inside the block (duplicate top-level keys
// refused, indented lines refused, a restricted YAML subset) is a separate and
// deliberate policy, unaffected by this predicate.
//
// Every reader of a record's bytes comes here rather than re-deriving the
// compare. capture kept a private byte-exact match (`--- ` was not a delimiter)
// while record-lint's ledger gate and the lifeboat graveyard read the same file
// through Fields, which trims: a trailing-space delimiter produced a lint-green
// issue file that every capture verb refused as malformed and that `abcd iss-N`
// reported as "not found in the issue ledger" while it sat in open/ — one format,
// two parsers, opposite verdicts, with the permissive one on the gate side
// (GitHub #338, the class iss-69 opened and left capture out of).
//
// The line may still carry its end-of-line bytes: callers that split keeping
// ends (internal/core/capture) pass them in, and callers that split on "\n" do
// not, so both are trimmed.
func IsDelimiter(line string) bool {
	return strings.TrimRight(line, " \t\r\n") == "---"
}

// Fields returns the top-level keys of the leading frontmatter block (the block
// between the first two `---` lines). Nested keys and list items are ignored,
// and the first occurrence of a key wins. An input whose first line is not `---`
// (or is empty) yields no fields. An unclosed block — a leading `---` with no
// closing `---` — is treated as no frontmatter (an empty map), so body prose is
// never harvested as fields.
func Fields(lines []string) map[string]Field {
	fields := map[string]Field{}
	// A delimiter line may carry trailing whitespace ("--- "); IsDelimiter trims
	// it, so a trailing-space closing delimiter is still seen as the close and
	// body lines after it do not leak in as fields. The BOM is trimmed HERE, at
	// line 0 — the only position where U+FEFF is a byte-order mark rather than an
	// ordinary ZWNBSP character (iss-2608270926036966).
	if len(lines) == 0 || !IsDelimiter(TrimBOM(lines[0])) {
		return fields
	}
	closed := false
	for i := 1; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		if IsDelimiter(line) {
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
	// The BOM is trimmed at line 0 only; see Fields.
	if len(lines) == 0 || !IsDelimiter(TrimBOM(lines[0])) {
		return nil
	}
	seen := map[string]bool{}
	var dups []Dup
	closed := false
	for i := 1; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		if IsDelimiter(line) {
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

// Split separates a record file's leading frontmatter block from its BODY. The
// two returned strings concatenate back to the input byte for byte: head runs
// from the opening delimiter through the closing one and its line ending, and
// body is everything after it. Nothing is trimmed, so a caller that splices an
// edited body back onto head gets the file it read.
//
// It exists so a record's writer and its readers can judge the SAME bytes. A
// writer handed the whole file looks for its section in the frontmatter too, and
// a `# Grounds` line there is a legal YAML comment \u2014 skipped by the block parser
// and matched by an ATX heading pattern \u2014 so the writer wrote into a
// pseudo-section the body reader never consults, agreed with itself on the
// read-back, and reported success about a value nothing could read
// (iss-2608301805069999).
//
// A text with no opening delimiter, and a block nothing closes, are BOTH all
// body: there is no frontmatter to hold back, and holding back prose that no
// reader treats as frontmatter would hide it from the caller that asked for the
// body. The block's interior is not parsed here \u2014 which lines close it is
// IsDelimiter's rule, and an indented line is never a close, exactly as the
// strict ledger parser reads it.
func Split(text string) (head, body string) {
	lines := strings.SplitAfter(text, "\n")
	if len(lines) == 0 || !IsDelimiter(TrimBOM(lines[0])) {
		return "", text
	}
	n := len(lines[0])
	for i := 1; i < len(lines); i++ {
		ln := lines[i]
		if !strings.HasPrefix(ln, " ") && !strings.HasPrefix(ln, "\t") && IsDelimiter(ln) {
			return text[:n+len(ln)], text[n+len(ln):]
		}
		n += len(ln)
	}
	return "", text
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

// Unquote reverses the backslash escaping a double-quoted frontmatter scalar
// carries — the mirror of the escaping capture's yamlScalar emits, where a
// backslash and a double quote are each written `\`-prefixed.
//
// This is the ONE decoder, for the reason IsNull above is the one null
// predicate. capture's reader and record-lint's schema gate each held a
// byte-identical private copy of this loop (iss-2608301212424896), which is the
// split-verdict shape in waiting: the gate exists to refuse exactly what the
// reader refuses, and two decoders that drift make a record the reader SKIPS go
// lint-green. Callers come here rather than re-deriving it.
//
// The argument is the scalar's INNER text, with the surrounding quotes already
// removed: whether a value is double-quoted at all is the caller's question,
// and each caller answers it differently for its own reasons.
func Unquote(s string) string {
	var b strings.Builder
	esc := false
	for _, r := range s {
		if esc {
			b.WriteRune(r)
			esc = false
			continue
		}
		if r == '\\' {
			esc = true
			continue
		}
		b.WriteRune(r)
	}
	if esc {
		b.WriteRune('\\')
	}
	return b.String()
}
