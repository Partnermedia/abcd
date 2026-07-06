---
id: itd-43
slug: epic-to-spec-terminology
spec_id: fn-65-terminology-epic-to-spec-sweep
kind: null
suggested_kind: standalone
reclassification_history: []
related_adrs: []
created: 2026-05-16
updated: 2026-05-16
prd_path: null
---

# abcd Speaks One Word for a Specced Block of Work, and That Word Is "Spec"

## Press Release

> **abcd finishes the rename flow-next started: a specced block of work is a `spec` everywhere a product thinker or contributor reads — no `epic` left behind in a heading, a review type, a schema field, or the glossary.** flow-next 1.1.1 renamed its own `.flow/` data from `epics` to `specs`, and abcd already renamed its `epic_id` intent-frontmatter field to `spec_id`. But abcd's *surfaces* still say `"epic"` in two registers: the reviews subsystem (`epic-review` type, `## Epic:` headings, `epic_id` review-directory identifiers), and prose across the brief, docs, command help, and the canonical `terminology/core/epic.md` term file. This intent makes the vocabulary single. It renames abcd-owned surfaces to `spec`, leaves the vendored flow-next `flowctl.py` untouched (its `epic`/`spec` aliases are an external boundary, not abcd's to edit), and updates the reviews subsystem only after confirming the change stays coherent with the `flow-next:epic-review` type flowctl still emits.
>
> "I'd renamed the field and thought I was done — then a contributor opened `terminology/core/epic.md` and asked which word was real," said Alex, framework author. "abcd's whole pitch is that each concept has one canonical term. Having `spec` in the schema and `epic` in the glossary was exactly the drift the glossary exists to prevent. One sweep, one word, and the term file is the source of truth again."

## Status (fn-48 B6 decision)

This intent is **open** — it is the *remaining* terminology sweep, not a shipped
one. fn-7 shipped only the atomic `epic_id`→`spec_id` intent-frontmatter field
rename (it had to be atomic: schema + data + code together, or intent-lint
fails). The broader, non-atomic surface/prose/glossary sweep enumerated under
*What's In Scope* below was deliberately deferred and parked **in this intent**.
fn-48's intent-lifecycle backfill therefore does NOT move itd-43 to `shipped/`:
the remaining sweep is now delivered by **fn-65** (the intent's `spec_id` points
at `fn-65-terminology-epic-to-spec-sweep`), but the intent stays in `drafts/`
until fn-65 closes — only then does the on-close lifecycle hook move it to
`shipped/`. The earlier "shipped in fn-7" annotation in
`06-delivery/03-out-of-scope.md` was the B6 contradiction this decision
resolves — now corrected to name fn-65 as the delivering spec.

## Why This Matters

abcd's [terminology discipline](../../foundation/terminology/) exists to kill exactly one failure: the same concept named two ways, drifting until two readers mean different things. Right now abcd commits that failure about its own core noun. The schema and all 41 intent files say `spec_id`; `terminology/core/epic.md` still defines the concept as `"epic"`; the reviews subsystem still classifies `epic-review`. A framework that enforces ubiquitous language cannot itself be bilingual about its central term.

The `epic_id`→`spec_id` field rename was done separately and first, on purpose — it had to be atomic (schema + data + code, or intent-lint validation fails). What remains does **not** break anything: it is inconsistency, not breakage, which is why it is its own intent rather than an emergency fix. But unaddressed it erodes the glossary's authority and confuses every new contributor.

Two parts of the remaining surface carry real risk and are the reason this needs grilling, not a `sed`:

1. **The reviews subsystem renames against an external contract.** `verify_reviews.py`'s `VALID_REVIEW_TYPES` and `review_postprocess.py`'s type maps classify reviews against the **flow-next plugin**, which still emits `epic-review`. Renaming abcd's side without coordinating with what flowctl produces can desync review classification.
2. **`flowctl.py` is vendored.** It is the flow-next plugin's code, copied into `scripts/ralph/`. flow-next deliberately keeps `epic`/`spec` as backward-compat aliases. Editing it means the next plugin update silently reverts the work — a maintenance trap.

The single-source-of-truth rule decides the order: `terminology/core/epic.md` is canonical for the concept, so it is renamed to `spec.md` first and everything else conforms to it.

## What's In Scope

- **Rename the canonical term file** `terminology/core/epic.md` → `terminology/core/spec.md`, with `term: spec` and `epic` recorded as a `forbidden_synonyms` entry so the lint catches regressions.
- **Reviews subsystem rename** — `reviews_index.py`, `review_postprocess.py`, `verify_reviews.py`: `epic_id` parameters → `spec_id`, the `--epic` CLI flag, the `## Epic:` rendered heading, the `epic_id` JSON field, and the `epic-review`/`epic` review-type tokens — **conditional on** first confirming coherence with the `flow-next:epic-review` type flowctl emits. Where flowctl's external token cannot change, abcd accepts the flow-next token at its boundary and uses `spec` everywhere internal.
- **`issue.schema.json`** — `related_epics` → `related_specs`, nested `epic` key → `spec`, descriptions updated.
- **`grill-report.schema.json`** — prose mentions of `epic`/task ID updated.
- **Prose sweep** — `intents/README.md`, the brief (`04-surfaces/`, `02-constraints/`, etc.), `docs/reference/{commands,facilitator,review-schema}.md`, `commands/abcd/intent.md`, the grill `SKILL.md` boundary message, project READMEs: `epic` as a noun → `spec`.
- **`.flow/README.md`** — the directory-purpose prose still says `epics`; align to `specs`.

## What's Out of Scope

- **Editing the vendored `flowctl.py`** — its `epic`/`spec` aliases are flow-next's, not abcd's; abcd treats them as an external boundary. A plugin update would revert any edit.
- **The `epic_id`→`spec_id` intent-field rename** — already completed in a prior, separate change (it had to be atomic; this intent is the non-atomic remainder).
- **flow-next's own `.flow/` data** — already migrated by flow-next 1.1.1; not abcd's to touch.
- **Renaming the `fn-` ID prefix** — `fn-N-slug` is flow-next's identifier format; this intent renames the *concept word*, not the ID scheme.
- **Forcing abcd's word onto flow-next's external review-type token** — if `flow-next:epic-review` cannot change, abcd accepts it at the boundary rather than diverging from the plugin.

## Acceptance Criteria

> _BDD format, per the itd-1 discipline._

- **Given** the rename is complete, **when** a contributor greps abcd-owned files (excluding `scripts/ralph/flowctl.py` and any other vendored flow-next code) for `"epic"` as a standalone noun, **then** no live reference remains — only flow-next boundary tokens and historical git-tracked records.
- **Given** the terminology directory, **when** a contributor looks up the concept, **then** it resolves to `terminology/core/spec.md` with `term: spec`, and `epic` appears there only as a `forbidden_synonyms` entry.
- **Given** the reviews subsystem is renamed, **when** a review is classified, **then** classification still succeeds against the review type flow-next's `flowctl.py` emits — the rename did not desync abcd from the plugin.
- **Given** `issue.schema.json` is updated, **when** an issue links to a spec, **then** it uses `related_specs`, and an issue file using the old `related_epics` key fails schema validation.
- **Given** the prose sweep is complete, **when** `intent_lint.py` runs, **then** no forbidden-synonym (`GL002`) violation for `epic` is raised by any abcd-owned intent or doc.

## Open Questions

- Can the `flow-next:epic-review` review type be renamed on flow-next's side, or must abcd accept `epic-review` as a permanent external boundary token? This decides whether the reviews-subsystem rename is full or boundary-only — needs checking against the installed flow-next 1.1.1 and any 1.2 roadmap.
- Sequencing against the `intents/README.md` v1/v2/v3 → phase migration (logged separately in `.work/issues.md`): both rewrite `intents/README.md`. Run the README migration first and this sweep second, or merge them into one README pass?
- Should `terminology/core/epic.md` be renamed (git mv → `spec.md`) or kept as a deprecated stub pointing at `spec.md`? A stub preserves inbound links but adds a file the glossary must explain.
- Does `epic.md`'s definition body need rewriting, or only its `term` field and filename? The concept is unchanged; only the word changes.
- Are there vendored flow-next files beyond `scripts/ralph/flowctl.py` that must also be excluded from the grep-clean acceptance check? (fn-34.8 note: `scripts/ralph/flowctl.py` stays vendored/upstream-UNFORKED — keep excluding it. But `scripts/ralph/flowctl` — the thin WRAPPER, now the abcd-owned `$FLOWCTL` dispatcher — is abcd-owned and should NOT be excluded. abcd's own `rp` verbs moved to the standalone `abcd-rp` / `scripts/abcd/abcd_flowctl_ext.py`, which are abcd-owned and in scope for the grep.)

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._

## References

- Follows: the `epic_id`→`spec_id` intent-field rename (intent.schema.json, prd.schema.json, all 41 intent files, intent_lint.py, commands/abcd/intent.md) — the atomic part, done first; this intent is the non-atomic remainder.
- Sequenced with: the `intents/README.md` v1/v2/v3 → phase migration (logged in `.work/issues.md`, 2026-05-16 session) — both rewrite the same README; order or merge them.
- Triggered by: flow-next 1.1.1, which renamed `.flow/` `epics`→`specs` and the task `epic`/`epic_id` fields, leaving abcd's surfaces inconsistent.
