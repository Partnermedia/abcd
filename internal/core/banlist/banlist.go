// Package banlist is abcd's two-layer banned-names store (itd-74, spc-20). It
// performs no I/O beyond reading and writing files under a caller-supplied repo
// root — no printing, no os.Exit — so front doors under internal/surface/* format
// its results.
//
// The two layers differ in visibility, and that difference is the whole design:
//
//   - The PRIVATE layer is a gitignored per-machine file under the local tier
//     (PrivateRelPath). Its entries are KEY + PATTERN, and the key is the only
//     part any surface ever renders: a name that must never appear in public
//     content cannot be written into public config to ban it there. The committed
//     .githooks/pre-commit guard is the enforcement point, so this package's job is
//     the store and the format — parse, add, remove, list — not the matching.
//   - The PUBLIC layer is the banned_tokens family of the committed docs-lint
//     config (PublicConfigRelPath), enforced deterministically in CI with a
//     per-line escape. There is exactly one banned-token primitive: the
//     hand-curated families already in that file and the entries these verbs write
//     are the same mechanism, and verb-written entries are namespaced under
//     PublicIDPrefix so ownership is legible.
//
// Neither layer's writer ever rewrites a whole file from a decoded structure: the
// private store is edited line-wise and the public one by byte surgery on the
// located array, so hand-written comments, ordering, and formatting survive an
// edit and a review diff shows only the entry that changed.
package banlist

import (
	"errors"
	"regexp"
)

// PrivateRelPath is the gitignored per-machine private banlist, repo-relative. It
// sits in the local-ephemeral tier, which the three-tier layout gitignores as a
// whole — the one placement where a private pattern is safe from a `git add -A`.
const PrivateRelPath = ".abcd/.work.local/private-names.txt"

// PublicConfigRelPath is the committed docs-lint config that holds the public
// layer, repo-relative.
const PublicConfigRelPath = ".abcd/docs-lint.json"

// PublicIDPrefix namespaces the banned_tokens entries these verbs own. Entries
// outside it (the hand-curated harness/* and present_tense/* families) are listed
// but never edited or removed by a verb: the prefix is the ownership boundary.
const PublicIDPrefix = "names/"

// maxStoreBytes caps every banlist read (trust boundary), following the guard
// registry's precedent.
const maxStoreBytes = 256 * 1024

// Severity values a public entry may carry.
const (
	SeverityBlocker = "blocker"
	SeverityWarn    = "warn"
)

// Sentinel errors. A message never contains a pattern value: on the private layer
// the pattern is the secret, and an error is output like any other.
var (
	// ErrNoStore reports that the layer's file does not exist. It is never
	// flattened into an empty success: "no store" and "no entries" are different
	// states, and conflating them makes silence look like protection.
	ErrNoStore = errors.New("banlist store is absent")
	// ErrInvalidKey rejects a key outside the portable key charset.
	ErrInvalidKey = errors.New("invalid banlist key")
	// ErrInvalidPattern rejects an empty or uncompilable pattern.
	ErrInvalidPattern = errors.New("invalid banlist pattern")
	// ErrInvalidSeverity rejects a severity outside blocker|warn.
	ErrInvalidSeverity = errors.New("invalid banlist severity")
	// ErrDuplicateKey rejects a key the layer already carries.
	ErrDuplicateKey = errors.New("banlist key already exists")
	// ErrUnknownKey reports a key the layer does not carry.
	ErrUnknownKey = errors.New("unknown banlist key")
	// ErrNotManaged reports an entry outside the verb-owned id namespace.
	ErrNotManaged = errors.New("banlist entry is not verb-managed")
	// ErrMalformedStore reports a store whose bytes cannot be read as this layer's
	// format at all (as opposed to one unusable line, which is reported by number).
	ErrMalformedStore = errors.New("banlist store is malformed")
)

// keyRe is the portable key charset, shared by both layers and by the shell hook's
// own parser. It deliberately excludes every regular-expression metacharacter:
// that is what lets the hook split "KEY<whitespace>PATTERN" safely, because the
// head of a broken regex can never pass for a key and so falls back to the legacy
// whole-line reading instead of silently changing what an old banlist matches.
var keyRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)

// validKey reports whether key is usable as a banlist key.
func validKey(key string) bool { return keyRe.MatchString(key) }

// validPattern checks a pattern compiles under the case-insensitive reading both
// layers use (the hook's `grep -iE`, the linter's own compile). The compile error
// is DISCARDED rather than wrapped: Go's regexp errors quote the expression, which
// would leak a private pattern into an error message.
func validPattern(pattern string) bool {
	if pattern == "" {
		return false
	}
	_, err := regexp.Compile("(?i)" + pattern)
	return err == nil
}
