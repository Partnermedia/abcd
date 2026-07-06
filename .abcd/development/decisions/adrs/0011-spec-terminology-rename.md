---
id: adr-11
slug: spec-terminology-rename
status: accepted
date: 2026-05-18
supersedes: null
superseded_by: null
related_intents: [itd-43]
related_rfcs: []
related_adrs: [adr-1, adr-2, adr-9, adr-10, adr-26]
---

# ADR-11: One canonical word for a specced block of work — spec

## Context

abcd organises engineering work in a layer it calls the **spec** — the
*how-to-build* artefact, tracing back to an intent or a brief plumbing-phase
([adr-1](0001-three-layer-mental-model.md)). abcd's terminology discipline
exists to kill exactly one failure — the same concept named two ways — and the
*how* layer is a load-bearing core noun: it appears in the terminology file, ADR
titles and bodies, the reviews subsystem's identifiers, the JSON schemas, the
reserved-vocabulary enum tokens in the brief, and prose across docs and READMEs.

Naming abcd's core artefact is a *decision*, not a content edit — ADRs are
decision records, and the canonical noun is woven through schemas, lint, and the
glossary. So the choice earns its own ADR.

## Decision

abcd adopts **spec** as its single canonical word for a specced block of work —
the *how-to-build* layer. It is abcd's native canonical noun; no other word
names this concept on any abcd-owned surface.

- The terminology file `terminology/core/spec.md` defines the concept and
  carries `forbidden_synonyms: [epic]`, so the glossary lint catches any
  regression to an older word.
- The `spec_id` field names the artefact in intent frontmatter and schemas.
- The reviews subsystem's identifiers and review-type tokens, the
  `issue.schema.json` / `grill-report.schema.json` keys and prose, the
  reserved-vocabulary enum tokens in `brief/02-constraints/04-naming.md` and
  their consumers, and prose across the brief, docs, command help, READMEs, and
  ADRs all say "spec".

The native spec layer ([adr-26](0026-native-spec-layer-ccpm-backend.md)) is
abcd's own store, so the canonical noun answers to abcd alone — there is no
external tool whose vocabulary abcd must track and no vendored surface carved out
of the rule.

This ADR **amends [adr-1](0001-three-layer-mental-model.md)**, whose *how* layer
is named "spec".

## Alternatives Considered

1. **A deprecated stub — a second term file pointing at `spec.md`.** Rejected: a
   stub adds a file the glossary must explain — a mild violation of the
   one-canonical-term rule this decision exists to hold. A single clean term
   file is the honest move.
2. **Two words for the layer — one schema field name, a different prose word.**
   Rejected: this is the exact bilingual state the terminology discipline exists
   to end — the schema saying one word while the glossary and docs say another.
   Half a naming is worse than none; it makes the inconsistency permanent and
   load-bearing.
3. **Leave the canonical noun implicit — no ADR, no forbidden-synonym guard.**
   Rejected: without a guarded invariant, a future contributor reintroduces a
   synonym and it drifts in silently. The `forbidden_synonyms` entry turns the
   choice into a lint-enforced invariant, not a convention.

## Consequences

**Gains:**
- abcd has one word for the *how* layer. The schema, the glossary, the reviews
  subsystem, the brief, and the docs all say "spec" — the
  same-concept-two-ways failure is closed for abcd's core noun.
- The `forbidden_synonyms: [epic]` entry in `spec.md` makes the choice a
  *guarded* invariant: the terminology lint flags any regression.

**Costs / obligations:**
- The forbidden-synonym guard must stay wired into the terminology lint, or the
  invariant decays to a convention.

**Downstream decisions enabled:**
- The terminology-lint guard means a future reintroduction of an older concept
  noun fails CI rather than drifting in silently.
</content>
</invoke>
