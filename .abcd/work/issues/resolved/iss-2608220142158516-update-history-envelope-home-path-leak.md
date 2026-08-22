---
schema_version: 1
id: "iss-2608220142158516"
slug: "update-history-envelope-home-path-leak"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "bughunt-b-round-7"
found_at: "internal/core/update/update.go"
resolution: "update refusal detail/target_path and history record path redacted to ~ via fsutil.RedactHome at the render boundary"
impact: fix
resolved_by:
  commit: "643c010"
---

abcd update and abcd history list/show/capture leak absolute home-rooted paths into their success/refusal envelopes (text and --json) because the CLI scrubPaths runs only on the error value; the update refusal detail fires on the documented plugin-session path and the plugin relays target_path into agent chat