package banlist

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/REPPL/abcd-cli/internal/fsutil"
)

// privateHeader is written when a store is created from nothing. It documents the
// format and seeds NOTHING: an example value in a file whose whole purpose is to
// hold real private strings would be one careless edit from looking like an entry.
// (Scaffolding a documented stub with reserved-value examples is `ahoy`'s job.)
const privateHeader = `# abcd private banlist — LOCAL TO THIS MACHINE, never committed.
# One entry per line: KEY<whitespace>PATTERN
#   KEY      a stable, non-sensitive handle ([A-Za-z0-9][A-Za-z0-9._/-]*). It is
#            the only part of an entry any output ever names.
#   PATTERN  a POSIX extended regular expression, matched case-insensitively.
# Blank lines and lines starting with '#' are ignored. Machine identifiers —
# hostnames, IP addresses, CIDR prefixes, MAC addresses, device names — are
# ordinary entries. The committed pre-commit hook refuses any commit whose staged
# content matches, naming the key alone.
`

// Entry is one private banlist entry as any surface may see it: the key and the
// line it sits on. There is deliberately NO pattern field — the type is the
// redaction, so no future rendering can leak the value by accident (AC2/AC6).
type Entry struct {
	Key  string `json:"key"`
	Line int    `json:"line"`
}

// PrivateReport is the private layer's state.
type PrivateReport struct {
	// Path is the store's repo-relative location (reported whether or not it exists).
	Path string `json:"path"`
	// Present distinguishes "this machine has not opted in" from "opted in with no
	// entries" — the layer is local by construction, so its absence is a fact a
	// surface must be able to state plainly.
	Present bool `json:"present"`
	// Entries are the keys, in file order.
	Entries []Entry `json:"entries"`
	// Malformed lists the 1-based line numbers whose pattern the engine refuses.
	// Line numbers only: the content of such a line is withheld like any other.
	Malformed []int `json:"malformed_lines"`
}

// PrivateResult is the outcome of a private-layer mutation.
type PrivateResult struct {
	Path    string `json:"path"`
	Key     string `json:"key"`
	Entries int    `json:"entries"`
}

// AddPrivateRequest adds one private entry.
type AddPrivateRequest struct {
	RepoRoot string
	Key      string
	Pattern  string
}

// rawEntry is one parsed line. It never leaves the package: Pattern is the secret.
type rawEntry struct {
	key       string
	pattern   string
	line      int
	malformed bool
}

// parse reads the private banlist format. It is the Go half of a format with two
// readers — the other is the committed shell hook — and the shared fixture corpus
// under testdata/ is what holds the two in agreement.
//
// Blank lines and comment lines are skipped. A line whose first whitespace-
// delimited field is a valid key AND which has something after it is KEY+PATTERN;
// anything else is a bare pattern under the synthetic key entry-<line-number>, so
// a store in the older one-pattern-per-line format keeps meaning what it meant.
// Returned malformed line numbers are entries whose pattern does not compile;
// those entries are still listed (a bad line must not blind the store) but a
// caller must treat them as unusable rather than clean.
func parse(data []byte) (entries []rawEntry, malformed []int) {
	for i, raw := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		e := rawEntry{key: "entry-" + strconv.Itoa(i+1), pattern: line, line: i + 1}
		if cut := strings.IndexAny(line, " \t"); cut > 0 {
			first, rest := line[:cut], strings.TrimSpace(line[cut:])
			if rest != "" && validKey(first) {
				e.key, e.pattern = first, rest
			}
		}
		if !validPattern(e.pattern) {
			e.malformed = true
			malformed = append(malformed, e.line)
		}
		entries = append(entries, e)
	}
	return entries, malformed
}

// privatePath resolves the store's absolute path under repoRoot.
func privatePath(repoRoot string) string {
	return filepath.Join(repoRoot, filepath.FromSlash(PrivateRelPath))
}

// readPrivate returns the store's bytes, or ErrNoStore when it does not exist.
func readPrivate(repoRoot string) ([]byte, error) {
	data, err := fsutil.ReadGuarded(privatePath(repoRoot), maxStoreBytes)
	switch {
	case err == nil:
		return data, nil
	case os.IsNotExist(err):
		return nil, fmt.Errorf("%w: %s", ErrNoStore, PrivateRelPath)
	default:
		// The read failure's cause (symlinked, oversize, not a regular file) is
		// named without echoing any content.
		return nil, fmt.Errorf("%w: %s is unreadable", ErrMalformedStore, PrivateRelPath)
	}
}

// ListPrivate reports the private layer: its keys, its malformed lines, and
// whether this machine has opted in at all.
func ListPrivate(repoRoot string) (PrivateReport, error) {
	rep := PrivateReport{Path: PrivateRelPath, Entries: []Entry{}, Malformed: []int{}}
	data, err := readPrivate(repoRoot)
	switch {
	case err == nil:
	case errors.Is(err, ErrNoStore):
		return rep, nil
	default:
		return PrivateReport{}, err
	}
	rep.Present = true
	entries, malformed := parse(data)
	for _, e := range entries {
		rep.Entries = append(rep.Entries, Entry{Key: e.key, Line: e.line})
	}
	if malformed != nil {
		rep.Malformed = malformed
	}
	return rep, nil
}

// AddPrivate appends one entry, creating the store (and the local tier directory)
// when absent. The write is atomic and the store's mode is tightened to 0600: it
// holds the patterns whose literal text must not leave this machine.
func AddPrivate(req AddPrivateRequest) (PrivateResult, error) {
	if !validKey(req.Key) {
		return PrivateResult{}, fmt.Errorf("%w: %q (want [A-Za-z0-9][A-Za-z0-9._/-]*)", ErrInvalidKey, req.Key)
	}
	if !validPattern(req.Pattern) {
		return PrivateResult{}, fmt.Errorf("%w for key %q: empty, or not a usable regular expression (the value is withheld)", ErrInvalidPattern, req.Key)
	}

	data, err := readPrivate(req.RepoRoot)
	fresh := false
	switch {
	case err == nil:
	case errors.Is(err, ErrNoStore):
		fresh = true
		data = []byte(privateHeader)
	default:
		return PrivateResult{}, err
	}

	entries, _ := parse(data)
	for _, e := range entries {
		if e.key == req.Key {
			return PrivateResult{}, fmt.Errorf("%w: %q", ErrDuplicateKey, req.Key)
		}
	}

	body := string(data)
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	body += req.Key + " " + req.Pattern + "\n"

	if fresh {
		if err := os.MkdirAll(filepath.Dir(privatePath(req.RepoRoot)), 0o700); err != nil {
			return PrivateResult{}, err
		}
	}
	if err := fsutil.WriteFileAtomic(privatePath(req.RepoRoot), []byte(body), 0o600); err != nil {
		return PrivateResult{}, err
	}
	return PrivateResult{Path: PrivateRelPath, Key: req.Key, Entries: len(entries) + 1}, nil
}

// RemovePrivate drops the entry with the given key. The edit is a line deletion —
// every other byte of the store survives, so comments and alignment are preserved.
func RemovePrivate(repoRoot, key string) (PrivateResult, error) {
	data, err := readPrivate(repoRoot)
	if err != nil {
		return PrivateResult{}, err
	}
	entries, _ := parse(data)
	target := -1
	for _, e := range entries {
		if e.key == key {
			target = e.line
			break
		}
	}
	if target < 0 {
		return PrivateResult{}, fmt.Errorf("%w: %q", ErrUnknownKey, key)
	}

	lines := strings.Split(string(data), "\n")
	kept := make([]string, 0, len(lines))
	for i, line := range lines {
		if i+1 == target {
			continue
		}
		kept = append(kept, line)
	}
	if err := fsutil.WriteFileAtomic(privatePath(repoRoot), []byte(strings.Join(kept, "\n")), 0o600); err != nil {
		return PrivateResult{}, err
	}
	return PrivateResult{Path: PrivateRelPath, Key: key, Entries: len(entries) - 1}, nil
}
