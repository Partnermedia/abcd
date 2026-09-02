---
id: spc-2609020626039834
slug: a-comparative-reading-receives-the-widening-run-s-items-as-i
intent: itd-2609020625407419
origin: researcher-authored
production_mode: dictated-and-formatted
---
# The comparative channel: one widening run's candidates, two fields each, characterised before anyone admits one

## Summary

spc-2609020626039834 delivers
[itd-2609020625407419](../../intents/planned/itd-2609020625407419-a-comparative-reading-receives-the-widening-run-s-items-as-i.md).
`abcd reading assemble --position comparative` gains a fourth operand naming one
widening run, and the include table gains one row, admitted at the comparative
position alone, that reaches that run's reading records and projects each to
the two body fields the widening position declares, keyed by the item
identifier the comparative body must cite. At the comparative position the
include table is the whole account of what the reading sees: The candidates
pass through the table and the walk as every other row does, every other row
withdraws from the position except the criteria discipline, and the rest of the
readings store is excluded by rows derived from the ledger's own directory
constants. Nothing else from the store travels: No disposition, no admission,
no surprise, no reframe, no other run, no manifest. If any candidate already
carries a disposition or an admission the assembly refuses and names it,
because the candidate set is defined as pre-admission. If the run holds fewer
than two candidates the assembly refuses, names the fixed interpretation, and
stages a comparative run with an empty candidate set whose manifest names the
widening run; ingesting that run commits it, so the outcome of a widening run
is always the same thing: A committed comparative run whose manifest names it.
At ingest, a comparative item naming a candidate outside the recorded run or a
criterion the discipline does not declare is refused by name. The comparative
preset refusal of
[spc-69](../closed/spc-69-a-reading-is-about-something-narrower-than-everything-its.md)
is withdrawn, and the read-block eval gains the comparative rows and plants.

This spec builds under the two flagged readings and does not ship until the ADR
is adopted. The positional exception to the prior-run exhaust is a
trust-boundary rule and belongs in an ADR with a brief invariant amendment;
[iss-2609020626100041](../../../work/issues/open/iss-2609020626100041-adr-owed-the-comparative-channel-admits-two-fields-of-one-wi.md)
carries that decision and the withdrawal of itd-199's ac-10 with it. The same
ADR carries the fourth operand: It supersedes adr-58's "to that extent and to
no other" to the extent of one more closed operand, `--candidate-run <rdg-N>`,
and amends brief invariant 15's operand enumeration to name it; the binding
property, that no operand carries prose, is unchanged.

## Scope

In: The candidate-run operand and its refusals; the candidate row of the
include table, its two-field projection and its selector; the withdrawal of
every other row from the comparative position and the criteria narrowing of
the disciplines row; the ordering guard; the fewer-than-two branch and the
empty comparative run it stages; the candidate kind, the bundle and manifest
additions, and the version classes that move with them; the derived exclusion
rows the comparative position asserts; the committed preset entry and the
withdrawal of the loader's comparative refusal; the criteria discipline as a
required part of the comparative bundle; the two comparative ingest checks and
the empty-item carve-out; the comparative-run probe the admission gate reads;
the exported durable-tier writer; the definition's Object section; the eval
rows, plants and counts; the plugin page and the brief chapter.

Out: Admission itself and the verb that writes it, which are
[spc-2609020626040342](spc-2609020626040342-an-admission-and-a-surprise-are-written-by-a-verb-and-the-or.md)'s,
including the gate in the shared disposition writer that holds the ruled
ordering; any characterisation performed by the assembler; a candidate set
drawn from more than one run; the preset file's own version, which
spc-2609020626048722 owns; the ADR and the invariant amendment, which are the
maintainer's.

## Approach

### Landing order

The eight Iteration 2 specs land in this order: PRE (spc-2609020626048722),
CND (spc-2609020626046252), ORG (spc-2609020626042168), CMP
(spc-2609020626039834), ADM (spc-2609020626040342), RFM (spc-2609020626048705),
SCR (spc-2609020626045177), PRN (spc-2609020626042471). CND lands strictly
before PRN; CMP before ADM before SCR; RFM after ADM and after CMP. No spec
names a target version number: Each names the class of bump it makes, and the
merging change sets each constant from the merged base and updates every
pinned count in the same diff.

This spec's classes: `SchemaVersion` on the bundle and the manifest moves by
one, because the closed `Kind` vocabulary gains `candidate` and both shapes
gain fields (`DecodeManifest` refuses an unknown kind, so an old binary refuses
the new manifest); `AssemblerVersionCore` moves MINOR, because `Table`,
`Exclusions`, `Kinds()` and the bundle shape all move. The comparative preset
entry conforms to the preset file's committed version at landing, so it
carries the `window` spc-2609020626048722's version 2 requires.

### The operand

`AssembleRequest` in `internal/core/reading/assemble.go` gains
`CandidateRun string`, and the front door in `internal/surface/cli/reading.go`
gains `--candidate-run <rdg-N>`. The value is validated by
`recordid.ValidReadingRunID` before it is joined into any path, exactly as the
ingest verb validates a payload's run id. It is a closed form, never prose, so
adr-58's binding property holds with a fourth operand as it held with a third;
the ADR named above is what makes the fourth operand a ruled one rather than
an accretion.

The comparative position refuses without it, naming the operand and its shape.
Every other position refuses with it: A candidate set is the comparative
reading's object and nobody else's, so the operand elsewhere is a misdirected
invocation rather than a harmless extra. `readingOperands` in
`internal/surface/cli/regime_surface_test.go` is the pin that fails closed on
the addition, and this spec is the statement the pin asks for. The block in
`Assemble` that refuses the comparative position before anything is resolved is
removed, and `AssemblingPositions()` returns all four positions.

### The candidate row

`Table` in `internal/core/reading/include.go` gains one row: Positions
`PositionComparative` alone, Source `.abcd/work/issues/readings`, Store `rdi`,
Fields `configuration` and `what_admits_it`, Kind `KindCandidate`, and a Rule
citing the ADR iss-2609020626100041 carries. The row's selector is the
`--candidate-run` operand: The assembly narrows the row's enumeration to the
named run's directory before the scope filter runs, the way a record-id scope
narrows a row to one record, so the row reaches one run directory and never
the family. That directory is the leaf bucket the readings family keys on,
which is what assembler rule 1 permits an include row to name (ruling (18)),
and the row is refused at every other position by `AdmittedAt` as every row
is. Each item carries the item id as its key, and the two fields resolve
through `projectField` as frontmatter keys, which is what they are on a
reading record. The pattern, the manifest reference, the regime and every
other field of the record have no row and therefore no item: The projection is
the include table's, not a caller's memory.

Because the row is a table row, the rest follows without a second mechanism.
`Admits` answers true for the named run's records at the comparative position
and false everywhere else; `Render()` carries the row, so the charter names
the channel and `AssemblerVersion` digests it; the comparative definition's
source list is regenerated from the table, which `TestDefinitionHoldsItsFiveParts`
in `internal/core/reading/definitions_test.go` holds through `statedSources`
against `admittedSources`. `Kinds()` gains `KindCandidate` (`"candidate"`),
and the kind is refused as a `--scope` token by `ResolveScope` and as a preset
kind by `validatePresets`, naming it: A scope selects repository material, and
a candidate is selected by the operand, never by a scope.

The run must have committed: `Assemble` probes
`.abcd/development/readings/<rdg-N>/run.json` and refuses a run without its
commit marker, because a run without one never happened and its records are
the next sweep's to roll back. Every record the row enumerates is validated
through `validateReadingStrict`, and a record whose `position` is not
`widening` refuses the assembly by name: A run at any other position is not a
candidate set.

`internal/core/reading/candidates.go` turns each enumerated record into two
`candidate` items, one per projected field, ordered by id then by field so two
assemblies of one run produce a byte-identical bundle.

### Every other row withdraws

At the comparative position the include table is the whole account, and the
readings companion admits no source but the candidates and the criteria. Every
row whose Positions is `allPositions` today loses the comparative position:
The three brief rows, the shipped-intents row, the specs row and the four
shipped-tree rows. The disciplines row keeps it, and at the comparative
position the assembler narrows that row's enumeration to `CriteriaDiscipline`
(`"itd-191"`) before the scope filter, so no scope token can widen the
comparative material past the criteria. The consequence is deliberate: `--scope
source` or `--scope warm` at the comparative position selects nothing outside
the criteria discipline, and the scope's meaning at this position is exactly
what the preset entry below names. `TestThreePositionsCarryDistinctItemSets`
extends to four positions on that basis.

### The ordering guard

`capture` gains `ItemFate(repoRoot, run, item string) (ItemFate, error)`, the
one admitted-proposal probe in this binary: It returns the item's standing
disposition ids (through `standingDispositions` and
`issueschema.StandingDispositionIDs`) and the ids of every admission under
`admissions/<run>/` whose `proposal` names the item, keyed on the (run,
proposal) pair exactly as `admittedProposals` in
`internal/core/lint/readingoutstanding.go` keys it. The judgement, given a
disposition set and an admission set, lives in `issueschema` beside the
standing-disposition judgement; the walk lives in `capture` once, and
spc-2609020626040342's `Admit` calls the same function rather than probing the
store again. `Assemble` asks it for every candidate before anything is minted
or written. Any standing disposition, in any state, refuses and names the item
and the disposition; any admission refuses and names the item and the
admission; a contested or cyclic disposition set refuses and names the item,
because a candidate whose fate cannot be read is not demonstrably uncommitted.
The refusal states the rule: The candidate set is defined as pre-admission, and
characterisation precedes admission.

The ruled ordering itself is not held here. spc-2609020626040342 puts one gate
in the shared disposition writer, `writeDispositionLocked`, so that at the
widening position no disposition in any state can be written until a committed
comparative run names the item's run. This guard is therefore unreachable
through any verb and stays as a guard against hand-written records: The
assembler cannot know a record was placed by hand, so it refuses what it finds.
There is no deadlock between the two, because the writer's gate opens on the
comparative run this assembly produces, and the closing run at position 4 is
repeatable for the same reason: Dispositions cannot exist before a comparative
run, and a second comparative run over the same widening run before any
disposition is a second run. A comparative reading commissioned after
dispositions is not a repeat of this position; recurrence is warm work.

### The fixed interpretation and the empty comparative run

Fewer than two candidates refuses the assembly, and the refusal names the
interpretation: The comparative reading has nothing to compare and is not
exercised. Unless the invocation is a dry run, the assembly still stages a run
directory, on the same rules as any other, whose bundle carries no candidate
item and whose manifest carries `candidate_run: <rdg-N>` and `candidates: 0`,
and the result reports `NotExercised` with the count the run holds. That run
is committed by `abcd reading ingest` with an empty item list, which is the
clean-run idiom ruling (3) already fixed: A run recorded with an empty item
set. `checkEnvelope` in `internal/core/reading/ingest.go` refuses an empty item
set today; the refusal gains one carve-out, taken after the parked manifest is
resolved, for a comparative manifest whose `candidates` is zero, and no other
run may be empty. The outcome of a widening run is therefore one thing whether
the position ran or was not exercised: A committed comparative run whose
manifest names that widening run. The durable tier stays write-once at
emission; no other run's directory is ever amended, and a second assembly over
the same one-item run is a second run.

`ReadingsRecordDir` moves to `issueschema` in the same change and
`reading.ReadingsRecordDir` becomes an alias of it, so the two packages name
one directory. `issueschema` declares `RunHead{RunID, Position, CandidateRun}`,
the strictly decoded subset of the run record that answers "which widening run
did this comparative run characterise". `capture` gains
`ComparativeRunFor(repoRoot, run string) (string, error)`, which walks
`.abcd/development/readings/*/run.json` behind the symlink-refusing walk that
package already carries, reads each behind `fsutil.ReadGuarded`, and returns
the lowest comparative run id whose `candidate_run` names the run, or nothing.
That is the one probe the disposition writer's gate reads, and `capture` can
call it because it does not import `reading`.

### The exported durable-tier writer

`reading` exports `WriteRunArtefact(repoRoot, runID, name string, v any)
(string, error)`, the one way a package outside `reading` writes beside a
committed run. It opens the containment root as `Ingest` does, requires the
run's `run.json` to exist, refuses a name that is one of the run's own three
files or that is not a plain `.json` basename, refuses an existing file
because the durable tier is write-once at emission, and writes through
`writeJSONIn`. spc-2609020626045177's `scribe ingest` promotes its manifest
through it; nothing in this spec calls it, and it exists here because the root
and the write are this package's.

### The bundle, the kind, the scope, and the criteria

`BundleItem` in `internal/core/reading/manifest.go` gains `Candidate string`
and `Field string`, both `omitempty`, both set only for a candidate item: The
reading is told which `rdi-N` each text belongs to and whether it is the
configuration or what admits it, and nothing else. Neither is a repository
path, so brief invariant 15 holds, and `TestNoBundleFieldIsAScopeSelector` in
`internal/core/reading/scope_test.go` gains the two members. At every other
position both are empty, and a test pins that. The closed bundle-item allow-set
in `TestBundleGainsNoFieldFromTheReport` in `internal/core/reading/size_test.go`
gains `candidate` and `field`; the bundle's own allow-set there does not move
for this spec.

`validatePresets` in `internal/core/reading/scope.go` stops refusing a
comparative entry, and `.abcd/config/reading-presets.json` gains one under
`cold` at `comparative` naming `records: ["itd-191"]` and no kind and no path,
in the shape the committed preset version requires at landing, so it carries a
`window` measured in a clean clone. `warm` adds nothing there, and the union
keeps warm containing cold. spc-2609020626048722's window eval iterates the
three assembling positions by name and states that the comparative entry is
exempt, because its object is bounded by the widening run rather than by the
tree; this spec's own eval covers the comparative position. The scope stays
required and keeps its meaning: It selects the repository material passed
beside the candidates, which at this position is the discipline and nothing
else.

The criteria are never supplied at invocation and a comparative bundle without
them characterises against nothing, so `reading` declares
`CriteriaDiscipline = "itd-191"` and `Assemble` refuses a comparative assembly
whose scoped items carry no item whose path names that record. The declared
criteria are read at assembly by `declaredCriteria` in
`internal/core/reading/criteria.go`: The bullets of the discipline's
`## The rule` section, each name being the text before the first em dash. A
test pins the committed
[itd-191](../../intents/disciplines/itd-191-the-selection-criteria-are-a-declared-recorded-discipline-a.md)
to its six names, so an amendment to the slate is a recorded change here too.

The comparative definition's Object section in `agents/cold-reading-comparative.md`
names both the candidate run and the discipline and loses its two
does-not-assemble paragraphs; the precedence sentence stays, and its source
list is regenerated from the table.
`TestComparativeObjectIsTheWideningPreAdmissionOutput` in
`internal/core/reading/definitions_test.go` moves with it.

### The manifest and the exclusions it asserts

`Manifest` gains `CandidateRun`, `Candidates` (the count of candidates the
bundle carries), `CandidateFields` (the two projected field names) and
`Criteria` (the parsed names), all `omitempty`, and `ManifestItem` gains
`Candidate` beside the `Field` it already carries. `RunRecord` in
`internal/core/reading/ingest.go` carries `CandidateRun` and `Candidates`
forward, so a committed comparative run says which widening run it
characterised and whether it was exercised.

The exclusion floor's row for `.abcd/work/issues` narrows to the three other
positions. At the comparative position the directory rows are derived, never
listed: `issueschema` gains `LedgerDirs()`, returning every ledger directory
constant it declares (`StatusDirs`, `ReadingsDir`, `DispositionsDir`,
`AdmissionsDir`, `SurprisesDir`, and spc-2609020626048705's `ReframesDir` once
it lands), and `Exclusions` gains one comparative-only directory row per entry
except `ReadingsDir`, whose row carries the signal `readings store` and the
detail "every run other than the candidate run, and every field of the named
run's items other than configuration and what_admits_it". A family the ledger
adds later is excluded here the day its constant is declared, and the scribe's
allow list in spc-2609020626045177 derives from the same function, so the two
instruments describe one set. `assertExclusions` enforces the directory rows
as it does today; `assertCandidateProjection` is the fail-closed half of the
signal row, refusing to emit any candidate item whose path is not under the
named run's directory or whose field is not one of the two. An assertion the
manifest declares and the assembler cannot violate is what adr-56 asks of an
exclusion control. `.abcd/development/readings` stays excluded at every
position: A run's own manifest and `run.json` never travel.

### Ingest at the comparative position

`validateItems` in `internal/core/reading/ingest_regime.go` takes the
manifest: Its signature becomes `validateItems(out Output, m Manifest, def
Definition)`, because the two checks below are against the manifest and the
function is handed none today. Both run after `checkItem`, at the comparative
position only. `candidate_id` must equal the `Candidate` of some manifest
item, refused as `unknown-candidate` naming the item's ordinal and the id.
`criterion` must equal, after `foldForMatching`, one of `manifest.Criteria`,
refused as `undeclared-criterion` naming the ordinal and the criterion. The
reserved names and the ordering signatures of the evaluative regime are
untouched.

On the commit path nothing outside the run's own directory is written: The run
record carries `candidate_run`, and the admission gate finds it through
`ComparativeRunFor`. No file in the widening run's directory is touched.

### The read-block eval

The fixture under `evals/testdata/cold-reading/baseline/` gains a committed
widening run of three items, a second widening run, dispositions, admissions
and a surprise on the second run's items, and a variant carrying a disposition
on one candidate of the named run. Three sentinel classes are added:
`CANDIDATE`, the carrier, planted in the `configuration` of the named run's
items and required to arrive at the comparative position; `ENVELOPE`, planted
in every candidate's `pattern` and required never to arrive; and `FATE`,
planted in the second run's dispositions and surprise (the leak check) and in
the variant's disposition on a candidate (the refusal check). `EXHAUST` gains
the second run's items as homes. The oracle stays transcribed: No file under
`evals/` imports the assembler, so `excludedFamilies` is edited by hand to
mirror the derived rows, and spc-2609020626048705 adds its row when it lands.

The pinned counts move as follows. `sentinelClasses` 18 to 21.
`excludedKeys` stays at 2: The envelope's keys are excluded by the candidate
row's projection and caught by plant, not by the key table. `excludedFamilies`
15 to 21: The `.abcd/work/issues` row gains the three other positions, and six
comparative-only rows are added for `open`, `resolved`, `wontfix`,
`dispositions`, `admissions` and `surprises`. `admittedRecordPaths` 8 to 9:
The fixture's candidate run directory at the comparative position. `coverage`
56 to 65, with these rows:

- The candidate run's items are admitted at the comparative position,
  projected to configuration and what admits it. Falsifier: Delete the
  candidate row from `Table`. Caught by the carrier floor; class `CANDIDATE`.
- The candidate row is admitted at no other position. Falsifier: Add the three
  other positions to the row. Caught by family absence; class `CANDIDATE`.
- The candidate projection drops the envelope. Falsifier: Empty the row's
  `Fields`. Caught by sentinel absence; class `ENVELOPE`.
- No run other than the candidate run reaches the comparative reading.
  Falsifier: Drop the run narrowing from the candidate selector. Caught by
  sentinel absence; class `EXHAUST`.
- Dispositions never reach the comparative reading. Falsifier: Delete the
  derived dispositions row and add an include row for it. Caught by sentinel
  absence; class `FATE`.
- Admissions never reach the comparative reading. Falsifier: Delete the
  derived admissions row and add an include row for it. Caught by sentinel
  absence; class `GROUNDS`.
- Surprises never reach the comparative reading. Falsifier: Delete the derived
  surprises row and add an include row for it. Caught by sentinel absence;
  class `FATE`.
- The status directories never reach the comparative reading. Falsifier:
  Delete the three derived status rows and add an include row under
  `.abcd/work/issues`. Caught by sentinel absence; class `DECISION`.
- A candidate carrying a standing disposition or an admission refuses the
  assembly. Falsifier: Delete the `ItemFate` call. Caught by refusal; class
  `FATE`.

`declaredGaps` stays at 7: Every row above is caught. `evals/coldreading_test.go`
gains the comparative cases named below; `TestComparativeRefusesToAssemble` is
retired with the refusal it guarded, and `TestEverySentinelIsPlanted` counts
the three new classes.

## How the Acceptance Criteria are satisfied

- **ac-1**: `Assemble` refuses a comparative request whose `CandidateRun` is
  empty, naming `--candidate-run` and the `rdg-N` shape. Test:
  `TestComparativeRequiresACandidateRun`.
- **ac-2**: The candidate row enumerates the named run and projects two fields;
  `candidates.go` emits two items per candidate; the bundle carries
  `Candidate` and `Field` and the discipline item, and no other row is
  admitted. Tests: `TestComparativeBundleCarriesTwoFieldsPerCandidate`,
  `TestComparativeBundleCarriesTheCriteriaDiscipline`,
  `TestComparativeAdmitsNoOtherRow`.
- **ac-3**: `ItemFate` reports a standing disposition; `Assemble` refuses
  naming the item. Test: `TestComparativeRefusesADispositionedCandidate`.
- **ac-4**: `ItemFate` reports an admission keyed on (run, proposal);
  `Assemble` refuses naming the item. Test:
  `TestComparativeRefusesAnAdmittedCandidate`.
- **ac-5**: The fewer-than-two branch refuses, names the interpretation, and
  stages a comparative run with an empty candidate set whose manifest names
  the widening run; `reading ingest` commits it with an empty item list.
  Tests: `TestOneCandidateStagesAnEmptyComparativeRun`,
  `TestIngestCommitsAnEmptyComparativeRun`.
- **ac-6**: The candidate row reaches one run directory and two fields; the
  derived exclusion rows assert the rest; `assertCandidateProjection` refuses
  any breach. Tests: `TestComparativeManifestAssertsTheDerivedExclusions`,
  `TestCandidateProjectionRefusesAForeignRun`.
- **ac-7**: The `unknown-candidate` check against the manifest's candidate set.
  Test: `TestIngestRefusesACandidateOutsideTheRun`.
- **ac-8**: The `undeclared-criterion` check against `manifest.Criteria`. Test:
  `TestIngestRefusesAnUndeclaredCriterion`.
- **ac-9**: The eval plants the `FATE` sentinel on a candidate's disposition
  and the `ENVELOPE` sentinel on every candidate's pattern; the assembly must
  refuse, and if it ever assembles the sentinel-absence oracle reports the leak
  by class and position. Test: `TestComparativeChannelCatchesAPlantedFate`.

## Tests

Watched fail before, pass after; each refusal proved by a mutation that removes
it.

- `internal/core/reading/candidates_test.go`:
  `TestComparativeRequiresACandidateRun`,
  `TestCandidateRunIsRefusedAtEveryOtherPosition`,
  `TestComparativeBundleCarriesTwoFieldsPerCandidate`,
  `TestComparativeBundleCarriesTheCriteriaDiscipline`,
  `TestComparativeAdmitsNoOtherRow`,
  `TestComparativeRefusesWithoutTheCriteriaDiscipline`,
  `TestComparativeRefusesADispositionedCandidate`,
  `TestComparativeRefusesAnAdmittedCandidate`,
  `TestComparativeRefusesAnUncommittedRun`,
  `TestComparativeRefusesANonWideningRun`,
  `TestOneCandidateStagesAnEmptyComparativeRun`,
  `TestAnEmptyComparativeRunIsNotStagedOnADryRun`,
  `TestComparativeManifestAssertsTheDerivedExclusions`,
  `TestCandidateProjectionRefusesAForeignRun`,
  `TestCandidateFieldsAreEmptyAtEveryOtherPosition`,
  `TestTwoComparativeAssembliesAreByteIdentical`.
- `internal/core/reading/include_test.go`: `TestCandidateRowIsComparativeOnly`,
  `TestEveryOtherRowWithdrawsFromComparative`,
  `TestComparativeExclusionRowsAreDerivedFromLedgerDirs`,
  `TestRenderCarriesTheCandidateRow`.
- `internal/core/reading/criteria_test.go`:
  `TestDeclaredCriteriaReadsTheCommittedDiscipline` (six names),
  `TestDeclaredCriteriaRefusesAnEmptySlate`.
- `internal/core/reading/scope_test.go`: `TestComparativeRefuses` is replaced by
  `TestComparativePresetIsAdmitted`; `TestWarmContainsCold` and
  `TestThreePositionsCarryDistinctItemSets` extend to four positions;
  `TestNoBundleFieldIsAScopeSelector` covers `Candidate` and `Field`;
  `TestCandidateIsNotAScopeToken` and `TestCandidateIsNotAPresetKind`.
- `internal/core/reading/size_test.go`: `TestBundleGainsNoFieldFromTheReport`
  with the item allow-set extended.
- `internal/core/reading/ingest_regime_test.go`:
  `TestIngestRefusesACandidateOutsideTheRun`,
  `TestIngestRefusesAnUndeclaredCriterion`,
  `TestIngestCommitsAnEmptyComparativeRun`,
  `TestIngestStillRefusesAnEmptyRunAtEveryOtherPosition`.
- `internal/core/reading/ingest_test.go`: `TestWriteRunArtefactIsWriteOnce`,
  `TestWriteRunArtefactRefusesAnUncommittedRun`,
  `TestWriteRunArtefactRefusesTheRunsOwnFiles`.
- `internal/core/capture/reading_test.go`: `TestItemFateReadsBothStores`,
  `TestComparativeRunForNamesTheLowestMatch`,
  `TestComparativeRunForIsEmptyBeforeAnyRun`.
- `internal/core/issueschema/disposition_test.go`: `TestItemFateJudgement`;
  `internal/core/issueschema/statusdirs_test.go`: `TestLedgerDirsNamesEveryConstant`.
- `internal/surface/cli/regime_surface_test.go`: `readingOperands` gains
  `candidate-run`; `TestCandidateRunIsAClosedForm` refuses a value that is not
  an `rdg-N`.
- `evals/coldreading_test.go`:
  `TestComparativeChannelCarriesCandidatesAndNothingElse` (dispositions,
  admissions, a surprise and a second run planted; the bundle carries the named
  run's two candidate fields and the discipline, the carrier arrives, no other
  sentinel does, and the manifest asserts the derived exclusions),
  `TestComparativeChannelCatchesAPlantedFate` (ac-9).
  `TestComparativeRefusesToAssemble` is retired with the refusal it guarded;
  `TestEverySentinelIsPlanted` and `TestEveryAssemblerRuleHasAFalsifier`
  carry the counts above.

## Out of scope

- The verb that admits a candidate, and the gate in the shared disposition
  writer that holds the ruled ordering, which spc-2609020626040342 builds.
- A candidate set from more than one run, or from a run at any other position.
- The pattern named by a widening item. It stays behind on the ground the
  intent's scope condition states; widening the projection is a manifest change
  and a recorded one.
- The preset file's version and the window eval, which spc-2609020626048722
  owns; this spec adds one entry in the committed shape.
- The ADR and the amendment to brief invariant 15 that
  iss-2609020626100041 carries, including the fourth operand's supersession of
  adr-58. This spec is buildable and not shippable until they are adopted.
- The re-disposition of itd-199's surviving condition cond-2608312031028702,
  which this spec's delivery moves. spc-2609020626046252's verb records it with
  the shipped intent as the occasion.
