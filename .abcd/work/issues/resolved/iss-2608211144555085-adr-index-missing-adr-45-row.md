---
schema_version: 1
id: "iss-2608211144555085"
slug: "adr-index-missing-adr-45-row"
severity: "minor"
category: "drift"
source: "agent-finding"
found_during: "bughunt-b-round-4"
found_at: ".abcd/development/decisions/adrs/README.md"
resolution: "Appended the adr-45 row to the ADR index. The durable fix (registering the index in index_drift with a fenced region, per iss-38) is left as a separate follow-up — it restructures the table and carries its own CHANGELOG entry."
impact: internal
---

The ADR index (.abcd/development/decisions/adrs/README.md) ends at the adr-44 row while adr-45 is accepted (2026-08-20) and cited as settled law by six live records (the invariants brief, itd-114, spc-33, .abcd/work/issues/README.md, two open issues, DECISIONS.md). 37 ADR files vs 36 index rows — exactly one missing. The index is the declared hand-maintained roll-call (adrs/README.md) and is not among index_drift's four registered indexes, so it is structurally ungated. Third recurrence of this class in the same table (after resolved iss-296 and iss-374), which iss-38's decision (hand-kept indexes must be deleted or gated) predicts.