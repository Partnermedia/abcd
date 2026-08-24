---
schema_version: 1
id: "iss-2608240815458212"
slug: "shipped-readme-empty-claim-false"
severity: "minor"
category: "drift"
source: "user-observation"
found_during: "bug-hunt loop round 5 (state issue #368), internal record-consistency hunt + independent adversarial refutation"
found_at: ".abcd/development/intents/shipped/README.md:3"
resolution: "Dropped the false 'is empty' claim; shipped/ front-door now states capabilities move in as their linked specs close, matching the parent index."
impact: internal
resolved_by:
  commit: "1a9d2b88837c2ec74b41d9ffd53a51b48ca1125a"
---

shipped-intents README front-door claims the directory is empty while it holds 18 shipped intents, contradicting the parent intents index which describes it as populated