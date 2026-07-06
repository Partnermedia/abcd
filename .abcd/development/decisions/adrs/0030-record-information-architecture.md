---
id: adr-30
slug: record-information-architecture
status: accepted
date: 2026-07-06
supersedes: null
superseded_by: null
related_intents: []
related_rfcs: []
related_adrs: [adr-3, adr-5]
---

# ADR-30: Design-record information architecture — flat artefact-type folders

## Context

The design record is the durable spec the build works from. It needs one
canonical home per concept, or facts drift across duplicated locations and
cross-references rot. A record that groups by a second axis (input vs process,
durable vs working) inside its own tree double-classifies every artefact and
forces a "where does this go?" judgement on every write.

The working tree already separates concerns on the durability axis: the durable
record in `.abcd/development/`, shared working notes in `.abcd/work/`, and
local-only ephemera in `.abcd/.work.local/`. A second such split inside
`development/` would be redundant.

## Decision

The design record under `.abcd/development/` is organised as **flat folders by
artefact type**, one canonical home per concept, with a `README.md` map:

- `brief/` — the living canvas (what abcd IS) plus the bounded-context `glossary/`.
- `intents/` — press-release intents, **directory-as-status**
  ([ADR-3](0003-directory-as-truth-for-lifecycle.md)):
  `disciplines/ drafts/ planned/ shipped/ superseded/`.
- `principles/` — distilled cross-cutting design principles.
- `decisions/` — ADRs (MADR), one canonical home, plus `notes/`.
- `roadmap/` — sequencing: `phases/` + `rfcs/`.
- `plans/` — dated design/implementation plans.
- `research/` — investigations: `notes/` + `spikes/`.

Two numbering schemes, by artefact nature:

- **ADRs use sequential `NNNN`** (`0007-title.md`) — a stable, order-free
  cross-reference handle that never moves.
- **Plans, research notes, and the Tier-2 `DECISIONS.md` log are date-prefixed**
  (`YYYY-MM-DD-topic.md`) — they are chronological artefacts read newest-first.

Issues are **not** a record folder: a design-significant issue graduates into
`intents/` or `principles/`; trivia lives in local `.work.local/` handover notes.

User-facing documentation is separate, under `docs/`, organised by **Diátaxis**
(tutorials / how-to / reference / explanation), one type per page. The CLI
reference under `docs/reference/cli/` is generated from the command tree and not
hand-edited. The design record does not live in `docs/`; `docs/` does not carry
design rationale.

## Alternatives Considered

- **An input/process (or durable/working) meta-grouping inside `development/`.**
  Rejected: it double-classifies against the `development/` ↔ `work/` ↔
  `.work.local/` tiering that already carries that axis, and makes filing
  ambiguous.
- **Keep the inherited nesting** (intents under `roadmap/`, a deep
  `foundation/terminology/` tree, design notes under `research/adr/`, a
  `research/phase/0/` tier). Rejected: it overloads "phase" against
  `roadmap/phases/`, buries top-level artefacts, and gives ADRs no single home.
- **Date-prefix ADRs too.** Rejected: an ADR's value is a stable handle other
  documents link to; a date prefix makes the handle chronological and unstable.

## Consequences

- Every artefact has exactly one home; cross-references target stable paths.
- ADRs keep a permanent `NNNN` identity even as the record is reorganised; the
  brief stays current-state ([ADR-5](0005-brief-is-current-state.md)) with
  history in git.
- The four Diátaxis directories give user docs a home distinct from the record,
  so a page is either a learning/task/reference/understanding artefact or a
  design record — never a mix.
