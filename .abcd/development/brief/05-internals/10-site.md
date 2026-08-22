# The Website — a Rendered Surface of the Record

> **The passages marked below are design targets; the rest describes what
> the binary does.** `abcd site build` renders the landing page, the record
> explorer's pages and `record.json`. The `check` gates and the deploy workflow
> are designs — abcdev.app serves the MkDocs rendering of
> `docs/` at its root, and the deploy is what moves it. Both halves rest on
> [adr-47](../../decisions/adrs/0047-abcdev-app-rendered-from-this-repository-alone.md)
> and [adr-48](../../decisions/adrs/0048-website-deploys-on-release-not-on-merge.md),
> with the [investigation cluster](../../research/abcdev-site/plan.md) and the
> composition rules' executable spec beside them. A shipping change removes
> the mark it lands.

**abcdev.app is a surface of this repository and of nothing else.** `/` is a
landing page for product thinkers, `/docs/` the MkDocs rendering of `docs/`
(SSG-agnostic; replaceable by a later ADR), and `/record/…`,
`/contributors/`, `/references/` a record explorer — every page rendered at
build time from repository text, structured data the repository already
maintains, and committed assets, under adr-47's single-source rule: no text
is written for the website, and the build fails on a text node it cannot
source. The record is **never bundled, rendered read-only** — the site is
the third publication surface adr-47 adds to adr-30's two trees, and the
adr-28 launch boundary is unchanged.

**Plumbing** — per this directory's rule, plumbing lives here and not in
intents:

- **`abcd site build`** walks `.abcd/site.json`, composes the landing page
  from repository text, and emits `record.json`: the record graph with each
  mirrored typed link collapsed once, body mentions deduplicated against it,
  counts by store, lifecycle and status, releases, authorship, the unresolved
  references measured against `.abcd/site-baseline.json`, and the two
  precomputed chart arrangements. It
  inlines committed SVG assets, copies rasters verbatim, and writes into
  `site/` and nowhere else. The graph comes from the record-lint engine's own
  scan (`lint.LoadRecordGraph`) and the dates from one
  `git log --reverse --name-status` pass, so the record has one parser and
  history one read; the frontmatter-free principle store joins the graph from a
  directory read, because there is no frontmatter for the scan to see. The
  render is deterministic — sorted inputs, a fixed
  layout seed, coordinates published at drawing precision, no clock read
  beyond the injected build stamp — which is what lets `record.json` be an
  artifact nobody commits. Transport-agnostic core, front doors per adr-23;
  the composition rules' executable spec is `compose.py`/`build_data.py`
  under [`research/abcdev-site/`](../../research/abcdev-site/), ported rather
  than reinvented. Raster optimisation is a ledger-recorded dependency
  decision; the build never draws.
- **The explorer's pages** are rendered from that one export: `/record/` (stat
  tiles, a state bar per store, release cadence, latest decisions, record
  health, each visual with a table twin), `/record/<type>/<id>/` (frontmatter,
  the body verbatim, typed links phrased from that record's own side, and the
  forge links to the file and to its commit history), `/record/graph/` (the
  chart's stage and its list twin, driven by `site-src/record.js`, reading
  `?focus=<id>`), `/record/timeline/` (the five-lane genealogy as one static SVG
  emitted in Go), `/record/foundations/` (principles and disciplines as cards
  that list and link), `/contributors/` and `/references/`. The bibliography is
  rendered by a stdlib CSL-JSON formatter and numbered identically to
  `ACKNOWLEDGEMENTS.md`, with a build check that the two agree entry for entry.
  A reference whose target has left the tree renders as a dashed stub — on the
  record page, in record health and on the genealogy — never as a dead link and
  never as an arc to an invented position. The Markdown subset carries what the
  record actually writes: nested lists, blockquotes with structure inside them,
  reference links, setext headings, rules, autolinks and CommonMark emphasis;
  anything still outside it is a build failure naming file and line.
- **`abcd site check`** (design target) — the provenance audit over every
  rendered text node, docs-lint's banned tokens over composed text, CLI-
  snippet drift against the generated reference, the
  `.abcd/site-baseline.json` ratchet, and the 390 px static mobile checks.
  The build measures the unresolved references and publishes the count; the
  ratchet that refuses a larger one is this verb's.
- **The generic/specific boundary** of the verb family is governed by the
  itd-140 discipline: repo-agnostic input contract, genericity demonstrated
  on a sparse second instance before it is claimed, working-tier ledger
  publication opt-in only. This repository opts in through
  `.abcd/site.json`'s `record.issue_ledger`.
- **The README migration** — README's product narrative lives in
  `docs/explanation/{rationale,roles,artefacts,process}.md` and
  `docs/how-to/install.md`, and README is a contributor page keeping the
  universal install one-liner, with a test that the one-liner, the
  `install.sh` script and install.md's per-OS forms agree.

The user-facing capabilities ride on this plumbing as intents: itd-135 (the
landing page, umbrella), itd-136 (the record explorer pages), itd-137 (the
relationship chart and genealogy), itd-138 (`install.sh`), itd-139 (the
generic explorer demonstrated on a second instance — the gate adr-47 decision 6
puts in front of any repo-agnostic claim, and the reason none is made here).
Deploy cadence and
trigger are adr-48's: production on `release: published` from the tag,
rendered by the released checksum-verified binary and attested; main to a
labelled preview; emergencies by `workflow_dispatch` from the latest tag,
never from main.
