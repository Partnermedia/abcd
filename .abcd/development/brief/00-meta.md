# Meta

This file holds **stable conventions** for the brief itself: directory structure, file-naming rules, and the brief↔lifeboat shape contract. The brief reflects the project's *current* state at all times — not versioned, not archived in this directory; see [adr-5](../decisions/adrs/adr-5-brief-is-current-state.md). History lives in `git log`; inflection-point rationale lives in [`../decisions/adrs/`](../decisions/adrs).

## Directory structure

The brief is split across numbered folders rather than a single `README.md`. Reasons:

1. **Concurrent editing** — multiple agents can work on different sections without serialising on one file.
2. **Diff legibility** — `git log brief/04-surfaces/02-disembark.md` tracks the evolution of one command's design, not a whole-brief blob.
3. **Agent context budget** — agents that need only one section can pull just that file (relevant to the [`04-surfaces/02-disembark.md § 3`](04-surfaces/02-disembark.md#3-agent-context-budget) budget rule).
4. **Reusable shape** — the same numbered-folder layout serves as a template for future projects (the lifeboat output shape mirrors this skeleton, see [`04-surfaces/02-disembark.md § 5`](04-surfaces/02-disembark.md#5-output-shape)).

## Naming convention

- **Numeric prefixes** use two-digit zero-padding (`01-`, `02-`, …, `10-`, …) so sort order survives past 9 entries.
- **Hyphens** (not underscores) separate words, matching the kebab-case norm elsewhere in the project (`fn-1-add-oauth`, `itd-N-<slug>`, `adr-N-<slug>`).
- **`README.md`** (the index) deliberately has *no* numeric prefix because it's the entry point, not a section. GitHub renders it automatically when browsing the folder.

## Brief vs lifeboat

The brief skeleton is, deliberately, the same shape as a populated lifeboat (see [`04-surfaces/02-disembark.md § 5`](04-surfaces/02-disembark.md#5-output-shape)). This is a design contract:

- **`/abcd:ahoy`** copies an empty version of this skeleton into a fresh repo (so a human fills it in).
- **`/abcd:disembark`** uses this skeleton's *shape* as a target schema for what lifeboat-composer agents must produce.
- **`/abcd:embark`** reads a lifeboat (which is structurally a populated brief + audit/provenance extras) and writes a new repo's brief by walking the brief↔lifeboat mapping in reverse — the amended press release becomes the new repo's initial brief.

There is one canonical skeleton, used three ways. The mapping table between brief and lifeboat sections is the contract; round-trip tests catch divergence.

## No archive directory

The brief does not maintain an `archive/<NN>/` directory of prior iterations. Per [adr-5](../decisions/adrs/adr-5-brief-is-current-state.md), the live `brief/` directory IS the brief; `git log` covers history; ADRs cover inflection-point rationale. Disembark snapshots (per [adr-4](../decisions/adrs/adr-4-lifeboat-as-regenerable-output.md)) provide the audit-traceable history chain when a forensic snapshot is needed.
