---
id: adr-16
slug: fn43-autodrain-boundary-and-gate-defaults
status: superseded
date: 2026-06-11
supersedes: null
superseded_by: adr-27
related_intents: [itd-53]
related_rfcs: []
related_adrs: []
---

# ADR-16: Autodrain fires at the Ralph post-iteration edge only (no Claude Code hook); the gate reports, never blocks; the drain is cost-bounded by processed entries

> Superseded by [ADR-27](0027-autonomous-run-pluggable-seam.md) — the autonomous
> run is a pluggable seam rather than a Ralph port, so the Ralph
> post-iteration edge this autodrain boundary depends on no longer exists.

## Context

fn-28 shipped the review-queue substrate: closing a spec moves its intent to
`shipped/` and the **pure** close-hook (`intent_lifecycle.reconcile` — a data
function, no subprocess/oracle/flowctl, because it runs inside the autonomous
loop) *enqueues* a fidelity-review entry. The review itself only ran when
someone manually invoked `flowctl review-queue drain`, so owed reviews
accumulated with nothing paying the queue down — "shipped" had drifted from
"audited". itd-53 resolved this with an opt-in drainer plus a consistency
gate, but left four open questions: which safe boundary drains, does the gate
report or block, how is drain cost bounded, and how does itd-50's policy layer
relate. fn-43 settled all four; this ADR records the settlements.

Constraints already locked when fn-43 was decided:

- The close-hook stays pure — running an oracle on the loop's critical path
  re-introduces the stall itd-53 exists to avoid.
- Every Claude Code hook in `hooks/hooks.json` runs under `timeout: 5`
  (seconds), while a single oracle review is "seconds-to-minutes"
  (`review_queue.run_one`).
- `drain()`'s two-lock + claim_id protocol and `deferred`-not-failed
  no-backend semantics (fn-28) are safety-hardened and untouched — fn-43
  wraps and triggers, never re-implements.
- `ralph.sh` is upstream-vendored, re-vendored by `/flow-next:ralph-init`,
  and runs `set -euo pipefail`.

## Decision

1. **The autodrain trigger is the Ralph post-iteration edge ONLY.** The
   `ralph.sh` loop body, after the timeout-wrapped `claude` subprocess
   returns — plain shell, no 5s budget, a boundary that already runs
   minute-long reviews. The Claude Code SessionEnd/Stop hook is an explicit
   **non-goal**: under the 5s hook timeout even `max_reviews=1` is SIGKILLed
   mid-review, leaving the queue entry `running` with a live claim until
   `stale_after_secs` (1800s) — a regression, not a feature. Background
   dispatch from a hook is also unreliable (the hook process is reaped at
   session teardown). So fn-43 makes **no `hooks/hooks.json` change and no
   `hooks/README.md` hook entry**. The cheap gate (no oracle dispatch) may
   additionally run at pre-commit/CI.
2. **Delivery is overlay-managed, never a fork.** The loop-body edit ships as
   the `patch:ralph-autodrain-trigger` manifest entry (precedent:
   `patch:ralph-quota-window`), anchored on stable code text with an
   `upstream_resolution_criterion`, re-applied by `/abcd:ralph-up` after every
   re-vendor. The invocation is `|| true`-guarded so a drain failure never
   aborts the `set -euo pipefail` loop.
3. **Default OFF, with an explicit off-path cost contract.** `review.autodrain`
   lives under `.abcd/config.json["review"]` (read via
   `config_io.read_config` with the safe-read pattern — absent file/block/key
   all mean off, never a crash), distinct from the `.flow/config.json`
   `review.body_max_bytes`/`render_max_bytes` block. With autodrain off, the
   Ralph edge still invokes the dedicated `autodrain` verb once per iteration:
   one cheap config-read subprocess that early-exits 0 (no oracle, no claim,
   no queue I/O beyond the config read). "Never fires" is a guarantee about
   `drain()`, not about the subprocess; the shell `|| true` is
   belt-and-braces, not load-bearing.
4. **The drain is cost-bounded, counting PROCESSED entries.** `drain()` gains
   a `max_reviews` cap (config `review.autodrain_max_reviews`, default 1) so a
   boundary firing can never fan out the entire backlog as oracle calls. The
   counter counts processed entries, **not successes** — K consecutive
   deferred-producing entries with `max_reviews=1` still process exactly 1 —
   so an unreachable backend cannot turn the cap into an unbounded retry loop
   within one firing. The manual `flowctl review-queue drain` stays unbounded
   by default (`--max-reviews` optional).
5. **The gate reports; it never blocks by default.** RC001/RC002 default to
   warn, RC003 to info; clean/info/warn exit 0. Blocking a commit on an
   unaudited intent contradicts the loop-purity intent, so blocker promotion
   is config-only (`lint.severity_overrides`), never the default posture.
6. **RC001's population is bounded strictly post-fn-12.** Review-absent fires
   by default only for intents whose `spec_id` numeric part is strictly
   `> 12` — an fn-12-linked intent is itself LEGACY, because the fidelity
   reviewer did not exist before its own spec closed. Older shipped intents
   (currently itd-27, itd-28) are a known finite backlog surfaced only behind
   `--include-legacy`; flagging them by default would train operators to
   ignore the signal. RC002 (`NOT_MET`) always fires — a failed audit is
   actionable regardless of era.
7. **itd-50 rides on this drainer and is out of scope.** The drainer just
   runs reviews; what a `NOT_MET` verdict *triggers*
   (loop-toward-acceptance, unachievable→replan) is itd-50's policy layer,
   built on top later — the drainer ships no hooks for it.

## Alternatives Considered

- **Claude Code SessionEnd/Stop hook as the drain point.** Rejected: the 5s
  hook timeout SIGKILLs any real review mid-flight, stranding a `running`
  claim for up to 1800s; background dispatch from a hook is reaped at session
  teardown. A boundary that corrupts queue state on every firing is worse
  than no boundary.
- **Drain inside the close-hook.** Rejected outright — `reconcile` is pure by
  design (itd-53's own premise); an oracle call on the loop's critical path
  stalls headless runs whenever the backend is unreachable.
- **Blocking gate by default.** Rejected: blocking commits on unaudited
  intents punishes the autonomous loop for a backlog it did not create.
  Report-first, promotion config-only.
- **Unbounded autodrain (drain the whole backlog per firing).** Rejected: one
  boundary firing fanning out N oracle calls is exactly the stall the
  boundary choice avoids. Cost-bound mandatory; default 1.
- **RC001 over the whole `intents/shipped/` corpus.** Rejected: every
  pre-reviewer intent fires forever, the finding becomes noise, operators
  learn to ignore it. Bounded population keeps the signal actionable; legacy
  stays reachable behind a flag.

## Consequences

- "Shipped" can now mean "shipped and audited" without any change to loop
  purity: enqueue stays pure, drain is a deliberate boundary step, and the
  gate tells the truth of record either way.
- Default-off means the back-edge only closes for operators who opt in; the
  gate (always available, oracle-free) is the floor everyone gets.
- One subprocess per Ralph iteration is the standing off-path cost — accepted
  as negligible against the loop's existing per-iteration work, and recorded
  here so it is never "discovered" as a leak.
- The post-iteration edge only fires while Ralph runs; a repo driven purely
  by interactive sessions never autodrains. Accepted: the manual verb and the
  (possible) pre-commit/CI gate cover that mode, and a future native upstream
  hook seam retires the patch per its `upstream_resolution_criterion`.
- The overlay patch adds one more `ralph.sh` splice to maintain across
  re-vendors (anchor: the `iter_end_ts` capture line; fail-closed on
  `no_anchor`).
- The strict `N > 12` bound leaves itd-27/itd-28 permanently legacy unless
  someone runs their reviews; `--include-legacy` keeps them visible.

## Related Documentation

- [`../../brief/05-internals/03-configuration.md`](../../brief/05-internals/03-configuration.md) — the `review.autodrain` config keys
- [`../../brief/05-internals/06-lint.md`](../../brief/05-internals/06-lint.md) — § 1 RC code registry (RC001–RC003 delivered, RC004–RC005 reserved)
- [`../../intents/drafts/itd-53-review-queue-auto-drain-fidelity-gate.md`](../../intents/shipped/itd-53-review-queue-auto-drain-fidelity-gate.md) — the intent whose open questions this ADR settles
- `../../../../scripts/abcd/overlay/README.md` — overlay patch mechanism (`patch:ralph-autodrain-trigger`)
