# fn-82.7a — Pre-claim doctor gate: SUBSUMED

Probe-only decision (R7). The fn-34.5 doctor gate
(`scripts/abcd/tools/gate.py::run_gate` :168) was never wired into the loop.
fn-71 has since made `/abcd:session` the default. This note records whether a
pre-claim readiness check is still needed, per the two pinned outcomes. **No
wiring lands in fn-82.**

## The constraint that forced a probe (from plan review)

`gate.py::apply_context(reduced, context)` (:149) downgrades `fail`/`infra` to
`warn` under `context == "session"` — "no backoff loop":

```python
if context == "session":
    if reduced in ("fail", "infra"):
        return "warn"   # no backoff loop — downgrade both to warn.
    return reduced
```

So wiring `run_gate` through the CURRENT session context **cannot** produce a
PAUSE: a `fail`/`infra` doctor outcome becomes `warn` (exit 0, non-blocking).
That is why this is a probe, not a wiring task — the decision is between
(i) SUBSUMED and (ii) NEEDS-NEW-POLICY-CONTEXT.

## Outcome: SUBSUMED

The session seam **already provides an equivalent pre-claim readiness check that
preserves PAUSE semantics** — at a different lifecycle point than gate.py, and
crucially WITHOUT the session-context downgrade.

**Component:** `scripts/abcd/session/launcher.py::evaluate_launch` driven by
`scripts/abcd/session/composer.py::compose_real_readiness` /
`RealReadinessProvider`.

**Evidence — the fail-closed pre-spawn gate:**

- `evaluate_launch` (`launcher.py` :183-219) runs BEFORE the detached
  supervisor spawns the worker. It validates config/budget shape, then the
  enforcement-stack readiness, and returns `LaunchDecision(permitted=False, …)`
  the instant ANY enforcement gate is not-ready:

  ```python
  missing = report.missing()
  if missing:
      return LaunchDecision(
          permitted=False,
          reason="enforcement stack incomplete: " + ", ".join(missing),
          missing=missing,
      )
  ```

- The launch path uses the REAL provider (`compose_real_readiness`,
  `composer.py` :140), which merges the four module `readiness_signals()`
  (watchdog budget validity, verdict_gate, boundary/`guard_preflight`,
  paths/`artifact_hygiene`). The merge is **fail-closed on every conflict** —
  a duplicate gate, an out-of-ownership emission, a non-bool state, a missing
  gate, or ANY provider exception forces the owned gate(s) NOT-ready
  (`RealReadinessProvider.report`, :82-137). `compose_real_readiness` also
  asserts the ownership partition equals `ENFORCEMENT_GATES` exactly, raising
  `UnderWired` on drift.

- A not-ready decision makes the launcher **refuse a real launch** — the caller
  never spawns an unenforced worker (`launcher.py` module docstring :8-16; the
  boundary-absent variant of the refusal surfaces `BARE_OPT_IN_HINT` :121).
  This is the PAUSE-equivalent gate.py could not deliver: readiness is a HARD
  gate here, not a `warn`-downgraded advisory.

**Why the other session components are NOT the pre-claim check** (they gate
different lifecycle points, so naming them would be wrong):

- `verdict_gate.py` — the SUPERVISOR-side *task-advance* permit (`$FLOWCTL done`
  needs a trusted SHIP). Post-work, not pre-claim.
- `watchdog.py` — a *runtime* budget / cancel-storm monitor over the child's
  live log stream. Post-spawn, not pre-claim.
- `boundary.py` — a *runtime* OS egress/fs isolation boundary wrapping the child
  argv. Post-spawn, not pre-claim.

Each of these DOES contribute a `readiness_signals()` to the pre-spawn merge
above — so their readiness is checked pre-claim — but the readiness GATE that
turns a not-ready signal into a refusal is `evaluate_launch`, and it is
fail-closed, unlike gate.py's session context.

## Consequence

- No pre-claim wiring of `gate.py::run_gate` is needed; the readiness gate the
  fn-34.5 idea reached for already exists at `evaluate_launch` with the correct
  (fail-closed) semantics.
- `gate.py` remains the doctor-tool aggregation surface for the `ralph`
  context (where fail/infra DO stay blocking, `apply_context` :151-152) and for
  the validation/doc-fidelity callers that consume it today; it is not dead.
- No follow-up spec candidate is required. If a future need arises to run the
  DOCTOR tool-status aggregation specifically as a pre-claim signal (distinct
  from the enforcement-stack readiness the four modules already provide), that
  would be a new `readiness_signals()` contributor feeding `evaluate_launch` —
  NOT a re-context of `gate.py`.

## Cross-links

- `scripts/abcd/tools/gate.py` — `run_gate` (:168), `apply_context` (:149).
- `scripts/abcd/session/launcher.py` — `evaluate_launch` (:183),
  `ENFORCEMENT_GATES` (:60), `ReadinessReport.missing` (:75).
- `scripts/abcd/session/composer.py` — `compose_real_readiness` (:140),
  `RealReadinessProvider` (:68), `_OWNERSHIP` partition (:44).
- fn-71 (session default), fn-34.5 (the never-wired gate), fn-37.2–.5 (the four
  `readiness_signals()` providers).
