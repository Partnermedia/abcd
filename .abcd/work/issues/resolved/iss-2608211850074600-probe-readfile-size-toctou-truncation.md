---
schema_version: 1
id: "iss-2608211850074600"
slug: "probe-readfile-size-toctou-truncation"
severity: "nitpick"
category: "security"
source: "user-observation"
found_during: "bughunt-round-6"
found_at: "internal/core/lifeboat/probe.go:242"
resolution: "probe ReadFile reads cap+1 and refuses an over-cap result"
impact: fix
resolved_by:
  commit: "aa78ef0"
---

lifeboat SourceContext.ReadFile sizes with fstat then reads exactly the cap, so a file that grows past the cap between check and read is silently truncated to a prefix rather than refused