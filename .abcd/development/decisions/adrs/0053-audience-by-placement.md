---
id: adr-53
slug: audience-by-placement
status: accepted
date: 2026-08-28
supersedes: null
superseded_by: null
related_intents: [itd-141]
related_rfcs: []
related_adrs: [adr-43]
---

# ADR-53: Audience is expressed by placement, never by a prose register

## Context

The repository has always behaved as if audience is a property of *where*
text lives, not *how* it is written: the three-tier layout keeps the
agent-heavy development record apart from user-facing `docs/`; the "JSON
internal, MD render" invariant assigns machine consumers to machine surfaces
(`--json` payloads, the generated CLI reference); and the rules loader
injects agent guidance as data, not as documentation prose. But no record
ratified that position, and on 2026-08-28 the question became live: a
proposal to extend the writing-style guide with a two-register audience
model — product prose written for humans, technical prose assumed to be
consumed mostly by AI agents.

The prefer-sota sequence ran in full: an independent fit-challenge over the
repo's record, then an independent SOTA research pass, recorded in the dated
research note
[2026-08-28-docs-audience-style-sota](../../research/notes/2026-08-28-docs-audience-style-sota.md).
Both converged. The field's evidence is that well-structured human
documentation serves agents as-is, that the adopted split is by artifact
class (docs for everyone; dedicated instruction files for agents), and that
the strongest dedicated-machine-surface experiment (llms.txt) failed on
adoption data. The fit-challenge found a register axis would duplicate the
placement boundary the tier layout already owns, and would open a
per-page "agent-register" exemption no path-scoped lint rule can see.

## Decision

Audience is expressed by placement and machine surfaces, never by a prose
register:

1. `docs/` is written for human readers, in one register. There are no
   parallel agent-facing versions of documentation pages.
2. The machine audience is served by machine surfaces: `--json` payloads,
   the generated CLI reference, the AGENTS.md router, and the rules loader.
   When a page seems to need an agent-specific rendering, the remedy is a
   machine surface, not restyled prose.
3. A page's density is set by its Diátaxis type, not by an audience guess. A
   page too dense for its reader is mis-typed or mis-placed, never
   re-registered.
4. No rule in the writing-style guide may be conditioned on a page's
   supposed audience; the guide binds `docs/` and `README.md` uniformly.

## Alternatives Considered

1. **Two prose registers inside `docs/`.** Rejected: it bifurcates
   maintenance, contradicts the `docs/` charter ("user-facing only"), has no
   SOTA support, and silently weakens every path-scoped machine-enforced
   rule by inviting register-based exemptions.
2. **A dedicated machine-readable docs surface (llms.txt class).** Rejected
   on adoption evidence (~10% adoption, ~97% zero traffic, no provider
   commitment). Revisit only if a major consumer commits to reading it; the
   repo's own consumers already read AGENTS.md and the rules loader.
3. **Keep the position unratified.** Rejected: the question recurs
   (recurrence-is-signal), and an unrecorded position invites relitigating
   it page by page.

## Consequences

The writing-style guide gains an Audience section citing this ADR, so the
reference page describes a ratified fact rather than an assumption. Future
"write this for agents" requests route to machine surfaces. Reversing this
position is a new ADR that supersedes this one.
