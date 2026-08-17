---
schema_version: 1
id: "iss-269"
slug: "the-hook-plane-s-fail-open-fix-covers-sub-verbs-only-not-fla"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
---

The hook plane's fail-open fix covers sub-verbs only, not flags or leaf positionals. iss-267 routes the guard and hook PARENTS through failOpenNoArgs so an unknown sub-verb refuses at exit 1 rather than the host's blocking status 2. Three neighbouring paths still exit 2: an unknown TOP-LEVEL token hits the root's Args validator (root's exit-2 contract is pinned by 04-surfaces/08-abcd.md, lint_surface_test.go and root_dispatch_test.go, so inverting it is a larger change), a stray positional on a leaf such as 'guard hook' exits 2, and ANY unknown flag exits 2 via FlagErrorFunc which failOpenNoArgs does not touch. None is reachable from today's hooks.json, but each becomes reachable the moment the manifest changes shape: if a future hooks.json adds a flag (e.g. 'abcd guard hook --strict') and skews against an older binary, iss-267 returns verbatim -- same brick, same wrapper blind spot -- and TestHooksManifestNamesLiveSubverbs would not catch it, because its regex stops at the first non-subverb token and never inspects flags. Also worth deciding: whether a top-level rename of 'guard' itself (this release already renames audit to lint) should be covered, since a skewed pair would then block every Bash call through the root validator. Options: extend failOpenNoArgs to the leaf and to FlagErrorFunc on the hook plane; extend the manifest test to pin flags too; or rely solely on the manifest-freeze doctrine and document it. Found by the security and ruthless reviews of the iss-267 fix.