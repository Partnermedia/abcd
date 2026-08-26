---
schema_version: 1
id: "iss-2608261133207814"
slug: "skew-test-inverted-assertion-messages"
severity: "nitpick"
category: "tech-debt"
source: "agent-finding"
found_during: "bughunt-round-8"
found_at: "internal/surface/cli/skew_test.go:97"
resolution: "both tripwires report the actual exit code with the sibling regression-tripwire framing; the live-[redacted-user] intent moves to the stderr assertion that carries it"
impact: internal
resolved_by:
  commit: "221ea5d1"
---

two session-start skew tests flipped their conditions to nonzero but kept messages reporting got exit 0, and both assertions are now unconditionally true