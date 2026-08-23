---
schema_version: 1
id: "iss-2608221126066632"
slug: "live-recurrence-2026-08-22-two-concurrent-sessions-in-two-cu"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
---

Live recurrence, 2026-08-22: two concurrent sessions in two current checkouts each minted itd-144 and itd-145 in the same window — four records, two ids — resolved by hand-renumbering one side to itd-146/147. This is the second, undocumented path into the sequential-id collision recorded as iss-2608220150157512, which describes the stale-view path adr-45 already anticipated. Staleness was NOT the cause here: one session fast-forwarded 46 commits immediately before minting and still collided. The structural fact is that intents and specs mint max+1 (nextIntentID) while withIntentMintLock in internal/core/intent/create.go is advisory and scoped to a single checkout, so it cannot see a sibling worktree; two current checkouts therefore allocate the same id by construction. The durable fix is the timestamp mint adoption recorded as iss-2608210737260468. Recorded deliberately because capture has no add-evidence path and itd-87 (recurrence escalation) is unbuilt, so this evidence would otherwise exist only in two session transcripts.