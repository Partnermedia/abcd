package banlist

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/REPPL/abcd-cli/internal/fsutil"
	"github.com/REPPL/abcd-cli/internal/gitutil"
)

// privateFormatDecl is the store's self-description, and it must be the file's
// FIRST line for the keyed format to apply. Without it every line is a whole-line
// pattern (the format the guard shipped with), and a reader may NEVER split one:
// the alternative — deciding per line whether its first field "looks like a key" —
// silently changes what an old store matches AND prints part of a pattern as a
// "key", which on this layer is the secret. One declaration, read once, settles it
// for the whole file.
const privateFormatDecl = "# abcd-banlist: keyed"

// privateHeader is written when a store is created from nothing. It documents the
// format and seeds NOTHING: an example value in a file whose whole purpose is to
// hold real private strings would be one careless edit from looking like an entry.
// (Scaffolding a documented stub with reserved-value examples is `ahoy`'s job.)
const privateHeader = privateFormatDecl + `
# abcd private banlist — LOCAL TO THIS MACHINE, never committed.
# The line above declares the KEYED format: every entry line below is
#     KEY<space-or-tab>PATTERN
#   KEY      a stable, non-sensitive handle ([A-Za-z0-9][A-Za-z0-9._/-]*). It is
#            the only part of an entry any output ever names.
#   PATTERN  a POSIX extended regular expression, matched case-insensitively by
#            grep. Inline flag groups and Perl escapes are NOT supported there:
#            no (?i) — matching is already case-insensitive — and no \d, \w, \b.
# Blank lines and lines starting with '#' are ignored; leading and trailing ASCII
# spaces and tabs are stripped. A line that does not parse is refused by number,
# never skipped. Machine identifiers — hostnames, IP addresses, CIDR prefixes, MAC
# addresses, device names — are ordinary entries. The committed pre-commit hook
# refuses any commit whose staged content matches, naming the key alone.
#
# Remove the first line and every line below becomes a whole-line pattern again.
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
	// Keyed reports the store's format: true when its first line declares the
	// keyed format, false for a legacy whole-line-pattern store. It changes what
	// every other line MEANS, so a diagnostic that hid it would be unreadable.
	Keyed bool `json:"keyed"`
	// Entries are the keys, in file order. A line that does not parse contributes
	// none: it has no key, and its text may be the secret.
	Entries []Entry `json:"entries"`
	// Malformed lists the 1-based line numbers the enforcement engine cannot use —
	// a line that does not parse in this store's format, or a pattern grep refuses.
	// The guard refuses every commit until they are fixed. Line numbers only: the
	// content of such a line is withheld like any other.
	Malformed []int `json:"malformed_lines"`
	// Inert lists the 1-based line numbers whose pattern the engine ACCEPTS but
	// reads differently from the language it was written in (a Perl-style escape,
	// an inline flag group). The guard does not refuse these — grep may read them
	// differently than written, which is the more dangerous failure and a different
	// remedy.
	Inert []int `json:"inert_lines"`
	// NotIgnored reports a present store at a path git does NOT ignore — the whole
	// layer rests on the file being untracked, so a not-ignored store is one
	// `git add -A` from committing the very patterns it exists to keep out of
	// history. The read path surfaces it (AddPrivate refuses outright); it is only
	// meaningful when the store is Present and inside a git repo.
	NotIgnored bool `json:"not_ignored"`
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
// An unparsed line carries its number and NOTHING else — deriving a key from a
// line the format does not accept would mean printing part of a pattern.
type rawEntry struct {
	key      string
	pattern  string
	line     int
	unparsed bool
}

// trimTrail strips a trailing run of ASCII space, tab, and carriage return, so a
// CRLF store and an aligned store read identically.
func trimTrail(s string) string {
	for len(s) > 0 {
		switch s[len(s)-1] {
		case ' ', '\t', '\r':
			s = s[:len(s)-1]
		default:
			return s
		}
	}
	return s
}

// trimLead strips a leading run of ASCII space and tab.
//
// ASCII ONLY, deliberately, here and in trimTrail: strings.TrimSpace would also
// strip U+00A0, U+000B and the rest of Unicode's space table, and bash's
// [[:space:]] strips a different set again. Any such difference makes one reader
// key a line the other cannot parse — a store that looks healthy in `abcd banlist`
// while the guard skips it, or the reverse. The two readers of this format share
// one rule: a separator is a space or a tab, nothing else.
func trimLead(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	return s
}

// parse reads the private banlist format. It is the Go half of a format with two
// readers — the other is the committed shell hook — and the shared fixture corpora
// under testdata/ are what hold the two in agreement.
//
// The FIRST line decides the whole file. Exactly privateFormatDecl (bar trailing
// blanks) means KEYED: every non-comment, non-blank line must parse as
// KEY<space-or-tab>PATTERN, and one that does not is returned unparsed — by number,
// with no key and no pattern. Anything else means LEGACY: every non-comment,
// non-blank line is one whole-line pattern under the synthetic key
// entry-<line-number>, and no line is ever split.
//
// Pattern usability is NOT decided here (see checkPattern): parsing is a cheap,
// deterministic, engine-free step that every caller needs, and only the diagnostic
// read pays for asking the engine.
func parse(data []byte) (entries []rawEntry, keyed bool) {
	lines := strings.Split(string(data), "\n")
	keyed = len(lines) > 0 && trimTrail(lines[0]) == privateFormatDecl
	for i, raw := range lines {
		line := trimLead(trimTrail(raw))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		e := rawEntry{line: i + 1}
		if !keyed {
			e.key, e.pattern = "entry-"+strconv.Itoa(i+1), line
			entries = append(entries, e)
			continue
		}
		cut := strings.IndexAny(line, " \t")
		if cut <= 0 {
			e.unparsed = true
			entries = append(entries, e)
			continue
		}
		first, rest := line[:cut], trimLead(line[cut:])
		if rest == "" || !validKey(first) {
			e.unparsed = true
			entries = append(entries, e)
			continue
		}
		e.key, e.pattern = first, rest
		entries = append(entries, e)
	}
	return entries, keyed
}

// readPrivate returns the store's bytes, or ErrNoStore when it does not exist. The
// read is contained: every path component resolves inside repoRoot, and a symlinked
// or non-regular leaf is refused rather than followed — a store swapped for a link
// is tampering, not a store.
func readPrivate(repoRoot string) ([]byte, error) {
	root, err := os.OpenRoot(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("%w: %s is unreadable", ErrMalformedStore, PrivateRelPath)
	}
	defer root.Close()
	data, err := fsutil.ReadGuardedInRoot(root, PrivateRelPath, maxStoreBytes)
	switch {
	case err == nil:
		return data, nil
	case os.IsNotExist(err):
		return nil, fmt.Errorf("%w: %s", ErrNoStore, PrivateRelPath)
	default:
		// The read failure's cause (symlinked, oversize, not a regular file, outside
		// the root) is named without echoing any content.
		return nil, fmt.Errorf("%w: %s is unreadable (not a regular file, oversize, or a symlink)", ErrMalformedStore, PrivateRelPath)
	}
}

// writePrivateStore commits the store's bytes atomically, CONTAINED inside
// repoRoot: rel-path resolution happens through os.Root, so a symlinked
// `.abcd/.work.local` — the shape that lands the private patterns outside the repo
// while every surface still reports the in-repo path — cannot redirect the write.
// The mode is 0600: this file holds the patterns whose literal text must not leave
// this machine.
func writePrivateStore(repoRoot string, data []byte) error {
	root, err := os.OpenRoot(repoRoot)
	if err != nil {
		return err
	}
	defer root.Close()
	return fsutil.WriteFileAtomicInRoot(root, PrivateRelPath, data, 0o600)
}

// mkdirLocalTier creates the local-ephemeral tier INSIDE repoRoot, mode 0700: the
// directory holding the private patterns is no more readable than they are, and
// every component resolves through os.Root so a symlink out of the tree is an
// error rather than a redirect.
func mkdirLocalTier(repoRoot string) error {
	root, err := os.OpenRoot(repoRoot)
	if err != nil {
		return err
	}
	defer root.Close()
	return root.MkdirAll(PrivateDirRelPath, 0o700)
}

// ListPrivate reports the private layer: its format, its keys, the lines the
// enforcement engine cannot use, the lines it accepts but reads differently, and
// whether this machine has opted in at all.
func ListPrivate(repoRoot string) (PrivateReport, error) {
	rep := PrivateReport{Path: PrivateRelPath, Entries: []Entry{}, Malformed: []int{}, Inert: []int{}}
	data, err := readPrivate(repoRoot)
	switch {
	case err == nil:
	case errors.Is(err, ErrNoStore):
		return rep, nil
	default:
		return PrivateReport{}, err
	}
	rep.Present = true
	// The layer's whole safety is that the store is untracked; a present store git
	// would track is a leak waiting for `git add -A`, and the read path must not
	// report it as healthy. requireIgnoredStore is the write-path enforcement; here
	// the same condition is reported rather than refused.
	if gitutil.InRepo(repoRoot) && !gitutil.IsIgnored(repoRoot, PrivateRelPath) {
		rep.NotIgnored = true
	}
	entries, keyed := parse(data)
	rep.Keyed = keyed
	for _, e := range entries {
		if e.unparsed {
			rep.Malformed = append(rep.Malformed, e.line)
			continue
		}
		rep.Entries = append(rep.Entries, Entry{Key: e.key, Line: e.line})
		// The diagnostic verb asks the ENFORCEMENT engine, not Go's: during an
		// incident its whole value is agreeing with the guard about which lines
		// work, and which way they fail.
		switch fault, _ := checkPattern(e.pattern); fault {
		case faultNone:
		case faultPerlish:
			rep.Inert = append(rep.Inert, e.line)
		default:
			rep.Malformed = append(rep.Malformed, e.line)
		}
	}
	return rep, nil
}

// AddPrivate appends one entry, creating the store (and the local tier directory)
// when absent. Every check happens before the write, and the load-modify-write runs
// under the store's lock so a concurrent add or an out-of-band refresh cannot drop
// an entry.
func AddPrivate(req AddPrivateRequest) (PrivateResult, error) {
	if !validKey(req.Key) {
		return PrivateResult{}, fmt.Errorf("%w: %q (want [A-Za-z0-9][A-Za-z0-9._/-]*)", ErrInvalidKey, req.Key)
	}
	if strings.IndexByte(req.Pattern, 0) >= 0 {
		// A NUL byte the two readers disagree on: bash reads to the NUL, the Go parser
		// keeps the whole string. Refuse before it can wedge one reader.
		return PrivateResult{}, fmt.Errorf("%w for key %q: contains a NUL byte, which the store's readers disagree on (the value is withheld)", ErrInvalidPattern, req.Key)
	}
	if fault, byEngine := checkPattern(req.Pattern); fault != faultNone {
		return PrivateResult{}, fmt.Errorf("%w for key %q: %s (the value is withheld)", ErrInvalidPattern, req.Key, patternRefusal(fault, byEngine))
	}
	// checkPattern validates the pattern in ISOLATION; it does not prove the COMPOSED
	// line `KEY<space>PATTERN` reads back as the same (key, pattern). A whitespace-only
	// pattern composes to `key  `, which re-reads as malformed and wedges every commit;
	// a leading/trailing-whitespace pattern is silently trimmed on re-read, so the
	// enforced pattern differs from the validated one. Parse the composed line back and
	// refuse (writing nothing) unless it round-trips exactly.
	if !composedLineRoundTrips(req.Key, req.Pattern) {
		return PrivateResult{}, fmt.Errorf("%w for key %q: the composed entry does not round-trip (a whitespace-only pattern, or one with leading/trailing whitespace the reader would trim; the value is withheld)", ErrInvalidPattern, req.Key)
	}
	if err := requireIgnoredStore(req.RepoRoot); err != nil {
		return PrivateResult{}, err
	}

	var res PrivateResult
	err := withPrivateLock(req.RepoRoot, func() error {
		data, err := readPrivate(req.RepoRoot)
		switch {
		case err == nil:
		case errors.Is(err, ErrNoStore):
			data = []byte(privateHeader)
		default:
			return err
		}

		entries, keyed := parse(data)
		if !keyed && len(entries) > 0 {
			return legacyStoreRefusal("add to")
		}
		for _, e := range entries {
			if e.key == req.Key {
				return fmt.Errorf("%w: %q", ErrDuplicateKey, req.Key)
			}
		}

		body := string(data)
		if !keyed {
			// An entryless legacy store has no whole-line pattern to reinterpret, so
			// declaring the format here is safe — and leaves the user's comments intact.
			body = privateFormatDecl + "\n" + body
		}
		if body != "" && !strings.HasSuffix(body, "\n") {
			body += "\n"
		}
		body += req.Key + " " + req.Pattern + "\n"

		if err := writePrivateStore(req.RepoRoot, []byte(body)); err != nil {
			return err
		}
		// countParsed, not len(entries): an unparsed line is reported by list under
		// Malformed, never as an entry, so the mutation's count must agree with what
		// list shows rather than counting lines the store cannot use.
		res = PrivateResult{Path: PrivateRelPath, Key: req.Key, Entries: countParsed(entries) + 1}
		return nil
	})
	if err != nil {
		return PrivateResult{}, err
	}
	return res, nil
}

// countParsed counts the entries a reader can actually use — the ones that are not
// unparsed. It is what a mutation's "N entries" count must report, so that count
// agrees with the keys `list` renders rather than including lines list files under
// Malformed.
func countParsed(entries []rawEntry) int {
	n := 0
	for _, e := range entries {
		if !e.unparsed {
			n++
		}
	}
	return n
}

// composedLineRoundTrips reports whether the line AddPrivate is about to write —
// `key + " " + pattern` — parses back to exactly (key, pattern). It runs the real
// keyed parser on the line alone (declaration prepended) so the check IS the reader:
// a pattern the writer would compose into an unparseable or silently-altered line is
// caught before any byte reaches the store.
func composedLineRoundTrips(key, pattern string) bool {
	entries, _ := parse([]byte(privateFormatDecl + "\n" + key + " " + pattern))
	for _, e := range entries {
		if e.unparsed {
			return false
		}
		if e.key == key && e.pattern == pattern {
			return true
		}
	}
	return false
}

// RemovePrivate drops the entry with the given key. The edit is a line deletion —
// every other byte of the store survives, so comments and alignment are preserved.
func RemovePrivate(repoRoot, key string) (PrivateResult, error) {
	var res PrivateResult
	err := withPrivateLock(repoRoot, func() error {
		data, err := readPrivate(repoRoot)
		if err != nil {
			return err
		}
		entries, keyed := parse(data)
		if !keyed && len(entries) > 0 {
			return legacyStoreRefusal("remove from")
		}
		// ALL lines carrying the key, not just the first: a hand-written store can hold
		// a key twice, and deleting only one leaves the second still blocking under a
		// key the report says is gone.
		targets := map[int]bool{}
		for _, e := range entries {
			if e.key == key {
				targets[e.line] = true
			}
		}
		if len(targets) == 0 {
			return fmt.Errorf("%w: %q", ErrUnknownKey, key)
		}

		lines := strings.Split(string(data), "\n")
		kept := make([]string, 0, len(lines))
		for i, line := range lines {
			if targets[i+1] {
				continue
			}
			kept = append(kept, line)
		}
		if err := writePrivateStore(repoRoot, []byte(strings.Join(kept, "\n"))); err != nil {
			return err
		}
		res = PrivateResult{Path: PrivateRelPath, Key: key, Entries: countParsed(entries) - len(targets)}
		return nil
	})
	if err != nil {
		return PrivateResult{}, err
	}
	return res, nil
}

// legacyStoreRefusal words the one migration a verb will not perform. Writing a
// KEY<space>PATTERN line into a legacy store would not merely add an entry: the
// next reader that sees a declaration reinterprets every OTHER line, so a store of
// whole-line patterns would silently start matching remainders. The user adds one
// line and keeps control of what their patterns mean.
func legacyStoreRefusal(verb string) error {
	return fmt.Errorf("%w: %s predates the keyed format, so every line in it is a whole-line pattern; "+
		"to %s it, add this as the file's FIRST line and give each existing line a key — %s",
		ErrLegacyStore, PrivateRelPath, verb, privateFormatDecl)
}

// requireIgnoredStore refuses to create or grow the private store at a path git
// would track. The whole layer rests on the file being untracked: a store that is
// not ignored is one `git add -A` from committing the very strings it exists to
// keep out of history, and the guard cannot catch it (the guard's own patterns are
// what the file holds). When git cannot answer — not a repo, git absent — the check
// is skipped: a verb cannot demand proof no one can supply.
func requireIgnoredStore(repoRoot string) error {
	if !gitutil.InRepo(repoRoot) {
		return nil
	}
	if gitutil.IsIgnored(repoRoot, PrivateRelPath) {
		return nil
	}
	return fmt.Errorf("%w: git does not ignore %s, so the patterns it holds would be one `git add -A` from tracked history; "+
		"add `%s` to .gitignore (the local-ephemeral tier) and re-run",
		ErrStoreNotIgnored, PrivateRelPath, PrivateDirRelPath+"/")
}

// withPrivateLock serialises a load-modify-write of the private store on the shared
// inter-process lock primitive, so a concurrent `abcd banlist add` and an
// out-of-band refresh cannot each write a body derived from the same stale read and
// silently drop one entry. The lock file lives beside the store, inside the
// gitignored local tier.
func withPrivateLock(repoRoot string, fn func() error) error {
	// The tier is created THROUGH the root, not with os.MkdirAll: a symlinked
	// `.abcd/.work.local` would otherwise be followed and the lock file — the first
	// thing this function creates — would land outside the repo before the contained
	// write ever got a chance to refuse.
	if err := mkdirLocalTier(repoRoot); err != nil {
		return err
	}
	dir := filepath.Join(repoRoot, filepath.FromSlash(PrivateDirRelPath))
	err := fsutil.WithFileLock(filepath.Join(dir, privateLockFilename), storeLockTimeout, fn)
	switch {
	case errors.Is(err, fsutil.ErrLockContention):
		return fmt.Errorf("%w: could not lock %s within %s", ErrStoreLocked, PrivateRelPath, storeLockTimeout)
	case errors.Is(err, fsutil.ErrLockPathUnsafe):
		return fmt.Errorf("%w: the lock path beside %s is not a regular file", ErrMalformedStore, PrivateRelPath)
	}
	return err
}
