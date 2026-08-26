---
id: spc-42
slug: every-change-starts-in-its-own-worktree-the-primary-checkout
intent: itd-148
---
# every-change-starts-in-its-own-worktree-the-primary-checkout

## Summary

Delivers itd-148: the primary checkout of an abcd-managed repo becomes a
read-only surface, every change starts in a per-change worktree, and the
worktree population is swept — merged trees removed on proof, abandoned trees
surfaced with a dossier the human rules on. All behaviour lives in the
transport-agnostic core; the CLI and the plugin markdown surface are thin
front doors, and the enforcement hooks are thin callers into the binary.

## Scope

1. **Core worktree package** (`internal/core`): tree classification (primary
   vs worktree, resolved via the git common directory, never a path prefix),
   worktree create/enter metadata, sweep classification, dossier composition,
   and the visibility ledger.
2. **Verbs**: `abcd worktree` (status render: which worktrees exist, who is
   in them, what each is about), `abcd worktree new <slug>`, `abcd worktree
   sweep`, and `abcd worktree check` (the hook entry point; exit code carries
   the block verdict). Bare `abcd worktree` performs zero writes.
3. **Hooks**: a PreToolUse-shaped hook routing the host's file-edit and shell
   tools through `worktree check`; a SessionStart contribution that surfaces
   sweep candidates and performs proven-merge removals; a git pre-commit
   backstop installed in the primary checkout via the ahoy defaults.
4. **Scaffolding**: the prepare/ahoy path installs the hooks and the
   pre-populated private banlist floor into abcd-managed repos.
5. **Record changes**: the AGENTS.md Concurrent sessions rewrite (blast-radius
   rule), and resolution of the bundled issues in the shipping changes
   (iss-370, iss-213, iss-2608230847432285, iss-2608230957104179,
   iss-2608210738378295).

## Approach, by acceptance criterion

**AC 1 — the mutation block with allow-set.** `worktree check` receives the
tool input (cwd, file path or command) and answers allow/refuse. Refusal only
when: the tree resolves to the primary checkout of an abcd-managed repo AND
the write targets a tracked file or the command is history-moving. The
allow-set is explicit: gitignored/untracked paths, `.abcd/.work.local/`,
fetch/fast-forward of the default branch, `git worktree` administration, and
the capture carve-out (a new timestamp-named file under
`.abcd/work/issues/open/` only). Shell interception is command parsing at the
guard-registry rung, stated as mitigation, not filesystem guarantee; the
pre-commit backstop (same check, commit-time) catches what parsing misses.
Refusal messages name the worktree route. Tests: table-driven check tests per
allow/refuse case; hook wiring exercised via the smoke harness.

**AC 2 — guard seeding and peer visibility on entry.** Seeding runs on
detection, not only creation: any hook fire that finds itself in a worktree
of an abcd-managed repo whose local tier lacks the name-guard layer seeds the
pointer to the primary checkout's store. The floor's four categories
(home/user paths, machine hostnames, personal email, real surname) are
populated by one-time setup prompt or from the user-level home; values are
machine-local, never derived silently, never committed. Entry/exit events
append to a visibility ledger in the primary checkout's local tier;
peer-session hook fires inject unseen entries (the record-then-inject shape
the rules loader already uses). Delivery is at-next-prompt by design.

**AC 3 — proven-merge removal.** Merge proof, in order: fast path, tip
reachable from the default branch; forge path, the remote PR's recorded merge
state where a PR exists; local fallback, patch-equivalence (`git cherry`
/ patch-id) of the branch's commits against the default branch. Squash and
rebase merges are covered by the forge and patch paths — sha reachability
alone cannot see them. Unprovable means dossier, never removal. Removal runs
at SessionStart or explicit sweep only, never as a side effect of other
verbs; residue tidy-work delegates to itd-118's mechanics.

**AC 4 — the abandoned dossier.** Candidate: unmerged, no open PR, no
activity for 14 days (repo-configurable). The dossier leads with dirty state,
then branch subjects, diffstat, referenced record ids, last activity, and a
recommendation whose vocabulary includes capture-or-commit-first for dirty
trees. Removal is per-item human-confirmed; `git worktree remove` without
force is the dirty-tree backstop. Orphan-capture adoption rides the sweep:
uncommitted ledger files in the primary checkout are offered into a change.

**AC 5 — mint visibility.** Sequential-family mints append a
family-and-checkout event to the same visibility ledger, injected to peers at
next hook fire — the mechanical form of the AGENTS.md convention, standing
until the adr-45 timestamp migration (schedule unchanged by this spec).

**AC 6 — zero cost when idle.** Bare verbs stay zero-write; sweep writes are
confined to SessionStart and explicit sweep. A repo with no worktrees and no
peers takes one tree-classification stat call in the hook path and nothing
else. Read-only verbs are untouched by construction (the block sits on write
paths only), verified by tests asserting no writes from bare invocations.

## Out of scope

Coordination claims (itd-33), post-merge remote/branch residue mechanics
(itd-118, consumed), presence leases (iss-2608220750029993), the timestamp
migration itself (adr-45 schedule holds).

## Delivery

Staged PRs, each preflight-clean with tests watched fail first: (1) core
classification + `worktree check` + hooks; (2) seeding + visibility ledger;
(3) sweep (merged half, then dossier); (4) scaffolding + AGENTS.md rewrite +
bundled-issue resolutions with `Resolves:` trailers in their fixing changes.
