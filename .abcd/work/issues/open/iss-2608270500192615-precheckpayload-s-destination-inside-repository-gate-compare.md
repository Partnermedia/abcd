---
schema_version: 1
id: "iss-2608270500192615"
slug: "precheckpayload-s-destination-inside-repository-gate-compare"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "github-ledger-dedup-2026-08-27"
found_at: "internal/core/launch/render.go"
---

PrecheckPayload's destination-inside-repository gate compares paths case-sensitively, so on a case-folding filesystem a case-variant --payload-dir slips the overlap gate and the render writes inside the working tree. GitHub mirror: #329