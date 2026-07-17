package identity

import (
	"os"
	"os/exec"
	"testing"
)

func mustGitID(t *testing.T, dir string, args ...string) {
	t.Helper()
	if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// TestEffectiveIdentityIgnoresInjectedConfig is the attack-input test for the
// gitConfig env scrub: the commit-identity gate reads user.name/user.email, so an
// injected GIT_CONFIG_* (or GIT_DIR) must not forge or redirect the identity the
// gate verifies. ScrubbedEnv strips the hijack vars while keeping global config.
func TestEffectiveIdentityIgnoresInjectedConfig(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	// GIT_AUTHOR_* must be unset so EffectiveIdentity falls through to gitConfig.
	t.Setenv("GIT_AUTHOR_NAME", "")
	t.Setenv("GIT_AUTHOR_EMAIL", "")
	dir := t.TempDir()
	mustGitID(t, dir, "init")
	mustGitID(t, dir, "config", "user.email", "real@example.com")
	mustGitID(t, dir, "config", "user.name", "Real Name")

	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "user.email")
	t.Setenv("GIT_CONFIG_VALUE_0", "forged@example.com")

	eff, err := EffectiveIdentity(dir)
	if err != nil {
		t.Fatalf("EffectiveIdentity: %v", err)
	}
	if eff.Email != "real@example.com" {
		t.Errorf("commit-identity gate honoured an injected GIT_CONFIG_* email: got %q, want real@example.com", eff.Email)
	}
}
