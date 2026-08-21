---
schema_version: 1
id: "iss-2608210934566223"
slug: "sessionend-must-never-bootstrap"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "plugin-update post-mortem 2026-08-21"
found_at: "hooks/hooks.json"
---

The SessionEnd hook attempts the blocking bootstrap download before falling back to a PATH binary. After a plugin update lands a fresh binary-less cache dir, an update-then-quit exits through SessionEnd, which starts the ~11MB download, gets cancelled by the harness during shutdown ('Hook cancelled'), and silently drops that session's transcript capture — observed 2026-08-21, session 8db3dbd6 lost. Exit is the one hook with no time budget: reorder to plugin-root binary, then PATH binary, then loud give-up, and never bootstrap at SessionEnd. The persistent-data-dir relocation shrinks the exposure to true first-install-then-quit, but the ordering fix stands on its own.