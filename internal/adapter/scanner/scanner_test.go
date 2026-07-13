package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, root, rel, content string) string {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return abs
}

func hasKind(f []Finding, kind string) bool {
	for _, fn := range f {
		if fn.Kind == kind {
			return true
		}
	}
	return false
}

// scanLine is a small helper: scan one line with the default patterns and no
// identity.
func scanLine(line string) []Finding {
	return ScanText(line, Identity{}, DefaultPatterns(), DefaultIdentitySeverities(), "f")
}

// TestSecretPatterns is a table test over one positive + one negative per §2.2
// pattern.
func TestSecretPatterns(t *testing.T) {
	r := strings.Repeat
	cases := []struct {
		kind string
		pos  string
		neg  string
	}{
		{"token:github_pat", "ghp_" + r("a", 36), "ghp_short"},
		{"token:github_server", "ghs_" + r("b", 36), "ghs_short"},
		{"token:github_oauth", "gho_" + r("c", 36), "gho_short"},
		{"token:github_user", "ghu_" + r("d", 36), "ghu_short"},
		{"token:github_refresh", "ghr_" + r("e", 36), "ghr_short"},
		{"token:anthropic", "sk-ant-" + r("A", 40), "sk-ant-short"},
		{"token:openai_project", "sk-proj-" + r("B", 40), "sk-proj-short"},
		{"token:openai_svcacct", "sk-svcacct-" + r("C", 40), "sk-svcacct-short"},
		{"token:aws_access_key", "AKIA" + r("A", 16), "AKIA-not-a-key"},
		{"token:slack", "xoxb-" + r("1", 12), "xoxb-x"},
		{"token:google_api", "AIza" + r("Z", 35), "AIza-too-short"},
		{"token:stripe_live", "sk_live_" + r("9", 24), "sk_live_short"},
		{"token:stripe_test", "sk_test_" + r("8", 24), "sk_test_short"},
		{"token:jwt_shaped", "eyJ" + r("a", 12) + "." + r("b", 12) + "." + r("c", 12), "eyJ-not-jwt"},
		{"rp_session_key", `"sessionKey": "real-uuid-value-here"`, `"sessionKey": ""`},
	}
	for _, c := range cases {
		t.Run(c.kind, func(t *testing.T) {
			if got := scanLine(c.pos); !hasKind(got, c.kind) {
				t.Errorf("positive %q did not flag %s: %+v", c.pos, c.kind, got)
			}
			if got := scanLine(c.neg); hasKind(got, c.kind) {
				t.Errorf("negative %q wrongly flagged %s", c.neg, c.kind)
			}
		})
	}
}

// TestRE2LookaroundPorts proves each ported lookaround predicate.
func TestRE2LookaroundPorts(t *testing.T) {
	// AWS docs example is skipped; a real AKIA key is caught.
	if hasKind(scanLine("AKIAIOSFODNN7EXAMPLE"), "token:aws_access_key") {
		t.Error("AWS docs example must NOT be flagged")
	}
	if !hasKind(scanLine("AKIA1234567890ABCDEF"), "token:aws_access_key") {
		t.Error("a real AKIA key must be flagged")
	}
	// A redacted sessionKey value is not re-flagged; a real one is.
	if hasKind(scanLine(`"sessionKey": "<RP-SESSION-UUID-REDACTED>"`), "rp_session_key") {
		t.Error("redacted sessionKey must NOT be re-flagged")
	}
	if !hasKind(scanLine(`"sessionKey": "abc-123-def-456"`), "rp_session_key") {
		t.Error("a real sessionKey value must be flagged")
	}
}

// TestScanBundleReadsContent proves the scanner inspects file CONTENTS: a
// planted secret inside an included file is caught and counts as a hard-fail.
func TestScanBundleReadsContent(t *testing.T) {
	root := t.TempDir()
	secret := "ghp_" + strings.Repeat("z", 40)
	abs := writeFile(t, root, "commands/x.md", "# doc\ntoken = "+secret+"\n")

	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	res, err := sc.ScanBundle([]BundleFile{{LogicalPath: "commands/x.md", ResolvedPath: abs}})
	if err != nil {
		t.Fatal(err)
	}
	if res.FilesScanned != 1 {
		t.Fatalf("expected 1 file scanned, got %d", res.FilesScanned)
	}
	if res.HardFails == 0 {
		t.Fatalf("planted secret not caught: %+v", res)
	}
	if !hasKind(res.Findings, "token:github_pat") {
		t.Errorf("expected github_pat finding, got %+v", res.Findings)
	}
	if res.Findings[0].File != "commands/x.md" || res.Findings[0].Line != 2 {
		t.Errorf("finding not reported under logical path/line: %+v", res.Findings[0])
	}
}

// TestScanBundleSkipsBinary proves a binary file is not scanned.
func TestScanBundleSkipsBinary(t *testing.T) {
	root := t.TempDir()
	abs := writeFile(t, root, "assets/blob.bin", "prefix\x00"+"ghp_"+strings.Repeat("a", 40))
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	res, _ := sc.ScanBundle([]BundleFile{{LogicalPath: "assets/blob.bin", ResolvedPath: abs}})
	if res.FilesScanned != 0 || res.HardFails != 0 {
		t.Errorf("binary file must be skipped: %+v", res)
	}
}

// TestIdentityEmailAndNoreply proves identity email detection and the noreply
// suppression (pure ScanText path).
func TestIdentityEmailAndNoreply(t *testing.T) {
	id := Identity{GitUserEmail: "person@example.com"}
	pats := DefaultPatterns()
	sev := DefaultIdentitySeverities()
	if got := ScanText("contact person@example.com today", id, pats, sev, "f"); !hasKind(got, kindRealEmail) {
		t.Errorf("real email not flagged: %+v", got)
	}
	// The noreply form of the SAME email must be suppressed.
	id2 := Identity{GitUserEmail: "12345+person@users.noreply.github.com"}
	if got := ScanText("author 12345+person@users.noreply.github.com", id2, pats, sev, "f"); hasKind(got, kindRealEmail) {
		t.Errorf("noreply email must be suppressed: %+v", got)
	}
}

// TestIdentityHomePath proves home_path_self detection with the boundary
// predicate.
func TestIdentityHomePath(t *testing.T) {
	id := Identity{HomePath: "/Users/someone", HomeUser: "someone"}
	pats := DefaultPatterns()
	sev := DefaultIdentitySeverities()
	got := ScanText("see /Users/someone/notes.txt for details", id, pats, sev, "f") // abcd-audit:allow
	if !hasKind(got, kindHomeSelf) {
		t.Errorf("home_path_self not flagged: %+v", got)
	}
}

// TestSeverityFloorHeld proves a config override cannot lower a bundled pattern
// below its built-in floor, but may raise an identity kind.
func TestSeverityFloorHeld(t *testing.T) {
	root := t.TempDir()
	// Try to downgrade github_pat hard_fail → warn, and raise github_username
	// warn → hard_fail.
	cfg := `{
	  "patterns": { "github_pat": { "regex": "\\bghp_[A-Za-z0-9]{36,}\\b", "severity": "warn" } },
	  "identity_severities": { "github_username": "hard_fail", "real_email": "warn" }
	}`
	writeFile(t, root, ".abcd/config/pii.json", cfg)
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	var ghp Pattern
	for _, p := range sc.patterns {
		if p.Name == "github_pat" {
			ghp = p
		}
	}
	if ghp.Severity != SeverityHardFail {
		t.Errorf("github_pat downgrade to warn must be clamped to hard_fail, got %s", ghp.Severity)
	}
	if sc.identSev[kindGithubUser] != SeverityHardFail {
		t.Errorf("github_username raise to hard_fail must be honoured, got %s", sc.identSev[kindGithubUser])
	}
	if sc.identSev[kindRealEmail] != SeverityHardFail {
		t.Errorf("real_email downgrade to warn must be clamped to hard_fail, got %s", sc.identSev[kindRealEmail])
	}
}

// TestScannerFailClosed proves an unreadable per-repo config marks the scanner
// unavailable (fail-closed) and ScanBundle surfaces it.
func TestScannerFailClosed(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".abcd/config/pii.json", "{ this is not valid json ")
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	res, _ := sc.ScanBundle(nil)
	if !res.Unavailable || res.UnavailableReason == "" {
		t.Errorf("malformed config must mark scanner unavailable: %+v", res)
	}
}

// TestBlankSkipEntriesRejected (Finding 1a) proves a blank/slash-only skip
// fragment or an empty skip-extension entry — each a substring/suffix of every
// path — is rejected rather than silently zeroing the scan's coverage: the
// planted secret is still scanned and caught.
func TestBlankSkipEntriesRejected(t *testing.T) {
	cases := []struct {
		name    string
		cfg     string
		logical string
		file    string
	}{
		{"empty_fragment", `{"skip_path_fragments": [""]}`, "commands/x.md", "commands/x.md"},
		{"slash_fragment", `{"skip_path_fragments": ["/"]}`, "commands/x.md", "commands/x.md"},
		// An empty extension entry would skip every extensionless file.
		{"empty_extension", `{"skip_extensions": [""]}`, "commands/LICENSE", "commands/LICENSE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, root, ".abcd/config/pii.json", tc.cfg)
			secret := "ghp_" + strings.Repeat("z", 40)
			abs := writeFile(t, root, tc.file, "token = "+secret+"\n")
			sc, err := New(root)
			if err != nil {
				t.Fatal(err)
			}
			res, _ := sc.ScanBundle([]BundleFile{{LogicalPath: tc.logical, ResolvedPath: abs}})
			if res.FilesScanned != 1 {
				t.Fatalf("blank skip entry must not zero coverage: %+v", res)
			}
			if res.HardFails == 0 {
				t.Fatalf("secret must still be caught with a rejected skip entry: %+v", res)
			}
			if res.Unavailable {
				t.Errorf("coverage was preserved, must not be unavailable: %+v", res)
			}
		})
	}
}

// TestWhitespaceFragmentDropped (Finding 1a) proves a whitespace-only skip
// fragment is dropped from the merged skip list rather than carried (it would
// never match a real path but must not persist as config, per the finding).
func TestWhitespaceFragmentDropped(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".abcd/config/pii.json", `{"skip_path_fragments": ["   "]}`)
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, frag := range sc.skipFragments {
		if strings.TrimSpace(frag) == "" {
			t.Errorf("whitespace-only skip fragment must be dropped, found %q in %v", frag, sc.skipFragments)
		}
	}
}

// TestZeroCoverageSentinel (Finding 1b) proves a bundle with files but zero
// scanned — here every file skipped by a valid, non-empty extension skip — is
// marked Unavailable so the launch path fails closed instead of publishing an
// unscanned bundle.
func TestZeroCoverageSentinel(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".abcd/config/pii.json", `{"skip_extensions": [".md"]}`)
	abs := writeFile(t, root, "commands/x.md", "clean content\n")
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	res, _ := sc.ScanBundle([]BundleFile{{LogicalPath: "commands/x.md", ResolvedPath: abs}})
	if res.FilesScanned != 0 {
		t.Fatalf("expected all files skipped, got %+v", res)
	}
	if !res.Unavailable || res.UnavailableReason == "" {
		t.Errorf("zero coverage with files present must be marked unavailable: %+v", res)
	}
}

// TestSerializedFindingRedactsSecret (Finding 2) proves a planted PAT is still
// counted as a hard-fail with its file+line preserved, but the raw token never
// appears in the serialized (JSON) scan result.
func TestSerializedFindingRedactsSecret(t *testing.T) {
	root := t.TempDir()
	secret := "ghp_" + strings.Repeat("Q", 40)
	abs := writeFile(t, root, "commands/x.md", "token = "+secret+"\n")
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	res, _ := sc.ScanBundle([]BundleFile{{LogicalPath: "commands/x.md", ResolvedPath: abs}})
	if res.HardFails == 0 {
		t.Fatalf("planted PAT must still be caught: %+v", res)
	}
	blob, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	js := string(blob)
	if strings.Contains(js, secret) {
		t.Errorf("serialized scan result must not contain the raw token: %s", js)
	}
	if !strings.Contains(js, `"file":"commands/x.md"`) || !strings.Contains(js, `"line":1`) {
		t.Errorf("file+line locator must survive redaction: %s", js)
	}
}

// TestSerializedFindingRedactsStraddlingSecret (Finding 2, straddle hole) plants
// a ghp_ PAT on a >200-byte line so the token crosses the snippet byte cap. The
// finding must still count as a hard-fail with its file/line/column intact, yet
// the serialized JSON must carry NO raw run of the token — not the whole token,
// and not any >=6-char window of it (the prefix a truncate-then-replace snippet
// would have leaked at the cut).
func TestSerializedFindingRedactsStraddlingSecret(t *testing.T) {
	token := "ghp_" + strings.Repeat("A", 40) // 44 chars; matches the github_pat pattern
	pad := strings.Repeat("x", 185)
	line := pad + " " + token // token starts at byte 186; byte 200 falls inside it
	if len(pad)+1+14 <= 200 && len(pad)+1 >= 200 {
		t.Fatal("test setup: token must straddle the 200-byte cap")
	}

	findings := ScanText(line, Identity{}, DefaultPatterns(), DefaultIdentitySeverities(), "commands/x.md")

	var pat *Finding
	for i := range findings {
		if findings[i].Kind == "token:github_pat" {
			pat = &findings[i]
			break
		}
	}
	if pat == nil {
		t.Fatalf("straddling PAT must still be caught: %+v", findings)
	}
	if pat.Severity != SeverityHardFail {
		t.Errorf("PAT must remain a hard-fail: %+v", pat)
	}
	wantCol := strings.Index(line, token) + 1
	if pat.Line != 1 || pat.Column != wantCol {
		t.Errorf("locator must be intact: got line=%d column=%d want line=1 column=%d", pat.Line, pat.Column, wantCol)
	}

	blob, err := json.Marshal(findings)
	if err != nil {
		t.Fatal(err)
	}
	js := string(blob)
	if strings.Contains(js, token) {
		t.Errorf("serialized result must not contain the raw token: %s", js)
	}
	// No >=6-char window (including any prefix) of the raw token may survive.
	for i := 0; i+6 <= len(token); i++ {
		if w := token[i : i+6]; strings.Contains(js, w) {
			t.Errorf("serialized result leaks a raw token window %q: %s", w, js)
		}
	}
	if !strings.Contains(js, `"file":"commands/x.md"`) || !strings.Contains(js, `"line":1`) {
		t.Errorf("file+line locator must survive redaction: %s", js)
	}
}

// TestSerializedShortIdentityFingerprinted proves a SHORT identity match (an
// email) is masked to a non-reversible fingerprint in the JSON: neither the raw
// value nor a leading fragment of it survives. A first-3 + last-2 window would
// expose most of such a short value, so short matches are fully starred.
func TestSerializedShortIdentityFingerprinted(t *testing.T) {
	email := "me@x.co" // 7 runes — short enough that head+tail would reveal it
	id := Identity{GitUserEmail: email}
	findings := ScanText("contact "+email+" now", id, DefaultPatterns(), DefaultIdentitySeverities(), "f")
	if !hasKind(findings, kindRealEmail) {
		t.Fatalf("real email must be flagged: %+v", findings)
	}
	blob, err := json.Marshal(findings)
	if err != nil {
		t.Fatal(err)
	}
	js := string(blob)
	if strings.Contains(js, email) {
		t.Errorf("serialized result must not contain the raw email: %s", js)
	}
	if strings.Contains(js, "me@") {
		t.Errorf("serialized result leaks a reversible fragment of the email: %s", js)
	}
	if !strings.Contains(js, `"matched":"`+strings.Repeat("*", len([]rune(email)))+`"`) {
		t.Errorf("short match must be fully starred (non-reversible): %s", js)
	}
}

// TestUnscannedBinaryRecorded (Finding 3) proves a leading-NUL file, classified
// binary and not scanned, is recorded in Unscanned rather than silently dropped.
func TestUnscannedBinaryRecorded(t *testing.T) {
	root := t.TempDir()
	abs := writeFile(t, root, "assets/blob.bin", "\x00ghp_"+strings.Repeat("a", 40))
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	res, _ := sc.ScanBundle([]BundleFile{{LogicalPath: "assets/blob.bin", ResolvedPath: abs}})
	if res.FilesScanned != 0 {
		t.Fatalf("leading-NUL file must not be scanned as text: %+v", res)
	}
	found := false
	for _, p := range res.Unscanned {
		if p == "assets/blob.bin" {
			found = true
		}
	}
	if !found {
		t.Errorf("unscannable binary file must be recorded, not silently dropped: %+v", res.Unscanned)
	}
}

// TestSerializedFindingRedactsSiblingSecret (iss-65 C14/C17, the BLOCK) plants
// TWO distinct secrets on one line. Each finding stores the whole line; the
// serialized snippet of finding A must not leak finding B's raw token, and vice
// versa — masking only a finding's own token leaves every sibling secret verbatim.
func TestSerializedFindingRedactsSiblingSecret(t *testing.T) {
	a := "ghp_" + strings.Repeat("A", 40)
	b := "ghp_" + strings.Repeat("B", 40)
	line := "gh=" + a + " other=" + b // both on one line, minified-env style

	findings := ScanText(line, Identity{}, DefaultPatterns(), DefaultIdentitySeverities(), "commands/x.md")
	if len(findings) < 2 {
		t.Fatalf("both planted PATs must be caught: %+v", findings)
	}
	blob, err := json.Marshal(findings)
	if err != nil {
		t.Fatal(err)
	}
	js := string(blob)
	for _, tok := range []string{a, b} {
		if strings.Contains(js, tok) {
			t.Errorf("serialized result leaks a sibling secret verbatim %q: %s", tok, js)
		}
		// No >=6-char window of either raw token may survive in any snippet.
		for i := 0; i+6 <= len(tok); i++ {
			if w := tok[i : i+6]; strings.Contains(js, w) {
				t.Errorf("serialized result leaks a raw token window %q: %s", w, js)
			}
		}
	}
}

// TestSerializedFindingRedactsOverlappingSecret (iss-65, the overlap variant of
// the BLOCK) plants two secrets whose matches PARTIALLY OVERLAP on one line: a
// greedy sk-ant- key that runs into a following JWT, both detected. Substring
// masking (longest-first) destroys the shorter match's substring and leaves its
// non-overlapping tail raw; only byte-span masking closes it. No >=6-char raw
// window of either token may survive in the serialized snippet.
func TestSerializedFindingRedactsOverlappingSecret(t *testing.T) {
	key := "sk-ant-" + strings.Repeat("X", 34)
	jwt := "eyJABCDEFGHIJ.KLMNOPQRST.UVWXYZ0123456"
	line := key + "-" + jwt // the `-` lets the key run into the JWT head; both fire

	findings := ScanText(line, Identity{}, DefaultPatterns(), DefaultIdentitySeverities(), "commands/x.md")
	if len(findings) < 2 {
		t.Fatalf("both an anthropic key and a JWT must be caught on the overlapping line: %+v", findings)
	}
	blob, err := json.Marshal(findings)
	if err != nil {
		t.Fatal(err)
	}
	js := string(blob)
	// The JWT payload/signature segments are the sensitive part — assert no raw
	// >=6-char window of either token survives anywhere in the serialized result.
	for _, tok := range []string{key, jwt} {
		for i := 0; i+6 <= len(tok); i++ {
			if w := tok[i : i+6]; strings.Contains(js, w) {
				t.Errorf("serialized snippet leaks a raw token window %q from an overlapping match: %s", w, js)
			}
		}
	}
}

// TestIsTextMultibyteRuneAtSniffBoundary (iss-65 C15) proves a valid UTF-8 file
// whose multibyte rune straddles the 8192-byte sniff cap is NOT misclassified as
// binary. The cut lands mid-rune, so a naive utf8.Valid(chunk[:8192]) fails.
func TestIsTextMultibyteRuneAtSniffBoundary(t *testing.T) {
	// '€' is 3 bytes (E2 82 AC) starting at index 8190, so the 8192 cut splits it.
	data := []byte(strings.Repeat("a", 8190) + "€" + strings.Repeat("z", 100))
	if !isText(data) {
		t.Fatal("a valid UTF-8 file with a rune straddling the sniff cap must read as text, not binary")
	}
	// A genuinely invalid encoding (a lone continuation byte early on) is still binary.
	bad := append([]byte("head "), 0x82)
	bad = append(bad, []byte(" tail")...)
	if isText(bad) {
		t.Error("genuinely invalid UTF-8 must still read as binary")
	}
}

// TestUnscannedUnreadableRecorded (iss-65 C18) proves a bundle file that cannot
// be read is surfaced in Unscanned, with the same visibility as a binary-skipped
// file, rather than silently dropped.
func TestUnscannedUnreadableRecorded(t *testing.T) {
	root := t.TempDir()
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	// ResolvedPath points at a file that does not exist → os.ReadFile errors.
	res, _ := sc.ScanBundle([]BundleFile{{LogicalPath: "commands/gone.md", ResolvedPath: filepath.Join(root, "nope", "gone.md")}})
	found := false
	for _, p := range res.Unscanned {
		if p == "commands/gone.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("an unreadable bundle file must be recorded in Unscanned, not silently dropped: %+v", res.Unscanned)
	}
}

// TestIdentityLocalUsernameSystemPathSuppressed proves the iss-31 fix: a machine
// username that collides with a system directory (here "dev" vs "/dev/null") is
// not flagged when it is the top segment of an absolute system path, while a
// genuine username occurrence is still caught.
func TestIdentityLocalUsernameSystemPathSuppressed(t *testing.T) {
	id := Identity{HomeUser: "dev"}
	pats := DefaultPatterns()
	sev := DefaultIdentitySeverities()

	// System path — must NOT flag (the /dev/null false positive).
	if got := ScanText(`run something 2>/dev/null || true`, id, pats, sev, "scripts/x.sh"); hasKind(got, kindLocalUser) {
		t.Errorf("system path /dev/null wrongly flagged as local username: %+v", got)
	}
	// Genuine leaks are still flagged (no false negatives from the suppression).
	if got := ScanText(`backup written to /home/dev/data`, id, pats, sev, "f"); !hasKind(got, kindLocalUser) { // abcd-audit:allow
		t.Errorf("nested username /home/dev not flagged (false negative): %+v", got)
	}
	if got := ScanText(`last commit authored by dev`, id, pats, sev, "f"); !hasKind(got, kindLocalUser) {
		t.Errorf("bare username not flagged (false negative): %+v", got)
	}
}
