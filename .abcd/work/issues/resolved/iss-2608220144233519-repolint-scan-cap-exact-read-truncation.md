---
schema_version: 1
id: "iss-2608220144233519"
slug: "repolint-scan-cap-exact-read-truncation"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "bughunt-b-round-7"
found_at: "internal/core/repolint/rule_privacy.go"
resolution: "readTrackedFile reads cap+1 and routes a grown file to the not-scanned oversize path"
impact: fix
resolved_by:
  commit: "a0a218c"
---

repolint privacy scanner readTrackedFile reads io.LimitReader(f, maxScanBytes) with no cap+1 probe, so a tracked file that grows past the cap between fstat and read is scanned as a truncated prefix and reported clean, bypassing the not-scanned warn channel the same file over-cap at stat would take