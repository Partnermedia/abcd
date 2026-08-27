---
schema_version: 1
id: "iss-2608270908346940"
slug: "two-adrs-sharing-a-filename-ordinal-make-the-second-unreacha"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "issue-sweep-2026-08-27"
found_at: "internal/core/record/record.go"
---

two ADRs sharing a filename ordinal make the second unreachable: describeADR routes on the first matching ordinal and no lint rule checks ADR id uniqueness, the backstop recordid.Resolver's first-in-scan-order comment assumes exists