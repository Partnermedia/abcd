package ahoy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The harness's persistent per-plugin data directory (spc-35) is always an
// absolute path, never inside a project checkout, and never world-writable.
// A CLAUDE_PLUGIN_DATA of any other shape is not that directory, so `ahoy
// install` must not treat what it finds there as a verified release artefact
// (sub-finding of GHSA-4q78-ccfv-f374): a relative value resolves against the
// checkout the verb runs in, an in-checkout value blesses committed bytes as
// the owned PATH binary, and a world-writable cache lets any local user swap
// both the artefact and its recorded hash. The design fork the parent record
// holds open (binding the cache to an attestation the env cannot supply) is
// untouched by these; they only refuse shapes the harness never produces.

// assertCacheIgnored: the owned copy was not written from the cache, no
// provenance record vouches for anything, and a note names the variable.
func assertCacheIgnored(t *testing.T, target string, res InstallResult) {
	t.Helper()
	if fi, err := os.Lstat(target); err == nil && fi.Mode().IsRegular() {
		if got, _ := os.ReadFile(target); string(got) == string(cacheArtefact) {
			t.Fatalf("install copied the cache artefact out of an untrusted data dir into %s", target)
		}
	}
	if _, err := os.Stat(userPathEntryPath()); err == nil {
		t.Fatalf("install recorded provenance for an untrusted cache; notes: %v", res.Notes)
	}
	said := false
	for _, n := range res.Notes {
		if strings.Contains(n, "CLAUDE_PLUGIN_DATA") {
			said = true
		}
	}
	if !said {
		t.Fatalf("no note names the ignored data directory; notes: %v", res.Notes)
	}
}

func adoptableRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	return repo
}

func TestInstallIgnoresRelativeDataCache(t *testing.T) {
	home, _ := setupUserScope(t)
	binDir := filepath.Join(home, ".local", "bin")
	t.Setenv("PATH", binDir)
	work := t.TempDir()
	seedDataCacheAt(t, filepath.Join(work, "data"), cacheArtefact)
	t.Chdir(work)
	t.Setenv("CLAUDE_PLUGIN_DATA", "data")

	res, err := Install(adoptableRepo(t), installOpts(), RefusingPrompter{})
	if err != nil {
		t.Fatal(err)
	}
	assertCacheIgnored(t, filepath.Join(binDir, "abcd"), res)
}

func TestInstallIgnoresDataCacheInsideTheCheckout(t *testing.T) {
	home, _ := setupUserScope(t)
	binDir := filepath.Join(home, ".local", "bin")
	t.Setenv("PATH", binDir)
	repo := adoptableRepo(t)
	data := filepath.Join(repo, ".plugin-data")
	seedDataCacheAt(t, data, cacheArtefact)
	t.Setenv("CLAUDE_PLUGIN_DATA", data)

	res, err := Install(repo, installOpts(), RefusingPrompter{})
	if err != nil {
		t.Fatal(err)
	}
	assertCacheIgnored(t, filepath.Join(binDir, "abcd"), res)

	// Detection agrees: the pinned symlink the install degraded to is not
	// offered a heal from a cache that lives inside the checkout.
	det, err := Detect(repo)
	if err != nil {
		t.Fatal(err)
	}
	if g := gapByID(det.Gaps, "symlink.legacy"); g != nil {
		t.Fatalf("detection offers a heal from an in-checkout cache: %+v", *g)
	}
}

func TestInstallIgnoresWorldWritableDataCache(t *testing.T) {
	home, _ := setupUserScope(t)
	binDir := filepath.Join(home, ".local", "bin")
	t.Setenv("PATH", binDir)
	data := seedDataCache(t, cacheArtefact)
	if err := os.Chmod(filepath.Join(data, "cache"), 0o777); err != nil {
		t.Fatal(err)
	}

	res, err := Install(adoptableRepo(t), installOpts(), RefusingPrompter{})
	if err != nil {
		t.Fatal(err)
	}
	assertCacheIgnored(t, filepath.Join(binDir, "abcd"), res)
}
