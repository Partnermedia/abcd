---
schema_version: 1
id: "iss-241"
slug: "research-prompting-agents-holds-prompt-baselines-for-chat-di"
severity: "minor"
category: "drift"
source: "agent-finding"
found_during: "intent-planning-prep"
found_at: ".abcd/development/research/prompting/agents"
---

research/prompting/agents/ holds prompt baselines for chat-distiller and embark-scaffolder, neither of which exists under agents/ any more — the baseline corpus has drifted from the shipped agent set.