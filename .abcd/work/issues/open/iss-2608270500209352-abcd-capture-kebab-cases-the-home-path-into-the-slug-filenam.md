---
schema_version: 1
id: "iss-2608270500209352"
slug: "abcd-capture-kebab-cases-the-home-path-into-the-slug-filenam"
severity: "major"
category: "security"
source: "agent-finding"
found_during: "github-ledger-dedup-2026-08-27"
found_at: "internal/core/capture"
---

abcd capture kebab-cases the home path into the slug/filename before redaction, so the username still lands in the issue filename even though the body is redacted. GitHub mirror: #485