---
schema_version: 1
id: "iss-2609012039210402"
slug: "an-override-pii-rules-passes-validate-internal-core-rules-ru"
severity: "minor"
category: "bug"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/rules/rules.go"
resolution: "Validate refuses a domain whose merged rules are empty: an override of {\"rules\": []} on a bundled domain, or a custom domain declared without any, fails loading with \"domain NAME: has no rules (it would inject a heading-only block; set state dormant to silence a domain)\", so the loader injects nothing rather than a heading-only \"## NAME\" block wearing the domain's name. The bundled defaults and a dormant state-only override still validate. Proven by TestValidateRefusesDomainWithoutRules, watched failing on v0.7.0."
impact: fix
---

An override `{"PII": {"rules": []}}` passes `Validate` (internal/core/rules/rules.go) and injects a heading-only `## PII` block — suppression wearing the domain's name, which the agent reads as a domain that says nothing; a custom domain declared with recall but no rules does the same. Reproduced at v0.7.0 (a 43-byte block, exit 0). The documented way to silence a domain is `{"state": "dormant"}`; a domain whose merged rules are empty should be refused loudly at load, naming the domain and the dormant remedy, in line with the empty-rule-body refusal (iss-2608261550497978). Sibling of GHSA-22f8-qf5r-gjgq.

## Grounds

- pursued: a loud refusal at load is the shape the empty-rule-body refusal (iss-2608261550497978) already set, and the error names the documented remedy
