---
id: adr-46
slug: persistence-never-weakens-the-verification-posture
status: accepted
date: 2026-08-21
supersedes: null
superseded_by: null
related_intents: [itd-105, itd-132]
related_rfcs: []
related_adrs: [adr-38]
---

# ADR-46: Persisting the hook binary never weakens the verification posture — every promotion re-verifies, and SessionEnd performs no network work

## Context

itd-105 (spc-21) bought the hook binary's provenance dearly: same-origin
checksums.txt pinned to one resolved release, HTTPS-only transport on every
fetch including redirects, `-q`-first curl with the proxy and CA environment
scrubbed before any request, and — after two adversarial-review defeats — no
environment seam of any kind for the fetch origins. Two properties bounded the
residual risk: the artefact lived in the harness's commit-stamped plugin directory,
which every plugin update replaced, so a tampered binary survived at most one
update cycle; and every trusted byte had just come off the verified download
path.

itd-132 (spc-35) moves the artefact into the harness's persistent per-plugin
data directory, where it survives every update and is deleted only on full
uninstall. Persistence is the point — it is what ends the re-download and the
missing-binary window on every update — but it changes the threat shape twice
over: the at-rest tampering window extends from one update cycle to unbounded,
and the artefact now travels again after fetch time, promoted from the cache
into fresh plugin roots, onto PATH, and (at migration) from a plugin root into
the cache. Each promotion is a fresh opportunity for the trusted path to pick
up bytes the verification never saw. The trust rule outlives any one feature
— it binds the bootstrap, `ahoy install`, `abcd update`, and every later verb
that moves the artefact — so per the itd-84 SPLIT routing it lives here rather
than in the intent's prose (iss-2608210934566226).

## Decision

1. **The fetch-time posture carries over verbatim, wherever the download
   lands.** Same-origin checksums pinned to one release, the HTTPS pin on
   every call, `-q`-first curl with the proxy/CA scrub, and no environment
   seam for origins are invariants of fetching the binary, not of the
   directory it lands in. Relocating the artefact's home changes none of them.
2. **Every promotion of a persisted artefact re-verifies against its recorded
   `binary_sha256` and refuses loudly on a mismatch.** Cache → plugin root,
   cache → PATH copy, PATH-copy refresh, and the one-way migration seed
   (plugin root → cache) each re-hash the bytes being moved against the
   recorded provenance before anything lands in a trusted location; a
   mismatch installs nothing and leaves the mismatching artefact in place as
   evidence. Verification at fetch time proves what was downloaded;
   re-verification at promotion time is what keeps an unboundedly-persistent
   file from becoming an unboundedly-stale trust decision.
3. **Ownership is recorded provenance, never content-guessing.** A persisted
   artefact or PATH entry is abcd's to refresh or remove only when a recorded
   hash vouches for its exact bytes; anything that stopped matching is
   foreign and is never replaced, refreshed, or deleted.
4. **SessionEnd performs no network work.** The exit path never bootstraps,
   fetches, or refreshes anything — a download belongs to session start and
   the explicit verbs, never to the moment the user is leaving
   (iss-2608210934566223, shipped separately; recorded here because the cache
   makes provisioning cheap enough to tempt back onto the exit path).

## Alternatives Considered

- **Trust the persistent location because the fetch was verified.** Rejected:
  that inference held (weakly) when the harness deleted the directory every
  update; a persistent directory extends the at-rest window without bound,
  so the recorded hash must be re-checked at every point where the artefact
  is promoted into a place something executes from.
- **Re-verify on every execution (hash in the hook fast path).** Rejected:
  the steady-state contract is one file test and no network (spc-21, kept by
  spc-35); hashing ~11 MB on every hook event buys little — the promotion
  points are where bytes move into trusted locations — and would be paid on
  every prompt.
- **Relocate execution into the data dir instead of copying per root.**
  Rejected in adversarial review (2026-08-21, recorded in itd-132): one
  shared mutable binary across roots creates cross-root skew and locking
  semantics, collides with the dev shim, and widens the at-rest tampering
  blast radius to every session at once.

## Consequences

- spc-35's bootstrap, `ahoy install`'s owned PATH copy, and the migration
  seed each carry an explicit re-verify step; the corrupt-cache refusal is
  pinned by tests at every promotion point.
- `abcd update` remains the one verb that legitimately changes the PATH
  copy's bytes, and it re-stamps the recorded digest only after proving the
  new content against the release's own checksums.
- The corresponding brief invariant is invariant 12 in
  [`02-constraints/03-invariants.md`](../../brief/02-constraints/03-invariants.md),
  landed with this record's acceptance.
- Any future verb that moves the persisted artefact (a repair verb, a
  multi-machine sync) inherits rules 2–3 as requirements, not suggestions.
