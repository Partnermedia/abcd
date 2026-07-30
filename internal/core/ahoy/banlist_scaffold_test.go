package ahoy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/adapter/scanner"
	"github.com/REPPL/abcd-cli/internal/core/banlist"
	"github.com/REPPL/abcd-cli/internal/gittest"
)

// TestInstallScaffoldsTheBanlistArtefacts is spc-20 AC5 at the install seam: a repo
// abcd configures inherits the whole two-layer arrangement — the public family in
// the docs-lint config, the guard hook in the repo's committed hooks, and the
// gitignored private stub — rather than a maintainer hand-wiring three files.
func TestInstallScaffoldsTheBanlistArtefacts(t *testing.T) {
	setupHermetic(t)
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := Install(repo, installOpts(), RefusingPrompter{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "clean" {
		t.Fatalf("install status = %q (remaining=%v), want clean", res.Status, res.Remaining)
	}

	hook := filepath.Join(repo, filepath.FromSlash(GuardHookRelPath))
	fi, err := os.Stat(hook)
	if err != nil {
		t.Fatalf("guard hook not scaffolded: %v", err)
	}
	if fi.Mode().Perm()&0o111 == 0 {
		t.Errorf("guard hook is not executable: mode %v", fi.Mode().Perm())
	}
	if _, err := os.Stat(filepath.Join(repo, filepath.FromSlash(banlist.PublicConfigRelPath))); err != nil {
		t.Errorf("public docs-lint config not scaffolded: %v", err)
	}
	pub, err := banlist.ListPublic(repo)
	if err != nil {
		t.Fatalf("scaffolded public config does not load: %v", err)
	}
	if !pub.Present {
		t.Errorf("public banned-names family absent after scaffolding")
	}
	priv, err := banlist.ListPrivate(repo)
	if err != nil {
		t.Fatalf("scaffolded private stub does not load: %v", err)
	}
	if !priv.Present {
		t.Errorf("private banlist stub absent after scaffolding")
	}
	if !priv.Keyed {
		t.Errorf("private stub does not declare the keyed format, so a later `add` is refused as legacy")
	}
	if len(priv.Entries) != 0 || len(priv.Malformed) != 0 {
		t.Errorf("stub seeded live entries (%v) or unusable lines (%v); its examples must be commented out",
			priv.Entries, priv.Malformed)
	}
}

// TestBanlistStubSeedsOnlyReservedIdentifiers is the AC5 seeding clause: the stub is
// scaffolded into every managed repo, so an example value that looked like a real
// host or address would teach the shape of a leak. Every illustrative identifier
// comes from a reserved documentation range (examples-use-reserved-identifiers), and
// the repo's own network-identifier detector is the judge — one primitive decides
// what "reserved" means, here and in the privacy-hygiene rule.
func TestBanlistStubSeedsOnlyReservedIdentifiers(t *testing.T) {
	// The stub's examples are regular expressions, so their dots are backslash-
	// escaped. Scanning them as written would prove nothing — no detector matches
	// `192\.0\.2\.17` — so the escaping is removed first and the plain identifiers
	// are what gets judged.
	plain := strings.ReplaceAll(privateStubContent(), `\`, "")
	scan := func(text string) []scanner.Finding {
		return scanner.ScanText(text, scanner.Identity{}, scanner.NetworkPatterns(), nil, "private-names.txt")
	}
	// The detector must be armed, or "no findings" is a test that cannot fail: a
	// control value from RFC 1918 has to be flagged before the stub's silence means
	// anything (fix-the-detector).
	control := "\n#   lab-ipv4      10.11.12.13\n" // abcd-audit:allow — the control the detector must flag
	if len(scan(plain+control)) == 0 {
		t.Fatal("the network-identifier detector reported nothing for a control RFC 1918 address; the assertion below is vacuous")
	}
	for _, f := range scan(plain) {
		t.Errorf("stub seeds a non-reserved identifier: kind=%s line=%d", f.Kind, f.Line)
	}
	// The positive half: the seeded examples actually exercise the reserved ranges a
	// user needs to see, so the stub teaches the convention rather than merely
	// avoiding a violation.
	for _, want := range []string{"192.0.2.", "2001:db8:", "example.com", "alice-laptop", "00:00:5e:00:53:"} {
		if !strings.Contains(plain, want) {
			t.Errorf("stub omits the reserved example %q", want)
		}
	}
}

// TestInstallNeverClobbersAHandWrittenGuardHook: a repo's pre-commit hook is the
// maintainer's, and it may carry gates abcd knows nothing about. Scaffolding creates
// the hook it is missing and leaves an existing one exactly as it found it.
func TestInstallNeverClobbersAHandWrittenGuardHook(t *testing.T) {
	setupHermetic(t)
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	hook := filepath.Join(repo, filepath.FromSlash(GuardHookRelPath))
	if err := os.MkdirAll(filepath.Dir(hook), 0o755); err != nil {
		t.Fatal(err)
	}
	const handWritten = "#!/usr/bin/env bash\nexit 0\n"
	if err := os.WriteFile(hook, []byte(handWritten), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(repo, installOpts(), RefusingPrompter{}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(hook)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != handWritten {
		t.Errorf("install overwrote a hand-written pre-commit hook")
	}
}

// TestInstallNeverClobbersAnExistingDocsLintConfig: the public config gates CI, so a
// repo that already carries one keeps it byte for byte.
func TestInstallNeverClobbersAnExistingDocsLintConfig(t *testing.T) {
	setupHermetic(t)
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(repo, filepath.FromSlash(banlist.PublicConfigRelPath))
	if err := os.MkdirAll(filepath.Dir(cfg), 0o755); err != nil {
		t.Fatal(err)
	}
	const existing = "{\n  \"roots\": [\"docs\"],\n  \"banned_tokens\": []\n}\n"
	if err := os.WriteFile(cfg, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(repo, installOpts(), RefusingPrompter{}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != existing {
		t.Errorf("install rewrote an existing docs-lint config:\n%s", got)
	}
}

// TestInstallNeverOverwritesAPopulatedPrivateStore is the one that matters most: the
// store holds the patterns whose literal text must not leave the machine, and a
// scaffold that re-seeded it would delete them.
func TestInstallNeverOverwritesAPopulatedPrivateStore(t *testing.T) {
	setupHermetic(t)
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	store := filepath.Join(repo, filepath.FromSlash(banlist.PrivateRelPath))
	if err := os.MkdirAll(filepath.Dir(store), 0o700); err != nil {
		t.Fatal(err)
	}
	existing := "# abcd-banlist: keyed\nlab-host alice-laptop\\.example\\.com\n"
	if err := os.WriteFile(store, []byte(existing), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(repo, installOpts(), RefusingPrompter{}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(store)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != existing {
		t.Errorf("install overwrote a populated private store")
	}
}

// TestScaffoldedGuardHookHoldsItsContract executes the artefact ahoy installs rather
// than reading it: the guard's value is entirely in what it does at commit time, and
// a scaffolded hook that parsed but refused nothing would look identical on disk
// (guards-prove-themselves). It pins the three behaviours the surface promises —
// refuse by key with the pattern withheld, refuse to stage the store itself, and
// warn loudly rather than block when the layer is absent.
func TestScaffoldedGuardHookHoldsItsContract(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash unavailable")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	setupHermetic(t)
	repo := t.TempDir()
	env := gittest.Env(t)
	git := func(args ...string) (string, error) {
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	if out, err := git("init"); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	if out, err := git("config", "user.name", "Alice Example"); err != nil {
		t.Fatalf("git config: %v\n%s", err, out)
	}
	if out, err := git("config", "user.email", "alice@example.com"); err != nil {
		t.Fatalf("git config: %v\n%s", err, out)
	}
	if _, err := Install(repo, installOpts(), RefusingPrompter{}); err != nil {
		t.Fatal(err)
	}
	// git runs .git/hooks/pre-commit; the scaffolded artefact is the committed
	// .githooks/pre-commit, so the test wires the two the way a clone's hooks path
	// does.
	src, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(GuardHookRelPath)))
	if err != nil {
		t.Fatal(err)
	}
	hooksDir := filepath.Join(repo, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hooksDir, "pre-commit"), src, 0o755); err != nil {
		t.Fatal(err)
	}

	// The scaffolded stub has no live entry, so the guard warns loudly and lets the
	// commit through: silence must never impersonate protection.
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := git("add", "README.md"); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
	out, err := git("commit", "-m", "seed")
	if err != nil {
		t.Fatalf("the scaffolded guard blocked a clean commit: %v\n%s", err, out)
	}
	if !strings.Contains(out, "NO ENTRIES") {
		t.Errorf("an entry-less store must warn loudly; got:\n%s", out)
	}

	// One real entry: a match refuses the commit and names the key alone.
	store := filepath.Join(repo, filepath.FromSlash(banlist.PrivateRelPath))
	if err := os.WriteFile(store, []byte("# abcd-banlist: keyed\nlab-host carol-server\\.example\\.net\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "notes.md"), []byte("ssh carol-server.example.net\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := git("add", "notes.md"); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
	out, err = git("commit", "-m", "leak")
	if err == nil {
		t.Fatalf("the scaffolded guard let a banned name through:\n%s", out)
	}
	if !strings.Contains(out, "lab-host") {
		t.Errorf("the refusal does not name the key:\n%s", out)
	}
	if strings.Contains(out, "carol-server") {
		t.Errorf("the refusal echoes the pattern or the matched text:\n%s", out)
	}

	// Staging the store itself is refused: the whole layer rests on it being
	// untracked, and the guard cannot catch its own source.
	if out, err := git("reset"); err != nil {
		t.Fatalf("git reset: %v\n%s", err, out)
	}
	if out, err := git("add", "-f", filepath.FromSlash(banlist.PrivateRelPath)); err != nil {
		t.Fatalf("git add -f: %v\n%s", err, out)
	}
	if out, err := git("commit", "-m", "stage the store"); err == nil {
		t.Fatalf("the scaffolded guard allowed the private store to be committed:\n%s", out)
	}
}
