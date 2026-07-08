---
schema_version: 1
id: "iss-21"
slug: "itd-launch-name-refs"
severity: "minor"
category: "inconsistency"
source: "agent-finding"
found_during: "intent-dependency-sweep"
found_at: ".abcd/development/intents/drafts/itd-16-hash-chain-merkle-audit.md"
---

itd-8 and itd-16 cite the launch intent as itd-launch by name, not number — a dangling-reference class the itd-78 graph lint should flag; re-point to the concrete itd