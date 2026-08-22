---
schema_version: 1
id: "iss-2608221126066379"
slug: "frontmatter-bom-tolerance-sibling-parsers-diverge"
severity: "minor"
category: "tech-debt"
source: "agent-finding"
found_during: "bughunt round 7 merge-gate dual review"
found_at: "internal/core/intent/intent.go"
---

frontmatter.Fields trims a leading UTF-8 BOM (iss-2608220134344680) but the sibling parsers that promise byte-exact parity with it do not: intent's setFrontmatterFields (internal/core/intent/intent.go:129, whose comment claims it matches frontmatter.Fields's delimiter tolerance exactly) refuses a BOM-led record the reader accepts, so intent mutations on such a record fail-closed with 'no leading frontmatter block'; changelog's bodyStart (internal/core/changelog/source.go:60, 'so the two never disagree') returns 0 and leaks the whole frontmatter block into the derived changelog body; lint's recordTitle (internal/core/lint/schema.go:554) hand-rolls the comment skip with neither TrimBOM nor the multi-line-comment state. Sweep the siblings onto the shared tolerance or narrow the parity comments to the truth.