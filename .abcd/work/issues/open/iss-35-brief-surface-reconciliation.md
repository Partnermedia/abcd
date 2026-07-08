---
schema_version: 1
id: "iss-35"
slug: "brief-surface-reconciliation"
severity: "critical"
category: "drift"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: ".abcd/development/brief/05-internals/08-skills.md"
---

brief-vs-shipped-surface reconciliation: 05-internals/08-skills.md claims abcd ships zero user-facing skills and six top-level commands while 04-surfaces/README.md itself tables nine and /abcd:consult and /abcd:ingest are shipped; the shipped skills violate the brief criterion that any artefact mutation is a command, not a skill; the skills/ layout described (abcd-ahoy, commit-attribution, secrets-and-pii) is fictional vs the real consult/ and ingest/; the implemented, user-reachable abcd docs lint and abcd history verbs have no home in 04-surfaces at all; the operator-internal paragraph contradicts the commands/ directory that exists. Detector (per spec-moves-with-the-surface): a record-lint cross-check that every entry under commands/ and skills/ resolves to a brief surface row, and every brief surface row resolves to a shipped or explicitly staged surface. Acceptance corpus: each falsified claim above — the check fails on all of them today. Fix amends the criterion or the surface in one change, never silently.