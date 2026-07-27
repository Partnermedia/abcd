---
schema_version: 1
id: "iss-80"
slug: "record-id-allocators-itd-n-spc-n-iss-n-are-branch-local-para"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "parallel-agent-run"
---

Record id allocators (itd-N, spc-N, iss-N) are branch-local: parallel agents on separate branches each scan for max+1 and mint the SAME id. Two intents both claimed itd-82 and both merged to main (PRs #46, #47). The iss-N allocator hit the same class before (iss-77 collision, manually renumbered to iss-79; class recorded as iss-74). Resolving a collision forces a renumber, which breaks the record's stated 'ids are capture-stable, never renumbered' invariant -- so the minting scheme, not the renumber, is the defect. This capture is the MINTING half: a collision-free allocation scheme is needed (forge-minted / random-suffix / timestamp / reserve-registry -- SOTA under research). The DETECTION half is armed separately: intent_lifecycle now blocks duplicate intent ids.

2026-07-27 instance (itd-101 arc): `intent plan itd-101` on a branch off main minted spc-16 while unmerged PR #156 already carried spc-16 (itd-103); the same session's `abcd capture` then minted iss-146 while #156 already carried iss-146. Both renumbered by hand (spc-17; the duplicate capture folded into this issue). Corpus: two collisions from one planning step, spc-N and iss-N allocators both live.