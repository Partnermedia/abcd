---
schema_version: 1
id: "iss-162"
slug: "the-surface-s-own-guidance-is-unreachable-under-the-double-n"
severity: "major"
category: "observation"
source: "user-observation"
found_during: "marketplace-install smoke test"
found_at: "commands/abcd/ahoy.md"
blocked_by: [iss-161]
---

The surface's own guidance is unreachable under the double namespace: ahoy's next-step output (and the command docs) reference /abcd:ahoy and skill-style args like 'install', but the harness registers the verb as /abcd:abcd:ahoy — following the printed instruction yields 'Unknown command: /abcd:ahoy' / 'Args from unknown skill: install'. The rendered guidance must match the registered command names, whatever iss-161 settles on.