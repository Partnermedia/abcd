---
schema_version: 1
id: "iss-94"
slug: "the-intent-corpus-still-specifies-the-pre-adr-35-lifeboat-mo"
severity: "minor"
category: "drift"
source: "agent-finding"
found_during: "itd-88-m0"
found_at: ".abcd/development/intents"
---

The intent corpus still specifies the pre-adr-35 lifeboat model: itd-2, itd-8, itd-9, itd-10, itd-13, itd-15, itd-19, itd-22 and itd-24 variously use the retired 'disembark to home' signature, the in-tree .abcd/lifeboat/ home, or the in-tree .abcd/development/voyage/ path (itd-9's acceptance writes voyage provenance in-tree — the exact path that would fail abcd's own privacy-hygiene audit rule). The brief, glossary and roadmap were reconciled to adr-35; the intents were deliberately NOT rewritten, because an intent is a proposal with its own lifecycle and silently rewriting nine of them inside an unrelated change is worse than tracking the drift. Each reconciles when it is next planned.