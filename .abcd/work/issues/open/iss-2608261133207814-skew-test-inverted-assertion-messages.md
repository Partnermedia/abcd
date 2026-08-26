---
schema_version: 1
id: "iss-2608261133207814"
slug: "skew-test-inverted-assertion-messages"
severity: "nitpick"
category: "tech-debt"
source: "agent-finding"
found_during: "bughunt-round-8"
found_at: "internal/surface/cli/skew_test.go:97"
---

two session-start skew tests flipped their conditions to nonzero but kept messages reporting got exit 0, and both assertions are now unconditionally true