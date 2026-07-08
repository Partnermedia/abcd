---
schema_version: 1
id: "iss-45"
slug: "surface-layer-boundary"
severity: "minor"
category: "architectural-insight"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "internal/surface/cli/cli.go"
---

surface-layer boundary leaks: slug derivation — business logic every front door needs — lives in the CLI surface (internal/surface/cli/cli.go:623) so each future front door must reimplement it; cmd/record-lint is a second binary front door duplicating abcd docs lint --config --root, and as a CI gate it has no tests; skills/consult and skills/ingest bypass the abcd binary entirely, their engine being unversioned scripts in the user home (converges with iss-27 corpus-tooling absorption and the script-first-mvp bounds). Detector: a transport-agnostic-core boundary check — decisions live below the surface, front doors only format; one binary front door per capability. Acceptance corpus: the three leaks above.