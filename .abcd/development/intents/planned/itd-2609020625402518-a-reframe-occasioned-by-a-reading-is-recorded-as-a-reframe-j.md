---
id: itd-2609020625402518
slug: a-reframe-occasioned-by-a-reading-is-recorded-as-a-reframe-j
spec_id: spc-2609020626048705
kind: standalone
suggested_kind: null
reclassification_history: []
builds_on: [itd-180, itd-189]
severity: minor
impact: additive
origin: researcher-authored
production_mode: dictated-and-formatted
---

# A reframe occasioned by a reading is recorded as a reframe, joined to what occasioned it, without carrying the construal it replaced

Typed links: `builds_on` [itd-180](../shipped/itd-180-a-cold-reading-s-findings-land-as-reading-records-and-the-re.md) (record families in the issue tier), [itd-189](../shipped/itd-189-what-the-widening-reading-proposes-is-admitted-or-declined-o.md) (the surprise entry as its own act); `refines` [adr-55](../../decisions/adrs/0055-the-construal-stands-in-the-record-its-history-does-not.md) (a reframe record beside the construal, flagged for the maintainer).

## Press Release

> **A change to the frame is a record, not a diff.** `abcd capture reframe --occasioned-by <rdi-N|dsp-N|srp-N> --grounds "<why>"` writes one frame-level revision record when the construal is rewritten because of a reading, naming what occasioned it, the hash of the construal before and after, and the grounds. The prior construal itself passes to ledger content on the local side, as adr-55 requires, so the committed record shows that a reframe happened, when, and why, without committing the framing it abandoned. `abcd <id>` on a reframe reports its occasion and both hashes.

> "When a detection sends me back to the frame rather than to the artefact, I need the record to say that is what happened," said an AI/agent researcher who keeps their own design record. "Otherwise a reframe looks like an edit to a paragraph, and the reading that caused it gets no credit and no blame."

## Why This Matters

The cold-reading design lists where an accepted detection may land: an intent, a discipline, an ADR, a brief passage, the construal section, and the frame. Every landing exists except the last, which the design schedules for Iteration 2 as a frame-level revision record, "without which a reframe occasioned by a reading cannot be recorded as a reframe".

[adr-55](../../decisions/adrs/0055-the-construal-stands-in-the-record-its-history-does-not.md) rules that the construal as it presently stands is committed record and that declined construals, superseded terms and the reasoning that settled a dispute stay on the local side and are read by nothing automated. The brief's framing chapter keeps no history in the section. A frame-level record therefore cannot carry the prior construal's text, and it need not: what is wanted is the join, from a detection to the reframe it occasioned and from a reframe to the reading that caused it, and the fact that the frame moved rather than the artefact.

Without it, Iteration 2's closing run cannot distinguish a tension that was answered by a reframe from one that was answered by a build, and the design's purpose-durability and convergence readings both depend on knowing which.

## Decisions flagged for the maintainer

- **The frame is the construal section, which is narrower than adr-55's construal.** adr-55 counts the framing section's statement, the glossary's committed terms and the committed scope as the construal as it presently stands; this record keys on the construal section alone, so a change to the glossary or to committed scope is not a reframe by its definition. That narrowing is flagged and carried on the same issue.
- **A reframe record beside the construal refines adr-55.** adr-55 is silent on recording that a rewrite happened; this record adds a committed pointer to an event whose content stays local. Whether every construal rewrite, reading-occasioned or not, must carry such a record is a standing rule over the brief and belongs in an ADR or a discipline, captured as an issue in the same change; this intent covers only reframes occasioned by a reading.

## What's In Scope

- **A reframe record family** in the working tier beside the other reading families, one record per reframe, minted on the timestamp seam, carrying: the occasion (a reading item, a disposition or a surprise), the construal's content hash before and after, and the grounds. It is warm and is excluded from every reading by positive inclusion, and the manifest asserts the exclusion.
- **`abcd capture reframe`** as the only writer, refusing an occasion that does not resolve, a ground below the substance floor, and a before-hash that does not match a committed state of the construal section. The verb reads the section's committed content itself; the operator supplies no hash. The join to the occasion is operator-asserted, with one check: the occasion predates the rewrite.
- **Dispatch** on the reframe id, reporting the occasion and the two hashes.
- The capture surface page documents the verb.

## What's Out of Scope

- The prior construal's text. It passes to the local side under adr-55 and this record does not carry it.
- Reframes not occasioned by a reading, and any lint over construal rewrites. Both wait on the flagged decision.
- The fourth audit verdict. A reframe changes what is built next; nothing here puts a delivered promise in question, which the design defers by decision.

## Mechanism

We expect a record keyed to the construal's committed hash to make reframes countable without committing framing traces because the hash identifies which construal a reading saw and which replaced it, and both are already committed content, while the reasoning between them is not. It fails if a rewrite and its record cannot be paired by hash, which happens when the section is rewritten twice between records; the verb refuses a before-hash it cannot find at HEAD, so the failure is loud.

## Scope Conditions

- The construal is one section of one brief chapter, and the record keys on that section's content. A repository with more than one construal surface is outside this scope. <!-- cond: cond-2609020626047674 -->
- The verb reads the section at HEAD. A reframe recorded before the rewrite is committed carries the pre-rewrite hash as its before-hash and is completed by a second write after the commit; the verb says which half it wrote. <!-- cond: cond-2609020626049770 -->

## Acceptance Criteria

- **Given** a reading item and a construal rewritten and committed, **when** `capture reframe` runs with the item as occasion and a ground, **then** one reframe record exists carrying the occasion, the before-hash matching the previously committed construal, the after-hash matching the current one, and the ground.
- **Given** a construal whose committed content matches no known prior state, **when** the verb runs, **then** it refuses and names the mismatch.
- **Given** an occasion that does not resolve to a reading item, a disposition or a surprise, **when** the verb runs, **then** it refuses.
- **Given** a repository holding reframe records, **when** any reading is assembled, **then** no reframe record reaches the bundle and the manifest asserts the family's exclusion.
- **Given** `abcd <reframe-id>`, **when** it runs, **then** it reports the occasion and both hashes.

## Prior Art

- [adr-55](../../decisions/adrs/0055-the-construal-stands-in-the-record-its-history-does-not.md); the brief's framing chapter; [itd-180](../shipped/itd-180-a-cold-reading-s-findings-land-as-reading-records-and-the-re.md) and [itd-189](../shipped/itd-189-what-the-widening-reading-proposes-is-admitted-or-declined-o.md).
- The cold-reading rulings of 2026-08-28 in the decision log.

## Open Questions

None beyond the flagged decision above. The family's identifier prefix is the spec's to fix.

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._

## Grounds

- pursued: we expect a reframe record keyed to the construal's committed hash to make reframes countable and joinable to the reading that occasioned them without committing framing traces; a reframe the record cannot pair with its before and after construal, or a prior construal's text reaching the committed record, would show it wrong
