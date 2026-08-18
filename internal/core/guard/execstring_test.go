package guard

import (
	"strings"
	"testing"
)

// The exec-string family: `su -c CMD`, `runuser -c CMD`, `script -c CMD FILE`,
// `flock FILE -c CMD`.
//
// Each takes a command STRING that a shell executes, so the payload is shell
// grammar the tokenizer already reads — but none of them IS a shell, so they
// belong to neither isShellFamily nor the wrapper walk. They were the last three
// of the ten silent allows iss-272 recorded, and Tier 2 cannot reach them: the
// payload is one opaque token, and no speculative start can see inside it.
//
// adr-42 decision 6 requires a TABLE rather than a generalisation of
// shellCPayload. The reason is the long spellings: shellCPayload matches short
// clusters, so generalising it would resolve `su -c` free and return shellNone
// for `su --command` — six new SILENT allows, which is the exact defect this
// whole change exists to close. The long spellings are therefore the first thing
// this file tests.

// execStringShort is the spelling everyone writes.
var execStringShort = []struct{ name, command string }{
	{"su", `su -c "gh repo delete owner/repo"`},
	{"su with user", `su bob -c "gh repo delete owner/repo"`},
	{"runuser", `runuser -c "git push --force origin main"`},
	{"script", `script -c "git push --force origin main" /dev/null`},
	{"flock", `flock /tmp/lock -c "git push --force origin main"`},
}

// execStringLong is the spelling a generalisation would have missed. Every one of
// these is a documented alias of the short form and runs the identical command.
var execStringLong = []struct{ name, command string }{
	{"su --command", `su --command "gh repo delete owner/repo"`},
	{"su --command=", `su --command="gh repo delete owner/repo"`},
	{"su --session-command", `su --session-command "gh repo delete owner/repo"`},
	{"runuser --command", `runuser --command "git push --force origin main"`},
	{"script --command", `script --command "git push --force origin main" /dev/null`},
	{"flock --command", `flock /tmp/lock --command "git push --force origin main"`},
}

// TestExecStringShortFormsAreRead closes the three bypasses part A could not
// reach, plus the two spellings that carry an operand before the flag.
func TestExecStringShortFormsAreRead(t *testing.T) {
	for _, c := range execStringShort {
		t.Run(c.name, func(t *testing.T) {
			d := verdictOf(t, c.command)
			if d.Verdict != VerdictBlock {
				t.Fatalf("verdict = %q, want %q — the payload was not read", d.Verdict, VerdictBlock)
			}
		})
	}
}

// TestExecStringLongFormsAreRead is the test the table exists for. A
// generalisation of shellCPayload passes every case above and fails every case
// here, silently — no verdict, no warn, no trace.
func TestExecStringLongFormsAreRead(t *testing.T) {
	for _, c := range execStringLong {
		t.Run(c.name, func(t *testing.T) {
			d := verdictOf(t, c.command)
			if d.Verdict == VerdictAllow {
				t.Fatalf("SILENT ALLOW on %s.\n"+
					"This is the six-new-holes case adr-42 decision 6 rejects a shellCPayload generalisation over.", c.command)
			}
			if d.Verdict != VerdictBlock {
				t.Errorf("verdict = %q, want %q", d.Verdict, VerdictBlock)
			}
		})
	}
}

// TestExecStringSkipsItsOwnFlagValues keeps the scan from reading a flag's value
// as the payload: `su -s /bin/sh -c <hazard>` carries the shell in -s.
func TestExecStringSkipsItsOwnFlagValues(t *testing.T) {
	for _, cmd := range []string{
		`su -s /bin/sh -c "gh repo delete owner/repo"`,
		`su --shell /bin/sh -c "gh repo delete owner/repo"`,
		`runuser -u bob -c "git push --force origin main"`,
		`flock -w 5 /tmp/lock -c "git push --force origin main"`,
		`script -q -c "git push --force origin main" /dev/null`,
	} {
		t.Run(cmd, func(t *testing.T) {
			if d := verdictOf(t, cmd); d.Verdict != VerdictBlock {
				t.Errorf("verdict = %q, want %q: the scan lost the payload behind another flag", d.Verdict, VerdictBlock)
			}
		})
	}
}

// TestExecStringWithoutItsValueFailsLoud holds the fail-safe. A payload flag with
// nothing after it is a command the guard cannot read, which is a warn — never
// the silent allow that shape produced before.
func TestExecStringWithoutItsValueFailsLoud(t *testing.T) {
	for _, cmd := range []string{"su -c", "runuser --command"} {
		t.Run(cmd, func(t *testing.T) {
			d := verdictOf(t, cmd)
			if d.Verdict == VerdictAllow {
				t.Fatalf("%q was allowed: a payload flag the guard cannot resolve must be loud", cmd)
			}
		})
	}
}

// TestExecStringEndOfOptionsIsRespected stops the scan reading a hazard that is
// an ARGUMENT rather than a payload. After `--` the tokens belong to the command
// being launched, and runuser's direct-exec grammar is the wrapper walk's job.
func TestExecStringEndOfOptionsIsRespected(t *testing.T) {
	// The direct-exec grammar: runuser is a wrapper here, so this is a Tier 1
	// match on git itself, not an exec-string payload.
	d := verdictOf(t, "runuser -u bob -- git push --force origin main")
	if d.Verdict != VerdictBlock {
		t.Fatalf("verdict = %q, want %q", d.Verdict, VerdictBlock)
	}
	if d.EntryID != "git-push-force" {
		t.Errorf("EntryID = %q, want the entry that names the hazard", d.EntryID)
	}

	// The discriminating shape: after `--`, a `-c` belongs to the launched
	// program. Without the stop the scan keeps looking, finds it, and blocks on a
	// string runuser never runs — so the verdict, not just the reasoning, changes.
	foreign := verdictOf(t, `runuser -u bob -- myprog -c "gh repo delete owner/repo"`)
	if foreign.Verdict == VerdictBlock {
		t.Errorf("FALSE BLOCK: the `-c` after `--` is myprog's, not runuser's:\n  %s", foreign.Message)
	}
}

// TestExecStringStopsAtTheLaunchedCommand is the false-block finding, verified
// against the real binary before it was fixed:
//
//	$ flock /tmp/lock /bin/echo -c 'git push --force origin main'
//	-c git push --force origin main
//
// flock handed `-c` and the string to /bin/echo as ordinary arguments. Nothing
// ran the string, and the guard blocked on it — falsifying, in the same change
// that wrote it, the claim that "a mis-parse yields a non-match, never a false
// block".
func TestExecStringStopsAtTheLaunchedCommand(t *testing.T) {
	for _, cmd := range []string{
		`flock /tmp/lock /bin/echo -c "git push --force origin main"`,
		`flock /tmp/lock myrunner -c cfg true`,
		`flock -w 5 /tmp/lock /bin/echo --command "git push --force origin main"`,
	} {
		t.Run(cmd, func(t *testing.T) {
			if d := verdictOf(t, cmd); d.Verdict == VerdictBlock {
				t.Fatalf("FALSE BLOCK: the `-c` belongs to the launched command, not to flock.\n  %s", d.Message)
			}
		})
	}
}

// TestForeignFlagDoesNotDisarmTheFailSafe is the same defect's other half, and
// the more dangerous one. Mis-reading a foreign `-c` as a payload also made the
// segment look understood, which switched Tier 2 off for it — so appending a
// `-c` token anywhere behind an exec-string verb was a one-token off switch for
// the fail-safe, the payload-level twin of the per-line suppression adr-42
// decision 4 rejects.
func TestForeignFlagDoesNotDisarmTheFailSafe(t *testing.T) {
	bare := verdictOf(t, "kubectl exec -c app pod -- gh repo delete owner/repo")
	if bare.Verdict != VerdictWarn {
		t.Fatalf("precondition: the unprefixed line should warn, got %q", bare.Verdict)
	}
	prefixed := verdictOf(t, "flock /tmp/lock kubectl exec -c app pod -- gh repo delete owner/repo")
	if prefixed.Verdict == VerdictAllow {
		t.Fatalf("SILENT ALLOW: a flock prefix disarmed the fail-safe for the whole segment")
	}
}

// TestExecStringIsReachedThroughTheWrapperChain pins the walk. Without it the
// two shapes execstring.go's own comment names degrade from a precise block to a
// Tier 2 warn — a downgrade quiet enough that every other test stays green.
func TestExecStringIsReachedThroughTheWrapperChain(t *testing.T) {
	for _, cmd := range []string{
		`sudo su -c "gh repo delete owner/repo"`,
		`nice runuser -c "git push --force origin main"`,
		`env -i su -c "gh repo delete owner/repo"`,
	} {
		t.Run(cmd, func(t *testing.T) {
			d := verdictOf(t, cmd)
			if d.Verdict != VerdictBlock {
				t.Fatalf("verdict = %q, want %q: the walk did not reach the verb behind its wrapper", d.Verdict, VerdictBlock)
			}
			if d.EntryID == speculativeEntryID {
				t.Errorf("the payload was found by the fail-safe guessing, not by reading the verb")
			}
		})
	}
}

// TestExecStringReadsAShortCluster pins clusteredPayload, which nothing else
// exercises: `su -lc CMD` is how getopt reads `-l -c CMD`, and `-c<value>` is the
// glued spelling of the same flag.
func TestExecStringReadsAShortCluster(t *testing.T) {
	for _, cmd := range []string{
		`su -lc "gh repo delete owner/repo"`,
		`su -c"gh repo delete owner/repo"`,
		`runuser -lc "git push --force origin main"`,
	} {
		t.Run(cmd, func(t *testing.T) {
			if d := verdictOf(t, cmd); d.Verdict != VerdictBlock {
				t.Errorf("verdict = %q, want %q: the clustered or glued payload was not read", d.Verdict, VerdictBlock)
			}
		})
	}
}

// TestExecStringSkipsAFlagShapedValue is the half of the value-flag skip that
// bites. A value that is not flag-shaped (`/bin/sh`, `bob`) is stepped as an
// operand whether or not the table names its flag, so only a FLAG-SHAPED value
// can show the skip working.
func TestExecStringSkipsAFlagShapedValue(t *testing.T) {
	// `-w --all` : --whitelist-environment consumes `--all`. Without the skip the
	// scan reads `--all` as a flag, keeps going, and the result is the same block —
	// so the case that discriminates is one where the skipped value would ITSELF
	// have been read as the payload flag.
	d := verdictOf(t, `su -w --command -c "gh repo delete owner/repo"`)
	if d.Verdict != VerdictBlock {
		t.Fatalf("verdict = %q, want %q", d.Verdict, VerdictBlock)
	}
	if !strings.Contains(d.Message, "repository") {
		t.Errorf("the payload read was not the -c string: %q", d.Message)
	}
}

// TestExecStringLeavesOrdinaryUseAlone is the false-positive side. These verbs
// are ordinary commands most of the time.
func TestExecStringLeavesOrdinaryUseAlone(t *testing.T) {
	for _, cmd := range []string{
		"su - bob",
		"script /tmp/typescript",
		"script -q /dev/null",
		"flock /tmp/lock true",
		`echo "su -c 'gh repo delete owner/repo'"`,
		`git commit -m "document su -c and runuser -c"`,
	} {
		t.Run(cmd, func(t *testing.T) {
			if d := verdictOf(t, cmd); d.Verdict != VerdictAllow {
				t.Errorf("verdict = %q, want %q (%s): %s", d.Verdict, VerdictAllow, d.EntryID, d.Message)
			}
		})
	}
}

// TestExecStringVerbsAreNotTheShellFamily pins the boundary the table draws.
// These four are NOT shells, so they must not be added to isShellFamily — that
// set is about POSIX shell grammar, and `su -c` in particular carries no such
// guarantee (it runs the TARGET user's login shell, overridable with -s, so a
// csh or fish login shell is not this grammar at all). The mismatch costs false
// negatives only: a mis-parse yields a non-match, never a false block.
func TestExecStringVerbsAreNotTheShellFamily(t *testing.T) {
	for verb := range execStringVerbs {
		if isShellFamily(verb) {
			t.Errorf("%q is in BOTH the exec-string table and the shell family; one of them is wrong", verb)
		}
	}
	for _, shell := range posixShellFamily {
		if _, ok := execStringVerbs[shell]; ok {
			t.Errorf("%q is a shell and must be read by the shell path, not the exec-string table", shell)
		}
	}
}

// TestExecStringFamilyNamesItself keeps the report legible: a reader who gets a
// verdict about a payload should be told which wrapper carried it.
func TestExecStringFamilyNamesItself(t *testing.T) {
	d := verdictOf(t, `su -c "gh repo delete owner/repo"`)
	if d.Verdict != VerdictBlock {
		t.Fatalf("verdict = %q, want %q", d.Verdict, VerdictBlock)
	}
	if !strings.Contains(d.Message, "gh repo delete") && !strings.Contains(d.Message, "repository") {
		t.Errorf("the message does not describe the hazard that was found: %q", d.Message)
	}
}
