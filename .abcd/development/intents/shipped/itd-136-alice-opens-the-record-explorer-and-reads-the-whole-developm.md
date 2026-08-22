---
id: itd-136
slug: alice-opens-the-record-explorer-and-reads-the-whole-developm
spec_id: spc-38
kind: standalone
suggested_kind: null
reclassification_history: []
builds_on: [itd-135]
severity: minor
impact: additive
---

# Alice opens the record explorer and reads the whole development record — every decision, intent, spec and issue as a page, with contributors and references — without reading YAML on GitHub

## Press Release

> **Alice opens the record explorer and reads the whole development record
> without reading YAML on GitHub.** Alice evaluates developer tooling for
> their team, and abcd's strongest evidence — thirty-eight ratified
> decisions, a hundred and thirty press-release intents, a structured issue
> ledger, an audited trail of what shipped — was invisible unless they were
> willing to read frontmatter in a code browser. Now `/record/` opens on a
> dashboard: stat tiles for releases, decisions, intents, specs, issues and
> principles, lifecycle bars, the release cadence, the latest decisions, and
> record health as a committed baseline that can only shrink. Every record in
> the tree is a page of its own — frontmatter, the body rendered from its
> Markdown, its inbound and outbound typed links, and the link out to GitHub.
> `/contributors/` names the humans as authors of record and presents the
> `Assisted-by:` trailer tallies as what they are — disclosure, not
> authorship — next to the policy that requires them. `/references/` renders
> the bibliography the research directory already keeps, numbered exactly as
> ACKNOWLEDGEMENTS.md numbers it. All of it is rendered at build time from
> one export of the record — no API calls, no hand-written summaries. "I
> stopped taking the project's word for it," said Alice. "The record showed
> me what was decided, when, and what it cost — from the same files the tool
> itself works from."

## Why This Matters

The record is abcd's product argument: intent-driven development leaves an
inspectable trail. A rendered explorer makes that argument self-demonstrating
— the binary rendering its own record is the strongest possible proof that
the record is machine-readable — and it arms a second detector: the build
fails on a cross-reference the tree cannot resolve, which turns record drift
from an invisible debt into a visible gate (the ratchet baseline is seeded
with the dangling references known today and can only shrink). The
contributors page makes the attribution convention legible to outsiders: the
human is the author of record; the trailer is disclosure.

## Acceptance Criteria

- Given `/record/`, then the dashboard's counts are derived at build time
  from the tree into `record.json` — a pure build artifact, never
  committed: determinism is asserted by a double-build diff in CI, and the
  published data cannot drift from the tree because production is rendered
  from the tag by the released binary (adr-48) — and every visual has a
  table twin for assistive tech.
- Given any record in the tree (`adr-N`, `itd-N`, `spc-N`, `iss-N`, a
  principle), then `/record/<type>/<id>/` renders its frontmatter, its body,
  its inbound and outbound typed links, and an open-on-GitHub link.
- Given a typed cross-reference whose target is not in the tree and not in
  the committed baseline `.abcd/site-baseline.json`, when the site builds,
  then the build fails naming the reference; given one in the baseline, then
  fixing it shrinks the baseline and a build that grows the baseline fails.
- Given the spec link recorded from both ends (intent `spec_id` ↔ spec
  `implements`) and `related` pairs listed in both files, then the build
  collapses mirrored references so each distinct link renders once.
- Given `/contributors/`, then authors of record come from `git shortlog`
  through `.mailmap`, bot and tool authors sit on a separate labelled row,
  the `Assisted-by:` share and per-model tallies are presented as disclosure
  with `CONTRIBUTING.md` linked, and model names are confined to this page
  under the sanctioned attribution escape.
- Given the repo declares principles (`principles/`) or active disciplines
  (`intents/disciplines/`), then a foundations page lists each as a card
  linking its record page — it lists and links, never explains (context
  belongs in `docs/` and is selected from there); given neither directory
  exists, then the page and its navigation entry are omitted.
- Given `/references/` ships, then the bibliography renders from
  `.abcd/development/research/references.csl.json`, numbered identically to
  `ACKNOWLEDGEMENTS.md`, with DOIs linked and the attribution line the CSL
  style requires; given no renderer compatible with adr-47's no-Node and
  no-committed-HTML constraints exists at build time, then the page and its
  navigation entry are omitted — never a broken or half-rendered page.
- Given a 390 px viewport, then every explorer route renders with no
  horizontal scroll, verified by the static checks in `abcd site check`
  plus the screenshot audit in CI.

## Open Questions

- Whether `/record/<type>/<id>/` pages render full bodies or summaries plus
  GitHub links — full bodies are the plan, and they make the working-tier
  issue ledger's wording more visible than the tree already is on GitHub.
- Whether retired ADR ids get tombstone files or stay baseline entries
  rendered as dashed stubs (the plan's §7; the genealogy in itd-137 hangs on
  the same answer).

## Audit Notes

<!-- abcd-review: OWED receipt=rcp-5a9275d115bd -->
Fidelity review OWED (receipt rcp-5a9275d115bd).
