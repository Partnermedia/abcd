---
schema_version: 1
id: "iss-2608300143315694"
slug: "itd-188-review-consistency-findings"
severity: "minor"
category: "inconsistency"
source: "impl-review"
found_during: "itd-188 adversarial review, 2026-08-30"
found_at: "internal/core/lint/scribecontract_test.go, .abcd/development/brief/05-internals/05-prompt-quality.md, .abcd/development/specs/open/spc-66"
---

itd-188 review consistency findings: spc-66 maps acceptance criterion 2 to TestReadingAndScribeSessionsAreDistinctRecords in internal/core/history, which the build neither adds nor declares dropped (deliver it or amend the mapping to say the criterion is procedural); 05-prompt-quality.md still counts ten prompts and six outside the roster after the eleventh landed; the guard test duplicates indexTokenRe and a readRepoFile helper, carries a control that compares a literal to itself, and repeats the canary presence checks the contract rule already runs; the canary's manifest path does not match spc-58's manifest home.
