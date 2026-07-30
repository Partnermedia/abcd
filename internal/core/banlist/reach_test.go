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
	// `git pull` fast-forwarding a branch that carries a banned name is the commonest
	// uncovered path of all — no commit is created, so no hook is asked — and it must
	// be named. A stash is NOT in the list: `git stash pop` followed by a commit runs
	// pre-commit exactly as any other commit does, and a bypass list that includes a
	// path which IS covered teaches the reader to distrust the rest of it.
	for _, want := range []string{"CI cannot enforce", "opted in", "e.g.", "rebase", "fast-forward", "no-verify"} {
		if !strings.Contains(PrivateReachNote, want) {
			t.Errorf("the reach note omits %q: %q", want, PrivateReachNote)
		}
	}
	if strings.Contains(PrivateReachNote, "stash") {
		t.Errorf("the reach note names a path that IS covered: %q", PrivateReachNote)
	}
}
