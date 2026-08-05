package scanner

import (
	"strings"
	"testing"
	"time"
)

// TestConcatenatedSecretsBothDetected is the repro for iss-185: a
// leading-\b-anchored pattern can never match a second same-family token that
// immediately abuts a first with no separator, because the byte before the
// second token's start is itself a word character (the first token's last
// byte) — a word/word transition, so \b never holds there. ScanText must
// still catch both tokens, and Redact must star both, not leave the second
// one raw.
func TestConcatenatedSecretsBothDetected(t *testing.T) {
	r := strings.Repeat
	token1 := "github_pat_" + r("a", 22) + "_" + r("b", 59)
	token2 := "github_pat_" + r("c", 22) + "_" + r("d", 59)
	line := token1 + token2

	findings := scanLine(line)
	count := 0
	for _, f := range findings {
		if f.Kind == "token:github_pat_finegrained" {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected both concatenated tokens detected, got %d findings: %+v", count, findings)
	}

	redacted, _ := Redact(line, findings)
	if strings.Contains(redacted, token2) {
		t.Errorf("second concatenated token survived redaction raw: %q", redacted)
	}
	rescan := scanLine(redacted)
	for _, f := range rescan {
		if f.Severity == SeverityHardFail {
			t.Errorf("hard_fail survived redaction of concatenated secrets: %+v (out=%q)", f, redacted)
		}
	}
}

// TestConcatenatedDifferentFixedLengthSecretsBothDetected extends the iss-185
// repair to a MIXED pair: two different fixed-length patterns (a GitHub
// fine-grained PAT and an AWS access key) glued together with no separator.
// The original fix only probed a match's OWN pattern at the adjacency point,
// which missed this case exactly as it missed the same-family one — a
// pre-PR adversarial review caught it before merge.
func TestConcatenatedDifferentFixedLengthSecretsBothDetected(t *testing.T) {
	r := strings.Repeat
	patToken := "github_pat_" + r("a", 22) + "_" + r("b", 59)
	awsToken := "AKIA" + r("Q", 16)
	line := patToken + awsToken

	findings := scanLine(line)
	if !hasKind(findings, "token:github_pat_finegrained") {
		t.Errorf("github_pat_finegrained not detected in mixed concatenation: %+v", findings)
	}
	if !hasKind(findings, "token:aws_access_key") {
		t.Errorf("aws_access_key not detected in mixed concatenation: %+v", findings)
	}

	redacted, _ := Redact(line, findings)
	if strings.Contains(redacted, awsToken) {
		t.Errorf("AWS key survived redaction raw after a different-family concatenation: %q", redacted)
	}
	rescan := scanLine(redacted)
	for _, f := range rescan {
		if f.Severity == SeverityHardFail {
			t.Errorf("hard_fail survived redaction of mixed concatenated secrets: %+v (out=%q)", f, redacted)
		}
	}
}

// TestGoogleAPIKeyDashJunctionNotDoubleCounted is the regression guard for a
// bug the adjacency fix itself introduced: google_api_key is fixed-length
// but its charset includes '-', a NON-word character. When the 35th body
// char happens to be '-', the trailing \b already holds at the junction, so
// FindAllStringIndex already finds a second concatenated token on its own —
// and the naive adjacency probe used to append it a second time, double
// counting a single real finding. It must be reported exactly once.
func TestGoogleAPIKeyDashJunctionNotDoubleCounted(t *testing.T) {
	r := strings.Repeat
	token1 := "AIza" + r("A", 34) + "-" // 35th body char is '-': a real word boundary
	token2 := "AIza" + r("B", 35)
	line := token1 + token2

	findings := scanLine(line)
	count := 0
	for _, f := range findings {
		if f.Kind == "token:google_api" {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected exactly 2 google_api findings (no double count), got %d: %+v", count, findings)
	}
}

// TestAdjacencyProbeStaysLinearOnLongLines is the regression guard for a
// performance bug the adjacency fix's first pass introduced: an unanchored
// probe re-scans the ENTIRE remainder of the line for every candidate match
// end, discarding the result unless it happened to start at offset 0 — an
// O(matches × patterns × line length) blow-up on exactly the large
// single-line input (a minified asset, a base64 blob) a secret scanner must
// handle. Pre-PR adversarial security review measured this at 14+ seconds on
// a 200KB line; anchoring the probe (adjacencyProbe) restores it to a single
// O(1) attempt per candidate.
func TestAdjacencyProbeStaysLinearOnLongLines(t *testing.T) {
	line := strings.Repeat("10.0.0.1 ", 300) + strings.Repeat("x", 200000)
	start := time.Now()
	scanLine(line)
	if elapsed := time.Since(start); elapsed > 15*time.Second {
		t.Errorf("scan of a 200KB line took %v, want well under 15s (unanchored-probe regression)", elapsed)
	}
}

// TestAdjacencyProbeWindowIsBounded is the regression guard for a SECOND,
// distinct performance bug a merge-gate review found in the anchored probe:
// `\A` bounds the probe to one start position, but not the cost of that one
// attempt. net_lan_hostname and net_device_hostname carry their own
// unbounded internal quantifier (`[a-z0-9-]*`); run against a long
// terminator-free alnum run, a single anchored attempt scans to the end of
// that run before failing. Repeated at every match junction on a line with
// many back-to-back matches, this is O(matches x remaining line length) —
// review measured multiple seconds on a few thousand back-to-back
// fixed-length tokens. maxAdjacencyProbeWindow bounds every single probe
// attempt to a small fixed window regardless of what follows it.
func TestAdjacencyProbeWindowIsBounded(t *testing.T) {
	r := strings.Repeat
	cases := []struct {
		name string
		line string
	}{
		{"aws_keys_back_to_back", r("AKIA"+r("Q", 16), 2000)},
		{"google_keys_back_to_back", r("AIza"+r("Z", 35), 2000)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			start := time.Now()
			scanLine(c.line)
			if elapsed := time.Since(start); elapsed > 15*time.Second {
				t.Errorf("scan of %d back-to-back fixed-length tokens (%d bytes) took %v, want well under 15s (unbounded-probe-window regression)", 2000, len(c.line), elapsed)
			}
		})
	}
}
