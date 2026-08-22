---
schema_version: 1
id: "iss-2608211849466878"
slug: "daysbetween-timezone-dependent-staleness"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "bughunt-round-6"
found_at: "internal/core/lint/citations.go:665"
resolution: "DaysBetween converts both ends to UTC before taking the calendar date"
impact: fix
resolved_by:
  commit: "e4561b5"
---

citation DaysBetween reads local calendar date then restamps UTC, so citation staleness and the release-gate overdue blocker depend on the maintainer timezone