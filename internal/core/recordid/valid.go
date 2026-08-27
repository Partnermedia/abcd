package recordid

import "regexp"

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
