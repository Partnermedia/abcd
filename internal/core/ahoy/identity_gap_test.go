package ahoy

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/identity"
	"github.com/REPPL/abcd-cli/internal/gittest"
)

func idGitRepo(t *testing.T, name, email string) string {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	dir := t.TempDir()
	idMustGit(t, dir, "init")
	if name != "" {
		idMustGit(t, dir, "config", "user.name", name)
	}
	if email != "" {
		idMustGit(t, dir, "config", "user.email", email)
	}
	return dir
}

func idMustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = gittest.Env(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func idWritePin(t *testing.T, root, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".abcd", "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".abcd", "config", "identity.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectGitIdentity_Mismatch(t *testing.T) {
	dir := idGitRepo(t, "Test User", "test@example.com")
	idWritePin(t, dir, `{"name":"Alex Reppel","email":"alex@example.com"}`)
	gaps := detectGitIdentity(dir)
	if len(gaps) != 1 || gaps[0].ID != "git_identity.mismatch" || !gaps[0].Required {
		t.Fatalf("want one required git_identity.mismatch gap, got %+v", gaps)
	}
}

func TestDetectGitIdentity_Match(t *testing.T) {
	dir := idGitRepo(t, "Alex Reppel", "alex@example.com")
	idWritePin(t, dir, `{"name":"Alex Reppel","email":"alex@example.com"}`)
	if gaps := detectGitIdentity(dir); len(gaps) != 0 {
		t.Fatalf("want no gap on match, got %+v", gaps)
	}
}

func TestDetectGitIdentity_Unpinned(t *testing.T) {
	dir := idGitRepo(t, "Alex Reppel", "alex@example.com")
	gaps := detectGitIdentity(dir)
	if len(gaps) != 1 || gaps[0].ID != "git_identity.unpinned" || gaps[0].Required {
		t.Fatalf("want one advisory git_identity.unpinned gap, got %+v", gaps)
	}
}

func TestStepIdentityPin_WritesFromCurrentIdentity(t *testing.T) {
	dir := idGitRepo(t, "Alex Reppel", "alex@example.com")
	a := &applyCtx{
		cwd:        dir,
		approved:   map[GapCategory]bool{ConfigChange: true},
		gapPresent: map[string]bool{"git_identity.unpinned": true},
	}
	a.stepIdentityPin()
	got, ok, err := identity.LoadPin(dir)
	if err != nil || !ok {
		t.Fatalf("pin not written: ok=%v err=%v", ok, err)
	}
	if got.Name != "Alex Reppel" || got.Email != "alex@example.com" {
		t.Fatalf("wrong pin written: %+v", got)
	}
	if len(a.writes) != 1 || a.writes[0] != identity.PinRelPath {
		t.Fatalf("expected one noted write of the pin, got %v", a.writes)
	}
}

func TestStepIdentityPin_SkipsUnderYes(t *testing.T) {
	dir := idGitRepo(t, "Test User", "test@example.com")
	a := &applyCtx{
		cwd:        dir,
		approved:   map[GapCategory]bool{ConfigChange: true},
		gapPresent: map[string]bool{"git_identity.unpinned": true},
		autoYes:    true,
	}
	a.stepIdentityPin()
	if _, ok, _ := identity.LoadPin(dir); ok {
		t.Fatal("--yes must not silently pin the current (possibly sandbox) identity")
	}
}

func TestStepIdentityPin_NeverAutoResolvesMismatch(t *testing.T) {
	dir := idGitRepo(t, "Test User", "test@example.com")
	a := &applyCtx{
		cwd:        dir,
		approved:   map[GapCategory]bool{ConfigChange: true},
		gapPresent: map[string]bool{"git_identity.mismatch": true},
	}
	a.stepIdentityPin()
	if _, ok, _ := identity.LoadPin(dir); ok {
		t.Fatal("a mismatch must never auto-write a pin")
	}
}
