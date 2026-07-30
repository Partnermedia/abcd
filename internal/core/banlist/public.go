package banlist

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/REPPL/abcd-cli/internal/core/lint"
	"github.com/REPPL/abcd-cli/internal/fsutil"
)

// docsLintAllowEscape is the per-line escape every entry in the family declares —
// the same regexp the hand-curated families use, so one comment suppresses any
// banned token on its line and a reader learns one convention, not two.
const docsLintAllowEscape = `(?i)<!--\s*docs-lint:\s*allow\b`

// defaultSuccessor is the machine-readable replacement a verb-written entry
// declares. The family's schema REQUIRES a successor (a ban with no successor left
// its replacement in prose only), so a default is what makes an argless add
// produce a valid entry rather than one the linter refuses to load.
const defaultSuccessor = "a generic term"

// PublicEntry is one entry of the public banned-token family. Public entries
// render IN FULL: they are committed, reviewable config, and the pattern is the
// reviewable part (AC6).
type PublicEntry struct {
	ID       string `json:"id"`
	Pattern  string `json:"pattern"`
	Severity string `json:"severity"`
	// Key is the id with the managed prefix stripped, empty for an entry outside
	// the verb-owned namespace.
	Key string `json:"key"`
	// Managed marks the entries these verbs own and may remove.
	Managed bool `json:"managed"`
}

// PublicReport is the public layer's state.
type PublicReport struct {
	Path    string        `json:"path"`
	Present bool          `json:"present"`
	Entries []PublicEntry `json:"entries"`
}

// PublicResult is the outcome of a public-layer mutation.
type PublicResult struct {
	Path  string      `json:"path"`
	Entry PublicEntry `json:"entry"`
}

// AddPublicRequest adds one public entry. Severity defaults to blocker; Successor
// defaults to defaultSuccessor.
type AddPublicRequest struct {
	RepoRoot  string
	Key       string
	Pattern   string
	Severity  string
	Successor string
}

// Report is both layers at once, for the read-only render.
type Report struct {
	Private PrivateReport `json:"private"`
	Public  PublicReport  `json:"public"`
}

// List reports both layers. It is the read-only status render's single call.
func List(repoRoot string) (Report, error) {
	priv, err := ListPrivate(repoRoot)
	if err != nil {
		return Report{}, err
	}
	pub, err := ListPublic(repoRoot)
	if err != nil {
		return Report{}, err
	}
	return Report{Private: priv, Public: pub}, nil
}

// publicPath resolves the docs-lint config's absolute path under repoRoot.
func publicPath(repoRoot string) string {
	return filepath.Join(repoRoot, filepath.FromSlash(PublicConfigRelPath))
}

// readPublic returns the config's bytes, or ErrNoStore when it does not exist.
func readPublic(repoRoot string) ([]byte, error) {
	data, err := fsutil.ReadGuarded(publicPath(repoRoot), maxStoreBytes)
	switch {
	case err == nil:
		return data, nil
	case os.IsNotExist(err):
		return nil, fmt.Errorf("%w: %s", ErrNoStore, PublicConfigRelPath)
	default:
		return nil, fmt.Errorf("%w: %s is unreadable", ErrMalformedStore, PublicConfigRelPath)
	}
}

// ListPublic reports the whole banned_tokens family — the hand-curated entries as
// well as the verb-managed ones. There is one primitive, so a list that hid the
// hand-curated half would misrepresent what gates the repo.
func ListPublic(repoRoot string) (PublicReport, error) {
	rep := PublicReport{Path: PublicConfigRelPath, Entries: []PublicEntry{}}
	data, err := readPublic(repoRoot)
	switch {
	case err == nil:
	case errors.Is(err, ErrNoStore):
		return rep, nil
	default:
		return PublicReport{}, err
	}
	rep.Present = true
	span, err := locateBannedTokens(data)
	if err != nil {
		return PublicReport{}, err
	}
	for _, el := range span.elems {
		var tok lint.BannedToken
		if err := json.Unmarshal(el.raw, &tok); err != nil {
			return PublicReport{}, fmt.Errorf("%w: %s has an unreadable banned_tokens entry", ErrMalformedStore, PublicConfigRelPath)
		}
		rep.Entries = append(rep.Entries, publicEntry(tok))
	}
	return rep, nil
}

// publicEntry renders one loaded token as a reportable entry.
func publicEntry(tok lint.BannedToken) PublicEntry {
	e := PublicEntry{ID: tok.ID, Pattern: tok.Pattern, Severity: tok.Severity}
	if key, ok := strings.CutPrefix(tok.ID, PublicIDPrefix); ok {
		e.Key, e.Managed = key, true
	}
	return e
}

// AddPublic inserts one entry into the banned_tokens family. The edit is byte
// surgery on the located array — one inserted line — so unrelated keys, entry
// order, and the file's formatting survive byte-for-byte and a review diff shows
// only the new entry.
func AddPublic(req AddPublicRequest) (PublicResult, error) {
	if !validKey(req.Key) {
		return PublicResult{}, fmt.Errorf("%w: %q (want [A-Za-z0-9][A-Za-z0-9._/-]*)", ErrInvalidKey, req.Key)
	}
	stored := storedPublicPattern(req.Pattern)
	if !validPublicPattern(stored) {
		return PublicResult{}, fmt.Errorf("%w for key %q: empty, or not a usable regular expression", ErrInvalidPattern, req.Key)
	}
	severity := req.Severity
	if severity == "" {
		severity = SeverityBlocker
	}
	if severity != SeverityBlocker && severity != SeverityWarn {
		return PublicResult{}, fmt.Errorf("%w: %q (want %s|%s)", ErrInvalidSeverity, req.Severity, SeverityBlocker, SeverityWarn)
	}
	successor := strings.TrimSpace(req.Successor)
	if successor == "" {
		successor = defaultSuccessor
	}

	data, err := readPublic(req.RepoRoot)
	if err != nil {
		return PublicResult{}, err
	}
	span, err := locateBannedTokens(data)
	if err != nil {
		return PublicResult{}, err
	}
	id := PublicIDPrefix + req.Key
	for _, el := range span.elems {
		var tok lint.BannedToken
		if err := json.Unmarshal(el.raw, &tok); err == nil && tok.ID == id {
			return PublicResult{}, fmt.Errorf("%w: %q", ErrDuplicateKey, req.Key)
		}
	}

	entry, err := encodeEntry(id, stored, severity, successor)
	if err != nil {
		return PublicResult{}, err
	}

	var out bytes.Buffer
	if len(span.elems) == 0 {
		indent := lineIndent(data, span.openEnd) + "  "
		out.Write(data[:span.openEnd])
		out.WriteString("\n" + indent)
		out.Write(entry)
		out.WriteString("\n" + lineIndent(data, span.openEnd))
		out.Write(data[span.closeStart:])
	} else {
		last := span.elems[len(span.elems)-1]
		out.Write(data[:last.end])
		out.WriteString(",\n" + lineIndent(data, last.start))
		out.Write(entry)
		out.Write(data[last.end:])
	}

	if err := writePublic(req.RepoRoot, out.Bytes()); err != nil {
		return PublicResult{}, err
	}
	return PublicResult{
		Path:  PublicConfigRelPath,
		Entry: PublicEntry{ID: id, Pattern: stored, Severity: severity, Key: req.Key, Managed: true},
	}, nil
}

// storedPublicPattern is the exact string an entry carries on disk. Every one of
// the config's hand-curated entries opens with (?i), and the docs promise the
// public layer matches case-insensitively — but the linter compiles the stored
// pattern RAW, so an entry without the flag is case-SENSITIVE and a mixed-case
// spelling of the banned name walks through CI. The prefix is added here, once, so
// a verb-written entry is enforced exactly as a hand-written one is; a pattern that
// already carries it is left alone rather than double-flagged.
func storedPublicPattern(pattern string) string {
	if pattern == "" || strings.HasPrefix(pattern, "(?i)") {
		return pattern
	}
	return "(?i)" + pattern
}

// RemovePublic drops the verb-managed entry with the given key. Entries outside the
// managed namespace are refused: the hand-curated families are edited in the config
// by a human, in a reviewable commit, never by a verb.
func RemovePublic(repoRoot, key string) (PublicResult, error) {
	data, err := readPublic(repoRoot)
	if err != nil {
		return PublicResult{}, err
	}
	span, err := locateBannedTokens(data)
	if err != nil {
		return PublicResult{}, err
	}

	id := PublicIDPrefix + strings.TrimPrefix(key, PublicIDPrefix)
	target := -1
	var removed lint.BannedToken
	for i, el := range span.elems {
		var tok lint.BannedToken
		if err := json.Unmarshal(el.raw, &tok); err != nil {
			continue
		}
		switch {
		case tok.ID == id:
			target, removed = i, tok
		case tok.ID == key:
			// Named exactly, but outside the namespace these verbs own.
			return PublicResult{}, fmt.Errorf("%w: %q is hand-curated; edit %s to change it", ErrNotManaged, key, PublicConfigRelPath)
		}
	}
	if target < 0 {
		return PublicResult{}, fmt.Errorf("%w: %q", ErrUnknownKey, key)
	}

	var out bytes.Buffer
	switch {
	case len(span.elems) == 1:
		out.Write(data[:span.openEnd])
		out.Write(data[span.closeStart:])
	case target == 0:
		out.Write(data[:span.elems[0].start])
		out.Write(data[span.elems[1].start:])
	default:
		out.Write(data[:span.elems[target-1].end])
		out.Write(data[span.elems[target].end:])
	}

	if err := writePublic(repoRoot, out.Bytes()); err != nil {
		return PublicResult{}, err
	}
	return PublicResult{Path: PublicConfigRelPath, Entry: publicEntry(removed)}, nil
}

// writePublic validates the edited bytes and commits them atomically. Validation
// before the write is the point: the config gates CI, so an edit that would leave
// it unloadable must fail with the file untouched.
func writePublic(repoRoot string, data []byte) error {
	var probe lint.Config
	if err := json.Unmarshal(data, &probe); err != nil {
		return fmt.Errorf("%w: the edit would leave %s unparseable", ErrMalformedStore, PublicConfigRelPath)
	}
	return fsutil.WriteFileAtomicPreserveMode(publicPath(repoRoot), data)
}

// encodeEntry renders one banned-token entry as a single compact line, matching
// the field order the config's existing entries use.
func encodeEntry(id, pattern, severity, successor string) ([]byte, error) {
	// A local ordered shape rather than lint.BannedToken: the struct's field order
	// is the on-disk key order, and the entry must read like its neighbours.
	entry := struct {
		ID           string   `json:"id"`
		Pattern      string   `json:"pattern"`
		Severity     string   `json:"severity"`
		Successor    string   `json:"successor"`
		AllowContext []string `json:"allow_context"`
		Message      string   `json:"message"`
	}{
		ID: id, Pattern: pattern, Severity: severity, Successor: successor,
		AllowContext: []string{docsLintAllowEscape},
		Message: "names a banned token in user-facing content (" + id + "); this repo's published surface does not name it. " +
			"Use " + successor + ", or add <!-- docs-lint: allow --> if naming it is genuinely necessary.",
	}
	// Compact, and with HTML escaping off: the patterns and the allow-context
	// regexp carry <, > and & (the escape is an HTML comment), which <-style
	// escaping would render unreadable beside the hand-written entries.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(entry); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// elemSpan is one array element's byte range: start is its first byte, end the
// byte after its last.
type elemSpan struct {
	start, end int
	raw        json.RawMessage
}

// arraySpan locates the banned_tokens array: the byte after its '[', the byte of
// its ']', and every element's range.
type arraySpan struct {
	openEnd    int
	closeStart int
	elems      []elemSpan
}

// locateBannedTokens finds the top-level banned_tokens array by streaming the
// document with the standard decoder and reading its input offsets. Nothing is
// re-marshalled: the caller edits the original bytes inside the located ranges,
// which is what keeps an edit to one entry from churning the whole file.
func locateBannedTokens(data []byte) (arraySpan, error) {
	malformed := func(detail string) (arraySpan, error) {
		return arraySpan{}, fmt.Errorf("%w: %s %s", ErrMalformedStore, PublicConfigRelPath, detail)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	depth := 0
	for {
		tok, err := dec.Token()
		if err != nil {
			return malformed("has no top-level banned_tokens array")
		}
		if d, ok := tok.(json.Delim); ok {
			if d == '{' || d == '[' {
				depth++
			} else {
				depth--
			}
			continue
		}
		key, ok := tok.(string)
		if !ok || depth != 1 || key != "banned_tokens" {
			continue
		}
		open, err := dec.Token()
		if err != nil {
			return malformed("has an unreadable banned_tokens value")
		}
		if d, ok := open.(json.Delim); !ok || d != '[' {
			return malformed("has a non-array banned_tokens value")
		}
		span := arraySpan{openEnd: int(dec.InputOffset())}
		for dec.More() {
			before := int(dec.InputOffset())
			var raw json.RawMessage
			if err := dec.Decode(&raw); err != nil {
				return malformed("has an unreadable banned_tokens entry")
			}
			span.elems = append(span.elems, elemSpan{start: valueStart(data, before), end: int(dec.InputOffset()), raw: raw})
		}
		if _, err := dec.Token(); err != nil {
			return malformed("has an unterminated banned_tokens array")
		}
		span.closeStart = int(dec.InputOffset()) - 1
		return span, nil
	}
}

// valueStart advances past the separator whitespace and comma the decoder reports
// as part of an element's leading offset, landing on the value's first byte.
func valueStart(data []byte, at int) int {
	for at < len(data) {
		switch data[at] {
		case ' ', '\t', '\r', '\n', ',':
			at++
		default:
			return at
		}
	}
	return at
}

// lineIndent returns the leading whitespace of the line containing at, so an
// inserted entry lands in the same column as the entries around it.
func lineIndent(data []byte, at int) string {
	if at > len(data) {
		at = len(data)
	}
	start := bytes.LastIndexByte(data[:at], '\n') + 1
	indent := 0
	for start+indent < len(data) && (data[start+indent] == ' ' || data[start+indent] == '\t') {
		indent++
	}
	return string(data[start : start+indent])
}
