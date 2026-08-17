package cli

import (
	"encoding/json"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// hookPlaneParents are the parents a host hook reaches. On the hook plane an exit
// status is not a diagnostic, it is an INSTRUCTION: the host reads 2 as "block
// this action". `guard hook` runs on PreToolUse before every shell command, and
// `hook prompt-router` runs on UserPromptSubmit before every prompt.
var hookPlaneParents = []string{"guard", "hook"}

// TestHookPlaneParentsFailOpenOnUnknownSubverb is iss-267. iss-266 made every
// parent refuse an unknown sub-verb at cobra's usage status, exit 2 — correct for
// a terminal, wrong here. `guard hook`'s contract (spc-16, itd-103 AC 1) is
// fail-open-loud: exit 2 is reserved for "the guard decided to block", and every
// path that is NOT a decision exits 1 so the command still runs and the warning
// is still seen. An unknown sub-verb is not a decision, so refusing it at 2 makes
// abcd claim a hazard verdict it never reached, and blocks every shell command in
// the session. The PreToolUse wrapper cannot catch this: it treats 2 as a
// recognised code, so its "FAILED TO RUN … UNGUARDED" net never fires.
//
// The refusal must still be LOUD and non-zero — iss-266's guarantee that a
// mistyped sub-verb never reads as success is unchanged. Only the code moves,
// from the host's blocking status to its non-blocking one.
func TestHookPlaneParentsFailOpenOnUnknownSubverb(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, parent := range hookPlaneParents {
		t.Run(parent, func(t *testing.T) {
			var out, errOut strings.Builder
			code := Run([]string{parent, "zzz-not-a-subverb"}, &out, &errOut)
			if code == 0 {
				t.Fatalf("abcd %s zzz-not-a-subverb exited 0; an unknown sub-verb must still refuse", parent)
			}
			if code == 2 {
				t.Fatalf("abcd %s zzz-not-a-subverb exited 2 — the host's BLOCKING status. "+
					"An unknown sub-verb is not a decision to block; it must fail open at exit 1 (iss-267).\nstderr:\n%s",
					parent, errOut.String())
			}
			if code != 1 {
				t.Fatalf("abcd %s zzz-not-a-subverb exited %d, want 1 (loud, non-blocking)", parent, code)
			}
			if errOut.Len() == 0 {
				t.Fatalf("abcd %s zzz-not-a-subverb said nothing on stderr; failing open silently is the "+
					"other half of the contract this must not break", parent)
			}
		})
	}
}

// binaryInvocationRe finds each place hooks.json runs the abcd binary and
// captures the sub-verb path that follows, stopping at the first shell
// metacharacter. The alternatives are the three spellings the wrappers use: the
// resolved-binary variable, the plugin-root variable, and the full plugin-root
// path SessionStart uses directly.
var binaryInvocationRe = regexp.MustCompile(`(?:"\$g"|"\$r/abcd"|"\$CLAUDE_PLUGIN_ROOT/abcd")((?:[[:space:]]+[a-z][a-z0-9-]*)+)`)

// TestHooksManifestNamesLiveSubverbs pins hooks/hooks.json against the live
// command tree.
//
// The skew this defends runs in ONE direction: a user's plugin clone updates by
// `git pull` while the binary updates by release download, so the manifest is the
// half that runs AHEAD. That decides the doctrine. An alias added to a new binary
// does nothing for a user whose binary is old, so the remedy is not "rename and
// alias the manifest across" — it is the reverse: **the manifest's spellings are
// frozen, and a rename is absorbed by an alias in the binary**. Editing
// hooks.json to a new spelling is the move that breaks every older binary.
//
// So this test asks only that every spelling the manifest names still resolves in
// this binary — through an alias just as happily as through a primary name. It
// fails the build in the same change that would strand a skewed pair, instead of
// in a user's session.
func TestHooksManifestNamesLiveSubverbs(t *testing.T) {
	raw, err := os.ReadFile("../../../hooks/hooks.json")
	if err != nil {
		t.Fatalf("reading the hooks manifest: %v", err)
	}
	// Decode rather than scan the raw bytes: the command is a JSON string, so the
	// shell quoting is escaped on disk and a regex over the file would match the
	// escaped spelling instead of the command the host actually runs.
	var manifest struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("the hooks manifest is not valid JSON: %v", err)
	}
	var events []string
	var commands []string
	for event, entries := range manifest.Hooks {
		for _, entry := range entries {
			for _, h := range entry.Hooks {
				events = append(events, event)
				commands = append(commands, h.Command)
			}
		}
	}
	if len(commands) == 0 {
		t.Fatal("the hooks manifest decoded to zero commands — the struct no longer matches its shape")
	}

	root := NewRootCommand()
	var seen []string
	// Scan EACH command separately and require every one to yield an invocation.
	// Joining them and checking the total only catches a scan that fails
	// completely; a quoting drift in one command would hide that one invocation
	// while the others kept the test green, leaving it unpinned forever — which is
	// the silent-pass this test exists to rule out.
	for i, command := range commands {
		matches := binaryInvocationRe.FindAllStringSubmatch(command, -1)
		if len(matches) == 0 {
			t.Errorf("the %s hook command contains no recognisable abcd invocation — "+
				"the scan is broken, not the manifest. Every hook command runs the binary; if this one "+
				"spells the invocation a new way, teach binaryInvocationRe that spelling rather than "+
				"leaving the command unchecked.", events[i])
			continue
		}
		for _, m := range matches {
			path := strings.Fields(m[1])
			joined := strings.Join(path, " ")
			if slices.Contains(seen, joined) {
				continue
			}
			seen = append(seen, joined)
			// Fully-consumed path, NOT a name comparison. Find resolves aliases but
			// Name() returns the primary, so comparing names would reject a rename
			// that carries a compatibility alias — the one remedy that survives the
			// skew — and the only way back to green would be editing the manifest,
			// the very change that stranding depends on. Leftover args mean Find
			// stopped at a parent, which is the prefix case this must still catch.
			cmd, rest, err := root.Find(path)
			if err != nil || cmd == nil || cmd == root || len(rest) > 0 {
				t.Errorf("hooks/hooks.json (%s) invokes `abcd %s`, which this binary does not answer.\n"+
					"The manifest runs ahead of the binary on a user's machine (clone pulls, binary lags), "+
					"so its spellings are frozen: absorb the rename with an Aliases entry on the renamed "+
					"command and leave hooks.json alone — editing it to the new spelling is what strands "+
					"every older binary (iss-267).", events[i], joined)
			}
		}
	}
	t.Logf("hook invocations checked: %v", seen)
}
