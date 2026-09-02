---
schema_version: 1
id: "iss-2609020352438590"
slug: "both-containment-checks-that-decide-whether-abcd-runs-a-fore"
severity: "minor"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "hooks/hooks.json"
---

Both containment checks that decide whether abcd runs a foreign binary or reads a foreign cache are weaker than the prose around them claims, in two shared ways. (1) The hook shims' PATH rung (hooks/hooks.json, all four PATH-resolving events) compares the candidate binary's directory against the shim's own `pwd -P`, not against the root of the repository the session is working on, so `<repo>/vendor/bin/abcd` is refused from `<repo>` and accepted from `<repo>/sub` — the same hostile clone, a different working directory. (2) Both that rung and dataDirHazard in internal/core/ahoy/data_dir.go judge only the containing directory's mode, so a world-writable file (0777) inside an ordinary 0755 directory passes every check; on a system where the directory's owner is not the only writer of its contents, the binary that is executed is still anyone's to replace. A fix must establish that containment is measured against the repository root the shim is protecting (git rev-parse --show-toplevel, or the harness's project dir), and that the trust test covers the artefact's own mode and ownership, not just its parent's.
