# README → docs migration map

Every paragraph below moves **verbatim**; only heading levels and relative links change. After the move, README.md is a contributor-facing page and abcdev.app renders the product narrative from `docs/`.

| Today (README.md) | Destination | Diátaxis type | Notes |
|---|---|---|---|
| Header block: logo, `<h1>`, tagline `<p>`, badges | stays in README.md | — | The tagline `<p>` is the `readme-strapline` surface in `.abcd/positioning.json`; it must remain the first `<p>`. The "Built with" badge stays under its `<!-- docs-lint: allow -->` escape. |
| § Purpose (paragraph + `intro.png`) | `docs/explanation/rationale.md` | Explanation | Page title "Who abcd is for" (new H1; the section was called "Purpose"). |
| Nav line (Roles — Artefacts — Process — Install — Resources) | dropped | — | The site's chapters a·b·c·d are this line. |
| § Roles, §§ Product thinker, Technical facilitator | `docs/explanation/roles.md` | Explanation | Adds one image per role (`role-product-thinker.png`, `role-facilitator.png`), pre-created from the cast illustration. |
| § Artefacts (table + four paragraphs) | `docs/explanation/artefacts.md` | Explanation | Unchanged. |
| § Process, §§ Capturing intents, Capturing issues & thoughts | `docs/explanation/process.md` | Explanation | Unchanged, including both code blocks and the Given/When/Then table. |
| § Install, §§ Plugin, CLI, Build | `docs/how-to/install.md` | How-to | Repo-relative links (`commands/`, `hooks/bootstrap.sh`, `.claude-plugin/`) become absolute GitHub URLs so they resolve on the site as well as on GitHub (the docs-lint `links_resolve` rule checks relative links only). § CLI is restructured under H3s: **macOS** and **Linux** each carry the one-liner specialised for that OS (`shasum -a 256` / `sha256sum`, `uname -s` resolved), **Windows** says in one sentence that a build is planned (and that WSL runs the Linux command meanwhile), **Afterwards** holds the PATH, older-install and inspect-before-running paragraphs unchanged. The site's tab row is exactly this structure. |
| The CLI one-liner | README.md § Install (universal form) and `docs/how-to/install.md` §§ macOS, Linux (specialised forms) | — | Three commands that must agree. A test asserts the two per-OS forms are the universal one with `uname -s` resolved — same release URLs, same checksum step, same install path — so a change in one fails the build unless all move. |
| § Resources | stays in README.md | — | Contributor-facing. |
| — | `README.md` (new) | — | Header block, the pitch (Identity block), one paragraph pointing at abcdev.app and the docs pages, Install (one-liner + link), Contributing (AGENTS, CONTRIBUTING, SECURITY, ACKNOWLEDGEMENTS), Resources. |

New assets under `docs/assets/img/` (all in this bundle): `role-product-thinker.png`, `role-facilitator.png`, `role-ai-agents.png` (203×203, cropped from the cast illustration; roles.md shows one per role, artefacts.md's table reuses the first two above its column labels); `artefact-brief.svg`, `artefact-intents.svg`, `artefact-audits.svg` (64×64 icons, one on the line before each lead-in paragraph in artefacts.md); `process-loop.svg` (the loop figure, referenced after the first paragraph of process.md, portraits embedded, every label a phrase on the page); plus a 1200×630 crop of `intro.png` for the OpenGraph card (not yet made). The SVGs use `var(--token, fallback)` colours: inlined by the site they follow the theme, on GitHub and in MkDocs the fallbacks render. Image assets are pre-created and committed; the site build only optimises rasters and inlines SVGs, it never draws new pictures.

Index pages updated: `docs/explanation/README.md` and `docs/how-to/README.md` list the new pages.

## What changes for tooling

- `docs-lint` now gates the moved text under `docs/` exactly as it gated it in README (same rules, same roots).
- `.abcd/positioning.json` keeps `readme-strapline`; the site adds a `site-hero` surface rendered from the same Identity block.
- Tests: nothing in `internal/` reads README's prose (only fixtures named README.md), so the slimmer README breaks no test; `positioning` still finds its `<p>`.
- `mkdocs.yml` needs no `nav:` change (auto-nav); the new pages appear under Explanation and How-to.
