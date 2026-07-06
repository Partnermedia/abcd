---
id: adr-2
slug: three-intent-kinds
status: accepted
date: 2026-05-07
supersedes: null
superseded_by: null
related_intents: [itd-34, itd-1, itd-5, itd-37]
related_rfcs: []
related_adrs: [adr-1, adr-3]
---

# ADR-2: Three intent kinds (standalone / bundle-member / discipline)

> **Terminology note.** The *how* layer is named the **spec**. This ADR's prose
> was updated by the spec-terminology-rename ADR
> ([adr-11](./adr-11-spec-terminology-rename.md)).

## Context

Per ADR-1, intents capture user-facing *why* in press-release shape, mapped 1:1 to specs that build them. The day-zero v5 audit (2026-05-07) surfaced a structural finding that wasn't drift-shaped: roughly 40% of the intent corpus had non-1:1 relationships with at least one other intent. Three patterns recurred:

1. **Coupled intents** — multiple intents that only made sense delivered together; splitting them across separate specs duplicated plumbing.
2. **Cross-cutting rules** — intents whose deliverable was "every other spec inherits this gate" (e.g., itd-1 acceptance-criteria gate, itd-5 prompt-quality additions). Forcing these into the standard `drafts/ → planned/ → shipped/` lifecycle produced specs whose only output was a rule applied elsewhere.
3. **Standalone intents** (~60%) fit the original 1:1 model cleanly.

The 1:1-only model calcified the wrong fits. Coupled work fragmented; cross-cutting rules became specs that didn't ship anything observable.

## Decision

Every intent declares a `kind` in frontmatter, set at `/abcd:intent plan` time. Three kinds exist:

- **`standalone`** — one press-release-shaped user moment, ships as one spec. The default; ~60% of the corpus. Lifecycle: `drafts/` → `planned/` → `shipped/`.
- **`bundle-member`** — multiple intents that share underlying delivery. Each member has its own press release (each captures a distinct user moment), but they ship together as one spec via multi-arg `/abcd:intent plan itd-A itd-B`. Members declare their bundle in frontmatter (`bundle: <bundle-id>`).
- **`discipline`** — a cross-cutting rule with no user moment of its own; applies to *every other spec* as an inherited acceptance gate. Lives in `disciplines/` (not `drafts/`); never gets its own spec; uses a `## Rule` template instead of a press release. Examples: itd-1, itd-5, itd-37.

The kind is **project-agnostic** — application projects produce their own disciplines (privacy-impact review, accessibility, code-style). The three kinds are a property of the intent framework, not abcd's particular subject matter.

Classification happens at plan time, with an advisory `suggested_kind` hint at capture time and continuous reclassification suggestions from the third role on `intent-fidelity-reviewer`.

Discipline subtypes are deferred — v1 introduces the kind itself; subtype taxonomy waits for empirical evidence (revisit triggers in `04-surfaces/05-intent.md § 1`).

## Alternatives Considered

1. **Keep 1:1-only; force coupled intents to share a spec name.** Rejected: the spec field is a single ID; representing "two intents, one spec" with naming convention only is ambiguous and breaks `intent_lint.py`'s reciprocal-link check.
2. **Two kinds: standalone + cross-cutting (no bundle-member).** Rejected: bundle-member captures a real pattern (coupled-but-each-has-its-own-press-release) that neither standalone nor discipline fits. Collapsing it into standalone forces an artificial split; collapsing into discipline strips the per-member press release.
3. **Free-text `kind` field (no closed enum).** Rejected: vocabulary drift compounds. Closed enum (`standalone | bundle-member | discipline`) lets `intent_lint.py` enforce kind-specific lifecycle (e.g., disciplines never go to `planned/`).
4. **Subtype taxonomy for disciplines now** (`methodology` / `documentation` / `audit`). Rejected: insufficient samples to cluster meaningfully. Free-text `kind_notes` field captures intent until the corpus warrants a closed enum.

## Consequences

**Gains:**
- Each intent's *delivery shape* matches its *forward-looking shape*.
- Coupled work ships as one spec without duplicating plumbing.
- Cross-cutting rules have a home (`disciplines/`) where they're not pretending to be features.
- The framework is project-agnostic; any abcd-managed project picks up the same three kinds.

**Costs / obligations:**
- `kind` enum becomes reserved vocabulary (registered in `02-constraints/04-naming.md`).
- `intent_lint.py` extends to verify `kind` + `bundle:` + directory-matching contracts.
- The `intent-fidelity-reviewer` agent grows a third role (continuous reclassification suggestions). Agent count stays at 15 — this is role extension, not agent splitting (per itd-31 precedent).
- Discipline-shaped intents need a `## Rule` template separate from the press-release template. Discipline-format guidance lives in `04-surfaces/05-intent.md`.
- Reclassification path (`/abcd:intent reclassify`) must move files between directories AND record `reclassification_history` in frontmatter — without it, the directory-as-truth principle (ADR-3) collapses.

**Downstream decisions enabled:**
- ADR-3 (directory-as-truth) — once kinds map to directories, the directory IS the lifecycle state for all kinds uniformly.
