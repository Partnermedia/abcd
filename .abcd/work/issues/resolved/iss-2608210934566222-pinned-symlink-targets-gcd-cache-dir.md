---
schema_version: 1
id: "iss-2608210934566222"
slug: "pinned-symlink-targets-gcd-cache-dir"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "plugin-update post-mortem 2026-08-21"
found_at: "internal/core/ahoy/apply.go"
related_intents: ["itd-132"]
resolution: "the default PATH entry is a data-dir owned copy written by installOwnedEntry; a cache-dir symlink is only a loud degraded fallback, so the GC-dangle class is dissolved"
impact: fix
---

The pinned PATH symlink written by ahoy install targets pluginBinaryPath(pluginRoot) — a commit-stamped plugin cache directory. Every plugin update orphans that directory and the harness garbage-collects it ~14 days after .orphaned_at, so for a marketplace user the ~/.local/bin/abcd symlink silently dangles within two weeks of every update. The heal path (symlink.dangling gap + re-running ahoy install) exists but nothing triggers it automatically. Fix follows the persistent-data-dir relocation: point the pinned symlink at the data-dir binary, whose path is stable across updates.