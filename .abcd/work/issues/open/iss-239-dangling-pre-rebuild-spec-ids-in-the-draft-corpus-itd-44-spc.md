---
schema_version: 1
id: "iss-239"
slug: "dangling-pre-rebuild-spec-ids-in-the-draft-corpus-itd-44-spc"
severity: "major"
category: "drift"
source: "agent-finding"
found_during: "intent-planning-prep"
found_at: ".abcd/development/intents/README.md"
---

Dangling pre-rebuild spec ids in the draft corpus: itd-44 (spc-56), itd-51 (spc-33/37), itd-57 (spc-48/60/62), itd-61 (spc-75) cite specs from the numbering adr-21/adr-26 reset (live store is spc-1..spc-22). Worse, intents/README.md and itd-34 assert the itd-44 lineage 'landed as itd-44 (spc-56 thin adoption)' — no such spec or decision-verdict code exists, so the record claims a delivery the tree cannot back.