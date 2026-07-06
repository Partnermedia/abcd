---
id: itd-17
slug: model-effectiveness-tracking
spec_id: null
kind: standalone
suggested_kind: null
reclassification_history: []
created: 2026-05-03
updated: 2026-05-08
related_adrs: [adr-8]
---

# Pick the Right Oracle for the Job, Automatically

> **Reframe pending (committed 2026-05-08 per idea-4 jagged-frontier review).** This intent will be reframed from "model-effectiveness tracking" to **"frontier mapping"** — sharper framing per Dell'Acqua et al. 2023 ("Navigating the Jagged Technological Frontier"). Reframe scope:
>
> - **Headline** rewritten around frontier observation + capability-aware dispatch + frontier rendering (not generic effectiveness tracking).
> - **Persona quote** rewritten — Carol is product-lead per `personas.json`; pick a researcher / dispatch-operator role for the frontier framing.
> - **`What's In Scope`** rewritten around `{task_class, agent, backend, model_id, outcome, failure_mode_tag}` schema + closed-enum failure-mode tags (`hallucination` / `scope_drift` / `stale_context` / `under_specification_blindness` / `format_violation`) + `/abcd:frontier` command (bare = render; sub-verbs `flag` / `history` / `explain` for actions bare can't do; no `show` per the bare-command-as-render discipline).
> - **Storage migration** (named as breaking change to itd-17's existing scope): from `.abcd/oracle-effectiveness.json` + `.abcd/oracle-scorecard.jsonl` (current paths) to logbook events at `.abcd/logbook/frontier/<ts>/frontier-events.{json,md}` + agent-frontmatter declarations (`capability_scope` per itd-5; `known_failure_modes` per the new Frontier Awareness intent) + aggregate frontier-map rendered on demand from the two source-of-truth locations. No new top-level namespace (`.abcd/toolchain-state/` was killed at idea-4 review per the namespace-creep precedent).
> - **Composition with Frontier Awareness intent** (capture-stable after itd-17 ships; no ID reserved per the idea-3 release-don't-reserve precedent). Frontier Awareness owns the dynamic half (events + plan-time check + pre-cascade selector + `/abcd:frontier` command); itd-17 either folds entirely into Frontier Awareness or remains as the implementation substrate (logbook event schema + aggregation algorithm). Decision deferred to itd-17's reframed plan-review.
> - **Cascade contract preserved**: capability-aware routing is a *pre-cascade selector* (a layer above the cascade defined in itd-2 + itd-6), NOT a modification. § 7 "fixed cascade" language in `04-universal-patterns.md` stays intact.
>
> **Until the reframe lands** (in itd-17's plan-review), the body of this intent below describes the original "model-effectiveness tracking" framing. Treat it as the prior shape; the new shape is documented in `.work/idea-assessments/4-jagged-frontier.md` and the itd-5 extension at `disciplines/itd-5-prompt-quality-additions.md § Add 4`.

## Press Release

> **abcd learns which oracle backend performs best for which agent task.** Every oracle audit (lifeboat-oracle, press-release-composer's product audit, intent-fidelity-reviewer) records its outcome — was the verdict accurate? Did revisions follow? Over time, abcd builds a per-backend per-agent effectiveness profile and biases backend selection automatically. RP for content-fidelity audits, Codex for product critique, in-session for fast iteration — chosen based on data, not configuration guesswork.
>
> "I had no way of knowing which backend was actually catching the most real issues," said Eve, ML engineer. "abcd tracks the data and just picks the right one. My oracle audits got better without me changing any config."

## Why This Matters

abcd's oracle backend resolution is configuration-driven: `oracle.backend` chooses RP > Codex > in-session subagent in a fixed order. There's no feedback loop — abcd never learns which backend produces audits that actually correlate with real issues vs noise.

Lifted directly from `~/.abcd/`'s `model-effectiveness.json` and `model-scorecard.jsonl` patterns: track per-model statistics (ship_count, revise_count, false_revise_count, accuracy ratios) and per-judge scoring (depth_score, calibration_score). Apply to oracle backends. Bias selection by measured effectiveness.

## What's In Scope

- `.abcd/oracle-effectiveness.json` — per-backend per-agent rolling statistics
- `.abcd/oracle-scorecard.jsonl` — per-audit detail (run_id, backend, agent, verdict, follow-up actions)
- Backend selection logic: weighted by recent effectiveness for the requesting agent type
- `abcd oracle [agent]` CLI subcommand to inspect effectiveness profiles. Bare `abcd oracle` renders the full per-backend per-agent table; `abcd oracle <agent>` scopes to one agent. No `stats` sub-verb — collapses to bare per the bare-command-as-render discipline (see `02-constraints/04-naming.md`).
- Manual override remains (config setting, per-call flag)

## What's Out of Scope

- Cross-project effectiveness sharing (each project has its own profile)
- Auto-retraining or fine-tuning models (abcd doesn't own the models)
- Cost-aware selection (covered by separate cost-tracking work if/when needed)

## Acceptance Criteria

> _BDD format, per `itd-1-acceptance-gates`. These gates are checked by `intent-fidelity-reviewer` when this intent moves to `shipped/`._

- **Given** a project with N completed oracle audits across multiple backends, **when** the user runs `abcd oracle` (bare), **then** the output reports per-backend per-agent rolling statistics (ship_count, revise_count, false_revise_count, accuracy ratio) sourced from `.abcd/oracle-effectiveness.json`.
- **Given** an oracle audit completes, **when** the audit-fix loop or downstream review action confirms or refutes the verdict, **then** a row is appended to `.abcd/oracle-scorecard.jsonl` recording `run_id`, `backend`, `agent`, `verdict`, `confirmed/refuted`, and `follow_up_action`.
- **Given** sufficient effectiveness data exists for a given agent type (`>=` configured threshold of past audits), **when** abcd selects a backend for a new oracle call with `oracle.backend = "auto"`, **then** the selection is biased by recent effectiveness AND the rationale ("Codex chosen for press-release audit; +12% accuracy over RP on this agent in last 20 calls") is logged in the run report.
- **Given** insufficient effectiveness data (cold start), **when** abcd selects a backend, **then** the selection falls back to the fixed order (RP > Codex > in-session) AND the run log records "cold-start fallback".
- **Given** the user runs an oracle call with an explicit backend override (`--backend rp`), **when** the call executes, **then** the override is respected unconditionally AND the call's outcome is still recorded in `oracle-scorecard.jsonl` for future statistics.
- **Given** effectiveness data exists for a previously-used backend that has since been disabled (e.g., user removed RP), **when** `abcd oracle` (bare) runs, **then** the stale rows are surfaced under "historical (backend not currently available)" rather than influencing live selection.

## Open Questions

- How is "effectiveness" measured — verdict accuracy? Followed-up actions? User feedback signal?
- Cold-start: how does selection work before any data exists? (Falls back to fixed order.)
- Does the user see selection decisions ("using Codex for this audit because it's outperformed RP on press-release audits")?

## Field Evidence

> Empirical observations to feed the reframed plan-review (per the reframe blockquote at the top of this file). Not yet structured into the `{task_class, agent, backend, model_id, outcome, failure_mode_tag}` schema — recorded here as raw input the reframe should consume.

### 2026-05-16 — review-backend frontier, observed over the `fn-5` dual-backend plan-review loop (12 rounds, 24 reviews)

A new **`task_class: review_backend`** worth tracking once itd-17 reframes. The `fn-5` spec plan review ran RepoPrompt and Codex CLI in parallel for rounds 16–27; their behaviour was asymmetric and consistent:

| Backend | Strength (high accuracy) | Weakness (failure mode) |
|---|---|---|
| **Codex CLI** (`gpt-5.2`, high effort, repo read access) | Repo-wide reasoning — found 2 Criticals only a repo scan could find (a `pytest`-breaking stale-string match; a concurrency race). Cross-file / cross-doc drift. | **`verdict_unreliable`** — returned `SHIP` 8× while its body still listed Major/Critical findings. Also one `hallucination` (fabricated citation) and one cross-round flip-flop on an SDK detail. Slow (8–13 min). |
| **RepoPrompt** (scoped chat, sees only selected files) | Honest verdict tag (`SHIP` ⇔ clean body, every round). Tight consistency / regression review — caught contradictions the loop's own fixes introduced. Fast (1–3 min), stateful re-review. | **`selection_blind`** — sees only attached files; never raises a repo-wide issue. Cannot find anything outside the curated selection. |

**Candidate failure-mode tags this surfaces** (for the reframe's closed enum): `verdict_unreliable` (self-assessment contradicts findings), `selection_blind` (reviewer scope too narrow to see the issue). Both are distinct from the draft enum (`hallucination` / `scope_drift` / `stale_context` / `under_specification_blindness` / `format_violation`) and worth considering as additions.

**Implication for capability-aware dispatch:** for review `task_class`, the selection signal is not "which backend is better" but "which axis the review needs" — repo-breadth → Codex; scoped-consistency + a trustworthy gate verdict → RepoPrompt; high-stakes → both. The qualitative decision is recorded in [adr-8](../../decisions/adrs/0008-dual-backend-review-asymmetric-trust.md); itd-17's reframe should make it data-driven.

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
