---
id: adr-1
slug: three-layer-mental-model
status: accepted
date: 2026-05-04
supersedes: null
superseded_by: null
related_intents: []
related_rfcs: []
related_adrs: [adr-2, adr-4, adr-5, adr-9, adr-11, adr-26]
---

# ADR-1: Three-layer mental model (brief / intent / spec)

> **Terminology note.** The *how* layer is named the **spec**. This ADR's title
> and body were updated by the spec-terminology-rename ADR
> ([adr-11](0011-spec-terminology-rename.md)), which records the rename
> decision and its rationale.

## Context

Early iterations of abcd's documentation conflated three distinct kinds of question into one corpus: *what is this project?* (canvas), *why does this user-facing change matter?* (forward-looking justification), and *how do we build this concrete thing?* (engineering plan). The original brief tried to hold all three at once, which made it large, hard to reorder, and impossible to ship in chunks. Scope creep entered through the engineering door because there was no surface that disciplined product clarity before engineering scope.

Plumbing work (adapters, agents, harness, hooks) had no natural home. Forcing press-release format on plumbing produced strained or mistargeted prose; treating plumbing as silent infrastructure left no acceptance gate for it.

## Decision

abcd organises development work in three layers, each tuned to the kind of question it answers:

- **Brief** — the *what*. Shared canvas; covers user-facing scope AND plumbing infrastructure. Read by anyone needing to understand the project as a whole.
- **Intent** — the *why* (user-facing). Press-release-shaped (Amazon working-backwards), with persona quote, scope, and Given-When-Then acceptance criteria. Lives as standalone documents so it can be reordered, bundled, deferred, or killed without disturbing the brief or each other.
- **Spec** — the *how*. A native minimal spec store — specs and tasks are directories whose location encodes status, over a dependency graph — with the companion harness `ccpm` as the primary deeper backend ([adr-26](0026-native-spec-layer-ccpm-backend.md)). Plan-reviewed before work; completion-reviewed after. Traces back to either an intent (user-facing) or a brief phase (plumbing) or a discipline (cross-cutting rule).

Acceptance discipline applies uniformly across both surfaces (intent and brief-phase), in Given-When-Then format. The `intent-fidelity-reviewer` agent compares delivered reality against intent acceptance; phase audit compares reality against brief-phase acceptance. The format is uniform; the *home* differs to match the nature of the work.

## Alternatives Considered

1. **Single unified "spec" surface** (one document type for all work). Rejected: forces press-release format on plumbing (mistargeted) and forces engineering-shape on user-facing capability (kills product clarity). The friction is load-bearing — different jobs need different shapes.
2. **Brief + intent only** (no separate spec layer). Rejected: specs are how-to-build artefacts; mixing them with why-it-matters intents collapses the press-release discipline. The native spec layer ([adr-26](0026-native-spec-layer-ccpm-backend.md)) provides a clean engineering surface; folding it into intents would duplicate.
3. **Intent + spec only** (no brief). Rejected: plumbing has no user moment to press-release. Without a brief, plumbing would either become silent (no acceptance gate) or be forced into intent shape (mistargeted prose). The brief carries the project-as-a-whole canvas that neither intent nor spec can hold.

## Consequences

**Gains:**
- Product-thinking is disciplined before engineering scope (intent's press-release format is the gate).
- Plumbing has a home with a uniform acceptance format (brief phases).
- Intents can be reordered/bundled/killed without rewriting the brief.
- The `intent-fidelity-reviewer` agent has a single audit shape to apply across two surfaces.

**Costs / obligations:**
- Three surfaces means three places to look — discoverability cost on first read. Mitigated by the reading guide in `brief/README.md`.
- Bidirectional links (intent ↔ spec, ADR ↔ intent) require lint enforcement (`internal/core/lint`).
- Cross-document fidelity becomes a real audit category (codes `XD001`–`XD007`); without it, drift between brief and intents goes undetected.

**Downstream decisions enabled:**
- ADR-2 (three intent kinds) — once the brief/intent/spec split exists, intents can have distinct shapes (standalone/bundle/discipline).
- ADR-4 (lifeboat as regenerable output) — disembark snapshots the brief; the layer model defines *what* gets snapshotted.
- ADR-5 (brief is current state) — once the brief is the canvas (not a versioned artefact), git covers history and intents cover forward-looking work.
