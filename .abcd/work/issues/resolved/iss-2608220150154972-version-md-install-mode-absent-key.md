---
schema_version: 1
id: "iss-2608220150154972"
slug: "version-md-install-mode-absent-key"
severity: "nitpick"
category: "inconsistency"
source: "agent-finding"
found_during: "bughunt-b-round-7"
found_at: "commands/version.md"
resolution: "version.md notes install_mode is omitted when no abcd-owned PATH entry is resolvable"
impact: fix
resolved_by:
  commit: "554f97f"
---

commands/version.md tells the agent to report install_mode from the JSON but the field is json omitempty and absent whenever no abcd-owned PATH entry is resolvable (fresh install, foreign/dangling entry, unresolved plugin root), so an agent following it reports a missing field or invents one