---
id: itd-132
slug: hook-binary-to-persistent-data-dir
spec_id: spc-35
kind: standalone
suggested_kind: null
reclassification_history: []
builds_on: [itd-105]
severity: major
impact: fix
promoted_from: iss-2608210934566221
---

# The hook binary survives plugin updates

## Press Release

> Updating the abcd plugin is now a non-event. The checksum-verified hook
> binary is no longer a casualty of the harness's plugin-update lifecycle —
> the commit-stamped cache directory that every update replaces and
> garbage-collection later deletes. The verified release artefact is kept
> once in the harness's persistent per-plugin data directory; an update
> copies it into the fresh plugin root instead of re-downloading ~11MB, so
> the first-hook-after-update window in which hooks run without their binary
> shrinks from once per plugin update to once per released binary. And the
> `abcd` command on PATH keeps working across any number of updates: it is a
> regular file abcd owns and refreshes, not a symlink into a directory the
> harness will delete. Each plugin root still runs its own binary, verified
> exactly as before, and every steady-state session pays one file test and
> no network.

## Why This Matters

The plugin cache directory is harness territory: re-cloned on every update,
documented as not preserving extra files, orphaned and garbage-collected
~14 days later. Storing the downloaded binary there meant fighting the
platform's lifecycle — the 2026-08-21 post-mortem (iss-2608210934566221)
showed the full cost in one update: a re-download, a first-hook window with
no binary, a binary copy per plugin version, and a PATH symlink 14 days from
dangling (iss-2608210934566222). The harness provides a persistent
per-plugin data directory (`${CLAUDE_PLUGIN_DATA}`, exported to all hook
processes, scope-agnostic, documented path shape) that survives updates and
is deleted only on full uninstall — the platform-sanctioned home for the
downloaded artefact.

## Decisions

- **Shape (maintainer-ruled at the planning interview, 2026-08-21).** The
  data dir is a **download cache, not an execution home**: bootstrap fetches
  and verifies the release artefact into `${CLAUDE_PLUGIN_DATA}` once, and
  copies a checksum-matched artefact into each plugin root; execution stays
  per-root, preserving per-root version pairing. The PATH entry becomes a
  **copied regular file** owned and refreshed by abcd (`ahoy
  install`/bootstrap), never a symlink into a harness-owned directory. The
  relocate-execution alternative was rejected for creating a shared mutable
  binary across roots (skew semantics, cross-root locking, dev-shim
  collision, unbounded at-rest tampering window — adversarial review,
  2026-08-21).
- **Routing (itd-84 SPLIT, human-confirmed 2026-08-21).** Capability — this
  intent. Trust rule (persistence must not weaken the spc-21 verification
  posture, at fetch time *and at rest*; SessionEnd performs no network
  work) — ADR + brief invariant, drafted alongside the spec; captured as
  iss-2608210934566226. Stance (store durable state where the platform
  documents it survives) — principle candidate; captured as
  iss-2608210934566227. Plumbing — spec detail.
- **Supersession (maintainer-confirmed, 2026-08-21).** itd-132 supersedes
  spc-21's "update into a fresh cache dir heals by re-fetch" acceptance
  criterion; the replacement contract is "an update never re-fetches unless
  the released binary changed". The spc-21 *verification* posture is
  untouched.
- **Refresh detector (maintainer-ruled, 2026-08-21).** Local state only:
  a plugin update (provisioning stamp differs) triggers one re-resolve of
  the latest release tag; fetch only when the tag changed. The known gap —
  a release cut with no plugin update never triggers — is **accepted**: the
  skew notice surfaces it, and `abcd update` (itd-130) remains the explicit
  refresh path.
- **Dev builds stay out of the trusted location.** The `--dev` shim builds
  and runs from the source checkout; a locally built binary is never
  written into the data-dir cache or over a verified artefact, and
  provenance metadata never describes a build it does not match.
- **Impact: `fix`** (maintainer-ruled) — the promise is repairing failure
  modes of the existing bootstrap contract, not new capability.

## Out of Scope

- The SessionEnd hook ordering fix (never bootstrap at exit) —
  iss-2608210934566223, fixed test-first on its own branch
  (`fix/iss-sessionend-never-bootstraps`), independent of this intent. This
  intent shrinks that window's *frequency*; the ordering fix removes the
  download from the exit path entirely, and both stand on their own.
- The missed-transcript recovery sweep (iss-2608210934566224) and the
  statusline surface (iss-2608210934566225) — held in the ledger as seeds.

## Scope Conditions

None stated.

## Acceptance Criteria

- Given a provisioned install with a cached release artefact, When the
  plugin updates into a fresh plugin root (simulated in tests as a second
  plugin root appearing), Then the next hook event provisions that root by
  checksum-verified copy from the cache — no network fetch, no
  missing-binary message.
- Given a plugin update whose latest release differs from the cached
  artefact's recorded release, When the local-state detector fires, Then
  the new artefact is fetched and checksum-verified into the cache under
  the unchanged spc-21 verification posture, and the root is provisioned
  from it.
- Given a steady-state session (root binary present), When any hook fires,
  Then the cost is one file test and no network.
- Given `ahoy install`, When the PATH entry is written, Then it is a
  regular file owned by abcd that survives plugin updates and
  cache-directory deletion (simulated in tests as the original plugin root
  being removed); a legacy abcd-owned symlink into a plugin root is
  detected as a heal-able gap and replaced.
- Given any promotion of a cached or adopted artefact (into a plugin root,
  onto PATH, or into the cache at migration), When the copy is made, Then
  the artefact is re-verified against its recorded `binary_sha256` first; a
  mismatch refuses loudly and installs nothing.
- Given a session on a plugin root other than the one that last fetched,
  When the session-start surface renders, Then the version-skew notice is
  computed against the *live* plugin root, never against a recorded
  provisioning-time root.
- Given a harness that provides no persistent data directory (dev/local
  installs — unverified harness behaviour), When hooks run, Then behaviour
  degrades loudly to the current per-root fetch, never silently.

## Open Questions

1. **Cache-write concurrency** (spec detail): the bootstrap lock and temp
   dirs are per-plugin-root today; concurrent roots fetching into the one
   shared cache need the lock and temp dir relocated into the data dir
   (same-filesystem rename atomicity preserved).
2. **Uninstall scope** (spec detail): `ahoy uninstall` removes the
   abcd-owned PATH file; whether it also empties the data-dir cache or
   leaves that to the harness's uninstall-from-all-scopes deletion — decide
   in the spec with the ADR.

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._
