package guard

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/Partnermedia/abcd/internal/fsutil"
)

// RepoRelPath is the per-repo override file, relative to the repo worktree. It
// is a dedicated committed file rather than a rules.json domain: the entry schema
// is structured, and the rules loader's kill switch must never silently disable a
// safety guard (spc-16, "Config home"). Disabling the guard is itself a
// committed, reviewable act.
const RepoRelPath = ".abcd/guard.json"

// maxGuardFileBytes caps the per-repo override (trust boundary).
const maxGuardFileBytes = 256 * 1024

// Load returns the bundled defaults merged with <repoRoot>/.abcd/guard.json when
// that file exists. An absent file yields the defaults unchanged. A file that
// cannot be read, parsed, or validated is an ERROR, never a silent fallback to
// the defaults or to no guard at all — core names the failure and the caller
// (the hook shim) decides what it means for the session.
func Load(repoRoot string) (Registry, error) {
	// Refuse a symlinked .abcd directory component before touching the leaf, so a
	// swapped .abcd cannot redirect the read (trust boundary).
	if di, err := os.Lstat(filepath.Join(repoRoot, ".abcd")); err == nil && di.Mode()&os.ModeSymlink != 0 {
		return Registry{}, fmt.Errorf("%w: .abcd is a symlink (refusing to follow)", ErrMalformedConfig)
	}
	data, err := fsutil.ReadGuarded(filepath.Join(repoRoot, ".abcd", "guard.json"), maxGuardFileBytes)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return Defaults(), nil
		case errors.Is(err, syscall.ELOOP):
			return Registry{}, fmt.Errorf("%w: %s is a symlink (refusing to follow)", ErrMalformedConfig, RepoRelPath)
		case errors.Is(err, fsutil.ErrNotRegular):
			return Registry{}, fmt.Errorf("%w: %s is not a regular file", ErrMalformedConfig, RepoRelPath)
		case errors.Is(err, fsutil.ErrTooBig):
			return Registry{}, fmt.Errorf("%w: %s exceeds the %d-byte cap", ErrMalformedConfig, RepoRelPath, maxGuardFileBytes)
		default:
			return Registry{}, fmt.Errorf("%w: reading %s failed", ErrMalformedConfig, RepoRelPath)
		}
	}
	over, err := parse(data)
	if err != nil {
		return Registry{}, fmt.Errorf("%s: %w", RepoRelPath, err)
	}
	// The override declares its schema version explicitly: an absent version is a
	// truncated or hand-mangled file, not an invitation to guess (a safety config
	// fails closed on unrecognised input).
	if over.SchemaVersion != SchemaVersion {
		return Registry{}, fmt.Errorf("%w: %s must declare schema_version %d, got %d", ErrSchemaVersion, RepoRelPath, SchemaVersion, over.SchemaVersion)
	}
	merged := Merge(Defaults(), over)
	if err := Validate(merged); err != nil {
		return Registry{}, fmt.Errorf("%s: %w", RepoRelPath, err)
	}
	return merged, nil
}

// Merge overlays over onto base. Entry fields are per-field: a field set on the
// override wins, an absent field inherits the bundled entry (so {"tier":"warn"}
// retiers an entry while keeping its pattern, successor, and why). New entry keys
// are added. The kill switch is sticky — either side can disable the guard, and
// neither can silently re-enable the other's refusal to run it.
func Merge(base, over Registry) Registry {
	out := cloneRegistry(base)
	if over.SchemaVersion != 0 {
		out.SchemaVersion = over.SchemaVersion
	}
	out.Disabled = base.Disabled || over.Disabled
	if out.Entries == nil && len(over.Entries) > 0 {
		out.Entries = make(map[string]Entry, len(over.Entries))
	}
	for id, oe := range over.Entries {
		out.Entries[id] = mergeEntry(out.Entries[id], oe)
	}
	return out
}

func mergeEntry(base, over Entry) Entry {
	r := cloneEntry(base)
	r.ID = over.ID
	if over.Tier != "" {
		r.Tier = over.Tier
	}
	if over.Successor != "" {
		r.Successor = over.Successor
	}
	if over.Why != "" {
		r.Why = over.Why
	}
	if over.Fixtures.KnownBad != nil {
		r.Fixtures.KnownBad = append([]string(nil), over.Fixtures.KnownBad...)
	}
	if over.Fixtures.KnownGood != nil {
		r.Fixtures.KnownGood = append([]string(nil), over.Fixtures.KnownGood...)
	}
	r.Pattern = mergePattern(r.Pattern, over.Pattern)
	return r
}

func mergePattern(base, over Pattern) Pattern {
	r := base
	if over.Command != "" {
		r.Command = over.Command
	}
	if over.Subcommand != "" {
		r.Subcommand = over.Subcommand
	}
	if over.Subcommand2 != "" {
		r.Subcommand2 = over.Subcommand2
	}
	if over.ValueFlags != nil {
		r.ValueFlags = append([]string(nil), over.ValueFlags...)
	}
	if over.Flags != nil {
		r.Flags = append([]string(nil), over.Flags...)
	}
	if over.ArgPrefixes != nil {
		r.ArgPrefixes = append([]string(nil), over.ArgPrefixes...)
	}
	if over.FlagValues != nil {
		r.FlagValues = cloneFlagValues(over.FlagValues)
	}
	if over.ArgPaths != nil {
		r.ArgPaths = append([]PathArg(nil), over.ArgPaths...)
	}
	// AfterCD is a pointer precisely so an override can set it to false — a
	// bool field could only ever tighten the requirement, never lift it.
	if over.AfterCD != nil {
		v := *over.AfterCD
		r.AfterCD = &v
	}
	return r
}
