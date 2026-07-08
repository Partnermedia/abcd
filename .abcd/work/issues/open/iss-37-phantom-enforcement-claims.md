---
schema_version: 1
id: "iss-37"
slug: "phantom-enforcement-claims"
severity: "major"
category: "documentation"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "docs/reference/cli/README.md"
---

phantom enforcement claims: docs/reference/cli/README.md describes CI-generated reference pages and a freshness check that exist nowhere; brief 06-lint.md makes present-tense Delivered claims for lint families internal/core/lint does not implement and cites phantom paths; AGENTS.md attributes a gofmt gate to make preflight that preflight does not run (Makefile:62); a stale ci.yml comment claims record-lint is non-blocking though the step blocks; README and CONTRIBUTING describe the gate suite incompletely. Detector (per enforcement-claims-are-facts): a gate cross-check lint — every named gate, lint code, Makefile target, or workflow step in the record resolves to a live definition; planned checks are written as intents, never present tense. Acceptance corpus: the five instances above.