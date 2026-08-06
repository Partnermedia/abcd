package scanner

import (
	"runtime/debug"
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

// TestConcatenatedOpenEndedSecretsBothDetected is the repro for iss-188, the
// gap iss-185's fix left open. github_pat's quantifier is open-ended
// (`[A-Za-z0-9]{36,}`), so on two abutting ghp_ tokens the first match greedily
// swallows the second token's own leading "ghp" — every byte of it is in the
// class — and stops only at the second token's '_'. The reported match end is
// therefore PAST the true junction, so probing at that end finds nothing and
// the second token's `_bbb…` tail used to survive Redact completely raw while
// the fail-closed residual re-scan reported the output clean.
func TestConcatenatedOpenEndedSecretsBothDetected(t *testing.T) {
	r := strings.Repeat
	token1 := "ghp_" + r("a", 36)
	token2 := "ghp_" + r("b", 36)
	line := token1 + token2

	findings := scanLine(line)
	if !hasKind(findings, "token:github_pat") {
		t.Fatalf("github_pat not detected at all: %+v", findings)
	}
	count := 0
	for _, f := range findings {
		if f.Kind == "token:github_pat" {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected both concatenated open-ended tokens detected, got %d findings: %+v", count, findings)
	}

	redacted, _ := Redact(line, findings)
	if strings.Contains(redacted, r("b", 36)) {
		t.Errorf("second concatenated token's tail survived redaction raw: %q", redacted)
	}
	rescan := scanLine(redacted)
	for _, f := range rescan {
		if f.Severity == SeverityHardFail {
			t.Errorf("hard_fail survived redaction of concatenated open-ended secrets: %+v (out=%q)", f, redacted)
		}
	}
}

// TestOpenEndedSecretSwallowingDifferentFamilyBothDetected is the mixed-family
// half of iss-188: an open-ended pattern's greedy class run can swallow a
// following token of a DIFFERENT family whole, when every byte of that token is
// in the first pattern's class (`AKIA`+16 upper-case is pure alnum). The
// junction is then 20 bytes before the reported end rather than at it.
func TestOpenEndedSecretSwallowingDifferentFamilyBothDetected(t *testing.T) {
	r := strings.Repeat
	ghpToken := "ghp_" + r("a", 36)
	awsToken := "AKIA" + r("Q", 16)
	line := ghpToken + awsToken

	findings := scanLine(line)
	if !hasKind(findings, "token:github_pat") {
		t.Errorf("github_pat not detected in mixed concatenation: %+v", findings)
	}
	if !hasKind(findings, "token:aws_access_key") {
		t.Errorf("aws_access_key swallowed by the open-ended match, not detected: %+v", findings)
	}

	redacted, _ := Redact(line, findings)
	if strings.Contains(redacted, awsToken) {
		t.Errorf("AWS key survived redaction raw after an open-ended swallow: %q", redacted)
	}
	rescan := scanLine(redacted)
	for _, f := range rescan {
		if f.Severity == SeverityHardFail {
			t.Errorf("hard_fail survived redaction of an open-ended swallow: %+v (out=%q)", f, redacted)
		}
	}
}

// TestStolenJunctionSearchDoesNotSkipPastRejectedCandidate is the repro for the
// gap iss-188's first fix left open, found independently by two adversarial
// reviews. junctionProbe is UNANCHORED, so one of its hits can SPAN the true
// junction — begin before it and end after it — without beginning AT it. The
// backward search used to resume at such a hit's END even when the hit's own
// offset had just been REJECTED by wholeMatch, stepping over every byte between
// the two, the real junction among them. Nothing revisits that range, so the
// second token was never recovered.
//
// Here the first `ghp_` match greedily runs to the second token's '_' at byte
// 47. The leftmost junction-probe hit inside it is google_api's body at byte 36
// (`AIza` + 35 more class bytes), which spans to byte 75 and is rejected — the
// prefix `ghp_…AIza` is not a whole github_pat match. The true junction is at
// byte 44, inside that span.
func TestStolenJunctionSearchDoesNotSkipPastRejectedCandidate(t *testing.T) {
	r := strings.Repeat
	token2 := "ghp_" + r("c", 36)
	line := "ghp_" + r("a", 32) + "AIza" + r("b", 4) + token2

	findings := scanLine(line)
	count := 0
	for _, f := range findings {
		if f.Kind == "token:github_pat" {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected both open-ended tokens detected across a rejected junction candidate, got %d findings: %+v", count, findings)
	}

	redacted, _ := Redact(line, findings)
	if strings.Contains(redacted, r("c", 36)) {
		t.Errorf("second token's tail survived redaction raw: %q", redacted)
	}
	rescan := scanLine(redacted)
	for _, f := range rescan {
		if f.Severity == SeverityHardFail {
			t.Errorf("hard_fail survived redaction past a rejected junction candidate: %+v (out=%q)", f, redacted)
		}
	}
}

// TestStolenJunctionSearchSkipsPastValidDecoySecret is the same defect reached
// without any filler: the rejected junction-probe hit that used to be skipped
// past is ITSELF a syntactically valid secret of a third family. The first
// `ghp_` match swallows a Google API key AND the head of an OpenAI project key;
// the leftmost candidate inside it is the Google key's own start at byte 14,
// whose span reaches byte 53 and is rejected (the prefix is shorter than
// github_pat's `{36,}` minimum). The real junction — the `sk-proj-` key at byte
// 40 — sits inside that span, so the whole 48-byte key used to survive Redact
// raw and reappear as a hard_fail on the fail-closed residual re-scan.
func TestStolenJunctionSearchSkipsPastValidDecoySecret(t *testing.T) {
	r := strings.Repeat
	openaiToken := "sk-proj-" + r("i", 40)
	line := "ghp_" + r("a", 10) + "AIza" + r("z", 22) + openaiToken

	findings := scanLine(line)
	if !hasKind(findings, "token:github_pat") {
		t.Errorf("github_pat not detected: %+v", findings)
	}
	if !hasKind(findings, "token:openai_project") {
		t.Errorf("openai project key behind a rejected decoy candidate not detected: %+v", findings)
	}

	redacted, _ := Redact(line, findings)
	if strings.Contains(redacted, r("i", 40)) {
		t.Errorf("openai project key survived redaction raw behind a decoy candidate: %q", redacted)
	}
	rescan := scanLine(redacted)
	for _, f := range rescan {
		if f.Severity == SeverityHardFail {
			t.Errorf("hard_fail survived redaction behind a decoy candidate: %+v (out=%q)", f, redacted)
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
	line := strings.Repeat("10.0.0.1 ", 300) + strings.Repeat("x", 200000) // abcd-audit:allow — adversarial perf fixture; the quad is stress input, not an identifier
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

// TestJunctionBacktrackIsBounded is the cost guard on iss-188's backward
// search. Recovering a junction a greedy quantifier ran past means looking
// BEHIND a match's reported end, and an open-ended pattern's match is
// legitimately allowed to be the whole line — a base64 blob, a minified asset.
// A per-byte backward walk over an arbitrarily long match, or a prefix
// re-validation per byte of it, would make exactly that input a
// resource-exhaustion cliff, trading one security bug for another. The search
// is instead capped at maxAdjacencyBacktrack behind the end regardless of how
// long the match is, and each candidate cut comes from one linear pass rather
// than a per-byte probe. Every case here is a single line of tens to hundreds
// of kilobytes, the shape that used to time out.
func TestJunctionBacktrackIsBounded(t *testing.T) {
	r := strings.Repeat
	n := scaleAdversarial
	cases := []struct {
		name string
		line string
	}{
		// One open-ended match spanning the whole line: the backward window
		// must not scale with it.
		{"one_very_long_open_ended_match", "ghp_" + r("a", n(200000))},
		{"one_very_long_jwt", "eyJ" + r("a", 20) + "." + r("b", 20) + "." + r("c", n(100000))},
		// Many open-ended matches, each of which backtracks.
		{"open_ended_tokens_back_to_back", r("ghp_"+r("a", 36), n(1000))},
		{"stripe_tokens_back_to_back", r("sk_live_"+r("a", 20), n(1000))},
		// One huge match densely seeded with candidate junctions, so the
		// backward search finds work at nearly every offset it looks at. This
		// is the case an unbounded backtrack blows up on: it measured 40s
		// against 0.3s bounded.
		{"dense_candidate_junctions", "ghp_" + r("AKIA", n(50000))},
		{"open_ended_chain_inside_open_ended", "xoxb-" + r("ghp_"+r("a", 36), n(1500))},
		// Dense short matches whose pattern is itself shrinkable.
		{"dense_dotted_quads", r("1.2.3.4", n(10000))}, // abcd-audit:allow — single-digit octets maximise match density; a reserved quad would weaken the stress
		// The worst case for resuming one byte past a REJECTED candidate rather
		// than past its whole span (see stolenJunctions): every match's backtrack
		// window is packed with junction-probe hits that ALL fail validation, so
		// the loop runs its maximum number of iterations and each one re-validates
		// a long prefix. Each token here is a JWT whose backtrack window falls
		// inside its middle segment, where no prefix can be a whole jwt_shaped
		// match (only one of the two required '.' separators is present), and the
		// segment is filled with `AKI` — the densest junction-probe hit spacing
		// the bundled set admits inside an alnum run, ~one candidate every three
		// bytes. Cost per match stays capped by the window, so the whole scan
		// stays linear in line length: doubling the line doubles the time, it does
		// not square it (measured 2.3s / 4.6s / 9.2s at 247KB / 494KB / 989KB).
		// The 15s bar is deliberately loose — it separates the unbounded-search
		// cliff (tens of seconds on inputs this size) from bounded work (~1s
		// here) without tracking machine speed.
		{"dense_rejected_candidates_in_backtrack_window", r("eyJ"+r("a", 10)+"."+r("AKI", 400)+"."+r("c", 20)+" ", n(100))},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			start := time.Now()
			scanLine(c.line)
			if elapsed := time.Since(start); elapsed > 15*time.Second {
				t.Errorf("scan of a %d-byte line took %v, want well under 15s (unbounded-backtrack regression)", len(c.line), elapsed)
			}
		})
	}
}

// raceDetector reports whether this test binary was built with -race. It is read
// from the build settings rather than a build tag so the whole guard stays in
// one file.
var raceDetector = func() bool {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return false
	}
	for _, s := range bi.Settings {
		if s.Key == "-race" {
			return s.Value == "true"
		}
	}
	return false
}()

// scaleAdversarial shrinks the adversarial cost-guard inputs when the race
// detector is on. The detector instruments every memory access the regexp
// engine makes and costs this package well over an order of magnitude in wall
// clock, which a fixed second-count budget cannot absorb. Shrinking the input
// is the right lever rather than loosening the budget: what each case tests is
// its SHAPE — one huge match, or many, densely seeded with candidate junctions
// — and every case's cost is linear in the input, so the shape survives a
// smaller multiplier. A budget stretched far enough to cover an instrumented
// 200KB line would stop discriminating a bounded search from an unbounded one
// on the uninstrumented run, which is where the guard has to bite.
func scaleAdversarial(n int) int {
	if !raceDetector {
		return n
	}
	if n < 8 {
		return 1
	}
	return n / 8
}
