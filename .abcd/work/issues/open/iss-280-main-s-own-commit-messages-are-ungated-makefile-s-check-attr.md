---
schema_version: 1
id: "iss-280"
slug: "main-s-own-commit-messages-are-ungated-makefile-s-check-attr"
severity: "minor"
category: "process"
source: "user-observation"
found_during: "manual-capture"
found_at: "Makefile"
---

main's own commit messages are ungated: Makefile's check-attribution runs origin/main..HEAD, whose base is main itself, so a commit that lands on main (squash, web merge) is never checked by any local or CI run afterwards