# Ideate verdict — abcdev-site

**Verdict: survives.** Recorded on 2026-08-22 by abcd's idea-admission protocol —
primary-source research, a grill against the existing record, and an
independent adversarial review. This record exists so the idea is not
re-litigated: it stands whether the idea lived or died.

## The idea

abcd generates its own website from the repository

## Leg 1 — Primary-source research

Every load-bearing claim checked against its primary source, never a
secondary citation.

| Claim | Primary source | Finding |
|---|---|---|
| Material for MkDocs entered maintenance mode in November 2025 (critical bug fixes and security updates for at least 12 months, no new features) and the successor path (Zensical, MkDocs 2.0) is unsettled, so the site generator must not depend on MkDocs plugins or hooks and the SSG rendering docs/ must stay replaceable | https://squidfunk.github.io/mkdocs-material/blog/ | verified |
| The Cloudflare Workers Builds image ships Go (1.24.3 default) alongside Python 3.13 and Node 24, so an `abcd site build` step written in Go runs in the existing build pipeline unchanged | https://developers.cloudflare.com/workers/ci-cd/builds/build-image/ | verified |
| GitHub serves a permanent releases/latest/download/< asset> redirect, and abcd's release workflow already publishes version-free asset names (abcd-darwin-arm64, ..., checksums.txt) that fit it, so install.sh needs no redirect service of its own | https://docs.github.com/en/repositories/releasing-projects-on-github/linking-to-releases | verified |
| The unauthenticated GitHub REST rate limit is 60 requests/hour per IP, so version data must be injected at build time and the site must make no runtime API calls | https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api | verified |
| The record's frontmatter already forms a typed cross-reference graph (supersedes, builds_on, spec_id/implements, related_*) sufficient to render a dashboard, relationship chart and genealogy from one build-time export; extracting it surfaces dangling references — adr-22 to adr-14/adr-15/adr-17, adr-25 to adr-8, adr-27 to adr-16, adr-28 to adr-18, adr-35 to adr-4, itd-3 to spc-1 — eight dangling targets where the investigation counted six | .abcd/development/ frontmatter, re-verified in-session by grep over decisions/adrs, intents and specs | verified |

## Leg 2 — Record grill

Does the brief, an intent, an ADR, or a principle already cover,
contradict, or supersede this idea? Every hit is cited by record id, and
every id resolved in this repository when the verdict was recorded.

| Record | Relation | Note |
|---|---|---|
| adr-28 | covered | One repository, curated release, no dev-to-public mirror — the site-in-one-repo half of the architecture extends this stance rather than opening a second tree; a separate abcd-web repo would re-create the mirror adr-28 abolished |
| adr-30 | covered | The record IA separates docs/ (user-facing) from .abcd/development/ (record, never in the launch payload). The site renders record content at /record/ without bundling it, which makes the site a third publication surface — the boundary needs a short ADR amendment, recorded in the new site ADR |
| adr-37 | covered | Changelog-driven releases give every docs fix a version and a date, which is what makes deploy-on-release viable: a docs-only fix is a patch release, not a blocked deploy |
| adr-38 | covered | abcd never fetches implicitly; the site's no-runtime-API rule (everything injected at build time) is the same posture applied to the website |
| itd-102 | covered | The Identity block already re-renders three surfaces via .abcd/positioning.json; the site hero becomes the fourth registered surface — the mechanism exists, the site extends it. The brief and principles carry no per-entry ids: the retire-the-name principle (retired ADRs leave the tree) is what created the dangling supersedes targets the site must render or tombstone, and ratchet-not-big-bang is the shape of the site-baseline ratchet |
| itd-100 | covered | Proof that a shipped intent with a MET audit exists, which the landing page's only-testimonial rule (quote the newest shipped intent with a MET audit) depends on |

## Leg 3 — Adversarial review

Conducted fresh-context and off-policy by an evaluator that did not carry
out the research and received the idea as an artefact of unknown
authorship — the evaluator-outside-the-loop principle applied to ideas.

- **survived** — A website belongs in a second repository: its own toolchain, cadence, contributors and secrets; abcd-web@sha building abcd@tag is a purer reproducibility story; Kubernetes, Go, Rust, GitLab and Atuin all keep their sites in separate repositories
- **survived** — Deploy on merge instead: it is the default of every hosted platform, keeps the record explorer current to the hour, and a release-bound site shows a record up to days stale (1,188 commits in 45 days against 10 releases)
- **partial** — A docs-only fix has to wait for a release
- **partial** — A site build failure could block or taint a release
- **partial** — The site could drift from the binary it documents if built from source rather than from the released bytes
- **partial** — Cloudflare's git integration keeps deploying main regardless of the release-only intent

## Rejected alternatives

- **A second repository (abcd-web) holding the site** — The single-source rule is a property of one tree at one commit: the composition manifest, ui.json allowlist, docs-lint and the record must be reviewable in one PR and checkable by one CI job on one commit. A split re-creates the dev-to-public mirror adr-28 abolished, adds a record-schema/renderer compatibility tax, and the peer group that looks like abcd (GoReleaser, golangci-lint, Task, mise, chezmoi, uv) keeps sites in-repo. The generator is product code: the binary rendering its own record
- **Production deploys on every merge to main** — The site documents an installable product; following main lets it document verbs no downloadable binary has. Production from the release tag, rendered by the released checksum-verified binary and attested, makes the site one verifiable statement; main gets a labelled preview, emergencies redeploy from the latest tag by workflow_dispatch
- **A Node-based SSG (Starlight, VitePress, Docusaurus) for the whole site** — Brings a Node toolchain into a repo that has none, against the single-binary ethos and the new-dependency gate; generation stays in Go, MkDocs Material stays for /docs/ behind an SSG-agnostic boundary replaceable by a later ADR
- **MkDocs plugins/hooks as the generation mechanism** — MkDocs 1.x is unmaintained, Material is in maintenance mode, and MkDocs 2.0 removes plugins; the generator writes plain files so the SSG is replaceable
- **Homebrew tap, /latest redirects, attestation page as additional distribution endpoints** — Each adds a publisher step or trust surface; install.sh on the project domain is the one new endpoint, everything else rides GitHub's permanent latest/download redirects

## What follows

The idea survives the gauntlet and may graduate to a draft intent through
the ordinary quoted-text create (`abcd intent "<text>"`). Ideate mints no
intent itself; graduation stays a deliberate act.
