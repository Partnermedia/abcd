---
id: itd-2609020625400169
slug: an-intent-that-a-reading-occasioned-says-so-in-its-origin-wi
spec_id: spc-2609020626042168
kind: standalone
suggested_kind: null
reclassification_history: []
builds_on: [itd-178, itd-180, itd-185]
severity: minor
impact: fix
origin: researcher-authored
production_mode: dictated-and-formatted
---

# An intent that a reading occasioned says so in its origin, with the run and the item that occasioned it

Typed links: `builds_on` [itd-178](../shipped/itd-178-every-record-written-through-a-command-carries-its-origin-an.md) (the origin key, its parser and its lint), [itd-180](../shipped/itd-180-a-cold-reading-s-findings-land-as-reading-records-and-the-re.md) (the promote path for a dispositioned item), [itd-185](../shipped/itd-185-one-ingest-verb-validates-every-cold-reading-output-includin.md) (the ingest that mints items); `refines` [itd-178](../shipped/itd-178-every-record-written-through-a-command-carries-its-origin-an.md) (the third arrival path gets its writer).

## Press Release

> **A reading's contribution is traceable from both ends.** When an accepted reading item is promoted into an intent draft, the draft's `origin` reads `contributed-by-reading <rdg-N>/<rdi-N>`, naming the run and the item, and the item carries `promoted_to` pointing forward. The provenance lint already resolves that pair to the reading record; now something writes it. From a reading item, the record shows what it caused; from an intent, whether a reading occasioned it. Promotion of an issue keeps saying `extracted-from-record`, because an issue is something a person noticed and a reading item is something an instrument returned.

> "I need to be able to ask, for any intent, whether a cold reading put it on the table, and for any reading item, what became of it," said an AI/agent researcher who keeps the loop's genealogy. "Both directions, from the record, without reading the commit history."

## Why This Matters

[itd-178](../shipped/itd-178-every-record-written-through-a-command-carries-its-origin-an.md) names three values for `origin`: `researcher-authored`, `contributed-by-reading` carrying the run and item identifiers, and `extracted-from-record`. Its second acceptance criterion requires the pair to resolve to a reading record, and the shipped lint checks exactly that. Its fidelity verdict recorded the criterion as having no producer: the stamp primitive refuses the kind, and promoting a dispositioned reading item stamps `extracted-from-record` with `promoted_from`. The linkage the design wants is both directions: an accepted item stamps forward to whatever it produced, and the resulting intent carries the item identifier in `origin`, with the run identifier.

No reading has run, so no record is wrong today. The first accepted item promoted under the current path would be stamped as extracted from a record, which is the wrong claim about where it came from, and the join that the closing run's convergence and purpose-durability readings rest on would be lost at the first use.

## What's In Scope

- **The promote path for a reading item** stamps `origin: contributed-by-reading <rdg-N>/<rdi-N>` on the draft it mints. When it links an existing draft with `--intent`, it writes `promoted_from` and `promoted_to` and leaves the draft's `origin` untouched, because an origin is stamped at mint and never rewritten.
- **The stamp primitive** accepts the kind when, and only when, the caller supplies a well-formed run and item pair; resolution to the readings store is the promote path's, which reads the store before it mints, so no command can write the value without the join.
- **`promoted_from`** keeps naming the item, and `promoted_to` on the item keeps pointing forward, so the pair is redundant by design and the lint can check it both ways.
- **The promoted draft's seed carries no item identifier.** The press release seed is projected to the entailment reading, and a prior item's identifier in it would be revision history; the back-edge lives in `promoted_from`, which is not projected. The read-block eval plants a promoted seed to prove it.
- **The issue promote path is unchanged** and keeps `extracted-from-record`.
- The plugin surface pages for capture and intent say which path writes which value.

## What's Out of Scope

- Backfilling any record. Population is forward-only.
- A reading item promoted to anything other than an intent draft. Other landings (a discipline, an ADR, a brief passage) carry the join by their own means.

## Mechanism

We expect stamping the run and item at promotion to preserve the join because promotion is the only command that moves a reading item toward an intent, and a value written by the command that performs the act cannot drift from the act. It fails if a draft can reach the record with the value typed by hand, which the lint reports when the pair does not resolve and cannot report when it does; that residual is disclosed by itd-178 and stays.

## Scope Conditions

- The value carries exactly one run and one item. An intent occasioned by several items is promoted from one; linking a further item to a draft that already names another source writes that item's `promoted_to`, skips the back-edge and reports it. <!-- cond: cond-2609020727241828 -->
- Promoting a widening item requires its `accepted` disposition, which the admission intent's gate withholds until a comparative run names the item's run, so this path is transitively gated on the comparative channel for widening items. <!-- cond: cond-2609020626045842 -->
- The join resolves item to run directory in the readings store, as the shipped lint already does; no run record is required beyond the item's own directory. <!-- cond: cond-2609020626041091 -->
- **The impact is `fix`, and the reasoning is stated.** The path exists and writes a value the record itself calls the wrong claim; nothing usable changes for an issue promotion, and no reading item has yet been promoted, so there is no working invocation to break. <!-- cond: cond-2609020626044296 -->

## Acceptance Criteria

- **Given** an accepted reading item, **when** `capture promote <rdi-N>` mints a draft, **then** the draft's `origin` is `contributed-by-reading <rdg-N>/<rdi-N>` naming the item's run and id, and the provenance lint resolves it.
- **Given** an accepted reading item and an existing draft, **when** `capture promote <rdi-N> --intent <itd-N>` runs, **then** the draft's `promoted_from` names the item, the item's `promoted_to` names the draft, and the draft's `origin` is unchanged.
- **Given** an open issue, **when** `capture promote <iss-N> --grounds "pursued: <text>"` runs, **then** the draft's `origin` is `extracted-from-record`, unchanged.
- **Given** a call to the stamp primitive with the reading kind and no well-formed run and item pair, **when** it runs, **then** it refuses; resolution to the readings store is the promote path's, which reads the store before it mints.

## Prior Art

- [itd-178](../shipped/itd-178-every-record-written-through-a-command-carries-its-origin-an.md) and its spec; [itd-180](../shipped/itd-180-a-cold-reading-s-findings-land-as-reading-records-and-the-re.md) (the promote path for a dispositioned item).

## Open Questions

None.

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._

## Grounds

- pursued: we expect stamping the run and item at promotion to preserve the detection-to-intent join the closing run rests on, because promotion is the only command that moves a reading item toward an intent; a promoted draft whose origin cannot be resolved back to the item that occasioned it would show it wrong
