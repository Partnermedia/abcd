package guard

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"
)

// gitGlobalValueFlags is git's COMPLETE set of global options that consume the
// FOLLOWING separate token, enumerated from git 2.43's handle_options in git.c.
// Every git entry's ValueFlags must contain all of them: operands() steps over a
// listed flag AND its value to find the subcommand at operand 0, so an omitted
// one leaves its value sitting at operand 0, the subcommand check misses, and the
// entry never fires (gh-299).
var gitGlobalValueFlags = []string{
	"-C", "-c", "--exec-path", "--git-dir", "--work-tree",
	"--namespace", "--super-prefix", "--config-env", "--attr-source",
}

// checkGuard runs one candidate through the bundled registry, failing the test on
// an evaluation fault rather than letting it read as a verdict.
func checkGuard(t *testing.T, candidate string) Decision {
	t.Helper()
	d, err := Defaults().Check(candidate)
	if err != nil {
		t.Fatalf("Check(%q): %v", candidate, err)
	}
	return d
}

// TestGitEntriesKnowEveryGlobalValueFlag is the structural half. The bypass is an
// INCOMPLETE ENUMERATION, so the test that matters asserts completeness against
// the registry rather than re-checking the two spellings that happened to be
// missing — a list that drifts again fails here rather than in a session.
func TestGitEntriesKnowEveryGlobalValueFlag(t *testing.T) {
	raw, err := os.ReadFile("defaults/guard.json")
	if err != nil {
		t.Fatalf("reading the bundled registry: %v", err)
	}
	// `entries` is an OBJECT keyed by entry id, and the command and its
	// value-flag list live under `pattern`, not at entry level.
	var reg struct {
		Entries map[string]struct {
			Pattern struct {
				Command    string   `json:"command"`
				ValueFlags []string `json:"value_flags"`
			} `json:"pattern"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(raw, &reg); err != nil {
		t.Fatalf("the bundled registry is not valid JSON: %v", err)
	}
	var checked int
	for id, e := range reg.Entries {
		if e.Pattern.Command != "git" {
			continue
		}
		checked++
		for _, want := range gitGlobalValueFlags {
			if !slices.Contains(e.Pattern.ValueFlags, want) {
				t.Errorf("entry %q omits the git global value-flag %q from value_flags.\n"+
					"An unlisted flag that consumes the next token leaves its VALUE at operand 0, "+
					"so the subcommand check misses and this entry silently allows the hazard (gh-299).",
					id, want)
			}
		}
	}
	if checked == 0 {
		t.Fatal("found no git entries in the bundled registry — the scan is broken, not the registry")
	}
	t.Logf("git entries checked: %d", checked)
}

// TestGitGlobalValueFlagsDoNotHideAHazard is the behavioural half, run through the
// real bundled registry. Each prefix below is accepted by git and executes the
// hazard normally, so the guard's parse must not diverge from what runs.
func TestGitGlobalValueFlagsDoNotHideAHazard(t *testing.T) {
	hazards := []string{
		"push --force origin main",
		"commit --no-verify -m x",
		"reset --hard",
	}
	// Both spellings of every value-flag: the separate-token form is the one that
	// evaded (the glued --opt=value form was already skipped as an ordinary flag),
	// and both must reach the same verdict as the bare command.
	prefixes := []string{
		"", "-C /tmp ", "-c user.name=x ",
		"--attr-source HEAD ", "--attr-source=HEAD ",
		"--config-env NAME=ENVV ", "--config-env=NAME=ENVV ",
		"--git-dir /tmp/.git ", "--work-tree /tmp ",
		"--namespace ns ", "--exec-path /usr/lib/git-core ", "--super-prefix p/ ",
	}
	for _, hazard := range hazards {
		bare := checkGuard(t, "git "+hazard)
		if bare.Verdict == VerdictAllow {
			t.Fatalf("`git %s` is not matched at all by the bundled registry — the fixture is wrong, "+
				"not the guard", hazard)
		}
		for _, prefix := range prefixes {
			candidate := "git " + prefix + hazard
			got := checkGuard(t, candidate)
			if got.Verdict != bare.Verdict {
				t.Errorf("%q\n  got verdict %q, want %q (same as the bare command).\n"+
					"  A global option that consumes the next token must not displace the subcommand (gh-299).",
					candidate, got.Verdict, bare.Verdict)
			}
		}
	}
}

// TestGitGlobalValueFlagsStillAllowBenignCommands guards the other direction: the
// widened list must not start blocking commands that were never hazards. A
// value-flag whose VALUE happens to read like a hazardous subcommand is the shape
// that would break if the flag were skipped without its argument.
func TestGitGlobalValueFlagsStillAllowBenignCommands(t *testing.T) {
	for _, candidate := range []string{
		"git --attr-source HEAD status",
		"git --attr-source push log",
		"git --config-env NAME=ENVV status",
		"git -c user.name=push status",
		"git status",
	} {
		if got := checkGuard(t, candidate); got.Verdict != VerdictAllow {
			t.Errorf("%q got verdict %q, want allow — the widened value-flag list must not "+
				"invent a hazard", candidate, got.Verdict)
		}
	}
}

// TestGitGlobalValueFlagsAreDocumented keeps the enumeration honest: the list in
// this file is the claim, and a reader must be able to see where it came from.
func TestGitGlobalValueFlagsAreDocumented(t *testing.T) {
	for _, f := range gitGlobalValueFlags {
		if !strings.HasPrefix(f, "-") {
			t.Errorf("%q is not a flag spelling", f)
		}
	}
	if len(gitGlobalValueFlags) != 9 {
		t.Errorf("the git global value-flag set has %d entries, expected 9 as enumerated from "+
			"git 2.43's handle_options; if git gained one, add it here AND to every git entry",
			len(gitGlobalValueFlags))
	}
}
