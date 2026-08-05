package scanner

import (
	"strings"
	"testing"
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
