# The Website — a Rendered Surface of the Record

> **Everything on this page is a design target, not yet shipped.** The site
> exists today as a design and a plan — the
> [investigation cluster](../../research/abcdev-site/plan.md), a clickable
> prototype, and a reference copy of the README-migration bundle — ratified
> by [adr-47](../../decisions/adrs/0047-abcdev-app-rendered-from-this-repository-alone.md)
> and [adr-48](../../decisions/adrs/0048-website-deploys-on-release-not-on-merge.md).
> The generator is not yet real; abcdev.app still serves the MkDocs
> rendering of `docs/` at the root. The shipping change removes the marks
> on this page.

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

**Plumbing** (design target; per this directory's rule, plumbing lives here
and not in intents):

- **The `site` verb family** — `abcd site build` (walk `.abcd/site.json`,
  compose the landing page and record pages from repo text, emit
  `record.json` with the deduplicated typed links and the two precomputed
  chart arrangements, inline SVG assets, optimise rasters, write `site/`)
  and `abcd site check` (the provenance audit, docs-lint over rendered text,
  CLI-snippet drift, the `.abcd/site-baseline.json` ratchet, the 390 px
  mobile audit). Transport-agnostic core, front doors per adr-23; the
  composition rules' executable spec is `compose.py`/`build_data.py` in the
  working-tier migration bundle (see
  [`research/abcdev-site/plan.md`](../../research/abcdev-site/plan.md)). The generic/specific boundary of the verb family
  is governed by the itd-140 discipline: repo-agnostic input contract,
  genericity demonstrated on a sparse second instance before it is claimed,
  working-tier ledger publication opt-in only.
- **The README migration** — README's product narrative moves verbatim to
  `docs/explanation/{rationale,roles,artefacts,process}.md` and
  `docs/how-to/install.md`; README becomes a contributor page keeping the
  universal install one-liner, with a test that the one-liner, the
  `install.sh` script and install.md's per-OS forms agree. The migration
  bundle sits unpacked in the shared working tier
  (`.abcd/work/site-plan/readme-migration/`) until Phase 1 of the build
  moves the files to their real paths and deletes the working copy.

The user-facing capabilities ride on this plumbing as intents: itd-135 (the
landing page, umbrella), itd-136 (the record explorer pages), itd-137 (the
relationship chart and genealogy), itd-138 (`install.sh`), itd-139 (the
generic explorer demonstrated on a second instance). Deploy cadence and
trigger are adr-48's: production on `release: published` from the tag,
rendered by the released checksum-verified binary and attested; main to a
labelled preview; emergencies by `workflow_dispatch` from the latest tag,
never from main.
