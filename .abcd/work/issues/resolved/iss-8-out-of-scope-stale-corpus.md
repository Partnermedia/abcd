---
schema_version: 1
id: "iss-8"
slug: "out-of-scope-stale-corpus"
severity: "major"
impact: internal
category: "drift"
source: "review-followup"
found_during: "roadmap-consistency-review"
found_at: ".abcd/development/brief/06-delivery/03-out-of-scope.md"
blocked_by: [iss-3]
resolution: "03-out-of-scope.md regenerated against the live drafts/ corpus under ADR-34: itd-43 removed (now planned+scheduled), ten missing drafts added (itd-57/59/60/61/62/64/70/73/74/75), derivation command simplified to no-exclusions since scheduled intents cannot live in drafts/; verified in lockstep with disk."
---

03-out-of-scope.md omits current drafts itd-74 and itd-75 while including scoped itd-43; regenerate from the actual draft corpus.