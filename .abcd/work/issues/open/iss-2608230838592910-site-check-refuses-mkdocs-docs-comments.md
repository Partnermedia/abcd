---
schema_version: 1
id: "iss-2608230838592910"
slug: "site-check-refuses-mkdocs-docs-comments"
severity: "critical"
category: "bug"
source: "agent-observation"
found_during: "site-triage-verification"
found_at: "internal/core/site/htmlscan.go"
---

site check refuses every mkdocs-built docs page: the scanner faults on HTML comments (mkdocs-material templates and the FontAwesome icon licence header emit them), so docs/** never parsed — before de19186 those pages silently dropped out of every gate (the ship-day green was a false green over docs), and after it every page-walking gate is loudly denied. The deploy chain's site-check step will fail the next release. Candidate fix: scope the no-comment parse fault to generator-authored pages (the rule's own rationale is 'the generator emits none'), so docs pages parse with comments tolerated and the other gates regain coverage of them.