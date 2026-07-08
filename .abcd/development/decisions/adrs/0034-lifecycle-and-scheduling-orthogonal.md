---
id: adr-34
slug: lifecycle-and-scheduling-orthogonal
status: accepted
date: 2026-07-08
supersedes: null
superseded_by: null
related_intents: [itd-7, itd-43]
related_rfcs: []
related_adrs: [adr-3, adr-9, adr-33]
---

# ADR-34: Intent lifecycle and phase scheduling are orthogonal axes

## Context

The record carried two contradicting definitions of what `intents/planned/`
means. `intents/README.md` defined it as "scoped into a roadmap phase", but the
corpus disagreed in both directions: most planned intents (itd-20, itd-24,
itd-29, itd-46, itd-48, itd-49, itd-50, itd-53, itd-58, itd-63, itd-65, itd-66,
itd-67, itd-69, itd-72) appear in no phase doc's `## Scope`, while two intents
named in a phase `## Scope` (itd-43 in Phase 0, itd-7 in Phase 6) still lived in
`drafts/`. With "planned", "phased", and "draft" disagreeing, the roadmap cannot
be used as a scheduler. (Captured as `iss-3`.)

The same README's own sequencing section already states the opposing model —
"an intent listed in no phase doc's `## Scope` is implicitly unscheduled" is a
valid state (per adr-9) — so the record was contradicting itself, not merely
drifting from the corpus.

## Decision

1. **Two orthogonal axes.** An intent's **lifecycle** is its commitment state,
   carried by directory location (adr-3): `drafts/` = captured, uncommitted;
   `planned/` = committed to build; `shipped/` = built. An intent's
   **scheduling** is phase membership, carried editorially by a phase doc's
   `## Scope` (adr-9). Neither axis is derived from the other.

2. **One directional invariant: scheduled ⇒ committed.** Any intent named in a
   phase doc's `## Scope` is committed by definition, so it lives in `planned/`
   (or `disciplines/` for discipline-kind intents). The converse does not hold:
   a planned intent with no phase is **committed but unscheduled** — a valid
   state, the committed bench awaiting sequencing (adr-9's unscheduled-intent
   rule, now explicitly extended beyond `drafts/`).

3. **Applied to the corpus.** itd-43 (Phase 0 `## Scope`) and itd-7 (Phase 6
   `## Scope`) move `drafts/` → `planned/`; the fifteen planned-but-unscheduled
   intents stay where they are, now legally. `intents/README.md`'s `planned/`
   definition drops the "scoped into a roadmap phase" clause.

4. **Enumerations derive from the axes, not from each other.** Any "later
   phase" or "out of scope" list is derived from `## Scope` membership (the
   scheduling axis), never from the `drafts/` directory (the lifecycle axis) —
   the two sets are different by design.

## Alternatives Considered

- **`planned/` = phased (enforce the README's old definition).** Rejected: it
  forces every committed-but-unsequenced intent back to `drafts/`, erasing the
  commitment signal, or forces premature phase authoring just to keep files in
  place — exactly the stale forecasting the phases README refuses ("naming
  which intents land in which future phase before the phase is planned would
  only go stale").
- **A `scheduled:` frontmatter field on intents.** Rejected: duplicates the
  phase docs' `## Scope` (adr-9 made that mapping editorial and single-sourced)
  and reintroduces the stored-status drift adr-3/adr-5 exist to prevent.
- **Full orthogonality with no invariant (a phase may bundle a draft).**
  Rejected: a phase `## Scope` is a commitment to build; letting it name an
  uncommitted draft makes "committed" unreadable from either axis.

## Consequences

- itd-43 and itd-7 relocate to `planned/`; `intent_lint`-style checks can later
  enforce the scheduled ⇒ committed invariant mechanically.
- `intents/README.md` describes `planned/` as committed (scheduled or awaiting
  sequencing) and its directory listings match disk.
- The out-of-scope/later-phase enumeration (`brief/06-delivery/03-out-of-scope.md`)
  is regenerated from `## Scope` membership rather than the drafts directory —
  tracked as `iss-8`.
- Hand-maintained intent counts keyed to "phased intents"
  (`brief/01-product/04-scope.md`) are re-derived on the scheduling axis —
  tracked as `iss-7`.
