package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// guardScopeClaim is the scope statement adr-42 requires the guard to make about
// itself. The guard is a MISTAKE FILTER: it catches accidents and casual evasion
// by a cooperating agent, and it does not withstand an author trying to get a
// command past it. A parse layer cannot be a security boundary — three separate
// incomplete-enumeration defects (gh-297, gh-299, iss-272) are the evidence — so
// the surfaces say so rather than letting a reader infer a boundary from a
// confident-looking verdict.
// The comparison is case-insensitive because the surfaces emphasise differently
// and correctly: terminal help has no bold, so it shouts (MISTAKE FILTER), while
// markdown uses `**`. Pinning the exact casing would force one convention onto
// both and buy nothing — the words are the claim.
const guardScopeClaim = "mistake filter, not a security boundary"

// guardScopeSuccessor is the other half: naming a parse layer's limits is only
// honest if the reader is told where the enforcing control lives instead.
const guardScopeSuccessor = "execution layer"

// guardScopeSurfaces are every place a reader meets the guard's own account of
// itself. The statement is worthless on one surface and absent from the next: a
// facilitator reads the plugin command, an operator reads `--help`, and a
// maintainer reads the brief. The list is asserted here so a future surface
// cannot quietly ship without it.
var guardScopeSurfaces = []string{
	filepath.Join("..", "..", "..", "commands", "guard.md"),
	filepath.Join("..", "..", "..", ".abcd", "development", "brief", "04-surfaces", "17-guard.md"),
	filepath.Join("..", "..", "..", "docs", "reference", "cli", "commands.md"),
}

// TestGuardCheckHelpStatesItsScope pins the statement on the CLI surface, read
// from the live command tree rather than from the generated page, so the claim
// cannot survive as documentation after the help text drops it.
func TestGuardCheckHelpStatesItsScope(t *testing.T) {
	root := NewRootCommand()
	check, _, err := root.Find([]string{"guard", "check"})
	if err != nil {
		t.Fatalf("guard check is not reachable from the command tree: %v", err)
	}
	if check.Name() != "check" {
		t.Fatalf("resolved %q, not the guard check verb", check.Name())
	}
	for _, want := range []string{guardScopeClaim, guardScopeSuccessor} {
		if !strings.Contains(strings.ToLower(check.Long), want) {
			t.Errorf("guard check --help does not state its scope: missing %q\n"+
				"adr-42 decision 1: the parse layer says what it is, on its own surfaces", want)
		}
	}
}

// TestEveryGuardSurfaceStatesItsScope holds the same line across the plugin
// command, the surface brief, and the generated reference. A scope statement a
// reader can miss by choosing a different door is not a scope statement.
func TestEveryGuardSurfaceStatesItsScope(t *testing.T) {
	for _, path := range guardScopeSurfaces {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("cannot read guard surface %s: %v", path, err)
		}
		if !strings.Contains(strings.ToLower(string(body)), guardScopeClaim) {
			t.Errorf("%s does not state the guard's scope: missing %q\n"+
				"adr-42 decision 1 and part D: every guard surface carries it", path, guardScopeClaim)
		}
	}
}
