---
schema_version: 1
id: "iss-345"
slug: "stranded-symlink-classifies-foreign"
severity: "major"
category: "observation"
source: "agent-observation"
found_during: "update-mechanism investigation"
found_at: "internal/core/ahoy/store.go"
resolution: "classifyBinTarget/detect/uninstall unified on strandedSiblingDest: the entry a plugin update strands is owned, healed by ahoy install, removed by ahoy uninstall, and never rewritten past a declined approval"
impact: fix
resolved_by:
  commit: "ab10fef"
---

ahoy symlink heal never fires for the entry a plugin update strands: classifyBinTarget owns a symlink only when its dest equals the current plugin root binary, so a PATH entry pointing into the deleted previous plugin cache dir classifies foreign+dangling — no symlink.dangling gap, stepSymlink refuses, and the user is told to resolve manually a link abcd itself wrote. The heal only covers same-root-binary-deleted, a state no real plugin update produces.