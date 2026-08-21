package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// spc-35: the harness's persistent per-plugin data directory
// ($CLAUDE_PLUGIN_DATA) is a download cache for the checksum-verified release
// artefact. The plugin ROOT is transient — every plugin update re-clones into a
// fresh commit-stamped directory and the old one is garbage-collected — so a
// binary stored only there is re-downloaded on every update. With the cache, a
// fresh root is provisioned by a checksum-verified COPY, and the network is
// asked for an artefact only when the released binary itself changed.

// runBootstrapWithData is runBootstrap with the persistent data directory set,
// which is how the harness invokes hooks in the field.
func runBootstrapWithData(t *testing.T, root, data string, fx *bootstrapFixture, extraPath string) (string, int) {
	t.Helper()
	bootstrapRequires(t)
	return runScript(t, bootstrapFixtureScript(t, fx.base), root,
		append(fx.env(), "CLAUDE_PLUGIN_DATA="+data), extraPath)
}

// runBootstrapWithDataHome is runBootstrapWithData with HOME pinned, so a test
// can seed and read the home-scoped owned-copy provenance record (spc-35 keeps
// it at $HOME/.abcd/path-entry, reachable from a terminal that has no
// CLAUDE_PLUGIN_DATA).
func runBootstrapWithDataHome(t *testing.T, root, data, home string, fx *bootstrapFixture, extraPath string) (string, int) {
	t.Helper()
	bootstrapRequires(t)
	return runScript(t, bootstrapFixtureScript(t, fx.base), root,
		append(fx.env(), "CLAUDE_PLUGIN_DATA="+data, "HOME="+home), extraPath)
}

// homePathEntry is $HOME/.abcd/path-entry, the home-scoped provenance record.
func homePathEntry(home string) string {
	return filepath.Join(home, ".abcd", "path-entry")
}

// seedHomePathEntry writes the provenance record under a pinned HOME.
func seedHomePathEntry(t *testing.T, home, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(home, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(homePathEntry(home), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// seedBootstrapCache plants a provisioned cache: the artefact plus the
// binary-meta record the bootstrap itself would have written for it.
func seedBootstrapCache(t *testing.T, data, tag string, body []byte) {
	t.Helper()
	cache := filepath.Join(data, "cache")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cache, bootstrapAsset()), body, 0o755); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	meta := "release_tag=" + tag + "\n" +
		"release_sha=" + bootstrapRelease + "\n" +
		"binary_sha256=" + hex.EncodeToString(sum[:]) + "\n" +
		"fetched_at=2026-08-01T00:00:00Z\n"
	if err := os.WriteFile(filepath.Join(cache, "binary-meta"), []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}
}

// cacheMetaValues parses the cache's binary-meta record.
func cacheMetaValues(t *testing.T, data string) map[string]string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(data, "cache", "binary-meta"))
	if err != nil {
		t.Fatalf("the cache must hold a binary-meta record: %v", err)
	}
	out := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if k, v, ok := strings.Cut(line, "="); ok {
			out[k] = v
		}
	}
	return out
}

// TestBootstrapProvisionsRootFromCacheWithoutDownload is itd-132's headline AC:
// a fresh plugin root plus a cached artefact whose recorded release matches the
// resolved latest is provisioned by verified copy — the refresh detector makes
// its one tag resolve and downloads NO artefact. The fixture serves different
// bytes than the cache holds, so a script that quietly re-downloaded could not
// pass by coincidence.
func TestBootstrapProvisionsRootFromCacheWithoutDownload(t *testing.T) {
	root := bootstrapRoot(t)
	data := t.TempDir()
	cached := []byte("#!/bin/sh\n# cached artefact\nexit 0\n")
	served := []byte("#!/bin/sh\n# served artefact\nexit 0\n")
	seedBootstrapCache(t, data, bootstrapTag, cached)
	// The published manifest authenticates the CACHED artefact (the cache is the
	// legitimate release for this tag), while the served asset bytes differ — so
	// a script that re-downloaded the asset could not pass by coincidence.
	fx := bootstrapServer(t, served, bootstrapManifest(cached))

	out, code := runBootstrapWithData(t, root, data, fx, "")
	if code != 2 {
		t.Fatalf("a cache-provisioned install must exit 2 (a notice the user is shown), got %d (output %q)", code, out)
	}
	got, err := os.ReadFile(filepath.Join(root, "abcd"))
	if err != nil {
		t.Fatalf("the root must be provisioned from the cache: %v", err)
	}
	if string(got) != string(cached) {
		t.Errorf("the root binary is not the cached artefact (it matches the served bytes: %v) — the cache was bypassed", string(got) == string(served))
	}
	if n := atomic.LoadInt32(fx.artefactHits); n != 0 {
		t.Errorf("a cache hit with an unchanged release must download no binary artefact, got %d asset request(s)", n)
	}
	// The published manifest IS fetched — that is how the cache is authenticated
	// before it is trusted (adr-46 decision 3).
	if n := atomic.LoadInt32(fx.manifestHits); n == 0 {
		t.Error("an online cache hit must fetch the published checksums.txt to authenticate the cache; no manifest request was recorded")
	}
	fi, err := os.Stat(filepath.Join(root, "abcd"))
	if err != nil || fi.Mode().Perm() != 0o755 {
		t.Errorf("the provisioned binary must be mode 0755: %v (%v)", fi, err)
	}
	if got := firstLine(out); !strings.HasPrefix(got, "abcd bootstrap: installed") {
		t.Errorf("the success must lead the first visible line; first line = %q", got)
	}
	// An online cache hit is authenticated against the published manifest, and
	// the notice says which trust it rests on (adr-46 decision 3).
	if !strings.Contains(out, "verified against the published") {
		t.Errorf("an online cache hit must name the manifest-verified trust; output %q", out)
	}
	// Cache-provisioned roots carry no root-local .binary-meta: the cache meta
	// is the one provenance record, and the skew notice reads the LIVE root.
	if _, err := os.Stat(filepath.Join(root, ".binary-meta")); !os.IsNotExist(err) {
		t.Errorf("a cache-provisioned root must not get a root-local .binary-meta: %v", err)
	}
}

// TestBootstrapCacheHitSurvivesOfflineResolve: the refresh detector's resolve is
// best-effort. When it fails (offline, or the release host is down), a verified
// cache still provisions the root — the alternative is a fresh plugin update
// with no working hooks for as long as the network is out.
func TestBootstrapCacheHitSurvivesOfflineResolve(t *testing.T) {
	root := bootstrapRoot(t)
	data := t.TempDir()
	cached := []byte("#!/bin/sh\n# cached artefact\nexit 0\n")
	seedBootstrapCache(t, data, bootstrapTag, cached)
	fx := bootstrapServer(t, []byte("never served"), bootstrapManifest([]byte("never served")))
	atomic.StoreInt32(fx.failLatest, 1)

	out, code := runBootstrapWithData(t, root, data, fx, "")
	if code != 2 {
		t.Fatalf("a cache hit must survive a failed resolve, got exit %d (output %q)", code, out)
	}
	got, err := os.ReadFile(filepath.Join(root, "abcd"))
	if err != nil || string(got) != string(cached) {
		t.Errorf("the root must be provisioned from the cache when the resolve fails; got %q (%v)", got, err)
	}
	if n := atomic.LoadInt32(fx.artefactHits); n != 0 {
		t.Errorf("no artefact may be fetched when the resolve fails and the cache holds one, got %d", n)
	}
	// Offline, no published manifest is reachable, so the cache is trusted at
	// corruption-evidence only — and the notice says so, never claiming a
	// manifest verification it did not perform (adr-46 decision 3).
	if !strings.Contains(out, "unauthenticated cache while offline") {
		t.Errorf("an offline cache hit must name the unauthenticated/offline trust; output %q", out)
	}
	if strings.Contains(out, "verified against the published") {
		t.Errorf("an offline cache hit must not claim manifest verification; output %q", out)
	}
}

// TestBootstrapSecondRootIsProvisionedFromCache is the plugin-update
// simulation: the first root's fetch populates the cache, and a SECOND fresh
// root — the directory a plugin update clones — is then provisioned with no
// further artefact download.
func TestBootstrapSecondRootIsProvisionedFromCache(t *testing.T) {
	data := t.TempDir()
	body := []byte("#!/bin/sh\nexit 0\n")
	fx := bootstrapServer(t, body, bootstrapManifest(body))

	root1 := bootstrapRoot(t)
	out, code := runBootstrapWithData(t, root1, data, fx, "")
	if code != 2 {
		t.Fatalf("the first root must install from the network, got %d (output %q)", code, out)
	}
	if n := atomic.LoadInt32(fx.artefactHits); n == 0 {
		t.Fatal("the first run downloaded nothing, so this case proves nothing")
	}
	after := atomic.LoadInt32(fx.artefactHits)
	if got, err := os.ReadFile(filepath.Join(data, "cache", bootstrapAsset())); err != nil || string(got) != string(body) {
		t.Fatalf("the fetch must land the verified artefact in the cache; got %q (%v)", got, err)
	}
	meta := cacheMetaValues(t, data)
	if meta["release_tag"] != bootstrapTag {
		t.Errorf("cache release_tag = %q, want %q", meta["release_tag"], bootstrapTag)
	}
	if meta["release_sha"] != bootstrapRelease {
		t.Errorf("cache release_sha = %q, want %q", meta["release_sha"], bootstrapRelease)
	}
	sum := sha256.Sum256(body)
	if want := hex.EncodeToString(sum[:]); meta["binary_sha256"] != want {
		t.Errorf("cache binary_sha256 = %q, want the verified hash %q", meta["binary_sha256"], want)
	}
	if _, err := time.Parse(time.RFC3339, meta["fetched_at"]); err != nil {
		t.Errorf("cache fetched_at = %q, want an RFC3339 UTC timestamp (%v)", meta["fetched_at"], err)
	}
	// One cache serves many roots, so a recorded provisioning-time root is
	// meaningless — the skew notice compares the LIVE root at render time.
	if _, has := meta["plugin_sha"]; has {
		t.Errorf("the cache meta must not record plugin_sha: %v", meta)
	}

	root2 := bootstrapRoot(t)
	out, code = runBootstrapWithData(t, root2, data, fx, "")
	if code != 2 {
		t.Fatalf("the second root must be provisioned, got %d (output %q)", code, out)
	}
	if got, err := os.ReadFile(filepath.Join(root2, "abcd")); err != nil || string(got) != string(body) {
		t.Errorf("the second root must hold the verified artefact; got %q (%v)", got, err)
	}
	if n := atomic.LoadInt32(fx.artefactHits); n != after {
		t.Errorf("the second root must be provisioned by copy, not re-download: artefact requests went %d -> %d", after, n)
	}
}

// TestBootstrapNewReleaseRefreshesCacheAndPathEntry: when the resolved latest
// tag differs from the cached one, the new artefact is fetched and verified
// into the cache under the unchanged spc-21 posture, the root is provisioned
// from it — and the abcd-owned PATH copy recorded in path-entry is refreshed in
// the same run, re-verified before the swap.
func TestBootstrapNewReleaseRefreshesCacheAndPathEntry(t *testing.T) {
	root := bootstrapRoot(t)
	data := t.TempDir()
	old := []byte("#!/bin/sh\n# old release\nexit 0\n")
	fresh := []byte("#!/bin/sh\n# new release\nexit 0\n")
	seedBootstrapCache(t, data, "v9.9.8", old)
	oldSum := sha256.Sum256(old)
	home := t.TempDir()
	pathDir := t.TempDir()
	pathCopy := filepath.Join(pathDir, "abcd")
	if err := os.WriteFile(pathCopy, old, 0o755); err != nil {
		t.Fatal(err)
	}
	entry := "path=" + pathCopy + "\nbinary_sha256=" + hex.EncodeToString(oldSum[:]) + "\nplugin_root=" + t.TempDir() + "\n"
	seedHomePathEntry(t, home, entry)
	fx := bootstrapServer(t, fresh, bootstrapManifest(fresh))

	out, code := runBootstrapWithDataHome(t, root, data, home, fx, "")
	if code != 2 {
		t.Fatalf("a new release must install, got %d (output %q)", code, out)
	}
	if got, err := os.ReadFile(filepath.Join(root, "abcd")); err != nil || string(got) != string(fresh) {
		t.Errorf("the root must hold the NEW artefact; got %q (%v)", got, err)
	}
	if got, err := os.ReadFile(filepath.Join(data, "cache", bootstrapAsset())); err != nil || string(got) != string(fresh) {
		t.Errorf("the cache must hold the NEW artefact; got %q (%v)", got, err)
	}
	meta := cacheMetaValues(t, data)
	freshSum := sha256.Sum256(fresh)
	if meta["release_tag"] != bootstrapTag || meta["binary_sha256"] != hex.EncodeToString(freshSum[:]) {
		t.Errorf("the cache meta must record the new release; got %v", meta)
	}
	if got, err := os.ReadFile(pathCopy); err != nil || string(got) != string(fresh) {
		t.Errorf("the owned PATH copy must be refreshed to the new release; got %q (%v)", got, err)
	}
	entryRaw, err := os.ReadFile(homePathEntry(home))
	if err != nil {
		t.Fatalf("path-entry must survive the refresh: %v", err)
	}
	if !strings.Contains(string(entryRaw), hex.EncodeToString(freshSum[:])) {
		t.Errorf("path-entry must record the refreshed hash; got %q", entryRaw)
	}
	// The record's plugin_root is re-stamped to the LIVE root each provision, so
	// a terminal (no CLAUDE_PLUGIN_ROOT) can still resolve the plugin root.
	if !strings.Contains(string(entryRaw), "plugin_root="+root+"\n") {
		t.Errorf("path-entry must re-stamp plugin_root to the live root %q; got %q", root, entryRaw)
	}
}

// TestBootstrapNewReleaseLeavesForeignPathFileAlone: the PATH refresh replaces
// only a file that still matches the provenance hash path-entry records.
// Anything else at that path is not abcd's to touch — whatever put it there
// owns it.
func TestBootstrapNewReleaseLeavesForeignPathFileAlone(t *testing.T) {
	root := bootstrapRoot(t)
	data := t.TempDir()
	old := []byte("#!/bin/sh\n# old release\nexit 0\n")
	fresh := []byte("#!/bin/sh\n# new release\nexit 0\n")
	foreign := []byte("#!/bin/sh\n# somebody else's abcd\nexit 0\n")
	seedBootstrapCache(t, data, "v9.9.8", old)
	oldSum := sha256.Sum256(old)
	home := t.TempDir()
	pathDir := t.TempDir()
	pathCopy := filepath.Join(pathDir, "abcd")
	if err := os.WriteFile(pathCopy, foreign, 0o755); err != nil {
		t.Fatal(err)
	}
	entry := "path=" + pathCopy + "\nbinary_sha256=" + hex.EncodeToString(oldSum[:]) + "\n"
	seedHomePathEntry(t, home, entry)
	fx := bootstrapServer(t, fresh, bootstrapManifest(fresh))

	out, code := runBootstrapWithDataHome(t, root, data, home, fx, "")
	if code != 2 {
		t.Fatalf("the install itself must proceed, got %d (output %q)", code, out)
	}
	if got, err := os.ReadFile(pathCopy); err != nil || string(got) != string(foreign) {
		t.Errorf("a file that does not match the recorded provenance hash must never be replaced; got %q (%v)", got, err)
	}
}

// TestBootstrapCorruptCacheRefusesLoudly is the trust bar at rest: every
// promotion out of the cache re-verifies against the recorded binary_sha256,
// and an artefact that no longer matches refuses loudly and installs nothing —
// a tampered or bit-rotted cache must never reach the guard path.
func TestBootstrapCorruptCacheRefusesLoudly(t *testing.T) {
	root := bootstrapRoot(t)
	data := t.TempDir()
	recorded := []byte("the recorded bytes")
	seedBootstrapCache(t, data, bootstrapTag, recorded)
	tampered := []byte("tampered bytes")
	if err := os.WriteFile(filepath.Join(data, "cache", bootstrapAsset()), tampered, 0o755); err != nil {
		t.Fatal(err)
	}
	// The meta is INTACT (records the recorded bytes' hash), so the online
	// authentication against the published manifest passes — and the
	// promotion-time re-hash of the corrupted artefact is what catches the
	// bytes-only corruption and refuses.
	fx := bootstrapServer(t, recorded, bootstrapManifest(recorded))

	out, code := runBootstrapWithData(t, root, data, fx, "")
	if code == 0 || code == 2 {
		t.Fatalf("a corrupted cache artefact must refuse loudly, got exit %d (output %q)", code, out)
	}
	if !strings.Contains(out, "SHA-256") {
		t.Errorf("the refusal must name the checksum mismatch; output %q", out)
	}
	if _, err := os.Stat(filepath.Join(root, "abcd")); !os.IsNotExist(err) {
		t.Error("a corrupted cache artefact must never be installed into the plugin root")
	}
	// The evidence is left in place, named for the human to remove — silently
	// re-fetching over it would heal tampering without anyone ever knowing.
	if got, err := os.ReadFile(filepath.Join(data, "cache", bootstrapAsset())); err != nil || string(got) != string(tampered) {
		t.Errorf("the mismatching artefact must be left as evidence; got %q (%v)", got, err)
	}
}

// TestBootstrapAuthenticatesCacheAgainstPublishedManifest is
// iss-2608210934566228: the cache promotion's "re-verify against recorded
// binary_sha256" is a CORRUPTION check, not a TAMPER check. The artefact and
// the binary-meta that records its expected hash both live in the cache,
// equally same-UID-writable, so an attacker who writes BOTH — a payload plus a
// meta recording its hash and the current release tag — passed the promotion
// gate and got an unverified binary installed at the guard path. When online
// (the tag resolved), the cached hash must be authenticated against the
// PUBLISHED checksums.txt for the resolved release before the cache is trusted;
// a mismatch discards the cache and falls to the download path.
func TestBootstrapAuthenticatesCacheAgainstPublishedManifest(t *testing.T) {
	root := bootstrapRoot(t)
	data := t.TempDir()
	poison := []byte("#!/bin/sh\n# forged payload\nexit 0\n")
	published := []byte("#!/bin/sh\n# the real release\nexit 0\n")
	// A self-consistent poisoned pair: the artefact AND the binary-meta that
	// vouches for it, both rewritten to record the poison's own hash under the
	// CURRENT release tag. Re-hashing the artefact against its co-located record
	// (the old design) passes — which is the whole defect.
	seedBootstrapCache(t, data, bootstrapTag, poison)
	fx := bootstrapServer(t, published, bootstrapManifest(published))

	out, code := runBootstrapWithData(t, root, data, fx, "")
	if code != 2 {
		t.Fatalf("the authenticated cache path must install the published release, got %d (output %q)", code, out)
	}
	got, err := os.ReadFile(filepath.Join(root, "abcd"))
	if err != nil {
		t.Fatalf("the root must be provisioned: %v", err)
	}
	if string(got) == string(poison) {
		t.Fatal("the POISONED artefact was installed at the guard path: the cache was trusted on its co-located record alone, which the same-UID attacker also wrote")
	}
	if string(got) != string(published) {
		t.Errorf("the root binary is neither poison nor published; got %q", got)
	}
	// The published checksums.txt was fetched to authenticate the cache, the
	// mismatch was detected, and the download path then fetched the real asset
	// to replace the tampered cache.
	if n := atomic.LoadInt32(fx.manifestHits); n == 0 {
		t.Error("the equal-tag cache path must fetch the published checksums.txt to authenticate the cache; no manifest request was made")
	}
	if n := atomic.LoadInt32(fx.artefactHits); n == 0 {
		t.Error("a rejected cache must be replaced by a real asset download; no asset request was made")
	}
	if c, err := os.ReadFile(filepath.Join(data, "cache", bootstrapAsset())); err != nil || string(c) != string(published) {
		t.Errorf("the tampered cache must be replaced by the verified download; got %q (%v)", c, err)
	}
	// The success notice rests its claim on the manifest, not on the cache.
	if !strings.Contains(out, "verified") {
		t.Errorf("the notice must name the verification the install rests on; output %q", out)
	}
}

// TestBootstrapDegradesLoudlyWithoutDataDir is AC 7: a harness that exports no
// persistent data directory gets today's per-root fetch — and the notice SAYS
// so, because a silent fallback would hide that every plugin update is paying
// a re-download that the platform's documented mechanism would have avoided.
func TestBootstrapDegradesLoudlyWithoutDataDir(t *testing.T) {
	root := bootstrapRoot(t)
	body := []byte("#!/bin/sh\nexit 0\n")
	fx := bootstrapServer(t, body, bootstrapManifest(body))

	out, code := runBootstrap(t, root, fx, "")
	if code != 2 {
		t.Fatalf("the degraded per-root fetch must still install, got %d (output %q)", code, out)
	}
	if got, err := os.ReadFile(filepath.Join(root, "abcd")); err != nil || string(got) != string(body) {
		t.Errorf("the degraded path must install the verified download; got %q (%v)", got, err)
	}
	if !strings.Contains(out, "persistent plugin data") {
		t.Errorf("the degradation must be said out loud, never a silent fallback; output %q", out)
	}
	// The degraded path is the spc-21 per-root fetch, root-local provenance
	// record included — the skew notice falls back to it.
	if _, err := os.Stat(filepath.Join(root, ".binary-meta")); err != nil {
		t.Errorf("the degraded path must still write the root-local provenance record: %v", err)
	}
}

// migrationRootMeta writes a pre-cache root provenance record for body.
func migrationRootMeta(t *testing.T, root string, body []byte, sum string) {
	t.Helper()
	if sum == "" {
		s := sha256.Sum256(body)
		sum = hex.EncodeToString(s[:])
	}
	meta := "release_tag=v9.9.7\n" +
		"release_sha=" + bootstrapRelease + "\n" +
		"binary_sha256=" + sum + "\n" +
		"fetched_at=2026-07-01T00:00:00Z\n" +
		"plugin_sha=" + bootstrapCommit + "\n" +
		"plugin_root_basename=" + bootstrapCommit + "\n"
	if err := os.WriteFile(filepath.Join(root, ".binary-meta"), []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestBootstrapMigratesVerifiedRootBinaryIntoCache: the one-way migration. A
// root provisioned before the cache existed holds a verified binary and its
// .binary-meta; the first run with an empty cache seeds the cache from it —
// re-verified against the recorded hash, because the seed is a promotion into
// the trusted location — with no network at all, and a second run no-ops.
func TestBootstrapMigratesVerifiedRootBinaryIntoCache(t *testing.T) {
	root := bootstrapRoot(t)
	data := t.TempDir()
	body := []byte("#!/bin/sh\n# pre-cache install\nexit 0\n")
	if err := os.WriteFile(filepath.Join(root, "abcd"), body, 0o755); err != nil {
		t.Fatal(err)
	}
	migrationRootMeta(t, root, body, "")
	fx := bootstrapServer(t, []byte("never served"), bootstrapManifest([]byte("never served")))

	out, code := runBootstrapWithData(t, root, data, fx, "")
	if code != 0 {
		t.Fatalf("the migration runs inside the fast path and must exit 0, got %d (output %q)", code, out)
	}
	if n := atomic.LoadInt32(fx.hits); n != 0 {
		t.Errorf("the migration must touch no network, got %d request(s)", n)
	}
	if got, err := os.ReadFile(filepath.Join(data, "cache", bootstrapAsset())); err != nil || string(got) != string(body) {
		t.Fatalf("the verified root binary must seed the cache; got %q (%v)", got, err)
	}
	meta := cacheMetaValues(t, data)
	if meta["release_tag"] != "v9.9.7" || meta["release_sha"] != bootstrapRelease {
		t.Errorf("the seeded meta must carry the root record's provenance; got %v", meta)
	}
	if _, has := meta["plugin_sha"]; has {
		t.Errorf("the seeded cache meta must drop plugin_sha: %v", meta)
	}
	if _, err := os.Stat(filepath.Join(data, ".bootstrap.lock")); !os.IsNotExist(err) {
		t.Error("the migration must release the cache lock")
	}

	// Second run: the cache is populated, so the fast path stays a no-op.
	before, err := os.Stat(filepath.Join(data, "cache", bootstrapAsset()))
	if err != nil {
		t.Fatal(err)
	}
	out, code = runBootstrapWithData(t, root, data, fx, "")
	if code != 0 || out != "" {
		t.Fatalf("the second run must be a silent no-op, got %d (output %q)", code, out)
	}
	after, err := os.Stat(filepath.Join(data, "cache", bootstrapAsset()))
	if err != nil {
		t.Fatal(err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Error("the second run re-seeded a cache that was already populated")
	}
}

// TestBootstrapMigrationIgnoresMismatchedRootBinary: a root binary that does
// not match its own recorded hash is not evidence of anything — it is ignored,
// nothing is seeded, and the next fresh root fetches from the release host as
// spc-21 always did.
func TestBootstrapMigrationIgnoresMismatchedRootBinary(t *testing.T) {
	root := bootstrapRoot(t)
	data := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "abcd"), []byte("#!/bin/sh\n# replaced\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	other := sha256.Sum256([]byte("the bytes that were actually verified"))
	migrationRootMeta(t, root, nil, hex.EncodeToString(other[:]))
	fx := bootstrapServer(t, []byte("never served"), bootstrapManifest([]byte("never served")))

	out, code := runBootstrapWithData(t, root, data, fx, "")
	if code != 0 {
		t.Fatalf("the fast path must still exit 0, got %d (output %q)", code, out)
	}
	if _, err := os.Stat(filepath.Join(data, "cache", bootstrapAsset())); !os.IsNotExist(err) {
		t.Errorf("a hash-mismatched root binary must never seed the cache: %v", err)
	}
}

// TestBootstrapMigrationRefusesDirectoryAtCacheBinary is iss-2608210934566229:
// the migration seed fast-path test `[ ! -f "$cache_binary" ]` is TRUE for a
// DIRECTORY, so a directory planted at the cache artefact path passed it, the
// seed `mv -f` moved the verified binary INTO the directory, and a lying
// binary-meta then vouched for it — every fresh plugin root downloaded ~11 MB
// and hit refuse, running the shell guard UNGUARDED every session until a human
// removed the directory by hand, and this survived every plugin update. The
// migration seed must refuse the non-regular-file shape exactly as the main
// install site does.
func TestBootstrapMigrationRefusesDirectoryAtCacheBinary(t *testing.T) {
	root := bootstrapRoot(t)
	data := t.TempDir()
	body := []byte("#!/bin/sh\n# pre-cache install\nexit 0\n")
	if err := os.WriteFile(filepath.Join(root, "abcd"), body, 0o755); err != nil {
		t.Fatal(err)
	}
	migrationRootMeta(t, root, body, "")
	cacheBinDir := filepath.Join(data, "cache", bootstrapAsset())
	if err := os.MkdirAll(cacheBinDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fx := bootstrapServer(t, []byte("never served"), bootstrapManifest([]byte("never served")))

	out, code := runBootstrapWithData(t, root, data, fx, "")
	if code == 0 {
		t.Fatalf("a directory at the cache artefact path must be refused, not seeded into; got exit 0 (output %q)", out)
	}
	if !strings.Contains(out, "not a regular file") {
		t.Errorf("the refusal must name the obstruction; output %q", out)
	}
	entries, err := os.ReadDir(cacheBinDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("the seed must never be moved INTO the obstructing directory: %v", entries)
	}
	if _, err := os.Stat(filepath.Join(data, "cache", "binary-meta")); !os.IsNotExist(err) {
		t.Error("a lying binary-meta must not be written for a directory-shaped cache artefact")
	}
	if n := atomic.LoadInt32(fx.hits); n != 0 {
		t.Errorf("the obstruction must be caught with no network, got %d request(s)", n)
	}
}

// TestBootstrapCacheModeSkipsPathRefreshOnCacheHit documents the refresh
// boundary: the PATH copy is refreshed when a NEW artefact lands in the cache,
// not on every cache-hit provisioning — a hit means nothing changed, so there
// is nothing to refresh.
func TestBootstrapCacheModeSkipsPathRefreshOnCacheHit(t *testing.T) {
	root := bootstrapRoot(t)
	data := t.TempDir()
	cached := []byte("#!/bin/sh\n# cached artefact\nexit 0\n")
	seedBootstrapCache(t, data, bootstrapTag, cached)
	stale := []byte("#!/bin/sh\n# stale path copy\nexit 0\n")
	staleSum := sha256.Sum256(stale)
	home := t.TempDir()
	pathDir := t.TempDir()
	pathCopy := filepath.Join(pathDir, "abcd")
	if err := os.WriteFile(pathCopy, stale, 0o755); err != nil {
		t.Fatal(err)
	}
	entry := "path=" + pathCopy + "\nbinary_sha256=" + hex.EncodeToString(staleSum[:]) + "\nplugin_root=" + t.TempDir() + "\n"
	seedHomePathEntry(t, home, entry)
	fx := bootstrapServer(t, cached, bootstrapManifest(cached))

	out, code := runBootstrapWithDataHome(t, root, data, home, fx, "")
	if code != 2 {
		t.Fatalf("the cache hit must provision the root, got %d (output %q)", code, out)
	}
	if got, err := os.ReadFile(pathCopy); err != nil || string(got) != string(stale) {
		t.Errorf("a cache hit must not touch the PATH copy (`abcd ahoy install` heals on demand); got %q (%v)", got, err)
	}
	// A cache hit does not touch the PATH COPY, but it still re-stamps the
	// record's plugin_root to the live root — the terminal's route home tracks
	// the latest root within one hook firing of any update.
	entryRaw, err := os.ReadFile(homePathEntry(home))
	if err != nil {
		t.Fatalf("path-entry must survive a cache hit: %v", err)
	}
	if !strings.Contains(string(entryRaw), "plugin_root="+root+"\n") {
		t.Errorf("a cache hit must re-stamp plugin_root to the live root %q; got %q", root, entryRaw)
	}
	if !strings.Contains(string(entryRaw), "binary_sha256="+hex.EncodeToString(staleSum[:])) {
		t.Errorf("the re-stamp must preserve the recorded path and hash; got %q", entryRaw)
	}
}

// TestBootstrapCachePathsStayOutOfMessagesRaw guards the cache-mode messages
// the same way the plugin-root messages are guarded: a data-dir path is not
// this script's own text, so control characters in it are stripped before any
// message renders.
func TestBootstrapCachePathsStayOutOfMessagesRaw(t *testing.T) {
	root := bootstrapRoot(t)
	data := filepath.Join(t.TempDir(), "data\x1b[31mdir")
	if err := os.MkdirAll(data, 0o755); err != nil {
		t.Fatal(err)
	}
	recorded := []byte("the recorded bytes")
	seedBootstrapCache(t, data, bootstrapTag, recorded)
	if err := os.WriteFile(filepath.Join(data, "cache", bootstrapAsset()), []byte("tampered"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Meta intact -> online authentication passes, and the promotion re-hash
	// catches the corrupted bytes and refuses (naming the cache path).
	fx := bootstrapServer(t, recorded, bootstrapManifest(recorded))

	out, code := runBootstrapWithData(t, root, data, fx, "")
	if code == 0 || code == 2 {
		t.Fatalf("expected the corrupt-cache refusal; got %d (output %q)", code, out)
	}
	if strings.ContainsRune(out, 0x1b) {
		t.Errorf("the refusal must strip control characters from the cache path it echoes; output %q", out)
	}
}
