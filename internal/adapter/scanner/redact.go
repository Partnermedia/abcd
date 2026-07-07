package scanner

import "strings"

// ScanText scans an in-memory string with THIS scanner's merged patterns,
// probed identity, and per-repo severity floors — the same detection
// ScanBundle runs on a file's content, but with no filesystem read. It is the
// entry point a write-time redactor (the history store) uses to find every
// secret/PII span in a transcript before it lands on disk. logicalName is the
// label stamped into each finding's File field.
//
// It intentionally exposes the merged config that the package-level ScanText
// cannot: a caller using the package-level function would bypass the
// .abcd/config/pii.json override that New folded in.
func (s *Scanner) ScanText(text, logicalName string) []Finding {
	return ScanText(text, s.identity, s.patterns, s.identSev, logicalName)
}

// Redact rewrites every finding's matched span out of text, returning the
// sanitised text and the count of spans actually changed. It is the shared
// write-time sanitiser: both the history transcript store and any other
// pre-disk redactor route through it so the masking discipline lives in ONE
// place (it reuses the same maskSecret fingerprint and the same per-line
// strings.ReplaceAll approach as Finding.MarshalJSON).
//
// Every occurrence of a finding's raw Matched token on its line is replaced —
// masked to a non-reversible fingerprint for secret tokens, and to a neutral
// placeholder for identity kinds (a self home path collapses to "~"). Because
// replacement is by exact substring, a length-changing rewrite (a home path to
// "~") is safe. Redact is only stage one; the caller MUST re-scan the result
// and fail closed if any hard_fail span survived.
func Redact(text string, findings []Finding) (string, int) {
	if len(findings) == 0 {
		return text, 0
	}
	lines := strings.Split(text, "\n")

	// Group by 1-based line, then apply longest-match-first so a token that is
	// a substring of a wider match (a username inside a home path) never
	// pre-empts the wider rewrite.
	byLine := map[int][]Finding{}
	for _, f := range findings {
		byLine[f.Line] = append(byLine[f.Line], f)
	}
	rewritten := 0
	for lineno, fs := range byLine {
		idx := lineno - 1
		if idx < 0 || idx >= len(lines) {
			continue
		}
		sortByMatchedLenDesc(fs)
		line := lines[idx]
		for _, f := range fs {
			if f.Matched == "" {
				continue
			}
			repl := redactionReplacement(f)
			next := strings.ReplaceAll(line, f.Matched, repl)
			if next != line {
				rewritten++
				line = next
			}
		}
		lines[idx] = line
	}
	return strings.Join(lines, "\n"), rewritten
}

// redactionReplacement maps a finding to the text that replaces its raw span.
// Identity kinds get readable placeholders (never the original value, so a
// re-scan cannot re-match); secret tokens get the maskSecret fingerprint.
func redactionReplacement(f Finding) string {
	switch f.Kind {
	case kindHomeSelf:
		return "~"
	case kindHomeOther:
		return "[redacted-path]"
	case kindRealEmail:
		return "[redacted-email]"
	case kindRealName:
		return "[redacted-name]"
	case kindGithubUser, kindLocalUser:
		return "[redacted-user]"
	default:
		return maskSecret(f.Matched)
	}
}

// sortByMatchedLenDesc orders findings by descending Matched byte length
// (stable, deterministic on ties via the existing sortFindings key would be
// overkill here — insertion is small and ties are handled by stability).
func sortByMatchedLenDesc(fs []Finding) {
	for i := 1; i < len(fs); i++ {
		for j := i; j > 0 && len(fs[j].Matched) > len(fs[j-1].Matched); j-- {
			fs[j], fs[j-1] = fs[j-1], fs[j]
		}
	}
}
