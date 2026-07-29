package audit

import (
	"path/filepath"

	"github.com/REPPL/abcd-cli/internal/fsutil"
	"github.com/REPPL/abcd-cli/internal/gitutil"
)

// threeTierLayout checks the committed three-tier .abcd/ layout: the durable
// record (.abcd/development/) and the shared working tier (.abcd/work/) must be
// present, the local-ephemeral tier (.abcd/.work.local/), when present, must
// be gitignored so per-worktree state never leaks into history, and the
// local tier's conventional artifacts (NEXT.md, scratch/, logs/) must not sit
// directly in a committed tier — a handover file carrying machine-local detail
// in .abcd/work/ is committed, pushed, and public before anything flags it.
//
// Presence of .work.local is NOT required — it is created on demand and a fresh
// clone has none. Requiring it would flag every clean checkout. The load-bearing
// assertion is "if present, ignored", which is what stops decisions and logs
// leaking. (Divergence from the plan's literal "present and gitignored",
// recorded in DECISIONS.md.)
type threeTierLayout struct{}

func (threeTierLayout) Meta() RuleMeta {
	return RuleMeta{
		ID:         "three-tier-layout",
		Severity:   SeverityError,
		Fix:        "create the missing .abcd/ tier; ensure .abcd/.work.local/ is listed in .gitignore",
		PolicyInfo: "the three-tier layout separates the durable record, shared working state, and per-worktree ephemera; the local tier must be gitignored so it never merge-conflicts or leaks",
	}
}

func (threeTierLayout) Where(Context) bool { return true }

func (threeTierLayout) Eval(ctx Context) ([]Finding, error) {
	var out []Finding

	for _, tier := range []struct{ rel, label string }{
		{".abcd/development", "durable-record tier .abcd/development/"},
		{".abcd/work", "shared-working tier .abcd/work/"},
	} {
		// A tier is a directory: a regular file at the tier path does not satisfy
		// the convention, so check the type, not mere presence. (IsDir follows a
		// symlink, so a symlink-to-directory does satisfy it — acceptable for a
		// layout check, which is not the owned-store trust boundary.)
		isDir, err := fsutil.IsDir(filepath.Join(ctx.RepoRoot, filepath.FromSlash(tier.rel)))
		if err != nil {
			return nil, err
		}
		if !isDir {
			out = append(out, Finding{
				RuleID:   "three-tier-layout",
				Severity: SeverityError,
				File:     tier.rel,
				Message:  "missing the " + tier.label + " (must be a directory)",
			})
			continue
		}

		// Local-tier artifacts in a committed tier: NEXT.md, scratch/ and logs/
		// are the local-ephemeral tier's conventional contents, so their presence
		// directly under a committed tier is a placement error of the leak class —
		// per-worktree ephemera about to enter (or already in) history. Presence
		// is checked on the filesystem, like the tiers themselves: an untracked
		// NEXT.md in .abcd/work/ is one `git add -A` from being committed.
		for _, artifact := range []struct{ name, kind string }{
			{"NEXT.md", "handover file"},
			{"scratch", "scratch directory"},
			{"logs", "logs directory"},
		} {
			rel := tier.rel + "/" + artifact.name
			present, err := fsutil.Exists(filepath.Join(ctx.RepoRoot, filepath.FromSlash(rel)))
			if err != nil {
				return nil, err
			}
			if present {
				out = append(out, Finding{
					RuleID:   "three-tier-layout",
					Severity: SeverityError,
					File:     rel,
					Message:  "local-tier " + artifact.kind + " " + artifact.name + " found in the " + tier.label + " — per-worktree ephemera must never enter a committed tier",
					Fix:      "move " + rel + " to the local-ephemeral tier .abcd/.work.local/",
				})
			}
		}
	}

	// The local tier: only an issue when it is present but not gitignored, and
	// only when git can actually answer — git-absent is "cannot tell", never a
	// silent pass claiming it is ignored.
	localRel := ".abcd/.work.local"
	localPresent, err := fsutil.Exists(filepath.Join(ctx.RepoRoot, filepath.FromSlash(localRel)))
	if err != nil {
		return nil, err
	}
	if localPresent && gitutil.InRepo(ctx.RepoRoot) {
		if !gitutil.IsIgnored(ctx.RepoRoot, localRel+"/") {
			out = append(out, Finding{
				RuleID:   "three-tier-layout",
				Severity: SeverityError,
				File:     localRel,
				Message:  "the local-ephemeral tier .abcd/.work.local/ is present but not gitignored — its contents would be committed",
			})
		}
	}

	return out, nil
}
