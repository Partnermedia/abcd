---
id: spc-35
slug: hook-binary-to-persistent-data-dir
intent: itd-132
---
# hook-binary-to-persistent-data-dir

## Summary

spc-35 delivers itd-132's maintainer-ruled shape: the harness's persistent
per-plugin data directory (`${CLAUDE_PLUGIN_DATA}`) becomes a **download
cache** for the checksum-verified release artefact, plugin roots are
provisioned from it by verified copy, and the PATH entry becomes an
abcd-owned **regular file** instead of a symlink into a harness-owned
directory. Execution stays per-root — the relocate-execution alternative was
rejected in adversarial review for creating a shared mutable binary across
roots. Supersedes spc-21's "update heals by re-fetch" criterion (replacement
contract: an update never re-fetches unless the released binary changed);
every other spc-21 invariant — same-origin checksums, HTTPS pin, `-q`-first
curl, proxy/CA unset, no environment seam for origins, refuse/notice
message discipline — carries over verbatim.

## Design

### 1. Cache layout (in `${CLAUDE_PLUGIN_DATA}`)

```
${CLAUDE_PLUGIN_DATA}/
  cache/abcd-<os>-<arch>     the verified release artefact (0755)
  cache/binary-meta          release_tag / release_sha / binary_sha256 /
                             fetched_at — the spc-21 format minus plugin_sha
  path-entry                 provenance of the abcd-owned PATH copy:
                             installed path + binary_sha256
  .bootstrap.lock            relocated concurrency lock (mkdir-atomic)
  .bootstrap.tmp.$$          relocated temp dirs (same-filesystem rename)
```

`plugin_sha` is dropped from the meta: with one cache serving many roots, a
recorded provisioning-time root is meaningless — the version-skew notice is
computed against the **live** `CLAUDE_PLUGIN_ROOT` basename at render time
(40-hex gate unchanged).

### 2. Provisioning flow (bootstrap.sh rework)

Steady state first, as in spc-21: a regular executable at
`$CLAUDE_PLUGIN_ROOT/abcd` exits in one file test, no network (AC 3).

On a missing root binary (fresh install or fresh post-update root):

1. If `CLAUDE_PLUGIN_DATA` is unset or unusable: degrade **loudly** to the
   current spc-21 per-root fetch — the loud-staging principle; never a
   silent fallback (AC 7). Deriving the documented path shape from
   `CLAUDE_PLUGIN_ROOT` is deliberately not attempted: the derivation is
   documented but not endorsed, and a wrong guess plants a trusted artefact
   in an untracked location.
2. Take the relocated lock in the data dir (per-root locks cannot serialise
   two roots writing one cache). The per-root `.bootstrap.attempt` throttle
   in the hook commands stays as-is — it only rate-limits invocation.
3. **Refresh detector** (maintainer-ruled, local state + one best-effort
   resolve): re-resolve the latest release tag exactly as spc-21 step 4.
   - Resolve answers a tag **equal** to `cache/binary-meta`'s
     `release_tag`, or resolve **fails** (offline): provision from cache —
     re-verify `cache/abcd-<os>-<arch>` against the recorded
     `binary_sha256`, copy into a data-dir temp dir, rename-install into
     the plugin root (AC 1). A cache-hash mismatch refuses loudly and
     installs nothing (AC 5).
   - Resolve answers a **different** tag: download + verify into the cache
     under the unchanged spc-21 steps 5–6, update `cache/binary-meta` by
     rename, then provision the root from the fresh cache (AC 2).
   - No cache and resolve fails: refuse with the spc-21 no-network message.
   The accepted gap: a release cut with no plugin update never triggers —
   the skew notice surfaces it and `abcd update` (itd-130) is the explicit
   refresh path.
4. The hook manifest (`hooks/hooks.json`) keeps invoking `$r/abcd`; only
   bootstrap.sh learns the cache. SessionEnd never reaches this path at all
   (iss-2608210934566223, shipped separately).

### 3. PATH entry: owned regular file (ahoy rework)

- `installPinnedSymlink` becomes `installOwnedCopy`: re-verify the cache
  artefact against `binary_sha256`, copy to the target (default
  `~/.local/bin/abcd`) as a regular file 0755, record path + hash in
  `path-entry` (AC 4, AC 5).
- **Ownership is recorded provenance, never content-guessing**: a regular
  file at the target matching `path-entry`'s hash is owned (idempotent /
  refreshable); anything else classifies foreign and is refused exactly as
  today. A legacy abcd-owned *symlink* (target inside a plugin root or
  cache dir — today's `classifyBinTarget` geometry) is a heal-able gap:
  replaced by the owned copy on `ahoy install` (AC 4).
- Refresh: when bootstrap installs a new release into the cache and
  `path-entry` exists, it refreshes the PATH copy in the same run (verified
  copy + rename); `ahoy install` heals it on demand. `ahoy uninstall`
  removes the owned copy and `path-entry`; the cache itself is left to the
  harness's uninstall-from-all-scopes deletion (open question 2 resolved:
  abcd removes what abcd owns, the harness removes what the harness owns).
- The `--dev` shim is untouched: it builds from the source checkout and
  never writes into the cache or over a verified artefact.

### 4. Migration (one-way, idempotent)

On the first bootstrap run with an empty cache and a verified binary already
in the plugin root: verify it against the root's existing `.binary-meta`
`binary_sha256`; on match, seed the cache (copy + meta rewrite minus
`plugin_sha`); on mismatch or missing meta, ignore it and fetch fresh (AC
5). Root-local `.binary-meta` files stop being written; the skew notice
reads the cache meta.

## Acceptance-criteria mapping

| itd-132 AC | Delivered by |
| --- | --- |
| Update → verified copy from cache, no fetch | Design 2 step 3 (equal-tag / offline branch) |
| New release + update → fetch-verify into cache | Design 2 step 3 (different-tag branch) |
| Steady state: one file test, no network | Design 2 fast path (spc-21 step 1 unchanged) |
| PATH regular file survives updates & GC; legacy symlink healed | Design 3 |
| Every promotion re-verifies binary_sha256; loud refusal | Design 2 step 3, Design 3, Design 4 |
| Skew notice vs live plugin root | Design 1 (plugin_sha dropped; render-time comparison) |
| No data dir → loud per-root fallback | Design 2 step 1 |

## Testing

Each watched fail first, per the spc-21 script-test pattern (fixture server,
temp `CLAUDE_PLUGIN_ROOT` **and** temp `CLAUDE_PLUGIN_DATA`):

- cache hit, resolve offline → root provisioned by copy, zero artefact
  downloads (fixture server records no asset request).
- second plugin root appears (update simulation) → provisioned from cache.
- resolve answers a new tag → one download, cache meta updated, root
  provisioned from the new artefact.
- corrupted cache artefact → loud refusal, nothing installed anywhere.
- `CLAUDE_PLUGIN_DATA` unset → per-root fetch with the loud degradation
  notice.
- migration: verified root binary seeds the cache once; second run no-ops;
  hash-mismatched root binary is ignored and fetched fresh.
- ahoy: owned-copy install; original plugin root deleted → PATH entry still
  executes; legacy owned symlink healed to a copy; foreign regular file
  refused; uninstall removes copy + `path-entry`, leaves the cache.
- skew notice: rendered against a live root differing from `release_sha`;
  silent when `release_sha` unknown (unchanged).

## Filed alongside (SPLIT routing)

- ADR + brief invariant from iss-2608210934566226: persistence must not
  weaken the spc-21 verification posture at fetch time or at rest — every
  promotion re-verifies; SessionEnd performs no network work.
- Principle from iss-2608210934566227: store durable state where the
  platform documents it survives.
- CHANGELOG entry at ship; iss-2608210934566221/222 resolved by the
  shipping change.
