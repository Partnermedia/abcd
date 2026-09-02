---
schema_version: 1
id: "iss-2609020259564193"
slug: "the-rules-root-fallback-still-accepts-a-foreign-uid-reposito"
severity: "major"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/rules/root.go"
---

The rules-root fallback still accepts a foreign-uid repository planted in a shared ancestor. ResolveRoot falls back to gitutil.RepoShapedRoot when git will not answer, and now requires that marker to look like a repository (a .git directory carrying HEAD, or a .git file beginning "gitdir: "), which closes the one-command plant but NOT a real repository: 'git init' in a shared temp directory owned by another uid makes git refuse on ownership under abcd's isolated environment (GIT_CONFIG_GLOBAL=/dev/null discards safe.directory), and that refusal is the SAME signal in the attack and in the motivating case a legitimate foreign-uid checkout or container bind mount presents. A session whose working directory is a plain directory beneath such a plant therefore resolves the planted tree as its root and reads that tree's .abcd/rules.json and .abcd/guard.json: injected rules, the loader kill switch and the hazard registry all come from a directory anyone can write. Evidence: ResolveRoot and plausibleRepository in internal/core/rules/root.go. The fix needs an ownership policy decided first (the run takes no design decisions): whether the fallback must refuse a marker root whose owner is not the caller, and what a shared CI checkout, a container bind mount and a multi-user host then do, or whether the git-refused fallback should be narrowed to a root the caller owns and fail closed otherwise. Nothing is changed here; the resolver's doc comment states the residual and names this record.
