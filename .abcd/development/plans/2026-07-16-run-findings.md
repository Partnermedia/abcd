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
## Burst 2 (2026-07-17, manual test B — delegated implementation, M2 itd-40)

**Setup:** orchestrator (Fable 5) keeps state recovery, item picking,
journaling, gates, record commits, and PRs; implementation is delegated — a
deliberate, recorded deviation from the protocol's "never delegate
implementation" — to one Opus 4.8 max-effort worker per item in an isolated
worktree, with the orchestrator re-running `make preflight` on the output.
Reviews stay fresh-context Opus subagents. Structured via the harness
Workflow tool (one implementation workflow, one review workflow per item).

### F9 — Run-instruction premises go stale against a moving repo · protocol gap

The burst-2 instructions asserted two plan files were "currently untracked"
and must be committed on the first item's branch. By burst start they were
already committed — inside the concurrent hunter's *unpushed foreign commit*
on the old item branch, which the orchestrator must not touch. The letter of
the instruction was unsatisfiable; its intent (durable record on
`origin/main`) was satisfied by committing copies on the item branch and
flagging the future both-added merge. **For `/abcd:run`:** treat every factual
premise in run instructions (file X is untracked, branch Y is merged) as a
hypothesis to re-verify at burst start; on divergence, act on the recorded
*intent* and journal the divergence rather than STOPping or blindly obeying.

### F10 — Unpushed foreign commits are a coordination blind spot · protocol gap

Burst 1's F1 covered foreign *uncommitted* edits in a shared checkout. Burst 2
found the successor hazard: the hunter's finished work sits as unpushed
commits on a local branch, touching `internal/core/ahoy` (M2's package) and
`internal/core/capture` (M3's). Origin-based collision checks ("an open PR or
fresh commit by the other agent") never see these; branch-scoped discipline
protects the branch but not the eventual merge. Remedy used here: extract the
foreign commits' changed *functions* (`git diff` hunk headers) and hand the
worker an explicit no-overlap list, downgrading the plan's package-level
collision fear to a function-level guard — plus a PR-body note where the same
file is added on both sides. **For `/abcd:run`:** at burst start, enumerate
local branches ahead of origin, diff them, and feed the touched-function set
into every worker's guard; a *required* edit inside a foreign-touched function
is the skip trigger, not mere package co-location.

### F11 — The plan names a `spec create` verb that does not exist · plan gap + tooling gap

M2's instructions offered "`abcd intent plan itd-40` path or `spec create` +
`intent link`". Neither literal path works for an intent that is already in
`planned/` with `spec_id: null` (the record-catch-up shape this whole run
exists for): `intent plan` refuses non-drafts, and there is no `spec create`
CLI verb — `spec.Create` is core-internal, reachable only through `intent
plan`. Resolution: transient `git mv` planned→drafts, then `intent plan`
(mints spc-5 under the mint lock, relinks, moves back); net churn is the
`spec_id` write, record lint green throughout. **For `/abcd:run`:** plans must
name only verbs that exist (a dry-run of every named verb at plan-authoring
time would have caught this); tooling-wise, either `intent plan` should accept
a planned+unlinked intent, or `spec create` should exist for record catch-up.

### F12 — The protocol's changelog-fragment rule contradicts the repo · protocol gap (minor)

The protocol mandates `changelog.d/<slug>.md` fragments; this repo has no
`changelog.d/` and appends to `CHANGELOG.md` directly (AGENTS.md's rule). The
delegated worker noticed and followed the repo, which is correct — AGENTS.md
overrides — but the protocol text should say "the repo's changelog mechanism,
as the PLAN records it" instead of hard-coding one mechanism.

### F13 — Delegated commits need an identity check · protocol gap

The Opus worker's commit arrived authored "Alex Reppel <...>" although the
repo's `git config user.name` is `REPPL` — the worker (or its sub-shell)
resolved identity from somewhere other than the repo config, and nothing in
the flow would have caught it before push. The orchestrator caught it by
inspection and amended the unpushed commit's author. **For `/abcd:run`:** when
implementation is delegated, the orchestrator's post-worker gate must include
an identity check on every new commit (the repo even ships one — `abcd ahoy
identity-check`); author metadata is part of the record, not a cosmetic.

### Delegation observations (test B vs burst 1's write-it-yourself baseline; running log)

- **TDD evidence survives delegation.** The worker returned verbatim
  watched-fail excerpts for both genuine gaps, and — unprompted — split
  "gap-filled (watched-fail)" from "met-already (characterization test
  added)", exactly the honesty the protocol wants. Quality of evidence is
  indistinguishable from burst 1's first-person TDD.
- **The worker respected every guard:** function-overlap list (edit landed in
  `newAhoyCommand`'s render closure, outside all guarded bodies), iss-101
  untouched, hermetic tests (temp HOME), no new deps, no record files touched,
  single atomic commit with a why-shaped body.
- **Orchestrator re-verification caught real deltas:** the author-identity
  drift (F13) — nothing the worker reported; found only by inspecting the
  commit. Re-running the gate myself reproduced green (test cache warm from
  the worker's own run, so the re-run cost seconds, not minutes).
- **Cost:** implementation phase = 1 Opus/max worker, ~95k subagent tokens,
  ~9.5 min wall, 47 tool uses, for a 200-line diff (8 production lines, 188
  test lines) — heavier per line than burst 1's in-context work, but the
  orchestrator's context stayed small enough to run the whole record/lifecycle
  half without compaction risk.
