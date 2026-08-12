package ahoy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupUserScope drives the DEFAULT (no-sudo) install location: HOME is
// redirected, the ABCD_BIN_TARGET test override is cleared so binTarget()
// resolves to ~/.local/bin/abcd, and PATH is exactly the dirs given — so a test
// can state, rather than inherit, what the machine has on PATH.
func setupUserScope(t *testing.T, pathDirs ...string) (home, pluginRoot string) {
	t.Helper()
	home, pluginRoot = setupHermetic(t)
	t.Setenv("ABCD_BIN_TARGET", "")
	t.Setenv("PATH", strings.Join(pathDirs, string(os.PathListSeparator)))
	return home, pluginRoot
}

// managedRepo returns a temp dir that classifies as a managed repo (a marker
// block fired), so the deeper gap checks run.
func managedRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	body := "# Project\n\n<!-- BEGIN ABCD -->\nx\n<!-- END ABCD -->\n"
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// linkOwned plants an owned install (a symlink to the plugin binary) at path.
func linkOwned(t *testing.T, path, pluginRoot string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(pluginBinaryPath(pluginRoot), path); err != nil {
		t.Fatal(err)
	}
}

func gapByID(gaps []Gap, id string) *Gap {
	for i := range gaps {
		if gaps[i].ID == id {
			return &gaps[i]
		}
	}
	return nil
}

// TestDetectAdoptsOwnedInstallOnPath is iss-171's headline defect: a working
// ~/.local/bin/abcd — the field-standard single-user location — must read as
// installed, not as symlink.missing, while the detector is itself running as
// abcd from PATH.
func TestDetectAdoptsOwnedInstallOnPath(t *testing.T) {
	home, pluginRoot := setupUserScope(t)
	binDir := filepath.Join(home, ".local", "bin")
	t.Setenv("PATH", binDir)
	linkOwned(t, filepath.Join(binDir, "abcd"), pluginRoot)

	det, err := Detect(managedRepo(t))
	if err != nil {
		t.Fatal(err)
	}
	if hasGap(det.Gaps, "symlink.missing") {
		t.Errorf("a working %s reported symlink.missing: %+v", filepath.Join(binDir, "abcd"), det.Gaps)
	}
	if m, _ := det.Signals["install_mode"].(string); m != "pinned" {
		t.Errorf("install_mode = %q, want pinned", m)
	}
}

// TestDetectAdoptsOwnedInstallAnywhereOnPath proves the detector scans PATH
// rather than recognising one blessed target: an owned install in any PATH
// directory counts.
func TestDetectAdoptsOwnedInstallAnywhereOnPath(t *testing.T) {
	_, pluginRoot := setupUserScope(t)
	other := filepath.Join(t.TempDir(), "opt", "bin")
	t.Setenv("PATH", other)
	linkOwned(t, filepath.Join(other, "abcd"), pluginRoot)

	det, err := Detect(managedRepo(t))
	if err != nil {
		t.Fatal(err)
	}
	if hasGap(det.Gaps, "symlink.missing") {
		t.Errorf("owned install at %s reported symlink.missing: %+v", other, det.Gaps)
	}
}

// TestDetectDanglingOwnedSymlinkIsItsOwnGap pins the shadowing failure mode: an
// owned symlink whose target has gone is NOT a healthy install, and it is not
// silence either — it is its own gap.
func TestDetectDanglingOwnedSymlinkIsItsOwnGap(t *testing.T) {
	home, pluginRoot := setupUserScope(t)
	binDir := filepath.Join(home, ".local", "bin")
	t.Setenv("PATH", binDir)
	linkOwned(t, filepath.Join(binDir, "abcd"), pluginRoot)
	if err := os.Remove(pluginBinaryPath(pluginRoot)); err != nil {
		t.Fatal(err)
	}

	det, err := Detect(managedRepo(t))
	if err != nil {
		t.Fatal(err)
	}
	g := gapByID(det.Gaps, "symlink.dangling")
	if g == nil {
		t.Fatalf("dangling owned symlink produced no symlink.dangling gap: %+v", det.Gaps)
	}
	if !g.Required {
		t.Errorf("symlink.dangling must be required: %+v", g)
	}
	if m, _ := det.Signals["install_mode"].(string); m == "pinned" {
		t.Errorf("a dangling symlink reported install_mode=pinned")
	}
}

// TestDetectBinDirNotOnPathIsALoudGap pins the script-first remedy: an install
// directory that is not on PATH is its own loud gap carrying the one-line export
// fix, and abcd never offers to patch a shell profile (the gap is not resolvable
// by an apply step).
func TestDetectBinDirNotOnPathIsALoudGap(t *testing.T) {
	home, pluginRoot := setupUserScope(t, filepath.Join(t.TempDir(), "elsewhere"))
	linkOwned(t, filepath.Join(home, ".local", "bin", "abcd"), pluginRoot)

	det, err := Detect(managedRepo(t))
	if err != nil {
		t.Fatal(err)
	}
	g := gapByID(det.Gaps, "path.bin_dir_not_on_path")
	if g == nil {
		t.Fatalf("no path.bin_dir_not_on_path gap: %+v", det.Gaps)
	}
	if !g.Required {
		t.Errorf("the PATH gap must be required (loud): %+v", g)
	}
	if g.Resolvable {
		t.Errorf("the PATH gap must NOT be resolvable — abcd prints the fix, it never patches a shell profile: %+v", g)
	}
	if want := `export PATH="$HOME/.local/bin:$PATH"`; !strings.Contains(g.FixHint, want) {
		t.Errorf("fix hint = %q, want it to carry %q", g.FixHint, want)
	}
}

// TestGapTextCarriesNoAbsoluteHomePath pins the receipt-hygiene half of the
// hand-off from iss-177: the moment the default install location moves under
// $HOME, any gap line embedding it carries the developer's username. Gap text
// renders user-scope paths in tilde form.
func TestGapTextCarriesNoAbsoluteHomePath(t *testing.T) {
	home, _ := setupUserScope(t, filepath.Join(t.TempDir(), "elsewhere"))

	det, err := Detect(managedRepo(t))
	if err != nil {
		t.Fatal(err)
	}
	sawTilde := false
	for _, g := range det.Gaps {
		for _, field := range []string{g.Title, g.Detail, g.FixHint} {
			if strings.Contains(field, home) {
				t.Errorf("gap %s leaks the absolute home path %q: %q", g.ID, home, field)
			}
		}
		if strings.Contains(g.Detail, "~/.local/bin/abcd") || strings.Contains(g.Title, "~/.local/bin") {
			sawTilde = true
		}
	}
	if !sawTilde {
		t.Errorf("no gap named the default install location in tilde form: %+v", det.Gaps)
	}
}

// TestInstallDefaultsToUserLocalBin is decision 5 of the install-experience
// plan: install writes ~/.local/bin/abcd, creating the directory, with no
// privilege escalation anywhere.
func TestInstallDefaultsToUserLocalBin(t *testing.T) {
	home, pluginRoot := setupUserScope(t)
	binDir := filepath.Join(home, ".local", "bin")
	t.Setenv("PATH", binDir)
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	res, err := Install(repo, installOpts(), RefusingPrompter{})
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(binDir, "abcd")
	fi, err := os.Lstat(target)
	if err != nil {
		t.Fatalf("install did not create %s: %v", target, err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("install wrote a regular file, want a symlink")
	}
	dest, _ := os.Readlink(target)
	if resolveSymlinkDest(target, dest) != resolvePath(pluginBinaryPath(pluginRoot)) {
		t.Errorf("symlink dest = %q, want %q", dest, pluginBinaryPath(pluginRoot))
	}
	// The write is recorded on the receipt through the same note seam as every
	// other apply step, so the receipt scrub owned by iss-177 covers it. Scrubbing
	// the receipt itself belongs to that change, not this one — what this change
	// owns is the two print sites OUTSIDE the seam (the gap text and the uninstall
	// receipt), asserted by their own tests.
	if !containsPath(res.Writes, target) {
		t.Errorf("the PATH entry was not recorded on the receipt: %v", res.Writes)
	}
	if m, _ := detectSignal(t, repo, "install_mode").(string); m != "pinned" {
		t.Errorf("install_mode = %q, want pinned", m)
	}
	_ = home
}

func containsPath(paths []string, want string) bool {
	for _, p := range paths {
		if p == want {
			return true
		}
	}
	return false
}

// detectSignal returns one detection signal for repo.
func detectSignal(t *testing.T, repo, key string) any {
	t.Helper()
	det, err := Detect(repo)
	if err != nil {
		t.Fatal(err)
	}
	return det.Signals[key]
}

// TestInstallAdoptsOwnedInstallInPlace proves install never plants a second
// install beside a working one: an owned entry already on PATH is adopted where
// it stands.
func TestInstallAdoptsOwnedInstallInPlace(t *testing.T) {
	home, pluginRoot := setupUserScope(t)
	other := filepath.Join(t.TempDir(), "opt", "bin")
	t.Setenv("PATH", other)
	linkOwned(t, filepath.Join(other, "abcd"), pluginRoot)
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := Install(repo, installOpts(), RefusingPrompter{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(home, ".local", "bin", "abcd")); !os.IsNotExist(err) {
		t.Errorf("install planted a second entry at ~/.local/bin beside the adopted one: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(other, "abcd")); err != nil {
		t.Errorf("install disturbed the adopted install: %v", err)
	}
}

// TestInstallRefusesToCreateDanglingSymlink is the shadowing refusal: with no
// binary at <plugin-root>/abcd, install must NOT write a symlink at all — a
// dangling link on PATH shadows whatever else would have answered.
func TestInstallRefusesToCreateDanglingSymlink(t *testing.T) {
	home, pluginRoot := setupUserScope(t)
	binDir := filepath.Join(home, ".local", "bin")
	t.Setenv("PATH", binDir)
	if err := os.Remove(pluginBinaryPath(pluginRoot)); err != nil {
		t.Fatal(err)
	}
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	res, err := Install(repo, installOpts(), RefusingPrompter{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(binDir, "abcd")); !os.IsNotExist(err) {
		t.Fatalf("install created a symlink to a non-existent target: %v", err)
	}
	joined := strings.Join(res.Notes, "\n")
	if !strings.Contains(joined, "does not exist") {
		t.Errorf("the refusal was silent; notes = %v", res.Notes)
	}
}

// TestInstallBinDirUnwritableFailsLoudly pins the system-wide path: an explicit
// --bin-dir abcd cannot write to is an error, never a silent skip and never a
// privilege escalation.
func TestInstallBinDirUnwritableFailsLoudly(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: every directory is writable, so the refusal cannot be observed")
	}
	setupUserScope(t)
	ro := filepath.Join(t.TempDir(), "ro")
	if err := os.Mkdir(ro, 0o555); err != nil {
		t.Fatal(err)
	}
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	opts := installOpts()
	opts.BinDir = ro

	_, err := Install(repo, opts, RefusingPrompter{})
	if err == nil {
		t.Fatalf("an unwritable --bin-dir was accepted silently")
	}
	msg := err.Error()
	for _, want := range []string{"not writable", "privilege"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error = %q, want it to carry %q", msg, want)
		}
	}
	if _, err := os.Lstat(filepath.Join(ro, "abcd")); !os.IsNotExist(err) {
		t.Errorf("something was written into the unwritable dir: %v", err)
	}
}

// TestInstallBinDirWritableInstallsThere pins the opt-in: an explicit, writable
// --bin-dir is where the entry lands.
func TestInstallBinDirWritableInstallsThere(t *testing.T) {
	_, pluginRoot := setupUserScope(t)
	dir := filepath.Join(t.TempDir(), "opt", "bin")
	t.Setenv("PATH", dir)
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	opts := installOpts()
	opts.BinDir = dir

	if _, err := Install(repo, opts, RefusingPrompter{}); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "abcd")
	dest, err := os.Readlink(target)
	if err != nil {
		t.Fatalf("--bin-dir install did not create %s: %v", target, err)
	}
	if resolveSymlinkDest(target, dest) != resolvePath(pluginBinaryPath(pluginRoot)) {
		t.Errorf("symlink dest = %q, want %q", dest, pluginBinaryPath(pluginRoot))
	}
}

// TestUninstallReceiptCarriesNoAbsoluteHomePath is the second half of the
// iss-177 hand-off: `ahoy uninstall` prints the receipt's symlink target
// verbatim, so the target must be rendered in tilde form once the default
// location lives under $HOME.
func TestUninstallReceiptCarriesNoAbsoluteHomePath(t *testing.T) {
	home, pluginRoot := setupUserScope(t)
	binDir := filepath.Join(home, ".local", "bin")
	t.Setenv("PATH", binDir)
	linkOwned(t, filepath.Join(binDir, "abcd"), pluginRoot)

	receipt, err := Uninstall(managedRepo(t))
	if err != nil {
		t.Fatal(err)
	}
	if !receipt.Symlink.Removed {
		t.Fatalf("uninstall did not remove the owned install: %+v", receipt.Symlink)
	}
	if strings.Contains(receipt.Symlink.Target, home) {
		t.Errorf("receipt leaks the absolute home path: %q", receipt.Symlink.Target)
	}
	if want := "~/.local/bin/abcd"; receipt.Symlink.Target != want {
		t.Errorf("receipt target = %q, want %q", receipt.Symlink.Target, want)
	}
}
