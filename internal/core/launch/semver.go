package launch

import (
	"fmt"
	"regexp"
	"strconv"
)

// Semver is a parsed strict SemVer 2.0.0 version.
type Semver struct {
	Major, Minor, Patch int
	Prerelease, Build   string
}

// semverRe is the strict SemVer 2.0.0 grammar (semver.org): MAJOR.MINOR.PATCH
// with no leading zeros and no leading v, plus optional -prerelease and +build.
// It is anchored with \A...\z (not ^...$) so a trailing newline is rejected —
// the exact bug the Python semver_grammar module guards against.
var semverRe = func() *regexp.Regexp {
	numID := `0|[1-9]\d*`
	alnumID := `\d*[A-Za-z-][0-9A-Za-z-]*`
	preID := `(?:` + numID + `|` + alnumID + `)`
	buildID := `[0-9A-Za-z-]+`
	return regexp.MustCompile(
		`\A(` + numID + `)\.(` + numID + `)\.(` + numID + `)` +
			`(?:-(` + preID + `(?:\.` + preID + `)*))?` +
			`(?:\+(` + buildID + `(?:\.` + buildID + `)*))?\z`,
	)
}()

// IsStrictSemver reports whether value is a strict SemVer 2.0.0 string.
func IsStrictSemver(value string) bool {
	return semverRe.MatchString(value)
}

// ParseSemver parses a strict SemVer string into its components. It returns an
// error for any non-conforming input (the boundary validator).
func ParseSemver(value string) (Semver, error) {
	m := semverRe.FindStringSubmatch(value)
	if m == nil {
		return Semver{}, fmt.Errorf("not a strict SemVer 2.0.0 version (semver.org, no leading 'v'): %q", value)
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	return Semver{Major: major, Minor: minor, Patch: patch, Prerelease: m[4], Build: m[5]}, nil
}

// Line is the MAJOR.MINOR retention line a version belongs to.
func (s Semver) Line() string { return fmt.Sprintf("%d.%d", s.Major, s.Minor) }

// String renders the core MAJOR.MINOR.PATCH (retention compares core only).
func (s Semver) String() string { return fmt.Sprintf("%d.%d.%d", s.Major, s.Minor, s.Patch) }

// Tag renders the v-prefixed git tag for the version core.
func (s Semver) Tag() string { return "v" + s.String() }

// coreLess orders two versions by (Major, Minor, Patch); prerelease/build are
// ignored (releases are core versions).
func coreLess(a, b Semver) bool {
	if a.Major != b.Major {
		return a.Major < b.Major
	}
	if a.Minor != b.Minor {
		return a.Minor < b.Minor
	}
	return a.Patch < b.Patch
}

// coreGreater reports whether a is strictly newer than b by core version.
func coreGreater(a, b Semver) bool { return coreLess(b, a) }
