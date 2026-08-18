package guard

import (
	"strings"
	"testing"
)

// verdictOf runs one candidate through the bundled registry and returns the
// decision, failing the test on an evaluation fault (which is never a clearance).
func verdictOf(t *testing.T, command string) Decision {
	t.Helper()
	d, err := Defaults().Check(command)
	if err != nil {
		t.Fatalf("guard could not evaluate %q: %v", command, err)
	}
	return d
}

// wrapperBypasses are the ten silent allows iss-272 recorded on the merged tree.
// Six are ordinary wrappers whose speculative suffix reaches the hazard directly;
// busybox reaches it through the interpreter path once the suffix is expanded.
// The remaining three are the exec-string family, which Tier 2 cannot reach at
// all — pinned separately below so part C's arrival is visible rather than
// assumed.
var wrapperBypasses = []string{
	"nice git push --force origin main",
	"nice -n 5 git push --force origin main",
	"setsid git push --force origin main",
	"flock /tmp/l git push --force origin main",
	"stdbuf -o0 git push --force origin main",
	"chroot / git push --force origin main",
	`busybox sh -c "gh repo delete owner/repo"`,
}

// execStringBypasses take a command STRING run by a shell they are not. Their
// payload stays one opaque token that isShellFamily will not open, so no
// speculative start can see the hazard inside it. Part C's table is the only
// thing that reaches them (adr-42 decision 6).
var execStringBypasses = []string{
	`su -c "gh repo delete owner/repo"`,
	`runuser -c "git push --force origin main"`,
	`script -c "git push --force origin main" /dev/null`,
}

// TestUnrecognisedLauncherWarnsInsteadOfAllowing is the whole point of the
// fail-safe: a hazard launched through a program the wrapper table does not name
// stops being a confident allow. It never becomes a block — an unknown command is
// not proof that it execs its arguments.
func TestUnrecognisedLauncherWarnsInsteadOfAllowing(t *testing.T) {
	for _, cmd := range wrapperBypasses {
		t.Run(cmd, func(t *testing.T) {
			d := verdictOf(t, cmd)
			if d.Verdict != VerdictWarn {
				t.Fatalf("verdict = %q, want %q — this is the silent allow iss-272 recorded", d.Verdict, VerdictWarn)
			}
			if d.EntryID != speculativeEntryID {
				t.Errorf("EntryID = %q, want the Tier 2 id %q", d.EntryID, speculativeEntryID)
			}
		})
	}
}

// TestSpeculationCarriesTheHazardsLesson keeps the warn worth reading: the
// successor and the why come from the entry the speculative suffix matched, so a
// Tier 2 warn teaches the same safe form a Tier 1 block would.
func TestSpeculationCarriesTheHazardsLesson(t *testing.T) {
	d := verdictOf(t, "nice git push --force origin main")
	if d.Successor == "" {
		t.Error("a speculative warn carries no successor: the refusal teaches nothing")
	}
	if !strings.Contains(d.Message, "git-push-force") {
		t.Errorf("message does not name the entry that matched: %q", d.Message)
	}
}

// TestTier2IsDistinguishableFromTier1 is adr-42 decision 2's stated requirement.
// The rationale for warning rather than blocking is that an unknown command is not
// proof it execs its arguments — a reader who cannot tell a speculative warn from
// a literal one has not been given that.
func TestTier2IsDistinguishableFromTier1(t *testing.T) {
	spec := verdictOf(t, "nice git clean -fd")
	lit := verdictOf(t, "git clean -fd")
	if lit.Verdict != VerdictWarn || spec.Verdict != VerdictWarn {
		t.Fatalf("both should warn; literal=%q speculative=%q", lit.Verdict, spec.Verdict)
	}
	if spec.EntryID == lit.EntryID {
		t.Fatal("a speculative warn is reported identically to a literal one")
	}
	if spec.Family == "" {
		t.Error("a speculative warn carries no family: nothing marks it as speculative")
	}
	if !strings.Contains(strings.ToLower(spec.Message), "does not recognise") {
		t.Errorf("the message does not say the verdict is speculative: %q", spec.Message)
	}
}

// TestSpeculationIsGatedPerSegment closes the suppression bypass adr-42 decision
// 4 records: Check merges matches across every segment, so a gate that asked
// "did anything match the command line" would let one warn-tier command disarm
// the fail-safe for everything after it.
func TestSpeculationIsGatedPerSegment(t *testing.T) {
	d := verdictOf(t, "git clean -fd ; nice git push --force origin main")
	if d.Verdict != VerdictWarn {
		t.Fatalf("verdict = %q, want %q", d.Verdict, VerdictWarn)
	}
	var sawLiteral, sawSpeculative bool
	for _, id := range d.Matches {
		switch id {
		case "git-clean":
			sawLiteral = true
		case speculativeEntryID:
			sawSpeculative = true
		}
	}
	if !sawLiteral {
		t.Errorf("the literal git-clean match is missing from %v", d.Matches)
	}
	if !sawSpeculative {
		t.Errorf("prefixing a warn-tier command disarmed the fail-safe: %v\n"+
			"the gate must be per segment, not per command line", d.Matches)
	}
}

// TestSpeculationNeverBlocks pins the demotion. expandPayloads raises blocker
// signals of its own, and a speculative suffix reaches them; honouring one at its
// own tier would block on two layers of uncertainty at once — an unknown command
// that may not exec its arguments, carrying a payload the guard could not read.
func TestSpeculationNeverBlocks(t *testing.T) {
	// Verified reachable: the same payload directly under env -S blocks today.
	if d := verdictOf(t, `env -S 'git push --force $X'`); d.Verdict != VerdictBlock {
		t.Fatalf("precondition failed: %q = %q, want a synthetic block", `env -S`, d.Verdict)
	}
	d := verdictOf(t, `myrunner env -S 'git push --force $X'`)
	if d.Verdict == VerdictBlock {
		t.Fatalf("a speculative suffix produced a BLOCK: %q\nadr-42 decision 2: Tier 2 never blocks", d.Message)
	}
	if d.Verdict != VerdictWarn {
		t.Fatalf("verdict = %q, want %q — the signal was dropped, not demoted", d.Verdict, VerdictWarn)
	}
}

// TestSpeculationDoesNotReachTheExecStringFamily pins a documented limit rather
// than a hope. These three stay silent until part C lands; if one starts warning,
// the record saying "still silent until part C" is what needs updating.
func TestSpeculationDoesNotReachTheExecStringFamily(t *testing.T) {
	for _, cmd := range execStringBypasses {
		t.Run(cmd, func(t *testing.T) {
			if d := verdictOf(t, cmd); d.Verdict != VerdictAllow {
				t.Fatalf("verdict = %q, want %q — part C has landed, so update adr-42's coverage claim", d.Verdict, VerdictAllow)
			}
		})
	}
}

// TestSpeculationLeavesOrdinaryWorkAlone is the false-positive floor in unit
// form. Quoting silences the fail-safe, and a command whose own entry matched is
// never speculated on.
func TestSpeculationLeavesOrdinaryWorkAlone(t *testing.T) {
	quiet := []string{
		`echo 'do not run git push --force'`,
		`rg 'git push --force' docs/`,
		`git commit -m 'ban git push --force'`,
		"make preflight",
		"go test ./...",
		"npm run build",
		"docker ps -a",
		"ls -la",
	}
	for _, cmd := range quiet {
		t.Run(cmd, func(t *testing.T) {
			if d := verdictOf(t, cmd); d.Verdict != VerdictAllow {
				t.Errorf("verdict = %q, want %q (entry %q): %s", d.Verdict, VerdictAllow, d.EntryID, d.Message)
			}
		})
	}
}

// TestSpeculationCapWarnsRatherThanSkipping holds the fail-loud half of the DoS
// bound. Past the start cap the guard has stopped looking, and stopping quietly
// would be the silent allow this whole change exists to remove.
func TestSpeculationCapWarnsRatherThanSkipping(t *testing.T) {
	long := "myrunner " + strings.Repeat("alpha ", maxSpeculativeStarts+8) + "beta"
	d := verdictOf(t, long)
	if d.Verdict != VerdictWarn {
		t.Fatalf("verdict = %q, want %q: a command line too long to inspect was waved through", d.Verdict, VerdictWarn)
	}
	if !strings.Contains(strings.ToLower(d.Message), "too long") {
		t.Errorf("the cap warn does not say why it fired: %q", d.Message)
	}
}

// TestReservedEntryIDsAreRefused stops a registry from claiming an id the
// synthetic verdicts use. A colliding entry would be indexed out of r.Entries by
// a synthetic winner and render a blank message — and, worse, would let a repo
// dress a real entry up as the guard's own voice.
func TestReservedEntryIDsAreRefused(t *testing.T) {
	for _, id := range []string{syntheticEntryID, speculativeEntryID} {
		t.Run(id, func(t *testing.T) {
			r := Defaults()
			r.Entries[id] = Entry{
				ID: id, Tier: TierWarn, Why: "x", Successor: "y",
				Pattern: Pattern{Command: "definitelynotacommand"},
			}
			if err := Validate(r); err == nil {
				t.Fatalf("Validate accepted an entry claiming the reserved id %q", id)
			}
		})
	}
}
