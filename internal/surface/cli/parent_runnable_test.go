package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// freeTextParents are the parents whose positional is prose by design: `abcd
// capture "…"` files an issue and `abcd intent "…"` files a draft, so an
// unrecognised token is text, not a typo. They are cobra.ArbitraryArgs on
// purpose and carry their own suspectedTypoedSubcommand guard instead, which has
// its own tests. Every other parent must refuse.
var freeTextParents = []string{"capture", "intent"}

// TestEveryParentRefusesAnUnknownSubverb is the iss-266 sweep, stated
// structurally. Two independent cobra facts have to hold for a parent to refuse a
// mistyped sub-verb, and missing EITHER one silently exits 0:
//
//   - The parent must be Runnable. A parent with no RunE returns flag.ErrHelp
//     before ValidateArgs ever runs, so cobra prints help and exits 0 without
//     looking at the args at all.
//   - The parent must declare an Args validator. With Args nil, cobra's
//     legacyArgs falls through to ArbitraryArgs for a non-root parent, so a
//     Runnable parent still accepts the stray token and exits 0.
//
// The property is asserted over the whole live tree rather than a hand-kept list,
// so a parent added later cannot reintroduce the hole by missing either half.
func TestEveryParentRefusesAnUnknownSubverb(t *testing.T) {
	var notRunnable, noArgsValidator []string
	for _, p := range parents(t, NewRootCommand()) {
		name := strings.Join(p.path, " ")
		if p.cmd.RunE == nil && p.cmd.Run == nil {
			notRunnable = append(notRunnable, name)
		}
		if p.cmd.Args == nil {
			noArgsValidator = append(noArgsValidator, name)
		}
	}
	if len(notRunnable) > 0 {
		t.Errorf("parents with no RunE print help and exit 0 on an unknown sub-verb (iss-266): %v\n"+
			"give each a `RunE: helpRunE`", notRunnable)
	}
	if len(noArgsValidator) > 0 {
		t.Errorf("parents with no Args validator accept an unknown sub-verb and exit 0 (iss-266): %v\n"+
			"give each a `cobra.NoArgs` (or an explicit validator that refuses a stray token)", noArgsValidator)
	}
}

// TestParentsRefuseAnUnknownSubverbAtExitTwo is the behavioural half. The
// structural check above guarantees the Args validator RUNS; this proves it
// actually refuses, and refuses with the exit code the CHANGELOG promises — a
// parent that regressed to exit 1 would be a different, worse contract. The list
// is DERIVED from the live tree minus freeTextParents, so a parent added later is
// exercised without anyone remembering to add it. Runs in a scratch cwd so a
// regression that does reach a write path cannot dirty the repository under test.
func TestParentsRefuseAnUnknownSubverbAtExitTwo(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, p := range parents(t, NewRootCommand()) {
		name := strings.Join(p.path, " ")
		if slices.Contains(freeTextParents, name) {
			continue
		}
		t.Run(strings.Join(p.path, "_"), func(t *testing.T) {
			args := append(append([]string{}, p.path...), "zzz-not-a-subverb")
			var out, errOut strings.Builder
			if code := Run(args, &out, &errOut); code != 2 {
				t.Fatalf("abcd %s exited %d, want 2 (usage refusal).\nstdout:\n%s\nstderr:\n%s",
					strings.Join(args, " "), code, out.String(), errOut.String())
			}
		})
	}
}

// TestFreeTextParentsAreStillExempt pins the exemption list to reality. A name
// left in freeTextParents after that verb stopped taking free text would silently
// drop a parent out of the sweep above — the exemption has to keep earning itself.
func TestFreeTextParentsAreStillExempt(t *testing.T) {
	byName := map[string]*cobra.Command{}
	for _, p := range parents(t, NewRootCommand()) {
		byName[strings.Join(p.path, " ")] = p.cmd
	}
	for _, name := range freeTextParents {
		cmd, ok := byName[name]
		if !ok {
			t.Errorf("freeTextParents names %q, which is not a parent in the command tree", name)
			continue
		}
		if err := cmd.Args(cmd, []string{"some prose"}); err != nil {
			t.Errorf("%q is exempt as a free-text verb but its Args validator refuses prose: %v", name, err)
		}
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
