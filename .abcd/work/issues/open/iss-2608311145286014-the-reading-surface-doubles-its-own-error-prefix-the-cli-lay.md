---
schema_version: 1
id: "iss-2608311145286014"
slug: "the-reading-surface-doubles-its-own-error-prefix-the-cli-lay"
severity: "minor"
category: "ux"
source: "agent-finding"
found_during: "itd-184 fidelity audit rcp-426034d44293"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/surface/cli/reading.go"
---

The reading surface doubles its own error prefix: the CLI layer prepends 'reading: ' and the locator's errors already carry it, so an operator sees 'abcd: reading: reading: agents/...' on a refusal. The path this appears on is the definition-refusal path spc-62 introduced as its own proof of wiring, so the first thing the new surface shows a human on failure is a stutter.
