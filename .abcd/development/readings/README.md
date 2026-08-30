# Readings

The charter of the readings family: the durable record of every cold-reading
assembly this repository performs.

A run's artefacts live at `.abcd/development/readings/<run-id>/`, where
`<run-id>` is minted as `rdg-<yymmddHHMMSS><rrrr>` (adr-45). The manifest
enumerates the items an assembly passed, by path, by field and by hash. It
carries no item content and no timestamp, so committing it needs no redaction
and two assemblies of one repository state produce one manifest.

The manifest's content is cold, which is why committing it for reader audit is
safe. The manifest as evidence is warm: it reveals run timing and target
selection. Both halves hold at once, and the second is why this family sits
inside the read block and no reading receives it.

## What a reading sees

The table below is generated from `internal/core/reading/include.go`, the
single source of truth, and a test holds the two to each other. Editing this
section by hand fails that test; edit the Go table and re-render.

Inclusion is positive at every grain: a position, a source directory, a file
match and, where a record travels as a projection, a named field list. What the
table does not name is absent, including a record family invented after the
table was written.

Two rules bind the table itself (ruled 2026-08-28):

1. No include names a directory that contains a record family. The deny is
   measured from a row's own source downward, so naming a family's leaf bucket
   individually is legal and reaching into a family from above is not.
2. A reading's object excludes the material whose state that reading exists to
   change. The drafts asymmetry and the Audit Notes exclusion are its two
   instances, which is what makes the table derivable rather than remembered.

A row's field list is a contract rather than a census. The intent projection
names five fields; `Scope Conditions` and `Mechanism` are headings no shipped
record carries yet, so three fields travel in this repository today and five
travel once those sections exist. A field a file does not carry contributes no
item.

A projected field ends where the redactor ends a section: at the next heading of
the same level or shallower. A subsection therefore travels with the field it
sits under, rather than being cut off at the first heading of any depth.

The exclusion floor is asserted into every manifest, each entry with the signal
by which a reader detects it, so the exclusions are checkable rather than taken
on trust. Its key and heading half is fail-closed like its path half: a file
that still carries an excluded key or heading after redaction refuses the run
rather than travelling with the manifest asserting otherwise. Prose-borne warmth inside an admitted chapter has no structural
signal: the chapter-level bound and the glossary discipline carry it, and it is
disclosed as residue.

<!-- BEGIN GENERATED: reading-include-table -->
### Include table

| Positions | Source | Matches | Fields | Admitting rule |
| --- | --- | --- | --- | --- |
| widening, entailment, comparative, detection | `.abcd/development/brief/01-product` | `.md` | the whole file | adr-55: the construal as it presently stands is committed record, admissible to every reader including a cold reading |
| widening, entailment, comparative, detection | `.abcd/development/brief/02-constraints` | `.md` | the whole file | The constraints chapter states the platform, the dependency stance, the invariants and the naming a reading reads against |
| widening, entailment, comparative, detection | `.abcd/development/brief/glossary` | `.md` | the whole file | adr-55: the glossary's committed terms are committed record; superseded terms and the reasoning that settled them are not |
| widening, entailment, comparative, detection | `.abcd/development/intents/disciplines` | `.md` | the whole file | A discipline is a standing commitment the record already holds, named individually inside the intent family |
| widening, entailment, comparative, detection | `.abcd/development/intents/shipped` | `.md` | `Press Release`, `Acceptance Criteria`, `Scope Conditions`, `Mechanism`, `spec_id` | Assembler rule 2: a shipped intent travels as its claim record, so the Audit Notes and dispositions it also carries stay behind |
| widening, entailment, comparative, detection | `.abcd/development/specs` | `.md` | the whole file | The design record a capability was built against |
| entailment | `.abcd/development/intents/drafts` | `.md` | `Press Release`, `Acceptance Criteria`, `Scope Conditions`, `Mechanism`, `spec_id` | Assembler rule 2: articulation precedes selection, so entailment sees the candidate set and the reading asked to widen it does not |
| entailment | `.abcd/development/intents/planned` | `.md` | `Press Release`, `Acceptance Criteria`, `Scope Conditions`, `Mechanism`, `spec_id` | Assembler rule 2: articulation precedes selection, so entailment sees the candidate set and the reading asked to widen it does not |
| widening, entailment, comparative, detection | `.` | `.go` | the whole file | Assembler rule 1: the shipped tree is source and tests, with the record, the definitions, the evals and the assembler's own package denied structurally |
| widening, entailment, comparative, detection | `.` | `.md` | the whole file | Assembler rule 1: the shipped tree is the delivered documentation and the root prose, with the record denied structurally |
| widening, entailment, comparative, detection | `.` | `.json`, `.yml`, `.yaml`, `.toml`, `.mod`, `.sum`, `Makefile` | the whole file | Assembler rule 1: the shipped tree is the delivered configuration and build files, with the record denied structurally |

### Exclusion floor

| Detail | Signal | Rule | Positions |
| --- | --- | --- | --- |
| `origin` | frontmatter key | field projection | every position |
| `production_mode` | frontmatter key | field projection | every position |
| `Audit Notes` | heading | field projection | every position |
| `Open Questions` | heading | field projection | every position |
| `Why This Matters` | heading | field projection | every position |
| `Scope Condition Dispositions` | heading | a reading's object excludes what it exists to change | every position |
| `.abcd/development/brief/03-evidence` | directory | absent from the positive walk | every position |
| `.abcd/development/decisions` | directory | absent from the positive walk | every position |
| `.abcd/development/roadmap/rfcs` | directory | absent from the positive walk | every position |
| `.abcd/development/intents/superseded` | directory | absent from the positive walk | every position |
| `.abcd/development/plans` | directory | absent from the positive walk | every position |
| `.abcd/development/research/notes` | directory | absent from the positive walk | every position |
| `.abcd/work/issues` | directory | no include names a directory containing a record family | every position |
| `.abcd/work/DECISIONS.md` | file | absent from the positive walk | every position |
| `the lapse log` | record type in a denied path | absent from the positive walk | every position |
| `admission and selection grounds` | record type in a denied path | absent from the positive walk | every position |
| `.abcd/development/readings` | directory | the instrument's own output is never its input | every position |
| `agents` | directory | the instrument's own output is never its input | every position |
| `evals` | directory | the instrument's own output is never its input | every position |
| `internal/core/reading` | directory | the instrument's own output is never its input | every position |
| `the session-transcript store` | unreachable path | the store sits outside the repository tree | every position |
| `.abcd/development/intents/drafts` | directory | a reading's object excludes what it exists to change | widening, comparative, detection |
| `.abcd/development/intents/planned` | directory | a reading's object excludes what it exists to change | widening, comparative, detection |

<!-- END GENERATED: reading-include-table -->

## What the assembler does not do

The bundle is pathless by construction: each item is a key, a material class
and its text, and only the manifest maps a key back to a path and a field. The
remaining half of the isolation, that a dispatching host grants the reader no
repository access, is a host obligation stated in each definition, never an
enforcement this binary performs.

Timing and target selection stay the operator's choice. The manifest and the
run record make that visible after the fact rather than preventing it.
