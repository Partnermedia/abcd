---
schema_version: 1
id: "iss-98"
slug: "what-didnt-only-reverts"
severity: "minor"
category: "tech-debt"
source: "agent-finding"
found_during: "M2 cross-repo gate probe"
found_at: "internal/core/lifeboat/sources_git.go"
---

disembark evidence/what-didnt grounds only on git reverts; repos that abandon via deletion or dead branches (not revert) report blank. Extend the Tier-0 adapter to also read files deleted after substantial history and branches abandoned unmerged — the graveyard adapter already gathers this.