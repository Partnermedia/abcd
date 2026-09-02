---
id: itd-194
slug: the-reading-include-table-admits-only-what-the-exclusion-flo
spec_id: spc-2609021003136831
kind: standalone
suggested_kind: standalone
reclassification_history: []
builds_on: [itd-183]
severity: major
impact: fix
origin: researcher-authored
production_mode: hand-written
---

# The reading include table admits only what the exclusion floor can read

Typed links: `builds_on` [itd-183](../shipped/itd-183-the-cold-reading-sees-exactly-what-the-assembler-passes-posi.md) (the assembler, the include table and the exclusion floor); `refines` [itd-183](../shipped/itd-183-the-cold-reading-sees-exactly-what-the-assembler-passes-posi.md) (the floor refuses what it cannot resolve and the manifest asserts per item) and [itd-199](../shipped/itd-199-a-reading-is-about-something-narrower-than-everything-its.md) (a preset opts source and tests in, never the table by default); related [adr-56](../../decisions/adrs/0056-an-exclusion-control-asserts-only-what-it-can-prove.md) (an exclusion control asserts only what it can prove).

## Press Release

> **abcd ships with an assembler whose manifest says exactly what was
> examined.** The cold-reading include table and the exclusion floor now
> describe one set, item by item: a markdown document the floor cannot resolve
> is refused at admission and named, never passed through unscanned, and a
> source or test file, which the floor does not parse, is admitted only when a
> committed preset entry opts it in, travels whole, and is marked `unscanned` in the
> manifest. The manifest's exclusion assertion is made per item and only for
> the items the floor parsed, so a scan that ran and a scan that never ran no
> longer produce the same artefact. The widening position stops receiving the
> shipped intents, which neither design document lists in its object.
>
> The floor previously recognised a construct by the shape of the container it
> sat in. A record-shaped document inside a Go fixture was not markdown, so it
> was not scanned, so it travelled intact while the manifest asserted its
> headings had been refused. Nine rounds of adding patterns closed nine
> spellings of that mistake. This closes the tenth by refusing what cannot be
> resolved and disclosing what was not examined, instead of enumerating what
> may walk down the silent path.

## Why This Matters

An exclusion floor is a control, and a manifest is its attestation. When the
two disagree, the manifest is the side that is wrong, and a reader holding the
bundle and the manifest has no way to know it. That is the defect the itd-183
audit found live on this repository's own corpus, not in a crafted document:
the assembler admitted `internal/core/site/fixture_test.go` at item `itm-0736`
carrying three literal `## Audit Notes` sections with their acceptance
rollups, while the same run's manifest asserted `Audit Notes` was refused
(iss-2608301450065320).

Ten records stand against the floor. They collapse into one mechanism and
three consequences of it. The mechanism is container-shape gating: the scan
is switched off for a whole region or file because the container does not
match an expected shape, and the manifest goes on asserting exclusion anyway.
The file extension is one container (iss-2608301450065320). A fenced region
wrongly extended over the frontmatter is another, and it switches off the very
refusal that exists to catch keys the field reader cannot see
(iss-2608301350533102, the set's only critical). A frontmatter block that
never opens because line 0 is not a delimiter is a third
(iss-2608301237456350).

The rest are the same error at smaller grain. A key that is nested rather than
line-anchored is invisible to patterns anchored to the line
(iss-2608301237450573, iss-2608301251398360, which states outright that the
two want one fix and not two). A character the file's other readers treat as
whitespace, but the scan's character class does not, mis-parses a heading
opener and admits the heading under it (iss-2608301350534164,
iss-2608301421380392).

Three of those are verified leaking on this round and on its parent. A fourth
is live on the corpus at HEAD. The records are unanimous about what that
means: `a spelling arms race needs a design, not more patterns`
(iss-2608301251398360), and `fixing them one at a time is how this floor has
spent nine rounds`.

The maintainer ruled on 2026-09-02, checked against the design framework and
the readings companion, which both name code, tests, documentation and
configuration as the shipped tree a reading may see: source and tests reach
a reading only where a committed preset entry names them, an item so named
travels whole and marked unscanned, and the floor refuses markdown it cannot
resolve
and marks what it does not parse, so the manifest asserts the exclusion only
for the items it parsed (brief invariant 16). The strict alternative, removing
the test kind from the table, was declined because it contradicts both
documents; widening the floor to scan every type stays declined (ruled
2026-08-30). The same interview settled iss-2609012259587904 from the
documents: the framework's widening object and the companion's section 5.2
both list the widening object without the shipped intents, so the shipped
intent projection row's positions exclude widening, and that change lands
here with the include-table work.

## Mechanism

We expect refusal-at-admission for markdown, and an `unscanned` mark for what
the floor does not parse, to close this class where nine rounds of added
patterns did not, because every leak in the set shares one shape rather than
ten: the floor declined to examine something, and nothing downstream knew the
examination had not run. Enumerating what may pass leaves the silent-admission
path in place and re-opens on the next unenumerated spelling. Refusing a
markdown document the floor cannot resolve removes the path for the material
the floor claims to cover, and marking every unparsed item makes the
manifest's claim a fact about each item rather than a claim about the run.

This is falsifiable in two moves. Find a leak in a document the floor parsed
and the manifest marked as parsed: such a leak would show the defect was in
the floor's patterns after all, and that refusal bought nothing. Or find an
item the floor did not parse arriving in a bundle without the mark: that
would show the manifest still asserting more than the examination behind it.

## Scope Conditions

- The corpus is this repository's committed record. A reading assembled over <!-- cond: cond-2609021003132586 -->
  an external or mixed corpus is a different population and is not claimed
  here.
- The floor's parseable set is markdown whose frontmatter and headings it can <!-- cond: cond-2609021003134464 -->
  resolve. Every other admitted kind is disclosed as unscanned rather than
  examined; widening the parseable set is a separate change, and this intent's
  claim does not survive it unexamined.
- The committed detection and widening entries name the `source` and `test` kinds, <!-- cond: cond-2609021003130127 -->
  on the object-set ruling of 2026-09-02, and every such item arrives whole
  and marked `unscanned`; the entailment entry names neither. An entry that
  names either kind by commit accepts the unscanned items it selects,
  disclosed by the mark.
- No reading runs before this intent ships; every criterion below is judged <!-- cond: cond-2609021003139989 -->
  against the assembler's output and its manifest.

## What's In Scope

- The `source` and `test` kinds stay in the include table, so the object stays
  as the design documents state it, which for the detection and widening
  positions closes with the shipped tree: The committed detection and widening
  entries name both kinds, on the object-set ruling of 2026-09-02, and the
  entailment entry names neither. An entry that names one gets each item
  whole, and the manifest marks each such item `unscanned`.
- The floor refuses a markdown document it cannot resolve, naming the document
  and the shape, and never admits one unscanned: a fence delimiter inside the
  frontmatter block; a delimited block that does not open at line 0 but would
  to a reader of the bundle, because only blank lines, whitespace or an HTML
  comment precede it; a compact mapping nested in a block sequence and an
  explicit key in a flow mapping, which the line-anchored key reader cannot
  see; an attribute whose value opens on the line after its equals sign,
  which declines the markup mask; and a raw heading opener that reaches the
  end of the document with no bound, which is how a CRLF document's blank line
  was read over. Each extends the answer
  `unresolvableFrontmatterShape` already gives for one class to the whole
  set.
- The manifest's exclusion assertion is made per item, only for the items the
  floor parsed: an item carries a closed mark saying whether the key and
  heading signals were examined over it, and the assertion rests on that
  mark rather than on the run.
- The narrowing the include table performs is stated in the table itself:
  each row declares whether the floor parses what it admits, the rendered
  charter carries the declaration, and admission and examination are one
  declaration rather than a row on one side and an extension test on the
  other (iss-2608301450065320 names this as the undeclared scope decision
  underneath the whole class).
- The shipped intent projection row's positions exclude widening, the
  exclusion floor asserts it at that position, and the widening definition's
  admitted sources are regenerated (iss-2609012259587904, resolved by this
  intent).
- The include table admits the brief's meta, product, constraints, surfaces,
  internals and delivery chapters as brief sections, one row per chapter, at
  every position that reads the brief, because the design documents name
  "brief current text" as a reading's object and the table admits two chapters
  of it today. The rows are named individually because rule 1 forbids naming
  `brief/` whole: The directory contains the glossary, which keeps its own row.
  The evidence chapter stays excluded as verdict material, on the ground the
  Audit Notes exclusion rests on, with that ground stated in the table's rule
  text, and the manifest asserts the exclusion (iss-2609021153264023, resolved
  by this intent on the corrections ruling of 2026-09-02).

## What's Out of Scope

Four findings in the evidence set are inherited by this intent and fixed by
none of it. They are recorded here so a later reader does not mistake their
survival for an oversight:

- The quadratic title render on an unbounded heading opener, and the test
  whose name over-claims the linearity it does not cover
  (iss-2608301421382564). A cost finding about the scan's algorithm, not
  about what it excludes; the refusal of an unbounded opener removes that
  shape's admission and claims nothing about the scan's cost, and the test's
  name and its coverage stay with the issue.
- The two hand-written definitions of what opens a tag
  (iss-2608301251394412). Tech debt under the one-canonical-primitive rule;
  the record asks only that this intent inherit it.
- The doc comment describing fence behaviour the code does not have
  (iss-2608301237450573, second half).
- The escaped-key refusal whose message asserts two things that need not be
  true (iss-2608301421381157). The refusal is correct; only its stated reason
  is wrong.

Widening the floor's parseable set is also out of scope. This intent makes
admission and comprehension agree by disclosing at admission; moving
comprehension instead is the alternative the itd-183 audit named and the
maintainer did not choose (ruled 2026-08-30, confirmed 2026-09-02). Removing
the `test` kind from the table is out of scope for the opposite reason: it was
declined because it contradicts both design documents. The presets themselves
are not recalibrated here; that is the preset-windows intent's work, and
which committed entries name the two tree kinds is that intent's to record:
The detection and widening entries name both, the entailment entry neither.

## Acceptance Criteria

- **Given** a markdown document the include table admits whose frontmatter or
  headings the exclusion floor cannot resolve, **when** the assembler runs,
  **then** the document is refused, the refusal names the document and the
  shape, and no part of it reaches the bundle.
- **Given** a document whose frontmatter the floor cannot resolve, including a
  fence delimiter inside the frontmatter block, a delimited block preceded
  only by blank lines, whitespace or an HTML comment, a compact mapping nested
  in a block sequence, and an explicit key in a flow mapping, **when** the
  floor scans it, **then** the document is refused rather than admitted, and
  the same holds for an attribute whose value opens on the line after its
  equals sign and for a raw heading opener that reaches the end of the
  document with no bound, once a CRLF blank line counts as one.
- **Given** a committed preset entry that opts the `source` or `test` kind in,
  **when** the assembler runs, **then** each such item travels whole and its
  manifest entry is marked `unscanned`, and no item the floor parsed carries
  that mark.
- **Given** a bundle and its manifest, **when** the manifest asserts a
  construct was excluded, **then** the assertion is made per item and only for
  the items marked as parsed, so it rests on a scan that ran rather than on
  one that may have declined, and a manifest whose item carries no mark is
  refused by the decoder.
- **Given** this repository's corpus at HEAD and the committed presets,
  **when** the assembler runs under the committed entry at any assembling
  position, **then** the Go fixture carrying three `## Audit Notes` sections
  is absent from the bundle, and **when** an entry opts tests in, **then** the
  item recorded at `itm-0736` arrives marked `unscanned`, so the leak does not
  reproduce as a leak.
- **Given** the include table, **when** it is inspected or rendered into the
  charter, **then** each row states whether the floor parses what it admits,
  and a reader can see which documents the floor is claimed to cover without
  reading the floor.
- **Given** the widening position, **when** the assembler runs, **then** no
  shipped intent reaches the bundle, the manifest asserts the exclusion at
  that position, and the entailment and detection positions still receive the
  shipped intents projected as before.
- **Given** the include table, **when** the assembler runs at any position
  that reads the brief, **then** a document under each of the brief's meta,
  product, constraints, surfaces, internals and delivery chapters is admitted
  as a brief section, a document under the evidence chapter is refused, and
  the manifest asserts that exclusion with the verdict-material ground stated
  in the table's rule text.

## Open Questions

None. The two questions the draft carried are answered by the ruling. The
narrowing costs the committed entries nothing they do not disclose: The
detection and widening entries name source and tests, on the object-set ruling
of the same day, and every such item arrives marked `unscanned`, while the
entailment entry names neither; the fixture leak at `itm-0736` is absent from
every committed entry, whose paths do not reach `internal/core/site`, and
disclosed as unscanned wherever an entry opts that path's tests in. A refusal
is per document for markdown the floor cannot resolve, which keeps a reading
possible over the rest of the corpus and names the one document that stopped
it; an opted-in non-markdown item is never refused, because refusing it would
contradict the object the design documents state, and is marked unscanned
instead.

## Audit Notes

<!-- abcd-review: OWED receipt=rcp-11891aee84a6 -->
Fidelity review OWED (receipt rcp-11891aee84a6).

## Grounds

- pursued: we expect refusing markdown the floor cannot resolve, and marking every item the floor did not parse as unscanned in the manifest, to close the container-shape class that nine rounds of added patterns did not, because every leak shared one shape, a scan that declined with nothing downstream told; a leak in a document the manifest marks as parsed, or an unparsed item arriving without the mark, would show it wrong
