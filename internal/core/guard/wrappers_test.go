package guard

import (
	"os/exec"
	"strings"
	"testing"
)

// DERIVED FROM THE BINARY, NOT FROM A DOCUMENT.
//
// gh-299 is why. Its git value-flag list was taken from the bug report, and it
// was wrong three ways — it omitted `--shallow-file` (a live force-push bypass,
// in git since 1.9) and counted two non-value flags as value-taking. It totalled
// nine either way, so a size assertion certified the wrong list as complete.
//
// The wrapper table repeats that shape at a larger scale, and documentation is
// not a way out: `nsenter --help` renders `-S/--setuid` as optional-argument and
// `unshare --help` renders the same letter, same package, same version, as
// required-argument, while BOTH binaries consume the following token. nsenter's
// help contradicts its own parser.
//
// So this file probes. Each candidate flag is handed a value and a command that
// prints a token; what the binary does with the following token is the answer,
// and the table has to agree with it.

// wrapperProbe describes one wrapper the guard steps over.
type wrapperProbe struct {
	// name is the program, and the key in the wrappers map.
	name string
	// operands is how many mandatory non-flag operands sit between the wrapper's
	// options and the command it launches (`chrt PRIORITY COMMAND`).
	operands int
	// operandArgs supplies those operands for the probe, and any flags the wrapper
	// needs before it will run at all.
	operandArgs []string
	// candidates are flags to classify. A flag the probe finds value-taking must
	// be in wrapperValueFlags; one it finds non-value-taking must not create a
	// miss by being listed (listing a non-value flag makes the walk step over the
	// COMMAND, which is a false negative, not a safe over-block).
	candidates []string
	// values supplies a plausible value per candidate flag; a flag whose value
	// the binary rejects tells us nothing.
	values map[string]string
	// needsRoot marks a wrapper that cannot run unprivileged. Its flags are still
	// asserted present in the table, but the classification cannot be re-derived
	// here, so the table is the claim and this note is the reason.
	needsRoot bool
	// unprobeable names flags this environment cannot classify, each with the
	// reason. A flag is allowed to be unclassified only by being written down
	// here — silence is what gh-299 shipped, and "we could not check it" is a
	// different statement from "we did not think about it".
	unprobeable map[string]string
}

// wrapperProbes covers every wrapper part B adds plus the ones already shipped
// whose grammar the same reasoning applies to. Platform: these are the Linux /
// util-linux / coreutils spellings. Several do not exist on macOS at all
// (`nsenter`, `unshare`, `chrt`, `taskset`, `ionice`, `setsid`, `flock`,
// `stdbuf`, `runuser`) and others carry BSD grammars there — macOS `su -c` means
// `-c class`, not a command string. A probe skips a binary it cannot find, so
// this file asserts on Linux and abstains elsewhere; the table itself is Linux's.
var wrapperProbes = []wrapperProbe{
	{name: "nice", candidates: []string{"-n"}, values: map[string]string{"-n": "5"}},
	{name: "setsid", candidates: []string{"-c", "-w", "-f"},
		unprobeable: map[string]string{
			"-c": "--ctty needs a controlling terminal; a test process has none " +
				"(`setsid -c /bin/echo hi` answers \"failed to set the controlling terminal\"). It is a boolean.",
		}},
	{name: "stdbuf", candidates: []string{"-i", "-o", "-e"},
		values: map[string]string{"-i": "0", "-o": "0", "-e": "0"}},
	{name: "ionice", candidates: []string{"-c", "-n", "-t"},
		values: map[string]string{"-c": "3", "-n": "4"}},
	{name: "eatmydata"},
	{name: "chrt", operands: 1, operandArgs: []string{"-f", "1"},
		candidates: []string{"-T", "-P", "-D", "-R", "-a"},
		values:     map[string]string{"-T": "100000", "-P": "1000000", "-D": "0"},
		unprobeable: map[string]string{
			"-T": "--sched-runtime is accepted only with SCHED_DEADLINE, which cannot be set here. " +
				"It consumes its value: `chrt -T -f 1 /bin/echo hi` answers \"invalid runtime argument: '-f'\".",
			"-P": "--sched-period, same restriction and same evidence as -T.",
			"-a": "--all-tasks only applies with -p PID, a grammar that launches no command.",
		}},
	// The mask operand is "1", valid both as a bitmask and as a -c CPU list, so
	// -c can be classified without the operand confounding it.
	{name: "taskset", operands: 1, operandArgs: []string{"1"},
		candidates: []string{"-c", "-a", "-p"},
		unprobeable: map[string]string{
			"-p": "--pid replaces the grammar entirely: it operates on a running process and launches no command.",
		}},
	{name: "flock", operands: 1, operandArgs: []string{"/tmp"},
		candidates: []string{"-w", "-E", "-n", "-s", "-x", "-u"},
		values:     map[string]string{"-w": "1", "-E": "9"}},
	{name: "chroot", operands: 1, operandArgs: []string{"/"},
		candidates: []string{"--userspec", "--groups", "--skip-chdir"},
		values:     map[string]string{"--userspec": "0:0", "--groups": "0"},
		needsRoot:  true},
	{name: "unshare", candidates: []string{"-S", "-G", "-w", "-R", "--map-user", "--map-group", "--propagation", "--setgroups", "-m", "-u", "-i", "-n", "-p", "-U", "-C", "-T", "-r", "-f"},
		values: map[string]string{"-S": "0", "-G": "0", "-w": "/tmp", "-R": "/",
			"--map-user": "0", "--map-group": "0", "--propagation": "private", "--setgroups": "deny"},
		unprobeable: map[string]string{
			"--setgroups": "needs a user namespace this test cannot create; documented as required-argument and listed on that basis.",
		}},
	// -W is the counterexample this file exists for. `nsenter --help` renders
	// `--wd[=dir]` and the probe agrees it consumes NOTHING, while `unshare -w`,
	// the same letter in the same package at the same version, IS required-argument
	// and consumes the next token. Listing -W would have made the walk step over
	// the COMMAND — a miss the table would have invented.
	{name: "nsenter", candidates: []string{"-t", "-S", "-G", "-W", "-m", "-u", "-i", "-n", "-p", "-U", "-C", "-T"},
		values:    map[string]string{"-t": "1", "-S": "0", "-G": "0", "-W": "/"},
		needsRoot: true},
	{name: "runuser", candidates: []string{"-u", "-g", "-G", "--session-command"},
		values:    map[string]string{"-u": "root", "-g": "root", "-G": "root", "--session-command": "true"},
		needsRoot: true},
}

// probeMarker is what the launched command prints. A wrapper that ran the command
// consumed exactly the tokens before it; one that ate the command name as a flag
// value prints nothing.
const probeMarker = "abcd-guard-probe"

// consumesNextToken reports whether the wrapper reads the token after flag as
// that flag's value. The discriminator is whether the launched command ran:
//
//	<wrapper> <flag> <value> <operandArgs...> /bin/echo MARKER   ran => value-taking
//	<wrapper> <flag> <operandArgs...> /bin/echo MARKER           ran => NOT value-taking
//
// The flag goes BEFORE the mandatory operands, because these wrappers stop
// parsing options at the first operand: `chrt -f 1 -T 100000 /bin/echo hi`
// answers "failed to execute -T", reading the flag as the command. Getting that
// order wrong makes every flag look unclassifiable.
//
// Both forms are tried, and they must disagree — if the wrapper refuses both (a
// permission it lacks, a value it dislikes) the probe learned nothing and says
// so rather than guessing, because guessing is what gh-299 cost.
func consumesNextToken(t *testing.T, p wrapperProbe, flag string) (consumes, learned bool) {
	t.Helper()
	value, hasValue := p.values[flag]

	run := func(args ...string) bool {
		cmd := exec.Command(p.name, args...)
		out, err := cmd.CombinedOutput()
		return err == nil && strings.Contains(string(out), probeMarker)
	}

	form := func(head ...string) []string {
		args := append([]string{}, head...)
		args = append(args, p.operandArgs...)
		return append(args, "/bin/echo", probeMarker)
	}

	ranWithout := run(form(flag)...)
	if !hasValue {
		// No plausible value to offer: the flag can only be classified by whether
		// it swallowed the command name.
		return !ranWithout, ranWithout
	}
	ranWith := run(form(flag, value)...)

	switch {
	case ranWith && !ranWithout:
		return true, true // the value was needed: the flag consumed it
	case ranWithout && !ranWith:
		return false, true // the extra token broke it: the flag took nothing
	default:
		return false, false // both or neither: inconclusive
	}
}

// TestWrapperValueFlagsMatchTheInstalledBinaries is the gh-299 lesson applied
// before the bug rather than after it. Every flag the table claims consumes a
// value must actually consume one, and — the direction gh-299's first fix left
// inert — every flag the probe finds value-taking must be in the table.
func TestWrapperValueFlagsMatchTheInstalledBinaries(t *testing.T) {
	var probed, skipped int
	for _, p := range wrapperProbes {
		p := p
		t.Run(p.name, func(t *testing.T) {
			if _, err := exec.LookPath(p.name); err != nil {
				skipped++
				t.Skipf("%s is not installed here; the table cannot be re-derived (it encodes the Linux grammar)", p.name)
			}
			if !wrappers[p.name] {
				t.Fatalf("%s is probed but absent from the wrappers map: the probe and the table have diverged", p.name)
			}
			if got := wrapperOperands[p.name]; got != p.operands {
				t.Errorf("wrapperOperands[%q] = %d, probe expects %d mandatory operands", p.name, got, p.operands)
			}
			listed := wrapperValueFlags[p.name]

			for _, flag := range p.candidates {
				consumes, learned := consumesNextToken(t, p, flag)
				inTable := containsString(listed, flag)
				switch {
				case !learned:
					if why, ok := p.unprobeable[flag]; ok {
						t.Logf("%s %s: not probed here — %s (table claims value-taking=%v)", p.name, flag, why, inTable)
						continue
					}
					if p.needsRoot {
						t.Logf("%s %s: inconclusive (needs privileges this test does not have); the table claims value-taking=%v", p.name, flag, inTable)
						continue
					}
					t.Errorf("%s %s: the probe could not classify it, and an unclassified flag is exactly what gh-299 shipped", p.name, flag)
				case consumes && !inTable:
					probed++
					t.Errorf("%s %s consumes the following token but is NOT in wrapperValueFlags.\n"+
						"Its value is read as the command, so every entry misses: `%s %s <value> <hazard>` is a silent allow.",
						p.name, flag, p.name, flag)
				case !consumes && inTable:
					probed++
					t.Errorf("%s %s takes NO value but IS in wrapperValueFlags.\n"+
						"The walk steps over the command itself, which is a false negative — listing a flag is not the safe direction.",
						p.name, flag)
				default:
					probed++
				}
			}
		})
	}
	t.Logf("probed %d flag classifications; %d wrappers unavailable on this platform", probed, skipped)
}

// TestEveryProbedWrapperIsInTheTable holds the other half on every platform,
// including the macOS leg where nothing can be probed: a wrapper this file knows
// about must be one the matcher steps, or the probe is documenting a table that
// does not exist.
func TestEveryProbedWrapperIsInTheTable(t *testing.T) {
	for _, p := range wrapperProbes {
		if !wrappers[p.name] {
			t.Errorf("%s is probed but missing from the wrappers map", p.name)
		}
	}
	// And the reverse, so a wrapper cannot be added to the matcher without a probe
	// or an explicit note saying why it needs none.
	for name := range wrappers {
		if containsString(wrappersWithoutProbes, name) {
			continue
		}
		found := false
		for _, p := range wrapperProbes {
			if p.name == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s is stepped by the matcher but has no probe and no exemption: add one, or say why it needs none", name)
		}
	}
}

// wrappersWithoutProbes are the wrappers whose grammar this file does not
// re-derive, each for a stated reason rather than by omission.
var wrappersWithoutProbes = []string{
	// Running sudo/doas in a test either prompts, escalates, or both. Their flag
	// lists predate this file and come from their own documented grammars.
	"sudo", "doas",
	// Shell builtins, not programs: `command`, `exec` are executed by the shell,
	// and there is no binary to probe.
	"command", "exec",
	// `env`'s -S is owned by the payload pre-pass, not by the wrapper walk, and
	// its other flags are probed there.
	"env",
	// POSIX-stable and value-flag-free (`nohup`) or already covered by their own
	// entries in the shipped table with no candidates left to classify.
	"nohup", "time", "xargs", "timeout",
	// A multiplexer, not a wrapper with flags: its first operand is the applet,
	// so stepping it leaves `sh -c …` in command position for the interpreter
	// path. Probing it would classify busybox's own applets, not a grammar.
	"busybox",
	// A loader shim with no options of its own.
	"proxychains",
}
