---
id: itd-2609020625407419
slug: a-comparative-reading-receives-the-widening-run-s-items-as-i
spec_id: spc-2609020626039834
kind: standalone
suggested_kind: null
reclassification_history: []
builds_on: [itd-199, itd-191, itd-186, itd-183, itd-180]
severity: major
impact: additive
origin: researcher-authored
production_mode: dictated-and-formatted
---

# A comparative reading receives the widening run's items as its candidate set, and characterises them before anyone admits one

Typed links: `builds_on` [itd-199](../shipped/itd-199-a-reading-is-about-something-narrower-than-everything-its.md) (the comparative refusal and the scope operand), [itd-191](../disciplines/itd-191-the-selection-criteria-are-a-declared-recorded-discipline-a.md) (the criteria discipline), [itd-186](../shipped/itd-186-the-read-block-eval-falsifies-the-firewall-planted-warm-cont.md) (the read-block eval's prior-run exhaust), [itd-183](../shipped/itd-183-the-cold-reading-sees-exactly-what-the-assembler-passes-posi.md) (field projection and the manifest), [itd-180](../shipped/itd-180-a-cold-reading-s-findings-land-as-reading-records-and-the-re.md) (reading records); `refines` [itd-199](../shipped/itd-199-a-reading-is-about-something-narrower-than-everything-its.md) (withdraws its refusal of a comparative preset, flagged for the maintainer) and [itd-186](../shipped/itd-186-the-read-block-eval-falsifies-the-firewall-planted-warm-cont.md) (a positional exception to the prior-run exhaust, flagged for the maintainer).

## Press Release

> **The comparative reading can now be commissioned.** `abcd reading assemble --position comparative` takes the identifier of a widening run and hands the reading that run's returned configurations, projected to the two fields the widening body carries, together with the six declared selection criteria. It hands over nothing else from the readings store: no disposition, no admission, no manifest, no other run. If any item of the named run already carries a disposition or an admission, the assembly refuses, because the design fixes the order as characterise first and admit second. If the run returned fewer than two configurations, the assembly refuses and records that the position was not exercised, which is the interpretation fixed before any run. At ingest, every item names a candidate from that run and a criterion from the discipline, and an ordering, a score or a recommendation across candidates is refused as it already is.

> "I want to see how each of the reading's proposals ordinarily behaves against the criteria we said we would select on, before I commit to any of them," said an AI/agent researcher who runs the four positions against their own record. "The reading that widened my candidate set must not be the one that tells me which candidate to take, and the one that characterises them must not see which I took."

## Why This Matters

The cold-reading rulings adopted on 2026-08-28 fix the comparative reading's object as the widening reading's output, being the configurations returned at Step 2 before admission, and admit no other source; they fix the ordering as Step 2 before Step 4, with admission a warm act performed after the comparative reading, because running admission first would leave the comparative reading nothing uncommitted to characterise; and they fix the interpretation in advance: where the widening reading returns fewer than two configurations, the comparative reading has nothing to compare and is not exercised, and that outcome is recorded as such.

v0.7.0 ships the definition, the regime gate and the criteria discipline ([itd-191](../disciplines/itd-191-the-selection-criteria-are-a-declared-recorded-discipline-a.md)), and the position refuses to assemble because its object "is not repository material and has no channel today" ([itd-199](../shipped/itd-199-a-reading-is-about-something-narrower-than-everything-its.md)). That refusal was the right move over serving the detection corpus, and it leaves one of the four positions structurally unavailable until the channel exists. itd-199 scoped the channel out as "a separate intent", and this is that intent.

The channel is delicate because the readings store sits inside the read block by ruling: prior manifests and reading records never reach a reading ([itd-186](../shipped/itd-186-the-read-block-eval-falsifies-the-firewall-planted-warm-cont.md), ac-4), and the general rule is that no reading sees another's output. The comparative object is the one ruled exception, and it is narrower than it sounds. What the reading needs is the candidate text as it was returned, which is cold: the widening reading produced it without ledger access. What it must not see is what has happened to those candidates since, which is warm: dispositions, admissions, surprises, and any other run. The channel therefore projects two fields out of one run's items and asserts the exclusion of everything else in the same store, which is exactly the field projection the assembler already performs on shipped intents.

## Decisions flagged for the maintainer

Two shipped rulings move if this intent is adopted, and the record must say so rather than have it discovered.

- **The prior-run exhaust gains one positional exception.** itd-186's ac-4 and the include table's exclusions state that the instrument's own output is never its input. This intent admits, at the comparative position only, two body fields of one named widening run's items. That is a trust-boundary rule and belongs in an ADR with a brief invariant amendment, captured as an issue in the same change; the intent carries the operand, the projection and the refusals, and does not ship until the ADR is adopted.
- **The comparative preset refusal is withdrawn.** itd-199's ac-10 refuses a preset naming a comparative scope. The committed presets gain a comparative entry naming the criteria discipline, carrying the window declaration the presets intent requires.
- **A fourth closed operand.** adr-58 admits a third operand "and no other", and brief invariant 15 enumerates three. The candidate-run operand is a fourth, shape-validated and never prose, so the ADR this intent waits on also supersedes adr-58 to that extent and amends the invariant's enumeration; the binding property, that no operand carries prose, is unchanged.

## What's In Scope

- **A candidate-run operand** on the comparative assembly, naming one widening run by its run identifier, in the closed shape the run mint produces. No other position accepts it, and the comparative position refuses without it. The operand is not prose, which keeps adr-58's binding property.
- **Projection of the named run's items** to the two widening body fields, configuration and what admits it, plus the item identifier the comparative body must cite. The item's pattern, its envelope, and every other field stay behind.
- **The criteria discipline** ([itd-191](../disciplines/itd-191-the-selection-criteria-are-a-declared-recorded-discipline-a.md)) passed with the candidates, as the declared criteria the reading characterises against, and the definition's Object section updated to name both.
- **The ordering as a refusal.** If any item of the named run has a standing disposition or an admission, the assembly refuses and names the item, because the candidate set is defined as pre-admission. The gate that makes this unreachable through any verb sits in the shared disposition writer (the admission intent): at the widening position no disposition can be written until a committed comparative run names the run.
- **The fixed interpretation as a refusal.** If the named run holds fewer than two items, the assembly still stages a comparative run with an empty candidate set, and ingesting it commits a comparative run with an empty item set naming that widening run, which is the clean-run idiom the design fixes in advance. The comparative outcome for a widening run is therefore always a committed comparative run naming it, never a mutable file, and the admission verb reads exactly that.
- **The manifest** records the candidate run, the projected fields, and the exclusion of the rest of the readings store, so a reader can see that the comparative reading saw candidates and not their fate.
- **Ingest at the comparative position** checks that every item's candidate identifier names an item of the run the manifest records, and that every criterion is one the discipline declares.
- **The read-block eval** gains a comparative case: with dispositions, admissions and a second run planted, the comparative bundle carries the named run's two candidate fields and nothing else from the store.
- **The include table is the whole account at this position.** The candidate channel is a table row of its own (source the readings store, two fields, kind candidate, comparative only), so the rendered charter and the assembler version carry it. Every other row excludes the comparative position except the disciplines row, which at this position is scoped to the criteria discipline, so no scope can hand the comparative reading the tree, the brief or the intents: no other source is admitted.

## What's Out of Scope

- Admission itself, which is itd-189's record and the admission verb's to write.
- Any characterisation the assembler performs. It passes candidates and criteria; the reading characterises.
- Widening beyond the named run. A comparative reading is about one widening run, and two runs' candidates are two comparative runs.

## Mechanism

We expect a field projection from one run's reading records to hold the read block where a path rule cannot, because the warm half of the readings store (dispositions, admissions, other runs) is separated from the cold half (returned candidate text) at field and directory granularity, and the assembler already asserts field exclusions fail-closed. This is falsifiable in one move: plant a disposition on a candidate and show its text in the comparative bundle.

## Scope Conditions

- The candidate set is exactly one widening run's items. A candidate set drawn from more than one run, or from anything other than a widening run, is out of scope and would need its own ruling. <!-- cond: cond-2609020626038511 -->
- The ordering refusal assumes the admission and disposition stores are the only places a candidate's fate is recorded. A new store recording something about a candidate is excluded from the reading by positive inclusion but is not, by itself, a reason to refuse assembly. <!-- cond: cond-2609020626033649 -->
- The pattern named by the widening item is withheld from the comparative reading on the ground that provenance is the envelope's and not the candidate's. If a criterion turns out to need it, the projection widens by one field and the manifest says so. <!-- cond: cond-2609020626032295 -->

## Acceptance Criteria

- **Given** a comparative invocation with no candidate run, **when** it runs, **then** the verb refuses and names the operand and its shape.
- **Given** a widening run with three items, none dispositioned or admitted, **when** the comparative assembly runs against it, **then** the bundle carries each item's identifier, configuration and what admits it, and the criteria discipline, and no other field of those items.
- **Given** the same run with one item carrying a standing disposition, **when** the comparative assembly runs, **then** it refuses and names that item.
- **Given** the same run with one item admitted, **when** the comparative assembly runs, **then** it refuses and names that item.
- **Given** a widening run with one item, **when** the comparative assembly runs, **then** it refuses, names the fixed interpretation, and a committed comparative run with an empty item set names that widening run.
- **Given** a repository holding two widening runs, dispositions and admissions, **when** the comparative assembly runs against one run, **then** no text from the other run, from any disposition or from any admission appears in the bundle, and the manifest asserts their exclusion.
- **Given** a comparative output whose item names a candidate not in the recorded run, **when** it is ingested, **then** the verb refuses and names the item.
- **Given** a comparative output whose item names a criterion the discipline does not declare, **when** it is ingested, **then** the verb refuses and names the item and the criterion.
- **Given** the read-block eval with a disposition planted on a candidate of the named run, **when** it runs, **then** it fails if the disposition's text reaches the comparative bundle.

## Prior Art

- The cold-reading rulings of 2026-08-28 in the decision log, and the design framework and its readings companion held outside the repository.
- [itd-199](../shipped/itd-199-a-reading-is-about-something-narrower-than-everything-its.md), [itd-186](../shipped/itd-186-the-read-block-eval-falsifies-the-firewall-planted-warm-cont.md), [itd-191](../disciplines/itd-191-the-selection-criteria-are-a-declared-recorded-discipline-a.md).

## Open Questions

None beyond the two flagged decisions above. The ordering and the fewer-than-two interpretation are ruled, and this intent implements them as refusals rather than reopening them.

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._

## Grounds

- pursued: we expect a field projection out of one widening run's items to give the comparative reading its ruled object without opening the read block to the rest of the readings store, because the cold half (returned text) and the warm half (dispositions, admissions, other runs) of that store are separable at field and directory grain; a planted disposition reaching the comparative bundle would show it wrong
