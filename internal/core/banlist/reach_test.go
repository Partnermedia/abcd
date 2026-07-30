package banlist

import (
	"strings"
	"testing"
)

// TestPrivateReachNoteNamesWhatTheGuardCannotSee is AC7's honesty clause held to
// the whole truth. "Machines that have opted in" is necessary and not sufficient:
// git runs no pre-commit hook for a rebase, `git am`, or a cherry-pick, and
// `--no-verify` turns it off outright. A reader who takes the sentence at face
// value must not come away believing an opted-in machine is fully covered.
func TestPrivateReachNoteNamesWhatTheGuardCannotSee(t *testing.T) {
	for _, want := range []string{"CI cannot enforce", "opted in", "rebase", "no-verify"} {
		if !strings.Contains(PrivateReachNote, want) {
			t.Errorf("the reach note omits %q: %q", want, PrivateReachNote)
		}
	}
}
