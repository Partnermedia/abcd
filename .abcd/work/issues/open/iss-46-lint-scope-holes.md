---
schema_version: 1
id: "iss-46"
slug: "lint-scope-holes"
severity: "minor"
category: "process"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "Makefile"
---

lint scope holes and gate parity: link-lint does not cover all committed markdown; the persona rule is absent from docs-lint; record-lint blocking semantics are inconsistent between local and CI and there is no warn-baseline ratchet; gofmt is missing from make preflight though attributed to it; repo-local hooks activation and its provisioning dependency are undocumented. Detector (per ratchet-not-big-bang): a lint scope matrix (which rule covers which tree, checked into the record) plus baseline-ratchet support in record-lint so new rules arm immediately against frozen violations. Acceptance corpus: the five holes above.