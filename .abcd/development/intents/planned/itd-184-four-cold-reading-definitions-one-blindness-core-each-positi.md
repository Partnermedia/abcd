---
id: itd-184
slug: four-cold-reading-definitions-one-blindness-core-each-positi
spec_id: spc-62
kind: bundle-member
suggested_kind: bundle-member
reclassification_history: []
builds_on: [itd-86]
severity: major
impact: additive
---

# Four cold-reading definitions, one blindness core — each position licenses a different output, and none may hold another's licence

Typed links: `refines` [itd-86](itd-86-cold-reading-surface.md) — the
single-document cold reading generalises to four positions, instances
within one detector context.

## Press Release

> **One definition with four objects cannot hold.** The prohibition against
> proposing is constitutive of the detection pass and would void the
> widening pass entirely — so there are four definitions, each holding its
> object, its question, and its regime value, with the blindness core
> byte-identical across all four: no project context · no ledger access ·
> no memory across runs · no ranking or prioritisation · no selection,
> explanation or commitment · named provenance on every item produced ·
> no passed input is authoritative.

| Definition | Regime value | Object |
| --- | --- | --- |
| Widening | `generative` | Brief current text incl. the construal statement; glossary; disciplines; specs; the shipped tree where one exists. Excludes `intents/drafts/` and `intents/planned/` |
| Entailment | `explicative` | The claim record — drafts and planned intents included — plus the constraint sources: disciplines, glossary, specs, brief current text |
| Comparative | `evaluative` | The candidate set (the widening reading's pre-admission output) against the declared selection criteria (the criteria discipline) |
| Detection | `registrative` | Shipped tree against the claim record |

## Ruled

- **Ruled (maintainer, 2026-08-28; decision log):** build four agent definitions, one per
  supply regime, as instances within one detector context — the count of
  definitions and the count of contexts are different countings and are
  compatible. This draft implements the ruling as stated.

## Per-instrument content (maintainer readings design, 2026-08-28)

Each definition holds five things and nothing else: its object, its
question, the blindness core verbatim, its regime value, and its item
shape.

Questions — widening: *given the situation as this design construes it,
what configurations does the construal admit that are not present in what
has been committed to?* Entailment: *what does this design commit to, by
being the kind of thing it is, that its articulation does not state?*
Comparative: *for each candidate and each declared criterion, how do
options of this shape ordinarily behave?* Detection: itd-86's question,
the shipped tree against the claim record.

Item shapes (validated per position by the output contract) — widening:
configuration · what admits it, and no third body field (no preference,
no comparison against what was built); entailment: claim surfaced · claim
type (criterion / causal / context) · what implies it — surfaced, never
dispositioned; comparative: one item per candidate-criterion pair —
candidate id · criterion · characterisation; detection: tension ·
constraint in play · why it is a tension. The pattern named is carried in
the record's envelope at every position, never in a body.

The object asymmetry over drafts is deliberate and is stated in the
assembler's include list: the widening reading must not see the candidate
set it is asked to widen; the entailment reading properly reads drafts,
since articulation precedes selection.

Widening's prohibitions are review-flags, not ingest refusals (the
generative licence is widest): a recommendation among configurations, or a
characterisation of one as better than another, flags for researcher
review — comparison belongs to the comparative position.

Adopted into the blindness core (2026-08-28; routed to the governing
document): **no passed input is authoritative** — no document passed to a reading
is designated the fixed side of any comparison; a discipline, a glossary
term or a declared criterion is as open to being named in an item as
anything else. The core carried verbatim in all four definitions includes
this condition, its seventh. Unlike the other six it is held by the
core's wording rather than by construction — the assembler cannot enforce
it — and it is disclosed as such.

## What's In Scope

- The four definition files under `agents/`, and the test holding the
  blindness core byte-identical across them.
- The regime value stated in each definition and not derivable from
  operator input.

## What's Out of Scope

- Enforcing the blindness — the assembler's job, checked by its evals.
- Validating what a reading produced against its regime — the output
  contract's supply-regime gate.

## Acceptance Criteria

- **Given** the four definitions, **when** they are diffed, **then** the
  blindness core is byte-identical across all four.
- **Given** any definition, **when** it is inspected, **then** its regime
  value is stated in the definition and not derivable from operator input.
