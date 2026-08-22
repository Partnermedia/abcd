---
schema_version: 1
id: "iss-2608221254566264"
slug: "disembark-maxagenttokens-is-documented-in-the-brief-05-inter"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "context-window SOTA investigation"
found_at: ".abcd/development/brief/05-internals/03-configuration.md"
---

disembark.maxAgentTokens is documented in the brief (05-internals/03-configuration.md) as a per-agent context budget with stream+summarise overflow behaviour, but no code reads the key and it is absent from .abcd/config.json — brief-vs-binary drift.