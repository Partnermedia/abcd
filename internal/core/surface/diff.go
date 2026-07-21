package surface

// BreakKind classifies one structural incompatibility between two snapshots.
//
// The set is exactly the break taxonomy of spc-10 and plan outcome 5 — nothing
// more. A kind absent from this list (a changed flag type, a flag losing its
// shorthand, a command becoming hidden) is deliberately NOT a break: the
// taxonomy is what a release contract can be held to, and widening it here
// would make the gate refuse releases the contract permits. Those are named as
// limits in Diff rather than silently folded in.
//
// The string values are stable identifiers a renderer or a machine-readable
// preview can switch on, so renaming a constant is safe and changing a value is
// a contract change.
type BreakKind string

const (
	// BreakCommandRemoved covers a removal AND a rename: a renamed command is a
	// removal at the path callers type, plus an addition nobody depends on yet.
	BreakCommandRemoved BreakKind = "command_removed"
	// BreakFlagRemoved covers a removed or renamed flag, for the same reason.
	BreakFlagRemoved BreakKind = "flag_removed"
	// BreakFlagRequired is a flag that was optional or absent becoming
	// mandatory — the one narrowing that adds surface rather than deleting it,
	// and the one most easily missed, because the flag is still there.
	BreakFlagRequired BreakKind = "flag_required"
	// BreakManifestRemoved is a declared manifest key path that is gone.
	BreakManifestRemoved BreakKind = "manifest_entry_removed"
)

// Break is one structural incompatibility, named precisely enough to act on.
//
// Surface is the whole point of the type. A gate that reports "a break was
// detected" tells the operator to go hunting; Surface names the command path,
// the command-and-flag, or the manifest file and key that changed, in the form
// the operator reads it in the tree.
type Break struct {
	Kind    BreakKind
	Surface string
}

// String renders one break as the line a failing gate names it with. It says
// "removed or renamed" for the two deletion kinds because a structural diff
// genuinely cannot tell them apart — the old path is gone either way, and
// claiming to know which happened would be a guess the operator has to check.
func (b Break) String() string {
	switch b.Kind {
	case BreakCommandRemoved:
		return "command removed or renamed: " + b.Surface
	case BreakFlagRemoved:
		return "flag removed or renamed: " + b.Surface
	case BreakFlagRequired:
		return "flag is now required: " + b.Surface
	case BreakManifestRemoved:
		return "manifest entry removed: " + b.Surface
	default:
		return string(b.Kind) + ": " + b.Surface
	}
}

// Diff reports every way current narrows the compatibility surface base
// declared — the structural half of the release guardrail (spc-10 AC 4).
//
// It is a pure comparison of two values: it reads no git, no files, and no
// clock, so the same pair of snapshots always yields the same breaks in the same
// order. Where the two snapshots come from is the CALLER's decision and is where
// the guardrail is easiest to get wrong: comparing the committed baseline
// against the current tree compares a file against itself, because a drift test
// keeps them equal. The baseline must be read out of the last release tag.
//
// What Diff reports is the taxonomy and only the taxonomy:
//
//   - a removed or renamed command;
//   - a removed or renamed flag on a command that still exists;
//   - a flag that was optional, or absent, becoming required on a command that
//     still exists;
//   - a removed declared manifest entry.
//
// Additions of any kind are silent — a new command, a new optional flag, a new
// manifest entry — as are reordering and every change the snapshot does not
// model at all (help text, descriptions, summaries), which is precisely why the
// snapshot does not model them.
//
// Two known limits, stated rather than hidden. A flag whose VALUE TYPE changes
// (string to stringSlice) is a compatibility event the taxonomy does not list,
// so it passes here and rests on the author's `breaking` judgement. And a
// behavioural break behind an unchanged surface is invisible to any structural
// diff. Both are the author's call; the gate backstops the structural ones.
//
// Results are ordered: commands in canonical path order (each command's own flag
// breaks before the next command's), then manifest entries in canonical order.
// The whole set is returned, never just the first, so one fix-and-rerun cycle
// shows the operator everything that changed.
func Diff(base, current Snapshot) []Break {
	// Canonicalise defensively rather than trusting the caller to have gone
	// through NewSnapshot or Decode: a hand-built Snapshot must not be able to
	// change the ORDER of a gate's findings.
	base, current = canonical(base), canonical(current)

	currentCommands := make(map[string]Command, len(current.Commands))
	for _, c := range current.Commands {
		currentCommands[c.Path] = c
	}

	var breaks []Break
	for _, was := range base.Commands {
		now, stillThere := currentCommands[was.Path]
		if !stillThere {
			// The command is gone; its flags went with it. Reporting them too
			// would turn one break into one-per-flag and bury the command that
			// is the thing to fix.
			breaks = append(breaks, Break{Kind: BreakCommandRemoved, Surface: was.Path})
			continue
		}
		breaks = append(breaks, flagBreaks(was, now)...)
	}

	currentEntries := make(map[ManifestEntry]struct{}, len(current.Manifest))
	for _, e := range current.Manifest {
		currentEntries[e] = struct{}{}
	}
	for _, was := range base.Manifest {
		if _, stillThere := currentEntries[was]; !stillThere {
			breaks = append(breaks, Break{Kind: BreakManifestRemoved, Surface: was.File + ":" + was.Key})
		}
	}
	return breaks
}

// flagBreaks compares the flags of ONE command that exists in both snapshots.
//
// Requiredness is judged here, and only here, which is what keeps the two
// taxonomy rows that meet at this point from contradicting each other: "a new
// command is not a break" and "an absent flag becoming required is a break".
// Every flag of a brand-new command was absent at the baseline, so judging
// requiredness across the whole tree would report each new command carrying a
// required flag as a break. Nobody can depend on surface that did not exist, so
// a new command's flags — required or not — are new surface, not narrowed
// surface. A required flag arriving on an EXISTING command is the real
// narrowing: every call that worked before now fails.
func flagBreaks(was, now Command) []Break {
	baseFlags := make(map[string]Flag, len(was.Flags))
	for _, f := range was.Flags {
		baseFlags[f.Name] = f
	}
	currentFlags := make(map[string]Flag, len(now.Flags))
	for _, f := range now.Flags {
		currentFlags[f.Name] = f
	}

	var breaks []Break
	for _, f := range was.Flags {
		if _, stillThere := currentFlags[f.Name]; !stillThere {
			breaks = append(breaks, Break{Kind: BreakFlagRemoved, Surface: flagSurface(was.Path, f.Name)})
		}
	}
	for _, f := range now.Flags {
		if !f.Required {
			continue
		}
		// Already required at the baseline: the narrowing shipped in an earlier
		// release and is not this cut's break to declare again.
		if before, existed := baseFlags[f.Name]; existed && before.Required {
			continue
		}
		breaks = append(breaks, Break{Kind: BreakFlagRequired, Surface: flagSurface(now.Path, f.Name)})
	}
	return breaks
}

// flagSurface names a flag the way an operator invokes it, so the failure can be
// pasted into a shell and understood without consulting the snapshot.
func flagSurface(commandPath, flagName string) string {
	return commandPath + " --" + flagName
}
