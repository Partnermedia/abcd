---
id: itd-58
slug: session-reviewer-verdict-ingestion
spec_id: fn-69-session-reviewer-verdict-ingestion-path
kind: standalone
suggested_kind: standalone
reclassification_history: []
related_adrs: ["adr-8"]
created: 2026-06-22
updated: 2026-06-22
prd_path: null
---

# A Real Reviewer's SHIP Verdict Reaches The Session Gate Through A Channel The Worker Cannot Forge

> **DELIVERED by `fn-69-session-reviewer-verdict-ingestion-path` (2026-06-26).** The
> supervisor-owned ingestion (`ReviewerRunner` → capture → `parse_trusted_verdict` →
> `AdvanceGate.record_verdict`) ships in `scripts/abcd/session/verdict_gate.py` (task .1),
> and the A4-positive round-trip is proven end-to-end through the real composer +
> a launcher-produced spec in `tests/abcd/test_session_a4_positive_e2e.py` (task .2) —
> no direct `record_verdict()` shortcut, no real-repo mutation. `.work/code-review.md`
> §A4 is upgraded to "closed (session path)".

## Press Release

> **abcd's session enforcement learns to ingest a live reviewer's verdict through a trusted production path — so a genuine SHIP advances the work, while a worker's self-written SHIP still cannot.** Activating the session layer (fn-62) closes the A4 *exploit*: a worker that writes its own `"verdict":"SHIP"` receipt is denied at the advance gate, because no trusted verdict was ever recorded. But that is only the negative half. The positive half — a real reviewer (RP or Codex) returning SHIP, parsed into a `TrustedVerdict`, and recorded so the legitimate advance proceeds — has no production path today: `record_verdict()` is a supervisor-only direct method, and nothing in the live launch path parses reviewer output and calls it. This intent builds that ingestion path and proves it end-to-end, so the trusted-verdict channel is a full round-trip, not a one-sided denial.

## Why This Matters

The A4 finding in the security SSOT is "the worker self-writes its own SHIP review receipt." fn-62 closes the dangerous direction of that: under `/abcd:session`, a forged receipt no longer advances the task, because the advance gate (`advance_adapter` → `verdict_gate`) only honors a verdict that was *recorded* through `record_verdict()`, and a worker cannot reach that method. That is the security win, and it is real.

What fn-62 deliberately does **not** do is wire the *legitimate* side: when a real reviewer returns SHIP, some production component must (a) capture the reviewer's output from the supervised run, (b) parse it into a `TrustedVerdict` (the `parse_trusted_verdict` machinery exists for this), and (c) call `record_verdict()` so the genuine advance is permitted. Without it, the only way to make a trusted SHIP "advance" in a test is to call `record_verdict()` directly — which is a test-only shortcut, not proof that the production channel works. A plan-review (Codex, fn-62 round 3) caught exactly this: requiring the A4 *positive* proof inside fn-62 would force inventing this ingestion path under a spec scoped to "wiring only," so it was split out here.

Until this lands, the session layer enforces A4 as a **denial without a counterpart**: nothing forged gets through, but the framework has not demonstrated that a real reviewer SHIP flows through the same trusted channel rather than some side door. Closing the loop is what lets `.work/code-review.md` mark A4 fully closed for the session path, not just "self-attestation exploit closed."

## What's In Scope

- A production component in the supervised launch path that captures the live reviewer's output, parses it via `parse_trusted_verdict` into a `TrustedVerdict`, and records it through `record_verdict()` (gating reviewer gates; advisory mined for findings, per ADR-8).
- An end-to-end A4 *positive* proof entering through `launcher.launch`: a trusted reviewer SHIP injected through the production ingestion endpoint advances the task, and the test performs no direct flow-state writes and no direct `record_verdict()` call.
- Pairing with fn-62's A4 negative proof so the channel is shown as a full round-trip (forged → denied; trusted → advances).
- `.work/code-review.md` A4 annotation upgraded from "self-attestation exploit closed (session path)" to "closed (session path)" once the positive proof is green.

## What's Out of Scope

- fn-62's wiring (readiness composition, supervisor hook composition, fail-closed launch, A1/A2/A3/B2 conversions, A4 negative) — this intent depends on fn-62, it does not redo it.
- Bare-Ralph: like all session-path conversions, this is scoped to `/abcd:session`; bare `ralph.sh` keeps the fn-58 interim posture.
- Any change to how the advance gate DENIES (that is fn-62 + the existing gate); this intent only adds the trusted INGESTION half.

## Acceptance Criteria

> _Given-When-Then per the itd-1 discipline._

- **Given** a supervised `/abcd:session` run where a real reviewer returns SHIP, **when** the run reaches the advance point, **then** the reviewer output is parsed into a `TrustedVerdict` and recorded through the production path (no direct `record_verdict()` from outside the supervisor), and the legitimate advance proceeds.
- **Given** the same run but a worker-forged SHIP receipt with no trusted reviewer verdict, **when** the worker attempts the advance, **then** it is denied (fn-62's negative behavior is preserved, now as one half of a demonstrated round-trip).
- **Given** the end-to-end positive test, **when** it runs, **then** it enters via `launcher.launch`, injects through the production ingestion endpoint, and asserts the advance with zero direct flow-state writes by the test.

## References

- `.work/code-review.md` §A4 (the finding) + §N1
- `scripts/abcd/session/verdict_gate.py:107` (`record_verdict`, supervisor-only — the gap), `scripts/abcd/session/verdict.py:158` (`parse_trusted_verdict`), `scripts/abcd/session/advance_adapter.py` (the gate)
- fn-62 (dependency — wiring + A4 negative), adr-8 (trusted-verdict policy resolution)
- Surfaced by: Codex plan-review of fn-62, round 3 (2026-06-22)
