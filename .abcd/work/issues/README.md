# Issue ledger

The per-repo issue ledger: abcd's structured replacement for a free-form
`issues.md`. It lives here, under the shared working tier (`.abcd/work/`), so it
is committed and travels with the repository. Each issue is a single
YAML-frontmatter + Markdown-body file named `iss-<N>-<slug>.md`, with an
unpadded, per-repo `iss-N` id namespace.

This document is the store contract. The write side is
`internal/core/capture`; the front door is `abcd capture`.

## The three states

An issue's status is its folder — there is no `status:` frontmatter field.
Membership of one of these three directories *is* the status signal:

- `open/` — live, unresolved issues.
- `resolved/` — issues closed by an action; each carries a non-empty
  `resolution` note.
- `wontfix/` — issues closed by an explicit decision not to act; each carries a
  non-empty `wontfix_reason`.

An issue moves between states by being relocated between these folders, never by
editing a field. Do not add `README.md` files inside `open/`, `resolved/`, or
`wontfix/`: only genuine `iss-N` files belong there (stray markdown is ignored
by the scanner, but keeping the folders clean keeps the contract honest).

## Schema fields

Frontmatter is validated strictly (unknown keys are rejected). The reader
handles `schema_version: 1`.

Required:

- `schema_version` — integer, currently `1`.
- `id` — `iss-N` (matches the filename's id).
- `slug` — kebab-case summary (matches the filename's slug).
- `severity` — one of `critical`, `major`, `minor`, `nitpick`.
- `category` — the loose taxonomy (`bug`, `documentation`, `drift`,
  `inconsistency`, `tech-debt`, `security`, `ux`, `process`,
  `architectural-insight`, `future-work-seed`, `observation`).
- `source` — the surfacing channel (`plan-review`, `impl-review`,
  `manual-test`, `review-followup`, `agent-finding`, `user-observation`,
  `drift-detection`, `memory-curation`).
- `found_during` — non-empty session or command context.

Optional:

- `found_at` — repo-relative path or conceptual location.
- `related_intents` — list of `itd-N` ids.
- `related_specs` — list of `fn-N` ids.
- `related_issues` — list of `iss-N` ids.
- `blocked_by` — list of `iss-N` ids this issue depends on (see below).
- `promoted_to` — the `itd-N` this issue graduated into.
- `resolution` — required and non-empty in `resolved/`; forbidden elsewhere.
- `wontfix_reason` — required and non-empty in `wontfix/`; forbidden elsewhere.
- `resolved_by` — optional pointer object (`intent`, `spec`, `commit`).
- `details`, `suggested_fix`, `synthesis_clusters` — free-form provenance.

There is no `created` or `updated` field. Git is the canonical source of an
issue's timeline; the ledger does not duplicate it.

## The capture verb

`abcd capture "<text>"` appends a new issue to `open/`, allocating the next
`iss-N`. Flags refine the frontmatter — `--severity`, `--category`, `--source`,
`--slug`, `--found-during`, `--found-at`, and `--blocked-by` (a comma-separated
list of `iss-N` ids). Bare `abcd capture` renders a read-only status board;
`abcd capture list` filters by state; `abcd capture resolve` and
`abcd capture wontfix` move an open issue to its closed folder with a note.

## Derived priority

`blocked_by` records a typed dependency edge in one direction only: the
dependent issue names the issues it waits on. There is no stored priority field.

Priority is a read-time projection computed by `list` and the status board:

1. **Unblocked issues come first.** An issue is *blocked* if any of its
   `blocked_by` targets is still in `open/`; once every target has moved to
   `resolved/` or `wontfix/`, the issue is unblocked again.
2. **Within each group, higher severity comes first**
   (`critical` > `major` > `minor` > `nitpick`).

Blocked rows are annotated with the still-open targets, for example
`[blocked-by iss-3,iss-7]`. Because the projection is derived, resolving a
blocker automatically re-prioritises everything that depended on it — nothing is
stored, so nothing goes stale.
