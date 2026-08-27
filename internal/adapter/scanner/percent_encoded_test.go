package scanner

import (
	"strings"
	"testing"
)

// Fake, non-live credential shapes. They match the bundled patterns' STRUCTURE
// (prefix + length) but are invented placeholders — never real credentials.
const (
	fakeGHP  = "ghp_" + "0123456789abcdefABCDEF0123456789abcd" // ghp_ + 36 alnum
	fakeAKIA = "AKIA" + "ABCDEFGHIJKLMNOP"                     // AKIA + 16 [0-9A-Z]
	fakeJWT  = "eyJ" + "hbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ" + "zdWIiOiIxMjM0NTY3ODkwIn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV0abcdefgh"
)

// TestPercentEncodedDelimiterDefeatsLeadingBoundary is the gh-370 regression
// guard: a live token embedded in a percent-encoded URL (the delimiter before
// the token encoded to a %XX whose final byte is a hex word char) must still be
// DETECTED and must NOT survive redaction into the stored bytes. On the unfixed
// scanner the leading \b lands on the hex byte and the token is missed entirely.
func TestPercentEncodedDelimiterDefeatsLeadingBoundary(t *testing.T) {
	cases := []struct {
		name  string
		line  string
		token string
	}{
		{
			// The canonical OAuth/redirect shape from the report:
			// redirect=%2Fauth%3Ftoken%3Dghp_...  ( = -> %3D, / -> %2F, ? -> %3F )
			name:  "ghp_in_encoded_redirect_url",
			line:  "GET /callback?redirect=%2Fauth%3Ftoken%3D" + fakeGHP + "&state=x HTTP/1.1",
			token: fakeGHP,
		},
		{
			name:  "aws_access_key_after_encoded_equals",
			line:  "curl 'https://api/x?key%3D" + fakeAKIA + "'",
			token: fakeAKIA,
		},
		{
			name:  "jwt_magic_link_after_encoded_equals",
			line:  "verify link%3Ftoken%3D" + fakeJWT + " sent",
			token: fakeJWT,
		},
		{
			// Double-encoded delimiter (%25 -> %, so %253D -> %3D -> =): the
			// bounded multi-pass decode must still reach the literal token.
			name:  "ghp_behind_double_encoded_equals",
			line:  "log redirect=%252Ftoken%253D" + fakeGHP + " end",
			token: fakeGHP,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			text := tc.line + "\n"
			findings := ScanText(text, Identity{}, DefaultPatterns(), DefaultIdentitySeverities(), "transcript")
			if len(findings) == 0 {
				t.Fatalf("no finding for a live token behind a percent-encoded delimiter:\n%s", tc.line)
			}
			redacted, n := Redact(text, findings)
			if n == 0 {
				t.Fatal("Redact rewrote nothing")
			}
			if strings.Contains(redacted, tc.token) {
				t.Errorf("live token survived redaction into the stored bytes:\n%s", redacted)
			}
			// Fail-closed: a re-scan of the stored bytes must find no residual.
			if residual := ScanText(redacted, Identity{}, DefaultPatterns(), DefaultIdentitySeverities(), "transcript"); len(residual) != 0 {
				t.Errorf("token still detectable after redaction: %+v", residual)
			}
		})
	}
}

// TestPlaintextTokensStillRedact guards against a regression in the ordinary
// (unencoded) path: a plaintext token=ghp_... must keep detecting and redacting.
func TestPlaintextTokensStillRedact(t *testing.T) {
	text := "export GITHUB_TOKEN=" + fakeGHP + "\n"
	findings := ScanText(text, Identity{}, DefaultPatterns(), DefaultIdentitySeverities(), "t")
	if len(findings) == 0 {
		t.Fatal("plaintext token no longer detected")
	}
	redacted, _ := Redact(text, findings)
	if strings.Contains(redacted, fakeGHP) {
		t.Errorf("plaintext token survived redaction:\n%s", redacted)
	}
}

// TestBenignPercentSequencesUnharmed proves the decode pre-pass does not invent
// findings: a percent-encoded URL carrying NO secret is left exactly as it was.
func TestBenignPercentSequencesUnharmed(t *testing.T) {
	text := "GET /search?q=hello%20world%3Ffoo%3Dbar&next=%2Fhome HTTP/1.1\n"
	findings := ScanText(text, Identity{}, DefaultPatterns(), DefaultIdentitySeverities(), "t")
	if len(findings) != 0 {
		t.Fatalf("benign percent sequences produced findings: %+v", findings)
	}
	redacted, n := Redact(text, findings)
	if n != 0 || redacted != text {
		t.Errorf("benign text was altered:\n%s", redacted)
	}
}
