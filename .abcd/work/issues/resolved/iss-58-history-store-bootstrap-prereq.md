---
schema_version: 1
id: "iss-58"
slug: "history-store-bootstrap-prereq"
severity: "minor"
category: "observation"
source: "agent-finding"
found_during: "autonomous-run"
found_at: "internal/core/history"
resolution: "Fixed history bootstrap remediation to name the real verb abcd ahoy install; also corrected sibling comments."
---

history capture requires a bootstrapped ~/.abcd/history store, so the autonomous run cannot dogfood transcript capture until ahoy install runs on this repo (ahoy doctor: 12 detection gaps). Also the error says 'run abcd install' but the real verb is 'abcd ahoy install' — misleading remediation text.