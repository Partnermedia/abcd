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
	var b strings.Builder
	from := 0
	for {
		i := strings.Index(text[from:], home)
		if i < 0 {
			break
		}
		at := from + i
		end := at + len(home)
		if homeStandsAsPath(text, at, end) {
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

// nameContinues reports whether the name ending at text[:end] goes on: a
// letter or digit continues it outright, and a '.', '_' or '-' continues it
// only when a letter or digit follows ("/root.old", "/root-cause"), so the
// sentence punctuation after "/root." or "/root," is a boundary and never an
// excuse to leave the home in place.
func nameContinues(text string, end int) bool {
	if end >= len(text) {
		return false
	}
	b := text[end]
	if isAlnumByte(b) {
		return true
	}
	if b == '.' || b == '_' || b == '-' {
		return end+1 < len(text) && isAlnumByte(text[end+1])
	}
	return false
}

func isAlnumByte(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}

// SurvivingCallerHome reports any absolute path in text that still reveals the
// caller's OWN home after SweepCallerHome: the $HOME literal standing as a path
// (defensive — the sweep removes exactly those, so this fires only if the two
// ever disagree), or a "/Users/<user>" / "/home/<user>" segment for the
// caller's local username (basename of $HOME), whatever punctuation follows
// it — only a name that goes on ("/Users/me" inside "/Users/metoo") is not a abcd-audit:allow
// match. It is a deterministic substring check with no dependency on the
// pattern heuristic.
// Returned findings carry only the kind (masked Matched), enough for a refusal
// to report without exposing raw material.
func SurvivingCallerHome(text, home string) []Finding {
	var out []Finding
	if home != "" && SweepCallerHome(text, home) != text {
		out = append(out, Finding{Kind: kindHomeSelf, Matched: "~"})
	}
	user := home
	if i := strings.LastIndex(home, "/"); i >= 0 {
		user = home[i+1:]
	}
	if user != "" {
		for _, prefix := range []string{"/Users/", "/home/"} {
			if containsUserSegment(text, prefix+user) {
				out = append(out, Finding{Kind: kindHomeSelf, Matched: "~"})
			}
		}
	}
	return out
}

// containsUserSegment reports whether needle ("/Users/<user>" or "/home/<user>")
// appears in text as a complete path segment, by the same rule the sweep
// anchors on (nameContinues): "/Users/me" does not falsely match "/Users/metoo" abcd-audit:allow
// (a different, longer username), while "/Users/me." at a sentence end does. abcd-audit:allow
func containsUserSegment(text, needle string) bool {
	from := 0
	for {
		i := strings.Index(text[from:], needle)
		if i < 0 {
			return false
		}
		if !nameContinues(text, from+i+len(needle)) {
			return true
		}
		from = from + i + 1
	}
}
