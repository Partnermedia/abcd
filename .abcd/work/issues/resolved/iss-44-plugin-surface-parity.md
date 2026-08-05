---
schema_version: 1
id: "iss-44"
slug: "plugin-surface-parity"
severity: "major"
category: "drift"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "commands/abcd/ahoy.md"
resolution: "All three acceptance-corpus instances close, none pre-fixed by an earlier round. (1) commands/ahoy.md drove only the bare read-only detect while the binary registered install, uninstall, doctor, dry-run and identity-check; it now carries a section per sub-verb on guard.md's pattern, with an explicit scoping note for identity-check, which stays CLI-only because its exit code is the point of wiring it into a hook. (2) Every 'not on PATH' remedy said 'make build', which cross-compiles bin/abcd-<goos>-<arch> and never a PATH-resolvable abcd; each now names the 'go run ./cmd/abcd' fallback the same paragraphs already carried. (3) Plain-CLI 'abcd memory ask' output was headed '# /abcd:memory ask' from internal/core, which is transport-agnostic — the heading, memory lint's heading, and six headroom remedies now name the binary invocation. Detector: internal/surface/cli/surfaceparity_test.go holds every binary sub-verb reachable from its command file or carrying a scoping note, and no remedy offering a build that cannot put abcd on PATH; internal/core/memory/rendernaming_test.go holds core renders to the binary spelling. Both watched failing first."
impact: fix
---

plugin surface parity: commands/abcd/ahoy.md drives only the bare read-only detect while the CLI registers install, uninstall, doctor, and dry-run — doctor and dry-run are read-only so a keep-mutation-away-from-agents policy cannot explain their absence, and the landing commit claimed the sub-verbs were wired through the plugin command; every plugin command remedy for abcd not on PATH says make build, which cannot produce an on-PATH abcd; plain-CLI abcd memory ask output is headed as a plugin slash command. Detector: a surface-parity check — every CLI sub-verb is reachable from the plugin markdown or carries an explicit scoping note, and remedy text is exercised by a test. Acceptance corpus: the three instances above.