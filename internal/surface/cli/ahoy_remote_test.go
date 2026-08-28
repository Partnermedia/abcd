package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAhoyRemoteRefusesAnUnmanagedFolderFromTheCLI is the front-door half of
// itd-153's trust boundary. The verb is the one thing in abcd that mutates state
// outside the machine, so the CLI must carry the refusal all the way out: a
// non-zero exit for the apply, and the reason on stdout rather than only in a
// result nobody rendered.
func TestAhoyRemoteRefusesAnUnmanagedFolderFromTheCLI(t *testing.T) {
	hermeticEnv(t)
	dir := t.TempDir()
	t.Chdir(dir)

	// The read is reachable and reports the refusal without exiting non-zero: it is
	// abcd's bare-render convention, and looking is never an error.
	out := string(runCLI(t, "ahoy", "remote"))
	if !strings.Contains(out, "refused") {
		t.Fatalf("`ahoy remote` did not refuse a folder abcd does not manage:\n%s", out)
	}
	if !strings.Contains(out, "note:") {
		t.Errorf("the refusal carries no reason on stdout:\n%s", out)
	}

	applied, err := runCLIErr(t, "ahoy", "remote", "apply")
	if err == nil {
		t.Fatalf("`ahoy remote apply` exited zero on a refusal; a write that did not happen must not look like one that did:\n%s", applied)
	}
	if !strings.Contains(string(applied), "refused") {
		t.Errorf("the apply's refusal is not rendered:\n%s", applied)
	}
	// Nothing was written, least of all the desired-state mirror.
	if _, statErr := os.Stat(filepath.Join(dir, ".abcd", "work", "rulesets", "repo-settings.json")); statErr == nil {
		t.Error("a refused apply wrote the desired-state mirror")
	}
}
