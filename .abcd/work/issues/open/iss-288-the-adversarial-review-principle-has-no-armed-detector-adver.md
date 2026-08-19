---
schema_version: 1
id: "iss-288"
slug: "the-adversarial-review-principle-has-no-armed-detector-adver"
severity: "minor"
category: "future-work-seed"
source: "user-observation"
found_during: "manual-capture"
found_at: ".abcd/development/principles/adversarial-review-scales-with-blast-radius.md"
---

The adversarial-review principle has no armed detector: adversarial-review-scales-with-blast-radius requires two independent reviews before an intent moves drafts/ to planned/ (and one before an ADR is accepted), but nothing refuses the move without them. The enforceable rung: intent ready (and the record checks) refuse the transition unless review receipts exist in the .abcd/work/reviews/ VSA shape for the draft at its reviewed content hash