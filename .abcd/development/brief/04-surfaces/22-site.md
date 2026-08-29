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
| `check` | gate | shipped |

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
abcd site build --preview    # stamp the render as unreleased · <commit>
abcd site check --out DIR    # gate an already-rendered tree; exit 1 on findings
```

`--preview` is for a build of an untagged tree: the stamp renders the
`ui.json` word `unreleased` with the commit in place of a version,
`record.json` carries `"preview": true`, and pinning `--version` alongside it
refuses. A build into a non-empty directory purges it only when the tree
carries the `.abcd-site-build` marker a previous build of this repository
wrote and git tracks nothing in it, and refuses loudly otherwise: the build
cannot remove a directory it did not write. An output path with a symlink at
its leaf or at an ancestor inside the repository, the repository root or a
directory containing it, and any directory holding `.git` are refused before
anything is read.

`build` reads, and reads nothing else:

| Input | What it supplies |
|---|---|
| `.abcd/site.json` | the composition: which span of which file becomes which block |
| `site-src/ui.json` | the closed allowlist of words the generator may add |
| `.abcd/record-lint.json` | where the record stores are, so the graph scan finds them |
| `.abcd/site-baseline.json` | the unresolved-reference ratchet the health block counts against (the path is `checks.unresolved_reference_baseline`'s, defaulting to this one) |
| `.abcd/development/**`, `.abcd/work/issues/**` | the record itself, through the record-lint engine's own scan — one parser, not a second |
| git history | one `git log --reverse --name-status --diff-merges=first-parent` pass, plus `shortlog` and the `Assisted-by:` trailers |
| `CHANGELOG.md` | the dated release headings |
| `docs/**` | the composed pages and their committed assets |
| `site-src/{site.css,site.js,record.js,redirects,headers}`, `site-src/install.sh.tmpl` | the static inputs copied into the output tree, and the install script the site serves |
| `CONTRIBUTING.md`, `ACKNOWLEDGEMENTS.md` | the credit sources the contributors and references pages render (the footer also probes `SECURITY.md` and `CITATION.cff` for their existence) |
| `.claude-plugin/plugin.json` | the package's forge URL, licence and author, for links and the footer |

It writes the landing page, the explorer's pages, `record.json`,
`install.sh` (the committed template plus one build-stamp comment), the
`_redirects` and `_headers` maps, the stylesheets and scripts, every
referenced raster, and the `.abcd-site-build` marker. Nothing else, nowhere
else, and no network at any point.

## The gates

`abcd site check` runs seven independent checks over a rendered tree and
reports every failure, not the first: the provenance walk (each visible text
node sits in a resolvable `data-src` span or matches the allowlist of
interface strings, numbers, dates, file and asset names), the hero against
the Identity block, docs-lint's banned tokens over composed text, `abcd …`
snippets against the generated CLI reference, the unresolved-reference
ratchet (growing fails, shrinking is invited), the static mobile checks over
every page, and the loop-figure labels against their page. Scope follows
adr-47 decision 3 exactly: composed surfaces are `/` and every
manifest-selected span, the verbatim record rendering under `/record/` is
exempt, and the attribution escape is a verification — a name on
`/contributors/` must match a trailer or contributor git actually carries.
The rendered-overflow screenshot audit is CI's optional, non-gating job; the
static gates here are what a browserless binary can assert, and the two are
complementary by design.

Two of those inputs are **declared deviations** from itd-140's generic-side
input contract, and are recorded here rather than argued away.
`.claude-plugin/plugin.json` is this repository's package manifest; a repo
without one renders without the forge links and the copyright line rather than
failing. `.abcd/site-baseline.json` is per-repo site configuration, and
`record.json`'s `health` block counts against it — which is the same opt-in
shape as `.abcd/site.json` itself, and the reason it is acceptable: the record
DATA proper stays record-format plus git plus `CHANGELOG.md`, and only the
health measurement consults a configured ratchet.

The rendered `<title>` and `<meta name="description">` carry Identity-block text
without a `data-src` attribute, because neither element can hold visible text a
provenance walk would reach. `abcd site check` special-cases both.

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

The renderer carries what the record actually writes: ATX and setext
headings, paragraphs, CommonMark emphasis and code spans, links (inline,
reference and autolink), images, fenced code, pipe tables, thematic breaks,
nested lists, and blockquotes with structure inside them. Anything else is a
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
