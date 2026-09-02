---
schema_version: 1
id: "iss-2609020735528927"
slug: "the-sessionend-hook-shim-in-hooks-hooks-json-does-not-self-p"
severity: "major"
category: "bug"
source: "review-followup"
found_during: "release-v0.7.1-crosscheck"
origin: researcher-authored
production_mode: hand-written
found_at: "hooks/hooks.json"
---

The SessionEnd hook shim in hooks/hooks.json does not self-provision: UserPromptSubmit, PreToolUse and PreCompact each carry the bootstrap.sh attempt with the .bootstrap.attempt throttle when the plugin-root binary is missing, and SessionEnd tests the plugin-root binary, falls to the vouched PATH abcd, then fails loudly. On a plugin root with no binary, four events attempt provisioning and SessionEnd silently loses the session transcript, which is the one event whose loss cannot be recovered on the next session. Both 01-ahoy.md (line 225) and 05-internals/03-configuration.md (line 374) document the four non-SessionStart shims as self-provisioning, so the brief describes the intended shape and the shim is the defect. Found by the v0.7.1 brief-surface crosscheck (Direction B, hook entrypoints).
