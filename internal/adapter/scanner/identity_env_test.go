package scanner

import (
	"os"
	"os/exec"
	"testing"

	"github.com/REPPL/abcd-cli/internal/gittest"
)

func mustGitIdentity(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = gittest.Env(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// TestProbeIdentityIgnoresInjectedConfig proves an injected GIT_CONFIG_* cannot
// forge the caller's identity: ProbeIdentity feeds the hard_fail identity gate, so
// if a hostile sandbox/CI exports a fake user.email, the probe must still resolve
// the repo's real identity — otherwise the caller's genuine identity in scanned
// content escapes redaction. ScrubbedEnv strips the injection while keeping global
// config (so a real global identity is still probed).
func TestProbeIdentityIgnoresInjectedConfig(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	dir := t.TempDir()
	mustGitIdentity(t, dir, "init")
	mustGitIdentity(t, dir, "config", "user.email", "real@example.com")
	mustGitIdentity(t, dir, "config", "user.name", "Real Name")

	// A hostile/injected identity that would displace the real one, as GIT_CONFIG_*
	// parameters outrank repo-local config.
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "user.email")
	t.Setenv("GIT_CONFIG_VALUE_0", "evil@example.com")

	id := ProbeIdentity(dir)
	if id.GitUserEmail != "real@example.com" {
		t.Errorf("ProbeIdentity honoured an injected GIT_CONFIG_* identity: got %q, want real@example.com", id.GitUserEmail)
	}
}
