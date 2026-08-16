---
schema_version: 1
id: "iss-236"
slug: "launch-dry-run-hard-fails-on-a-local-username-false-positive"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "intent-planning-prep"
found_at: "internal/core/launch"
---

launch --dry-run hard-fails on a local_username false positive: the scan matches the maintainer's short username against the documented '--dev' CLI flag text, so the release preview reports 5 hard fails on a clean tree. The matcher needs a word-boundary/flag-context exemption.