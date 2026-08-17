---
schema_version: 1
id: "iss-265"
slug: "task-classes-crosscheck-claim-unbacked"
severity: "minor"
category: "drift"
source: "impl-review"
found_during: "spc-28 build"
found_at: ".abcd/development/brief/02-constraints/04-naming.md"
---

02-constraints/04-naming.md's task_classes enum row claims 'Machine-readable source of truth: the Go binary's task_classes schema (internal/core/...) — a cross-check test fails if this table and the schema diverge', but no task_classes schema or cross-check test exists anywhere in the Go tree (grep confirms zero Go references). spc-28's step 3 instructed updating that test in the same commit as the intent_review→intent_audit token flip; nothing existed to update. Either build the cross-check the row promises or reword the row to name the real source of truth (agent frontmatter + this table).