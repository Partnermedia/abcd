---
id: itd-64
slug: benchmark-driven-config-optimization
spec_id: null
kind: null
suggested_kind: standalone
reclassification_history: []
related_adrs: []
created: 2026-06-29
updated: 2026-06-29
prd_path: null
---

# abcd Learns From Its Own Runs Which Configurations Work Best, And Tunes Its Configuration Layer Accordingly

## Press Release

> **abcd gains an internal benchmarking system: it records configurations, evals, oracle/review outputs, and outcomes across runs, learns which configuration choices produce the best results, and proposes (or applies) optimisations to the abcd configuration layer.** abcd's whole identity is a configuration layer over robust tools — which oracle backend, which review bar, which scanner thresholds, which agent-instruction revisions, which loop settings. Today those choices are set by hand and judged by feel. A benchmarking system makes them EVIDENCE-DRIVEN: it captures what each configuration produced (review verdicts, fidelity outcomes, retries, time-to-SHIP, gate pass rates), benchmarks configurations against each other on real abcd work, and feeds back tuned defaults — so the configuration layer gets measurably better over time instead of staying static.

> "We keep tuning abcd's knobs by intuition — this backend, that review threshold, this agent-instruction revision," said a maintainer. "I want abcd to tell me, from its own run history, which settings actually produced better outcomes, and to tighten its own defaults based on that. The configuration layer should learn."

## Why This Matters

abcd is a configuration layer bringing together robust tools — its value is in the QUALITY of how it composes and tunes them, not in the tools themselves. But the configuration choices (oracle backend, review bar, severity thresholds, agent-instruction revisions, loop and retry settings, scanner selection) are currently set by hand and never systematically measured. abcd already PRODUCES rich signal on every run — review verdicts, fidelity outcomes ([[itd-60-doc-fidelity-anti-drift]]), grill reports, gate results, retries, receipts — but nothing harvests that signal to ask "which configuration produced the better result?". A benchmarking system closes that loop: it makes the configuration layer self-improving and evidence-grounded, which is exactly the leverage point for a framework whose entire job is good configuration. It also future-proofs the safety/review machinery (itd-62, the dual-backend bar) by letting abcd tune thresholds against measured outcomes rather than guesses.

## What's In Scope

- A benchmarking substrate that records, per run, the CONFIGURATION used (backend, review bar, thresholds, agent-instruction revisions, loop settings) alongside the OUTCOMES produced (review verdicts, fidelity verdicts, gate results, retries, time-to-outcome, eval scores).
- Comparison/learning over that record: which configuration choices correlate with better outcomes on comparable work, surfaced as ranked, evidence-backed recommendations.
- A feedback path that PROPOSES tuned configuration defaults (human-approved before applied — like the fidelity reviewer's never-auto-mutate posture), or applies within a configured envelope.
- Reuse of the signal abcd already emits (receipts, grill reports, review/fidelity verdicts) rather than new instrumentation where avoidable.
- Local-first, no Claude-Code dependency for the recording + analysis mechanics.

## What's Out of Scope

- Auto-applying configuration changes without human review (proposals first; auto-apply only within an explicit envelope).
- Touching or tuning the wrapped DEPENDENCIES themselves — this tunes abcd's CONFIGURATION of them, never forks/patches the tools (per the wrap-only rule).
- Building a general ML platform — it learns over abcd's own run signal to tune abcd's own config, scoped to that.
- Sending run data to any external service — local-first, the run history stays local.

## Acceptance Criteria

> _Given-When-Then per the itd-1 discipline._

- **Given** abcd runs work under some configuration, **when** the run completes, **then** the benchmarking substrate records the configuration used and the outcomes produced (verdicts, gate results, retries, time, eval scores) in a local store.
- **Given** an accumulated run record, **when** the analysis runs, **then** it surfaces which configuration choices correlate with better outcomes on comparable work, as ranked evidence-backed recommendations.
- **Given** a recommendation, **when** abcd would change a default, **then** the change is proposed for human approval (or applied only within an explicitly configured envelope), never silently auto-applied beyond it.
- **Given** the wrap-only rule, **when** optimisation runs, **then** it tunes abcd's CONFIGURATION of a tool, never the tool itself.
- **Given** the recording + analysis mechanics, **when** invoked outside Claude Code, **then** they run with no Claude-Code dependency, and run data stays local.

## Open Questions

- What is the unit of "comparable work" for fair benchmarking (same intent? same spec shape? same task size)? Without it, configuration comparisons are confounded.
- Which configuration dimensions are in scope first (backend + review bar + thresholds are highest-leverage; agent-instruction revisions and loop settings are richer but noisier)?
- What is the local store + schema for the config↔outcome record, and how does it reuse existing receipts/grill-reports vs add new fields?
- How is "better outcome" defined and weighted (fidelity MET rate? fewer retries? time-to-SHIP? a composite)? This is the load-bearing value judgement.
- Relationship to any evals work — does this subsume or feed an evals harness?

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._

## References

- Originating requests (2026-06-29): "optimise configurations based on internal benchmarks"
  + "sophisticated internal benchmarking system that learns from configurations, evals,
  outputs, etc. to optimise the abcd configuration layer" — merged here (optimisation is the
  PURPOSE of the benchmarking system).
- Consumes signal from: the fidelity reviewer, grill reports, review verdicts, gate results,
  receipts — the run artifacts abcd already produces.
- Governing rule: tunes abcd's CONFIGURATION of robust tools, never the tools themselves
  (wrap-only); local-first, run data stays local.
