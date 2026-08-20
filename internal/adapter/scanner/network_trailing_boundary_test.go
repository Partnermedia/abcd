package scanner

import (
	"strings"
	"testing"
)

// A hostname abutting a word character on its right must still be caught
// (iss-344): `_` is a word character, so a trailing `\b` silently drops the
// whole match on shapes like a snake_case filename — the dead corner the token
// patterns retired in the secret sweep and iss-307 retired for ipv4/mac. The
// hostnames are assembled at runtime per this corpus's discipline.
func TestNetworkHostnameSurvivesWordCharSuffix(t *testing.T) {
	cases := []struct {
		name string
		line string
	}{
		{"lan underscore suffix", host("printer", "local") + "_backup"},
		{"lan underscore suffix in filename", "see " + host("scans", "printer", "lan") + "_2026.log"},
		{"device underscore suffix", dash("alexs", "macbook") + "_2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := scanNetLine(t, tc.line); len(got) == 0 {
				t.Fatalf("hostname in %q survived the scan: no findings", tc.line)
			}
		})
	}
}

// The suppression the trailing `\b` used to provide incidentally — an alnum
// right after the suffix means the regex truncated a LONGER label
// ("printer.local2", "alexs-macbookpro" name no .local host and no macbook) —
// must be preserved by the positional skip, not the boundary.
func TestNetworkHostnameTruncatedLabelStaysQuiet(t *testing.T) {
	for _, line := range []string{
		host("printer", "local") + "2",
		dash("alexs", "macbook") + "pro",
	} {
		if got := scanNetLine(t, line); len(got) != 0 {
			t.Fatalf("truncated label %q flagged: %v", line, got)
		}
	}
}

func scanNetLine(t *testing.T, line string) []Finding {
	t.Helper()
	var hits []Finding
	for _, f := range scanNet(line) {
		if strings.Contains(f.Kind, "net:") {
			hits = append(hits, f)
		}
	}
	return hits
}
