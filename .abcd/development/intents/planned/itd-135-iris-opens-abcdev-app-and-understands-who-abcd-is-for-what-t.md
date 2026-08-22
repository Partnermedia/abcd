---
id: itd-135
slug: iris-opens-abcdev-app-and-understands-who-abcd-is-for-what-t
spec_id: spc-37
kind: standalone
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
impact: additive
---

# Iris opens abcdev.app and understands who abcd is for, what they and their facilitator own, and how to install it — from one page rendered from the repository

## Press Release

> **Iris opens abcdev.app and understands who abcd is for, what they and
> their facilitator own, and how to install it — from one page rendered from
> the repository.** Iris thinks in products, not in repositories. Until now,
> abcdev.app greeted them with a documentation filing system, and the story of
> who abcd serves lived in a README they would only find by already knowing
> where to look. Now the front page is that story: a hero built from the
> rationale page and the repository's canonical Identity block, four chapters
> — Roles, Artefacts, Process, Install — rendered from the four documentation
> pages that carry them, joined by one thread line, and, as the only
> testimonial, the newest shipped intent whose audit verdict is MET, quoted
> verbatim from the record. The install chapter puts the plugin path on the
> left and the CLI group on the right with Iris's own system starred, every
> command a fenced block from the install page. A Beta badge sits by the brand
> for as long as the major version is 0 — a rule on the release version, not
> copy. Not one sentence on the page was written for the website: every span
> is selected from a repository file through `.abcd/site.json`, and the build
> fails on any text it cannot source. "I read one page and knew whose job
> abcd does and whose it protects," said Iris, a product thinker. "And
> nothing on it could be marketing, because nothing on it was written for it."

## Why This Matters

The repository already holds an unusually rich, machine-readable account of
what abcd is — the brief's Identity block, the README's product narrative
moving into `docs/`, a record with audited shipped intents — and none of it
reaches a visitor. A landing page assembled by selection instead of authorship
turns that account into the front door while making marketing drift
structurally impossible: if a sentence would improve the site, it must be
written into the documentation, where it must also read true on GitHub. The
single-source rule and its build gates are recorded in adr-47; the
release-bound deploy that keeps the page describing an installable product is
adr-48. The README→docs migration and the `abcd site build` generator that
this page rides on are plumbing, recorded in the brief rather than filed as
intents.

## Acceptance Criteria

- Given the Identity block under `.abcd/development/brief/01-product/README.md`
  changes, when the site rebuilds, then the hero at `/` renders the new
  tagline and pitch with no template edit — the hero selector in
  `.abcd/site.json` names the block, and `abcd site check` verifies the
  rendered hero against it at build time (the site-hero analogue of the
  `.abcd/positioning.json` surfaces, which check committed files and so
  cannot carry a build-time surface themselves).
- Given any visible text node on `/`, when `abcd site check` runs, then the
  node sits inside an element carrying a `data-src` provenance attribute that
  names a repository file span, or matches an interface string in
  `site-src/ui.json`, a number, a date, a file name or an asset name — and
  the check fails naming any node it cannot source.
- Given the four chapters a–d on `/`, then each is rendered from its
  documentation page (`docs/explanation/roles.md`,
  `docs/explanation/artefacts.md`, `docs/explanation/process.md`,
  `docs/how-to/install.md`) through `.abcd/site.json`, and the only
  testimonial on the page is the newest shipped intent whose audit verdict is
  MET, quoted verbatim.
- Given any `abcd …` snippet on `/`, when the site builds, then the snippet
  matches the generated CLI reference (`docs/reference/cli/commands.md`) or
  the build fails on the stale snippet.
- Given rendered text from any tree, when the site builds, then the docs-lint
  banned-token rules run over it, so record text cannot reintroduce on the
  site what docs-lint keeps out of `docs/`.
- Given the latest release has major version 0, then the Beta badge renders
  beside the brand; given a v1 release, then it is absent — with no copy
  change in between.
- Given a 390 px viewport, when `/` renders, then nothing scrolls
  horizontally and no element is wider than the viewport — the static
  checks (viewport meta, overflow containers on wide elements, max-width
  on images) run in `abcd site check`; the rendered-overflow screenshot
  audit runs as a CI job, since a browserless binary cannot measure
  layout.
- Given every picture on `/`, then it is a committed asset under
  `docs/assets/img/` referenced from a documentation page; inlined SVGs use
  `var(--token, fallback)` colours so they follow the site theme while GitHub
  renders the fallbacks.

## Open Questions

- Which rule picks the featured intent when several shipped intents carry a
  MET audit from the same day (the plan's §7 leaves this to the team).

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._
