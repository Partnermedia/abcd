---
schema_version: 1
id: "iss-77"
slug: "launch-payload-omits-agents-hooks"
severity: "major"
impact: fix
category: "bug"
source: "agent-finding"
found_during: "2026-07-12 /abcd:run iss-31"
resolution: "launch payload now includes agents/ and hooks/; bundle-completeness detector guards every plugin-surface dir"
---

launch payload completeness: .abcd/config/launch-payload.json includes only [.claude-plugin, commands, scripts, docs, README, LICENSE, .gitignore] and omits agents/ and hooks/, both of which exist and are auto-discovered Claude-Code plugin surfaces (agents are referenced by the harness; hooks are the prompt-router injection transport). The iss-31 'omits skills/' instance is now STALE (no skills/ dir — reclassified to commands/abcd/ on 2026-07-11, already shipped via the commands include). Deciding the exact public includes set (do agents/ and hooks/ ship as-is? do hooks reference a binary path that must resolve post-install?) is design-shaped — maintainer call. Detector: a bundle-completeness test asserting every auto-discovered plugin surface dir is either included or explicitly excluded-with-reason.