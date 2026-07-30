package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/banlist"
)

// TestAhoyStatusStatesThePrivateBanlistReach is spc-20 AC7 on the status board.
// `abcd ahoy` now reports the two-layer name guard, and a reader who sees "hook
// installed" beside a list of private keys would reasonably assume a pull request
// is covered. It is not, and it cannot be: a pattern in CI config is a published
// pattern. The status surface says so in the same words the verb does.
func TestAhoyStatusStatesThePrivateBanlistReach(t *testing.T) {
	hermeticEnv(t)
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)

	out := string(runCLI(t, "ahoy"))
	if !strings.Contains(out, "banlist:") {
		t.Fatalf("bare ahoy does not report the name-banlist layers:\n%s", out)
	}
	if !strings.Contains(out, banlist.PrivateReachNote) {
		t.Errorf("the status board does not state the private layer's local-only reach (%q):\n%s",
			banlist.PrivateReachNote, out)
	}
}

// TestAhoyStatusReportsEachScaffoldedArtefact keeps the line useful as well as
// honest: a reader has to be able to tell which of the three artefacts is missing,
// and whether this machine has opted into the private layer at all.
func TestAhoyStatusReportsEachScaffoldedArtefact(t *testing.T) {
	hermeticEnv(t)
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)

	bare := string(runCLI(t, "ahoy"))
	if !strings.Contains(bare, "INACTIVE") {
		t.Errorf("an absent private store must read as inactive, never as silence:\n%s", bare)
	}

	if _, err := runCLIErr(t, "ahoy", "install", "--yes", "--adopt",
		"--visibility", "private", "--docs-target", "both",
		"--oracle-backend", "host-delegated", "--scan-deep", "false"); err != nil {
		t.Fatalf("install: %v", err)
	}
	after := string(runCLI(t, "ahoy"))
	for _, want := range []string{"hook installed", "public family", "private store"} {
		if !strings.Contains(after, want) {
			t.Errorf("the banlist line omits %q after scaffolding:\n%s", want, after)
		}
	}
	if strings.Contains(after, "INACTIVE") {
		t.Errorf("the private store is scaffolded, so the line must not report it inactive:\n%s", after)
	}
}

// TestAhoyEnvelopeCarriesTheReach: the JSON envelope is a report surface too, and a
// machine consumer that reads the booleans without the caveat would draw exactly the
// wrong conclusion. The reach travels with the state it qualifies.
func TestAhoyEnvelopeCarriesTheReach(t *testing.T) {
	hermeticEnv(t)
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)

	var env struct {
		Banlist struct {
			Reach string `json:"reach"`
		} `json:"banlist"`
	}
	if err := json.Unmarshal(runCLI(t, "ahoy", "dry-run"), &env); err != nil {
		t.Fatalf("dry-run envelope does not parse: %v", err)
	}
	if env.Banlist.Reach != banlist.PrivateReachNote {
		t.Errorf("envelope reach = %q; want %q", env.Banlist.Reach, banlist.PrivateReachNote)
	}
}
