---
schema_version: 1
id: "iss-2608211849468791"
slug: "attribution-crlf-body-false-red"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "bughunt-round-6"
found_at: "scripts/check-attribution.sh:39"
resolution: "check_text normalises CRLF to LF before every rule; corpus gained CRLF cases"
impact: fix
resolved_by:
  commit: "2a2b6b6"
---

attribution gate trailer/None regexes are dollar-anchored over a class excluding CR, so a CRLF pull-request body false-reds the required attribution check