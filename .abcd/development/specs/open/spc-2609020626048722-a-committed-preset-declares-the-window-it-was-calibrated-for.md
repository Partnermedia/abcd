---
id: spc-2609020626048722
slug: a-committed-preset-declares-the-window-it-was-calibrated-for
intent: itd-2609020625400445
origin: researcher-authored
production_mode: dictated-and-formatted
---
# A declared window per preset and position: schema version 2 of the preset file, populated record lists, recalibrated cold presets, and an eval that re-measures every preset on every committed change

## Summary

spc-2609020626048722 delivers
[itd-2609020625400445](../../intents/planned/itd-2609020625400445-a-committed-preset-declares-the-window-it-was-calibrated-for.md).
The committed preset file `.abcd/config/reading-presets.json` moves to schema
version 2, where every position entry declares the estimated-token window it
was calibrated for, together with the figure it measured and the commit it
measured on. A version 1 file goes on loading unchanged. The size report
[spc-68](../closed/spc-68-an-assembly-reports-what-it-would-cost-before-a-reading-is.md)
added carries the declaration beside the measurement, and a new eval in the
cold-reading lane assembles every preset at each of the three assembling
positions by dry run and fails when a result exceeds its declaration. The
entailment preset names the claim record it reads as a populated record list,
and the cold presets are recalibrated so that each position keeps every kind
its definition names as its object and declares the window it measures. Two of
the three cold entries measure above 200 thousand estimated tokens, and this
spec says so rather than narrowing the object to make the figure fit: The
record-list narrowing is the lever, named below, for the change that needs the
window to fit that reader.

The assembler still enforces no budget at invocation. The preset declares, the
eval checks, and the report says what a run would cost. The scope grammar of
[spc-69](../closed/spc-69-a-reading-is-about-something-narrower-than-everything-its.md)
is not extended: A scope is still a record id, a material kind or a preset
name, and `.abcd` stays in `denySegments`, so no preset path can reach a record
bucket.

## Scope

In: The version 2 shape and its loader in `internal/core/reading/scope.go`; the
declaration on `SizeReport` in `assemble.go`; the CLI and plugin renderings; the
populated entailment record list; the recalibration of `cold` and the matching
additions to `warm`; the eval `evals/coldreading_window_test.go` and the
guard that proves it can fail; the unit tests for a record scope's positive
half.

Out: A budget at invocation; a tokenizer; the comparative position, whose
object arrives by its own channel
(spc-2609020626039834, the comparative channel) and is bounded by the widening
run rather than by the tree; item-set distinctness across positions
([iss-2608311501240566](../../../work/issues/open/iss-2608311501240566-three-of-the-four-reading-positions-receive-a-byte-identical.md));
the eval lane becoming a required check
([iss-2608311632382737](../../../work/issues/open/iss-2608311632382737-the-pre-push-gate-is-blind-to-both-eval-lanes-so-the-read-bl.md),
[iss-2608311051046981](../../../work/issues/open/iss-2608311051046981-the-new-cold-reading-evals-ci-job-is-not-a-required-status-c.md)),
which is the gate work the enforcement claim below depends on.

## Approach

### The preset file at schema version 2

```json
{
  "schema_version": 2,
  "presets": {
    "cold": {
      "comment": "Each entry keeps every kind its position's definition names and declares what it measured. ...",
      "positions": {
        "entailment": {
          "kinds": ["brief-section", "glossary-term", "discipline", "spec"],
          "records": ["itd-177", "itd-178", "..."],
          "paths": [],
          "window": {
            "tokens_est": 230000,
            "measured_tokens_est": 227261,
            "measured_bytes": 874956,
            "measured_at": "<sha>"
          }
        }
      }
    }
  }
}
```

`PositionScope` gains `Window *Window` (`json:"window,omitempty"`) and
`Comment string` (`json:"comment,omitempty"`); `Preset` gains `Comment` too.
`Window` is `TokensEst int`, `MeasuredTokensEst int`, `MeasuredBytes int`,
`MeasuredAt string`. `tokens_est` is the declaration, on the size report's own
basis (bytes divided by 3.85, per spc-68). The three `measured_*` keys are
disclosure and nothing gates on them beyond shape: `measured_at` must match the
assembler's `targetRe`, and its reachability is deliberately not checked,
because a squash or rebase merge rewrites a branch sha out of existence and a
disclosure that fails the build after one would teach people to omit it.
`comment` is free text nothing reads except a reviewer; the loader declares it
so the strict decoder admits it.

This spec owns preset schema version 2. An entry another spec adds to the file
conforms to the committed version at its own landing: The comparative entry
that the comparative channel adds carries a `window` like every other entry.

### Loading: two versions, one strict

`PresetSchemaVersion` becomes 2 and a `presetSchemaVersions` set holds 1 and 2.
`LoadPresets` accepts either, then applies the version-2 rule: Every position
entry of every preset must carry `window` with `tokens_est` greater than zero,
`measured_tokens_est` and `measured_bytes` at least zero, and a well-formed
`measured_at`; a missing window refuses with the preset name and the position
("preset cold at entailment declares no window; at schema_version 2 every
position states the window it was calibrated for"). A version 1 file declaring
a window is refused as an unknown field by the decoder it already has, so the
two shapes cannot be mixed. The effective window of a preset at a position is
its own entry's; a preset that inherits a position through `extends` without an
entry of its own inherits the parent's window as well, since the union is then
the parent's selectors exactly. A new `PresetWindow(pf, position, token) *Window`
beside `ResolveScope` returns the effective declaration, nil for a record-id or
kind token, which are not presets and were calibrated for nothing. `Scope`
itself is untouched, so the manifest's shape and hash do not move.

### The size report carries the declaration

`SizeReport` in `assemble.go` gains `Window *Window` (`json:"window,omitempty"`)
and `ExceedsWindow bool`, filled from `PresetWindow` in `Assemble`. The report
lives on `AssembleResult`, not on either artefact, so `SchemaVersion` and
`AssemblerVersionCore` do not move. `renderSizeReport` in
`internal/surface/cli/reading.go` gains one line after the total:
`window:        230,000 tokens declared (measured ~227,300 at <sha>); this run is within it`,
or `window:        none declared (preset schema version 1)`, or
`window:        none declared (a record id or kind is not a preset)`; a run over
the declaration says `EXCEEDS` on the same line. `commands/reading.md` adds
`size.window` and `size.exceeds_window` to the fields the host reports.

### Which records populate the entailment preset, and why

The entailment definition names as its object "the claim record, the draft and
planned intents included", read against the constraint sources. Under `cold`
the claim record is named as a record list, because that is the one selector
form that says which claims a reading is about, and because a selector on the
`intent-projection` kind would sweep in every draft in the store: The
entailment reading of Iteration 2's closing run is about the ledger design, not
about every draft anyone has filed. The list is every intent of the
cold-reading workstream: The fifteen shipped intents `itd-177` to `itd-189`,
`itd-198` and `itd-199`, and the eight planned Iteration 2 intents
(`itd-2609020625400169`, `itd-2609020625400194`, `itd-2609020625400445`,
`itd-2609020625402518`, `itd-2609020625402599`, `itd-2609020625405170`,
`itd-2609020625405251`, `itd-2609020625407419`). A record selector matches the
record's filename by id whatever bucket it sits in, so a listed intent follows
its record from `planned/` to `shipped/` without an edit.

The list is maintained by hand and held by review: An intent added to the
workstream is added here in the same change, and the preset's `comment` says
so. The eval is silent on that rule. It measures what the listed ids select and
`TestShippedEntailmentRecordsAllResolve` refuses a listed id that names no
record, so a mistyped id cannot select nothing under cover of the union; but
nothing checks that the list covers the workstream, and a workstream intent
left off the list is a gap review has to see. The alternative, deriving the
list from a named root's `builds_on` closure, is not taken: The closure is not
a workstream (itd-177 builds on records outside it), and a derived list is one
nobody reads. The override stamp remains the safety valve for a run that
departs from the list.

The kinds beside the list are the constraint sources the definition names:
`brief-section`, `glossary-term`, `discipline` and `spec`. None is dropped;
the intent's In Scope says the recalibration drops no kind the position's
definition names, and the `spec` kind is what carries the design record the
claims were built against. The sixth criterion is read per selector: Every
`intent-projection` item at this position names a listed record, because the
claim record arrives through the record selector alone; the specs and the
other three kinds arrive through their kind selectors as constraint sources,
not as claim record. The record-list narrowing that would close the letter of
that criterion over specs as well, naming the workstream's own specs by id in
place of the `spec` kind (a record selector already admits `spc-N`), is the
lever this spec names and does not pull.

### The recalibration, and how it was measured

Every figure below is an estimate on the size report's basis. The baseline is
the committed version 1 file, measured by dry-run assembly on the tree this
spec was written against (`13121721`, which merges main at `24c7ac87`; the
intent's figures at `8f68ffb3` are the same tree one merge earlier), per
material kind:

| kind | items | bytes | est. tokens |
|---|---:|---:|---:|
| brief-section | 12 | 94,815 | 24,627 |
| glossary-term | 20 | 50,271 | 13,057 |
| discipline | 14 | 112,853 | 29,312 |
| intent-projection at entailment (shipped, drafts, planned) | 467 | 407,613 | 105,873 |
| intent-projection at detection (shipped only) | 120 | 92,800 | 24,103 |
| spec (34 closed, 35 open) | 69 | 553,435 | 143,749 |
| source, test, doc, config (the whole tree) | 121 | 1,788,177 | 464,459 |

Under the committed file: `cold` widening 257,939 bytes (66,997), entailment
961,048 (249,622), detection 761,103 (197,689); `warm` widening 892,838
(231,905), entailment 1,073,901 (278,935), detection 2,549,280 (662,150).

The recalibrated file was measured in a clean clone of that commit, the file
committed there so the dirty gate passed by construction rather than by
skipping, under the version 1 shape of the same selectors, because the
binary in the clone predates the version 2 loader. The measuring commit in the
clone is `c7e18694`; a second commit `b660f3fe` adds the eight planned
Iteration 2 intents as they stand in the working tree, since they are not yet
in the measured tree and the entailment list names them. The implementer
re-measures on the commit the change lands on and records those figures as
`measured_*`; the declaration is the figure measured, rounded up to the next
ten thousand, and it moves with the measurement.

| preset | position | items | bytes | est. tokens | declared | by kind (est. tokens) |
|---|---|---:|---:|---:|---:|---|
| cold | widening | 46 | 257,939 | 66,997 | 70,000 | brief-section 24,627; glossary-term 13,057; discipline 29,312 |
| cold | entailment | 213 | 874,956 | 227,261 | 230,000 | brief-section 24,627; glossary-term 13,057; discipline 29,312; spec 143,749; intent-projection 16,514 (98 items over the 23 listed ids; 9,092 over the fifteen shipped alone, at `c7e18694`: 219,839 in all) |
| cold | detection | 317 | 2,348,817 | 610,082 | 620,000 | brief-section 24,627; discipline 29,312; spec 143,749; intent-projection 24,103 (120 shipped intents); source 174,578 (38 files); test 209,972 (63 files); doc 3,737 (`commands/reading.md`) |
| warm | widening | 92 | 892,838 | 231,905 | 240,000 | cold widening plus doc 164,908 |
| warm | entailment | 622 | 1,247,561 | 324,041 | 330,000 | cold entailment plus every other intent projection (507 items, 113,295 in all) |
| warm | detection | 383 | 3,156,071 | 819,758 | 820,000 | cold detection plus doc 164,908 (46 files) and config 48,504 (21 files) |

- **`cold` widening**: Unchanged (`brief-section`, `glossary-term`,
  `discipline`); 66,997 tokens, declared 70,000.
- **`cold` entailment**: `brief-section`, `glossary-term`, `discipline` and
  `spec`, plus the record list above. Measured 227,261 tokens with the eight
  planned intents present, declared 230,000. That is above 200 thousand,
  and this spec says so plainly: The object is the claim record read against
  every constraint source the definition names, and the specs alone are
  143,749 of it. The lever that keeps the object and cuts the size is the
  record-list narrowing named above (the workstream's specs by id in place of
  the `spec` kind), which takes the entry to 139,031 tokens, measured at a
  third commit in the clone, `5bc54d36`: The fifteen workstream specs
  `spc-55` to `spc-69` weigh 55,519 against the 143,749 of every spec. It
  is not pulled here because the intent's In Scope holds the recalibration to
  dropping no kind the definition names, and whether a location narrowing
  within the specs is a narrowing of the object is a question this spec leaves
  as it found it.
- **`cold` detection**: `brief-section`, `discipline`, `spec`,
  `intent-projection` and a bounded slice of the shipped tree by path:
  `internal/core/capture`, `internal/core/issueschema`,
  `internal/surface/cli/reading.go`, `commands/reading.md` and
  `internal/core/lint` (the cold-reading lint rules and the lint package they
  sit in). The detection definition's object is "the shipped tree read against
  the claim record", and the committed cold entry handed the design record
  while handing no shipped intent and no tree at all; the tree is the object
  the definition names first, and a location narrowing within it is what
  itd-199's `paths` selector exists for. Measured 610,082 tokens, declared
  620,000: Three times the 200-thousand reader, said plainly. The tree slice
  is 388,287 of it, dominated by the lint package's tests (209,972 across the
  63 test files under the five paths), and the specs a further 143,749. The
  record-list narrowing is again the lever on the record side (the same
  workstream specs by id: 521,852 tokens at `5bc54d36`, 88,230 fewer), and a
  shorter path list is the lever on the tree side; neither is pulled here for
  the reason given above.
- **`warm`** keeps `extends: cold`. Its own entries become widening `doc`
  (unchanged), entailment `intent-projection` (so `warm` at entailment hands
  what it handed before, the whole claim record including every draft, plus
  the two kinds `cold` gained), and detection `doc`, `config` with the path
  `internal/core/lint`, unchanged. A path selector selects every candidate at
  or beneath the directory whatever its kind, so the committed `warm` detection
  entry already hands the source and tests under `internal/core/lint` without
  naming the `source` or `test` kind; naming either kind would select the whole
  tree, which is not what `warm` handed before, so neither is added. Measured
  231,905, 324,041 and 819,758 tokens; declared 240,000, 330,000 and
  820,000.

The declaration is what the entry measured, rounded up to the next ten
thousand, so the eval speaks as soon as the tree grows past the rounding; that
is the point of it. A declaration is a disclosure of what a preset was
measured at, not a claim about a reader, and the `cold` comment says which of
its entries fit a 200-thousand-token reader (widening) and which do not. The
divisor's disclosed spread (record and doc material over-stated by 6 to 7 per
cent, test material under-stated by 8 per cent) is a property of the estimate
and applies to the declaration and the measurement alike, so it does not
move the comparison. Every figure keeps its estimate label.

### The eval, and what keeps it from passing vacuously

`evals/coldreading_window_test.go`, under `//go:build smoke || coldreading`,
joins the lane with no Makefile or workflow edit, exactly as the harness
README says a later eval does. `TestEveryCommittedPresetFitsItsDeclaredWindow`:

1. Resolves the module root (`..`) and its `HEAD` sha, clones it into a
   temporary directory and checks out that sha detached, so the measured tree
   is the checkout's commit with no uncommitted edit in it; the dirty gate and
   the tracked-file check are then satisfied by construction rather than
   skipped. `HOME` is redirected to an empty temporary directory, as every eval
   in the lane redirects it.
2. Parses the clone's `.abcd/config/reading-presets.json` with the eval's own
   minimal struct (version, presets, `extends`, positions, `window.tokens_est`).
   The oracle rule holds: `bannedImports` already refuses
   `internal/core/reading`, so the eval's account of the declaration is
   independent of the assembler's.
3. For each preset (sorted) and each of the three assembling positions,
   named in the eval as `widening`, `entailment` and `detection`, with an
   effective window, runs
   `abcd reading assemble --position <p> --target HEAD --scope <preset> --dry-run --json`
   through `runIn`, requires the binary's `size.window.tokens_est` to equal the
   eval's own parse, and collects a breach when `size.tokens_est` exceeds it.
   The comparative position is exempt from this eval by name: Its object is
   bounded by the widening run it is handed, not by the tree, so a window over
   the tree measures nothing about it, and the comparative channel's own eval
   covers it.
4. Fails naming, for each breach, the preset, the position, the measured
   figure (tokens and bytes) and the declaration.

Non-vacuity: The test requires the number of pairs measured to equal the number
of position entries the file declares at the three named positions, and
refuses zero. The negative control, `TestTheWindowCheckReportsABreach`, lowers
one declaration in the clone to one below the figure just measured, commits it
there under the fixture identity so the dirty gate admits it, re-runs, and
requires exactly one breach naming all four facts; `windowBreaches` is the one
function both tests call.

The enforcement this delivers is what the intent now claims: Drift past a
window is caught by the cold-reading eval lane, which is not yet a required
check. The eval runs on every committed change in the eval lane; it does not
run in `make preflight`, and a pull request can merge without it until the two
issues named under Scope are closed. That is the gate work this depends on for
the claim to become "caught by the build".

### Landing order

The eight Iteration 2 specs land in this order: This spec (PRE), then the
condition verb (CND), the reading-occasioned origin (ORG), the comparative
channel (CMP), admission and surprise (ADM), the reframe record (RFM), the
scribe (SCR) and principles (PRN). CND lands strictly before PRN; CMP before
ADM before SCR; RFM after ADM and after CMP. This spec lands first because it
owns preset schema version 2 and the window eval, which every later preset
entry conforms to. It moves neither `SchemaVersion` nor `AssemblerVersionCore`:
the preset file and the size report are outside both artefacts.

### The target is stated, not enforced

Two hundred thousand estimated tokens is the target a cold preset aims at; the maintainer ruled on 2026-09-02 that any figure inside the reader's window is acceptable and that a figure over the target is stated to the operator. The size report therefore carries one over-target line and no refusal, and the declared window stays what the preset measures.

## How the Acceptance Criteria are satisfied

- **ac-1 (version 2 without a window refuses).** The version-2 rule in
  `LoadPresets`; `TestPresetV2RefusesAPositionWithoutAWindow` asserts the
  message names the preset and the position.
- **ac-2 (version 1 loads, report says none).** `presetSchemaVersions` admits
  1; `Scope.Window` is nil; the report renders "none declared".
  `TestPresetV1LoadsAndReportsNoWindow`.
- **ac-3 (committed presets pass).** `TestEveryCommittedPresetFitsItsDeclaredWindow`
  over the recalibrated file, each of the three assembling positions under
  each preset at or below the declaration it carries.
- **ac-4 (an undersized declaration fails, naming four facts).**
  `TestTheWindowCheckReportsABreach`.
- **ac-5 (a scope naming one shipped intent carries it alone).** The record
  selector and `pathNamesRecord`; `TestARecordScopeCarriesThatRecordAlone`, a
  unit test in `internal/core/reading` over a fixture holding two shipped
  intents, asserts every manifest item names the scoped intent, the projected
  fields present are exactly that intent's, and the manifest lists exactly
  those items.
- **ac-6 (a populated record list selects only listed records).**
  `TestAPresetRecordListSelectsOnlyListedRecords` over a version-2 fixture
  preset with `records: [itd-A]` and `kinds: [glossary-term]` at entailment:
  Every item whose path names a record names `itd-A`, and the glossary items
  are present. `TestShippedEntailmentRecordsAllResolve` holds the committed
  list to ids that exist.
- **ac-7 (each declaration carries its figure and commit).**
  `TestShippedPresetsDeclareMeasuredFigures` loads the committed file and
  requires every window to carry `measured_tokens_est`, `measured_bytes` and a
  well-formed `measured_at`.
- **ac-8 (over-target line).** `renderSizeReport` compares the total estimate with `TargetTokens = 200000`, a constant beside the divisor in `assemble.go`, and appends one line, `over target: <estimate> estimated tokens against a target of 200,000; the reader's window decides whether this is acceptable`, when the total exceeds it; under the target the line is absent. The target is a statement to the operator, never a refusal, and the JSON result carries `over_target: true` beside the figures. `TestSizeReportNamesAnOverTargetTotal` renders a result above and one below the constant and asserts the line's presence and absence.

## Tests

Watched fail before, pass after. `internal/core/reading`:
`TestPresetV2RefusesAPositionWithoutAWindow`, `TestPresetV1LoadsAndReportsNoWindow`,
`TestPresetV1RefusesAWindow`, `TestInheritedPositionTakesTheParentWindow`,
`TestSizeReportCarriesTheDeclarationAndTheVerdict` (mutation: A declaration one
below the measured figure flips `ExceedsWindow`),
`TestARecordScopeCarriesThatRecordAlone`, `TestAPresetRecordListSelectsOnlyListedRecords`,
`TestShippedEntailmentRecordsAllResolve`, `TestShippedPresetsDeclareMeasuredFigures`;
the existing `TestWarmContainsCold`, `TestThreePositionsCarryDistinctItemSets`
and the three committed-file tests of spc-69 stay green over the new file.
`internal/surface/cli`: `TestRenderSizeReportSaysWhetherAWindowIsDeclared`.
`evals`: The two tests above. Both eval lanes are run explicitly and once under
`TMPDIR=/tmp`; the lane runs on every committed change and is not in
`make preflight`.

## Out of scope

- A budget enforced at invocation, and any refusal for size.
- A tokenizer; the basis stays bytes divided by 3.85 and says so.
- The comparative position, whose object arrives by its own channel and is
  bounded by the widening run; its window eval is the comparative channel's.
- Pulling the record-list lever on the two cold entries that measure above
  200 thousand tokens; this spec declares what they measure.
- Checking that the entailment record list covers the workstream; the eval is
  silent on it and review holds it.
- Item-set distinctness across positions, and the eval lane becoming a
  required status check, each with the issue named above.
- Selecting a record bucket from a preset. `.abcd` is a denied segment, so a
  path clause cannot name `specs/closed`; a bucket selector would be new
  grammar, which adr-58 closed.
