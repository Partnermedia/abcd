---
schema_version: 1
id: "iss-2608211147403129"
slug: "rfc-frontmatter-contract-disagreement"
severity: "nitpick"
category: "inconsistency"
source: "agent-finding"
found_during: "bughunt-b-round-4"
found_at: ".abcd/development/roadmap/rfcs/README.md"
resolution: "Reconciled rfcs/README.md up to adrs/README.md: widened spawned_from to itd-N or adr-N, added the related_adrs field to the RFC frontmatter spec, and added the RFC->ADR and reverse ADR->RFC rows to the bidirectional table. Conformed rfc-2's frontmatter (spawned_intents/related_intents/related_adrs: [adr-43]/authors), closing adr-43's one-sided pair. The reciprocal lint stays future work per both READMEs."
impact: internal
---

The ADR-RFC pairing contract disagreed across three surfaces. adrs/README.md declared RFCs carry 'related_adrs: [adr-N, ...]', but rfcs/README.md (the RFC frontmatter spec and its bidirectional table) defined no such field; and rfc-2 carried 'spawned_from: adr-43' — a field rfcs/README.md documented as itd-N only — while omitting spawned_intents/related_intents/authors/related_adrs, so adr-43's declared pair (related_rfcs: [rfc-2]) was one-sided. The RFC store is ungated (record_schema registers adr/itd/spc/iss only). rfc-2 was the sole outlier in the two-file store; landed with adr-43 (2026-08-19) with no recorded follow-up. Structurally identical to resolved iss-360 (adr-7 outlier).