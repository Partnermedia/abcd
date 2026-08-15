package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core"
)

// TestSessionStartReportsVersionTransition proves AC6: when the running binary's
// version differs from the version recorded when the repo was last set up, the
// session-start hook reports the transition. The comparison is disk-only.
func TestSessionStartReportsVersionTransition(t *testing.T) {
	orig := core.Version
	core.Version = "v9.9.9"
	t.Cleanup(func() { core.Version = orig })

	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".abcd", "config.json"),
		[]byte(`{"meta":{"setup_version":"v1.0.0"}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// The hook exits non-zero when it emits notices (so SessionStart shows them);
	// the notice text is on the captured stream regardless.
	out, _ := runCLIStdinErr(t, `{"cwd":"`+repo+`"}`, "hook", "session-start")
	s := string(out)
	if !strings.Contains(s, "v9.9.9") || !strings.Contains(s, "v1.0.0") {
		t.Fatalf("transition notice missing running/recorded versions:\n%s", s)
	}
}

// TestSessionStartSilentWhenVersionMatches proves the report does not fire when
// the recorded and running versions agree.
func TestSessionStartSilentWhenVersionMatches(t *testing.T) {
	orig := core.Version
	core.Version = "v9.9.9"
	t.Cleanup(func() { core.Version = orig })

	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".abcd", "config.json"),
		[]byte(`{"meta":{"setup_version":"v9.9.9"}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, _ := runCLIStdinErr(t, `{"cwd":"`+repo+`"}`, "hook", "session-start")
	if strings.Contains(string(out), "last set up with") {
		t.Fatalf("transition reported when versions match:\n%s", out)
	}
}
