package guard

import (
	"encoding/json"
	"os"
	"os/exec"

	"github.com/REPPL/abcd-cli/internal/gittest"
	"slices"
	"strings"
	"testing"
)

// gitGlobalValueFlags are the git global options after which git does NOT read
// the next token as the subcommand, so the operand walk must step over that token
// to find the subcommand at operand 0. An omitted one leaves the flag's VALUE
// sitting at operand 0, the subcommand check misses, and the entry never fires —
// which is the whole of gh-299.
//
// DERIVED FROM GIT, NOT FROM A DOCUMENT. The first cut of this list was taken
// from the bug report and was wrong three ways: it omitted --shallow-file (a live
// force-push bypass, in git since 1.9), and counted --exec-path and --super-prefix
// as value-taking when neither is. It totalled nine either way, so a size
// assertion certified the wrong list as complete.
var gitGlobalValueFlags = []string{
	"-C", "-c", "--git-dir", "--work-tree", "--namespace",
	"--config-env", "--attr-source", "--shallow-file",
}

// gitNonConsumingButListed are spellings the registry steps over even though git
// 2.43 takes no value after them. Stepping over a non-value flag's neighbour can
// only over-block, and both are retained deliberately: --exec-path prints its
// path and exits, so no subcommand runs either way; --super-prefix DID take a
// value on git <= 2.42, where a repo may still be. Named here so the difference
// is a recorded decision rather than an unexplained entry.
var gitNonConsumingButListed = []string{"--exec-path", "--super-prefix"}

// gitProbeCandidates is deliberately BROADER than the registry: the point is to
// catch a value-taking global that nobody listed, so it must contain names the
// registry does not.
//
// It is hand-maintained, and that limitation is worth stating rather than hiding.
// `git --help` does not mention --attr-source, --shallow-file or --super-prefix
// at all — the three that caused or complicated this defect — so deriving the
// candidate set from git's own output would have missed exactly the flags that
// matter. Exhaustiveness against git is not reachable from here; what this buys
// is that every option anyone has thought of is CLASSIFIED against the installed
// git rather than asserted.
var gitProbeCandidates = []string{
	// value-taking, documented in `git --help`
	"-C", "-c", "--git-dir", "--work-tree", "--namespace", "--config-env",
	// value-taking or value-taking-historically, NOT documented there
	"--attr-source", "--shallow-file", "--super-prefix",
	// early-exit or no value
	"--exec-path", "--html-path", "--man-path", "--info-path",
	"--paginate", "--no-pager", "--bare", "--no-replace-objects",
	"--literal-pathspecs", "--glob-pathspecs", "--noglob-pathspecs",
	"--icase-pathspecs", "--no-optional-locks",
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

// gitEntryValueFlags decodes defaults/guard.json and returns each git entry's
// value_flags, keyed by entry id.
//
// `entries` is an OBJECT keyed by entry id, and the command and its value-flag
// list live under `pattern`, not at entry level. An earlier draft used the wrong
// shape, decoded nothing, and passed — which is why callers guard on an empty
// result.
func gitEntryValueFlags(t *testing.T) map[string][]string {
	t.Helper()
	raw, err := os.ReadFile("defaults/guard.json")
	if err != nil {
		t.Fatalf("reading the bundled registry: %v", err)
	}
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
	out := map[string][]string{}
	for id, e := range reg.Entries {
		if e.Pattern.Command == "git" {
			out[id] = e.Pattern.ValueFlags
		}
	}
	if len(out) == 0 {
		t.Fatal("found no git entries in the bundled registry — the scan is broken, not the registry")
	}
	return out
}

// registryGitValueFlags is every flag the SHIPPED git entries step over, read
// from the registry rather than restated here.
//
// Reading the file is the whole point. An earlier version returned this test's
// own constants, so the "wrongly listed" direction could never fire against
// defaults/guard.json: adding one non-consuming global to the six entries left
// the suite green while producing a live force-push bypass, because the walk then
// steps over the real subcommand. The registry is the thing under test, so it has
// to be the thing read.
func registryGitValueFlags(t *testing.T) []string {
	t.Helper()
	byEntry := gitEntryValueFlags(t)
	// The six lists are hand-duplicated in the JSON, which is how gh-299 happened
	// in the first place. Assert they agree, so a partial edit is a failure rather
	// than a silent divergence between entries.
	var ids []string
	for id := range byEntry {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	first := byEntry[ids[0]]
	for _, id := range ids[1:] {
		if !slices.Equal(byEntry[id], first) {
			t.Errorf("git entries %q and %q carry different value_flags:\n  %v\n  %v\n"+
				"The six lists are duplicated by hand, so a partial edit leaves the others "+
				"bypassable — the exact shape of gh-299.", ids[0], id, first, byEntry[id])
		}
	}
	return append([]string{}, first...)
}

// TestGitEntriesKnowEveryGlobalValueFlag is the structural half. The bypass is an
// INCOMPLETE ENUMERATION, so the test asserts completeness against the registry
// rather than re-checking the spellings that happened to be missing.
func TestGitEntriesKnowEveryGlobalValueFlag(t *testing.T) {
	raw, err := os.ReadFile("defaults/guard.json")
	if err != nil {
		t.Fatalf("reading the bundled registry: %v", err)
	}
	// `entries` is an OBJECT keyed by entry id, and the command and its value-flag
	// list live under `pattern`, not at entry level. An earlier draft used the
	// wrong shape, decoded nothing, and passed — which is why the "found no git
	// entries" guard below exists.
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
					"An unlisted flag that takes a value leaves that VALUE at operand 0, so the "+
					"subcommand check misses and this entry silently allows the hazard (gh-299).",
					id, want)
			}
		}
	}
	if checked == 0 {
		t.Fatal("found no git entries in the bundled registry — the scan is broken, not the registry")
	}
	t.Logf("git entries checked: %d", checked)
}

// TestGitGlobalValueFlagsMatchThisGit re-derives the classification from the git
// actually installed, so the list is a CHECKABLE claim rather than an asserted
// one. The previous shape — a hard-coded size plus a superset-tolerant membership
// check — could never notice a missing member, and duly certified a list holding
// a live bypass as "complete against git 2.43's handle_options".
//
// The question the guard needs answered is not "does this flag take a value" but
// "does git read the NEXT token as the subcommand". Those differ: --exec-path
// takes no value, it prints and exits, so git runs no subcommand either way and
// stepping over the token is harmless. Only a flag after which git really does
// dispatch on the next token must be absent from the walk.
//
// Probe: `git <opt> ZZZMARK status`. If git reports ZZZMARK as an unknown git
// command it dispatched on it; anything else means the token was swallowed or git
// exited first.
func TestGitGlobalValueFlagsMatchThisGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed; the classification cannot be re-derived here")
	}
	// A REAL repository, because the three behaviours are only distinguishable by
	// whether the subcommand actually ran.
	// Every git call here routes through gittest.Env so the probe reads the
	// installed git's OPTION HANDLING and never a developer's global config
	// (iss-28 hermeticity).
	env := gittest.Env(t)
	repo := t.TempDir()
	initCmd := exec.Command("git", "-C", repo, "init", "-q")
	initCmd.Env = env
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Skipf("cannot init a probe repository: %v: %s", err, out)
	}

	const (
		consumes   = "consumes"   // git swallowed the token, then ran the subcommand
		harmless   = "harmless"   // git printed a path or refused the option; nothing ran
		dispatches = "dispatches" // git read the token AS the subcommand
	)
	// CONSUMES IS THE DEFAULT, deliberately. This is a security control, so a flag
	// whose behaviour the probe cannot recognise must be treated as taking a value
	// — that demands it be listed, which over-blocks at worst. Defaulting the other
	// way would let an unrecognised flag be silently ignored, which is the exact
	// shape of the defect this test exists to catch. Only the two shapes that
	// provably run nothing are excused.
	classify := func(opt string) (string, string) {
		cmd := exec.Command("git", opt, "ZZZMARK", "status")
		cmd.Dir = repo
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		text := strings.TrimSpace(string(out))
		switch {
		case strings.Contains(text, "ZZZMARK' is not a git command"),
			strings.Contains(text, "ZZZMARK is not a git command"):
			return dispatches, text
		case err == nil && !strings.ContainsAny(text, " \n") && strings.HasPrefix(text, "/"):
			// A bare path on a clean exit: --exec-path, --html-path and friends
			// print and stop, so no subcommand runs whether or not the walk steps
			// over the token.
			return harmless, text
		case strings.Contains(text, "unknown option"):
			// git refused the option outright; nothing ran.
			return harmless, text
		default:
			return consumes, text
		}
	}

	listed := registryGitValueFlags(t)
	var sawConsuming, sawDispatching bool
	for _, opt := range gitProbeCandidates {
		kind, text := classify(opt)
		switch kind {
		case consumes:
			sawConsuming = true
			if !slices.Contains(listed, opt) {
				t.Errorf("git CONSUMES the token after %q and then runs the subcommand, yet no git "+
					"entry steps over it — a bypass of exactly the gh-299 shape: the flag's value "+
					"lands at operand 0 and every git entry misses.\n  probe: %s", opt, text)
			}
		case dispatches:
			sawDispatching = true
			if slices.Contains(listed, opt) {
				t.Errorf("git DISPATCHES on the token after %q, yet the registry steps over it — "+
					"the walk would skip a real subcommand.\n  probe: %s", opt, text)
			}
		case harmless:
			// git printed and exited, so no subcommand runs whether or not the walk
			// steps over the token. Listing is harmless; not listing is harmless.
			// Recorded rather than asserted, since either choice is safe.
			t.Logf("%q runs nothing after the token (listed=%v) — harmless either way",
				opt, slices.Contains(listed, opt))
		}
	}
	// Both poles must be observed, or the probe proves nothing: one classifies
	// everything as swallowed and agrees with any list, the other the reverse.
	if !sawConsuming || !sawDispatching {
		t.Fatalf("the probe did not observe both poles (consuming=%v dispatching=%v) — it has "+
			"rotted and would accept any value-flag list", sawConsuming, sawDispatching)
	}
}

// TestGitGlobalValueFlagsDoNotHideAHazard is the behavioural half, through the
// real bundled registry. Each prefix is accepted by git and runs the hazard
// normally, so the guard's parse must not diverge from what executes.
func TestGitGlobalValueFlagsDoNotHideAHazard(t *testing.T) {
	hazards := []string{
		"push --force origin main",
		"commit --no-verify -m x",
		"reset --hard",
	}
	// Both spellings: the separate-token form is the one that evaded, since the
	// glued --opt=value form was already skipped as an ordinary flag.
	prefixes := []string{
		"", "-C /tmp ", "-c user.name=x ",
		"--attr-source HEAD ", "--attr-source=HEAD ",
		"--config-env NAME=ENVV ", "--config-env=NAME=ENVV ",
		"--shallow-file /dev/null ", "--shallow-file=/dev/null ",
		"--git-dir /tmp/.git ", "--work-tree /tmp ", "--namespace ns ",
		// Composed, in both orders: a residual must not survive by pairing with a
		// flag that was already handled.
		"--attr-source HEAD --shallow-file /dev/null ",
		"--shallow-file /dev/null --attr-source HEAD ",
	}
	for _, hazard := range hazards {
		bare := checkGuard(t, "git "+hazard)
		if bare.Verdict == VerdictAllow {
			t.Fatalf("`git %s` is not matched at all by the bundled registry — the fixture is wrong", hazard)
		}
		for _, prefix := range prefixes {
			candidate := "git " + prefix + hazard
			if got := checkGuard(t, candidate); got.Verdict != bare.Verdict {
				t.Errorf("%q\n  got verdict %q, want %q (same as the bare command).\n"+
					"  A global option taking a value must not displace the subcommand (gh-299).",
					candidate, got.Verdict, bare.Verdict)
			}
		}
	}
}

// TestGitGlobalValueFlagsStillAllowBenignCommands guards the other direction: the
// widened list must not start blocking commands that were never hazards. A flag
// whose VALUE reads like a hazardous subcommand is the shape that would break if
// a flag were stepped over without its argument.
func TestGitGlobalValueFlagsStillAllowBenignCommands(t *testing.T) {
	for _, candidate := range []string{
		"git --attr-source HEAD status",
		"git --attr-source push log",
		"git --shallow-file /dev/null status",
		"git --shallow-file push log",
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

// TestGitGlobalValueFlagsAreDocumented keeps the sets coherent. There is
// deliberately NO size assertion: a count is what froze the previous, wrong list
// — it read as authoritative while being superset-tolerant, so it could never
// notice the member that was missing.
func TestGitGlobalValueFlagsAreDocumented(t *testing.T) {
	for _, f := range append(registryGitValueFlags(t), gitProbeCandidates...) {
		if !strings.HasPrefix(f, "-") {
			t.Errorf("%q is not a flag spelling", f)
		}
	}
	for _, f := range gitNonConsumingButListed {
		if slices.Contains(gitGlobalValueFlags, f) {
			t.Errorf("%q is in both the value-taking and the non-value-taking set", f)
		}
	}
	// Every claim must be re-derivable, so anything asserted to take a value has
	// to be probed against git.
	for _, f := range gitGlobalValueFlags {
		if !slices.Contains(gitProbeCandidates, f) {
			t.Errorf("%q is claimed to take a value but is not among the probe candidates", f)
		}
	}
}
