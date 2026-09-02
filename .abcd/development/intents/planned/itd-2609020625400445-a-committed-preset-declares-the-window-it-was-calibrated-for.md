---
id: itd-2609020625400445
slug: a-committed-preset-declares-the-window-it-was-calibrated-for
spec_id: spc-2609020626048722
kind: standalone
suggested_kind: null
reclassification_history: []
builds_on: [itd-199, itd-198, itd-196]
severity: major
impact: fix
origin: researcher-authored
production_mode: dictated-and-formatted
---

# A committed preset declares the window it was calibrated for, and the record holds every preset to its declaration

Typed links: `builds_on` [itd-199](../shipped/itd-199-a-reading-is-about-something-narrower-than-everything-its.md) (the committed presets), [itd-198](../shipped/itd-198-an-assembly-reports-what-it-would-cost-before-a-reading-is.md) (the size report), [itd-196](../disciplines/itd-196-a-capability-is-rehearsed-end-to-end-before-it-ships-over-a.md) (rehearsal before shipping); `refines` [itd-199](../shipped/itd-199-a-reading-is-about-something-narrower-than-everything-its.md) (one entry per position in three parts, a declared window per entry).

## Press Release

> **A reading can be handed to a reader.** The committed preset file holds one entry per position, and each entry states three things: the object set the run is about (which records and which delivered paths), the kinds admitted within it (every kind the position's definition reads, while the object set fits the reader's window; a kind that proves useless leaves by a commit), and the estimated-token window it was calibrated for. The assembler applies the entry for the position it was invoked at, with no operand, and its own eval assembles the entry for every assembling position by dry run and fails when a result exceeds its declaration, so drift past the window is caught by the cold-reading eval lane, which is not yet a required check, rather than by the reader. An entry whose object set names one record is proven to carry that record's material and nothing else. The assembler still enforces no budget at invocation; the entry states the bound, the eval checks it, and the size report goes on saying what a run would cost, stating the figure whenever it is over the two-hundred-thousand-token target. Changing any part of an entry is a commit, which is how the object set grows and how a kind leaves the default over time.

> "I want to point the entailment reading at the claim record and know before I spend a run that the bundle fits the reader," said an AI/agent researcher who commissions cold readings against their own repository. "And when the repository grows past what the entry was measured at, I want the build to tell me, not the reader."

## Why This Matters

The instrument was found correct and undeliverable one day before release: about 9.8 MB per position, roughly 2.45 million estimated tokens ([iss-2608311501186646](../../../work/issues/open/iss-2608311501186646-the-assembled-input-for-a-real-reading-of-this-repository-is.md)). [itd-198](../shipped/itd-198-an-assembly-reports-what-it-would-cost-before-a-reading-is.md) added the size report and [itd-199](../shipped/itd-199-a-reading-is-about-something-narrower-than-everything-its.md) the committed presets. Measured on the v0.7.0 tree (commit 8f68ffb3) by dry-run assembly under the narrower of the two named presets that file then carried, the widening position assembles 257,835 bytes (about 67 thousand estimated tokens), the entailment position 942,074 bytes (about 245 thousand), and the detection position 751,612 bytes (about 195 thousand). Under the wider preset the detection position is 2,528,029 bytes (about 657 thousand). The entailment and detection assemblies exceed or sit at the edge of a 200-thousand-token reader, and nothing in the record says what window any preset was calibrated for.

Two gaps in itd-199's delivery compound this. Every committed preset carries an empty record list, and the only assembly under a record-id selector in the test corpus selects nothing, so the positive half of the selector grammar has never been shown to carry a record's material. And the widening reading's bound, stated in advance by the design, is a construal of one or two sentences and a glossary of three to six terms, which is a statement about what is passed, not about what the tree weighs.

The design fixes the invocation at a position and a target state and places what a reading is handed in the record; adr-2609021016286571 restores that, and the preset file is the record it points at. One entry per position, in the three parts the framework names, is the shape that lets a reviewer see in one place what a position reads, what it weighs and where it came from.

This intent lands in Phase A, before any reading runs, after the two-operand invocation and itd-194 and before the comparative channel and the reading-occasioned origin; the origin lands in Phase A too, before the first promotion of an accepted reading item, because the origin key is written only by commands.

This intent closes the critical issue above, resolves [iss-2609012259585189](../../../work/issues/open/iss-2609012259585189-nothing-reports-the-proportion-of-intents-that-carry-a-mecha.md), which asks that the entailment reading's yield bound be stated beside its findings, and resolves [iss-2608311501240566](../../../work/issues/open/iss-2608311501240566-three-of-the-four-reading-positions-receive-a-byte-identical.md), whose item-set distinctness itd-199 addressed by a selector-set proxy: the maintainer ruled on 2026-09-02 that the preset entry is the one configuration surface for a position's object set, kinds and window, so the entry each position is handed is what makes its item set its own, and the spec pins each default item set by digest.

## What's In Scope

- **One entry per position, in three committed parts.** The preset file moves to schema version 2, where each position's entry states its object set (records by id and delivered paths), the kinds admitted within it, and the window it was calibrated for. The named presets and their inheritance go with the operand that chose between them. At version 2 a position without a declaration is refused at load, and a version 1 file holding one preset goes on loading, so an adopter's committed file keeps working until they move it.
- **A declared window per entry** as an estimated-token figure the entry was calibrated to, using the size report's own byte-derived basis, together with the figure measured and the commit measured on.
- **An eval that assembles the entry for every assembling position by dry run** and fails when a result exceeds its declared window, naming the position, the measured figure and the declaration. It runs in the cold-reading eval lane; that lane becoming a gate that blocks a merge is the two recorded issues on it, which this intent names and does not close ([iss-2608311632382737](../../../work/issues/resolved/iss-2608311632382737-the-pre-push-gate-is-blind-to-both-eval-lanes-so-the-read-bl.md), [iss-2608311051046981](../../../work/issues/resolved/iss-2608311051046981-the-new-cold-reading-evals-ci-job-is-not-a-required-status-c.md)).
- **The default object set for Iteration 2 is Iteration 1's shipped state**: the fifteen workstream intents itd-177 to itd-189, itd-198 and itd-199, their specs spc-55 to spc-69, and the packages and pages they delivered that the deny list does not exclude. A test proves that an entry whose object set names one record carries that record's projected material and no other record's.
- **The default kinds are every kind the position's definition reads.** A reading of the object set includes all kinds when the object set fits the reader's window, on the ruling of 2026-09-02. The detection entry names `brief-section`, `glossary-term`, `discipline`, `spec`, `intent-projection`, `doc`, `config`, `source` and `test`, the glossary among them because the design framework's section 9 lists it among what the assembler reads for the position; measured on 2026-09-02, it comes to about 624,000 estimated tokens. The widening entry names `brief-section`, `glossary-term`, `discipline`, `spec`, `doc`, `config`, `source` and `test`, every kind the readings companion's section 5.2 lists for the position, "Brief current text including the construal section; `brief/glossary/`; `intents/disciplines/` including the selection-criteria discipline; `specs/`; the shipped tree where one exists, scoped per 4.5", with the object set narrowing the specs to spc-55 to spc-69 and the tree to the delivered paths; it comes to about 615,000. Both are over the target, stated by the size report on every assembly, and inside a million-token reader, and every source and test item is marked `unscanned` in the manifest. The entailment entry names the claim record and the constraint sources and no tree kind, because its definition's object holds no tree. A kind that proves useless (tests, say) leaves the default entry by a commit, which is recorded; two leaner entries are one such commit away at either tree position. The choice is recorded in each entry's own comment field. The `brief-section` kind reaches every entry that names it as the six chapters itd-194 admits, meta, product, constraints, surfaces, internals and delivery, so the figures stated here were measured over two chapters and are re-measured at landing by the window eval.
- **An assembly over a 200 thousand token target says so.** Two hundred thousand estimated tokens is the target an entry aims at, and any figure inside a reader's window is acceptable; a dry run or an assembly whose estimate exceeds the target prints one line naming the figure and the target, as the maintainer ruled on 2026-09-02, so the operator learns it from the tool rather than from the reader. The committed detection and widening entries are over the target and their declarations say so; the kinds are the lever that keeps the object set while cutting size.
- **A change to any part of an entry is a commit**, reviewed and inside the dirty gate, and the manifest names the entry applied and its hash. That is the mechanism by which the object set grows, a kind such as `test` enters or leaves the default for a position, and a window moves: each is recorded, and none is a flag.
- **The per-position entry resolves the byte-identical item sets.** The maintainer ruled on 2026-09-02 that the preset entry is the one configuration surface for a position's object set, its kinds and its window, so the change that delivers one entry per position resolves [iss-2608311501240566](../../../work/issues/open/iss-2608311501240566-three-of-the-four-reading-positions-receive-a-byte-identical.md): each position is handed its own entry, the default item set at each position is pinned by a digest recorded before the change, and two assemblies of one entry produce identical bundles.
- **The entailment size report states the mechanism proportion.** The readings companion bounds the entailment reading by how many intents carry a mechanism claim and asks that the proportion be reported beside the findings. At the entailment position the size report states how many projected intents carried a `## Mechanism` section, how many carried the `None stated.` nullity, and how many carried neither, as the maintainer ruled on 2026-09-02; the workstream's own fifteen shipped intents keep their absent claims as the Iteration 1 baseline and are reported, never backfilled.

## What's Out of Scope

- A budget enforced at invocation. The assembler reports; the entry declares; the eval checks. Nothing is refused for size, and there is nothing at the invocation to widen an entry with.
- A tokenizer. The basis stays byte-derived and labelled as such.
- The comparative position, whose object arrives by its own channel and is bounded by the widening run rather than by the tree; its entry is that channel's, in the shape fixed here.

## Mechanism

We expect a declared window checked by a dry-run eval to keep the instrument deliverable because the failure that reached release was one nobody measured, and a measurement that runs on every committed change in the eval lane is not one a builder forgets. It fails if the byte-derived estimate diverges from a reader's real count by more than the margin the declaration leaves, which the calibration sample recorded in spc-68 bounds.

## Scope Conditions

- The declared window is a property of the entry, not of any reader. Which reader a run is handed to is the operator's choice and is recorded on the run. <!-- cond: cond-2609020626042294 -->
- The eval measures the repository it runs in. A repository that adopts the presets measures its own tree; the figures recorded here are this repository's at the commit named. <!-- cond: cond-2609020626042487 -->
- Recalibration narrows what a position is handed, kind by kind, and is recorded as such in the preset file: A kind that proves useless leaves the default entry by a commit. A narrowing that changes what the position's definition names as its object is a ruling and is not made here. <!-- cond: cond-2609020626045182 -->
- **The impact is `fix`.** The change closes a critical defect, and the schema move is opt-in by version, so no adopter's committed file stops loading. <!-- cond: cond-2609020626048715 -->
- The widening reading receives the whole committed glossary, twenty-four term files across four contexts (ten core, three distribution, two interview and nine ledger), because the readings companion's section 5.2 names `brief/glossary/` whole; the companion's section 5.6 states the bound in advance as "the glossary three to six terms", so the run record's `bounds` list, which `reading ingest` populates from the manifest, states that bound as exceeded, on ruling M5 that a departure is "stated as a bound rather than passed off". The glossary figure the spec records was measured before the ledger context existed and is re-measured at landing. <!-- cond: cond-2609021140329660 -->
- The closing run uses the preset entry at the same hash the opening run's manifest records, because the design framework's section 13 requires the closing run to be over "the same object set"; where a prior run at the same position recorded a different preset hash, the closing run record's `bounds` list, which `reading ingest` populates from the manifest, states the mismatch as a bound on the comparison rather than repairing it. <!-- cond: cond-2609021140328523 -->

## Acceptance Criteria

- **Given** a preset file at schema version 2 in which one position carries no declared window, **when** it is loaded, **then** the assembler refuses and names the position.
- **Given** a preset file at schema version 1 holding one preset, **when** it is loaded, **then** its entries apply per position as they do today and the size report says no window is declared.
- **Given** the committed entries on the current tree, **when** the eval runs, **then** each of the three assembling positions' entries (widening, entailment and detection) measures at or below its declaration, and the eval passes.
- **Given** an entry whose declaration is set below its measured figure, **when** the eval runs, **then** it fails and names the position, the measured figure and the declaration.
- **Given** an entry whose object set names one shipped intent, **when** the entailment assembly runs, **then** the bundle carries that intent's projected fields and no other intent's, and the manifest lists exactly those items.
- **Given** a committed entry with a populated object set, **when** the assembly runs, **then** the bundle carries the listed records' material as the position admits it, the admitted kinds under the listed paths, and no unlisted record's.
- **Given** the preset file, **when** it is read, **then** each declaration carries the figure it was measured at and the commit it was measured on.
- **Given** an assembly or dry run whose estimate exceeds two hundred thousand tokens, **when** the size report renders, **then** it carries one line naming the figure and the target, and an assembly under the target carries no such line.
- **Given** an assembly or dry run at the entailment position, **when** the size report renders, **then** it states how many projected intents carried a `## Mechanism` section, how many carried the `None stated.` nullity, and how many carried neither, and the report at any other position carries no such statement.

## Prior Art

- [itd-198](../shipped/itd-198-an-assembly-reports-what-it-would-cost-before-a-reading-is.md) (the size report), [itd-199](../shipped/itd-199-a-reading-is-about-something-narrower-than-everything-its.md) (the committed presets), [itd-196](../disciplines/itd-196-a-capability-is-rehearsed-end-to-end-before-it-ships-over-a.md) (rehearsal before shipping).
- The cold-reading rulings of 2026-08-28 in the decision log; adr-2609021016286571 (the invocation is a position and a target state, and the committed preset for the position supplies the rest).

## Open Questions

None. Whether an object set should be expressible as a set of record families was narrowed at itd-199's verdict and is not reopened here.

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._

## Grounds

- pursued: we expect a declared window per preset checked by a dry-run eval on every change to keep the instrument deliverable, because the size defect reached release only because nobody measured; a preset that passes its eval and still overflows the reader it was calibrated for by more than the recorded margin would show it wrong
