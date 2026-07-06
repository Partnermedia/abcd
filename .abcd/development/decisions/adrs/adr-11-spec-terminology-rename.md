---
id: adr-11
slug: spec-terminology-rename
status: accepted
date: 2026-05-18
supersedes: null
superseded_by: null
related_intents: [itd-43]
related_rfcs: []
related_adrs: [adr-1, adr-2, adr-6, adr-8, adr-9, adr-10]
---

# ADR-11: One canonical word for a specced block of work — the spec-terminology rename

## Context

abcd organises engineering work in a layer it has, until now, called the
**epic** — the *how-to-build* artefact, a flow-next spec that traces back to an
intent or a brief plumbing-phase (adr-1). The word "epic" entered abcd because
the underlying flow-next plugin used it: `.flow/epics/` held the spec files and
the data carried an `epic_id` field.

flow-next 1.1.1 renamed its own data: `.flow/epics/` became `.flow/specs/`,
and the plugin's canonical word for the artefact became **spec**. abcd's
terminology discipline exists to kill exactly one failure — the same concept
named two ways — and after the flow-next 1.1.1 migration abcd was committing
that failure about its own core noun. The terminology file
`terminology/core/epic.md` still defined the concept as "epic" while the data
underneath it, and the tool that produces it, said "spec".

The rename is non-trivial because the old word is woven through abcd-owned
surfaces: a terminology file, ADR-1's title and body, the reviews subsystem's
internal identifiers, two JSON schemas, the reserved-vocabulary enum tokens in
the brief, and prose across docs and READMEs. It is also a *decision*, not a
content edit — ADRs are dated decision records, and a silent find-replace of an
accepted ADR's title would erase the decision trail. So the rename earns its
own ADR.

## Decision

abcd adopts **spec** as its single canonical word for a specced block of work —
the *how-to-build* layer. The old word is retired from every abcd-owned live
surface.

The rename is delivered in two parts, deliberately split:

1. **The atomic first part — done.** The `epic_id` → `spec_id` intent
   frontmatter field rename shipped on `main` in commit `5b3fd24` (the
   flow-next 1.1.1 migration merge). Renaming a schema field is a single
   coherent change that must land all at once — schema, linter, fixtures, and
   every intent file together — so it could not be deferred or partialled.

2. **The terminology sweep — this ADR's subject.** Everything else: the
   terminology file (`terminology/core/epic.md` → `spec.md`, with the old word
   kept only as a `forbidden_synonyms` entry so the glossary lint catches
   future regressions), the reviews subsystem's abcd-internal identifiers and
   review-type tokens, the `issue.schema.json` / `grill-report.schema.json`
   keys and prose, the reserved-vocabulary enum tokens in
   `brief/02-constraints/04-naming.md` and their consumers, and a prose sweep
   across the brief, docs, command help, READMEs, and ADRs. This part is
   non-atomic — it is a collection of independent surface edits — and is
   tracked as the fn-7 spec.

**The vendored flow-next boundary is left alone.** Everything under
`scripts/ralph/` (`flowctl.py`, `ralph.sh`, the `prompt_*.md` files, hooks) is
the flow-next plugin copied in; a plugin update reverts any edit. Its
`epic`/`spec` aliases stay as they are and are excluded from the sweep.

**The reviews-subsystem rename is fully unblocked — there is no external
`epic_review` token to preserve.** itd-43's central open question was whether
flow-next emits a review-receipt `type` of `epic_review` that abcd must keep as
a boundary. It does not. The installed flowctl 1.1.1
(`scripts/ralph/flowctl.py`) emits review-receipt `type` values `impl_review`,
`plan_review`, and `completion_review` — never `epic_review`. abcd's own
`epic`/`epic_review` review-type triad is purely abcd-internal and dead
relative to current flowctl: nothing external produces it. So abcd's reviews
subsystem can be renamed in full with no boundary token to preserve. (The lone
`epic_review` string inside `flowctl.py` is an unrelated local variable holding
`default_review`, not a review-type token.)

This ADR **amends ADR-1**, whose title named the *how* layer "epic". ADR-1's
title is corrected here; its body prose is swept as part of the terminology
sweep.

## Alternatives Considered

1. **A deprecated stub — keep `epic.md` as a pointer to `spec.md`.** Rejected:
   a stub preserves inbound links but adds a file the glossary must explain — a
   mild violation of the one-canonical-term rule the rename exists to restore.
   Inbound links are few and are swept directly; a clean `git mv` is the
   honest move.
2. **A silent find-replace across all surfaces, no ADR.** Rejected: the rename
   touches ADR-1's title, and accepted ADRs are dated decision records. Editing
   one silently erases the decision trail. The rename is a decision and gets a
   decision record.
3. **Rename only the schema field, leave "epic" as the prose word.** Rejected:
   this is the exact bilingual state the terminology discipline exists to end —
   the schema would say `spec_id` while the glossary and docs said "epic". Half
   a rename is worse than none; it makes the inconsistency permanent and
   load-bearing.
4. **Boundary-only rename of the reviews subsystem — keep `epic_review` as an
   assumed external token.** Rejected on evidence: flowctl 1.1.1 emits no
   `epic_review` type, so there is no boundary to honour. A boundary-only
   rename would leave dead `epic` identifiers in place for a token nothing
   produces.

## Consequences

**Gains:**
- abcd has one word for the *how* layer. The schema, the glossary, the reviews
  subsystem, the brief, and the docs all say "spec" — the same-concept-two-ways
  failure is closed for abcd's core noun.
- The `forbidden_synonyms: [epic]` entry in `spec.md` turns the rename into a
  *guarded* invariant: `lint_terminology.py` flags any future regression to the
  old word.
- The reviews subsystem sheds a dead internal token (`epic_review`) it never
  needed.

**Costs / obligations:**
- The sweep is non-atomic — a collection of surface edits across many files —
  so it ships as a multi-task spec (fn-7) rather than one commit. Until the
  whole spec lands, abcd is briefly mid-rename.
- Inbound links to `terminology/core/epic.md` must be repointed in the same
  spec, or they dangle.
- ADR-1's body prose still reads "epic" until the prose-sweep task runs; this
  ADR corrects only ADR-1's title up front.
- The vendored `scripts/ralph/` surface keeps the old word. That is an accepted
  permanent asymmetry: abcd does not own that code, and the grep-clean
  acceptance check excludes it.

**Downstream decisions enabled:**
- The terminology-lint guard (`forbidden_synonyms`) means a future
  reintroduction of "epic" as the concept noun fails CI rather than drifting in
  silently.
