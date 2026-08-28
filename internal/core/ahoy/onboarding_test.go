package ahoy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// onboardingArtefacts are the committed files the adopt phase reads and applies
// from. Every asset one of them names must resolve from inside this record or from
// the binary; a path outside both is a template only the machine that wrote it has.
var onboardingArtefacts = []string{
	filepath.Join("..", "..", "..", "commands", "prepare-this-repo.md"),
}

// machineLocalRefs returns the 1-based line numbers of every machine-local path
// reference in data: a home-relative `~/` path, or the private templates directory
// the adopt phase used to reach into.
//
// It is the detector iss-87 asked for, and it is deliberately blunt. A path under
// `~` is unavailable to every reader but the one who wrote it, so an adopt step
// that reaches for one degrades to nothing on a fresh clone — silently, because the
// step was written "if present". Blunt is the point: there is no legitimate `~`
// reference in an onboarding record, so there is no exception to reason about.
func machineLocalRefs(data []byte) []int {
	var hits []int
	for i, line := range strings.Split(string(data), "\n") {
		switch {
		case strings.Contains(line, "~/"),
			strings.Contains(line, "$HOME/"),
			strings.Contains(line, ".agents/templates"):
			hits = append(hits, i+1)
		}
	}
	return hits
}

// TestOnboardingIsSelfContained is itd-162's gate. The adopt phase reached for a
// pre-commit config and a prepare-commit-msg hook under a maintainer-local
// templates directory, both "if present" — so on any machine but that one the step
// did nothing at all, and the adoption silently degraded against loud-staging.
// Every asset it applies now resolves from this record or from the binary.
func TestOnboardingIsSelfContained(t *testing.T) {
	for _, rel := range onboardingArtefacts {
		data, err := os.ReadFile(rel)
		if err != nil {
			t.Fatalf("cannot read onboarding artefact %s: %v", rel, err)
		}
		if hits := machineLocalRefs(data); len(hits) > 0 {
			t.Errorf("%s references a machine-local path on line(s) %v; every adopt-phase asset must resolve "+
				"from this record or from the binary, never from a path only one machine has", rel, hits)
		}
	}
}

// TestMachineLocalRefsCatchesAReintroduction is the detector's own must-fail half.
// A gate nobody has watched fire is an assertion, not a check: without this,
// deleting the scan's body would leave the gate above green for ever.
func TestMachineLocalRefsCatchesAReintroduction(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{"tilde template path", "1. **Commit gates.** template at\n   `~/ABCDevelopment/.agents/templates/pre-commit-config.yaml`, if present\n"},
		{"HOME-relative path", "install the hook from `$HOME/.agents/templates/` if present\n"},
		{"the templates directory alone", "copy it out of the .agents/templates directory\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if hits := machineLocalRefs([]byte(tc.body)); len(hits) == 0 {
				t.Errorf("the self-containment scan missed a machine-local reference in:\n%s", tc.body)
			}
		})
	}
	// And the must-pass half: the shapes the rewritten record actually uses must not
	// trip it, or the gate would push the record back towards the `~` path it fixes.
	clean := "run `\"${CLAUDE_PLUGIN_ROOT}/abcd\" ahoy install --attribution`, then\n" +
		"point git at the committed hooks: `git config core.hooksPath .githooks`\n"
	if hits := machineLocalRefs([]byte(clean)); len(hits) > 0 {
		t.Errorf("the scan flags the binary-resolved form on line(s) %v; it must only catch machine-local paths", hits)
	}
}
