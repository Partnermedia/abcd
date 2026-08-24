---
schema_version: 1
id: "iss-100"
slug: "conventions-naming-glossary-internals"
severity: "minor"
category: "tech-debt"
source: "agent-finding"
found_during: "M2 cross-repo gate probe"
found_at: "internal/core/lifeboat/sources_conventions.go"
resolution: "glossary, naming and internals conventions adapters built and wired into disembark (itd-96)"
impact: additive
resolved_by:
  intent: "itd-96"
---

disembark has no conventions-tier adapters for constraints/naming, glossary, or internals, so these ground only from an authored abcd brief. Add adapters reading GLOSSARY.md / docs/glossary (naming, glossary) and docs/architecture / package layout (internals) so a repo with conventional docs but no abcd record still grounds them.