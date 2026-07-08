---
schema_version: 1
id: "iss-2"
slug: "release-version-source-of-truth"
severity: "critical"
category: "inconsistency"
source: "review-followup"
found_during: "roadmap-consistency-review"
found_at: ".abcd/development/brief/04-surfaces/04-launch.md"
resolution: "Aligned to ADR-31: roadmap README now tracks releases by the v* tag with the number derived from shipped intents' impact; 04-launch.md bump-tier rule rewritten to impact derivation + surface-diff guardrail (phase-completion tiering and --version dropped); verification matrix row updated."
---

release-version source of truth conflicts: roadmap README tracks releases by plugin.json; ADR-31 derives versions from shipped intents; 04-launch.md still describes --version and phase-completion tiering. ADR-bound.