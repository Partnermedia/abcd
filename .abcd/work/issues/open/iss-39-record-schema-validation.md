---
schema_version: 1
id: "iss-39"
slug: "record-schema-validation"
severity: "major"
category: "drift"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: ".abcd/development/decisions"
---

record schema validation: adr-6 is cited as a live design input but does not exist and no successor records superseding it; adr-12 is not marked superseded by adr-32; itd-47 and itd-49 sit in superseded/ without supersession frontmatter; the superseded/ lint exemption is broad enough to hide content-rule violations. Detector: a mechanical record-schema check — every prose handle (adr-N, itd-N, iss-N) resolves to a file; supersession is bidirectional (superseded-by and supersedes both present); filenames match their patterns; every lifecycle directory is covered by a catch-all so no state escapes linting; the superseded/ exemption narrows to content rules only. Acceptance corpus: the four instances above.