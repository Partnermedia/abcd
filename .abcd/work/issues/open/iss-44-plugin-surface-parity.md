---
schema_version: 1
id: "iss-44"
slug: "plugin-surface-parity"
severity: "major"
category: "drift"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "commands/abcd/ahoy.md"
---

plugin surface parity: commands/abcd/ahoy.md drives only the bare read-only detect while the CLI registers install, uninstall, doctor, and dry-run — doctor and dry-run are read-only so a keep-mutation-away-from-agents policy cannot explain their absence, and the landing commit claimed the sub-verbs were wired through the plugin command; every plugin command remedy for abcd not on PATH says make build, which cannot produce an on-PATH abcd; plain-CLI abcd memory ask output is headed as a plugin slash command. Detector: a surface-parity check — every CLI sub-verb is reachable from the plugin markdown or carries an explicit scoping note, and remedy text is exercised by a test. Acceptance corpus: the three instances above.