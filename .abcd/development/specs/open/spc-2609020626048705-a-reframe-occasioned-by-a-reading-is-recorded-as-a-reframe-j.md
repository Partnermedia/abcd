---
id: spc-2609020626048705
slug: a-reframe-occasioned-by-a-reading-is-recorded-as-a-reframe-j
intent: itd-2609020625402518
origin: researcher-authored
production_mode: dictated-and-formatted
---
# The reframe record: `capture reframe` writes one `rfm-N` keyed on the construal's committed hash before and after, joined to the reading item, disposition or surprise that occasioned it

## Summary

spc-2609020626048705 delivers
[itd-2609020625402518](../../intents/planned/itd-2609020625402518-a-reframe-occasioned-by-a-reading-is-recorded-as-a-reframe-j.md).
A new record family, `rfm-N`, lives flat under `.abcd/work/issues/reframes/`
beside the surprise family, one record per reframe occasioned by a reading. It
carries the occasion, the SHA-256 of the construal section as it stood before
and after the rewrite, and the grounds, and nothing of the abandoned framing's
text, which stays local under
[adr-55](../../decisions/adrs/0055-the-construal-stands-in-the-record-its-history-does-not.md).
`abcd capture reframe` is its only writer. It reads the section itself, at
`HEAD` and in the working tree, so the operator supplies no hash; when the
record precedes the commit it writes the first half and says so, and a second
invocation completes it after the commit. The occasion is asserted by the
operator and checked in one respect, that it predates the rewrite. The family
is warm, reaches no reading, and its exclusion is asserted in every manifest.
`abcd rfm-N` reports the occasion and both hashes. The frame this record keys
on is the `## Construal` section of the framing chapter, which is narrower
than adr-55's construal; that narrowing is flagged for the maintainer below.

## Scope

In: The family's constants and schema in `internal/core/issueschema`, the
directory constant registered with the ledger's directory list; the writer,
the section reader and the hash in `internal/core/capture/reframe.go`; the
`capture reframe` sub-verb and its plugin page section; the record store in
`.abcd/record-lint.json` and `internal/core/lint/schema.go`; an exclusion-floor
row, the regenerated charter and a planted sentinel in the read-block eval; the
dispatcher in `internal/core/record`.

Out: The prior construal's text; reframes no reading occasioned; a lint over
construal rewrites; the fourth audit verdict; a second construal surface; a
rewrite of the glossary or of committed scope, which is not a reframe by this
record's definition.

## Approach

### The family: prefix, store and shape

The prefix is `rfm`, fixed here as the intent left it to the spec; it collides
with no family in `recordid`, `issueschema` or the record-lint stores, and
`grep` finds no `rfm-` in the tree. `issueschema/reframe.go` declares
`ReframeFamily = "rfm"`, `ReframesDir = "reframes"` (flat, like
`SurprisesDir`: The record is keyed by what it carries, not by a directory),
`ReframeRequired = ["schema_version", "id", "occasioned_by", "construal_before", "grounds"]`,
`ReframeKnown` as that set plus `construal_after`, and
`ReframeOccasionFamilies = [ReadingItemFamily, DispositionFamily, SurpriseFamily]`.
`ReframesDir` (registered in `issueschema.LedgerDirs()`, the one list the comparative exclusion rows and the scribe allow list derive from) is registered in the ledger's directory list in `issueschema`,
beside `ReadingsDir`, `DispositionsDir`, `AdmissionsDir` and `SurprisesDir`,
which is the list the comparative channel derives its per-directory exclusion
rows from at the comparative position and the scribe derives its allow list
from; registering the constant is what makes both pick the family up without
a literal edit in either. The record is frontmatter only, rendered through
`buildIssueText(fields, "")` as a disposition is: `schema_version: 1`, `id`,
`occasioned_by`, `construal_before`, `construal_after` (absent while the
record is open, a 64-hex scalar once complete), `grounds` (free text, held to
`grounds.ValidateText` over `grounds.Fold`, the floor every grounds-shaped
field is held to, then redacted through the ledger redactor before it is
written). The disclosure pair is not carried: It belongs to intent, spec and
issue records, and both pages say records of other families carry neither.
The store path is `.abcd/work/issues/reframes/rfm-<N>.md`, minted through
`recordid.Minter.Mint(issueschema.ReframeFamily)` under `withLedgerLock`,
provisioned through `safeMkdirLeaf`, and refused if the path already exists.

### The construal section and its hash

The construal is one section of one chapter, per the intent's first scope
condition: `capture.ConstrualPath = ".abcd/development/brief/01-product/06-framing.md"`
and `capture.ConstrualHeading = "Construal"`. `capture.ConstrualHash(doc string) (string, error)`
strips the frontmatter with `site.StripFrontmatter`, walks `site.Sections`, takes
the H2 titled `Construal` and the lines to the next heading of level two or
lower or to the end of the file, normalises line endings to LF, trims blank
edges, and returns the SHA-256 hex of those bytes. A chapter with no such
section, or with two, is refused by name. The not-yet-real marker the chapter
may open the section with is part of the section: A change to it is a change to
what the construal states about itself, and the hash says so rather than
guessing. Three readers share the one function: The committed content through
`gitutil.RunCapped(repoRoot, issueschema.RecordReadLimit, "show", "HEAD:"+ConstrualPath)`,
the working-tree content through the guarded read the ledger already uses, and
prior states through `git log --format=%H -n 64 -- <path>` followed by
`git show <sha>:<path>` for each. `capture.construalHistory` returns that
bounded list of (commit, hash) pairs from `HEAD` backwards, and both the whole
write and the completion walk it; sixty-four commits touching the chapter is
the bound, and a search that reaches it without finding what it looks for is
refused rather than extended.

### The verb, and the two halves

`abcd capture reframe --occasioned-by <rdi-N|dsp-N|srp-N> --grounds "<why>" [--open]`
and `abcd capture reframe --complete <rfm-N>`, dispatching to
`capture.Reframe(ReframeRequest{RepoRoot, IssuesRoot, OccasionedBy, Grounds, Open bool, Complete string})`
returning `ReframeResult{ID, Path, OccasionedBy, Before, After, Half string, Commits int, Redacted int, Degraded string}`.
`Half` is `whole`, `open` or `completed`, and both renderings print it, so
the verb says which half it wrote; `Commits` is how many commits the walk
crossed between the two hashes.

- **Whole, after the commit** (no flag): The working-tree section must hash
  equal to `HEAD`'s, or the verb refuses ("the construal has uncommitted
  changes; commit the rewrite, or record the first half with --open"). It then
  walks the history to the previous distinct committed state, the first hash
  that differs from `HEAD`'s; none within the bound refuses ("the construal at
  HEAD matches no prior committed state, so there is no reframe to record").
  It writes `construal_before` as that prior hash and `construal_after` as
  `HEAD`'s, in one record. The rewrite commit, for the predate check below, is
  the earliest commit in the walk whose hash equals `HEAD`'s.
- **First half, before the commit** (`--open`): `construal_before` is `HEAD`'s
  hash, `construal_after` is absent, and the render says "first half written;
  commit the rewrite, then `abcd capture reframe --complete rfm-N`". A second
  open record is refused while one is open, so a half can never be completed
  against the wrong rewrite. The rewrite is not yet committed, so the predate
  check requires the occasion to be committed at `HEAD`.
- **Second half** (`--complete rfm-N`): The record must be open, the
  working-tree section must equal `HEAD`'s, and `HEAD`'s hash must differ from
  the record's `construal_before` ("the construal at HEAD is still the state
  the record opened against; nothing was rewritten"). Then the verb walks the
  chapter's history back from `HEAD` until it finds the record's
  `construal_before`, within the bound. Distinct states between the two are
  crossed and counted, not refused: A rewrite committed in two commits, or
  brought in by a merge commit rather than a squash, pairs the same way, so
  the merge strategy does not decide the outcome; `Commits` reports the
  crossing and the render says "completed across N commits". A walk that
  reaches the bound without finding the before-hash refuses naming both
  hashes: The section's history no longer contains the state the record
  opened against, which is the one failure the intent's Mechanism names, and
  it is loud. On success `construal_after` is set through `setScalarField`
  under the ledger lock, in place, atomically.

Whole and first-half writes validate and redact `grounds`, mint under the
lock, refuse an existing path, and validate the rendered record through
`validateReframeStrict` (required set, allow-list, hash shape, occasion shape)
before the write, so the writer refuses what the gate would refuse.

### Resolving the occasion, and the one check on the join

The occasion is resolved through `readingitem.ResolveOccasion(issuesRoot, id, issueschema.ReframeOccasionFamilies)`
in the `internal/core/readingitem` leaf, the one resolver the admission,
reframe and condition verbs share; it refuses an id outside the families it is
handed by shape before any path is built. An `rdi-N` resolves through
`readingitem.Locate`; a `dsp-N` through `readingitem.LocateDisposition`, which
walks `dispositions/<item>/dsp-N.md` across every item directory with
`refuseSymlinkedDir` at both levels, exactly as `readingitem.Paths` walks runs,
and refuses zero or more than one match; an `srp-N` is `surprises/srp-N.md`,
admitted only as a regular file by `Lstat`. Resolution is by presence in the
store, so a surprise written by hand today and by the sibling verb tomorrow
resolves the same way. `capture.findReadingItem` stays as a thin wrapper over
`readingitem.Locate` until every caller has moved; this spec calls the leaf.

The join from occasion to rewrite is operator-asserted: Nothing in the record
can prove that this reading item is why the construal was rewritten, and the
verb does not pretend to. It checks one thing, that the occasion predates the
rewrite. `capture.occasionCommit` takes the resolved path and asks
`git log --diff-filter=A --format=%H -n 1 -- <path>` for the commit that added
the record; an occasion not yet committed refuses ("the occasion rdi-N is not
committed; a reframe cannot be occasioned by a record that does not yet
exist in history"). For the whole write and the completion, that commit must
be a strict ancestor of the rewrite commit (`git merge-base --is-ancestor`),
or the verb refuses naming both commits ("the occasion was committed after
the rewrite; a reframe cannot be occasioned by what came later"). For the
first half the rewrite commit does not yet exist, so the occasion must be
committed at `HEAD`, which the same check delivers with `HEAD` as the second
operand. The check is a floor, not a proof, and the capture page says so.

### Exclusion from every reading, asserted

The exclusion floor's row for `.abcd/work/issues` binds at the widening,
entailment and detection positions once the comparative channel narrows it,
and at the comparative position the floor's per-directory rows are derived
from the ledger's directory list, which now carries `ReframesDir`; so no
include row can reach a reframe at any position. Since the record is a new
family and the floor is a declaration a reader checks, `Exclusions` in
`internal/core/reading/include.go` gains one row:
`{Rule: "absent from the positive walk", Signal: "record type in a denied path", Detail: "the reframe record"}`,
beside the lapse log's. `Render()` moves, so `AssemblerVersion()` moves with it
and the charter is regenerated in the same change, which
`TestReadingsCharterCarriesTheRenderedIncludeTable` requires. The read-block
eval gains a plant: A reframe record carrying the sentinel class
`LEDGER-REFRAME` under `testdata/cold-reading/baseline/.abcd/work/issues/reframes/`,
its row in the oracle table citing this spec, and its row in the coverage
matrix, so `TestEveryAssemblerRuleHasAFalsifier` and the declared table sizes
move together. The pinned counts this spec moves: `sentinelClasses` by one,
`coverage` by one, and `excludedFamilies` by one at the comparative position
by derivation from the directory list; `excludedKeys` and `declaredGaps` do
not move.

### Versions

This spec makes a MINOR change to `AssemblerVersionCore`: An `Exclusions` row
is part of the assembly contract, and the contract's own comment says the
number moves when the contract moves. `SchemaVersion` on the manifest and
the bundle does not move: No artefact field and no kind changes. The merging
change sets the constant from the merged base and updates every pinned count
in the same diff; this spec names the class of bump and no number.

### Dispatch

`record.IDRe` becomes `^(iss|itd|spc|adr|adm|srp|rfm)-[0-9]+$`; the admission
and surprise spec makes that edit and the matching edit to the dispatch-family
list in `commands/abcd.md`, and this spec inherits both, since it lands after
that spec. `Describe` gains `describeReframe`, which reads
`.abcd/work/issues/reframes/rfm-N.md` through `readRecordHead`. `Family` is
`reframe`; `Status` is `open` when `construal_after` is absent and `complete`
otherwise; `Title` is "reframe occasioned by <id>"; `Links` carries
`occasioned_by`, `construal_before` and, when present, `construal_after`.
`NextMoves` on an open record is "commit the rewrite, then
`abcd capture reframe --complete rfm-N`", through a new
`verbCaptureReframe = "capture reframe"` in `RecommendedVerbPaths`, so the
anti-drift test covers it; a complete record reports none.

### The flagged decisions, and what is built under them

Two things are flagged for the maintainer, both carried on
[iss-2609020626118168](../../../work/issues/open/iss-2609020626118168-ruling-owed-whether-every-rewrite-of-the-construal-must-carr.md).

The first is the one the intent flags: Whether every construal rewrite must
carry a reframe record. This spec builds under the second reading that issue
names: Only a reading-occasioned rewrite carries a record, so the verb
requires `--occasioned-by` and no lint judges a construal edit. That is why
the verb derives `construal_before` from the chapter's history rather than
from the last record's `construal_after`: A chain would refuse every
reading-occasioned reframe that followed an unrecorded one, deciding the
ruling by implementation. If the ruling goes the other way, the chain check
and a lint over the chapter are the additions, and the record shape here
already carries what they would read.

The second is a narrowing this spec makes against adr-55. adr-55's construal
is the framing statement, the glossary's committed terms, and committed scope
and vocabulary together, "as it presently stands"; the governing framework
lists the construal and the frame as two landings. This record keys on the
`## Construal` section of the framing chapter alone, because that is the one
surface whose content a hash can name before and after, and a change to the
glossary or to committed scope is therefore not a reframe by this record's
definition. The intent's flagged decision names this narrowing. Widening the
record to adr-55's construal (a hash over the glossary and the scope chapters
beside the section, or one record per surface moved) is a ruling this spec
does not make; the shape here carries one before and one after, and a wider
construal would carry one pair per surface.

### Surfaces

`newCaptureCommand` in `internal/surface/cli/cli.go` gains the sub-verb with
flags `--occasioned-by`, `--grounds`, `--open`, `--complete`; its usage line
joins the capture page's `argument-hint`. `commands/capture.md` gains a
"Record a reframe" section stating the two halves, the three occasion
families, the predate check and its limit, and that the prior construal's text
never enters the record. `.abcd/development/brief/04-surfaces/06-capture.md`
gains the row `` `reframe` | — | shipped `` in its sub-verbs table, the
release surface snapshot is regenerated, and the CLI reference page is
regenerated with `go generate ./internal/surface/cli`.

### Landing order

The eight Iteration 2 specs land in this order: The preset window (PRE), the
condition verb (CND), the reading-occasioned origin (ORG), the comparative
channel (CMP), admission and surprise (ADM), this spec (RFM), the scribe (SCR)
and principles (PRN). CND lands strictly before PRN; CMP before ADM before
SCR; RFM after ADM and after CMP. This spec lands after ADM because it
inherits ADM's `record.IDRe` edit and resolves the surprise family ADM writes,
and after CMP because its directory constant feeds the per-directory exclusion
rows CMP derives at the comparative position; it lands before SCR so the
scribe's derived allow list carries the family from the start.

## How the Acceptance Criteria are satisfied

- **ac-1 (one record with occasion, both hashes and the ground).** The whole
  write after a committed rewrite: `construal_before` from the chapter's
  previous distinct state, `construal_after` from `HEAD`. Proved by
  `TestReframeRecordsACommittedRewrite`, which commits two states of the
  chapter in a fixture repository and compares both hashes to
  `ConstrualHash` over each committed text.
- **ac-2 (no known prior state refuses, naming the mismatch).** The whole
  write refuses when no distinct prior state exists, and the completion refuses
  when the walk reaches its bound without finding the record's
  `construal_before`, naming both hashes. `TestReframeRefusesAConstrualWithNoPriorState`
  and `TestCompleteRefusesARewriteItCannotPair`.
- **ac-3 (an unresolvable occasion refuses).** `readingitem.ResolveOccasion`;
  proved by `TestReframeRefusesAnUnresolvableOccasion` over an unknown
  `rdi-N`, a `dsp-N` in no item directory, an `srp-N` that is a symlink, and
  an `iss-N`.
- **ac-4 (no reframe reaches any reading; the manifest asserts it).** The
  denied segment, the derived per-directory row at comparative, the new
  exclusion row, and the planted sentinel. Proved by
  `TestReframeRecordsNeverReachTheBundle` in `internal/core/reading`, over all
  four positions, and by the read-block eval's baseline and holed runs over
  the new plant.
- **ac-5 (dispatch reports the occasion and both hashes).**
  `describeReframe`; proved by `TestDescribeReframeReportsOccasionAndHashes`
  and the root-dispatch JSON contract test.

## Tests

Watched fail before, pass after. `internal/core/capture`:
`TestConstrualHashIsStableAcrossLineEndingsAndBlankEdges`,
`TestConstrualHashRefusesAChapterWithoutTheSection`,
`TestReframeRecordsACommittedRewrite`, `TestReframeOpensAHalfBeforeTheCommit`,
`TestCompleteFinishesAnOpenRecord`,
`TestCompleteCrossesATwoCommitRewrite` (two commits between the halves, and
a merge commit bringing the rewrite in, both complete and report the
crossing), `TestCompleteRefusesARewriteItCannotPair`,
`TestReframeRefusesAConstrualWithNoPriorState`,
`TestReframeRefusesUncommittedChangesWithoutOpen`,
`TestReframeRefusesASecondOpenRecord`,
`TestReframeHoldsTheGroundToTheFloor` (blank, and a value `ValidateText`
refuses), `TestReframeRedactsTheGround`,
`TestReframeRefusesAnUnresolvableOccasion`,
`TestReframeRefusesAnUncommittedOccasion`,
`TestReframeRefusesAnOccasionCommittedAfterTheRewrite`,
`TestReframeSerialisesOnTheLedgerLock`. `internal/core/readingitem`:
`TestLocateDispositionRefusesASymlinkedItemDir`,
`TestResolveOccasionRefusesAFamilyItIsNotHanded`. `internal/core/issueschema`:
`TestReframeKnownIsRequiredPlusAfter`, `TestLedgerDirectoriesCarryReframes`.
`internal/core/lint`: `TestRecordSchemaJudgesAReframeRecord` (blank ground,
missing before-hash, an unknown key, a mis-shaped hash, an occasion naming
nothing). `internal/core/reading`: `TestReframeRecordsNeverReachTheBundle`,
`TestExclusionFloorNamesTheReframeRecord`. `internal/core/record`:
`TestDescribeReframeReportsOccasionAndHashes`, `TestRecommendedVerbPathsClosed`
(existing, extended). `internal/surface/cli`: `TestCaptureReframeSurface`
(both halves' renders name the half), `TestNextMoveVerbsResolveInLiveTree`
(existing). `evals`: The read-block lane over the new plant, run explicitly.

## Out of scope

- The prior construal's text, and any path by which it could reach the record.
- Reframes no reading occasioned, and any lint over construal rewrites; both
  wait on the ruling above.
- A reframe of the glossary or of committed scope, which adr-55 counts as
  construal and this record does not; the narrowing is flagged above.
- A proof that the occasion caused the rewrite. The record carries the
  operator's assertion and the one check that the occasion came first.
- The fourth audit verdict; a reframe changes what is built next and puts no
  delivered promise in question.
- A repository with more than one construal surface; the path and heading are
  constants, per the intent's first scope condition.
- Writing the surprise family, and the `record.IDRe` and `commands/abcd.md`
  edits; the admission and surprise spec owns those, and this one inherits
  them and resolves what that verb wrote.
