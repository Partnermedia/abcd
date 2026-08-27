---
schema_version: 1
id: "iss-2608270500202696"
slug: "validatesourceblock-s-single-source-branch-requires-only-cla"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "github-ledger-dedup-2026-08-27"
found_at: "internal/core/memory (validateSourceBlock)"
---

validateSourceBlock's single-source branch requires only class, so an external_* page written without source_hash passes validation and abcd memory lint reports only an MQ003 info whose 'no external source' reason is false, so the MQ001/MQ002 quotation-budget blockers never run. GitHub mirror: #362