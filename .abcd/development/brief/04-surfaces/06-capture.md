# `/abcd:capture` — Issue Ledger

`/abcd:capture` is the lightweight write side of a structured issue ledger, the replacement for the gitignored, free-form `.work/issues.md`. Every captured issue gets a stable `iss-N` ID, a frontmatter schema, and folder-as-status (`open/`, `resolved/`, `wontfix/`). Cross-corpus synthesis (`/abcd:dredge`) comes in a later phase as itd-25 — capture earns its keep on day one; synthesis only earns its keep once a meaningful ledger has accumulated.

See itd-4 for the full intent. Ledger schema: `scripts/abcd/schemas/issue.schema.json`.

## 1. Subcommands

| Subcommand | Purpose | File movement |
|---|---|---|
| `/abcd:capture` (no args) | Help + status: shows recent captures grouped by state, suggests next actions. Bare invocation owns the default status/help render — there is no implicit-default filtered list. | — |
| `/abcd:capture "<text>"` | Fast path: appends a structured issue entry to the ledger with auto-assigned `iss-N`, timestamp, source provenance (which session / command / file the user was in) | writes `.abcd/development/activity/issues/open/iss-N-<slug>.md` |
| `/abcd:capture list --open` | Query the ledger for currently-open issues (flag immediately adjacent — earned SD001 exception) | — |
| `/abcd:capture list --resolved` | Query the ledger for resolved issues | — |
| `/abcd:capture list --wontfix` | Query the ledger for wontfix issues | — |
| `/abcd:capture list --all` | Query the ledger across all three states | — |
| `/abcd:capture promote <iss-N>` | **LIVE (fn-30, itd-46).** Promote an issue to an intent draft: reads the issue, runs the agent-driven intent-new interview seeded with the issue body, and writes the four-field back-link (`source_issue`/`promoted_to` scalars + `related_issues` ↔ `related_intents` lists). Skill-orchestrated — never a `capture promote` CLI sub-verb. | (issue stays; intent created in `drafts/`) |
| `/abcd:capture resolve <iss-N> "<resolution-note>"` | Mark issue resolved | `open/` → `resolved/` |
| `/abcd:capture wontfix <iss-N> "<reason>"` | Explicit non-action decision | `open/` → `wontfix/` |

The unfiltered CLI shape `abcd capture list` (no flag) is rejected with
exit 2 and a "choose a filter: --open / --resolved / --wontfix / --all"
message. The four flagged forms above are the only earned SD001
exception under the capture surface; each flag must appear immediately
adjacent to `list`, never pipe-joined into a single token. There is no
implicit-default filtered list — bare `/abcd:capture` is what renders
status, recent captures, and next-action hints.

## 2. Ledger structure

Frontmatter fields (per `scripts/abcd/schemas/issue.schema.json`):

```yaml
---
schema_version: 1
id: iss-N                  # unpadded, mirrors itd-N
slug: <kebab-case>
severity: nitpick|minor|major|critical
category: bug|documentation|drift|inconsistency|tech-debt|security|ux|process|architectural-insight|future-work-seed|observation
source: plan-review|impl-review|manual-test|review-followup|agent-finding|user-observation|drift-detection|memory-curation
found_during: <session-or-command-context>
found_at: <path-or-conceptual>
related_intents: [itd-N, ...]
related_specs: [fn-N, ...]
created: YYYY-MM-DD
updated: YYYY-MM-DD
wontfix_reason: "<text>"   # required when in wontfix/
resolution: "<one-line>"   # required when in resolved/
---
```

Enum values above mirror `scripts/abcd/schemas/issue.schema.json`
exactly; the schema is the single source of truth.

Body is free-form: details, suggested fix, links to context.

## 3. `.work/issues.md` migration

On first run of `abcd dev-sync work` after install (or first `/abcd:ahoy` upgrade), the command parses `.work/issues.md` entry-by-entry and promotes each to a corresponding `.abcd/development/activity/issues/open/iss-N-<slug>.md`. Idempotent. The original `.work/issues.md` is preserved as a staging buffer (still works for ad-hoc scribbles; subsequent entries promoted on the next `abcd dev-sync work`).

## 4. Acceptance

- **Given** an abcd-installed repo, **when** the user runs `/abcd:capture "review nitpick: T7 cache_ttl_days dead-config alternative"`, **then** a new file `.abcd/development/activity/issues/open/iss-N-<slug>.md` exists with frontmatter populated and the captured text in the body.
- **Given** an existing issue at `.abcd/development/activity/issues/open/iss-3-foo.md`, **when** the user runs `/abcd:capture resolve iss-3 "fixed in fn-7 task 4"`, **then** the file moves to `.abcd/development/activity/issues/resolved/iss-3-foo.md` with the resolution recorded.
- **Given** an existing issue, **when** the user runs `/abcd:capture promote iss-N` (LIVE as of fn-30, itd-46), **then** the agent-driven intent-new interview runs seeded with the issue body; the resulting draft intent's frontmatter has `source_issue: iss-N` and `related_issues: [iss-N]`; the issue's frontmatter has `promoted_to: itd-M` and `related_intents: [itd-M]` — the four-field back-link written transactionally by `link_promoted_issue`.
- **Given** a fresh `/abcd:ahoy` upgrade with an existing `.work/issues.md`, **when** `dev-sync` runs, **then** every entry in `.work/issues.md` is promoted to the structured ledger with provenance noting "migrated from .work/issues.md".
- **Given** a ledger containing 5 open issues, **when** the user runs `/abcd:capture list --open` (the flag is explicit — there is no implicit default), **then** the output lists all 5 with id, state, severity, and slug, sorted numerically by `iss-N`.
- **Given** a ledger with a mix of open, resolved, and wontfix issues, **when** the user runs `/abcd:capture list --all`, **then** every issue across all three states is listed; the equivalent unfiltered CLI form `abcd capture list` (no flag) instead exits 2 with a "choose a filter" message.
- **Given** an abcd-installed repo, **when** the user runs bare `/abcd:capture` (no args), **then** the output is a read-only status render — counts (`open N · resolved N · wontfix N`), up to 10 most recent open issues, and suggested next actions — and no `iss-*.md` file is created, moved, or field-mutated by the invocation itself.

## 5. Implementation status

- **Library primitives:** delivered by `fn-20-issue-ledger-primitives-iss-n-allocator`.
  `_issue_lib` (allocator, find, read, build, mutate) and `issue_workflow`
  (capture, resolve, wontfix, update_field, list_issues) are the API the
  command surface consumes.
- **Command flow:** delivered by `fn-21-abcdcapture-command-flow-text-ingest`.
- **`.work/issues.md` migration:** delivered by `fn-22-workissuesmd-migration-promote-legacy`.
- **intent-fidelity-reviewer cross-check:** delivered by `fn-23-intent-fidelity-reviewer-extension`.
- **`promote <iss-N>` bridge:** delivered by `fn-30-symmetric-abcdintent-and-abcdcapture`
  (itd-46). Skill-orchestrated issue → intent-new interview → four-field
  back-link, with the `PR000`/`PR001`/`PR002` lint family guarding the
  contract.
