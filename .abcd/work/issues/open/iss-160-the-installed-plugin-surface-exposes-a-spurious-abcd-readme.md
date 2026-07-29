---
schema_version: 1
id: "iss-160"
slug: "the-installed-plugin-surface-exposes-a-spurious-abcd-readme"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "marketplace-install smoke test"
found_at: "commands/README.md"
---

The installed plugin surface exposes a spurious /abcd:README command: the harness registers every markdown file under commands/ as a slash command, so the folder's own README.md leaks into the command list. The doc needs a home or form that the command loader ignores.