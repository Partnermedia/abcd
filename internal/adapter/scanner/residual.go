package scanner

import (
	"os"
	"strings"
)

// residual.go — the stage-two discipline the committed stores share.
//
// Every store that writes free text into a committed artefact (the history
// transcript store, the memory page store) redacts through this package and
// then asks the same three questions before the write lands: what is the
// caller's home, did it survive, and which rescan findings must refuse the
// write. Each store used to carry its own copy of the answers, so the stores
// that are meant to agree on what counts as a leak could be fixed apart. The
// answers live here once; the stores call them.

// CallerHome resolves the caller's home directory as ProbeIdentity does — $HOME
// first (so tests and redirected runs agree), then os.UserHomeDir — trimmed of
// any trailing slash. Empty when neither resolves.
func CallerHome() string {
	home := os.Getenv("HOME")
	if home == "" {
		if h, err := os.UserHomeDir(); err == nil {
			home = h
		}
	}
	return strings.TrimRight(home, "/")
}

// BlockingResidual filters a stage-two rescan to the findings that must refuse
// a write. Severity alone is the wrong gate: the hostname patterns are shape
// heuristics and therefore warn by design, so a LAN host or device name that
// survived stage-one redaction would otherwise be committed in silence — the
// very class of leak the stores exist to stop. Any surviving IDENTITY or
// NETWORK span refuses the write whatever its severity; everything else still
// gates on hard_fail. After the stage-one detector fixes this path is rarely
// reachable, which is what a backstop is for.
func BlockingResidual(findings []Finding) []Finding {
	var out []Finding
	for _, f := range findings {
		if f.Severity == SeverityHardFail || IsIdentityKind(f.Kind) {
			out = append(out, f)
		}
	}
	return out
}

// SweepCallerHome is the deterministic literal $HOME backstop, independent of
// the pattern heuristic: every occurrence of home that stands as a path is
// collapsed to "~". An occurrence stands as a path when nothing path-like
// precedes it (the start of the text, or a byte that cannot continue a path
// segment) and nothing name-like follows it (the end, a separator, or any
// byte outside the username-continuation set). An unanchored replace turned
// "/rootfs/etc/hosts" into "~fs/etc/hosts" under HOME=/root and "/home/abc/x"
// into "~bc/x" under HOME=/home/a, silently corrupting the committed text; the
// anchor is what lets a short home coexist with the paths that merely share
// its prefix. An empty home sweeps nothing.
func SweepCallerHome(text, home string) string {
	if home == "" {
		return text
	}
	urls := urlSpans(text)
	var b strings.Builder
	from := 0
	for {
		i := strings.Index(text[from:], home)
		if i < 0 {
			break
		}
		at := from + i
		end := at + len(home)
		if homeSweepable(text, at, end, urls) {
			b.WriteString(text[from:at])
			b.WriteByte('~')
			from = end
			continue
		}
		b.WriteString(text[from : at+1])
		from = at + 1
	}
	if from == 0 {
		return text
	}
	b.WriteString(text[from:])
	return b.String()
}

// homeStandsAsPath reports whether text[at:end] — an occurrence of the home —
// is a path of its own rather than the prefix of another: the byte before it
// cannot continue a path segment (so "/var/root" and "~/root" are not "/root";
// a preceding '/' is an empty segment, as in "file:///root", and does not
// count) and the name does not continue past it (so "/rootfs" and
// "/root-cause" are not "/root" either, while "/root/x", "/root" at the end
// and "/root." before a space are).
func homeStandsAsPath(text string, at, end int) bool {
	if at > 0 && text[at-1] != '/' && (isPathSegmentByte(text[at-1]) || text[at-1] == '~') {
		return false
	}
	return !nameContinues(text, end)
}

// homeSweepable is homeStandsAsPath with the leading half of the anchor
// waived inside a URL span: behind a URL host the byte before the home is the
// host's last letter, which is a path-segment byte, yet the path IS the
// caller's home ("https://ci.example.com/Users/me/build.log"). The trailing abcd-audit:allow
// half still holds there, so a longer name behind a host is not swept either.
func homeSweepable(text string, at, end int, urls []span) bool {
	if inAnySpan(at, urls) {
		return !nameContinues(text, end)
	}
	return homeStandsAsPath(text, at, end)
}

// nameContinues reports whether the name ending at text[:end] goes on: a
// letter or digit continues it outright, and a '.' or '-' continues it only
// when a letter or digit follows ("/root.old", "/root-cause"), so the
// sentence punctuation after "/root." or "/root," is a boundary and never an
// excuse to leave the home in place. '_' is a boundary outright: it is a word
// rune to the local_username rule, so a home followed by '_' is one that rule
// would not catch either, and over-sweeping "/root_2" is the safe side of
// committing the home in "/root_backup/x".
func nameContinues(text string, end int) bool {
	if end >= len(text) {
		return false
	}
	b := text[end]
	if isAlnumByte(b) {
		return true
	}
	if b == '.' || b == '-' {
		return end+1 < len(text) && isAlnumByte(text[end+1])
	}
	return false
}

func isAlnumByte(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}

// SurvivingCallerHome is the deterministic backstop after SweepCallerHome,
// with no dependency on the pattern heuristic. It REWRITES what it can and
// reports what it cannot: a "/Users/<user>" / "/home/<user>" segment for the
// caller's local username (basename of $HOME) is rewritten to the
// local_username placeholder wherever the name does not go on — whatever
// punctuation follows it, and whatever precedes it, since "/Users/me" behind a abcd-audit:allow
// URL host or under a longer root is still the caller's name — and the $HOME
// literal standing as a path is reported (defensive: the sweep removes exactly
// those, so this fires only if the two ever disagree). A backstop that refused
// the write on a segment it could rewrite stopped every page that quoted the
// caller's own path; one that rewrites keeps the write and drops the name.
// Returned findings carry only the kind (masked Matched), enough for a
// refusal to report without exposing raw material.
func SurvivingCallerHome(text, home string) (string, []Finding) {
	var out []Finding
	if home == "" {
		return text, nil
	}
	user := home
	if i := strings.LastIndex(home, "/"); i >= 0 {
		user = home[i+1:]
	}
	if user != "" {
		for _, prefix := range []string{"/Users/", "/home/"} {
			text = sweepUserSegment(text, prefix+user, prefix+redactionReplacement(Finding{Kind: kindLocalUser}))
		}
	}
	if SweepCallerHome(text, home) != text {
		out = append(out, Finding{Kind: kindHomeSelf, Matched: "~"})
	}
	return text, out
}

// sweepUserSegment rewrites every occurrence of needle ("/Users/<user>" or
// "/home/<user>") that stands as a complete path segment, by the trailing
// half of the sweep's anchor (nameContinues): "/Users/me" does not falsely abcd-audit:allow
// match "/Users/metoo" (a different, longer username), while "/Users/me." at abcd-audit:allow
// a sentence end does. Only the trailing half — a leading anchor here would
// trade the old refusal for a leak, since a name behind a host or under a
// longer root is still the caller's name.
func sweepUserSegment(text, needle, repl string) string {
	var b strings.Builder
	from := 0
	for {
		i := strings.Index(text[from:], needle)
		if i < 0 {
			break
		}
		at := from + i
		end := at + len(needle)
		if !nameContinues(text, end) {
			b.WriteString(text[from:at])
			b.WriteString(repl)
			from = end
			continue
		}
		b.WriteString(text[from : at+1])
		from = at + 1
	}
	if from == 0 {
		return text
	}
	b.WriteString(text[from:])
	return b.String()
}
