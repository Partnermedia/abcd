---
schema_version: 1
id: "iss-41"
slug: "lifecycle-delivery-state"
severity: "major"
category: "process"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: ".abcd/development/intents"
---

lifecycle cannot represent delivered work: v0.1.0 shipped capability from a drafts-stage intent while shipped/ sits empty, so the record goes quiet exactly where it matters; the canonical later-phase list in 03-out-of-scope has drifted from the filesystem it claims lockstep with (superseded itd-47 still listed; drafts itd-76/77/78 missing). Detector (per reality-is-filable): define the interim delivery-state rule (how shipped capability is represented before Phase 4), then a lifecycle lint — no CHANGELOG delivery entry whose intent still sits in drafts/, and out-of-scope lists derived from the filesystem. Acceptance corpus: the v0.1.0 drafts-stage shipment and the two out-of-scope list drifts.