# Ideate verdict — deterministic-delivery-pipeline

**Verdict: reframed.** Recorded on 2026-08-20 by abcd's idea-admission protocol —
primary-source research, a grill against the existing record, and an
independent adversarial review. This record exists so the idea is not
re-litigated: it stands whether the idea lived or died.

## The idea

Build a deterministic orchestration harness that runs abcd's own software-delivery pipeline end to end — research -> itd-84 decomposition -> draft -> an independent adversarial-review pair -> planning -> test-first implementation -> a second adversarial-review pair -> automated gates -> commit/PR -> close-out — with fixed stage order, mandatory review-pair gates before plan and before merge, and host hand-offs for the human-in-the-loop steps. Captured as iss-381 after the itd-130 delivery was hand-run this way.

## Leg 1 — Primary-source research

Every load-bearing claim checked against its primary source, never a
secondary citation.

| Claim | Primary source | Finding |
|---|---|---|
| LLMs cannot reliably self-correct their reasoning without an external signal, so self-review is a weak defect check. | https://arxiv.org/abs/2310.01798 | verified |
| Multi-agent debate between independent LLM instances improves reasoning and factuality over a single instance. | https://arxiv.org/abs/2305.14325 | verified |
| An adversarial LLM review-PAIR measurably improves software CODE-DEFECT detection specifically (as opposed to reasoning/factuality benchmarks). | https://arxiv.org/abs/2305.14325 | unverifiable |
| Workflow/deterministic orchestration (predefined code paths) is more reliable than model-driven orchestration for well-defined, decomposable tasks. | https://www.anthropic.com/engineering/building-effective-agents | verified |
| A deterministic orchestration harness makes the pipeline's OUTPUTS reproducible. | https://www.anthropic.com/engineering/building-effective-agents | falsified |
| Stripping authorship (blind review) reduces evaluator bias. | https://arxiv.org/abs/2101.02701 | verified |
| A pipeline cannot be more automated than its least-automated stage. | https://www.anthropic.com/engineering/building-effective-agents | falsified |
| Staged pipelines with mandatory quality gates (CI) reduce escaped defects. | https://arxiv.org/html/2309.10205 | verified |

## Leg 2 — Record grill

Does the brief, an intent, an ADR, or a principle already cover,
contradict, or supersede this idea? Every hit is cited by record id, and
every id resolved in this repository when the verdict was recorded.

| Record | Relation | Note |
|---|---|---|
| itd-107 | covered | The versioned routine template already carries fixed stage order, gate commands, worktree-isolated review pairs, hand-offs and close-out for the bug-hunt and plan-drain archetypes. The proposed pipeline is a NEW ARCHETYPE of this template (a config artefact), not a new engine. |
| adr-27 | contradicted | Accepted: the autonomous run is a pluggable seam with a thin native fallback, explicitly NOT a bespoke engine that hard-codes one loop shape. A standalone fixed-stage-order orchestration engine reverses this decision rather than extending the seam. |
| itd-104 | covered | Shipped (spc-18): /abcd:ideate already runs a deterministic fixed-order gated mini-pipeline (research -> record-grill -> adversarial review) with a fresh-context reviewer hand-off. It owns the front stages AND is the working precedent for the exact shape the idea claims as novel; its fidelity review also found that checking HOW an agent was run is theatre the binary cannot observe. |
| itd-84 | covered | Owns the decompose-before-filing stage as a discipline, staged on the promotion ladder (documented protocol now, agent later) per script-first-mvp — which contradicts building the whole harness up front. |
| itd-83 | covered | Owns the review-pair FIRING, but puts enforcement/hard-gating explicitly OUT of scope: enforcing an uncalibrated judge is the failure the discipline exists to prevent (hard-prereq: itd-81 calibration). The idea's mandatory gate is exactly the piece itd-83 refuses. |
| itd-82 | covered | Owns the close-out/triage stage: drain sorts the ledger, opens PRs, and promotes design work into intent drafts (seed-the-deferrals). It is itself the closest existing orchestrator. |
| itd-78 | covered | Owns 'what to build first' (severity + dependency edges -> derived build order) — the sequencing input the pipeline would consume, not re-derive. |
| itd-29 | covered | Owns the run/execution operator surface (start/status/pause/resume/ship/resolve, preflight budget, merge gates) that the pipeline's execution and commit/PR stages ride. |
| itd-98 | contradicted | 'Solo vs duo is measured, not debated.' The idea MANDATES a review-pair as a hard gate; the project's own stance is that pairing value is an empirical question to measure on matched tasks first, not a design claim to assert. |
| itd-50 | covered | Owns the close-out audit loop that drives a shipped intent to acceptance or UNACHIEVABLE/replan. |
| itd-28 | covered | Owns the SHA-pinned native review store (review.json, verdict enum) where a gate's review-pair receipts already land — the substrate, not a gap the idea fills. |

## Leg 3 — Adversarial review

Conducted fresh-context and off-policy by an evaluator that did not carry
out the research and received the idea as an artefact of unknown
authorship — the evaluator-outside-the-loop principle applied to ideas.

- **fatal** — Novelty: something here is genuinely new after the grill. Refuted — every stage maps to an existing owner (itd-104 shipped, itd-84 discipline, itd-83/82/78 drafts, itd-29 planned) and the composition itself is already the itd-107 template's shape; nothing is left unowned except an archetype config slot.
- **fatal** — Architecture: building it now respects the record. Refuted — a standalone deterministic orchestration engine is the thing adr-27 accepted AGAINST (pluggable seam, not a bespoke engine), and script-first-mvp says build the smallest documented protocol before the Go core absorbs the proven contract. Either sinks 'build the engine now'.
- **fatal** — The adversarial review-PAIR earns its place as a mandatory gate. Refuted three ways — no primary shows review-pairs improve code-defect detection; two same-model reviewers are a correlated-failure hazard the repo already refuses to assume (itd-98 measures, itd-83 requires itd-81 calibration); and adversarial-review-scales-with-blast-radius makes the armed gate a post-runs seed, not a now-build.
- **partial** — 'Deterministic scaffold' survives scrutiny. Partial — determinism buys reproducible control-flow/render (already delivered by itd-107's byte-identical render), not reproducible outputs; the stochastic LLM stages and itd-104's 'checking how an agent ran is theatre' finding refute the stronger reading, and the false 'least-automated stage caps everything' claim rides along here.
- **fatal** — Sequencing/ROI holds if built first. Refuted — most composed stages are still drafts/planned, so the harness would orchestrate hand-offs between manual steps: an automated outer loop around a manual body, negative ROI against finishing the stage verbs and composing later for near-free as an itd-107 archetype. The pitch concedes the automation cap, which is itself the refutation.

## Rejected alternatives

- **Build a standalone deterministic delivery-pipeline orchestration engine now, with the full fixed-order control flow as bespoke code.** — Reverses accepted adr-27 (the run is a pluggable seam, not a bespoke engine) and violates script-first-mvp; the adversarial review found multiple independent fatals with no single-point fix. The surviving scaffold is an itd-107 archetype (config), not an engine.
- **Ship the mandatory review-pair as an armed hard-gate that refuses the plan and merge stages without review receipts, immediately.** — adversarial-review-scales-with-blast-radius makes the armed detector a post-runs seed; itd-83 puts enforcement out of scope pending itd-81 judge calibration; and no code-defect evidence plus correlated-failure risk means itd-98's 'measure, don't mandate' governs. It ships as the documented protocol, with the armed detector deferred.
- **Claim the harness makes the pipeline reproducible (a 'deterministic pipeline').** — Determinism reaches only control-flow and prompt assembly (itd-107's byte-identical render); the substantive LLM stages stay stochastic, and itd-104's own fidelity review calls checking how an agent was run theatre. The honest claim is deterministic ASSEMBLY, already delivered.

## What follows

The idea as posed does not survive, but the reframing recorded above
does. Any graduation to a draft intent carries the reframing, not the
original wording.
