---
schema_version: 1
id: "iss-2608311039586922"
slug: "commands-reading-md-describes-the-bare-verb-s-definitions-fi"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
origin: researcher-authored
production_mode: hand-written
---

commands/reading.md describes the bare verb's 'definitions' field as 'the reading definitions present under agents/'. Since itd-184 that field is what the definition locator RESOLVES, not what is present: a cold-reading-<name>.md naming no closed position is invisible to it, and a definition silent about its position or its regime refuses the whole verb with exit 2 rather than being listed. The plugin surface should say resolved rather than present, and say that a malformed definition is a refusal. The itd-184 builder's lane did not include commands/, so the correction was captured rather than taken.
