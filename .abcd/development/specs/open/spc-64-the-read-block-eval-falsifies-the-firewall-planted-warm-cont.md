---
id: spc-64
slug: the-read-block-eval-falsifies-the-firewall-planted-warm-cont
intent: itd-186
---
# The read-block eval falsifies the firewall

## Summary

spc-64 delivers [itd-186](../../intents/planned/itd-186-the-read-block-eval-falsifies-the-firewall-planted-warm-cont.md):
a repository eval that plants sentinel warm content across every warm location
class in a fixture repository state, runs the cold-reading input assembler
([itd-183](../../intents/planned/itd-183-the-cold-reading-sees-exactly-what-the-assembler-passes-posi.md),
spec [spc-61](spc-61-the-cold-reading-sees-exactly-what-the-assembler-passes-posi.md))
over it, and asserts that no sentinel and no excluded field reaches the
assembled input. It is the only component in the workstream capable of
falsifying the read-block rather than restating it, so its oracle is the planted
corpus, never the assembler's include table.

The eval lands in the existing smoke harness package
([`evals/`](../../../../evals/README.md)), behind the `smoke` build tag, so
`make smoke` and CI's `smoke` job pick it up with no change to
[`Makefile`](../../../../Makefile) or
[`.github/workflows/ci.yml`](../../../../.github/workflows/ci.yml). Passing in
CI is a cycle exit criterion. spc-61 is being written concurrently; where it is
still a stub, this spec designs against itd-183's own text and the assembler row
of the cycle build plan, and states under [Approach](#approach) the four demands
it places on the assembler's interface, which are deliberately the whole of the
coupling.

## Scope

In: the sentinel fixture corpus and its plants, one per warm location class; the
independent oracle and the import guard that keeps it independent; the
field-level and content-level assertions over the assembled input at every
position the assembler accepts; the negative-control fixture that proves the
eval fails when the firewall is holed; the anti-vacuity guard that proves every
plant is still planted. Out: everything under
[Out of scope](#out-of-scope), and in particular any assertion about what a
reading *produced*, which belongs to the output contract
([itd-185](../../intents/planned/itd-185-one-ingest-verb-validates-every-cold-reading-output-includin.md)).

## Approach

### The oracle, and how it stays independent of the include table

Four mechanisms, each structural rather than promised:

- **Out of process.** The eval invokes the assembler through the built binary,
  reusing the harness's existing `TestMain` build and `run` helper in
  [`evals/smoke_test.go`](../../../../evals/smoke_test.go). It never links the
  assembler's package, so it cannot read the include table even by accident.
- **An import guard.** `TestOracleImportsNothingFromTheAssembler` parses every
  Go file in `evals/` with `go/parser` and fails if any import path names the
  assembler's package or `internal/core/launch`'s include list. The
  independence claim is therefore checked, not asserted in a comment.
- **A hand-transcribed exclusion table.** The eval's table of excluded
  frontmatter keys, headings, and record families is transcribed by hand from
  itd-183's exclusion list, brief invariants 14 and 15
  ([`03-invariants.md`](../../brief/02-constraints/03-invariants.md)), and
  [adr-55](../../decisions/adrs/0055-the-construal-stands-in-the-record-its-history-does-not.md).
  Transcription is the point: a generated table would agree with the assembler
  by construction and could only ever confirm it.
- **A content oracle, not a path oracle.** Every assertion is over the bytes and
  the parsed fields of the assembled input. A plant that moves is still caught,
  because nothing in the assertion mentions where the plant was.

### The fixture corpus

Fixtures live under `evals/testdata/cold-reading/`, not `evals/data/`: the
latter is reserved for the auto-discovered per-command shape of itd-75, and
`testdata` is the tree the Go toolchain already excludes from builds. Three
variants, all inert on disk and materialised into a temporary directory per run:

- `baseline/`: a whole miniature repository state, every plant in its canonical
  home. The eval copies it out, runs `git init`, and makes one commit under a
  fixture identity in a reserved example domain, so the manifest's commit
  reference resolves. `HOME` is redirected to a temporary home carrying a
  planted transcript store at `.abcd/history/<root-sha>/`, keyed on the
  fixture's own root-commit SHA, so the transcript class of invariant 15 is
  genuinely reachable if the assembler ever walked it.
- `holed/`: the negative control. Two plants are relocated into positively
  included material, one into a `01-product/` brief chapter and one into
  `docs/`, which the shipped tree includes wholesale. A correct assembler passes
  both through, and the eval must report exactly those two violations.
- `order/`: shared with spc-65, which owns its contents.

Plants, one per warm location class, each carrying a distinct token of the form
`ABCD-EVAL-SENTINEL-<CLASS>` so a failure names the class that leaked:

| Warm location class | Where the baseline plants it |
|---|---|
| Framing traces, declined construals, and the local ledger side | `.abcd/.work.local/ledger/` |
| Transcript class | `.abcd/.work.local/scratch/`, and the fixture home's `.abcd/history/<root-sha>/transcripts/` |
| Decisions, the issue ledger in all three states, a wontfix reason, and a lapse entry | `.abcd/work/DECISIONS.md`, an ADR, and `.abcd/work/issues/{open,resolved,wontfix}/` |
| Deliberation chapters and superseded records | `brief/03-evidence/`, `intents/superseded/`, `plans/`, `research/notes/`, and `roadmap/rfcs/` |
| Warm fields on an *included* record type | An Audit Notes heading and a scope-condition disposition on a shipped intent |
| Warm frontmatter keys on *included* record types | `origin` on a spec and a discipline, and the production-mode key on a brief chapter |
| The instrument's own exhaust | A prior manifest under `.abcd/development/readings/<run-id>/`, a prior reading record under `.abcd/work/issues/readings/<run-id>/`, and its disposition |
| Grounds records | Admission and selection grounds on the selection surfaces |
| Draft-intent body, position-dependent | `intents/drafts/`, carrying a second token in its `origin` key |

### Assertions

Per position, over the assembled input the assembler wrote:

1. **Sentinel absence:** No planted token appears in the raw serialisation.
2. **Field absence:** A recursive walk of the parsed document reports no key on
   the excluded-key list at any depth, and no excluded heading in any projected
   body. Depth-agnostic key matching is what makes a warm field landing in a new
   place a failure, which a path assertion would miss.
3. **Family absence:** No item's recorded path lies inside an excluded record
   family, checked against the eval's own list of families.

The draft plant is the one position-sensitive case: its body token is cold at
entailment and warm at widening, so assertion 1 exempts it at entailment alone,
while its `origin` token stays banned everywhere. That is itd-183's drafts
asymmetry, tested rather than trusted.

### Decisions this spec makes that the record did not

- **The eval binds to the verb, not the package**, for the independence reason
  above. The verb's name lives in one constant at the top of the eval file, the
  single place a rename touches.
- **Four demands on spc-61's interface**, stated here so spc-61 can meet them:
  a dry-run invocation that writes the assembled input and the manifest as two
  separate artefacts into an operator-named output directory; a JSON
  serialisation of the assembled input, per brief invariant 4; selection of the
  position and the target state as flags, with no free text, per itd-183's ruled
  interface; and a non-zero exit on failure.
- **Positions exercised:** widening, entailment, and detection are asserted in
  full. At the comparative position the eval asserts only the read-block over
  the instrument's own stored prior outputs; the in-cycle candidate set arrives
  by whatever channel spc-61 defines, and the eval makes no claim about it.
- **The corpus is owned here** and shared with spc-65, so the two evals cannot
  drift apart in the repository state they exercise.

### Wiring

The eval is Go test files in package `evals` with `//go:build smoke`, so
`make smoke` (`go test -tags smoke ./evals/...`) runs it and CI's `smoke` job,
a required status check on the protected branch, runs `make smoke`. No Makefile
or workflow edit is required. One caveat is recorded rather than worked around:
the diff classifier stands the `smoke` job down on a pull request confined to
`docs/`, `.abcd/development/`, `.abcd/work/`, and the root prose files, so a
record-only pull request does not exercise this eval. The merge-queue entry runs
the full set on the would-be merge result, so the property still gates the merge.

## Acceptance criteria mapping

| itd-186 criterion | How spc-64 satisfies it | Pinned by |
|---|---|---|
| Given the fixture state, when the eval runs, then it passes only if the output contains no planted warm content and no field on the exclusion list | Assertions 1 to 3 run over the baseline fixture at every asserted position | `TestReadBlockBaselineIsClean` |
| Given a ledger path moved to a new location holding a plant, when the eval runs, then it fails loudly | The content oracle names no path, so a relocated plant is caught wherever it lands; the `holed/` variant relocates two plants into included material | `TestReadBlockCatchesAHoledFirewall` |
| Given a warm field introduced on a record type already on the include list, when the eval runs, then it fails | The `origin` and production-mode plants sit on included record types, and the Audit Notes plant sits inside an included file; each is a real leak path if the key filter or the projection is absent | `TestReadBlockCatchesWarmFieldsOnIncludedTypes` |
| Given a repository state containing manifests and reading records from prior runs, when the eval runs, then none of them appears in the output | The exhaust plants, asserted at every position including comparative | `TestPriorRunExhaustNeverReaches` |

## Tests

Every test below is watched red before the change and green after.

- `TestReadBlockBaselineIsClean` is watched red against an assembler whose
  exclusion of the local ledger tier is temporarily removed by a one-line local
  patch. The sentinel that must be caught is
  `ABCD-EVAL-SENTINEL-LEDGER-FRAMING`; the run is recorded in the pull request
  body, and the patch is reverted before the branch is pushed.
- `TestReadBlockCatchesAHoledFirewall` is red until the `holed/` variant exists,
  and it is the permanent proof that the assertion can fail: it demands exactly
  two violations, naming the classes, so an assertion that stops detecting
  anything fails here rather than passing quietly.
- `TestEverySentinelIsPlanted` walks the materialised fixture and fails if any
  class's token is absent or planted more than its declared number of times. It
  is the anti-vacuity guard: without it, a corpus that lost a plant would pass
  silently, the one failure mode an absence assertion cannot see.
- `TestOracleImportsNothingFromTheAssembler` is red the moment the eval's own
  package imports the assembler, which keeps the independence claim true after
  this spec closes.
- `TestPriorRunExhaustNeverReaches` and
  `TestReadBlockCatchesWarmFieldsOnIncludedTypes` are watched red against the
  same temporary hole technique, each removing the single exclusion it covers.

## Grounds (pursued)

_Pre-tooling: recorded in the plan record until the grounds argument (itd-179) ships._

Every other component in the workstream asserts the blindfold; this one is the
only component capable of falsifying it, and an eval that read the assembler's
own include table could only ever confirm the table rather than test the
property. It is pursued now because the assembler exists to hold a firewall
whose failure is otherwise silent: warm content that reaches a reading
contaminates the reading invisibly, and nothing else in the cycle would notice.

## Out of scope

- The assembler itself, its include table, its projection, and its manifest
  (spc-61); the four reading definitions and the blindness core
  ([spc-62](spc-62-four-cold-reading-definitions-one-blindness-core-each-positi.md));
  validating what a reading was licensed to produce
  ([spc-63](spc-63-one-ingest-verb-validates-every-cold-reading-output-includin.md));
  and determinism of the assembled input
  ([spc-65](spc-65-amnesia-is-a-repository-property-proven-by-an-eval-the-same.md)).
- Prose-borne warmth inside an included chapter, which carries no structural
  signal. The eval catches a planted token wherever it sits, so it is stronger
  than the assembler's filter here, but unplanted prose warmth stays the
  disclosed residue itd-183 records, and so does the timing and
  target-selection leakage that intent discloses.
