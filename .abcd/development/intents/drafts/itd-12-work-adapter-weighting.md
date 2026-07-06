---
id: itd-12
slug: work-adapter-weighting
spec_id: null
kind: standalone
suggested_kind: null
reclassification_history: []
---

# Working Notes Don't Outrank Settled Decisions

## Press Release

> **abcd weights `.abcd/development/activity/notes/` (curated working scratch) below ADRs, specs, and memory in the principle distiller.** Quick scribbles in `.work/notes/` no longer compete on equal footing with carefully-considered architecture decisions. The distiller's confidence scoring per principle source is exposed in the lifeboat brief so reviewers can see why a particular principle made the cut.
>
> "I had a half-formed thought in `.work/scratch.md` that ended up appearing as a top-ranked principle in my lifeboat next to an ADR I'd spent days on," said Carol, engineering manager. "abcd now ranks them by source authority. The ADR wins; my scratch note becomes a candidate principle to revisit."

## Why This Matters

abcd promotes content from `.work/notes/` into `.abcd/development/activity/notes/` via `dev-sync`, then feeds those into `principle-distiller` alongside ADRs, specs, memory, code-rescuer principles, and review findings. All sources are weighted equally by default. Working scratch is highly variable in quality — drafts, brainstorming, status snapshots — and shouldn't be treated as authoritative.

This intent introduces source-authority weighting and exposes the reasoning to reviewers.

## What's In Scope

- Per-source weight configuration in `.abcd/config.json` (defaults: ADRs > specs > memory > reviews > notes > code-extracted)
- Confidence scoring per emitted principle
- Source-attribution shown in `principles.md` (each principle cites its sources with weights)
- Override mechanism for users who want to flip weights for their context

## What's Out of Scope

- Auto-detecting whether a `.work/notes/` file is "polished" vs "scratch" (heuristic, fragile)
- Cross-source contradiction resolution (a separate problem)
- Removing `.work/notes/` from inputs entirely (we want low-weight inclusion, not exclusion)

## Acceptance Criteria

> _BDD format, per `itd-1-acceptance-gates`. These gates are checked by `intent-fidelity-reviewer` when this intent moves to `shipped/`._

- **Given** a project with both an ADR and a `.abcd/development/activity/notes/` entry covering the same architectural concern, **when** principle-distiller runs in Pass C, **then** the resulting principle in `principles.md` cites the ADR as primary source AND the notes entry as a candidate-to-revisit (not equally weighted).
- **Given** `.abcd/config.json` declares per-source weights overriding the defaults (e.g., `principle_distiller.weights.notes = 0.8`), **when** principle-distiller runs, **then** the configured weights replace the defaults AND the resulting `principles.json` records the active weight table for audit.
- **Given** every principle in the lifeboat brief, **when** a reviewer reads `principles.md`, **then** each principle is annotated with its source list (ADR / spec / memory / review / notes / code-extracted) and the per-source weight that contributed to its acceptance.
- **Given** a low-weight source (notes) and a high-weight source (ADR) emit contradictory principles, **when** the distiller resolves the conflict, **then** the high-weight source wins by default AND the suppressed low-weight principle appears under "candidate-principles-to-revisit" rather than being silently dropped.
- **Given** the default weight hierarchy (ADRs > specs > memory > reviews > notes > code-extracted), **when** any abcd documentation describes the distiller, **then** the hierarchy is documented exactly once (SSOT) at `05-internals/04-universal-patterns.md` (or wherever the distiller lives) and every other reference links back rather than re-stating.
- **Given** a user wants ADRs and specs treated as equally authoritative for a particular project, **when** they set `principle_distiller.weights.adr = 1.0` and `weights.spec = 1.0` in `.abcd/config.json`, **then** the override takes effect on the next disembark and is recorded in the principle-distiller report.

## Open Questions

- Are the default weights right, or do they need to be tuned per-project?
- Should weights be a flat hierarchy or a context-dependent ordering (e.g., ADRs trump specs ONLY for architecture decisions, not naming conventions)?
- How does `intent-fidelity-reviewer` interact with weighted principles — does it audit higher-weight principles more strictly?

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
