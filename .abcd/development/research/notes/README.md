# Research Notes

Free-standing research notes produced during design and review work — the
evidence and synthesis behind decisions, kept distinct from the decisions
themselves.

## What belongs here

- **SOTA / external research** — surveys of prior art, library/tool comparisons,
  best-practice synthesis with cited sources (e.g. `factoring-large-files-sota.md`,
  `spec-kit-vs-flow-next.md`).
- **Review and audit syntheses** — cross-cutting findings that span the codebase
  and don't belong to one spec (e.g. `abcd-lineage.md`, `related-work.md`).
- **Spike / investigation write-ups** — manual scaffolding notes, divergence
  audits, evidence collected while answering a design question
  (e.g. `ahoy-history-store-manual-scaffolding.md`).

## What does NOT belong here

- **Decisions** — go to `.abcd/development/decisions/adrs/` (ADRs). A note here
  may *inform* an ADR; the ADR is where the choice is recorded.
- **Intents** — go to `.abcd/development/intents/`. A note may seed an
  intent; the intent is the forward-looking record.
- **Phase-scoped research** — goes to `research/phase/<N>/` (design inputs a
  given phase consumes).
- **Prompt R&D** — goes to `research/prompting/` (agent prompt drafts, templates).
- **Run logs / ephemeral acceptance output** — go to `.abcd/logbook/`.

## Naming convention

`<topic>-<kind>.md` — kebab-case topic, then a kind suffix that says what the
note *is*: `-sota` (state-of-the-art survey), `-audit` (review/divergence
finding), `-study` / `-notes` (investigation), or a spec/fn tag when the note is
spec-scoped (`fn-34-...`). Notes are not a numbered record-type; they are
folder-convention artefacts promoted into ADRs/intents by hand.

> A subdirectory groups a cluster of notes produced for
> one piece of work.

## Related

- [`../decisions/adrs/`](../../decisions/adrs) — decisions a note may inform
- [`../intents/`](../../intents) — intents a note may seed
- [`../research/`](..) — sibling research directories (`phase/`, `prompting/`, `adr/`)
