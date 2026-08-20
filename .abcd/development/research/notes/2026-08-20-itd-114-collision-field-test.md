# itd-114 field test: four minters, one morning — pre-registered

The morning of 2026-08-20 is a natural experiment for
[itd-114](../../intents/drafts/itd-114-abcd-mints-collision-proof-record-ids-across-parallel-agents.md)
(collision-proof record ids), run deliberately after two live collisions in the
preceding day (the standing record is iss-330; the second occurrence is logged
in the round-2 entry of `.abcd/work/DECISIONS.md`). The maintainer's decision,
same date: **itd-114 is pulled forward** — this note is pre-registration for
its planning interview, predictions written before the outcomes.

## The set-up (as of ~08:00, recorded before outcomes)

Concurrent minters against the ledger's next-free-integer allocator:

1. **A reconciliation agent** (isolated worktree) landing hunt A's round 1,
   carrying the stranded range iss-305..325, under a defensive protocol:
   just-in-time id verification twice (at ledger commit and at push, after
   fresh fetches), a scan of unmerged `bughunt-*` branch ledgers, and
   union-merge treatment of the append-only surfaces (CHANGELOG
   `[Unreleased]`, `DECISIONS.md` tail).
2. **Bughunt cloud routine 1** — running now, branch not yet pushed; a round
   historically mints a ~20-id block.
3. **Bughunt cloud routine 2** — fires at 11:00 local.
4. **The maintainer's local agent** — working the codebase, may capture ad hoc.

Ledger state at pre-registration: main's highest id is iss-343; the range
305..325 is unallocated on main and carried only by the reconciler.

## The allocator's observed behaviour (evidence, not prediction)

The mint is **max+1, not gap-filling**: on 2026-08-19 captures minted iss-326
onward while 305..325 were absent everywhere — the allocator never reused the
gap. Corollary: a *gap* range is structurally safe against fresh mints; only
two minters starting from the **same max** collide.

## Pre-registered predictions

- **P1 — the reconciler's range survives.** 305..325 sits below main's max, so
  max+1 minters cannot take it. Falsified if any concurrent mint lands in
  305..325.
- **P2 — hunt-vs-hunt collision at 344+, conditional.** Both hunts mint from
  their checkout-time max. If hunt 1 has not *merged* by the time hunt 2
  checks out (~11:00), both start at 344 and the day's third collision
  follows. If hunt 1 merges first, hunt 2 starts above it and no collision
  occurs. This is the experiment's sharpest test: it predicts collision from
  *merge timing alone*, which is exactly the property a collision-proof mint
  must not have.
- **P3 — detection stays late.** Any collision is caught by `issue_id_unique`
  at the first CI run that sees both sides (PR or merge-queue), never at
  capture time — the capture-time validator does not exist (itd-84's future
  rung; iss-265). Cost per occurrence so far: a renumber-and-remap commit, or
  a full release re-cut when frozen receipts are involved.
- **P4 — the defensive protocol holds but does not scale.** The reconciler's
  double-check avoids collision at the cost of a page of standing
  instructions and two extra fetch/verify rounds — overhead every future
  parallel task would have to re-carry by hand. (The protocol is itself the
  argument: safety should live in the mint, not in every minter's prompt.)

## Results — TO BE GRADED after the morning resolves

_Pre-registered and deliberately empty until the reconciler's PR, hunt 1's
round, and hunt 2's 11:00 round have all landed or failed. Grade each
prediction CONFIRMED / FALSIFIED with the evidence (branch, PR, run), count
collisions and near-misses, and record where each was detected and what it
cost. This section is the input the itd-114 planning interview opens with._
