# Prior Art — `sync-to-public.sh` (harvested from `scripts/`)

Reference artifact for whoever builds Phase 7 (`/abcd:launch ship`). Not a
living document; a snapshot of a publish mechanism that briefly lived in
`abcdDev/scripts/` and was removed for conflicting with the brief's design.

> **Read alongside:** [`publish-implementation-spec.md`](publish-implementation-spec.md)
> — the canonical Phase 7 hand-off spec — and
> [`../../brief/04-surfaces/04-launch.md`](../../brief/04-surfaces/04-launch.md),
> the brief's launch design.

## What it was

`/abcd:workspace:repair` (workspace bootstrap, 2026-05-15) installed two files
into `abcdDev/scripts/`:

- `sync-to-public.sh` — a 331-line bash publish script, ported and "generalised
  from `idelphiDev/scripts/sync-to-public.sh`".
- `public-manifest.txt` — a hand-maintained allowlist of paths to publish.

They were removed on 2026-05-16. Recover verbatim from git:

```
git show 045a766:scripts/sync-to-public.sh
git show 045a766:scripts/public-manifest.txt
```

## Why it was removed — conflict with the brief

The script implements an *older, different* publish model than the brief's
`/abcd:launch`. The mismatches:

| Aspect | `sync-to-public.sh` | Brief (`04-launch.md`) |
|---|---|---|
| Payload model | **Default-allow** — hand-maintained `public-manifest.txt` allowlist | **Default-deny** — fixed include/exclude lists in § 2 + `.abcd/launch.allow` override |
| Orchestration | Standalone bash | `launch-gatekeeper` agent + `scan.py` |
| PII / secret scan | Hardcoded `grep` pattern list | gitleaks + Presidio + custom regex + optional TruffleHog |
| Pre-flight | Dirty-tree check + grep PII | Full gate suite incl. marker-block, `plugin.json` parse, doc-auditor |
| Mirror modes | Single (clear-then-copy) | `overlay` / `clean` / `branch` (§ 3) |
| Versioning | `--bump beta\|patch\|minor\|major` from public tags | Default patch bump; `--version` override (§ 4) |

`publish-implementation-spec.md` already cites the **idelphi** copy of this
script as "the working precedent" — so the precedent was already recorded.
The `abcdDev/scripts/` copy was a redundant, divergent duplicate that risked
being mistaken for abcd's real publish path. The brief's § 6 bootstrap
exception is an explicit *manual `git push`*, not a sync script.

## What is still worth harvesting

The brief's design supersedes the script's *architecture*, but two pieces of
its mechanics may be reusable when Phase 7 is built:

1. **`--bump` version logic** — parsing the latest public semver tag and
   incrementing `beta`/`patch`/`minor`/`major` (script §§ "Auto-bump
   version"). `04-launch.md` § 4 specifies a default patch bump but does not
   pin down the parse/increment routine; this is a worked reference.
2. **Sibling-repo resolution** — the `resolve_public_root` function
   (flag → `$PUBLIC_ROOT` env → scan siblings with a `.git/` and prompt).
   Useful for locating the public target when no path is configured.

Everything else — the manifest model, the bash orchestration, the grep PII
scan — is superseded; do not carry it forward.
