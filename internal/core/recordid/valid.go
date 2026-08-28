package recordid

import (
	"regexp"
	"strings"
)

// The ONE well-formedness rule for the two path-building record ids. A record id
// becomes a filesystem path in the intent and spec stores, so the loaders refuse
// anything but `<family>-<digits>` before they will build one — and the
// record-lint gate has to refuse exactly the same set, or a lint-green record
// fail-closes a loader for everyone who pulls it (iss-2608270500198764,
// iss-2608270500207987). Both sides call these predicates rather than restating
// the pattern, so the gate and the runtime cannot drift apart.
var (
	intentIDRe = regexp.MustCompile(`^itd-[0-9]+$`)
	specIDRe   = regexp.MustCompile(`^spc-[0-9]+$`)
)

// ValidIntentID reports whether id is a well-formed intent id (itd-N).
func ValidIntentID(id string) bool { return intentIDRe.MatchString(id) }

// ValidSpecID reports whether id is a well-formed spec id (spc-N).
func ValidSpecID(id string) bool { return specIDRe.MatchString(id) }

// recordFilenameRe splits a record filename into its family prefix (group 1,
// with its hyphen; empty for the ADR store's bare numeric form), its id number
// (group 2), and its slug segment (group 3, empty when the name carries none).
// The slug alternative is the SAME kebab-case shape the issue schema's slug
// property is validated against, so a name whose slug segment could not have
// been written by a store does not split at all.
var recordFilenameRe = regexp.MustCompile(`^([a-z]+-)?([0-9]+)(?:-([a-z0-9]+(?:-[a-z0-9]+)*))?\.md$`)

// SplitRecordFilename splits a record filename of the given family into the id
// it claims and the slug segment that follows it. family is the prefix without
// its hyphen ("iss", "itd", "spc") or "" for the ADR store's zero-padded numeric
// names; a filename of any other family does not match, so a stray record in a
// store's directory is not read as one of that store's.
//
// It is the ONE splitter for that question, called by both the ledger reader
// (core/capture's filename ↔ frontmatter invariants) and the record-lint gate
// that judges the committed corpus — the same reason ValidIntentID and
// ValidSpecID live here rather than being restated on each side. Two hand-kept
// copies would let a record pass the gate and then fail the reader, which is the
// split this package exists to prevent.
//
// The slug it returns is compared EXACTLY, never as a prefix. Every store applies
// its length cap while deriving the slug — before that one value forks into the
// filename and the frontmatter — so a filename is never a truncated form of a
// longer field, and a prefix tolerance would license the drift the comparison is
// there to catch.
func SplitRecordFilename(family, name string) (id, slug string, ok bool) {
	m := recordFilenameRe.FindStringSubmatch(name)
	if m == nil {
		return "", "", false
	}
	if strings.TrimSuffix(m[1], "-") != family {
		return "", "", false
	}
	return m[1] + m[2], m[3], true
}
