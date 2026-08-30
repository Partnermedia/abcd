---
schema_version: 1
id: "iss-2608301649337965"
slug: "isabsentvalue-reads-an-empty-flow-sequence-as-absent-but-not"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "itd-189-round-5-security"
found_at: "internal/core/lint/schema.go"
---

isAbsentValue reads an empty flow sequence as absent but not an empty flow mapping or an explicit null tag so a blank grounds passes the gate and silences the report

Found by the round-5 security review. Branch-introduced.

`isAbsentValue` special-cases the empty flow SEQUENCE and not the empty flow
MAPPING, and knows no explicit null tag:

```
if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") { ... }
```

So `grounds: {}`, `grounds: { }`, `grounds: {}  # c` and `grounds: !!null`
reach `checkRecordRequiredFields` as present and non-blank, draw zero findings,
and the admission is accepted. `admittedProposals` then keys the pair as
admitted and the proposal vanishes from the outstanding report: a groundless
admission answers a proposal and nothing says so.

What makes this an asymmetry rather than a missing feature is the siblings.
`[]`, `[ ]`, `""`, `''`, `" "`, `~`, `Null`, `|`, `|+` and `>-` are ALL caught.
One flow collection is handled and the other is not.

It also falsifies two claims this branch has already shipped. The resolution
note on iss-2608300935218982 says an empty, whitespace-only or quoted-whitespace
required field is refused in EVERY store that shares the check; and
`commands/capture.md` tells the user that `record_schema` refuses an admission
with a blank `grounds`. Both are currently false.
`capture/blankrequired_parity_test.go` enumerates exactly four blank spellings,
and neither of these is among them.

Remedy: one branch beside the flow-sequence one, plus the two spellings added to
`blankSpellings` and to the schema tests.
