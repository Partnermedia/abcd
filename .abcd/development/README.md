# Development record

The abcd design record — the durable "what / why" the build works from. Kept in the
repo (transparent) but excluded from the release artifact; user-facing docs live
under [`../../docs/`](../../docs/). Organised **flat by artefact type**, one
canonical home per concept:

| Folder | What it holds |
|--------|---------------|
| [`brief/`](brief) | The living canvas: what abcd IS (product … delivery) + the [glossary](brief/glossary). |
| [`intents/`](intents) | Press-release intents — the WHY of each user-facing change. Lifecycle by directory: `disciplines/` `drafts/` `planned/` `shipped/` `superseded/`. |
| [`principles/`](principles) | Distilled cross-cutting design principles (first-class — the lifeboat packs these). |
| [`decisions/`](decisions) | ADRs (MADR) — ratified architecture decisions, one canonical home; plus `notes/`. |
| [`roadmap/`](roadmap) | Sequencing: `phases/` + `rfcs/` (an accepted RFC produces an ADR). |
| [`plans/`](plans) | Dated design / implementation plans (`YYYY-MM-DD-*`). |
| [`research/`](research) | Investigations: `notes/` (dated) + `spikes/` (prototypes). |

Also here: `personas.json` — data for press-release quote attribution (migrates to
embedded Go data when the intent surface is built).

**Conventions.** Durable-vs-working is the `development/` ↔ `../work/` ↔
`../.work.local/` tiering. Issues graduate into `intents/` or `principles/` rather
than a ledger. ADRs use sequential `NNNN` (stable cross-reference handles); plans and
research notes are date-prefixed (chronological). Present tense only; history lives
in git.
