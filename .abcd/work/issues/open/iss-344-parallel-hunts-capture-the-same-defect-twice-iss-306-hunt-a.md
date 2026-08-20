---
schema_version: 1
id: "iss-344"
slug: "parallel-hunts-capture-the-same-defect-twice-iss-306-hunt-a"
severity: "minor"
category: "architectural-insight"
source: "user-observation"
found_during: "hunt-A round-1 reconciliation"
---

Parallel hunts capture the same defect twice: iss-306 (hunt A round 1, landed late) and iss-339 (hunt B round 2) record the identical githubRemoteRe case-sensitivity bug, minted independently on unmerged branches invisible to each other. Collision-proof ids (itd-114) do not solve this class — it is the semantic sibling of the id collision, sharing the root cause that parallel minters cannot see each other's unmerged mints. Candidate rungs: a capture-time similarity check (itd-84's validator rung), hunt briefings that read open PRs' ledgers, or a shared mint registry as a side effect of itd-114's forge-backed option