---
id: itd-194
slug: the-reading-include-table-admits-only-what-the-exclusion-flo
spec_id: null
kind: null
suggested_kind: standalone
reclassification_history: []
builds_on: [itd-183]
severity: major
impact: fix
origin: researcher-authored
production_mode: hand-written
---

# The reading include table admits only what the exclusion floor can read

## Press Release

> **abcd ships with an assembler that refuses what it cannot read.** The
> cold-reading include table and the exclusion floor now describe the same set
> of documents: a file the table admits is a file the floor can parse, and a
> file the floor cannot parse is refused at admission rather than passed
> through unscanned. The manifest's exclusion claim stops being the output of a
> scan that may have declined to run, and becomes true by construction.
>
> The floor previously recognised a construct by the shape of the container it
> sat in. A record-shaped document inside a Go fixture was not markdown, so it
> was not scanned, so it travelled intact while the manifest asserted its
> headings had been refused. Nine rounds of adding patterns closed nine
> spellings of that mistake. This closes the tenth by removing the path
> instead of enumerating what may walk down it.

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

## Mechanism

We expect refusal-at-admission to close this class where nine rounds of added
patterns did not, because every leak in the set shares one shape rather than
ten: the floor declined to scan something, and nothing downstream knew the
scan had not run. Enumerating what may pass leaves the silent-admission path
in place and re-opens on the next unenumerated spelling. Refusing what cannot
be parsed removes the path itself, and makes the manifest's claim structural
rather than empirical.

This is falsifiable in one move: find a leak in a document the floor did
parse. Such a leak would show the defect was in the floor's patterns after
all, and that narrowing admission bought nothing.

## Scope Conditions

- The corpus is this repository's committed record. A reading assembled over
  an external or mixed corpus is a different population and is not claimed
  here.
- The floor's parseable set is markdown whose frontmatter it can resolve.
  Widening that set is a separate change, and this intent's claim does not
  survive it unexamined.
- No reading runs this cycle. Every criterion below is judged against the
  assembler's output and its manifest, never against a reader's use of them.

## What's In Scope

- The include table refuses a document whose type the floor cannot parse,
  naming it, rather than admitting it unscanned.
- The floor answers a construct it cannot resolve with a refusal, extending
  the pattern `unresolvableFrontmatterShape` already holds for one class to
  every class in the set above.
- The narrowing the include table performs is stated in the table, rather
  than inherited unstated by the floor (iss-2608301450065320 names this as the
  undeclared scope decision underneath the whole class).

## What's Out of Scope

Four findings in the evidence set are inherited by this intent and fixed by
none of it. They are recorded here so a later reader does not mistake their
survival for an oversight:

- The quadratic title render on an unbounded heading opener, and the test
  whose name over-claims the linearity it does not cover
  (iss-2608301421382564). A cost finding about the scan's algorithm, not
  about what it excludes.
- The two hand-written definitions of what opens a tag
  (iss-2608301251394412). Tech debt under the one-canonical-primitive rule;
  the record asks only that this intent inherit it.
- The doc comment describing fence behaviour the code does not have
  (iss-2608301237450573, second half).
- The escaped-key refusal whose message asserts two things that need not be
  true (iss-2608301421381157). The refusal is correct; only its stated reason
  is wrong.

Widening the floor's parseable set is also out of scope. This intent makes
admission and comprehension agree by moving admission; moving comprehension
instead is the alternative the itd-183 audit named and the maintainer did not
choose (ruled 2026-08-30).

## Acceptance Criteria

- **Given** a document the include table would admit whose type the exclusion
  floor cannot parse, **when** the assembler runs, **then** the document is
  refused and named in the refusal, and no part of it reaches the bundle.
- **Given** a bundle and its manifest, **when** the manifest asserts a
  construct was excluded, **then** every document in the bundle is one the
  floor parsed, so the assertion rests on a scan that ran rather than on one
  that may have declined.
- **Given** this repository's corpus at HEAD, **when** the assembler runs,
  **then** the Go fixture carrying three `## Audit Notes` sections is refused
  rather than admitted, and the leak recorded at item `itm-0736` does not
  reproduce.
- **Given** a document whose frontmatter the floor cannot resolve, including a
  fence delimiter inside a block scalar and a frontmatter block that does not
  open at line 0, **when** the floor scans it, **then** the document is
  refused rather than silently admitted.
- **Given** the include table, **when** it is inspected, **then** the
  narrowing it performs is stated in the table itself, and a reader can see
  which documents the floor is claimed to cover without reading the floor.

## Open Questions

- What the narrowing costs. Refusing what cannot be parsed shrinks the corpus
  a reading sees, and nobody has measured by how much. The measurement wants
  taking before the first reading runs, not after.
- Whether a refusal is per-document or refuses the whole assembly. The
  per-document answer keeps a reading possible over a mixed tree; the
  whole-assembly answer is the louder one. The records do not settle it.

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._
