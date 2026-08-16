---
schema_version: 1
id: "iss-243"
slug: "sota-per-intent-s-template-slot-and-intent-sota-lint-were-ne"
severity: "major"
category: "tech-debt"
source: "agent-finding"
found_during: "intent-planning-prep"
found_at: ".abcd/development/principles/sota-per-intent.md"
---

sota-per-intent's template slot and intent_sota lint were never built: only 9 of 107 intents carry a '## SOTA' section, and 5 of 7 freshly-prepped drafts lack one. The principle's promotion path names the lint; until the detector exists the convention silently under-enforces (fix-the-detector).