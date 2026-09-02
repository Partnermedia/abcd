# Readings

The charter of the readings family: the durable record of every cold-reading
assembly this repository performs.

A run's artefacts live at `.abcd/development/readings/<run-id>/`, where
`<run-id>` is minted as `rdg-<yymmddHHMMSS><rrrr>` (adr-45). The manifest
enumerates the items an assembly passed, by path, by field and by hash. It
carries no item content, so committing it needs no redaction.

It carries no timestamp field, but it is not timestamp-free: the run identifier
embeds a mint stamp by construction. What holds across two assemblies of one
repository state at one commit is that the bundle is byte-identical and the
manifest's items and exclusions are identical, the two manifests differing in
the run identifier and in nothing else. That is the determinism a re-run is
checked against, and it is why the manifest sits outside the amnesia eval's
byte comparison rather than inside it.

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

## One pile, and a position given its own

The four positions share **one assembly** by default. Three of them — widening,
comparative and detection — therefore receive a byte-identical item set, and
only entailment differs, by the drafts asymmetry rule 2 above states. That is
the default rather than an accident: the readers' definitions already tell each
position what to attend to, and a shared pile keeps the four readings
comparable (ruled 2026-09-01, `iss-2608311501240566`).

A position can be handed its own. The table's per-position section,
`PositionTables` beside the shared `Table`, declares a set of rows that
**replaces** the shared pile at one position, together with the rule saying why
that position is handed its own object. A position absent from the section
assembles from the shared table, byte for byte. Declaring a pile for one
position moves that position's assembly and no other's.

An own pile is held to everything the shared table is held to, through the same
validator rather than a copy of it: the two closed vocabularies, positive
selection at every grain, a stated admitting rule, rule 1 and rule 2 as the
exclusion floor states them. It has one rule of its own — every row must be
admitted at the position the pile is given to, because a pile is assembled at
that position and a row admitted elsewhere would be declared and never read.
The assembler validates the pile before it walks anything.

Each run says which pile it drew from. The manifest carries a `pile` stamp —
`shared` or `own`, with the hash of the rows — so a closing-run comparison can
tell a shared assembly from a narrowed one without re-deriving either, and
`abcd reading` reports the same for all four positions before any run happens.
The stamp is required rather than omitted when shared: a manifest that omits
it cannot distinguish a shared assembly from a stamp nobody wrote.

The section is Go data for the reason the shared table is. The table is the
whole of what a reading may see, its rendering is digested into the stamped
assembler version, and a test holds it against the rendering below. A runtime
configuration file able to add a row would be a channel for widening what a
reading sees that no committed record had to pass through.

**The comparative position is the case the section exists for, and the one
entry it cannot yet carry.** Its natural own pile is the widening reading's
admitted output, which lands at
`.abcd/work/issues/admissions/<run-id>/adm-N.md`. Two of the table's own rules
refuse that path, and both refuse it correctly: it sits inside the `.abcd`
namespace the structural deny refuses at every depth, and inside a record
family no include may name from above. The assembler has no channel for a prior
run's output at all, which is why the position refuses to assemble rather than
being served a corpus that is not its object (itd-199). So the entry below is
an **example of the shape**, not a declaration, and it does not assemble today:

```go
PositionComparative: {
    Rule: "the comparative position reads the widening run's admitted output, " +
        "not the shipped tree",
    Rows: []Row{{
        Positions: []Position{PositionComparative},
        Source:    ".abcd/work/issues/admissions", // refused: inside a record family
        Match:     []string{".md"},
        Kind:      KindIntentProjection,
        Rule:      "the candidate set is the prior reading's admitted output",
    }},
},
```

Supplying that object needs a channel for a prior run's output, which is a
separate change and not one an include row can make.

A row's field list is a contract rather than a census, and the two differ by
position today. The intent projection names five fields. `Scope Conditions` and
`Mechanism` are headings no shipped intent carries, so the shipped rows yield
three; two drafts already carry them, so the entailment position — the only one
that reads drafts — projects all five. A field a file does not carry contributes
no item.

A projected field ends where the redactor ends a section: at the next heading of
the same level or shallower. A subsection therefore travels with the field it
sits under, rather than being cut off at the first heading of any depth.

Two headings are the same heading when they fold to one another or render to
one another, the second measured by the site's own anchor slug — the equality the
whole floor is stated against.

Three shapes stay outside it and are disclosed rather than claimed: a heading
nested inside a blockquote or a list item, which no scan here reads as a
heading; one reaching an excluded title through a homoglyph or an invisible
format character, which slugs differently because the comparison is over code
points; and a double-encoded character reference, which decodes to the literal
text it renders as and is therefore a different heading, left as one on purpose.

Two further residues are shared with the site renderer and pre-date this
instrument. A fence delimiter indented four spaces is accepted here as a fence,
where CommonMark reads an indented code block, so a heading between two such
lines is masked and not seen. And a double-quoted scalar continued across lines,
carrying a brace on a continuation line, refuses: the scan reads one line at a
time, and the closing quote is not on it.

The heading signal is scoped to markdown. A heading and a frontmatter key are
things a record carries, and a source or configuration file carries neither, so
those travel whole with no section scan run over them. Within markdown a heading
is recognised however it is spelled: a closing sequence of hashes is normalised
away, a heading inside a fenced block is an example rather than a field and is
left alone, and an underlined heading is refused rather than redacted, because
the scan this projection spans by does not model one.

The bare verb reads run-directory NAMES under the assembler's own scratch
directory, and nothing else from `.abcd/.work.local/`. That is a listing of this
instrument's own staging, not a reader of the local ledger tier, and it is
stated here so it is not mistaken for one: no reading, and no part of this
assembler, opens a file in that tier.

A walk row's source directory must exist, and its absence refuses the run: a
brief chapter or the glossary going missing would otherwise enumerate nothing
and report clean. A record store's lifecycle BUCKET is different — an empty
bucket is a legitimate state of the record — so those rows stay silent, which is
disclosed residue rather than a check.

The exclusion floor is asserted into every manifest, each entry with the signal
by which a reader detects it, so the exclusions are checkable rather than taken
on trust. Its key and heading half is fail-closed like its path half: a file
that still carries an excluded key or heading after redaction refuses the run
rather than travelling with the manifest asserting otherwise. Prose-borne warmth inside an admitted chapter has no structural
signal: the chapter-level bound and the glossary discipline carry it, and it is
disclosed as residue.

<!-- BEGIN GENERATED: reading-include-table -->

### Include table

| Positions | Source | Matches | Suffixes | Fields | Store | Bucket | Kind | Admitting rule |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| widening, entailment, comparative, detection | `.abcd/development/brief/01-product` | `.md` | none | the whole file | every | every | `brief-section` | adr-55: the construal as it presently stands is committed record, admissible to every reader including a cold reading |
| widening, entailment, comparative, detection | `.abcd/development/brief/02-constraints` | `.md` | none | the whole file | every | every | `brief-section` | The constraints chapter states the platform, the dependency stance, the invariants and the naming a reading reads against |
| widening, entailment, comparative, detection | `.abcd/development/brief/glossary` | `.md` | none | the whole file | every | every | `glossary-term` | adr-55: the glossary's committed terms are committed record; superseded terms and the reasoning that settled them are not |
| widening, entailment, comparative, detection | `.abcd/development/intents/disciplines` | `.md` | none | the whole file | `itd` | `disciplines` | `discipline` | A discipline is a standing commitment the record already holds, named individually inside the intent family |
| widening, entailment, comparative, detection | `.abcd/development/intents/shipped` | `.md` | none | `Press Release`, `Acceptance Criteria`, `Scope Conditions`, `Mechanism`, `spec_id` | `itd` | `shipped` | `intent-projection` | Assembler rule 2: a shipped intent travels as its claim record, so the Audit Notes and dispositions it also carries stay behind |
| widening, entailment, comparative, detection | `.abcd/development/specs` | `.md` | none | the whole file | `spc` | every | `spec` | The design record a capability was built against |
| entailment | `.abcd/development/intents/drafts` | `.md` | none | `Press Release`, `Acceptance Criteria`, `Scope Conditions`, `Mechanism`, `spec_id` | `itd` | `drafts` | `intent-projection` | Assembler rule 2: articulation precedes selection, so entailment sees the candidate set and the reading asked to widen it does not |
| entailment | `.abcd/development/intents/planned` | `.md` | none | `Press Release`, `Acceptance Criteria`, `Scope Conditions`, `Mechanism`, `spec_id` | `itd` | `planned` | `intent-projection` | Assembler rule 2: articulation precedes selection, so entailment sees the candidate set and the reading asked to widen it does not |
| widening, entailment, comparative, detection | `.` | none | `_test.go` | the whole file | every | every | `test` | Assembler rule 1: the shipped tree is source and tests, counted apart because tests are the largest single class and admitted identically |
| widening, entailment, comparative, detection | `.` | `.go` | none | the whole file | every | every | `source` | Assembler rule 1: the shipped tree is source and tests, with the record, the definitions, the evals and the assembler's own package denied structurally |
| widening, entailment, comparative, detection | `.` | `.md` | none | the whole file | every | every | `doc` | Assembler rule 1: the shipped tree is the delivered documentation and the root prose, with the record denied structurally |
| widening, entailment, comparative, detection | `.` | `.json`, `.yml`, `.yaml`, `.toml`, `.mod`, `.sum`, `Makefile` | none | the whole file | every | every | `config` | Assembler rule 1: the shipped tree is the delivered configuration and build files, with the record denied structurally |

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
