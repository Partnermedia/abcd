---
schema_version: 1
id: "iss-2608300205343890"
slug: "itd-188-second-round-residue"
severity: "minor"
category: "inconsistency"
source: "impl-review"
found_during: "itd-188 second-round adversarial reviews, 2026-08-30"
found_at: "internal/core/lint/scribecontract_test.go, agents/scribe/fixtures/injection-canary.json, .abcd/development/specs/open/spc-66"
---

itd-188 second-round review residue: spc-66 still asserts the record_schema validation path retracted elsewhere and names the guard root as .abcd/work/ while its amended table says .abcd/work/issues/; the guard's prose exemption (fewer than two separators and no dot) admits a bare shipped-tree directory such as internal/core or a spaced-separator path; the canary's must_not_contain token is one the verbatim-transcription rule requires the output to carry, so no faithful run can satisfy it; the extractor does not fold U+2571, U+29F8, U+FF3C, HTML-entity or percent-encoded separators, does not refuse format (Cf) code points, and its comment overclaims every spelling; legitimate prose such as version pairs and URLs is flagged as outside paths without a comment saying so; the canary's input items carry no identifier while the expected output refers to item-1 and item-2.
