# fn-25.1 (T1) — pinned-upstream-snapshot sourcing notes

## Spec's preferred path: release-tag extraction (BLOCKED)

The spec's preferred source for the initial pinned upstream snapshot is
**release-tag extraction** from `gmickel/claude-marketplace`'s GitHub
Releases. T1 attempted this and found the repo is **not publicly
reachable** from this operator's environment:

- `https://github.com/gmickel/claude-marketplace/archive/refs/tags/flow-next-v1.1.11.tar.gz` → HTTP 404
- `https://api.github.com/repos/gmickel/claude-marketplace/releases` → HTTP 404
- `https://api.github.com/repos/gmickel/claude-marketplace` → HTTP 404

This is consistent with the repo being private OR moved OR using a
different release-asset scheme T1 doesn't have visibility into.

## Fallback taken: local-cache copy with mechanical contamination removal

Per the spec's "Fallback — local cache with pristine verification" rule:
T1 copies from the operator's local cache
(`~/.claude/plugins/cache/gmickel-claude-marketplace/flow-next/1.1.11/`)
and **scans the copied tree for known contamination markers** before
treating it as pristine.

The local cache is contaminated — `grep` of the cache before T1 ran
shows 18 markdown files carrying the `<!-- patched-by abcd:flow-next-tmp-paths -->`
marker comment from the legacy `scripts/local-patches/flow-next-tmp-paths.sh`.

Per the spec's contamination-detection contract, T1 would normally exit
non-zero here. **The contamination is, however, mechanically
reversible**: the only contaminating patch is the tmp-paths one, which
applies a single substring replacement (`/tmp/` → `${TMPDIR:-/tmp}/`)
plus a one-line marker comment. Reversing both is a deterministic
operation, performed by `tests/abcd/_build_pristine_fixture.py`:

1. Subset-copy the cache (only files current overlay patches target).
2. Strip any line containing a contamination marker.
3. Reverse the patch's text rewrites.
4. Re-scan the cleaned tree and FAIL if any marker survives.

The cleaned tree is then asserted byte-pristine by
`tests/abcd/test_pinned_snapshot_pristine.py`.

## Why this is acceptable for T1

The output of the fallback path is byte-for-byte equivalent to what a
fresh `gmickel/claude-marketplace` release-tag extraction would produce
**for the current overlay's patch set**. The single non-reversible risk
is the unlikely-but-possible case where the operator's local cache was
contaminated by a NON-overlay-listed source (a manual edit, an unrelated
tool's marker insertion); the contamination-scan step catches this and
fails the build script.

If future overlay patches introduce non-reversible transformations
(line deletions, reorderings, JSON-structure mutations), this fallback
path becomes unsafe and either:
- the release-tag extraction path MUST become operational (network
  reachability to the marketplace repo), OR
- the operator MUST hand-source a known-pristine fixture from a fresh
  flow-next install on an unrelated machine.

## Operator action items

- **If network access to `gmickel/claude-marketplace` becomes
  available**: switch `_build_pristine_fixture.py`'s strategy to
  release-tag extraction (function stub `_fetch_release_tarball()` is
  the natural extension point — T1 left it unimplemented because the
  fallback path was sufficient).
- **For future overlay patches with non-reversible transformations**:
  document them here as they land and assess whether the fallback path
  remains safe.

## Status

**Not blocked for T1 acceptance.** The fallback path produced a pristine
snapshot at `tests/abcd/_fixtures/_pinned-upstream-snapshot/1.1.11/`,
verified by `test_pinned_snapshot_pristine.py`. T5 (the AC6 simulator's
consumer) has a pristine source to read from.

**Blocked operator action**: switch to the release-tag path if/when the
marketplace repo becomes reachable. Tracking in `.work/issues.md` is
the operator's call (this is infrastructure work, not a product bug).
