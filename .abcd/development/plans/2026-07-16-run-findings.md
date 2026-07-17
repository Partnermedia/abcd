# Run findings — planned-intent drain (manual test of the future `/abcd:run`)

**Status:** live findings log for the run driven by
[`2026-07-12-abcd-run-protocol.md`](2026-07-12-abcd-run-protocol.md) over
[`2026-07-16-planned-intent-drain-run.md`](2026-07-16-planned-intent-drain-run.md).
This run doubles as a manual test of the C-phase protocol before it becomes
`commands/abcd/run.md` (itd-29 reconciliation). Every burst appends here, as
findings occur, not at burst end.

Classification: **protocol gap** (the generic protocol needs a rule) ·
**plan gap** (the PLAN template needs a field) · **tooling gap** (the abcd
binary should do what the agent improvised) · **human stop** (a boundary
autonomy must not cross).

---

## Burst 1 (2026-07-16, M1 itd-89/spc-4 — shipped as PR #83, then STOP)

### F1 — Shared working tree with a concurrent agent · protocol gap + human stop

The plan's coordination rules were branch-scoped ("never touch a branch you
did not create"), but the concurrent bug-hunter edited the **same checkout's
working tree** while the run's branch was checked out — uncommitted foreign
edits accrued mid-burst across five packages. Branch discipline cannot protect
a shared checkout: any `git checkout`/`add`/`stash` risks sweeping in or
clobbering a second writer's work.

- The run survived because the item's files happened not to overlap
  (path-scoped staging), then stopped — correctly — rather than continue.
- **For `/abcd:run`:** (a) the protocol must require an exclusive checkout —
  own worktree or clone — as a *precondition*, not a coordination note;
  (b) cheap detector: snapshot `git status --porcelain` at burst start and
  re-check before every commit/checkout; any foreign delta is a STOP;
  (c) which agent relocates when two share a repo is a **human stop** — the
  agent cannot adjudicate ownership of a checkout it doesn't own.

### F2 — Fidelity-review provenance hashes are undocumented · tooling gap

`intent review ingest` hard-requires `policy.rubric_hash` + `prompt_hash`, but
nothing states what the host must hash. The agent reverse-engineered the
convention from itd-80's ingested provenance line plus git history (rubric =
reviewer agent file, prompt = emitted request file). A wrong-but-non-empty
hash would ingest fine — the requirement is currently un-checkable theatre.
- **For `/abcd:run`:** the emit path (`spec close` / `intent review`) should
  print or embed the expected hashes in the `.request.md` it writes; ingest
  could then verify instead of merely requiring non-empty.

### F3 — "delivered" input to the fidelity reviewer needs judgment · plan gap

The request file says "delivered: the diff/commit range that realised spc-4
(host supplies the range)". For record-catch-up items the implementation
merged long ago across many PRs, so "the range" is not mechanical; the agent
substituted "repo state at HEAD plus the load-bearing files". Worked, but a
PLAN driving record catch-up should state per item what "delivered" means.

### F4 — Named reviewers are prompt files, not harness agent types · protocol gap (resolved pattern)

`ruthless-reviewer`/`security-reviewer`/`intent-fidelity-reviewer` exist as
`agents/*.md` role definitions, not as dispatchable agent types in this
harness. The working pattern: spawn a general-purpose subagent whose first
instruction is to read and adopt the role file, and to return the terse
structured verdict only. The protocol should name this fallback explicitly so
"reviewer unavailable" is never silently equated with "skip the lens".
Corollary that worked well: the correctness reviewer independently chose a
detached worktree to escape the foreign uncommitted edits (F1) — reviewers
should always review the *committed* range, never the working tree.

### F5 — Run-level journal entries have no schema slot · protocol gap

`run-journal.json` is keyed on item stable-ids; the burst-1 STOP was run-level
(shared checkout), not item-level. The agent minted a `"RUN"` pseudo-id. Fine,
but the protocol should bless a reserved key for run-level outcomes so strike
counting never confuses an environmental STOP with an item failure.

### F6 — Worktree resume loses the handoff state · protocol gap

`.abcd/.work.local/` is per-worktree by design, so resuming in a fresh
worktree (the F1 remedy) starts blind: NEXT.md and run-journal.json stay in
the abandoned checkout. The restart prompt had to include a manual bootstrap
copy. `/abcd:run` needs a state-location rule for worktree moves (copy-once on
worktree creation, or a run-state directory addressed by repo, not checkout).

### F7 — Redundant review-emit step in the protocol flow · protocol gap (minor)

The plan said "close, then run `abcd intent review itd-89`" — but `spec close`
already emits the OWED receipt and request file. The extra `intent review`
call is harmless (idempotent re-emit) but the canonical flow should be stated
as: close → (auto-emit) → delegate reviewer → ingest.

### F8 — Interval semantics: 90m is not cron-expressible · tooling note (harness)

The harness loop rounded `/loop 90m` to 2h. Immaterial here (bursts are
self-bounded at 45m), but `/abcd:run` documentation should not promise
arbitrary intervals if the scheduling substrate quantises them.

### What ran fully autonomously in burst 1 (the positive finding)

State recovery → item pick → TDD gap-fill tests → gate → spec authoring →
lifecycle close → host-delegated fidelity review → deterministic ingest →
path-scoped commits → two PROMOTE reviews → push → PR, with zero human input.
The two places autonomy correctly ended: the shared-checkout collision (F1)
and the standing do-not-merge policy on PRs. Both are the right shape for
permanent human stops; nothing else in M1 needed one.
