---
id: itd-130
slug: abcd-update-completes-a-chosen-update-in-one-verb-it-fetches
spec_id: spc-32
kind: standalone
suggested_kind: null
reclassification_history: []
builds_on: [itd-105, itd-108]
severity: major
impact: additive
---

# `abcd update` completes a chosen update in one verb

## Press Release

> **Updating abcd is now one verb, and it is always the user's verb.** `abcd
> version --check` has said "update available: v0.6.0 -> v0.6.1" since the
> staleness work landed — and then left the user to find the README one-liner.
> With this intent, `abcd update` completes what `--check` reports: it names
> the release it resolved (or takes an explicit tag), fetches the binary for
> this platform over a pinned transport, verifies it against the same
> release's `checksums.txt` before anything is replaced, swaps the
> PATH-installed copy atomically, and prints a receipt naming the origin, the
> old and new versions, and the digest it verified. On a terminal it shows
> download progress; piped or hooked, it is silent and loud only on failure.
> Nothing about the [[adr-38-implicit-checks-are-disk-only]] posture moves:
> abcd still never looks for updates on its own — the check is a verb the
> user runs, and now the fix is too.

> "The tool told me I was stale and then wished me luck," said Alice, who had
> typed `abcd version --check` on the machine where they first installed the
> one-liner. "Now the next line is `abcd update`, and the receipt tells me
> what it verified before it touched anything." Bob, whose platform team pins
> the developer tooling on every machine they hand out, cared about the
> refusals as much as the swaps: "Eight error messages in the binary say
> 'upgrade abcd'. Now there is a verb they can name — and on a plugin install
> it correctly says 'take a plugin update' instead of touching the plugin
> root." Carol, who reviews what runs on their team's machines, read the
> receipt line: "Fetched from the pinned release origin, verified against the
> release's own checksums, or it refuses. That is the same bar the bootstrap
> already meets — now it is one implementation, not three."

## Why This Matters

[[adr-38-implicit-checks-are-disk-only]] built the grammar this verb
completes: implicit operations never touch the network (tier 1), the network
answers only an explicit ask whose documented meaning IS that fetch (tier 2),
and provisioning completes a chosen update rather than discovering one
(tier 3). `abcd version --check` is tier 2 and already exists
(`internal/core/vintage/release.go` resolves the tag from the unfollowed
`releases/latest` redirect under the urlguard policy). The action half is
missing: the binary tells users to "upgrade abcd" in eight schema-too-new
error sites and gives them no verb to do it.

The provisioning logic exists today in three places at two bars:
`hooks/bootstrap.sh` (hardened: pinned origins, proxy/CA scrubbing,
tag-pinned binary+manifest, same-origin checksum refusal, atomic rename),
the README one-liner (checksummed but hand-run), and — without this intent —
nothing at all for the machine whose `~/.local/bin/abcd` is a real copy that
no plugin update will ever refresh. One canonical primitive says the fourth
implementation must not happen: the fetch/verify/swap core lands once in
`internal/core`, the CLI renders it, and the end-state this verb enables is
`hooks/bootstrap.sh` demoted to the cold-start trampoline — the only state a
Go updater can never serve is the state where no binary exists yet. Windows
is the forcing function: the `.sh` files were always placeholders for
getting mac/linux right, a compiled verb is the only updater that ever runs
on Windows, and the host harness's own updater already demonstrates the
rename-dance this verb needs there.

The install-shape dispatch is the part a naive updater gets wrong, and the
repo record already names each shape: a plugin-root binary is refreshed by
the plugin update itself ([[itd-108]]'s one-cut coherence — an updater that
bumped it independently would recreate the exact surface/binary skew that
intent exists to make structurally impossible, so this verb REFUSES there
and names the host's plugin-update path); an abcd-owned PATH symlink
stranded by a plugin update is healed by `ahoy install` (iss-345); the dev
shim is a chosen install mode, never silently replaced with a release
binary; a real copy at `~/.local/bin` is this verb's home case, owned by
provenance (see Decisions); and a package-manager path is refused with the
manager's own upgrade command (Homebrew recorded-and-parked in the decision
log, 2026-08-20 — the refusal ships with the verb, the tap ships later;
detection is a resolved-path-under-brew-prefix test, never a substring
match).

## Decisions (grilled 2026-08-20)

The maintainer resolved the open questions at the planning interview; these
are commitments, not options.

- **Ownership of a real copy is provenance: digest-matches-a-release.** At
  update time the existing copy is hashed and treated as abcd's when its
  digest appears in a published release's `checksums.txt`. No new state, no
  receipt file; the network use is already granted (the user typed the
  verb). A copy matching no release stays foreign and is refused with the
  occupant described.
- **The binary's home stays the plugin root.** [[itd-108]]'s one-cut
  coherence (fresh root → matching binary) is preserved structurally;
  cold-start once per plugin update is the accepted cost — the 2026-08-20
  experiments showed it cheap and reliable. The persistent-data-dir
  alternative is rejected, not deferred.
- **Verification ships at the checksums bar; in-process attestation is
  deferred.** Same-origin `checksums.txt` (the [[itd-105]] bar) gates the
  swap. In-process build-provenance verification is recorded as a future
  extension, twin to [[itd-108]]'s deferred offline signing key.
- **The swap uses `minio/selfupdate`.** Dependency sign-off given at the
  2026-08-20 interview (the AGENTS.md gate): minimal Apply-on-stream with
  checksum verification and the Windows rename-dance built in; abcd keeps
  its own release resolution and transport. `go get` lands with the
  implementation, not before.
- **The trampoline demotion is a follow-on rung, not this cut.**
  Script-first: the verb ships alone and proves itself. When the script
  delegates, the seam is hardened as specified here: delegate only to a
  binary passing an ownership check, keep the raw fetch path as the
  fallback when delegation fails, and the release-asset layout becomes a
  frozen contract — an old PATH binary provisioning a new plugin root must
  keep working across cuts.
- **No [[adr-38-implicit-checks-are-disk-only]] amendment.** The recorded
  reading: `abcd update <tag>` satisfies tier 3 literally (a named update,
  completed); bare `abcd update` is a tier-2 verb whose documented meaning
  is the fetch, and it names the resolved tag before acting (TTY
  confirmation; in the receipt always). The ADR's text is untouched.

## What's In Scope

- The fetch/verify/swap core in `internal/core` returning a structured
  report; the CLI surface renders it. The origin agreement already stated at
  `internal/core/vintage/release.go` (checker and provisioner resolve
  "latest" identically) extends to the updater.
- An explicit transport policy, specified rather than inherited: the
  updater's client ignores proxy and CA overrides from the environment (the
  seam `hooks/bootstrap.sh` scrubs — `HTTPS_PROXY`, `SSL_CERT_FILE`,
  `SSL_CERT_DIR` — because same-origin checksums are vacuous when one config
  surface supplies both payload and manifest), pins its redirect policy to
  the release origin's own asset hosts, streams the body under a size bound,
  and runs under the urlguard SSRF policy like the checker.
- Verification bar: same-origin `checksums.txt` for the resolved release,
  refusing on mismatch. The swap is atomic and same-directory via
  `minio/selfupdate`; the running binary is never overwritten in place; a
  failed download or verify leaves no partial file.
- Target selection: an explicit tag argument, or bare `abcd update`
  resolving `releases/latest` and naming the resolved tag before acting
  (TTY confirmation; named in the receipt always).
- Install-shape dispatch, keyed on what actually runs (the first PATH
  occupant / the running executable, not one blessed location): plugin-root
  → refuse, name the host's plugin-update path; dev shim → refuse, name the
  shim's contract (it tracks the source tip; `ahoy install` switches modes);
  owned dangling symlink → name `abcd ahoy install` (the iss-345 heal);
  real copy whose digest matches a published release → swap; foreign
  occupant or package-manager path → refuse with the occupant described and
  the right command printed; a shadowed entry is reported as shadowed, never
  as a completed update. Every refusal is loud and names its remedy.
- Progress on a TTY, silence otherwise; the receipt (origin, tag, digest,
  old→new) prints in both modes.

## What's Out of Scope

- Ambient discovery in any form — a background check, a session-start nudge,
  an auto-apply. Forever out, unless [[adr-38-implicit-checks-are-disk-only]]
  is superseded in the open.
- The Homebrew tap itself (parked in the decision log 2026-08-20; this verb
  ships the install-channel refusal that makes the tap safe to add later).
- The plugin surface's delivery — [[itd-108]] owns it; this verb never
  touches a plugin root.
- The trampoline demotion of `hooks/bootstrap.sh` (follow-on rung per the
  Decisions above) and the Windows cold-start shim (its own, later intent;
  this verb's core must merely be portable to it).
- In-process attestation verification (deferred per the Decisions above).

## Scope Conditions

None stated.

## Acceptance Criteria

> _Confirmed by the maintainer at the 2026-08-20 planning interview: all
> nine accepted as written._

- **Given** a pinned `~/.local/bin/abcd` copy at vN and a published release
  vN+1, **when** the user runs `abcd update`, **then** the binary is replaced
  only after its SHA-256 matches the same release's `checksums.txt`, the
  receipt names the origin, tag, digest, and old→new versions, and the exit
  code is 0.
- **Given** a checksum mismatch, **when** the update runs, **then** no file
  is replaced, the refusal names both digests, and the exit code is non-zero.
- **Given** a hostile environment (`HTTPS_PROXY`, `SSL_CERT_FILE`,
  `SSL_CERT_DIR` pointing at attacker-controlled infrastructure), **when**
  the update runs, **then** the fetch either ignores those overrides or
  refuses — it never verifies an attacker's payload against an attacker's
  checksums.
- **Given** the binary resolves into a plugin root, **when** the user runs
  `abcd update`, **then** it refuses and names the host's plugin-update path,
  and the plugin-root binary is untouched.
- **Given** the dev shim occupies the PATH entry, **when** the user runs
  `abcd update`, **then** it refuses, names the shim's track-latest contract,
  and changes nothing.
- **Given** another `abcd` precedes the updated entry on PATH, **when** the
  update completes, **then** the receipt reports the shadowing occupant
  rather than claiming the machine now runs the new version.
- **Given** no network, **when** the update runs, **then** it fails loudly,
  names exactly what was not done, and leaves no partial file anywhere.
- **Given** the zero-network test harness, **when** any verb other than
  `update` (and `version --check`) runs, **then** no release-origin request
  is observed — the ADR-38 seam extended to the new verb.
- **Given** stdout is not a TTY, **when** the update downloads, **then** no
  progress output is emitted; the receipt still prints.

## Prerequisites

- **iss-232 lands first**: `ahoy install`'s staleness refusal keys on
  BuildInfo `vcs.revision`; an updater that ever ships a differently-built
  binary flips that check for every pinned user. The guard asserting the
  shipped binary carries `vcs.revision` precedes the first `abcd update`
  release.
- **iss-345 is the detection-side companion**, staged in the same 2026-08-20
  change-set: the entry a plugin update strands classifies as abcd-owned and
  heals via `ahoy install` instead of reading as a foreign occupant.
- **[[itd-108]]'s central open question is narrowed by the 2026-08-20
  experiments recorded there** (local rig, harness v2.1.237: the archive is
  re-fetched on every plugin update in both the digest-versioned and
  declared-version shapes; the install decision is version-keyed, so a cut
  that bumps the version propagates). The real-cut confirmation itd-108
  still owes at Cut B stands; this intent's plugin-root refusal is designed
  against the confirmed shape and does not depend on that residue.

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._
