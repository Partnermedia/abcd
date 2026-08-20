---
schema_version: 1
id: "iss-345"
slug: "stranded-symlink-classifies-foreign"
severity: "major"
category: "observation"
source: "agent-observation"
found_during: "update-mechanism investigation"
found_at: "internal/core/ahoy/store.go"
---

ahoy symlink heal never fires for the entry a plugin update strands: classifyBinTarget owns a symlink only when its dest equals the current plugin root binary, so a PATH entry pointing into the deleted previous plugin cache dir classifies foreign+dangling — no symlink.dangling gap, stepSymlink refuses, and the user is told to resolve manually a link abcd itself wrote. The heal only covers same-root-binary-deleted, a state no real plugin update produces.