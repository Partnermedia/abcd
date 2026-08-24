---
schema_version: 1
id: "iss-2608210934566229"
slug: "spc35-migration-mv-onto-directory"
severity: "major"
category: "security"
source: "user-observation"
found_during: "spc-35 adversarial review 2026-08-21"
found_at: "hooks/bootstrap.sh"
resolution: "the migration seed refuses a non-regular file at the cache binary path with the same [-e && ! -f] form as the main install site, guarded both pre-lock and re-checked under the lock"
impact: fix
resolved_by:
  commit: "4ee5726f5032b323f7ca30cf6613dadb214fbc96"
---

spc-35 BLOCK (security review): the migration seed re-opens the mv-onto-directory hazard the main path already refuses. The seed fast-path test is '[ ! -f "$cache_binary" ] || exit 0', which is TRUE for a directory at that path. So 'mkdir -p .../cache/abcd-<os>-<arch>' (one same-UID command) makes the seed 'mv -f' the file INTO the directory (succeeds) and write a lying binary-meta -> every fresh plugin root then downloads ~11MB and hits refuse, so shell commands run UNGUARDED, every session, until a human removes the directory by hand — and this survives every plugin update. Demonstrated (exit 0, no message). Fix: the migration path needs the '[ -e "$cache_binary" ] && [ ! -f "$cache_binary" ] -> refuse' form already used at the main install site (bootstrap.sh:409).