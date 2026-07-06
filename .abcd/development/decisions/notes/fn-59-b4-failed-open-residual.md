# B4 residual: the abcd guard's healthy-loader seal fail-open

Accepted residual + blast-radius bound for the two `failed_open` sites in the
abcd sibling guard's seal lookup. Recorded under fn-59.2 (R4b). This is a
**note**, not an ADR: it documents a gap we have deliberately decided to live
with, with a precise bound — not a reversible architecture choice.

## What is fail-open

`is_abcd_path_sealed` (`scripts/abcd/hooks/abcd_ralph_guard.py`) returns
`sealed=False, mode="failed_open"` at two sites, both reached only with an
otherwise **healthy** protected-path loader:

1. **`flowctl_path is None`** (`:913-917`) — no flowctl binary resolved, so the
   seal (task-status + receipt verdict) cannot be computed.
2. **`_query_task_status(...) is None`** (`:921-925`) — flowctl resolved but the
   owner task-status query produced no usable result (flowctl crash, timeout,
   unparseable output).

In both cases the guard returns "not sealed", and the seal-gated abcd-owned tier
branch (`:2767-2782`) then **allows** the Edit/Write. The trade is deliberate:
treating an unavailable / flaky flowctl as "everything sealed" would block every
abcd-owned path the moment flowctl hiccups, bricking the autonomous loop itself
(flowctl IS how Ralph advances). The fail-open keeps the loop alive at the cost
of not enforcing the seal during the outage.

R4a (malformed-stdin) is a **separate** decision and is now fixed (fail-closed
under `FLOW_RALPH=1`); it is not part of this residual.

## Why the existing mitigations do NOT cover it

Two mechanisms partially bound the exposure, but neither closes it for a healthy
loader with unavailable/flaky flowctl:

- **The fn-40 fail-closed fallback** (`_fallback_block_for_target`, `:2716-2727`)
  fires **only when the protected-path LOADER is degraded** (loader error or an
  empty required tier). A healthy loader with merely unavailable flowctl does
  NOT trip it — so the manifest-derived fallback set is not consulted on this
  path.
- **The present-only `.flow/bin/flowctl` block** (`:2738-2745`) and the
  **unconditional upstream tiers** (`REQUIRED_UPSTREAM` / `OPTIONAL_UPSTREAM`,
  `:2748-2762`) run BEFORE the seal logic and need no flowctl — so those classes
  stay blocked regardless. They do not protect the seal-gated abcd-owned tier.

So the residual is precisely: **seal-gated abcd-owned paths that are NOT also
covered by an unconditional tier, when the loader is healthy but flowctl is
unavailable or flaky.**

## Blast-radius bound (precise)

### Stays unconditionally blocked (NOT exposed by this residual)

- **`REQUIRED_UPSTREAM_PROTECTED_PATHS`** — `scripts/ralph/flowctl.py`,
  `scripts/ralph/ralph.sh`, `scripts/ralph/ralph_once.sh`,
  `scripts/ralph/hooks/ralph-guard.py`, `scripts/ralph/config.env`. Unconditional
  block; no seal, no flowctl needed.
- **`OPTIONAL_UPSTREAM_PROTECTED_PATHS`** — unconditional block when present.
- **`.flow/bin/flowctl`** — dedicated present-only block, seal-independent.
- **External plugin-cache paths** — matched and blocked before any seal lookup.

### Remains fail-open during a healthy-loader + flowctl-unavailable window

- The **seal-gated abcd-owned tier** (`ALL_ABCD_PROTECTED_PATHS`, ~100 paths:
  the `scripts/abcd/**` siblings, overlay machinery, the `protected_paths.d/**`
  fragments, tool configs, session/* runtime, etc.) **minus** the paths above.
- **Notably this INCLUDES** two paths that are in the fn-40 floor but are NOT in
  an unconditional upstream tier: **`scripts/ralph/flowctl`** (the abcd
  dispatcher) and **`scripts/abcd/hooks/abcd_ralph_guard.py`** (the guard
  itself). Their floor membership only protects them when the LOADER is degraded;
  under a healthy loader + no flowctl they fall into the seal-gated branch and
  fail open. This is the sharpest edge of the residual — during the window an
  in-loop edit to the dispatcher or the guard would NOT be blocked.

### Bound on severity

- **Trigger is narrow and self-limiting:** it requires a *healthy* loader (so the
  degraded-loader fallback is the common failure mode, and that one fails CLOSED)
  AND flowctl simultaneously unavailable/flaky. A flowctl outage also stalls the
  loop's own task advancement, so the exposed window tends to be short rather
  than persistent (though, per the next point, the fail-open allow itself is
  silent while it lasts).
- **Silent at the hook boundary (no compensating observability):** the
  `failed_open` reason IS recorded in the `SealLookupResult.diag` object — but
  ONLY on the internal return value (visible in direct unit tests of
  `is_abcd_path_sealed`). The runtime allow path does NOT surface it: when
  `seal.sealed` is false, `_check_protected_edit_or_write` returns `(0, "")`
  (discarding `seal.diag`), and `main()` prints nothing for an rc==0 result. So
  for the normal agent Edit/Write path the fail-open is **silent** — the edit is
  allowed with no stderr trace. This is part of what is accepted here, NOT a
  mitigation. If observability is later wanted, the closure is to emit the
  `failed_open` diag to stderr (or a receipt) on the unsealed-allow branch;
  that is a separate, non-blocking improvement, not a fix to the seal contract.
- **Not a privilege-escalation primitive:** the exposed paths are abcd's own
  tooling; the residual is "the seal is not enforced during a flowctl outage",
  not "an attacker gains new capability". The threat model here is an autonomous
  loop editing its own tooling mid-outage, not an external adversary.

## Decision

**Accepted as residual.** Tightening these two sites to fail-closed under
`FLOW_RALPH=1` was considered (mirroring the R4a malformed-stdin decision) but
rejected for R4b: unlike a malformed stdin payload (anomalous, never sent by the
harness), a transiently unavailable/flaky flowctl is an *expected* infra
condition, and failing closed there would re-introduce the infra-flakiness
false-block that bricks the loop — the exact failure the fail-open exists to
avoid. The narrow trigger and the unconditional protection of the upstream
tiers + `.flow/bin/flowctl` bound the exposure to abcd-owned tooling during a
short window — accepting that the fail-open allow is silent at the hook boundary
(see "Bound on severity"), which an optional follow-up could make observable
without changing the seal contract.

If the dispatcher (`scripts/ralph/flowctl`) or the guard
(`scripts/abcd/hooks/abcd_ralph_guard.py`) specifically warrant
healthy-loader-independent protection in future, the cleanest closure is to add
them to a seal-independent unconditional block (the way `.flow/bin/flowctl` is
present-only blocked), NOT to fail the seal closed on flowctl unavailability.
