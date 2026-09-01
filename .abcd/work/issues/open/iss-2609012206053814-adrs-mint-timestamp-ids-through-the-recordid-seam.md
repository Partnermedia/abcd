---
schema_version: 1
id: "iss-2609012206053814"
slug: "adrs-mint-timestamp-ids-through-the-recordid-seam"
severity: "major"
category: "process"
source: "user-observation"
found_during: "design-decisions-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: ".abcd/development/decisions/adrs"
---

Ruled 2026-09-01 (DECISIONS.md, that date): ADRs take timestamp ids through the same recordid seam as issues, because two branches minted 0055 and 0056 on the same day for different decisions. Implementation owed: a verb that mints adr-<timestamp> and writes the ADR skeleton under decisions/adrs with a filename ordered by the stamp; record-lint and the wiki-link resolver admit both the ordinal shape (0001-0058, kept as is) and the timestamp shape; the ADR index and site export sort ordinals first then stamps; adr-45's rollout note names this as the deferred family. Companion to the intent/spec adoption tracked by iss-2608210737260468.
