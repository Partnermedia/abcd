---
schema_version: 1
id: "iss-2608301519253368"
slug: "the-join-s-blocker-message-asserts-per-bucket-id-minting-tha"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "itd-189-round-3-security"
found_at: "internal/core/lint/schema.go"
---

the join's blocker message asserts per-bucket id minting that the same rule refuses as a fault and that the branch base already made false

Found by the round-3 security review. Must change before merge.

The bucket leg emits, as a BLOCKER, that reading ids "are minted per bucket and
collide across them, so the pair this record forms is what identifies it". They
are not. `mintUnusedItemID` (capture/reading.go:856-874) probes EVERY run
directory under the ledger lock and redraws on a hit — the fix recorded in
iss-2608300227228575, present in the branch base 932629f9.

The same rule contradicts itself 440 lines away: `record_schema`'s
duplicate-ordinal leg (schema.go:441-449) refuses a cross-run `rdi` collision
outright, and the reviewer fixtured one to confirm it fires. So the rule
simultaneously refuses the collision AND says it happens by construction.

Note iss-2608301411017768 (major) closed the PREVIOUS spelling of this defect by
SCOPING the claim to `rdi` — where it is equally false — rather than removing
it. That is the pattern to break here.

Four sites carry it: the message at schema.go:880, the doc comments at
schema.go:210-216 and :812-816, and readingoutstanding.go:57 ("Reading ids are
minted per run and collide"), which cites the very issue whose fix made it
false.

Remedy: state only what the walk establishes. `admittedProposals` keys the
admitted set on (directory run, proposal), so an admission filed under another
run is keyed on a pair nothing queries — that consequence is TRUE and
SUFFICIENT. Delete the minting clause from all four sites in one change, so one
rule holds one story.
