# Ideate verdict — record-explorer-generalisation

**Verdict: reframed.** Recorded on 2026-08-22 by abcd's idea-admission protocol —
primary-source research, a grill against the existing record, and an
independent adversarial review. This record exists so the idea is not
re-litigated: it stands whether the idea lived or died.

## The idea

abcd renders a record explorer for any repository it manages

## Leg 1 — Primary-source research

Every load-bearing claim checked against its primary source, never a
secondary citation.

| Claim | Primary source | Finding |
|---|---|---|
| Every input of the record-explorer pages (dashboard, relationship chart, genealogy, contributors, record health) is the abcd record format itself (.abcd/development/** frontmatter and lifecycle directories, .abcd/work/issues/**), git history, or CHANGELOG.md dated headings — created and maintained by abcd's own verbs in any managed repo, so the explorer's input contract is repo-agnostic | .abcd/development/ record layout and docs/reference/cli/commands.md (capture, intent, spec, identity verbs operate per-repo), checked in-session | verified |
| The Identity block is optional per-repo but abcd-managed where present (abcd identity init/render work in any managed repo), so a hero fallback needs graceful absence, not configuration | docs/reference/cli/commands.md — abcd identity init defaults (.abcd/development/IDENTITY.md) | verified |
| A references renderer compatible with the repo's constraints (no Node toolchain; HTML is disposable build output, never committed) exists today for the CSL-JSON bibliography page | 2026-08-21 plan §5.4 (offers citation-js — Node — or pandoc-in-an-Action committing HTML) against .abcd/development/brief/02-constraints/02-dependencies.md and the documentation convention | falsified |
| The landing page's inputs (rationale, roles, artefacts, process docs pages; the cast imagery; the install tab row) exist only in this repository — no generic renderer can produce them, so the landing page must be opt-in composition | 2026-08-21 plan §5.4 content inventory against the docs/ tree | verified |
| abcd already scaffolds rather than owns per-repo automation — abcd launch scaffold writes the release workflows into a managed repo and never publishes — so deployment-stays-with-the-repo has an in-tree precedent | docs/reference/cli/commands.md — abcd launch scaffold | verified |
| A second, sparse managed repo exists somewhere in the rollout to demonstrate the generic claim before it ships | 2026-08-21 plan §6 phased rollout — every phase is abcdev.app; no second instance appears | falsified |

## Leg 2 — Record grill

Does the brief, an intent, an ADR, or a principle already cover,
contradict, or supersede this idea? Every hit is cited by record id, and
every id resolved in this repository when the verdict was recorded.

| Record | Relation | Note |
|---|---|---|
| itd-93 | covered | Scaffolds hardened release workflows into a managed repo, parity-tested against the live workflows so the template cannot drift — the exact seam a scaffolded site-deploy workflow would reuse, answering the scaffold-rot attack |
| itd-106 | covered | abcd sets up the CI a repo requires (builds_on itd-93) — per-repo scaffolding with a drift-report audit mode is an established direction, not a new stance |
| adr-30 | covered | Defines the record format the explorer renders; the explorer is generic exactly to the extent this format is the contract |
| itd-102 | covered | The Identity block re-renders registered surfaces via .abcd/positioning.json in any managed repo — the mechanism a generic hero would extend. The brief and principles carry no per-entry ids: evaluate-at-the-user-surface and enforcement-claims-are-facts are the principles the adversarial leg's decisive attack rests on (a genericity claim needs a demonstrated instance, not construction), and script-first-mvp is satisfied for the composition contract by the prototype's compose.py reference implementation |
| adr-32 | contradicted | The issue ledger is deliberately working-tier data, not record; a zero-configuration explorer that renders it onto a website by default exports a publication decision to repos whose captures were never written for a front page. Resolved by the reframing: the zero-config default renders the durable record only, and ledger publication is an explicit per-repo opt-in in .abcd/site.json |
| itd-9 | covered | abcd stamps schema_version everywhere and ships no migrators (deliberately deferred); a generic renderer breaks the binary-and-record-pinned-to-one-commit story abcdev.app enjoys, so the generic claim inherits itd-9's migration obligation the moment schema_version moves |
| adr-28 | covered | The launch payload's record exclusion is about the release artifact, not visibility; rendering the committed record to a site breaches nothing adr-28 protects |

## Leg 3 — Adversarial review

Conducted fresh-context and off-policy by an evaluator that did not carry
out the research and received the idea as an artefact of unknown
authorship — the evaluator-outside-the-loop principle applied to ideas.

- **partial** — This turns a Go CLI into a static-site generator with a permanent web-UI maintenance surface (templates, CSS, canvas graph, mobile audits, accessibility twins) that a configuration-layer product should not carry
- **partial** — The explorer is forever coupled to a record schema that is still moving, and in the generic case the record-writing binary and the rendering binary are different versions with no migration story (itd-9)
- **partial** — The references page has no renderer the repo's own constraints permit: citation-js brings Node, pandoc-in-an-Action commits HTML
- **partial** — Most managed repos have sparse records, no bibliography, no Identity block — the generic explorer renders embarrassing near-empty pages, and graceful absence (page omitted) is a different, easier property than graceful sparsity (page present but thin), which the idea does not specify
- **survived** — A scaffolded deployment workflow rots in the managed repo the way scaffolded CI rots
- **partial** — The generic claim ships untested: the only planned instance is abcd's own maximal record, written by the format's own designers — 'repo-agnostic from day one' is construction, not demonstration, which evaluate-at-the-user-surface and enforcement-claims-are-facts refuse
- **partial** — Zero configuration renders any managed repo's working-tier issue ledger onto a website by default, publishing text adr-32 deliberately keeps out of the durable record
- **survived** — The idea violates script-first-mvp: a Go verb absorbed before the contract is discovered
- **survived** — The dashboard and per-record pages re-implement primitives with canonical homes (status board, frontmatter parsing per itd-128, record-lint health checks) — the third copy one-canonical-primitive forbids
- **survived** — Rendering .abcd/** to a published site inverts adr-28's single enforcement point keeping the record out of the shipped artifact

## Rejected alternatives

- **Repo-agnostic by construction, claimed from day one with zero demonstration beyond abcd's own site** — abcd's record is the one record written by the format's designers; nothing in the phased rollout exercises sparse, malformed or absent inputs. The reframing makes genericity a demonstrated property: the verb may not describe itself as generic until a second, sparse managed instance (a committed fixture repo at minimum, driven through the same CLI surface in CI with golden-file output) exercises every graceful-absence and graceful-sparsity path
- **Zero configuration renders everything, working-tier issue ledger included** — Exports a publication decision to repos whose captures were never written for a front page, against adr-32's tiering. The zero-config default renders the durable record; ledger publication is an explicit per-repo opt-in in .abcd/site.json, beside the landing-page composition
- **Keep the generator abcd-specific, structured so the explorer could be extracted later** — The input contract genuinely is the record format plus git plus CHANGELOG — the research leg verified no abcd-only input in the explorer pages — so specificity would be a choice, not a constraint; the demonstrated-genericity gate captures the honest middle without foreclosing the adoption story
- **abcd becomes a general-purpose static-site generator for arbitrary repo content** — Out of scope by construction: the explorer renders the record format and nothing else; the landing page is opt-in composition of named docs pages; deployment stays with the repo as at most a scaffolded, parity-tested workflow

## What follows

The idea as posed does not survive, but the reframing recorded above
does. Any graduation to a draft intent carries the reframing, not the
original wording.
