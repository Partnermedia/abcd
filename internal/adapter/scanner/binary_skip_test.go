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

// TestBinaryScanExemptsIdentityRules: prose/identity rules are meaningless on
// binary bytes, so the caller's login inside an image is not a finding — only
// the secret rules run there.
func TestBinaryScanExemptsIdentityRules(t *testing.T) {
	root := t.TempDir()
	sc, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	user := sc.identity.HomeUser
	if user == "" {
		t.Skip("no home user probed on this machine")
	}
	abs := writeFile(t, root, "a.png", "\x89PNG\r\n\x1a\n"+user+" "+user+"\n")
	res, _ := sc.ScanBundle([]BundleFile{{LogicalPath: "a.png", ResolvedPath: abs}})
	if len(res.Findings) != 0 {
		t.Fatalf("identity rules must not run on skip-listed binary bytes: %+v", res.Findings)
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
