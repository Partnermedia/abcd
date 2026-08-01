package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// bootstrapTag / bootstrapCommit are the fixture release the test server serves.
// The commit is a made-up 40-hex string: the script only ever tests its shape.
const (
	bootstrapTag    = "v9.9.9"
	bootstrapCommit = "0123456789abcdef0123456789abcdef01234567"
)

// bootstrapScript locates the committed hooks/bootstrap.sh from this test file's
// own on-disk position (internal/surface/cli/ -> three levels up), so the test
// drives the script that actually ships rather than a copy.
func bootstrapScript(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh unavailable")
	}
	if _, err := exec.LookPath("curl"); err != nil {
		t.Skip("curl unavailable")
	}
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "darwin/amd64", "darwin/arm64", "linux/amd64", "linux/arm64":
	default:
		t.Skipf("no released binary for %s/%s: the supported-matrix case covers this", runtime.GOOS, runtime.GOARCH)
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed to locate the test source file")
	}
	script := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "hooks", "bootstrap.sh"))
	if _, err := os.Stat(script); err != nil {
		t.Fatalf("the committed bootstrap script must exist: %v", err)
	}
	return script
}

// bootstrapAsset is the release asset name the script derives from uname.
func bootstrapAsset() string { return "abcd-" + runtime.GOOS + "-" + runtime.GOARCH }

// bootstrapServer stands in for the release host: the asset path redirects to a
// tag-stamped download URL (so the script can read the tag off curl's effective
// URL), checksums.txt carries whatever manifest the case wants, and the commits
// endpoint answers the release-sha lookup. hits counts every request, which is
// how the no-network cases are proved rather than asserted.
func bootstrapServer(t *testing.T, body []byte, manifest string) (base string, hits *int32) {
	t.Helper()
	var count int32
	asset := bootstrapAsset()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		switch r.URL.Path {
		case "/" + asset:
			http.Redirect(w, r, "/releases/download/"+bootstrapTag+"/"+asset, http.StatusFound)
		case "/releases/download/" + bootstrapTag + "/" + asset:
			_, _ = w.Write(body)
		case "/checksums.txt":
			_, _ = w.Write([]byte(manifest))
		case "/commits/" + bootstrapTag:
			_, _ = w.Write([]byte(`{"sha":"` + bootstrapCommit + `","node_id":"x"}`))
		default:
			http.NotFound(w, r)
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL, &count
}

// bootstrapManifest renders a sha256sum-style manifest over body, listing the
// platform asset plus one unrelated entry so a match is never accidental.
func bootstrapManifest(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]) + "  " + bootstrapAsset() + "\n" +
		strings.Repeat("0", 64) + "  abcd-other-platform\n"
}

// bootstrapRoot makes a plugin root whose basename is a 40-hex commit stamp —
// the shape the harness's plugin cache uses, and the input for plugin_sha.
func bootstrapRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), bootstrapCommit)
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

// runBootstrap executes the script with an explicitly constructed environment —
// no inherited proxy or plugin variables — and returns its merged output and
// exit code. extraPath is prepended to PATH so a case can shim a command.
func runBootstrap(t *testing.T, root, base, extraPath string) (string, int) {
	t.Helper()
	pathValue := os.Getenv("PATH")
	if extraPath != "" {
		pathValue = extraPath + string(os.PathListSeparator) + pathValue
	}
	cmd := exec.Command("sh", bootstrapScript(t))
	cmd.Env = []string{
		"PATH=" + pathValue,
		"HOME=" + t.TempDir(),
		"CLAUDE_PLUGIN_ROOT=" + root,
		"ABCD_BOOTSTRAP_BASE_URL=" + base,
		"ABCD_BOOTSTRAP_API_URL=" + base,
	}
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		var coded *exec.ExitError
		if e, ok := err.(*exec.ExitError); ok {
			coded = e
			code = coded.ExitCode()
		} else {
			t.Fatalf("running the bootstrap: %v (output %s)", err, out)
		}
	}
	return string(out), code
}

// metaValues parses the key=value .binary-meta the bootstrap writes.
func metaValues(t *testing.T, root string) map[string]string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".binary-meta"))
	if err != nil {
		t.Fatalf("the bootstrap must write .binary-meta on a successful install: %v", err)
	}
	out := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if k, v, ok := strings.Cut(line, "="); ok {
			out[k] = v
		}
	}
	return out
}

// TestBootstrapFastPathTouchesNoNetwork is the steady-state AC: a plugin root
// that already holds an executable binary costs one file test and nothing else.
func TestBootstrapFastPathTouchesNoNetwork(t *testing.T) {
	root := bootstrapRoot(t)
	base, hits := bootstrapServer(t, []byte("payload"), bootstrapManifest([]byte("payload")))
	binary := filepath.Join(root, "abcd")
	if err := os.WriteFile(binary, []byte("already here"), 0o755); err != nil {
		t.Fatal(err)
	}

	out, code := runBootstrap(t, root, base, "")
	if code != 0 {
		t.Errorf("a present binary must exit 0, got %d (output %q)", code, out)
	}
	if n := atomic.LoadInt32(hits); n != 0 {
		t.Errorf("a present binary must cost no network, got %d request(s)", n)
	}
	got, err := os.ReadFile(binary)
	if err != nil || string(got) != "already here" {
		t.Errorf("the fast path must leave the existing binary untouched; got %q (%v)", got, err)
	}
}

// TestBootstrapInstallsVerifiedBinary is the fresh-install AC: download, verify
// against the release manifest, install executable, and record the provenance
// the skew notice reads.
func TestBootstrapInstallsVerifiedBinary(t *testing.T) {
	root := bootstrapRoot(t)
	body := []byte("#!/bin/sh\nexit 0\n")
	base, _ := bootstrapServer(t, body, bootstrapManifest(body))

	out, code := runBootstrap(t, root, base, "")
	if code != 0 {
		t.Fatalf("a verified download must exit 0, got %d (output %q)", code, out)
	}
	binary := filepath.Join(root, "abcd")
	fi, err := os.Stat(binary)
	if err != nil {
		t.Fatalf("the bootstrap must install the binary into the plugin root: %v", err)
	}
	if fi.Mode().Perm() != 0o755 {
		t.Errorf("the installed binary must be mode 0755, got %v", fi.Mode().Perm())
	}
	got, err := os.ReadFile(binary)
	if err != nil || string(got) != string(body) {
		t.Errorf("the installed binary must be the verified download; got %q (%v)", got, err)
	}
	if !strings.Contains(out, "ahoy install") {
		t.Errorf("the success message must suggest `abcd ahoy install` for terminal PATH setup; output %q", out)
	}

	meta := metaValues(t, root)
	if meta["release_tag"] != bootstrapTag {
		t.Errorf("release_tag = %q, want %q", meta["release_tag"], bootstrapTag)
	}
	if meta["release_sha"] != bootstrapCommit {
		t.Errorf("release_sha = %q, want the commit the release API reports", meta["release_sha"])
	}
	if meta["plugin_sha"] != bootstrapCommit {
		t.Errorf("plugin_sha = %q, want the plugin root's commit stamp", meta["plugin_sha"])
	}
	if _, err := time.Parse(time.RFC3339, meta["fetched_at"]); err != nil {
		t.Errorf("fetched_at = %q, want an RFC3339 UTC timestamp (%v)", meta["fetched_at"], err)
	}
	if _, err := os.Stat(filepath.Join(root, ".bootstrap.lock")); !os.IsNotExist(err) {
		t.Error("the lock must be removed on the success path")
	}
}

// TestBootstrapRefusesChecksumMismatch is the trust bar: a download whose SHA-256
// disagrees with the release manifest never reaches the binary path, and the
// message says so.
func TestBootstrapRefusesChecksumMismatch(t *testing.T) {
	root := bootstrapRoot(t)
	base, _ := bootstrapServer(t, []byte("tampered"), bootstrapManifest([]byte("the published bytes")))

	out, code := runBootstrap(t, root, base, "")
	if code == 0 {
		t.Errorf("a checksum mismatch must fail loudly, got exit 0 (output %q)", out)
	}
	if !strings.Contains(out, "SHA-256") {
		t.Errorf("the refusal must name the checksum mismatch; output %q", out)
	}
	assertNothingInstalled(t, root, out)
}

// TestBootstrapRefusesAbsentManifestEntry is the other half of the same bar: a
// manifest that does not list this platform's asset is unverifiable, so nothing
// is installed.
func TestBootstrapRefusesAbsentManifestEntry(t *testing.T) {
	root := bootstrapRoot(t)
	body := []byte("payload")
	base, _ := bootstrapServer(t, body, strings.Repeat("0", 64)+"  abcd-other-platform\n")

	out, code := runBootstrap(t, root, base, "")
	if code == 0 {
		t.Errorf("an unlisted asset must fail loudly, got exit 0 (output %q)", out)
	}
	if !strings.Contains(out, bootstrapAsset()) {
		t.Errorf("the refusal must name the asset the manifest lacks; output %q", out)
	}
	assertNothingInstalled(t, root, out)
}

// assertNothingInstalled is the shared refusal contract: no binary, no meta file,
// no leftover lock or temp dir, and an actionable message rather than a raw
// shell error.
func assertNothingInstalled(t *testing.T, root, out string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, "abcd")); !os.IsNotExist(err) {
		t.Error("a refused download must never be installed")
	}
	if _, err := os.Stat(filepath.Join(root, ".binary-meta")); !os.IsNotExist(err) {
		t.Error("a refused download must write no .binary-meta")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".bootstrap") {
			t.Errorf("a refused run must leave no %s behind", e.Name())
		}
	}
	if strings.Contains(out, "No such file or directory") {
		t.Errorf("a raw shell error must never be the message a user gets; output %q", out)
	}
	if !strings.Contains(out, "go build ./cmd/abcd") {
		t.Errorf("the refusal must name the build-from-source recovery; output %q", out)
	}
}

// TestBootstrapUnsupportedPlatformChangesNothing shims uname so the script sees a
// platform no release covers. It must say so plainly, name the matrix, change
// nothing, and exit 0 — an unsupported platform is a reported condition, not a
// hook error to retry every session.
func TestBootstrapUnsupportedPlatformChangesNothing(t *testing.T) {
	root := bootstrapRoot(t)
	base, hits := bootstrapServer(t, []byte("payload"), bootstrapManifest([]byte("payload")))
	shim := t.TempDir()
	fake := "#!/bin/sh\ncase \"$1\" in\n-s) echo Haiku ;;\n-m) echo sparc64 ;;\n*) echo Haiku ;;\nesac\n"
	if err := os.WriteFile(filepath.Join(shim, "uname"), []byte(fake), 0o755); err != nil {
		t.Fatal(err)
	}

	out, code := runBootstrap(t, root, base, shim)
	if code != 0 {
		t.Errorf("an unsupported platform must exit 0, got %d (output %q)", code, out)
	}
	if n := atomic.LoadInt32(hits); n != 0 {
		t.Errorf("an unsupported platform must download nothing, got %d request(s)", n)
	}
	for _, want := range []string{"darwin", "linux", "amd64", "arm64", "go build ./cmd/abcd"} {
		if !strings.Contains(out, want) {
			t.Errorf("the message must name %q (the supported matrix and the way out); output %q", want, out)
		}
	}
	if entries, err := os.ReadDir(root); err != nil || len(entries) != 0 {
		t.Errorf("an unsupported platform must change nothing in the plugin root; entries %v (%v)", entries, err)
	}
}

// TestBootstrapLockContentionExitsQuietly: a second concurrent session that loses
// the mkdir race does nothing at all — no download, no output, no exit code that
// would render as a hook failure.
func TestBootstrapLockContentionExitsQuietly(t *testing.T) {
	root := bootstrapRoot(t)
	base, hits := bootstrapServer(t, []byte("payload"), bootstrapManifest([]byte("payload")))
	lock := filepath.Join(root, ".bootstrap.lock")
	if err := os.Mkdir(lock, 0o755); err != nil {
		t.Fatal(err)
	}

	out, code := runBootstrap(t, root, base, "")
	if code != 0 {
		t.Errorf("losing the lock race must exit 0, got %d (output %q)", code, out)
	}
	if out != "" {
		t.Errorf("losing the lock race must be silent; output %q", out)
	}
	if n := atomic.LoadInt32(hits); n != 0 {
		t.Errorf("losing the lock race must cost no network, got %d request(s)", n)
	}
	if _, err := os.Stat(lock); err != nil {
		t.Error("the loser must not remove the winner's live lock")
	}
}

// TestBootstrapStaleLockIsBroken: a lock left by a crashed run would otherwise
// make the plugin root unprovisionable forever, so one older than ten minutes is
// taken over.
func TestBootstrapStaleLockIsBroken(t *testing.T) {
	root := bootstrapRoot(t)
	body := []byte("payload")
	base, _ := bootstrapServer(t, body, bootstrapManifest(body))
	lock := filepath.Join(root, ".bootstrap.lock")
	if err := os.Mkdir(lock, 0o755); err != nil {
		t.Fatal(err)
	}
	stale := time.Now().Add(-30 * time.Minute)
	if err := os.Chtimes(lock, stale, stale); err != nil {
		t.Fatal(err)
	}

	out, code := runBootstrap(t, root, base, "")
	if code != 0 {
		t.Fatalf("a stale lock must be broken and the install proceed, got %d (output %q)", code, out)
	}
	if _, err := os.Stat(filepath.Join(root, "abcd")); err != nil {
		t.Errorf("a stale lock must not block provisioning: %v", err)
	}
}
