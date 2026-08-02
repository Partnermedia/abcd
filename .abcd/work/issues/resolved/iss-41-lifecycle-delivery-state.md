---
schema_version: 1
id: "iss-41"
slug: "lifecycle-delivery-state"
severity: "major"
category: "process"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: ".abcd/development/intents"
resolution: "delivered capability is represented by the intent leaving drafts/ — stated in intents/README.md and gated by the new delivery_state record-lint rule, which blocks an itd-N citation in a CHANGELOG delivery section (Added/Changed) whose intent still sits in drafts/. The two live instances are corrected: the v0.1.0 abcd docs lint entry cited itd-60 (only its deterministic layer shipped; the semantic layer is five open questions) and the v0.2.0 abcd audit entry cited itd-85 (shipped whole, never promoted — tracked as iss-180, blocked on the shipped/ schema wanting a spec id the audit verb never had). Both now describe the delivered capability without the citation. The brief's later-phase list is derived again: index_drift gains dir_entry so a listing can enumerate records by id, and 03-out-of-scope's bullet list sits in a marked region held to drafts/ — regenerated from a nineteen-entry drift (itd-47/73/74 departed, sixteen captures missing), with the four never-ID'd items moved out of a list that could not derive them."
impact: additive
---

lifecycle cannot represent delivered work: v0.1.0 shipped capability from a drafts-stage intent while shipped/ sits empty, so the record goes quiet exactly where it matters; the canonical later-phase list in 03-out-of-scope has drifted from the filesystem it claims lockstep with (superseded itd-47 still listed; drafts itd-76/77/78 missing). Detector (per reality-is-filable): define the interim delivery-state rule (how shipped capability is represented before Phase 4), then a lifecycle lint — no CHANGELOG delivery entry whose intent still sits in drafts/, and out-of-scope lists derived from the filesystem. Acceptance corpus: the v0.1.0 drafts-stage shipment and the two out-of-scope list drifts.