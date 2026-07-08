---
schema_version: 1
id: "iss-28"
slug: "hermetic-git-test-env"
severity: "minor"
category: "future-work-seed"
source: "agent-finding"
found_during: "external agent finding, recorded 2026-07-08"
---

scaffold a hermetic git environment for tests that shell out to git: any test invoking git as a subprocess can read the developer's real ~/.gitconfig (identity, aliases, includeIf, hooks path) — non-determinism and identity leakage into fixtures — or, via one un-scoped call (missing -C/cwd) or corrupted repo state, mutate the ambient repo's config, refs, or history. Per-call temp-repo scoping is defence-in-breadth only; the robust fix also isolates the git environment. Ship a shared hermetic-git helper in the generated test scaffolding, per target language (Go gitIsolatedEnv(), pytest fixture, shell shim), that pins HOME and XDG_CONFIG_HOME to a per-test temp dir, sets GIT_CONFIG_GLOBAL/GIT_CONFIG_SYSTEM to /dev/null (or GIT_CONFIG_NOSYSTEM=1), supplies identity via GIT_AUTHOR_*/GIT_COMMITTER_* env rather than git config, sets GIT_TERMINAL_PROMPT=0 and core.hooksPath=/dev/null, and is always combined with scoping to the temp repo — documented in the scaffolded repo's testing conventions/AGENTS.md so it is the path of least resistance. Optional enforcement: a lint/pre-commit check flagging git subprocesses in test code that bypass the helper, turning convention into guarantee. Exactly the cross-repo hygiene the scaffolder exists to standardise; removes a whole class of flaky-test and repo-pollution bugs.