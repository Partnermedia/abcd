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

Typed links: `builds_on` [itd-199](../shipped/itd-199-a-reading-is-about-something-narrower-than-everything-its.md) (the scope operand and presets), [itd-198](../shipped/itd-198-an-assembly-reports-what-it-would-cost-before-a-reading-is.md) (the size report), [itd-196](../disciplines/itd-196-a-capability-is-rehearsed-end-to-end-before-it-ships-over-a.md) (rehearsal before shipping); `refines` [itd-199](../shipped/itd-199-a-reading-is-about-something-narrower-than-everything-its.md) (populated record lists, a declared window per preset).

## Press Release

> **A reading can be handed to a reader.** Each committed preset declares, per position, the estimated-token window it was calibrated for, and the assembler's own eval assembles every preset at every assembling position by dry run and fails when a result exceeds its declaration, so drift past the window is caught by the cold-reading eval lane, which is not yet a required check, rather than by the reader. The presets carry populated record lists where a position's object is a set of records, and a scope naming one record is proven to carry that record's material and nothing else. The assembler still enforces no budget at invocation; the preset states the bound, the eval checks it, and the size report goes on saying what a run would cost.

> "I want to point the entailment reading at the claim record and know before I spend a run that the bundle fits the reader," said an AI/agent researcher who commissions cold readings against their own repository. "And when the repository grows past what the preset was measured at, I want the build to tell me, not the reader."

## Why This Matters

The instrument was found correct and undeliverable one day before release: about 9.8 MB per position, roughly 2.45 million estimated tokens ([iss-2608311501186646](../../../work/issues/open/iss-2608311501186646-the-assembled-input-for-a-real-reading-of-this-repository-is.md)). [itd-198](../shipped/itd-198-an-assembly-reports-what-it-would-cost-before-a-reading-is.md) added the size report and [itd-199](../shipped/itd-199-a-reading-is-about-something-narrower-than-everything-its.md) the scope operand with committed presets. Measured on the v0.7.0 tree (commit 8f68ffb3) by dry-run assembly under the committed `cold` preset, the widening position assembles 257,835 bytes (about 67 thousand estimated tokens), the entailment position 942,074 bytes (about 245 thousand), and the detection position 751,612 bytes (about 195 thousand). Under `warm` the detection position is 2,528,029 bytes (about 657 thousand). The entailment and detection cold assemblies exceed or sit at the edge of a 200-thousand-token reader, and nothing in the record says what window any preset was calibrated for.

Two gaps in itd-199's delivery compound this. Every committed preset carries an empty record list, and the only assembly under a record-id scope in the test corpus selects nothing, so the positive half of the scope grammar has never been shown to carry a record's material. And the widening reading's bound, stated in advance by the design, is a construal of one or two sentences and a glossary of three to six terms, which is a statement about what is passed, not about what the tree weighs.

This intent closes the critical issue above; it does not close [iss-2608311501240566](../../../work/issues/open/iss-2608311501240566-three-of-the-four-reading-positions-receive-a-byte-identical.md), whose item-set distinctness itd-199 addressed by a selector-set proxy and which stays open until a gate compares item sets.

## What's In Scope

- **A declared window per preset and position** in the committed preset file, as an estimated-token figure the preset was calibrated to, using the size report's own byte-derived basis, together with the figure measured and the commit measured on. The preset file moves to schema version 2; at version 2 a position without a declaration is refused at load, and a version 1 file goes on loading as it does today, so an adopter's committed file keeps working until they move it.
- **An eval that assembles every committed preset at every assembling position by dry run** and fails when a result exceeds its declared window, naming the preset, the position, the measured figure and the declaration. It runs in the cold-reading eval lane; that lane becoming a gate that blocks a merge is the two recorded issues on it, which this intent names and does not close ([iss-2608311632382737](../../../work/issues/open/iss-2608311632382737-the-pre-push-gate-is-blind-to-both-eval-lanes-so-the-read-bl.md), [iss-2608311051046981](../../../work/issues/open/iss-2608311051046981-the-new-cold-reading-evals-ci-job-is-not-a-required-status-c.md)).
- **Populated record lists** where a position's object is a set of records: the entailment preset names the claim record it reads, and a test proves that a scope naming one record carries that record's projected material and no other record's.
- **An assembly over a 200 thousand token target says so.** Two hundred thousand estimated tokens is the target a preset aims at, and any figure inside a reader's window is acceptable; a dry run or an assembly whose estimate exceeds the target prints one line naming the figure and the target, as the maintainer ruled on 2026-09-02, so the operator learns it from the tool rather than from the reader.
- **Presets recalibrated** without dropping any kind the position's definition names as its object: the detection preset keeps a bounded slice of the shipped tree by path, and the entailment preset keeps its constraint sources. Where a cold entry measures above two hundred thousand estimated tokens the declaration says so and the record list is the lever that keeps the object while cutting size. The choice of kinds, records and paths is recorded in the file's own comment field.

## What's Out of Scope

- A budget enforced at invocation. The assembler reports; the preset declares; the eval checks. An operator overriding a preset at invocation is stamped as overridden, as now, and is not refused for size.
- A tokenizer. The basis stays byte-derived and labelled as such.
- The comparative position, whose object arrives by its own channel and is bounded by the widening run rather than by the tree.
- Item-set distinctness across positions, which stays with the issue named above.

## Mechanism

We expect a declared window checked by a dry-run eval to keep the instrument deliverable because the failure that reached release was one nobody measured, and a measurement that runs on every committed change in the eval lane is not one a builder forgets. It fails if the byte-derived estimate diverges from a reader's real count by more than the margin the declaration leaves, which the calibration sample recorded in spc-68 bounds.

## Scope Conditions

- The declared window is a property of the preset, not of any reader. Which reader a run is handed to is the operator's choice and is recorded on the run. <!-- cond: cond-2609020626042294 -->
- The eval measures the repository it runs in. A repository that adopts the presets measures its own tree; the figures recorded here are this repository's at the commit named. <!-- cond: cond-2609020626042487 -->
- Recalibration narrows what a position is handed and is recorded as such in the preset file. A narrowing that changes what the position's definition names as its object is a ruling and is not made here. <!-- cond: cond-2609020626045182 -->
- **The impact is `fix`.** The change closes a critical defect, and the schema move is opt-in by version, so no adopter's committed file stops loading. <!-- cond: cond-2609020626048715 -->

## Acceptance Criteria

- **Given** a preset file at schema version 2 in which one position carries no declared window, **when** it is loaded, **then** the assembler refuses and names the preset and the position.
- **Given** a preset file at schema version 1, **when** it is loaded, **then** it loads as it does today and the size report says no window is declared.
- **Given** the committed presets on the current tree, **when** the eval runs, **then** every assembling position under every preset measures at or below its declaration, and the eval passes.
- **Given** a preset whose declaration is set below its measured figure, **when** the eval runs, **then** it fails and names the preset, the position, the measured figure and the declaration.
- **Given** a scope naming one shipped intent, **when** the entailment assembly runs, **then** the bundle carries that intent's projected fields and no other intent's, and the manifest lists exactly those items.
- **Given** a committed preset with a populated record list, **when** the assembly runs, **then** the bundle carries the listed records' material as the position admits it and no unlisted record's.
- **Given** the preset file, **when** it is read, **then** each declaration carries the figure it was measured at and the commit it was measured on.
- **Given** an assembly or dry run whose estimate exceeds two hundred thousand tokens, **when** the size report renders, **then** it carries one line naming the figure and the target, and an assembly under the target carries no such line.

## Prior Art

- [itd-198](../shipped/itd-198-an-assembly-reports-what-it-would-cost-before-a-reading-is.md) (the size report), [itd-199](../shipped/itd-199-a-reading-is-about-something-narrower-than-everything-its.md) (the scope operand and presets), [itd-196](../disciplines/itd-196-a-capability-is-rehearsed-end-to-end-before-it-ships-over-a.md) (rehearsal before shipping).
- The cold-reading rulings of 2026-08-28 in the decision log.

## Open Questions

None. Whether a scope should be expressible as a set of record families was narrowed at itd-199's verdict and is not reopened here.

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._

## Grounds

- pursued: we expect a declared window per preset checked by a dry-run eval on every change to keep the instrument deliverable, because the size defect reached release only because nobody measured; a preset that passes its eval and still overflows the reader it was calibrated for by more than the recorded margin would show it wrong
