---
schema_version: 1
id: "iss-43"
slug: "readme-capability-currency"
severity: "major"
category: "documentation"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "README.md"
---

README capability currency: the Status section claims Phase 0 while v0.1.0 shipped through Phase 2 (README.md:40); the surface list claims a review oracle, spec/task engine, and autonomous run each ship as a native default — none exist (README.md:29); the curated-release-excluding-.abcd claim is aspirational — no publish path exists and the live plugin channel ships the whole repo. Detector: extend the docs-currency gate scope to README capability and status claims (every claimed-shipped capability resolves to a wired verb; phase claims cross-checked against the roadmap), with loud-staging wording for the unwired publish path. Acceptance corpus: the three claims above.