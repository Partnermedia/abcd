package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestEveryParentIsRunnable is the iss-266 sweep, stated structurally. A cobra
// parent with no RunE is not Runnable, so cobra prints help and exits 0 WITHOUT
// running the command's Args validator: `abcd docs nonsense` then reads as
// SUCCESS to a script, which is exactly how a typo or a retired sub-verb spelling
// slips a CI gate. spc-30 fixed the disembark parent by giving it a RunE that
// prints help — bare invocation still shows help, and cobra.NoArgs now refuses a
// stray token. The property is asserted over the WHOLE tree rather than a
// hand-kept list, so a parent added later cannot reintroduce the hole silently.
func TestEveryParentIsRunnable(t *testing.T) {
	var offenders []string
	for _, p := range parents(t, NewRootCommand()) {
		if p.cmd.RunE == nil && p.cmd.Run == nil {
			offenders = append(offenders, strings.Join(p.path, " "))
		}
	}
	if len(offenders) > 0 {
		t.Fatalf("parents with no RunE print help and exit 0 on an unknown sub-verb (iss-266): %v\n"+
			"give each a `RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() }`",
			offenders)
	}
}

// TestReadOnlyParentsRefuseAnUnknownSubverb is the behavioural half: the
// structural check above guarantees the Args validator RUNS, and this proves the
// validator actually refuses. `capture` and `intent` are deliberately absent —
// their positional is free text by design (`abcd capture "…"` files an issue), so
// an unrecognised token is prose, not a typo, and they carry their own
// suspected-typo guard instead. Runs in a scratch cwd so a regression that DOES
// write a record cannot dirty the repository under test.
func TestReadOnlyParentsRefuseAnUnknownSubverb(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, parent := range []string{"docs", "docs cite", "guard", "disembark", "embark", "spec", "history", "memory", "launch", "banlist", "ahoy", "hook", "ideate"} {
		t.Run(strings.ReplaceAll(parent, " ", "_"), func(t *testing.T) {
			args := append(strings.Fields(parent), "zzz-not-a-subverb")
			var out, errOut strings.Builder
			if code := Run(args, &out, &errOut); code == 0 {
				t.Fatalf("abcd %s exited 0; an unknown sub-verb must refuse.\nstdout:\n%s\nstderr:\n%s",
					strings.Join(args, " "), out.String(), errOut.String())
			}
		})
	}
}

type parentCmd struct {
	path []string
	cmd  *cobra.Command
}

// parents walks the command tree and returns every command that has
// sub-commands, with the argv path that reaches it. cobra's own generated `help`
// and `completion` commands are skipped: they are not part of abcd's surface.
func parents(t *testing.T, root *cobra.Command) []parentCmd {
	t.Helper()
	var found []parentCmd
	var walk func(cmd *cobra.Command, prefix []string)
	walk = func(cmd *cobra.Command, prefix []string) {
		for _, sub := range cmd.Commands() {
			if sub.Name() == "help" || sub.Name() == "completion" {
				continue
			}
			path := append(append([]string{}, prefix...), sub.Name())
			if len(sub.Commands()) > 0 {
				found = append(found, parentCmd{path: path, cmd: sub})
				walk(sub, path)
			}
		}
	}
	walk(root, nil)
	if len(found) == 0 {
		t.Fatal("walked the command tree and found no parents — the walk is broken")
	}
	return found
}
