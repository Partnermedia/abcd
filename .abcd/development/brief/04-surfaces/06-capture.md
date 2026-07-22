# `/abcd:capture` — Issue Ledger

`/abcd:capture` is the lightweight write side of a structured issue ledger, the structured replacement for gitignored, free-form scratch notes under `.abcd/.work.local/`. Every captured issue gets a stable `iss-N` ID, a frontmatter schema, and folder-as-status (`open/`, `resolved/`, `wontfix/`). Cross-corpus synthesis (`/abcd:dredge`) comes in a later phase as itd-25 — capture earns its keep on day one; synthesis only earns its keep once a meaningful ledger has accumulated.

See itd-4 for the full intent. Ledger schema lives in the Go binary (`internal/core`).

## 1. Subcommands

| Subcommand | Purpose | File movement |
|---|---|---|
| `/abcd:capture` (no args) | Help + status: shows the most recent open issues and closes with a capture-vs-intent routing hint. Bare invocation owns the default status/help render — there is no implicit-default filtered list. | — |
| `/abcd:capture "<text>"` | Fast path: appends a structured issue entry to the ledger with auto-assigned `iss-N`; provenance and taxonomy are caller-supplied flags, each with a default — `--severity`, `--category`, `--source`, `--found-during`, `--found-at`, `--slug`, `--blocked-by` (comma-separated `iss-N` dependency edges; blocked/priority status is derived from `blocked_by`, never stored) | writes `.abcd/work/issues/open/iss-N-<slug>.md` |
| `/abcd:capture list --open` | Query the ledger for currently-open issues (flag immediately adjacent — earned SD001 exception) | — |
| `/abcd:capture list --resolved` | Query the ledger for resolved issues | — |
| `/abcd:capture list --wontfix` | Query the ledger for wontfix issues | — |
| `/abcd:capture list --all` | Query the ledger across all three states | — |
| `/abcd:capture promote <iss-N>` | Promote an issue to an intent draft: command-orchestrated (never a `capture promote` CLI sub-verb), it hands the issue body to `abcd intent "<text>"`, which files a new draft under `intents/drafts/` seeded from that text. The reciprocal back-link (`promoted_to` scalar + `related_issues` ↔ `related_intents` lists) onto the `iss-N` record is written by hand — no engine verb writes that edge. | (issue stays; intent created in `drafts/`) |
| `/abcd:capture resolve <iss-N> "<resolution-note>"` | Mark issue resolved | `open/` → `resolved/` |
| `/abcd:capture wontfix <iss-N> "<reason>"` | Explicit non-action decision | `open/` → `wontfix/` |

The unfiltered CLI shape `abcd capture list` (no flag) is rejected with
exit 2 and a "choose a filter: --open / --resolved / --wontfix / --all"
message. The four flagged forms above are the only earned SD001
exception under the capture surface; each flag must appear immediately
adjacent to `list`, never pipe-joined into a single token. There is no
implicit-default filtered list — bare `/abcd:capture` is what renders
status and recent captures, closing with a capture-vs-intent routing
hint.

## 2. Ledger structure

Frontmatter fields (per the issue ledger schema in `internal/core`):

```yaml
---
schema_version: 1
id: iss-N                  # unpadded, mirrors itd-N
slug: <kebab-case>
severity: nitpick|minor|major|critical
impact: additive|breaking|fix|internal   # required in resolved/; drives the derived version and changelog inclusion
category: bug|documentation|drift|inconsistency|tech-debt|security|ux|process|architectural-insight|future-work-seed|observation
source: plan-review|impl-review|manual-test|review-followup|agent-finding|agent-observation|user-observation|drift-detection|memory-curation
found_during: <session-or-command-context>
found_at: <path-or-conceptual>
details: "<text>"          # optional structured detail
suggested_fix: "<text>"    # optional proposed remedy
related_intents: [itd-N, ...]
related_specs: [fn-N, ...]
related_issues: [iss-N, ...]
synthesis_clusters: [<label>, ...]  # optional dredge/synthesis grouping
blocked_by: [iss-N, ...]   # dependency edges; blocked/priority is derived, never stored
promoted_to: itd-M         # set when the issue is promoted to an intent
wontfix_reason: "<text>"   # required when in wontfix/
resolution: "<one-line>"   # required when in resolved/
resolved_by:               # optional structured pointer to what resolved it
  intent: itd-M
  spec: spc-N
  commit: <sha>
---
```

Enum values above mirror the issue ledger schema in `internal/core`
exactly; the schema is the single source of truth.

Body is free-form: details, suggested fix, links to context.

## 3. Legacy scratch migration

A later phase, not yet built — the migration rides the `abcd dev-sync work` surface ([`08-abcd.md`](08-abcd.md)); the shipped Go ledger engine reserves a migrator-only `ForceID` seam for it. On first run of `abcd dev-sync work` after install (or first `/abcd:ahoy` upgrade), the command parses a free-form scratch buffer under `.abcd/.work.local/` entry-by-entry and promotes each to a corresponding `.abcd/work/issues/open/iss-N-<slug>.md`. Idempotent. The original scratch buffer under `.abcd/.work.local/` is preserved as a staging buffer (still works for ad-hoc scribbles; subsequent entries promoted on the next `abcd dev-sync work`).

## 4. Acceptance

- **Given** an abcd-installed repo, **when** the user runs `/abcd:capture "review nitpick: T7 cache_ttl_days dead-config alternative"`, **then** a new file `.abcd/work/issues/open/iss-N-<slug>.md` exists with frontmatter populated and the captured text in the body.
- **Given** an existing issue at `.abcd/work/issues/open/iss-3-foo.md`, **when** the user runs `/abcd:capture resolve iss-3 "fixed in spc-7 task 4"`, **then** the file moves to `.abcd/work/issues/resolved/iss-3-foo.md` with the resolution recorded.
- **Given** an existing issue, **when** the user runs `/abcd:capture promote iss-N`, **then** the issue body is handed to `abcd intent "<text>"`, which files a new draft intent under `intents/drafts/` seeded from that text; the reciprocal back-link (`promoted_to` scalar + `related_issues` ↔ `related_intents` lists) is recorded by hand, since no engine verb writes that edge.
- **Given** a fresh `/abcd:ahoy` upgrade with an existing scratch buffer under `.abcd/.work.local/`, **when** `dev-sync` runs (a later phase, not yet built — § 3), **then** every entry in that scratch buffer is promoted to the structured ledger with provenance noting "migrated from `.abcd/.work.local/` scratch".
- **Given** a ledger containing 5 open issues, **when** the user runs `/abcd:capture list --open` (the flag is explicit — there is no implicit default), **then** the output lists all 5 with id, state, severity, and slug, in derived-priority order — unblocked issues first, then severity (`critical` → `nitpick`); rows blocked by an open dependency are demoted and annotated with their open blockers.
- **Given** a ledger with a mix of open, resolved, and wontfix issues, **when** the user runs `/abcd:capture list --all`, **then** every issue across all three states is listed; the equivalent unfiltered CLI form `abcd capture list` (no flag) instead exits 2 with a "choose a filter" message.
- **Given** an abcd-installed repo, **when** the user runs bare `/abcd:capture` (no args), **then** the output is a read-only status render — counts (`open N · resolved N · wontfix N`), up to 10 most recent open issues, and a capture-vs-intent routing hint — and no `iss-*.md` file is created, moved, or field-mutated by the invocation itself.

## 5. Implementation status

- **Library primitives:** delivered by `spc-20-issue-ledger-primitives-iss-n-allocator`.
  The API the command surface consumes is the Go package
  `internal/core/capture` (allocator, find, read, build, mutate; capture,
  resolve, wontfix, list, status) — a port of the predecessor's `_issue_lib`
  and `issue_workflow` primitives.
- **Command flow:** delivered by `spc-21-abcdcapture-command-flow-text-ingest`.
- **Legacy `.abcd/.work.local/` scratch migration:** design target per `spc-22-workissuesmd-migration-promote-legacy` — a later phase, not yet built (rides the `dev-sync` surface, § 3).
- **intent-fidelity-reviewer cross-check:** delivered by `spc-23-intent-fidelity-reviewer-extension`.
- **`promote <iss-N>` bridge:** the command-orchestrated flow leans on
  `abcd intent "<text>"`, delivered by `spc-7-abcd-intent-quoted-text-create-symmetric`
  (itd-46). The issue body is handed to that create path, which files a new draft;
  the reciprocal back-link onto the `iss-N` record is written by hand —
  no engine verb writes that edge.
