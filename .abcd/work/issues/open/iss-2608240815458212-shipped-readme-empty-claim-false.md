---
schema_version: 1
id: "iss-2608240815458212"
slug: "shipped-readme-empty-claim-false"
severity: "minor"
category: "drift"
source: "user-observation"
found_during: "bug-hunt loop round 5 (state issue #368), internal record-consistency hunt + independent adversarial refutation"
found_at: ".abcd/development/intents/shipped/README.md:3"
---

shipped-intents README front-door claims the directory is empty while it holds 18 shipped intents, contradicting the parent intents index which describes it as populated