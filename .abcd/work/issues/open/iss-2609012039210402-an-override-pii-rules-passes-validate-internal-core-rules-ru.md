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
---

An override `{"PII": {"rules": []}}` passes `Validate` (internal/core/rules/rules.go) and injects a heading-only `## PII` block — suppression wearing the domain's name, which the agent reads as a domain that says nothing; a custom domain declared with recall but no rules does the same. Reproduced at v0.7.0 (a 43-byte block, exit 0). The documented way to silence a domain is `{"state": "dormant"}`; a domain whose merged rules are empty should be refused loudly at load, naming the domain and the dormant remedy, in line with the empty-rule-body refusal (iss-2608261550497978). Sibling of GHSA-22f8-qf5r-gjgq.
