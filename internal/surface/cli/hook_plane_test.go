package cli

import (
	"encoding/json"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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
// metacharacter. The two spellings are the resolved-binary variable the wrapper
// builds and the plugin-root path SessionStart uses directly.
var binaryInvocationRe = regexp.MustCompile(`(?:"\$g"|"\$CLAUDE_PLUGIN_ROOT/abcd")((?:[[:space:]]+[a-z][a-z0-9-]*)+)`)

// TestHooksManifestNamesLiveSubverbs pins hooks/hooks.json against the live
// command tree. hooks.json ships with the plugin git clone while the binary is
// fetched from the latest release, so the two can skew and a rename is only
// visible once a user's session breaks. This turns that into a build failure in
// the same change that renames the verb — which is what makes "a hook sub-verb
// rename must carry an alias" enforceable rather than a note in a record.
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
	var commands []string
	for _, entries := range manifest.Hooks {
		for _, entry := range entries {
			for _, h := range entry.Hooks {
				commands = append(commands, h.Command)
			}
		}
	}
	if len(commands) == 0 {
		t.Fatal("the hooks manifest decoded to zero commands — the struct no longer matches its shape")
	}

	matches := binaryInvocationRe.FindAllStringSubmatch(strings.Join(commands, "\n"), -1)
	if len(matches) == 0 {
		t.Fatal("found no abcd invocations in hooks/hooks.json — the scan is broken, " +
			"not the manifest (it would silently pass forever)")
	}

	root := NewRootCommand()
	var seen []string
	for _, m := range matches {
		path := strings.Fields(m[1])
		joined := strings.Join(path, " ")
		if slices.Contains(seen, joined) {
			continue
		}
		seen = append(seen, joined)
		if cmd, _, err := root.Find(path); err != nil || cmd == nil || !slices.Equal(commandPath(cmd), path) {
			t.Errorf("hooks/hooks.json invokes `abcd %s`, which is not a command in this binary.\n"+
				"A hook sub-verb is a compatibility contract: the manifest ships with the plugin clone "+
				"while the binary comes from the latest release, so a rename without an alias breaks "+
				"live sessions against any skewed pair (iss-267).", joined)
		}
	}
	t.Logf("hook invocations checked: %v", seen)
}

// commandPath returns cmd's argv path below the root, so a Find that merely
// resolved a PREFIX (returning the parent, with the rest left as args) is not
// mistaken for a hit.
func commandPath(cmd *cobra.Command) []string {
	var path []string
	for c := cmd; c != nil && c.HasParent(); c = c.Parent() {
		path = append([]string{c.Name()}, path...)
	}
	return path
}
