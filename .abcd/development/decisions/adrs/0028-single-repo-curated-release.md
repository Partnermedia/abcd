---
id: adr-28
slug: single-repo-curated-release
status: accepted
date: 2026-07-06
supersedes: [adr-18]
superseded_by: null
related_intents: []
related_rfcs: []
related_adrs: [adr-19, adr-20]
---

# ADR-28: One repository, a curated release artifact — no dev→public mirror

## Context

abcd shipped through a two-repo mirror: active work in a development repo, a
curated snapshot published to a separate public repo via the launch flow. That
split forced a payload-exclusion policy at the mirror boundary — the launch
payload dropped `.abcd/memory/**` and scoped a restrictive-licence gate to the
lifeboat (ADR-18)
— and it created two trees to keep version-consistent, which
[ADR-19](0019-plugin-json-version-carve-out.md) and
[ADR-20](0020-manifest-version-lockstep.md) exist to police (dev stays
unversioned; the published snapshot carries the version). A dev→public mirror is
a maintenance liability and a drift source: two histories, two identities, a
copy step that can leak or lag.

## Decision

abcd lives in **one repository** and ships a **curated release artifact** cut
from it — there is **no dev→public mirror**.

- **`.abcd/**` stays in-tree** — the design record travels with the code — but
  is **excluded from the release artifact by packaging**. Exclusion is a
  build-time filter over one tree, not a copy between two repos.
- **`launch` cuts a curated GitHub Release** from the single repo. The release
  artifact is the published surface; **the repo is the marketplace**.

This retires the mirror, so this ADR **supersedes**
ADR-18: the
payload-exclusion policy is no longer a mirror-boundary concern but a packaging
filter — `.abcd/**` (memory included) is excluded from the release artifact by
the same mechanism, on one tree.

It **amends** [ADR-19](0019-plugin-json-version-carve-out.md) and
[ADR-20](0020-manifest-version-lockstep.md) rather than superseding them: their
**dev-unversioned / release-versioned polarity survives**, now realised on a
**single tree** — the working tree stays unversioned, and the version is stamped
into the curated release artifact at cut time. The anti-drift intent of ADR-20
still holds; it just no longer spans two repositories.

## Alternatives Considered

- **Keep the dev→public mirror.** Preserves the shipped launch flow. Rejected:
  two repos is a discouraged anti-pattern here — duplicated history, a copy step
  that can leak private record or drift, and two identities to keep consistent —
  for no benefit a packaging filter does not already give.
- **One repo, publish everything (no curation).** Simplest. Rejected: the design
  record and local memory are not part of the shipped product; publishing them
  bloats the artifact and re-opens the exposure ADR-18 guarded against.
- **Chosen: one repo, `.abcd/**` excluded by packaging, `launch` cuts a curated
  GitHub Release.** Single history and identity; curation is a build filter, not
  a repo boundary.

## Consequences

- The launch flow reduces from mirror-and-publish to package-and-release: filter
  `.abcd/**` out of the artifact, stamp the version, cut the GitHub Release.
- The release-artifact exclusion is the single enforcement point for what stays
  private; there is no second repo whose contents can diverge from the source.
- The dev-unversioned / release-versioned invariant
  ([ADR-19](0019-plugin-json-version-carve-out.md),
  [ADR-20](0020-manifest-version-lockstep.md)) now applies within one tree: the
  version lives only in the cut artifact, and the anti-drift check runs over the
  paths that get packaged.
- The repo being the marketplace means discovery, install, and the design record
  share one location; contributors and consumers see the same history.
