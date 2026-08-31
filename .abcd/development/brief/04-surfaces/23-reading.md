# `/abcd:reading` — The Cold-Reading Input Assembler

`/abcd:reading` assembles the input a cold reading is handed. The location
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

## The invocation carries no free text

`assemble` takes exactly two operands, both closed in shape.

| Operand | Grammar |
|---|---|
| `--position` | one of `widening`, `entailment`, `comparative`, `detection` |
| `--target` | `HEAD`, or a hexadecimal commit sha of 7 to 40 digits |

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

Ruling (18) is held on both sides of the run. An output directory the include
table can reach is refused when it is named, because writing a run where the
table reaches it commits the next run's contamination. And both artefacts are
refused as input wherever an admitted path holds one, recognised by the
top-level `_type` tag they carry, so a run committed before that refusal existed
cannot ride in either.

A run identifier is `rdg-<yymmddHHMMSS><rrrr>`, minted per adr-45: the mint
reads no maximum, so two checkouts assembling in the same window cannot
converge on one id.

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
