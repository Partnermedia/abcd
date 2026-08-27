---
schema_version: 1
id: "iss-2608270908341622"
slug: "ahoy-s-home-redaction-tests-hasprefix-p-home-sep-case-sensit"
severity: "nitpick"
category: "bug"
source: "agent-finding"
found_during: "issue-sweep-2026-08-27"
found_at: "internal/core/ahoy/fsutil.go"
---

ahoy's home redaction tests HasPrefix(p, home+sep) case-sensitively, so a case-variant HOME spelling leaks the identity path in rendered output on a case-folding filesystem