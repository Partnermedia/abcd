---
schema_version: 1
id: "iss-2608211849582190"
slug: "contributing-fence-carveout-overscoped"
severity: "minor"
category: "inconsistency"
source: "user-observation"
found_during: "bughunt-round-6"
found_at: "CONTRIBUTING.md:76"
resolution: "CONTRIBUTING scopes the fence carve-out to the pull-request body"
impact: fix
resolved_by:
  commit: "1395e72"
---

CONTRIBUTING grants the fenced-quotation carve-out for commit messages and pull-request bodies alike, but the gate strips fences on the body arm only, so a fenced footer in a commit message still fails CI