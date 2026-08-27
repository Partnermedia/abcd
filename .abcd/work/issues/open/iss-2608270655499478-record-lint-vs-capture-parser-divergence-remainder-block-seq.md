---
schema_version: 1
id: "iss-2608270655499478"
slug: "record-lint-vs-capture-parser-divergence-remainder-block-seq"
severity: "minor"
category: "tech-debt"
source: "agent-finding"
found_during: "security-cut-agent-flagged-siblings-2026-08-27"
found_at: "internal/core/frontmatter"
---

record-lint vs capture parser divergence remainder: block-sequence frontmatter fields are legitimate and used in 21+ intent/adr records but only capture's strict ledger parser rejects them, so a correct fix is store-scoped (share capture's typed strict parser through the canonical frontmatter package) rather than a universal rejection. The duplicate-key and space-before-colon halves of #357 are fixed; this block-sequence remainder is the follow-up. Flagged by the lint-integrity fix agent.