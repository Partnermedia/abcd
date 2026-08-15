---
id: itd-111
slug: a-stale-abcd-never-answers-silently-every-surface-that-runs
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: [itd-108]
severity: minor
impact: additive
---

# A Stale abcd Never Answers Silently

## Press Release

> **Every abcd surface now knows its own vintage and says so.** The binary
> carries its build revision; at session start the plugin compares it against
> what should be running — the source tip in a dogfood checkout, the
> plugin manifest's pinned version everywhere else — and a mismatch is
> announced with the one-command fix, never discovered a session later
> through misbehaviour. `abcd version` and `abcd ahoy` always show install
> mode and vintage. And the one verb where stale logic writes state refuses
> outright: `ahoy install` run through a binary older than its own source
> declines to touch the machine until the binary is rebuilt. abcd never asks
> the network "what's new?" on its own: implicit checks read only what is on
> disk, the network answers only an explicit `--check`, and when a plugin
> update names a new binary version, provisioning fetches exactly that
> pinned, checksum-verified artifact — completing an update the user already
> chose, not phoning home.
>
> "A month-stale binary spent a morning confidently applying month-stale
> install logic at my machine," said Alice, a maintainer. "Now the session
> opens by telling me the binary is behind the tip, and the install verb
> won't even run through it."

## Why This Matters

The 2026-08-15 install session is the evidence (iss-228): the repo-root
plugin binary predated the no-sudo install work, so it targeted root-owned
`/usr/local/bin`, rejected the `--bin-dir` flag its own skill documentation
described (iss-225/226 — the docs were current, the binary was not), and sent
a whole session into detective work before `make build` revealed the trivial
root cause. Nothing in the system knew — or could say — that the binary was a
month behind the surface fronting it. The failure class is invisible by
construction: a stale binary behaves plausibly, just wrongly, and the newer
the documentation the more misleading the combination. iss-227 (error paths
that report partial with no note) compounds it: silence on top of silence.

Detection is cheap and already half-present — Go stamps the VCS revision into
the binary, the checkout knows its tip, the plugin cache knows its pinned
version. What is missing is only the comparison and the refusal to stay quiet.

## Design Decisions (grilled 2026-08-15)

1. **Separate intent, detection-only.** itd-108 ships a distribution channel;
   this ships a standing invariant across both channels. itd-111 never
   writes and never heals — rebuild is the dev shim's job, re-download is
   itd-105/108's provisioning job. It detects and refuses to be silent.
2. **The network trichotomy.** Implicit checks are disk-only (embedded
   revision vs checkout tip; binary version vs plugin-cache manifest, which
   the harness's own update cycle refreshes). The network answers only an
   explicit user-invoked `--check`. Provisioning after a plugin update may
   fetch — but only the manifest-named, checksum-verified version: no
   version-discovery request exists anywhere in abcd.
3. **Warning surfaces.** SessionStart notice (the existing gap-notice
   channel) names the stale binary, its vintage, and the one-command fix.
   `abcd version` and `abcd ahoy` always print install mode + vintage.
   Ordinary verbs stay silent — with one deliberate exception: `ahoy
   install` run through a stale-against-tip binary refuses with the mismatch
   named and writes nothing, because stale install logic mutating the
   machine is the trap the evidence session actually fell into.
4. **Doc/binary version-match is a corollary, not a mechanism.** The plugin
   ships docs and binary as one unit, so drift exists only where vintage
   drift exists (dev checkouts, failed provisioning); the vintage warning is
   the doc-drift warning. No separate doc-version machinery.
5. **Platform parity is verified, not assumed.** macOS and Linux share the
   same paths and semantics by design; the claim becomes a machine-class
   criterion in the itd-109 calibration set (fresh Linux box on the (b)
   page), not an assumption in prose.

## SOTA

Anchors: the `update-notifier` pattern (npm ecosystem), Homebrew's
auto-update-on-use, and Go's embedded build info (`runtime/debug.BuildInfo`
VCS stamping). **Declared path: 2 — native floor.** The anchors' implicit
network checks fail the fit-challenge outright (a privacy-positioned tool
making background HTTP calls is the posture this record scores against
elsewhere); what survives is their UX grammar — cached comparison, gentle
nudge, one-command fix — implemented over disk-only sources. The seam: the
vintage comparison is one documented function over (embedded revision,
checkout tip, manifest version); any future channel or harness reuses it.
The independent fit-challenge of this declaration runs at plan time.

## Acceptance Criteria

- Given a dogfood checkout whose plugin-root binary predates the source tip,
  when a session starts, then the SessionStart notice names the binary, its
  embedded revision, the tip it trails, and the one-command rebuild fix.
- Given a binary stale against its checkout tip, when `abcd ahoy install`
  runs through it, then the install refuses before any write, naming the
  vintage mismatch; a fresh binary proceeds normally.
- Given any install mode, when `abcd version` or `abcd ahoy` runs, then the
  output includes install mode and vintage (embedded revision or pinned
  version) and staleness relative to the disk references.
- Given any verb without an explicit check flag, when it runs, then no
  network request for version discovery is made — verifiable in the
  zero-network test harness the citation gate already uses.
- Given the explicit check flag, when the user invokes it, then the latest
  release is fetched once, compared, and reported with its source named.
- Given a plugin update that names a new required binary version, when the
  next session starts, then provisioning fetches that pinned,
  checksum-verified artifact and the session reports the version transition
  it performed.
- Given the same staleness scenarios on macOS and on Linux, when the itd-109
  calibration runs, then observed behaviour matches — recorded as a
  machine-class criterion, human-verified per platform.

## Open Questions

- **Sampled re-surfacing (the harness-survey pattern).** After the first
  SessionStart notice, an unacted staleness should not repeat every session
  (wallpaper) nor vanish; a pseudo-random low-probability re-surface — the
  cadence Claude Code's own session-quality survey uses — may be the right
  anti-wallpaper mechanism. Whether a general one-tap micro-prompt channel
  (press 1/2/3) is worth building is bigger than this intent: it could also
  carry itd-109's part-(b) verdicts and lightweight feedback capture. Park
  for its own capture if the pattern proves wanted.
- Naming and home of the explicit check (`abcd version --check` vs an ahoy
  flag).
- Harness portability of the SessionStart channel (OpenCoder equivalent —
  ties to the itd-22 lineage and the host-delegated boundary).
- Whether the refusal-on-stale extends beyond `ahoy install` to other
  state-writing verbs (launch scaffold?) or stays deliberately narrow.

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
