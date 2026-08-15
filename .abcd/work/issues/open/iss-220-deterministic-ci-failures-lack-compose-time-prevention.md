---
schema_version: 1
id: "iss-220"
slug: "deterministic-ci-failures-lack-compose-time-prevention"
severity: "minor"
category: "process"
source: "user-observation"
found_during: "manual-capture"
---

simple deterministic CI failures have no compose-time prevention or self-repair path: the attribution gate detects a missing PR-body Assisted-by trailer post-hoc (PR 230), but nothing composes or verifies the body at creation time and no loop is authorised to apply the mechanical fix. PR-body composition is facilitator-owned manual work and therefore automation backlog: abcd should compose/verify the PR surfaces it requires for managed repos, and an allowlisted class of deterministic check failures (trailer append, formatting) should be self-repairing