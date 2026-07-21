---
schema_version: 1
id: "iss-20"
slug: "launch-cluster-unscheduled"
severity: "major"
impact: internal
category: "inconsistency"
source: "agent-finding"
found_during: "intent-dependency-sweep"
found_at: ".abcd/development/roadmap/phases/README.md"
resolution: "Recorded-why-not, not scheduled: adr-33 (accepted 2026-07-08) already decides the launch deepenings are deliberately unscheduled until sequenced (Phase 1 owns only the curated-release cut), and adr-9 makes no-Scope a valid implicit state. The residue was discoverability: the phase index never named the cluster. phases/README.md § Beyond Phase 6 now names the full cluster (itd-65/66/67/70/72/73, covering itd-67 which adr-33's list omitted) with the adr-33 pointer and the itd-78 derived-priority note."
---

the launch cluster (itd-65/66/67 critical, itd-72 major) is scheduled in no phase Scope section — the silent-stall pattern itd-78's lint should catch; schedule or record why not