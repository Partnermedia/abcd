# `/abcd:reading` — The Cold-Reading Instrument Surface

`/abcd:reading` assembles the input a cold reading is handed and validates the
output it returns. The location
tiering the repository already has is organisational, not an access control:
nothing in it prevents a reading reaching ledger content, so every claim an
instrument makes about what it saw rests on a disclosure taken on trust. This
surface is the point at which that becomes checkable.

Blindness becomes a property of the input. A positive include table names what
may travel, at field granularity; a bundle carries no repository path; and a
hashed manifest records what was passed, so a third party can re-run the
assembly and diff the result.

## Sub-verbs

> _Machine-checked (`surface_coverage`, spc-27): each row records the verb's
> adr-40 bucket (`lint` / `review` / `audit` / `gate`, or `—` for a
> non-assessment verb) and its existence (`shipped` / `staged`), verified
> against the committed command-tree snapshot in both directions._

| Verb | Bucket | Status |
|---|---|---|
| `assemble` | — | shipped |
| `ingest` | gate | shipped |

## The invocation carries no free text

`assemble` takes exactly three operands, every one closed in shape. It carries
no prose: that is the property the 2026-08-28 rulings protected by closing the
invocation at two, and [adr-58](../../decisions/adrs/0058-a-reading-is-commissioned-about-something-so-the-invocation-takes-a-scope.md) restates it as the rule that binds with a third
operand admitted.

| Operand | Grammar |
|---|---|
| `--position` | one of `widening`, `entailment`, `comparative`, `detection` — though `comparative` does not assemble and refuses |
| `--target` | `HEAD`, or a hexadecimal commit sha of 7 to 40 digits |
| `--scope` | a record id (`itd-N`, `spc-N`), a material kind, or a committed preset |

**No repository path is accepted at the invocation.** A path may be named only
inside the committed preset file, where it is reviewed, shape-validated and
inside the dirty gate.

A branch name or a tag is refused as mutable: the manifest's re-runnability
rests on a reference that cannot move. A positional argument is refused
outright. The reading's object and its question come from its definition, so
there is no channel through which ledger content can travel in the framing of a
request.

Assembly reads the working tree, so it refuses unless HEAD resolves to the
target and no included path is uncommitted — a dirty tree cannot be described by
a commit reference, and the manifest would promise a re-run it could not
deliver. Every refusal exits 2.

## Two rules bind the include table

1. **No include names a directory that contains a record family.** The deny is
   measured from a row's own source downward, so a family's leaf bucket may be
   named individually while a directory above a family may not. A family added
   later, the readings family itself included, is excluded by construction.
2. **A reading's object excludes the material whose state that reading exists to
   change.** The drafts asymmetry and the Audit Notes exclusion are its two
   instances: the widening position cannot see the candidate set it is asked to
   widen, and a shipped intent travels as its claim record rather than with the
   audit written against it.

The table itself is Go data in `internal/core/reading/include.go`, rendered into
the readings family's charter under a test holding the two to each other. The
exclusion floor rides in every manifest, each entry with the signal by which a
reader detects it.

## Two artefacts, and where they land

`assemble` writes the assembled input (`bundle.json`) and the manifest
(`manifest.json`) as two separate files: the input goes to a reader, the
manifest stays with the auditor. `--out <dir>` names the directory, which must
be empty or absent — one run's artefacts are one run's evidence — and each file
is written through a temporary name and renamed into place. Without `--out` they
land in
`.abcd/.work.local/scratch/reading-runs/<run-id>/`. With `--dry-run` and no
`--out`, nothing is written and the result is rendered only.

## What a reading would cost

Every assembly, `--dry-run` included, reports the size of the item text it
assembled — bytes and an estimated token count, in total and per kind — so what
a reading would be handed is known before one is commissioned (itd-198). The
estimate is byte-derived (bytes divided by 3.85), never a tokenizer's count, and
the render says so beside the number rather than letting a reader take it for a
measurement.

Ruling (18) is held on both sides of the run. An output directory the include
table can reach is refused when it is named, because writing a run where the
table reaches it commits the next run's contamination. And both artefacts are
refused as input wherever an admitted path holds one, recognised by the
top-level `_type` tag they carry, so a run committed before that refusal existed
cannot ride in either.

A run identifier is `rdg-<yymmddHHMMSS><rrrr>`, minted per adr-45: the mint
reads no maximum, so two checkouts assembling in the same window cannot
converge on one id.

## `ingest` checks what the reading was licensed to produce

`ingest --reading-json <path>` validates the JSON a reading returned and writes
its reading records. It is the output-contract idiom the repository already
carries — an agent emits JSON, a deterministic verb validates it, the verb
writes the record — and it adds a check no structural schema performs: what the
reading was LICENSED to produce, not only what it saw.

Three properties distinguish it. **Item identifiers are minted by the verb**, so
the payload carries none and one it supplies is refused as an unknown field.
**The supply regime is the definition's**, read from the position's definition
file and compared against the output's own claim, with no operand and no
configuration key able to reach it. And **named provenance is enforced at every
regime without exception**: an item whose `pattern` is empty or absent is
refused, which the definitions instruct and nothing else checks.

Refusal has two granularities. An item-level violation refuses that item and
lands the rest, naming the ordinal, the rule and the field. A list-level
violation refuses the run, and **a refused run still leaves a durable record**:
`refusal.json` carries the run metadata and the named reason and no items, so a
rerun is a new run with a new run id rather than an amendment.

**The layer that refuses is structural.** Each regime declares reserved names,
and an item carrying one as a field of its own is refused with the licence
stated. The table is read at the run's own regime, one row per regime, and the
generative regime has no row — its licence is the widest, and the constraint on
it falls at admission. That layer carries no bound of the kind prose does,
because a field is present or it is not.

**Nothing an item's prose says refuses it.** A registry of named signatures
reads the item's text and records a hit on the run record as a review flag; all
four ship observed rather than enforcing, because a registry cannot tell a
reading that proposes from one reporting that the document proposes, and an
enforcing one refuses a reading for quoting its own material. Its bound is
disclosed on itd-185: a fix proposal or a disposition phrased outside the
registry's signatures is not seen at all.

Writes are staged. No OTHER run's durable state is written to or deleted from
until the whole payload validates — a refusal after the run is proven writes its
refusal record and nothing else, and the one delete it makes is on its own run
id, rolling back the records of an earlier attempt at it that never committed;
the reading records land in the reading-record family; and the run metadata is
written **last**, as the commit marker, so a run without one never happened. An
ingest interrupted before that marker leaves a stage in the local tier. Every
later invocation names the orphan; the next one whose payload validates rolls
that run back and clears the stage, and a stage left by a run that did commit is
a leftover — that run stands, and only the stage goes. A refused run reports the
orphans it left in place: the sweep is a delete in the committed tier, and a
refused run never reaches one.

## What this surface does not claim

It never runs a reading. It produces the input a reading would be given;
dispatching that input to a reader is host work.

The bundle is pathless by construction, which is the half of the isolation the
binary enforces. The other half — that the dispatching host grants the reader no
repository access — is a host obligation stated in the plugin surface, and it is
disclosed as an obligation rather than claimed as an enforcement.

Timing and target selection stay the operator's choice. The manifest and the run
record make that visible after the fact rather than preventing it. Prose-borne
warmth inside an admitted chapter has no structural signal: the chapter-level
include bound and the glossary discipline carry it, and it is disclosed as
residue.

## References

- Plugin command: [`commands/reading.md`](../../../../commands/reading.md)
- The family's charter and the rendered include table:
  [`.abcd/development/readings/README.md`](../../readings/README.md)
- The construal's admissibility:
  [adr-55](../../decisions/adrs/0055-the-construal-stands-in-the-record-its-history-does-not.md)
- Invariants 14 and 15: [`03-invariants.md`](../02-constraints/03-invariants.md)
