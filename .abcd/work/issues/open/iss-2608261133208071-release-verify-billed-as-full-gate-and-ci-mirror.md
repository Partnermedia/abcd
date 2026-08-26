---
schema_version: 1
id: "iss-2608261133208071"
slug: "release-verify-billed-as-full-gate-and-ci-mirror"
severity: "nitpick"
category: "documentation"
source: "agent-finding"
found_during: "bughunt-round-8"
found_at: ".github/workflows/release.yml:3"
---

release.yml bills verify as the full verification gate mirroring the ci check and record-lint jobs while it deliberately runs a nine-gate subset; the adjective overclaims and the mirror sentence is stale