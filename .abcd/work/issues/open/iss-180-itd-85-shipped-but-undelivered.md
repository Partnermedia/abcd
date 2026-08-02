---
schema_version: 1
id: "iss-180"
slug: "itd-85-shipped-but-undelivered"
severity: "major"
category: "process"
source: "agent-finding"
found_during: "iss-41 delivery-state pass"
found_at: ".abcd/development/intents/drafts"
---

itd-85 shipped whole but still sits in intents/drafts/: abcd audit ships with all five v1 rules, the tri-state exit and the prepare-this-repo wiring the intent scopes, yet the intent was never promoted, so the record cannot say the capability is delivered. Promotion is blocked on the shipped/ schema, which requires a non-null spec_id and an impact value the intent has never carried (there is no spc-N for the audit verb). Disposition: mint or retro-link a spec via /abcd:intent link, set impact, run the fidelity review, and move drafts/ -> shipped/. Until then the v0.2.0 CHANGELOG entry describes abcd audit without citing itd-85 (iss-41).