---
schema_version: 1
id: "iss-99"
slug: "open-questions-no-todo-scan"
severity: "minor"
category: "tech-debt"
source: "agent-finding"
found_during: "M2 cross-repo gate probe"
found_at: "internal/core/lifeboat/sources_conventions.go"
---

disembark has no adapter scanning code for TODO/FIXME markers, so evidence/open-questions is blank on any repo without a native record even when the code is full of TODOs. Add a conventions-tier adapter that greps tracked files for TODO/FIXME/XXX.