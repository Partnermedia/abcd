package guard

import "testing"

// TestUnquotedBraceGroupIsRefused is the repro for iss-2608221457227161. Bash
// expands `{--force,}` to byte-identical `--force` argv, but the guard's
// tokenizer does not expand braces: it read the literal token `{--force,}`, no
// blocker matched it, and a Tier-1 hazard was a silent allow — the same
// mutate-the-flag-token shape the redirection fix closed. The guard does not
// expand the group either; it REFUSES it, because a token it cannot expand is a
// token whose argv it cannot check.
func TestUnquotedBraceGroupIsRefused(t *testing.T) {
	for _, cmd := range []string{
		// The reported shape: a single-element group with an empty alternative.
		`git push {--force,} origin main`,
		// The same trick spelled the other way round, and on the long flag.
		`git push {,--force} origin main`,
		`git push {--force-with-lease,} origin main`,
		// A real two-alternative group, and a nested one whose comma is not at
		// the group's own top level.
		`git push {--force,--dry-run} origin main`,
		`git push {{--force,--dry-run}} origin main`,
		// A range, the other expansion form.
		`rm -rf dir{1..9}`,
		// The group hidden one execute-a-string layer down, where the outer
		// quotes exempt it but the payload's own tokenize does not.
		`sh -c 'git push {--force,} origin main'`,
	} {
		t.Run(cmd, func(t *testing.T) {
			if d := verdictOf(t, cmd); d.Verdict != VerdictBlock {
				t.Errorf("verdict = %q, want %q: an unquoted brace group expands to argv the guard cannot check", d.Verdict, VerdictBlock)
			}
		})
	}
}

// TestBraceRefusalIsFailClosed pins the refusal POSTURE, not just the verdict.
// Returning ErrUnparsableCommand would have been the obvious route and is the
// wrong one: the `guard check` verb maps a tokenize error to a blocking exit,
// but the pre-tool-use hook maps it to fail-OPEN — so the bypass would have
// survived on the surface that matters. Only a real VerdictBlock is fail-closed
// on both front doors.
func TestBraceRefusalIsFailClosed(t *testing.T) {
	const cmd = `git push {--force,} origin main`
	d, err := Defaults().Check(cmd)
	if err != nil {
		t.Fatalf("Check(%q) returned an error: %v — a tokenize error fails OPEN on the hook", cmd, err)
	}
	if d.Verdict != VerdictBlock {
		t.Fatalf("verdict = %q, want %q", d.Verdict, VerdictBlock)
	}
	if d.Tier != TierBlocker {
		t.Errorf("tier = %q, want %q", d.Tier, TierBlocker)
	}
	if d.EntryID != braceEntryID {
		t.Errorf("EntryID = %q, want the reserved brace id %q", d.EntryID, braceEntryID)
	}
	if _, isEntry := Defaults().Entries[braceEntryID]; isEntry {
		t.Errorf("the reserved id %q must never be a registry entry", braceEntryID)
	}
	if d.Reason == "" || d.Why == "" || d.Successor == "" || d.Message == "" {
		t.Errorf("a synthetic verdict must carry Reason/Why/Successor/Message for the report, got %+v", d)
	}
	if !contains(d.Matches, braceEntryID) {
		t.Errorf("Matches = %v, must list the reserved brace id", d.Matches)
	}
}

// TestBraceHandlingLeavesEveryOtherShapeAlone is the false-positive half. A
// brace sequence that bash does NOT expand — quoted, parameter expansion, a
// group with no alternative, a reserved-word group command — must keep exactly
// the verdict it had before braces were handled at all.
func TestBraceHandlingLeavesEveryOtherShapeAlone(t *testing.T) {
	cases := []struct {
		cmd  string
		want Verdict
	}{
		// Quoted: the bytes never reach the tokenizer's structural branches, so
		// bash hands the child the literal token and so does the guard.
		{`git push '{--force,}' origin main`, VerdictAllow},
		{`git push "{--force,}" origin main`, VerdictAllow},
		{`echo '{a,b}'`, VerdictAllow},
		// Parameter expansion, not a brace group.
		{`echo ${HOME}`, VerdictAllow},
		{`echo ${x:-a,b}`, VerdictAllow},
		{`git commit -m ${MSG}`, VerdictAllow},
		// A brace with no alternative is a literal in bash too.
		{`echo {a}`, VerdictAllow},
		{`echo {`, VerdictAllow},
		{`awk {print}`, VerdictAllow},
		// Ordinary commands, unaffected.
		{`git status`, VerdictAllow},
		{`ls -la`, VerdictAllow},
		// A reserved-word group command still puts its inner command in command
		// position, so the blocker inside it still fires — the brace handling
		// must not turn that into a synthetic refusal instead.
		{`{ git push --force origin main; }`, VerdictBlock},
	}
	for _, c := range cases {
		t.Run(c.cmd, func(t *testing.T) {
			d := verdictOf(t, c.cmd)
			if d.Verdict != c.want {
				t.Errorf("verdict = %q, want %q (%+v)", d.Verdict, c.want, d)
			}
			if c.want != VerdictBlock && d.EntryID == braceEntryID {
				t.Errorf("a shape bash does not brace-expand was refused as one: %+v", d)
			}
			if c.cmd == `{ git push --force origin main; }` && d.EntryID == braceEntryID {
				t.Errorf("a group command was refused as a brace expansion: %+v", d)
			}
		})
	}
}

// TestBraceGroupIsRecordedOnTheSegment pins the tokenizer half directly: the
// flag rides on the segment that carried the group, so Check folds one signal
// however many segments the line has.
func TestBraceGroupIsRecordedOnTheSegment(t *testing.T) {
	segs, err := tokenize(`echo hi && git push {--force,} origin main`)
	if err != nil {
		t.Fatalf("tokenize: %v", err)
	}
	if len(segs) != 2 {
		t.Fatalf("got %d segments, want 2: %v", len(segs), render(segs))
	}
	if segs[0].braceGroup {
		t.Errorf("the brace-free segment was marked: %v", segs[0].tokens)
	}
	if !segs[1].braceGroup {
		t.Errorf("the segment carrying the brace group was not marked: %v", segs[1].tokens)
	}
}
