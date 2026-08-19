# PR #294 — verified findings (most severe first)

## F1 — the "gates cannot disagree" guarantee is false for double-quoted nulls

`internal/core/capture/parse.go:188`. An issue record carries
`impact: "NULL"`. capture's `decodeScalar` strips the double quotes before
`validateStrict` calls `frontmatter.IsNull`
(`internal/core/capture/validate.go:104`), so the bare `NULL` reads as absent
and `abcd capture resolve` accepts the record. record-lint's
`checkIssueImpact` (`internal/core/lint/lint.go:1817`) reads the raw value
with quotes intact via `frontmatterFields`, `isNull("\"NULL\"")` is false,
`changelog.ParseImpact` fails, and the `issue_impact_valid` blocker fires on
the record capture just accepted. Pre-PR this split existed only for
`"null"`; the PR adds `"Null"` and `"NULL"` while its own comment at
`internal/core/lint/lint.go:2397` claims the split is impossible. Fix at the
right depth: normalise quoting on one side (or unquote consistently in the
shared predicate's callers), not just widen both literal sets.

## F2 — two tests added in this PR pin opposite verdicts for the same bytes

`internal/core/lifeboat/graveyard_abandoned_test.go:167` asserts a file with
`superseded_by: "NULL"` yields no finding (`gvSupersededADRs` unquotes first,
`internal/core/lifeboat/graveyard_abandoned.go:112`), while
`internal/core/frontmatter/frontmatter_test.go:76` asserts
`IsNull("\"NULL\"")` must stay false ("a quoted scalar is not a YAML null").
Post-PR the same bytes are null to `disembark pack`, but `abcd record
describe` (`internal/core/record/record.go:172`/`294`) publishes
`Links["superseded_by"] = "NULL"` and suggests reading a record named
`"NULL"`, and lint's schema check calls it "not a record handle". A future
unifying fix must break one of the two new tests.

## F3 — lint's `isNull` is a hand-synced byte-identical copy

`internal/core/lint/lint.go:2400`. lint already imports
`internal/core/frontmatter` (which has zero internal deps — no cycle), and
the same file already delegates its scanner for exactly this reason
(`frontmatterFields`, `lint.go:2357`). GitHub #290 happened because a
widening landed nowhere; keeping the copy re-arms the same bug shape.
Replace the body with `return frontmatter.IsNull(v)` and delete the
duplicated table test.

## F4 — the mirrored tests cannot detect drift, and have already diverged

`internal/core/lint/isnull_test.go:17`. frontmatter's negative table guards
quoted/case-folded forms; lint's stops at `nullish`. Nothing asserts the two
functions agree, so a future widening of lint's `isNull` to strip quotes
passes lint's test while only the unchanged package's test would catch it.
If the copy is not deleted (F3), add one equivalence test over a shared
table comparing `isNull(v) == frontmatter.IsNull(v)`.

## F5 — capture's agreement test was not widened with the predicates

`internal/core/capture/parse_test.go:126`. `TestValidateStrictImpact` still
iterates only `{"", "null", "~"}`. If `frontmatter.IsNull` is later narrowed
(or capture stops delegating), the GitHub #290 regression re-lands undetected
in the one package whose test should catch it.

## F6 — two of four independent YAML scalar decoders patched; the class remains

`internal/core/memory/yaml.go:526`. memory's `parseScalar` is the only
complete resolver (nulls, bools, quotes, numbers) but is unexported; capture's
`decodeScalar` strips quotes but knows no nulls or bools; lifeboat's
`gvUnquote` strips quotes only; `frontmatter.IsNull` knows nulls only. The
bool axis is untouched (`parseScalar` accepts true/True/TRUE while
`internal/core/memory/vintage.go` tests `== "true"` exactly) and
`yaml.go:949`'s dumper holds a fourth null-literal list on the write path.
Exporting one scalar-resolution helper (frontmatter is the dependency-free
home) fixes the class, not the instance. → routed to an intent.

## F7 — the new doc comment miscounts its own enumeration

`internal/core/frontmatter/frontmatter.go:66`: "an empty value and the four
YAML nulls ""/"null"/"Null"/"NULL"/"~"" names the empty scalar twice and
lists five items under "four"; the lint-side twin at `lint.go:2397` phrases
it correctly, so the two copies now document themselves differently. Also,
the CHANGELOG's "byte-identical" claim is true only of the boolean
expression, not the functions as written.

## F8 — a third impact consumer diagnoses `impact: NULL` differently

`internal/core/changelog/shipped.go:277`. `newRecord` calls `ParseImpact`
with no `IsNull` gate: record-lint reports "impact must be set explicitly"
(missing) while the release cut records "invalid impact" (malformed) for the
same line — operator-facing message inconsistency introduced by this diff.
An `IsNull` gate at the call site restores parity.

## F9 — `t.Fatalf` inside the six-spelling matrix loop

`internal/core/lifeboat/graveyard_abandoned_test.go:174`. A partial
regression reports as a single-spelling failure and the positive control at
lines 178–184 never runs. `t.Errorf` (or a subtest per spelling) reports the
whole matrix in one run.

## F10 — the CHANGELOG entry cites a bare GitHub number

`CHANGELOG.md:33` cites "(#290)" — the file's only non-ledger citation among
~70 record handles — and no corresponding iss-N exists in
`.abcd/work/issues/`, leaving the fix invisible to record-lint, capture, and
the graveyard readers. Capture the bug retroactively and cite the handle.
