---
schema_version: 1
id: "iss-278"
slug: "five-agent-prompts-read-attacker-influenceable-input-without"
severity: "minor"
category: "security"
source: "user-observation"
found_during: "manual-capture"
found_at: "agents/"
---

Five agent prompts read attacker-influenceable input without the itd-5 contract (ruthless-reviewer, security-reviewer, docs-currency-reviewer, intent-auditor, sota-researcher) and agents/ sits outside both lint roots, so no detector exists for the class; the PQ linter (agents/README.md) is the missing detector