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