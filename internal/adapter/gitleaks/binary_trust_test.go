package gitleaks

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// The tests in this file pin the binary-admission rule that closes
// GHSA-fg9r-3f8g-89m6: a committed .abcd/config/gitleaks.json must never be able
// to point execution at repository content, and a PATH lookup must never
// resolve into the checkout either. Every refusal is loud (an error the history
// store fails closed on) and never a silent fallback to PATH.

// writeExecutable creates a mode-0755 regular file at path and returns it.
func writeExecutable(t *testing.T, path string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// trustProbe is an Adapter whose fake LookPath and Runner both record that they
// were reached, so a test can assert a refused path neither spawned anything
// nor fell back to PATH.
type trustProbe struct {
	*Adapter
	runner *fakeRunner
	looked *bool
}

func newTrustProbe(lookPathResult string) trustProbe {
	looked := false
	runner := &fakeRunner{report: "[]"}
	return trustProbe{
		Adapter: &Adapter{
			LookPath: func(string) (string, error) { looked = true; return lookPathResult, nil },
			Runner:   runner,
		},
		runner: runner,
		looked: &looked,
	}
}

// assertRefused checks that Augment refused with ErrConfiguredPathRefused and
// touched neither the runner nor the PATH lookup.
func assertRefused(t *testing.T, p trustProbe, repo string, cfg Config) {
	t.Helper()
	_, err := p.Augment(context.Background(), repo, cfg, "x\n", "transcript")
	if err == nil {
		t.Fatal("Augment accepted the binary; want a loud refusal")
	}
	if !errors.Is(err, ErrConfiguredPathRefused) {
		t.Fatalf("error is not ErrConfiguredPathRefused: %v", err)
	}
	if p.runner.called {
		t.Error("runner was invoked for a refused binary")
	}
	if cfg.Path != "" && *p.looked {
		t.Error("a refused configured path fell back to PATH lookup")
	}
}

// TestRefusesCommittedRelativePath is the advisory's exact attack: a committed
// config naming a relative path, plus a mode-100755 file at that path in the
// checkout, with the process working directory inside the checkout.
func TestRefusesCommittedRelativePath(t *testing.T) {
	repo := t.TempDir()
	writeExecutable(t, filepath.Join(repo, ".abcd", "tools", "g"))
	t.Chdir(repo)
	assertRefused(t, newTrustProbe("/fake/bin/gitleaks"), repo, Config{Enabled: true, Path: filepath.Join(".abcd", "tools", "g")})
}

// TestRefusesAbsolutePathInsideRepo: an absolute path is not enough — the
// checkout is attacker content wherever it is mounted.
func TestRefusesAbsolutePathInsideRepo(t *testing.T) {
	repo := t.TempDir()
	bin := writeExecutable(t, filepath.Join(repo, "tools", "g"))
	assertRefused(t, newTrustProbe("/fake/bin/gitleaks"), repo, Config{Enabled: true, Path: bin})
}

// TestRefusesSymlinkInsideRepoPointingOutside: a committed symlink under the
// checkout that targets a real executable elsewhere is still repository content
// (the link is the attacker's) and is refused on its lexical location.
func TestRefusesSymlinkInsideRepoPointingOutside(t *testing.T) {
	repo := t.TempDir()
	outside := writeExecutable(t, filepath.Join(t.TempDir(), "gitleaks"))
	link := filepath.Join(repo, ".abcd", "tools", "g")
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	assertRefused(t, newTrustProbe("/fake/bin/gitleaks"), repo, Config{Enabled: true, Path: link})
}

// TestRefusesOutsideSymlinkResolvingIntoRepo: a path outside the checkout whose
// symlink target resolves INTO it is refused on its resolved location.
func TestRefusesOutsideSymlinkResolvingIntoRepo(t *testing.T) {
	repo := t.TempDir()
	target := writeExecutable(t, filepath.Join(repo, "tools", "g"))
	link := filepath.Join(t.TempDir(), "gitleaks")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	assertRefused(t, newTrustProbe("/fake/bin/gitleaks"), repo, Config{Enabled: true, Path: link})
}

// TestRefusesNonExecutableFile: a regular file outside the repo that is not
// executable is refused rather than handed to exec to fail later.
func TestRefusesNonExecutableFile(t *testing.T) {
	repo := t.TempDir()
	plain := filepath.Join(t.TempDir(), "gitleaks")
	if err := os.WriteFile(plain, []byte("not a binary"), 0o644); err != nil {
		t.Fatal(err)
	}
	assertRefused(t, newTrustProbe("/fake/bin/gitleaks"), repo, Config{Enabled: true, Path: plain})
}

// TestRefusesNonRegularFile: a directory (or any non-regular file) is refused.
func TestRefusesNonRegularFile(t *testing.T) {
	repo := t.TempDir()
	assertRefused(t, newTrustProbe("/fake/bin/gitleaks"), repo, Config{Enabled: true, Path: t.TempDir()})
}

// TestLookPathResultInsideRepoRefused: the PATH fallback is held to the same
// rule — a PATH entry pointing into the checkout (a direnv-style
// $PWD/bin) must not make repository content executable.
func TestLookPathResultInsideRepoRefused(t *testing.T) {
	repo := t.TempDir()
	inRepo := writeExecutable(t, filepath.Join(repo, "bin", "gitleaks"))
	assertRefused(t, newTrustProbe(inRepo), repo, Config{Enabled: true})
}

// TestLookPathRelativeResultRefused: a relative lookup result (a "." or ""
// PATH entry) is refused whatever the working directory holds.
func TestLookPathRelativeResultRefused(t *testing.T) {
	repo := t.TempDir()
	writeExecutable(t, filepath.Join(repo, "gitleaks"))
	t.Chdir(repo)
	assertRefused(t, newTrustProbe("gitleaks"), repo, Config{Enabled: true})
}

// TestAdmitsAbsoluteExecutableOutsideRepo is the allow-shape: an absolute path
// to a regular executable outside the checkout runs, and what runs is the
// RESOLVED path — the bytes that were judged — not the configured spelling.
// The candidate is reached through a symlink outside the repo (the Homebrew
// shape: /usr/local/bin/gitleaks -> …/Cellar/…/gitleaks) to pin that.
func TestAdmitsAbsoluteExecutableOutsideRepo(t *testing.T) {
	repo := t.TempDir()
	target := writeExecutable(t, filepath.Join(t.TempDir(), "cellar", "gitleaks"))
	link := filepath.Join(t.TempDir(), "gitleaks")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	p := newTrustProbe("/fake/bin/gitleaks")
	if _, err := p.Augment(context.Background(), repo, Config{Enabled: true, Path: link}, "x\n", "transcript"); err != nil {
		t.Fatalf("admissible binary refused: %v", err)
	}
	if !p.runner.called {
		t.Fatal("runner was not invoked for an admissible binary")
	}
	want, err := filepath.EvalSymlinks(target)
	if err != nil {
		t.Fatal(err)
	}
	if p.runner.binPath != want {
		t.Errorf("runner executed %q, want the resolved path %q", p.runner.binPath, want)
	}
	if *p.looked {
		t.Error("a configured path must not consult PATH")
	}
}

// TestRefusalWithoutRepoRootFailsClosed: a configured path with no known
// repository root cannot be judged, so it is refused rather than trusted.
func TestRefusalWithoutRepoRootFailsClosed(t *testing.T) {
	bin := writeExecutable(t, filepath.Join(t.TempDir(), "gitleaks"))
	assertRefused(t, newTrustProbe("/fake/bin/gitleaks"), "", Config{Enabled: true, Path: bin})
}

// TestRefusesRelativeRepoRoot: a relative repository root cannot anchor the
// containment check (filepath.Rel would fail and read as "outside"), so it is
// refused rather than silently disabling the rule.
func TestRefusesRelativeRepoRoot(t *testing.T) {
	repo := t.TempDir()
	bin := writeExecutable(t, filepath.Join(repo, "tools", "g"))
	t.Chdir(repo)
	assertRefused(t, newTrustProbe("/fake/bin/gitleaks"), ".", Config{Enabled: true, Path: bin})
}

// TestCaseFoldingRefusesCaseVariantAndNFDSpellings: on a case-folding
// filesystem (APFS/HFS+, Windows) a respelt path — a component in another
// case, or the NFD form of a non-ASCII name — addresses the SAME in-repo file,
// so a byte-exact containment check would judge it "outside" and execute
// repository content. The predicate is forced so the branch is proved on any
// host (the seam ahoy, launch and lifeboat use for the same reason).
func TestCaseFoldingRefusesCaseVariantAndNFDSpellings(t *testing.T) {
	restore := caseFoldingFS
	t.Cleanup(func() { caseFoldingFS = restore })
	caseFoldingFS = func() bool { return true }

	// The repo root's last component is mixed-case and non-ASCII so both
	// respellings are meaningful.
	parent := t.TempDir()
	repo := filepath.Join(parent, "MyRépo") // NFC é
	bin := writeExecutable(t, filepath.Join(repo, ".abcd", "tools", "g"))

	upper := filepath.Join(parent, "MYRÉPO", ".abcd", "tools", "g") // upper-cased, NFC É
	nfd := filepath.Join(parent, "MyRépo", ".abcd", "tools", "g")  // NFD: e + combining acute

	for name, path := range map[string]string{"case-variant": upper, "nfd": nfd} {
		t.Run(name, func(t *testing.T) {
			if _, err := os.Stat(path); err != nil {
				t.Skipf("host filesystem does not fold this spelling: %v", err)
			}
			assertRefused(t, newTrustProbe("/fake/bin/gitleaks"), repo, Config{Enabled: true, Path: path})
		})
	}
	// The exact spelling is refused too, independent of the fold.
	assertRefused(t, newTrustProbe("/fake/bin/gitleaks"), repo, Config{Enabled: true, Path: bin})
}

// TestCaseSensitiveKeepsExactMatch: with the fold off (Linux), a sibling that
// differs only by case is a genuinely different directory and is admitted.
func TestCaseSensitiveKeepsExactMatch(t *testing.T) {
	restore := caseFoldingFS
	t.Cleanup(func() { caseFoldingFS = restore })
	caseFoldingFS = func() bool { return false }

	parent := t.TempDir()
	repo := filepath.Join(parent, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	sibling := filepath.Join(parent, "REPO", "gitleaks")
	if _, err := os.Stat(filepath.Join(parent, "REPO")); err == nil {
		t.Skip("host filesystem folds case; the sibling would be the repo itself")
	}
	bin := writeExecutable(t, sibling)
	p := newTrustProbe("/fake/bin/gitleaks")
	if _, err := p.Augment(context.Background(), repo, Config{Enabled: true, Path: bin}, "x\n", "transcript"); err != nil {
		t.Fatalf("case-distinct sibling refused on a case-sensitive host: %v", err)
	}
}
