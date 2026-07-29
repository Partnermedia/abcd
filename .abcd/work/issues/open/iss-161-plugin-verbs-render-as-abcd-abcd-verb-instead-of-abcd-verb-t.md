---
schema_version: 1
id: "iss-161"
slug: "plugin-verbs-render-as-abcd-abcd-verb-instead-of-abcd-verb-t"
severity: "major"
category: "observation"
source: "user-observation"
found_during: "marketplace-install smoke test"
found_at: "commands/abcd"
---

Plugin verbs render as /abcd:abcd:<verb> instead of /abcd:<verb>: the verb files live in commands/abcd/, and the harness maps a commands/ subdirectory to a namespace segment, so the directory name duplicates the plugin prefix. Verbs need to live flat under commands/ (or the subdirectory renamed away from the plugin name) for the documented /abcd:<verb> surface.