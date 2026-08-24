---
schema_version: 1
id: "iss-2608240806090544"
slug: "planned-readme-overclaims-every-intent-has-a-spec-link"
severity: "minor"
category: "documentation"
source: "agent-finding"
found_during: "bughunt-a/round-5"
found_at: ".abcd/development/intents/planned/README.md"
---

planned/README.md categorically claims every intent carries a spec_id: spc-N link, but 20 of 43 planned intents have spec_id: null — a legal lint-enforced state (intent_lifecycle) per the surface contract, parent intents/README, and adr-34; the subheading silently reintroduces the over-narrow framing adr-34 already corrected in the parent README (resolved iss-3).