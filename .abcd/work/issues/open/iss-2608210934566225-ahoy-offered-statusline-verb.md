---
schema_version: 1
id: "iss-2608210934566225"
slug: "ahoy-offered-statusline-verb"
severity: "minor"
category: "future-work-seed"
source: "user-observation"
found_during: "plugin-update post-mortem 2026-08-21"
---

The harness statusline is a user-level setting no plugin can inject, but ahoy install could offer it as an owned transparent-confirm ConfigChange (same pattern as the PATH symlink): point statusLine at a new 'abcd statusline' verb consuming the harness stdin JSON and appending guard health, version skew, and update-available. Event-driven refresh (300ms debounce, min 1s interval) rules out live download progress; steady-state health only.