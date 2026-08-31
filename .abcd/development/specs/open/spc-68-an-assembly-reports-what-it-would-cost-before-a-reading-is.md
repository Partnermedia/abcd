---
id: spc-68
slug: an-assembly-reports-what-it-would-cost-before-a-reading-is
intent: itd-198
origin: researcher-authored
production_mode: hand-written
---
# The size report: a per-kind cost on every assembly, a test kind split from source, and a version that cannot misreport its table

## Summary

spc-68 delivers itd-198: every assembly reports what it would cost before anyone
dispatches a reading. The result carries bytes and a byte-derived token estimate
per material kind and in total, on both the CLI and the plugin page, whether or
not an artefact is written. A `test` kind splits from `source`, because on this
repository test files are 53 per cent of the source and nobody had counted them.

Three mechanism changes make the report checkable rather than asserted, and all
three close holes that exist before this intent rather than holes it opens: the
include table's rendering gains the kind column it never had, the manifest
records each item's kind, and the assembler version stops being able to
misreport the table it was built from.

No budget is enforced and none is invented. The assembler cannot know what a
given reader accepts, so a threshold here would be a guess with a gate attached.
It reports; the operator decides.

## Scope

In:

- A per-kind and total size report on every `Assemble` result.
- A `test` material kind, and the match form the include table needs to express it.
- `Kind` added to `Render()`, closing a latent gap in the version's coverage.
- `Kind` recorded per `ManifestItem`.
- `SchemaVersion` 1 to 2; `AssemblerVersion` 1.0.0 to 1.1.0.
- The assembler version made structurally unable to misreport its table.

Out:

- Making a reading fit. No selection, no budget, no refusal. itd-199 carries that.
- A tokenizer in the binary. The estimate is bytes over a measured constant and says so.
- Reclassifying test-support packages that are not `_test.go` files.
- Splitting `SchemaVersion` into two constants. Named as residue below, not done here.
- Resolving the existing case asymmetry between the two match forms
  (`iss-2608311949421873`).

## Approach

### The `test` kind and the match form it needs

`Kind` gains a ninth member, `KindTest = "test"`, in `include.go`'s closed
vocabulary, ordered immediately after `KindSource` so `Kinds()` reads
source-then-test.

The match grammar cannot express `_test.go`. An entry beginning with a dot is an
extension, matched against `path.Ext`; anything else is an exact basename. Since
`path.Ext("foo_test.go")` is `.go`, the existing `.go` row already claims every
test file, and first-row-wins means a test row must sit above it in `Table`.

The suffix form is carried by **its own row field**, not by a third convention
inside `Match`:

```go
type Row struct {
	// ...
	Match       []string // extension (leading dot) or exact basename
	MatchSuffix []string // basename suffix, matched case-sensitively
	// ...
}
```

A form named by the field it sits in cannot be confused with a form inferred
from a string's first character, so no disambiguation rule against the two
existing forms is needed and none is written. `Row.matches` gains a suffix pass
before the extension pass:

```go
for _, s := range r.MatchSuffix {
	if strings.HasSuffix(base, s) {
		return true
	}
}
```

The comparison is `strings.HasSuffix`, not a case-folded equivalent, because the
Go toolchain recognises only a lowercase `_test.go` as a test file. A report
that labelled `Foo_TEST.go` a test would disagree with the thing it counts.

The new row is inserted immediately above the `.go` source row, admitted at
every position, `MatchSuffix: []string{"_test.go"}`, `Kind: KindTest`, with its
own admitting rule. Both rows admit; only the label differs. `Match` and
`MatchSuffix` are both empty on no row, and a row with neither still admits
every file, as today.

### The size report

`AssembleResult` gains a report computed from the collected candidates, before
any write and independently of whether one happens:

```go
type KindSize struct {
	Kind       Kind `json:"kind"`
	Items      int  `json:"items"`
	Bytes      int  `json:"bytes"`
	TokensEst  int  `json:"tokens_est"`
}

type SizeReport struct {
	ByKind    []KindSize `json:"by_kind"`
	Items     int        `json:"items"`
	Bytes     int        `json:"bytes"`
	TokensEst int        `json:"tokens_est"`
	Basis     string     `json:"basis"`
}
```

`Bytes` counts the UTF-8 length of the item text that actually travels in the
bundle, not the file on disk, so a projected record is counted at what the
reading receives rather than at what the record weighs. `ByKind` is ordered by
`Kinds()`, so the report's row order is the vocabulary's order and is stable
across runs. A kind that passed no item is omitted rather than reported as zero,
because a zero row and an absent kind are different facts and the manifest can
settle which.

`Basis` is a literal string naming the method and the divisor, carried in the
artefact rather than only in the rendering, so a report read out of context
still says what it is. The report is on the result, not on the bundle: ac-8
requires the bundle to gain no field, and a reading has no use for its own
weight.

### The estimate and its divisor

`TokensEst` is `Bytes / tokenBytesPerToken`, integer division, where the divisor
is a package constant measured during this spec build rather than assumed.

### Measurement

Every tracked file in this repository was tokenized — a full census, not a
sample — through `tiktoken` 0.14.0 in a throwaway virtual environment outside
the repository, under both `cl100k_base` and `o200k_base`. 2,575 files,
17,119,789 bytes.

| kind | files | bytes | tokens (cl100k) | bytes/token |
|---|---:|---:|---:|---:|
| source | 277 | 3,740,055 | 974,509 | 3.838 |
| test | 429 | 4,372,000 | 1,230,931 | 3.552 |
| record | 1,667 | 6,951,013 | 1,697,680 | 4.094 |
| doc | 108 | 772,264 | 186,939 | 4.131 |
| config | 94 | 1,284,457 | 338,820 | 3.791 |
| total | 2,575 | 17,119,789 | 4,428,879 | 3.865 |

`o200k_base` differs by under 0.3 per cent on every kind, so the choice of
encoding does not carry the result.

**The divisor is 3.85**, byte-weighted across the corpus and rounded. Three
things about it are stated here rather than discovered later:

- **It is calibrated on a superset of what an assembly passes.** The census
  counts every tracked file; the include table denies most of `.abcd/` and
  reaches only named leaf buckets. A bytes-per-token ratio is a property of the
  material and not of the selection, so the ratio transfers; the totals do not,
  and the census totals are not comparable to the assembly figures in itd-198.
- **A single divisor mis-states each kind, in known directions.** At 3.85:
  source −0.3 per cent, config −1.5 per cent, test −7.7 per cent (under-stated),
  record +6.3 per cent and doc +7.3 per cent (over-stated). The spread is about
  11 percentage points end to end. Per-kind divisors would remove it, and are
  not adopted: itd-198 says a constant, the estimate exists to judge plausibility
  at the order of magnitude, and a table of five tuned constants is exactly the
  tuning the intent forbids. The bias is disclosed rather than removed, which is
  what brief invariant 16 asks of a figure whose examination cannot establish
  more than this.
- **`tiktoken` is OpenAI's tokenizer, so this is a proxy.** No reader's actual
  tokenization is measured here, and none is claimed. It is a stable,
  reproducible reference for calibrating a byte-derived estimate, which is all
  the estimate claims to be. This is why the figure is labelled an estimate at
  every surface it reaches.

The divisor is fixed from evidence, never tuned. itd-198 states that if the
estimate ever changes a decision it should not have, it is replaced by a real
tokenizer rather than adjusted; this spec adds nothing to that and provides no
configuration knob for it.

### `Kind` into `Render()`

`Render()` emits positions, source, matches, fields and rule, and no kind. So a
kind reassignment on an existing row changes every bundle while the version the
manifests carry stands still. That is true before this intent.

The include table's rendering gains a Kind column, and a Suffixes column beside
Matches so the new field is covered too. The readings charter carries the
rendered table between its markers, so `TestReadingsCharterCarriesTheRenderedIncludeTable`
requires the charter to be regenerated in the same change.

`Store` and `Bucket` remain outside the rendering. They route a row through the
record graph and do not decide what a reading receives, so they are not part of
the contract the version names. This is stated rather than left silent, because
the reasoning that admits Kind would otherwise appear to admit them.

### The version that cannot misreport its table

`TestAssemblerVersionCoversTheIncludeTable` compares `sha256(Render())` to a
standalone literal and never reads `AssemblerVersion`. Editing the table and
updating only the literal is green with the version unmoved
(`iss-2608311949385350`). The gate asks a human to move the version; it does not
make them.

A map of version to digest does not fix this — the current version's entry is as
editable as the literal was. The fix is structural: the stamped version carries
the digest of the table it was built from.

```go
// AssemblerVersionCore is the hand-set semver of the assembly contract.
const AssemblerVersionCore = "1.1.0"

// AssemblerVersion is the core semver with the rendered include table's digest
// as build metadata. A table change moves it whether or not a human notices,
// so a manifest can no longer name a version that does not describe its table.
func AssemblerVersion() string {
	sum := sha256.Sum256([]byte(Render()))
	return AssemblerVersionCore + "+" + hex.EncodeToString(sum[:])[:12]
}
```

The core stays hand-set and keeps a pin, because it signals a contract change a
digest cannot: a rewritten rule text moves the digest without changing what the
assembler promises, and a projection change alters the promise without touching
the table. The pin on the core remains advisory. That is now tolerable where it
was not, because the advisory half can no longer make an artefact lie about its
table — which is the property brief invariant 16 requires, and the property the
digest-only literal never delivered.

Call sites move from a constant to a call: `assemble.go` (manifest stamp, result
echo) and `status.go`. Determinism is unaffected: the digest is a pure function
of the table, carries no timestamp, host or absolute path, and two assemblies of
one state at one commit remain byte-identical but for `RunID`.

### `Kind` on the manifest item, and the schema version

```go
type ManifestItem struct {
	ItemKey string `json:"item_key"`
	Path    string `json:"path"`
	Field   string `json:"field,omitempty"`
	Kind    Kind   `json:"kind"`
	SHA256  string `json:"sha256"`
}
```

`Kind` is not `omitempty`: an item without a kind is a defect, and a shape that
can omit the field cannot distinguish one from an old manifest. `DecodeManifest`
is already strict on unknown fields, trailing content and schema version, so it
refuses a v1 manifest against v2 without further work.

`SchemaVersion` moves 1 to 2. It is one constant shared by both artefacts an
assembly writes, so the bundle is restamped although ac-8 holds its shape
unchanged. That is a known consequence of the shared constant, accepted rather
than fixed: splitting it is a larger change than this intent, and making the
split silently, inside a change that needed only one half, is how a shape
version stops meaning anything. Recorded as residue, not as a gap closed here.

### Rendering on both surfaces

`renderAssembleResult` gains the report beneath the existing lines, one row per
kind that passed plus a total, with the basis stated once:

```
itd-198-run: 953 item(s) assembled at the detection position of a1b2c3d4e5f6
  manifest hash: ...
  size:          9.8 MB, ~2,295,107 tokens (estimated: bytes / 3.85, not a tokenizer's count)
    source       446 items   3.6 MB   ~890,436 tokens
    test         407 items   4.1 MB   ~995,776 tokens
    ...
  written:       nothing (dry run; name --out to write the two artefacts)
```

`commands/reading.md` gains the report to its list of fields the host reports
from the JSON. No new operand is added, so the `readingOperands` pin at
`internal/surface/cli/regime_surface_test.go:78` is untouched by this intent —
that pin is itd-199's to move.

## Acceptance criteria mapping

| ac | delivered by |
|---|---|
| 1 | `SizeReport` on `AssembleResult`, computed from candidates before any write |
| 2 | `renderAssembleResult` rows plus the `commands/reading.md` field list; dry-run path unchanged |
| 3 | `SizeReport.Basis`, and the parenthetical in the CLI rendering |
| 4 | the test row and the source row both admit; only `Kind` differs. Pinned by a before-and-after path-set test at all four positions |
| 5 | `MatchSuffix: ["_test.go"]`, `strings.HasSuffix`, row ordered above the `.go` row |
| 6 | Kind and Suffixes columns in `Render()`; charter regenerated |
| 7 | `ManifestItem.Kind`, not `omitempty`; `DecodeManifest` strict |
| 8 | report lives on the result, not the bundle; `BundleItem` unchanged |
| 9 | `AssemblerVersion()` computed from `Render()`'s digest |

## Tests

Every test below is watched fail before the change and pass after.

- `TestTestFilesCarryTheTestKind` — a fixture holding `a.go`, `a_test.go` and
  `A_TEST.go` yields kinds source, test, source.
- `TestKindSplitDoesNotMoveAdmission` — the admitted path set at each of the four
  positions is identical before and after. **Proved by mutation**: deleting the
  new row must leave the path set unchanged and change only the kinds.
- `TestSizeReportSumsToTotal` — per-kind bytes and items sum to the totals, and
  the reported bytes equal the sum of bundle item text lengths.
- `TestSizeReportOmitsKindsThatPassedNothing` — a position passing no test file
  reports no `test` row rather than a zero row.
- `TestDryRunCarriesTheSizeReport` — a dry run writes nothing and still reports.
- `TestRenderIncludesKindAndSuffix` — **proved by mutation**: reassigning an
  existing row's `Kind`, and adding a suffix to a row, each move `Render()`.
  This is the vacuity the old pin had, so it is proved by breaking it, not by
  passing.
- `TestAssemblerVersionCarriesTheTableDigest` — the stamped version's build
  metadata equals the digest of the current rendering; **proved by mutation**:
  a table edit moves the stamped version with no other edit.
- `TestManifestItemRoundTripsKind` — strict decode of an encoded manifest
  preserves every item's kind; a manifest with an item lacking a kind is refused.
- `TestManifestSchemaVersionRefusesV1` — a v1 manifest is refused against v2.

The eval lanes are not run by `make preflight` (`iss-2608311632382737`), so
`make evals-cold-reading` and `make smoke` are run explicitly, and once under
`TMPDIR=/tmp`. The read-block eval is run at a preset that keeps every carrier:
three of its eleven carriers (`main.go`, `fence.go`, `go.mod`) are shipped-tree
files, and `fence.go` is the sole corpus behind the body-redaction row.

## Out of scope

- Any selection, budget or refusal. itd-199.
- A tokenizer in the binary.
- Splitting `SchemaVersion` per artefact. Residue, stated above.
- The case asymmetry between the two existing match forms (`iss-2608311949421873`).
- Making the advisory pin on `AssemblerVersionCore` mechanical. The structural
  digest removes the consequence that made it urgent; the core's own pin remains
  a convention, and this spec says so rather than implying otherwise.
