package scanner

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// GHSA-9wv7-88w3-f77m (iss-2608291807454357): a filename-keyed skip must never
// be an unverified allow. A file on the binary skip list still enters the
// payload with its bytes, so its bytes are scanned by the secret rules and the
// file is reported under ScannedBinary — never waved through by its name.

// fakeToken builds a secret-SHAPED value at runtime (iss-2608281752471145: no
// verbatim secret-shaped literal ever enters source).
func fakeToken() string { return "ghp_" + strings.Repeat("q", 40) }

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// TestSkipListedBinaryIsSecretScanned plants a token in notes.png — no image
// data at all, the bytes are the secret — and requires a hard-fail finding,
// the file reported as ScannedBinary, and nothing in Unscanned.
func TestSkipListedBinaryIsSecretScanned(t *testing.T) {
	root := t.TempDir()
	abs := writeFile(t, root, "docs/assets/notes.png", "token="+fakeToken()+"\n")
	clean := writeFile(t, root, "docs/README.md", "clean documentation\n")
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	res, _ := sc.ScanBundle([]BundleFile{
		{LogicalPath: "docs/README.md", ResolvedPath: clean},
		{LogicalPath: "docs/assets/notes.png", ResolvedPath: abs},
	})
	if res.HardFails == 0 {
		t.Fatalf("a secret inside a skip-listed .png must hard-fail: %+v", res)
	}
	if !contains(res.ScannedBinary, "docs/assets/notes.png") {
		t.Errorf("the .png must be reported as ScannedBinary: %+v", res)
	}
	if contains(res.Unscanned, "docs/assets/notes.png") {
		t.Errorf("a readable .png is verified, not an Unscanned gap: %+v", res)
	}
	if res.Unavailable {
		t.Errorf("a bundle with a fully scanned text file is not zero coverage: %+v", res)
	}
	if res.FilesScanned != 1 {
		t.Errorf("FilesScanned counts full-rule-set coverage only, want 1 got %d", res.FilesScanned)
	}
}

// TestSkipListedBinaryPEMKeyIsCaught covers the other secret class the skip
// exempted: a private-key block dropped into a .pdf.
func TestSkipListedBinaryPEMKeyIsCaught(t *testing.T) {
	root := t.TempDir()
	body := "%PDF-1.4\n-----BEGIN " + "RSA PRIVATE KEY-----\nMIIE\n"
	abs := writeFile(t, root, "docs/paper.pdf", body)
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	res, _ := sc.ScanBundle([]BundleFile{{LogicalPath: "docs/paper.pdf", ResolvedPath: abs}})
	if !hasKind(res.Findings, "token:pem_private_key") || res.HardFails == 0 {
		t.Fatalf("a private key inside a skip-listed .pdf must hard-fail: %+v", res)
	}
}

// pngBytes encodes a small valid RGBA image with the standard library.
func pngBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 16), uint8(y * 16), uint8(x ^ y), 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// TestValidPNGWithoutSecretIsClean: a genuine image yields zero findings, and
// is reported as ScannedBinary so the report says what was verified.
func TestValidPNGWithoutSecretIsClean(t *testing.T) {
	root := t.TempDir()
	abs := writeFile(t, root, "docs/assets/img/logo.png", string(pngBytes(t)))
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	res, _ := sc.ScanBundle([]BundleFile{{LogicalPath: "docs/assets/img/logo.png", ResolvedPath: abs}})
	if len(res.Findings) != 0 {
		t.Fatalf("a real image must not trip any rule: %+v", res.Findings)
	}
	if !contains(res.ScannedBinary, "docs/assets/img/logo.png") {
		t.Errorf("the image must be reported as ScannedBinary: %+v", res)
	}
}

// synthIdentity is a fabricated caller identity for the binary-branch tests —
// never the real account (iss-2608291444328326 shows why the literal login is
// the wrong fixture).
func synthIdentity() Identity {
	return Identity{
		GitUserName:       "Zed Q Eight",
		GitUserEmail:      "zq8@example.test",
		GitRemoteUsername: "zq8handle",
		HomePath:          "/Users/zq8home",
		HomeUser:          "zq8home",
	}
}

// TestBinaryScanKeepsLongIdentityRulesDropsShortOnes pins the rule for the
// skip-listed branch: the LONG-literal identity rules (home_path_self,
// real_email) run on raw bytes because their chance collision is negligible
// and a home path or an address in image or PDF metadata is the same leak it
// is in prose; the SHORT/GENERIC ones (local_username, real_name,
// github_username, home_path_other) do not, because on binary bytes they are
// noise.
func TestBinaryScanKeepsLongIdentityRulesDropsShortOnes(t *testing.T) {
	root := t.TempDir()
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	sc.identity = synthIdentity()
	id := sc.identity
	body := "\x89PNG\r\n\x1a\ntEXt " + id.HomePath + "/deck " + id.GitUserEmail + " " +
		id.GitUserName + " " + id.GitRemoteUsername + " " + id.HomeUser + " /home/someone/x\n"
	abs := writeFile(t, root, "a.png", body)
	res, _ := sc.ScanBundle([]BundleFile{{LogicalPath: "a.png", ResolvedPath: abs}})
	for _, kind := range []string{kindHomeSelf, kindRealEmail} {
		if !hasKind(res.Findings, kind) {
			t.Errorf("long-literal identity rule %s must fire on binary bytes: %+v", kind, res.Findings)
		}
	}
	for _, kind := range []string{kindLocalUser, kindRealName, kindGithubUser, kindHomeOther} {
		if hasKind(res.Findings, kind) {
			t.Errorf("short/generic identity rule %s must not fire on binary bytes: %+v", kind, res.Findings)
		}
	}
}

// TestBinaryHomePathHardFailsLikeText: renaming deck.md to deck.pdf must not
// turn a release-blocking home path in plaintext metadata into a pass.
func TestBinaryHomePathHardFailsLikeText(t *testing.T) {
	root := t.TempDir()
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	sc.identity = synthIdentity()
	body := "%PDF-1.4\n/Creator (" + sc.identity.HomePath + "/deck.key)\n"
	md := writeFile(t, root, "deck.md", body)
	pdf := writeFile(t, root, "deck.pdf", body)
	text, _ := sc.ScanBundle([]BundleFile{{LogicalPath: "deck.md", ResolvedPath: md}})
	bin, _ := sc.ScanBundle([]BundleFile{{LogicalPath: "deck.pdf", ResolvedPath: pdf}})
	if text.HardFails == 0 {
		t.Fatalf("control: the home path must hard-fail as text: %+v", text)
	}
	if bin.HardFails != text.HardFails {
		t.Errorf("deck.pdf hardfails=%d, deck.md hardfails=%d — the extension must not change the verdict", bin.HardFails, text.HardFails)
	}
}

// TestGenericHomePathInBinaryIsNotAFinding: home_path_other runs outside the
// identity guards, so it must be dropped explicitly on the binary branch.
func TestGenericHomePathInBinaryIsNotAFinding(t *testing.T) {
	root := t.TempDir()
	abs := writeFile(t, root, "b.png", "\x89PNG\r\n\x1a\n/home/runner/work/repo/x\n")
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	res, _ := sc.ScanBundle([]BundleFile{{LogicalPath: "b.png", ResolvedPath: abs}})
	if len(res.Findings) != 0 {
		t.Fatalf("a third-party path in binary bytes is noise, not a finding: %+v", res.Findings)
	}
}

// TestContainerIsNotVerified: a compressed container is byte-scanned (an
// uncompressed tar still yields its secret) but reported as
// ContainerUnverified, never as ScannedBinary — the deflate stream makes the
// byte scan vacuous, and the report must not claim otherwise
// (iss-2608291832160371).
func TestContainerIsNotVerified(t *testing.T) {
	root := t.TempDir()
	plainTar := writeFile(t, root, "docs/plain.tar", ".env\x00\x00TOKEN="+fakeToken()+"\n")
	gz := writeFile(t, root, "docs/pack.tgz", "\x1f\x8b\x08\x00opaque deflate bytes\n")
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	res, _ := sc.ScanBundle([]BundleFile{
		{LogicalPath: "docs/plain.tar", ResolvedPath: plainTar},
		{LogicalPath: "docs/pack.tgz", ResolvedPath: gz},
	})
	for _, p := range []string{"docs/plain.tar", "docs/pack.tgz"} {
		if !contains(res.ContainerUnverified, p) {
			t.Errorf("%s must be reported as ContainerUnverified: %+v", p, res)
		}
		if contains(res.ScannedBinary, p) {
			t.Errorf("%s must not claim ScannedBinary: %+v", p, res)
		}
	}
	if res.HardFails == 0 {
		t.Errorf("an uncompressed tar's plaintext secret is still caught by the byte scan: %+v", res)
	}
}

// TestBinaryScanCapMatchesPrivacyLint pins the cap to the sibling privacy
// scan's 4 MiB: ScanText's percent-decode map costs ~20-30x the input, so a
// larger cap is a memory cliff, not a convenience.
func TestBinaryScanCapMatchesPrivacyLint(t *testing.T) {
	if maxBinaryScanBytes != 4<<20 {
		t.Fatalf("maxBinaryScanBytes = %d, want 4 MiB (rule_privacy.go maxScanBytes)", maxBinaryScanBytes)
	}
}

// TestOversizedBinaryIsLoudRefusal: a skip-listed file past the byte cap is
// not read (memory bound) and lands in Unscanned — the fail-closed gap the
// launch gate refuses — never a quiet skip.
func TestOversizedBinaryIsLoudRefusal(t *testing.T) {
	root := t.TempDir()
	abs := filepath.Join(root, "big.zip")
	f, err := os.Create(abs)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(maxBinaryScanBytes + 1); err != nil { // sparse: no real bytes written
		t.Fatal(err)
	}
	f.Close()
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	res, _ := sc.ScanBundle([]BundleFile{{LogicalPath: "big.zip", ResolvedPath: abs}})
	if !contains(res.Unscanned, "big.zip") {
		t.Fatalf("an oversized skip-listed file must be an Unscanned refusal: %+v", res)
	}
	if contains(res.ScannedBinary, "big.zip") {
		t.Errorf("an unread file must not claim to be scanned: %+v", res)
	}
}
