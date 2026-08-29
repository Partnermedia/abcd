---
schema_version: 1
id: "iss-2608292036125100"
slug: "own-home-nested-under-a-longer-root-escapes-for-non-users-homes"
severity: "minor"
category: "security"
source: "impl-review"
found_during: "v0.6.9-security-pass"
found_at: "internal/adapter/scanner/residual.go"
---

v0.6.9 security pass: the caller's own home nested under a longer root — /Volumes/T7<HOME>/x, /mnt/backup<HOME>_snapshot/x — is reported by nothing when HOME is not under /Users or /home (a temp or CI home such as /var/folders/.../<user> or /workspaces/<user>). home_path_self declines it at the leading anchor (the byte before the home is a path-segment byte), home_path_other's regex does not match such a root, local_username fails its word boundary inside the longer token, and the backstop's user-segment rewrite covers only /Users/<user> and /home/<user>. For a /Users or /home home the backstop rewrites the segment, so the shape is covered there. main's unanchored sweep took it everywhere. Fix: extend SurvivingCallerHome's user-segment rewrite to the caller's actual home root (basename plus its parent segment), or waive the leading anchor for the literal home when the byte before it is a path-segment byte and the home is at least two segments long.
