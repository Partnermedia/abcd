---
schema_version: 1
id: "iss-2608270908348042"
slug: "five-delimiter-compare-variants-remain-beside-the-canonical"
severity: "minor"
category: "tech-debt"
source: "agent-finding"
found_during: "issue-sweep-2026-08-27"
found_at: "internal/core/frontmatter/frontmatter.go"
---

five delimiter-compare variants remain beside the canonical frontmatter.IsDelimiter: gate-side TrimSpace compares in lint and glossary accept an indented delimiter the canonical rule refuses, intent and changelog carry tolerant local copies, memory keeps its own close predicate, and site tests a bare HasPrefix — one consolidation pass onto the canonical predicate closes the family