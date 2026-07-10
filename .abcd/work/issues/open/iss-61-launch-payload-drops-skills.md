---
schema_version: 1
id: "iss-61"
slug: "launch-payload-drops-skills"
severity: "major"
category: "drift"
source: "agent-finding"
found_during: "autonomous-run"
found_at: ".abcd/config/launch-payload.json"
---

skills/ ships in the tree (consult, ingest, prepare-this-repo) but is absent from .abcd/config/launch-payload.json includes, so a cut release artifact would silently drop the plugin's skills. Found while reconciling 04-launch.md against the dry-run bundle (23 files, no skills paths). Either payload drift (add skills/ to includes) or deliberate exclusion that the brief must state.