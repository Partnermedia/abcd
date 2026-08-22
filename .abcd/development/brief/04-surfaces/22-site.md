# `/abcd:site` — The Website as a Rendered Surface

`/abcd:site` renders abcdev.app from this repository and from nothing else
([adr-47](../../decisions/adrs/0047-abcdev-app-rendered-from-this-repository-alone.md)).
The bare form is **strictly read-only**: it reports what the repository has
declared and what the output directory holds. `build` is the render, and it
writes only inside the directory it is given.

It answers a different question from `/abcd:launch`: `launch` prepares what a
release ships to users who install the binary; `site` prepares what a reader
sees who never installs anything. The cadence that connects them is
[adr-48](../../decisions/adrs/0048-website-deploys-on-release-not-on-merge.md)'s:
production renders from the tag, with the released bytes.

## Sub-verbs

> _Machine-checked (`surface_coverage`, spc-27): each row records the verb's
> adr-40 bucket (`lint` / `review` / `audit` / `gate`, or `—` for a
> non-assessment verb) and its existence (`shipped` / `staged`), verified
> against the committed command-tree snapshot in both directions._

| Verb | Bucket | Status |
|---|---|---|
| `build` | — | shipped |

## The single-source rule

No text is written for the website. Every sentence it renders is a span of a
repository file, selected by path and heading through `.abcd/site.json`; the
only words the generator may add are the interface strings in
`site-src/ui.json`, plus numbers, dates, file names and asset names. A sentence
that would improve the site is written into `docs/` or the record, where it must
also read true on the forge.

Two mechanisms keep that from being a promise nobody can check. Every rendered
block carries `data-src="path#heading"`, so each one names its own source in the
markup. And `site-src/ui.json` is decoded against a closed struct with unknown
fields refused, so a word added to that file which no field reads fails the
build rather than reaching a reader unreviewed.

Every picture is a committed asset under `docs/assets/img/`, referenced from a
docs page like any other image. SVGs are inlined so their `var(--token,
fallback)` colours follow the reader's theme; rasters are copied verbatim. The
build never draws.

## Behaviour

```bash
abcd site                    # what is declared, and what the last build left; exit 0
abcd site build              # render into ./site
abcd site build --out DIR    # render into DIR
```

`build` reads the composition manifest, the interface-string allowlist, the
record (through the record-lint engine's own scan — one parser, not a second),
one `git log --reverse --name-status` pass, and `CHANGELOG.md`. It writes
`index.html`, `record.json`, the `_redirects` and `_headers` maps, the
stylesheet and script, and every referenced raster. Nothing else, nowhere else,
and no network at any point.

The render is **deterministic**: sorted inputs, a fixed layout seed, coordinates
published at the precision the chart draws them, and no clock read beyond the
build stamp the caller injects. Two builds of one tree are byte-identical, which
is what lets `record.json` be a build artifact nobody commits.

Three things are derived rather than decided. The featured quotation is the
newest shipped intent whose audit rollup records met criteria and none unmet,
dated by the day its file entered `shipped/`. The Beta badge renders while the
newest dated changelog version's major is 0. The footer's version and commit are
the build stamp.

Graceful absence throughout: no `CHANGELOG.md` omits the release badge, the
release pill and the releases list, and the build succeeds; no Identity block
omits the hero's eyebrow, tagline and pitch, and the headline and lede that
carry the page stay.

## The markdown subset

The renderer carries ATX headings, paragraphs, bold/italic/code spans, links,
images, fenced code, pipe tables, blockquotes and flat lists. Anything else is a
build error naming file and line. Passing an unknown construct through
unrendered publishes raw markdown to readers; dropping it publishes a hole. A
build that stops and says which line is the only outcome anybody can act on.

## `record.json`

The record graph as one file: nodes with their lifecycle, title, dates and
degree; typed links with each mirrored pair collapsed once (an intent's
`spec_id` and its spec's `intent` are one link); body mentions deduplicated
against those links; counts by store and lifecycle; releases; authorship and the
`Assisted-by:` tallies; the unresolved references measured against the committed
baseline; and both precomputed chart arrangements. It is a build artifact and is
never committed.

## References

- Plugin command: [`commands/site.md`](../../../../commands/site.md)
- Decisions: [`adr-47`](../../decisions/adrs/0047-abcdev-app-rendered-from-this-repository-alone.md), [`adr-48`](../../decisions/adrs/0048-website-deploys-on-release-not-on-merge.md)
- Internals: [`05-internals/10-site.md`](../05-internals/10-site.md)
- Composition rules: [`research/abcdev-site/`](../../research/abcdev-site/)
- Release surface: [`04-launch.md`](04-launch.md)
