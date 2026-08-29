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

// SurvivingCallerHome reports any absolute path in text that still reveals the
// caller's OWN home after the literal $HOME sweep: the $HOME literal itself
// (defensive — the sweep should have removed it), or a "/Users/<user>" /
// "/home/<user>" segment for the caller's local username (basename of $HOME),
// regardless of the character that follows it (trailing punctuation must never
// excuse a leak). It is a deterministic substring check with no dependency on
// the pattern heuristic. Returned findings carry only the kind (masked
// Matched), enough for a refusal to report without exposing raw material.
func SurvivingCallerHome(text, home string) []Finding {
	var out []Finding
	if home != "" && strings.Contains(text, home) {
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
// appears in text as a complete path segment: the rune following it must not be
// a username-continuation rune ([A-Za-z0-9._-]), so "/Users/me" does not falsely abcd-audit:allow
// match "/Users/metoo" (a different, longer username). abcd-audit:allow
func containsUserSegment(text, needle string) bool {
	from := 0
	for {
		i := strings.Index(text[from:], needle)
		if i < 0 {
			return false
		}
		end := from + i + len(needle)
		if end >= len(text) || !isPathUserByte(text[end]) {
			return true
		}
		from = from + i + 1
	}
}

func isPathUserByte(b byte) bool {
	return b == '.' || b == '_' || b == '-' ||
		(b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}
