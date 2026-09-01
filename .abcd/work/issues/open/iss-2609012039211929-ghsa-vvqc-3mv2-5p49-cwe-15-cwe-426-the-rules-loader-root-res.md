---
schema_version: 1
id: "iss-2609012039211929"
slug: "ghsa-vvqc-3mv2-5p49-cwe-15-cwe-426-the-rules-loader-root-res"
severity: "major"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/surface/cli/cli.go"
---

GHSA-vvqc-3mv2-5p49 (CWE-15, CWE-426): the rules-loader root resolver `rulesRoot` (internal/surface/cli/cli.go) walks every ancestor of cwd and takes the first one holding a `.abcd` directory, consulting the git toplevel only as a fallback, so a `.abcd/rules.json` or `.abcd/guard.json` planted above the working tree — a world-readable `/tmp/.abcd`, or the user-scope `~/.abcd` for any session under the home directory — governs the injected rules, the refresh backstop and the shell guard of every session in a repository that has no `.abcd` of its own; a planted `{"disabled":true}` guard.json is a successful load that disarms the guard. Reproduced at v0.7.0 (8f68ffb3): `hook prompt-router` run from a fresh inner git repository injected the ancestor plant, and `guard hook` reported the registry switched off. The fix must bound the resolution at the git working tree — toplevel first, the nearest `.abcd` inside the tree wins, a tree with none resolves to its toplevel, and a non-git cwd resolves to itself with no walk — so an ancestor plant is unreachable from any repository; `banlistRoot`, `LoadBackstop` and both guard loaders share the resolver and are fixed by the same change. Behaviour changes to state out loud: a non-git directory no longer walks to the home directory; a nested repository or submodule resolves its own toplevel. Advisory severity medium.
