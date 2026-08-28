package scanner

import (
	"strings"
	"testing"
)

// testIdent is the caller-identity fixture for the percent-decode identity
// sweep: an invented home path, its login, and an invented email. None are live.
func testIdent() Identity {
	return Identity{
		HomePath:     "/home/alice",
		HomeUser:     "alice",
		GitUserEmail: "alice@example.com",
	}
}

// TestPercentEncodedIdentitySurvivesRedaction — iss-2608270720336165. The
// percent-decode pre-pass added for gh-370 re-ran only the TOKEN matchers on the
// decoded copy; the identity matchers (home path, $HOME backstop, email, name,
// username) ran on the RAW line alone. So a percent-encoded home path or email —
// `%2Fhome%2Falice`, `%61lice%40example.com` — never appears literally on the
// raw line, defeats every identity matcher, and survives redaction into a
// COMMITTED memory/intent/capture artifact. The identity matchers must now run
// over the decoded copy too, with each hit mapped back to its raw span so Redact
// masks the live bytes where they sit on disk.
func TestPercentEncodedIdentitySurvivesRedaction(t *testing.T) {
	cases := []struct {
		name  string
		line  string
		token string // the raw, encoded bytes that must not survive
		kind  string
	}{
		{
			name:  "encoded_home_path_self",
			line:  "next=%2Fhome%2Falice&state=x",
			token: "%2Fhome%2Falice",
			kind:  kindHomeSelf,
		},
		{
			name:  "encoded_real_email",
			line:  "contact=%61lice%40example.com&ok=1",
			token: "%61lice%40example.com",
			kind:  kindRealEmail,
		},
		{
			// Double-encoded home path: the bounded multi-pass decode must still
			// reach it. %252F -> %2F -> '/'.
			name:  "double_encoded_home_path",
			line:  "redir=%252Fhome%252Falice end",
			token: "%252Fhome%252Falice",
			kind:  kindHomeSelf,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			text := tc.line + "\n"
			findings := ScanText(text, testIdent(), DefaultPatterns(), DefaultIdentitySeverities(), "memory")
			if !hasKind(findings, tc.kind) {
				t.Fatalf("no %s finding for a percent-encoded identity leak:\n%s\ngot %+v", tc.kind, tc.line, findings)
			}
			redacted, n := Redact(text, findings)
			if n == 0 {
				t.Fatal("Redact rewrote nothing")
			}
			if strings.Contains(redacted, tc.token) {
				t.Errorf("encoded identity token survived redaction into the stored bytes:\n%s", redacted)
			}
			// Fail-closed: a re-scan of the stored bytes finds no residual leak.
			if residual := ScanText(redacted, testIdent(), DefaultPatterns(), DefaultIdentitySeverities(), "memory"); hasKind(residual, tc.kind) {
				t.Errorf("identity still detectable after redaction: %+v", residual)
			}
		})
	}
}

// TestUnencodedIdentityUnaffected guards the common path: a line with NO
// percent-encoding must detect and redact identity exactly as before, and a
// benign encoded line carrying no identity must invent nothing.
func TestUnencodedIdentityUnaffected(t *testing.T) {
	// Plaintext home + email still caught on the raw line.
	text := "path /home/alice mail alice@example.com\n"
	findings := ScanText(text, testIdent(), DefaultPatterns(), DefaultIdentitySeverities(), "memory")
	if !hasKind(findings, kindHomeSelf) || !hasKind(findings, kindRealEmail) {
		t.Fatalf("plaintext identity no longer detected: %+v", findings)
	}
	redacted, _ := Redact(text, findings)
	if strings.Contains(redacted, "/home/alice") || strings.Contains(redacted, "alice@example.com") {
		t.Errorf("plaintext identity survived redaction:\n%s", redacted)
	}

	// A benign encoded URL carrying no identity produces no identity finding.
	benign := "GET /s?q=hello%20world&next=%2Ftmp%2Fcache HTTP/1.1\n"
	bf := ScanText(benign, testIdent(), DefaultPatterns(), DefaultIdentitySeverities(), "memory")
	for _, f := range bf {
		if IsIdentityKind(f.Kind) {
			t.Errorf("benign encoded line invented an identity finding: %+v", f)
		}
	}
}
