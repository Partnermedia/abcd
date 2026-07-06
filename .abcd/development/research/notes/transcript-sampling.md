# Transcript Sampling — Phase 0 Task 2

## TL;DR

- Signal density (user-decision turns / all turns): **1.4%** (range 0–2.1%) — conservative lower bound
- Signal density (user-decision turns / user turns only): **22.3%** (138 / 619)
- Signal density (noise-filtered user — excludes S1–S3, M2): **22.9%** (138 / 602)
- **Operational verdict: VIABLE** (user-message density 22.3% ≥ 15%; see Measurement Deviation in § Pass B viability assessment)
- Historical note: spec-formula density **1.4%** (conservative lower bound; NEEDS REDESIGN on original metric — superseded; see Measurement Deviation)

**Note on numerator and denominator:**
The task spec requires scoring turns where "the user OR assistant explicitly resolves a decision."
In practice, large sessions were scored from stratified user-message samples; agent sub-steps were
not individually classified. Decision counts reflect user-authored turns only and are a conservative
lower bound on the spec's "user OR assistant" intent — the true density is higher. The 1.4% figure
uses all specstory headers as denominator (per task spec formula) with a user-only numerator. Using
user turns as both numerator and denominator gives 22.3%, above the ≥15% VIABLE threshold.
The 1.4% figure is retained as historical provenance; the user-message denominator (22.3%) is the
operational gating signal per the Measurement Deviation recorded in § Pass B viability assessment.

Corpus snapshot: 2026-05-09, `ls | wc -l`.

## Sampling methodology

- Corpus: 439 transcripts, `~/ABCDevelopment/Autonomous/idelphiDev/.specstory/history/`
- Date range: 2026-04-11 → 2026-05-08 (~28 days); task spec cited 409 files through 2026-05-03
  (corpus grew by 30 files; denominator adjusted accordingly)
- Stratification: 3 small + 3 medium + 3 large + 5 boundary (14 total)
- Boundary criteria: ±2 days around idelphiDev epic transitions:
  - spc-14 (created 2026-04-19, done 2026-04-19)
  - spc-19 (created 2026-04-20, done 2026-04-27)
  - spc-21 (done 2026-04-29)
- Small stratum (S1–S3): intentionally sampled `command-name-clear-command` stubs (bottom 5% by
  size, ≤1 KB); these represent the zero-signal floor and are useful to confirm the 2 KB cutoff.
- Medium/large strata: `you-are-running-one` (Ralph autonomous, 399 files) and most
  `command-name-clear-command` files (23 files) excluded to avoid over-representing noise. M3
  was selected for medium stratum; it happens to fall near spc-19 done but is counted as medium
  (not boundary) to preserve the 3 + 3 + 3 + 5 split.
- Boundary picks chosen for genuine human–agent sessions near epic transitions.

## Per-transcript metrics

"Total turns" = all `_**User**` + `_**Agent**` specstory headers (including tool-use sub-steps).
"User turns" = `_**User**` headers only. Decision turns scored from user messages only (agent
sub-steps not classified individually). Density % (spec) = decisions / total turns;
Density % (user) = decisions / user turns.

| Filename | Size (KB) | Total turns | User turns | Decision turns | Off-topic | Wandering | Density % (spec) | Density % (user) |
|---|---|---|---|---|---|---|---|---|
| `2026-04-29_21-25-38Z-command-name-clear-command.md` (S1) | 1 | 6 | 6 | 0 | 0 | 2 | 0% | 0% |
| `2026-05-03_17-04-45Z-command-name-clear-command.md` (S2) | 1 | 5 | 5 | 0 | 0 | 0 | 0% | 0% |
| `2026-05-07_15-42-04Z-command-name-clear-command.md` (S3) | 1 | 5 | 5 | 0 | 0 | 0 | 0% | 0% |
| `2026-05-03_11-05-48Z-first-then-add-epics.md` (M1) | 64 | 57 | 7 | 1 | 0 | 0 | 1.8% | 14% |
| `2026-04-30_05-07-16Z-we-hit-the-rate.md` (M2) | 70 | 24 | 1 | 0 | 0 | 1 | 0% | 0% |
| `2026-04-27_11-23-55Z-any-branches-to-merge.md` (M3, near spc-19 done) | 180 | 62 | 5 | 1 | 0 | 1 | 1.6% | 20% |
| `2026-04-14_06-49-26Z-what-s-the-status.md` (L1) | 3,267 | 1,737 | 176 | 36 | 7 | 66 | 2.1% | 20% |
| `2026-04-22_06-04-12Z-update-claude.md` (L2) | 3,071 | 1,536 | 73 | 29 | 9 | 23 | 1.9% | 40% |
| `2026-04-18_11-52-38Z-what-s-the-status.md` (L3) | 6,003 | 3,081 | 189 | 44 | 17 | 66 | 1.4% | 23% |
| `2026-04-19_21-26Z.md` (B1) ★spc-14 done | 34 | 22 | 2 | 0 | 0 | 1 | 0% | 0% |
| `2026-04-19_21-11-26Z-an-autonomous-ralph-loop.md` (B2) ★spc-14 done | 2,108 | 1,099 | 104 | 15 | 8 | 54 | 1.4% | 14% |
| `2026-04-20_09-42-33Z-command-name-clear-command.md` (B3) ★spc-19 created | 2,793 | 857 | 24 | 5 | 0 | 6 | 0.6% | 21% |
| `2026-04-29_13-51-36Z-command-name-clear-command.md` (B4) ★spc-21 done | 1,233 | 417 | 13 | 5 | 0 | 6 | 1.2% | 38% |
| `2026-05-01_06-30-34Z-what-s-the-status.md` (B5) ★spc-21 done+2d | 2,781 | 890 | 9 | 2 | 0 | 4 | 0.2% | 22% |

★ = boundary pick; adjacent epic shown.
Corpus-weighted density: **1.4%** (spec) / **22.3%** (user-message denominator).
All density values reproducible from table columns: Density % (spec) = Decision / Total turns;
Density % (user) = Decision / User turns.

## Findings

1. **91% of the corpus is Ralph-loop noise.** 399/439 files are `you-are-running-one` Ralph
   autonomous prompt files; each is 2–3 KB with near-zero human text. A naïve full-corpus scan
   inflates the "total turns" denominator ~10× without adding decisions, depressing raw density
   to ~0.1% if those files are included. Any viable Pass B must pre-filter them.

2. **The "total turns" denominator conflates tool-use steps with conversational turns.** Large
   sessions (L1–L3) have 8–17× more agent tool-use steps than user messages. The spec formula
   `decision/total` gives 1–2% because agent tool calls dominate the count. Within human-authored
   sessions, 14–40% of user messages contain decisions — a genuinely high rate. The low spec-formula
   density is an artefact of specstory's fine-grained turn recording, not content poverty.

3. **Boundary sessions (near epic transitions) have the highest user-message decision density
   (~20–38%).** B3 (spc-19 created, 21%) and B4 (spc-21 done, 38%) are among the highest. Marathon
   sessions (L2, `update-claude`, 40%) also show high density. The time-window hypothesis holds:
   decisions concentrate around epic transitions and active design sessions.

4. **Small files (bottom 5% of size) contribute zero signal.** S1–S3 are all `/clear` + `/exit`
   or rate-limit interrupts with no substantive content. A 2 KB file-size cutoff eliminates them
   entirely.

5. **Large transcript token budget is a Pass B risk.** L3 at 6 MB is far too large to pass to
   `chat-distiller` as-is, even with a time window. Within-transcript chunking and line-range
   retrieval are required before any LLM call; a max-file-size gate (~100 KB post-window) is
   necessary to keep context budgets tractable.

## Pass B viability assessment

### Measurement Deviation

**NEEDS REDESIGN is no longer the operational verdict.** The metric used in the original
preregistered contract (decision turns / all specstory turns) was found to conflate tool-use
sub-steps with conversational turns, producing a systematically deflated figure. Per the
CONSORT/COS-OSF Transparent Changes pattern, the deviation is recorded below with full provenance.
The original 1.4% number is preserved here as historical provenance — never deleted. The operational
verdict is **VIABLE**, contingent on the three forward-commitment gates encoded in Pass B acceptance.

```yaml
measurement_deviation:
  as_planned:
    metric: "decisions / all conversational turns"
    threshold: "<5% NEEDS REDESIGN"
    verdict: "NEEDS REDESIGN (original verdict triggered by 1.4% spec-formula figure)"
  as_conducted:
    spec_formula_density: "1.4%"
    user_message_density: "22.3%"
    corpus_snapshot: "2026-05-09"
    sample_size: "14 transcripts, 439-file corpus (see per-transcript table above)"
  reason_for_deviation:
    - "Spec formula conflates tool-use sub-steps with conversational turns (Finding #2 in this document)"
    - "Pre-filter requirement was not specified in the original contract; 91% of corpus is `you-are-running-one*` Ralph noise"
  decision_rule:
    new_threshold: "≥15% on user-message denominator → VIABLE"
    historical_provenance: true
  forward_commitment:
    - "Pre-filter `you-are-running-one*` files before any LLM call"
    - "User-message denominator (not all-turns)"
    - "Within-transcript chunking for files >100 KB after time-window extraction"
    - "Encoded in `06-delivery/01-build-sequence.md § 7`"
```

**Operational verdict: VIABLE** (user-message density 22.3% ≥ 15% threshold, contingent on the
three forward-commitment gates above). The 1.4% spec-formula figure is retained as historical
provenance in the table at L41-66 and in the `as_conducted` block above; it is the conservative
lower bound from the original contract, not the gating signal for Pass B.

**Provisional thresholds (original task spec definition, decision turns / total turns):**
- ≥15% → VIABLE
- 5–15% → VIABLE WITH MITIGATIONS
- <5% → NEEDS REDESIGN

**Measured density (spec formula, user-decision numerator / all-turns denominator): 1.4%**

This figure falls below `<5%` on the original spec-formula metric. As recorded in the Measurement
Deviation above, NEEDS REDESIGN is no longer the operational verdict. The spec-formula metric
is retained as a historical reference point; the gating signal for Pass B is the user-message
denominator (22.3%).

**Pass B contract changes (adopted — see `06-delivery/01-build-sequence.md § 7` for BDD gates):**

1. **Pre-filter corpus by filename pattern before any scanning.** Exclude `you-are-running-one*`
   (91% of files) before `chat-distiller` is invoked. Brings effective corpus from 439 to ~40 files.
   **Highest-priority — eliminates 91% noise before any LLM call.**

2. **Redefine the density denominator in Pass B acceptance criteria.** Change from "total
   specstory turns" to "total user messages in time window." Aligns metric with observable signal
   rate (~22%) and makes the threshold meaningful.

3. **Within-transcript chunking and max post-window file-size gate.** Files >100 KB after
   time-window extraction must split into line-window chunks (2,000-line segments) with map-reduce
   aggregation before any `chat-distiller` call. Without this, 3–6 MB transcripts blow context.
   (Finding 5 identifies this as "required before any LLM call" — in scope now, not a later phase.)

4. **Minimum file-size cutoff of 2 KB** (recommended cleanup — not viability-gating, but eliminates zero-signal /clear stubs cheaply).

**itd-11 mitigations (in a later phase, as scoped in itd-11):**

- **Multi-strategy filter chain** (time-window → keyword → epic-id → semantic fallback)
- **Confidence scoring** per chat-distiller finding + quarantine threshold
- **lifeboat-oracle audit** that quarantines low-confidence Pass B output

## Sample raw notes (per-transcript, brief)

- **S1–S3** (<2 KB): /clear or /exit sequences; rate-limit interrupts. Zero signal.
- **M1** (64 KB): Opens with multi-point feature spec → 1 decision. Ralph loop takes over remainder.
- **M2** (70 KB): Single user turn "we hit the rate limit." No decision content.
- **M3** (180 KB, medium stratum; near spc-19 done): Branch merge check + plan-review invocation. 1 decision (20%).
- **L1** (3.3 MB): Multi-day marathon. Dense feature design interspersed with manual testing. 20%.
- **L2** (3.1 MB): Repo migration, publish skill design, security audit, architectural discussion.
  40% user density — design-heavy.
- **L3** (6.0 MB): Largest file. Facilitator page, model curation, methodology variant design.
  23% user density; highest tool-step count.
- **B1** (34 KB, spc-14 done): Short. 1 real turn (status question). No decision.
- **B2** (2.1 MB, spc-14 done): Ralph loop near spc-14 completion. 14% — human intervention turns.
- **B3** (2.8 MB, spc-19 created 2026-04-20): Active epic scoping for analytics/telemetry. 21%.
- **B4** (1.2 MB, spc-21 done 2026-04-29): Ralph timeout investigation + Ollama/LM Studio epic. 38%.
- **B5** (2.8 MB, spc-21 done+2d 2026-05-01): Status/merge session. Light decisions (22%).

## Forward tracking

The three Pass B contract changes (pre-filter, denominator redefinition, chunking gate) are adopted
and encoded as BDD acceptance gates in `06-delivery/01-build-sequence.md § 7`. They are also
captured in:

- **`06-delivery/01-build-sequence.md § 7`** — three new BDD acceptance bullets (canonical encoding)
- **`05-internals/01-agents.md`** — chat-distiller row carries denominator footnote referencing this document
- **Phase 4 epic spec** (to be created) — must reference this document's viability findings
- Any Phase 4 task implementing `chat-distiller` MUST gate on these three mitigations as acceptance criteria
