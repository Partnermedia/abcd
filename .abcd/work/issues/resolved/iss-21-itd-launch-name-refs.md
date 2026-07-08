---
schema_version: 1
id: "iss-21"
slug: "itd-launch-name-refs"
severity: "minor"
category: "inconsistency"
source: "agent-finding"
found_during: "intent-dependency-sweep"
found_at: ".abcd/development/intents/drafts/itd-16-hash-chain-merkle-audit.md"
resolution: "itd-16's 'per itd-launch' re-pointed to itd-66 (the launch payload manifest, default-deny excludes, and launch.allow opt-in are itd-66's render contract). itd-8 carries no dangling name today — its launch references are to the /abcd:launch command surface and 'the launch payload manifest', with no itd-launch literal left in the corpus (grep clean outside the sweep note and this issue)."
---

itd-8 and itd-16 cite the launch intent as itd-launch by name, not number — a dangling-reference class the itd-78 graph lint should flag; re-point to the concrete itd