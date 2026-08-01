package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// bootstrapTag / bootstrapCommit / bootstrapRelease are the fixture release the
// test server serves. The commits are made-up 40-hex strings (the script only
// ever tests their shape) and deliberately DIFFER from each other, so a script
// that confused the plugin root's stamp with the release's commit — the two
// values the skew notice compares — could not pass by coincidence.
const (
	bootstrapTag     = "v9.9.9"
	bootstrapCommit  = "0123456789abcdef0123456789abcdef01234567"
	bootstrapRelease = "fedcba9876543210fedcba9876543210fedcba98"
)

// bootstrapReleaseOrigin / bootstrapAPIOrigin are the two fetch origins the
// shipped script hardcodes. They are the ONLY seam these tests have, and they
// are a textual one: the script itself reads no environment variable to decide
// where to fetch from, because two successive attempts to keep such a variable
// behind a loopback allowlist were both defeated in adversarial review. A test
// therefore rewrites these literals in a throwaway COPY of the script (see
// bootstrapFixtureScript) rather than driving a redirection seam that would
// then also exist in production.
const (
	bootstrapReleaseOrigin = "https://github.com/REPPL/abcd-cli"
	bootstrapAPIOrigin     = "https://api.github.com/repos/REPPL/abcd-cli"
)

// bootstrapScript locates the committed hooks/bootstrap.sh from this test file's
// own on-disk position (internal/surface/cli/ -> three levels up), so the tests
// derive from the script that actually ships rather than from a hand-written
// stand-in. Callers that need to RUN it against a fixture go through
// bootstrapFixtureScript.
func bootstrapScript(t *testing.T) string {
	t.Helper()
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

// bootstrapRequires skips a case whose host cannot run the script at all.
func bootstrapRequires(t *testing.T) {
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
}

// bootstrapFixtureScript writes a per-test COPY of the committed script with its
// two hardcoded fetch origins rewritten to the fixture's base URL, and returns
// the copy's path.
//
// This is the whole test seam, and it exists on THIS side of the boundary on
// purpose: the shipped script must contain no branch, variable, or flag that can
// point a fetch somewhere else, because the binary and the checksums.txt that
// "verifies" it come from the same origin and the result is run unattended as
// the Bash shell guard on every tool call. Rewriting a copy keeps the production
// bytes identical to what runs on a user's machine while still exercising the
// full download/verify/install path.
//
// Every substitution is checked, both before and after. A silent no-op — the
// literal drifting in the script, a typo in the constant — would leave every
// case below pointed at the REAL GitHub while reporting a pass, which is exactly
// the fixture-hides-the-bug failure an earlier round already shipped once.
func bootstrapFixtureScript(t *testing.T, base string) string {
	t.Helper()
	src, err := os.ReadFile(bootstrapScript(t))
	if err != nil {
		t.Fatalf("reading the committed bootstrap script: %v", err)
	}
	body := string(src)
	// The QUOTED, complete assignment value — never a bare hostname and never a
	// bare prefix. A prefix match would happily rewrite a longer origin
	// (…/abcd-cli-renamed) into a mangled URL and carry on; requiring the closing
	// quote means the literal is matched whole or not at all. Each must appear
	// exactly once, or the script has changed shape in a way this harness no
	// longer models and would otherwise silently test against the real GitHub.
	for _, origin := range []string{bootstrapAPIOrigin, bootstrapReleaseOrigin} {
		quoted := `"` + origin + `"`
		if n := strings.Count(body, quoted); n != 1 {
			t.Fatalf("the fixture substitution expects exactly one %s in hooks/bootstrap.sh, found %d — the script's fetch origins have changed and this harness would otherwise silently test against the real GitHub", quoted, n)
		}
		body = strings.Replace(body, quoted, `"`+base+`"`, 1)
	}
	for _, origin := range []string{bootstrapAPIOrigin, bootstrapReleaseOrigin} {
		if strings.Contains(body, origin) {
			t.Fatalf("the rewritten script still names %q, so the fixture substitution did not take effect", origin)
		}
	}
	if n := strings.Count(body, base); n != 2 {
		t.Fatalf("the rewritten script must name the fixture base exactly twice (release origin + API origin), found %d", n)
	}
	path := filepath.Join(t.TempDir(), "bootstrap.sh")
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("writing the rewritten bootstrap script: %v", err)
	}
	return path
}

// bootstrapAsset is the release asset name the script derives from uname.
func bootstrapAsset() string { return "abcd-" + runtime.GOOS + "-" + runtime.GOARCH }

// bootstrapFixture is the stand-in release host plus the trust material a real
// curl needs to talk to it.
type bootstrapFixture struct {
	base string // https://127.0.0.1:<port>
	ca   string // PEM file holding the fixture's self-signed certificate
	hits *int32 // every request the script makes, counted
}

// bootstrapServer stands in for the release host, shaped like the real one
// rather than like the script's convenience: /releases/latest answers a 302 to
// /releases/tag/<tag> (the only place the tag is legible), and the tag-pinned
// asset URL redirects to an OPAQUE CDN path carrying no tag segment — which is
// what GitHub actually does (release-assets.githubusercontent.com/<numbers>?…),
// and the reason reading the tag off curl's effective URL resolved to "unknown"
// in production while passing against a friendlier fixture. checksums.txt is
// served only from the tag-pinned path, so a script that fetched it from
// /releases/latest/download instead gets a 404 rather than a silent pass.
// hits counts every request, which is how the no-network cases are proved.
//
// It speaks TLS, not plain http, because the script pins `--proto '=https'
// --proto-redir '=https'` on every call unconditionally and nothing can clear
// that. Trust is supplied through CURL_CA_BUNDLE (see runBootstrap): a CA bundle
// only ever ADDS an issuer curl will accept — it can never name a different
// server — so it is not the class of seam that was removed from the script.
func bootstrapServer(t *testing.T, body []byte, manifest string) *bootstrapFixture {
	t.Helper()
	var count int32
	asset := bootstrapAsset()
	pinned := "/releases/download/" + bootstrapTag + "/"
	mux := http.NewServeMux()
	// Deliberately outside the counter: the harness's own reachability probe
	// must not read as a fetch the script made.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		switch r.URL.Path {
		case "/releases/latest":
			http.Redirect(w, r, "/releases/tag/"+bootstrapTag, http.StatusFound)
		case pinned + asset:
			http.Redirect(w, r, "/cdn-blob/1292161760/10d722b0?sig=x", http.StatusFound)
		case "/cdn-blob/1292161760/10d722b0":
			_, _ = w.Write(body)
		case pinned + "checksums.txt":
			_, _ = w.Write([]byte(manifest))
		case "/commits/" + bootstrapTag:
			_, _ = w.Write([]byte(`{"sha":"` + bootstrapRelease + `","node_id":"x"}`))
		default:
			http.NotFound(w, r)
		}
	})
	srv := httptest.NewTLSServer(mux)
	t.Cleanup(srv.Close)

	ca := filepath.Join(t.TempDir(), "fixture-ca.pem")
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
	if err := os.WriteFile(ca, pemBytes, 0o600); err != nil {
		t.Fatalf("writing the fixture CA: %v", err)
	}
	fx := &bootstrapFixture{base: srv.URL, ca: ca, hits: &count}

	// Prove the harness itself works before any case blames the script. Without
	// this, a curl build that ignored CURL_CA_BUNDLE would surface as "the latest
	// release tag could not be resolved" and read like a script defect.
	probe := exec.Command("curl", "-fsS", "--proto", "=https", fx.base+"/healthz")
	probe.Env = append(fx.env(), "PATH="+os.Getenv("PATH"))
	if out, err := probe.CombinedOutput(); err != nil {
		t.Fatalf("the fixture TLS server is unreachable from curl with CURL_CA_BUNDLE set, so the harness cannot exercise the pinned-HTTPS script: %v (output %q)", err, out)
	}
	return fx
}

// env is the trust material the fixture needs, and nothing else. CURL_CA_BUNDLE
// and SSL_CERT_FILE are curl/OpenSSL's own variables, honoured by any curl
// invocation in the process environment; the script neither names nor branches
// on them. Both are set because the two curl TLS backends in play read different
// ones.
func (f *bootstrapFixture) env() []string {
	return []string{"CURL_CA_BUNDLE=" + f.ca, "SSL_CERT_FILE=" + f.ca}
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
	return bootstrapRootNamed(t, bootstrapCommit)
}

// bootstrapRootNamed makes a plugin root with an arbitrary basename, which is
// how the "the commit-stamped-cache warrant did not hold" case is reached.
func bootstrapRootNamed(t *testing.T, name string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

// runBootstrap executes a fixture-pointed copy of the script with an explicitly
// constructed environment — no inherited proxy or plugin variables — and returns
// its merged output and exit code. extraPath is prepended to PATH so a case can
// shim a command.
func runBootstrap(t *testing.T, root string, fx *bootstrapFixture, extraPath string) (string, int) {
	t.Helper()
	bootstrapRequires(t)
	return runScript(t, bootstrapFixtureScript(t, fx.base), root, fx.env(), extraPath)
}

// runScript executes a bootstrap script directly (not via `sh <path>`, so the
// executable bit and the shebang are exercised too) under a constructed
// environment.
func runScript(t *testing.T, script, root string, extraEnv []string, extraPath string) (string, int) {
	t.Helper()
	pathValue := os.Getenv("PATH")
	if extraPath != "" {
		pathValue = extraPath + string(os.PathListSeparator) + pathValue
	}
	cmd := exec.Command(script)
	cmd.Env = append([]string{
		"PATH=" + pathValue,
		"HOME=" + t.TempDir(),
		"CLAUDE_PLUGIN_ROOT=" + root,
	}, extraEnv...)
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			code = e.ExitCode()
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
	fx := bootstrapServer(t, []byte("payload"), bootstrapManifest([]byte("payload")))
	binary := filepath.Join(root, "abcd")
	if err := os.WriteFile(binary, []byte("already here"), 0o755); err != nil {
		t.Fatal(err)
	}

	out, code := runBootstrap(t, root, fx, "")
	if code != 0 {
		t.Errorf("a present binary must exit 0, got %d (output %q)", code, out)
	}
	if n := atomic.LoadInt32(fx.hits); n != 0 {
		t.Errorf("a present binary must cost no network, got %d request(s)", n)
	}
	got, err := os.ReadFile(binary)
	if err != nil || string(got) != "already here" {
		t.Errorf("the fast path must leave the existing binary untouched; got %q (%v)", got, err)
	}
}

// TestBootstrapInstallsVerifiedBinary is the fresh-install AC: download, verify
// against the release manifest, install executable, and record the provenance
// the skew notice reads. The exit code is 2, not 0: a SessionStart hook's stdout
// is model context and only a non-zero exit shows its stderr to the human, so
// the one-time `abcd ahoy install` suggestion would otherwise never be read by
// anyone who could act on it. SessionStart does not block on a non-zero exit.
func TestBootstrapInstallsVerifiedBinary(t *testing.T) {
	root := bootstrapRoot(t)
	body := []byte("#!/bin/sh\nexit 0\n")
	fx := bootstrapServer(t, body, bootstrapManifest(body))

	out, code := runBootstrap(t, root, fx, "")
	if code != 2 {
		t.Fatalf("a verified download must exit 2 (a notice the user is shown), got %d (output %q)", code, out)
	}
	// The substitution is proved BEHAVIOURALLY as well as textually: a copy that
	// still named the real origin could not have reached this fixture at all.
	if n := atomic.LoadInt32(fx.hits); n == 0 {
		t.Fatal("the fixture served no request, so the script was not pointed at it — the substitution silently missed")
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
	// release_tag has to be resolved from the release endpoint, not scraped off
	// the asset download's effective URL: the real one redirects to an opaque CDN
	// path, so a scraped tag is always "unknown" and the whole skew feature — and
	// the tag-pinned checksums URL that keeps the binary and its manifest from
	// coming from two different releases — dies with it.
	if meta["release_tag"] != bootstrapTag {
		t.Errorf("release_tag = %q, want %q resolved from the release endpoint", meta["release_tag"], bootstrapTag)
	}
	if meta["release_sha"] != bootstrapRelease {
		t.Errorf("release_sha = %q, want %q (the commit the release API reports)", meta["release_sha"], bootstrapRelease)
	}
	if meta["plugin_sha"] != bootstrapCommit {
		t.Errorf("plugin_sha = %q, want the plugin root's commit stamp", meta["plugin_sha"])
	}
	if meta["plugin_root_basename"] != bootstrapCommit {
		t.Errorf("plugin_root_basename = %q, want the raw plugin-root basename %q", meta["plugin_root_basename"], bootstrapCommit)
	}
	// The hash the verification just accepted is recorded, so a human (or a later
	// verb) can tell the binary that was verified from one that replaced it.
	// Nothing re-checks it today; the fast path deliberately stays a file test.
	sum := sha256.Sum256(body)
	if want := hex.EncodeToString(sum[:]); meta["binary_sha256"] != want {
		t.Errorf("binary_sha256 = %q, want the verified hash %q", meta["binary_sha256"], want)
	}
	if _, err := time.Parse(time.RFC3339, meta["fetched_at"]); err != nil {
		t.Errorf("fetched_at = %q, want an RFC3339 UTC timestamp (%v)", meta["fetched_at"], err)
	}
	if _, err := os.Stat(filepath.Join(root, ".bootstrap.lock")); !os.IsNotExist(err) {
		t.Error("the lock must be removed on the success path")
	}
}

// TestBootstrapRecordsTheRawPluginRootBasename is the observability half of the
// commit-stamped-cache WARRANT. plugin_sha is gated to exactly forty lowercase
// hex characters because itd-105 assumes the harness names each plugin cache
// directory for the commit it was cloned from. Nothing in this repository
// verifies that against the real harness. If it ever stops holding, the gate
// records `unknown` for every install, the version-skew notice goes silent
// permanently, and — without this raw field — there is nothing anywhere for a
// human to look at. The raw value is recorded, never compared and never
// rendered, so the notice's behaviour is unchanged.
func TestBootstrapRecordsTheRawPluginRootBasename(t *testing.T) {
	const notACommit = "abcd-plugin-v0.5.0"
	root := bootstrapRootNamed(t, notACommit)
	body := []byte("payload")
	fx := bootstrapServer(t, body, bootstrapManifest(body))

	out, code := runBootstrap(t, root, fx, "")
	if code != 2 {
		t.Fatalf("a plugin root that is not commit-stamped must still install, got %d (output %q)", code, out)
	}
	meta := metaValues(t, root)
	if meta["plugin_sha"] != "unknown" {
		t.Errorf("plugin_sha = %q, want unknown: %q is not a commit", meta["plugin_sha"], notACommit)
	}
	if meta["plugin_root_basename"] != notACommit {
		t.Errorf("plugin_root_basename = %q, want the raw basename %q — otherwise a warrant that stopped holding is invisible", meta["plugin_root_basename"], notACommit)
	}
}

// TestBootstrapRefusesChecksumMismatch is the trust bar: a download whose SHA-256
// disagrees with the release manifest never reaches the binary path, and the
// message says so.
func TestBootstrapRefusesChecksumMismatch(t *testing.T) {
	root := bootstrapRoot(t)
	fx := bootstrapServer(t, []byte("tampered"), bootstrapManifest([]byte("the published bytes")))

	out, code := runBootstrap(t, root, fx, "")
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
	fx := bootstrapServer(t, []byte("payload"), strings.Repeat("0", 64)+"  abcd-other-platform\n")

	out, code := runBootstrap(t, root, fx, "")
	if code == 0 {
		t.Errorf("an unlisted asset must fail loudly, got exit 0 (output %q)", out)
	}
	if !strings.Contains(out, "lists no entry") {
		t.Errorf("the refusal must say the manifest lists no entry for this platform, not a download or mismatch failure; output %q", out)
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
// nothing, and exit 2 — an unsupported platform is a reported condition, not a
// hook fault, but a SessionStart hook that exits 0 has only put its message into
// model context, where the human who would build from source never sees it.
// A non-zero SessionStart exit is non-blocking, so nothing is disrupted; the
// accepted cost is that this notice then repeats every session for as long as
// the platform is unsupported (itd-105's "every hook reports, for now" posture,
// with the one-notice-per-session aspiration tracked in iss-168).
func TestBootstrapUnsupportedPlatformChangesNothing(t *testing.T) {
	root := bootstrapRoot(t)
	fx := bootstrapServer(t, []byte("payload"), bootstrapManifest([]byte("payload")))
	shim := t.TempDir()
	fake := "#!/bin/sh\ncase \"$1\" in\n-s) echo Haiku ;;\n-m) echo sparc64 ;;\n*) echo Haiku ;;\nesac\n"
	if err := os.WriteFile(filepath.Join(shim, "uname"), []byte(fake), 0o755); err != nil {
		t.Fatal(err)
	}

	out, code := runBootstrap(t, root, fx, shim)
	if code != 2 {
		t.Errorf("an unsupported platform must exit 2 (a notice the user is shown), got %d (output %q)", code, out)
	}
	if n := atomic.LoadInt32(fx.hits); n != 0 {
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
	fx := bootstrapServer(t, []byte("payload"), bootstrapManifest([]byte("payload")))
	lock := filepath.Join(root, ".bootstrap.lock")
	if err := os.Mkdir(lock, 0o755); err != nil {
		t.Fatal(err)
	}

	out, code := runBootstrap(t, root, fx, "")
	if code != 0 {
		t.Errorf("losing the lock race must exit 0, got %d (output %q)", code, out)
	}
	if out != "" {
		t.Errorf("losing the lock race must be silent; output %q", out)
	}
	if n := atomic.LoadInt32(fx.hits); n != 0 {
		t.Errorf("losing the lock race must cost no network, got %d request(s)", n)
	}
	if _, err := os.Stat(lock); err != nil {
		t.Error("the loser must not remove the winner's live lock")
	}
}

// TestBootstrapUnusableLockPathRefusesLoudly is the other side of that silence.
// A lock that cannot be taken and is NOT there afterwards is not a race: the
// plugin root is unwritable, or something that is not a directory occupies the
// lock path. That condition is permanent and repeats every session, so sharing
// the loser's silent exit 0 with it would leave a plugin root that can never be
// provisioned and never says why.
func TestBootstrapUnusableLockPathRefusesLoudly(t *testing.T) {
	root := bootstrapRoot(t)
	fx := bootstrapServer(t, []byte("payload"), bootstrapManifest([]byte("payload")))
	// A regular FILE at the lock path: mkdir fails exactly as it does on a
	// read-only plugin root, and no lock directory exists afterwards. Unlike a
	// permission bit this reproduces for root as well as an ordinary user.
	if err := os.WriteFile(filepath.Join(root, ".bootstrap.lock"), []byte("not a lock"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, code := runBootstrap(t, root, fx, "")
	if code == 0 {
		t.Fatalf("an unusable lock path must not be mistaken for a lost race; got exit 0 (output %q)", out)
	}
	if !strings.Contains(out, "not writable") {
		t.Errorf("the refusal must say the plugin root is not writable; output %q", out)
	}
	if !strings.Contains(out, root) {
		t.Errorf("the refusal must name the plugin root %q; output %q", root, out)
	}
	if n := atomic.LoadInt32(fx.hits); n != 0 {
		t.Errorf("a run that never took the lock must fetch nothing, got %d request(s)", n)
	}
}

// TestBootstrapSweepsOrphanedTempDirs: SIGKILL runs no trap, so a killed run
// leaves its PID-stamped temp directory behind holding a partially downloaded,
// UNVERIFIED binary. Nothing else ever removes it, so it accumulates in the
// plugin root forever. A run that holds the lock sweeps them.
func TestBootstrapSweepsOrphanedTempDirs(t *testing.T) {
	root := bootstrapRoot(t)
	body := []byte("payload")
	fx := bootstrapServer(t, body, bootstrapManifest(body))
	orphan := filepath.Join(root, ".bootstrap.tmp.99999")
	if err := os.MkdirAll(orphan, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(orphan, "abcd-unverified"), []byte("not verified"), 0o755); err != nil {
		t.Fatal(err)
	}

	out, code := runBootstrap(t, root, fx, "")
	if code != 2 {
		t.Fatalf("the install must still proceed, got %d (output %q)", code, out)
	}
	if _, err := os.Stat(orphan); !os.IsNotExist(err) {
		t.Errorf("a temp directory orphaned by a killed run must be swept, %s survives (%v)", orphan, err)
	}
}

// TestBootstrapFastPathRejectsADirectory: `[ -x ]` is true of a DIRECTORY too, so
// a plugin root holding a directory named abcd would take the fast path forever
// and never provision anything. The check has to be a regular file, matching
// internal/core/ahoy's isExecutableFile — the two must agree about what
// "reachable binary" means or ahoy reports a gap the bootstrap refuses to fill.
// It must also REFUSE outright rather than proceed: `mv -f` onto an existing
// directory moves the downloaded file INTO it instead of replacing it, which
// would otherwise report success while the hooks stay broken.
func TestBootstrapFastPathRejectsADirectory(t *testing.T) {
	root := bootstrapRoot(t)
	fx := bootstrapServer(t, []byte("payload"), bootstrapManifest([]byte("payload")))
	if err := os.MkdirAll(filepath.Join(root, "abcd"), 0o755); err != nil {
		t.Fatal(err)
	}

	out, code := runBootstrap(t, root, fx, "")
	if code == 0 {
		t.Fatalf("a directory at the binary path must not read as a provisioned binary; output %q", out)
	}
	if !strings.Contains(out, "not a regular file") {
		t.Errorf("the refusal must name the obstruction; output %q", out)
	}
	if n := atomic.LoadInt32(fx.hits); n != 0 {
		t.Errorf("a pre-existing obstruction must be caught before any download, got %d request(s)", n)
	}
	if _, err := os.Stat(filepath.Join(root, "abcd", bootstrapAsset())); !os.IsNotExist(err) {
		t.Error("the downloaded asset must never be moved INTO the obstructing directory")
	}
}

// TestBootstrapStaleLockIsBroken: a lock left by a crashed run would otherwise
// make the plugin root unprovisionable forever, so one older than ten minutes is
// taken over.
func TestBootstrapStaleLockIsBroken(t *testing.T) {
	root := bootstrapRoot(t)
	body := []byte("payload")
	fx := bootstrapServer(t, body, bootstrapManifest(body))
	lock := filepath.Join(root, ".bootstrap.lock")
	if err := os.Mkdir(lock, 0o755); err != nil {
		t.Fatal(err)
	}
	stale := time.Now().Add(-30 * time.Minute)
	if err := os.Chtimes(lock, stale, stale); err != nil {
		t.Fatal(err)
	}

	out, code := runBootstrap(t, root, fx, "")
	if code != 2 {
		t.Fatalf("a stale lock must be broken and the install proceed, got %d (output %q)", code, out)
	}
	if _, err := os.Stat(filepath.Join(root, "abcd")); err != nil {
		t.Errorf("a stale lock must not block provisioning: %v", err)
	}
}

// TestBootstrapStripsControlCharactersFromMessages: both message sinks
// interpolate values the script does not author — the plugin-root path, uname's
// output, the release tag read off a redirect. A raw ESC in one of them can
// recolour, reposition, or visually rewrite what the reader is shown, on a
// message whose entire job is to be believed. A bare CR is the same class of
// attack by a different mechanism (it returns the cursor to column zero and
// overprints what the reader has already been shown — internal/termsafe's
// SanitizeBlock names this specifically), and DEL is stripped alongside it.
func TestBootstrapStripsControlCharactersFromMessages(t *testing.T) {
	t.Run("refuse", func(t *testing.T) {
		root := bootstrapRootNamed(t, "plugin\x1b[31mroot\rEVIL\x7f")
		fx := bootstrapServer(t, []byte("tampered"), bootstrapManifest([]byte("published")))

		out, code := runBootstrap(t, root, fx, "")
		if code == 0 {
			t.Fatalf("expected a refusal; output %q", out)
		}
		for _, r := range []rune{0x1b, '\r', 0x7f} {
			if strings.ContainsRune(out, r) {
				t.Errorf("the refusal must strip control character %U from the values it echoes; output %q", r, out)
			}
		}
	})
	t.Run("notice", func(t *testing.T) {
		root := bootstrapRoot(t)
		fx := bootstrapServer(t, []byte("payload"), bootstrapManifest([]byte("payload")))
		shim := t.TempDir()
		fake := "#!/bin/sh\ncase \"$1\" in\n-s) printf 'Hai\\033[31mku\\n' ;;\n-m) echo sparc64 ;;\n*) printf 'Hai\\033[31mku\\n' ;;\nesac\n"
		if err := os.WriteFile(filepath.Join(shim, "uname"), []byte(fake), 0o755); err != nil {
			t.Fatal(err)
		}

		out, code := runBootstrap(t, root, fx, shim)
		if code != 2 {
			t.Fatalf("expected the unsupported-platform notice; got %d (output %q)", code, out)
		}
		if strings.ContainsRune(out, 0x1b) {
			t.Errorf("the notice must strip control characters from the values it echoes; output %q", out)
		}
	})
}

// TestBootstrapSanitizesPluginRootBasename: plugin_root_basename is the RAW,
// unvalidated plugin-cache directory name, recorded so "why has the skew
// notice never fired" has an answer in the file when the commit-stamped-cache
// warrant does not hold. .binary-meta is a key=value text file another process
// parses with last-line-wins semantics, so a basename is not an inert label —
// an embedded newline could forge an additional key into that file, and an
// unbounded one could push the file past the guarded read size and silence the
// notice by a second route. Both must be neutralised before the value is ever
// written.
func TestBootstrapSanitizesPluginRootBasename(t *testing.T) {
	t.Run("embedded newline cannot forge a key", func(t *testing.T) {
		forged := strings.Repeat("f", 40)
		root := bootstrapRootNamed(t, "evil\nrelease_sha="+forged)
		body := []byte("payload")
		fx := bootstrapServer(t, body, bootstrapManifest(body))

		out, code := runBootstrap(t, root, fx, "")
		if code != 2 {
			t.Fatalf("the install must still succeed despite the hostile basename, got %d (output %q)", code, out)
		}
		meta := metaValues(t, root)
		if got := meta["release_sha"]; got == forged {
			t.Errorf("a newline in the plugin-root basename forged an extra release_sha line: %q", got)
		}
		lines := strings.Count(strings.TrimSpace(mustReadFile(t, filepath.Join(root, ".binary-meta"))), "\n") + 1
		if lines != 6 {
			t.Errorf(".binary-meta must hold exactly its six declared fields, got %d lines", lines)
		}
	})
	t.Run("length is capped", func(t *testing.T) {
		// 200 bytes: past the 120-byte cap this test exists to prove, but under the
		// 255-byte filename limit most filesystems enforce.
		root := bootstrapRootNamed(t, strings.Repeat("a", 200))
		body := []byte("payload")
		fx := bootstrapServer(t, body, bootstrapManifest(body))

		out, code := runBootstrap(t, root, fx, "")
		if code != 2 {
			t.Fatalf("the install must still succeed despite the oversized basename, got %d (output %q)", code, out)
		}
		got := metaValues(t, root)["plugin_root_basename"]
		if len(got) != 120 {
			t.Errorf("plugin_root_basename must be capped at 120 bytes, got %d", len(got))
		}
	})
}

// mustReadFile is a small helper so the tests above can assert on
// .binary-meta's raw shape, not just its parsed values.
func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// TestBootstrapScriptIsExecutable: the harness invokes the committed script by
// path, not as `sh <path>`, so a lost executable bit is a fresh-install failure
// no other test would catch — every case above runs a rewritten COPY this test
// file chmods itself.
func TestBootstrapScriptIsExecutable(t *testing.T) {
	fi, err := os.Stat(bootstrapScript(t))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm()&0o111 == 0 {
		t.Errorf("hooks/bootstrap.sh must be committed executable, got mode %v", fi.Mode().Perm())
	}
}

// TestBootstrapFetchOriginsAreConstants is the structural half of the security
// bar the two rounds before this one failed. The shipped script must hold its
// fetch origins as literals with NO environment-driven way to redirect them, and
// must pin HTTPS (redirects included) on every call with no toggle: the binary
// and the checksums.txt that "verifies" it come from one origin, so anything
// that can name that origin supplies both the payload and its own manifest, and
// the result is executed unattended as the Bash shell guard on every tool call.
func TestBootstrapFetchOriginsAreConstants(t *testing.T) {
	src, err := os.ReadFile(bootstrapScript(t))
	if err != nil {
		t.Fatal(err)
	}
	body := string(src)

	if strings.Contains(body, "ABCD_BOOTSTRAP") {
		t.Error("the shipped script must name no ABCD_BOOTSTRAP override: the seam was removed, not narrowed")
	}
	// A DENYLIST of the names a redirection seam has used so far (ABCD_BOOTSTRAP_*)
	// only ever catches a seam that reuses one of those names — a bypass that
	// renamed the variable (e.g. $GH_MIRROR) would sail through unnoticed. The
	// real invariant this test exists to hold is narrower and checkable directly:
	// every $NAME / ${NAME} reference the script contains, after subtracting the
	// names it assigns to itself, must be exactly {CLAUDE_PLUGIN_ROOT} — a
	// filesystem destination, never a fetch origin. This is an ALLOWLIST: an
	// unrecognised environment reference under any name fails it.
	envRefPattern := regexp.MustCompile(`\$\{?([A-Za-z_][A-Za-z0-9_]*)`)
	assignPattern := regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)=`)
	assigned := map[string]bool{}
	tainted := map[string]bool{}
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		m := assignPattern.FindStringSubmatch(trimmed)
		if m == nil {
			continue
		}
		name := m[1]
		// A SELF-referential assignment (name="${name:-default}") reads the
		// environment under its own name before falling back — exactly the
		// mirror-variable bypass that walked through an earlier version of this
		// allowlist, because the name was in `assigned` by the time its own
		// defining line was checked. Such a name is never treated as safe, on
		// ANY of its assignment lines, so both the self-reference itself and any
		// later use of the name are flagged as an unrecognised environment read.
		rhs := trimmed[len(m[0]):]
		if regexp.MustCompile(`\$\{?` + regexp.QuoteMeta(name) + `\b`).MatchString(rhs) {
			tainted[name] = true
			continue
		}
		assigned[name] = true
	}
	for name := range tainted {
		delete(assigned, name)
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		for _, m := range envRefPattern.FindAllStringSubmatch(line, -1) {
			name := m[1]
			if name == "CLAUDE_PLUGIN_ROOT" || assigned[name] {
				continue
			}
			if !seen[name] {
				seen[name] = true
				t.Errorf("unrecognised environment reference $%s in a non-comment line — every environment read must be CLAUDE_PLUGIN_ROOT or a name this script assigns itself: %q", name, line)
			}
		}
	}
	for _, origin := range []string{bootstrapReleaseOrigin, bootstrapAPIOrigin} {
		if n := strings.Count(body, origin); n != 1 {
			t.Errorf("the script must hold %q as a single literal, found %d", origin, n)
		}
	}
	// Every curl invocation pins the transport, unconditionally. A variable
	// holding the flags is what the previous round used, and clearing it on one
	// path is what reopened redirect downgrade.
	calls := 0
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || !strings.Contains(trimmed, "curl -") {
			continue
		}
		calls++
		if !strings.Contains(trimmed, "--proto '=https' --proto-redir '=https'") {
			t.Errorf("every curl call must pin HTTPS literally; %q does not", trimmed)
		}
	}
	if calls < 4 {
		t.Errorf("expected the script's four curl calls to be found, saw %d — the scan above is no longer matching", calls)
	}
}

// TestBootstrapIgnoresEveryEnvironmentOrigin is the behavioural half, and the
// direct regression for the defect two review rounds failed to close: the
// COMMITTED script (no substitution) is run with the removed override variables
// set to every shape that previously worked — a bare loopback base, the URL
// userinfo form that walked straight through the loopback glob
// (http://127.0.0.1:1@host), and a file:// origin. curl is shimmed so nothing
// leaves the machine and every invocation is recorded; the assertions are that
// the recorded URLs name GitHub and only GitHub, that the fixture server was
// never contacted, and that the transport pin is present on each call.
func TestBootstrapIgnoresEveryEnvironmentOrigin(t *testing.T) {
	bootstrapRequires(t)
	fx := bootstrapServer(t, []byte("attacker payload"), bootstrapManifest([]byte("attacker payload")))
	hostile := []struct{ name, value string }{
		// The plain loopback base the removed allowlist accepted by design.
		{"loopback base", fx.base},
		// The userinfo form an independent reviewer walked through the loopback
		// glob twice: everything before the @ is a username, so the real host is
		// abcd-bootstrap.invalid while the pattern reads 127.0.0.1.
		{"userinfo past the loopback glob", "http://127.0.0.1:1@abcd-bootstrap.invalid"},
		{"userinfo over https", "https://127.0.0.1:1@abcd-bootstrap.invalid"},
		{"local file origin", "file:///dev/null"},
	}
	for _, tc := range hostile {
		value := tc.value
		t.Run(tc.name, func(t *testing.T) {
			root := bootstrapRoot(t)
			shim := t.TempDir()
			log := filepath.Join(shim, "curl.log")
			// The shim records the argv it was handed and fails, standing in for
			// "no network": the script then refuses at tag resolution, and the log
			// is the evidence of where it TRIED to go.
			fake := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> " + log + "\nexit 1\n"
			if err := os.WriteFile(filepath.Join(shim, "curl"), []byte(fake), 0o755); err != nil {
				t.Fatal(err)
			}

			out, code := runScript(t, bootstrapScript(t), root, []string{
				"ABCD_BOOTSTRAP_BASE_URL=" + value,
				"ABCD_BOOTSTRAP_API_URL=" + value,
				"ABCD_BOOTSTRAP_URL=" + value,
			}, shim)
			if code == 0 {
				t.Fatalf("with no reachable network the script must refuse; got exit 0 (output %q)", out)
			}
			recorded, err := os.ReadFile(log)
			if err != nil {
				t.Fatalf("the script must have attempted at least one fetch: %v", err)
			}
			calls := strings.Split(strings.TrimSpace(string(recorded)), "\n")
			for _, call := range calls {
				if !strings.Contains(call, bootstrapReleaseOrigin) && !strings.Contains(call, bootstrapAPIOrigin) {
					t.Errorf("a fetch went somewhere other than the compiled origins: %q", call)
				}
				if !strings.Contains(call, "--proto =https --proto-redir =https") {
					t.Errorf("a fetch was made without the HTTPS pin: %q", call)
				}
			}
			for _, forbidden := range []string{"127.0.0.1", "abcd-bootstrap.invalid", "file://"} {
				if strings.Contains(string(recorded), forbidden) {
					t.Errorf("the environment redirected a fetch to %q; recorded calls: %q", forbidden, recorded)
				}
			}
			if n := atomic.LoadInt32(fx.hits); n != 0 {
				t.Errorf("the fixture server was contacted %d time(s): the override still redirects", n)
			}
			assertNothingInstalled(t, root, out)
		})
	}
}
