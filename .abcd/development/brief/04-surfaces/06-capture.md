# `/abcd:capture` — Issue Ledger

`/abcd:capture` is the lightweight write side of a structured issue ledger, the structured replacement for gitignored, free-form scratch notes under `.abcd/.work.local/`. Every captured issue gets a stable `iss-N` ID, a frontmatter schema, and folder-as-status (`open/`, `resolved/`, `wontfix/`). Cross-corpus synthesis (`/abcd:dredge`) comes in a later phase as itd-25 — capture earns its keep on day one; synthesis only earns its keep once a meaningful ledger has accumulated.

See itd-4 for the full intent. Ledger schema lives in the Go binary (`internal/core`).

## Sub-verbs

> _Machine-checked (`surface_coverage`, spc-27): each row records the verb's
> adr-40 bucket (`lint` / `review` / `audit` / `gate`, or `—` for a
> non-assessment verb) and its existence (`shipped` / `staged`), verified
> against the committed command-tree snapshot in both directions._

| Verb | Bucket | Status |
|---|---|---|
| `list` | — | shipped |
| `promote` | — | shipped |
| `resolve` | — | shipped |
| `wontfix` | — | shipped |


## 1. Subcommands

| Subcommand | Purpose | File movement |
|---|---|---|
| `/abcd:capture` (no args) | Help + status: shows the most recent open issues and closes with a three-way routing hint — capture vs intent, plus an ideate route (a big, unproven idea? `abcd ideate` runs the optional admission gauntlet and records the verdict either way). Bare invocation owns the default status/help render — there is no implicit-default filtered list. | — |
| `/abcd:capture "<text>"` | Fast path: appends a structured issue entry to the ledger with auto-assigned `iss-N`; provenance and taxonomy are caller-supplied flags — `--severity`, `--category`, `--source`, and `--found-during` each carry a default; `--found-at`, `--slug`, and `--blocked-by` (comma-separated `iss-N` dependency edges; blocked/priority status is derived from `blocked_by`, never stored) have none | writes `.abcd/work/issues/open/iss-N-<slug>.md` |
| `/abcd:capture list --open` | Query the ledger for currently-open issues (flag immediately adjacent — earned SD001 exception) | — |
| `/abcd:capture list --resolved` | Query the ledger for resolved issues | — |
| `/abcd:capture list --wontfix` | Query the ledger for wontfix issues | — |
| `/abcd:capture list --all` | Query the ledger across all three states | — |
| `/abcd:capture promote <iss-N\|rdi-N> [--intent <itd-N>]` | Promote an issue to an intent draft, native CLI sub-verb (spc-24, itd-119): one invocation mints a draft under `intents/drafts/` — slug reused from the issue, body carrying a by-id pointer, never a copy (SSOT) — and stamps the issue's `promoted_to` with the minted `itd-N`; the draft's `promoted_from` is the reciprocal edge. Works from any status folder (promotion is orthogonal to fix-status). `--intent` is the stamp-only mode that links an existing draft — the repair path after a post-mint stamp failure, which the error names. A dispositioned reading item (`rdi-N`) graduates through the same verb; an undispositioned one is refused before anything is minted (spc-58). | (issue stays; intent created in `drafts/`) |
| `/abcd:capture resolve <iss-N> "<resolution-note>" --impact <additive\|breaking\|fix\|internal>` | Mark issue resolved (`--impact` is required — resolving without it exits 1). Three optional `resolved_by` provenance flags name what fixed it: `--intent itd-N` (must exist), `--spec spc-N` (must exist), `--commit <sha>` (shape-checked hex) | `open/` → `resolved/` |
| `/abcd:capture wontfix <iss-N> "<reason>"` | Explicit non-action decision | `open/` → `wontfix/` |

The unfiltered CLI shape `abcd capture list` (no flag) is rejected with
exit 2 and a "choose a filter: --open / --resolved / --wontfix / --all"
message. The four flagged forms above are the only earned SD001
exception under the capture surface; each flag must appear immediately
adjacent to `list`, never pipe-joined into a single token. There is no
implicit-default filtered list — bare `/abcd:capture` is what renders
status and recent captures, closing with a three-way routing hint
(capture vs intent, plus an ideate route for a big, unproven idea).

## 2. Ledger structure

Frontmatter fields (per the issue ledger schema in `internal/core`):

```yaml
---
schema_version: 1
id: iss-N                  # unpadded, mirrors itd-N
slug: <kebab-case>
severity: nitpick|minor|major|critical
impact: additive|breaking|fix|internal   # required in resolved/; drives the derived version and changelog inclusion
category: bug|documentation|drift|inconsistency|tech-debt|security|ux|process|architectural-insight|future-work-seed|observation|lapse
source: plan-review|impl-review|manual-test|review-followup|agent-finding|agent-observation|user-observation|drift-detection|memory-curation
found_during: <session-or-command-context>
found_at: <path-or-conceptual>
details: "<text>"          # optional structured detail
suggested_fix: "<text>"    # optional proposed remedy
related_intents: [itd-N, ...]
related_specs: [spc-N, ...]
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

**Verify a `--commit` stamp is reachable before writing it.** The flag is
shape-checked and nothing more: `^[0-9a-f]{7,64}$` proves the value looks
like a sha, not that the commit exists or that it is reachable from the
default branch. A stamp that points at nothing is indistinguishable from a
good one when read, so the check belongs at write time:

```sh
git merge-base --is-ancestor <sha> origin/main
```

The reason to check rather than assume is that whether a branch's own shas
survive a merge depends on the merge method, and that is a repository
setting which can change without announcement. A habit resting on either
answer is correct only until the setting moves; the command above is
correct under both. Where a merge produces two reachable candidates, prefer
the commit that carries the change over the merge commit, whose diff is the
whole pull request rather than the fix.

Body is free-form: details, suggested fix, links to context.

## 3. Legacy scratch migration

A later phase, not yet built — the migration rides the `abcd dev-sync work` surface ([`08-abcd.md`](08-abcd.md)); the shipped Go ledger engine reserves a migrator-only `ForceID` seam for it. On first run of `abcd dev-sync work` after install (or first `/abcd:ahoy` upgrade), the command parses a free-form scratch buffer under `.abcd/.work.local/` entry-by-entry and promotes each to a corresponding `.abcd/work/issues/open/iss-N-<slug>.md`. Idempotent. The original scratch buffer under `.abcd/.work.local/` is preserved as a staging buffer (still works for ad-hoc scribbles; subsequent entries promoted on the next `abcd dev-sync work`).

## 4. Acceptance

- **Given** an abcd-installed repo, **when** the user runs `/abcd:capture "review nitpick: T7 cache_ttl_days dead-config alternative"`, **then** a new file `.abcd/work/issues/open/iss-N-<slug>.md` exists with frontmatter populated and the captured text in the body.
- **Given** an existing issue at `.abcd/work/issues/open/iss-3-foo.md`, **when** the user runs `/abcd:capture resolve iss-3 "fixed in spc-7 task 4" --impact fix` (`--impact` is required and has no default), **then** the file moves to `.abcd/work/issues/resolved/iss-3-foo.md` with the resolution recorded.
- **Given** an existing issue in any status folder, **when** the user runs `/abcd:capture promote iss-N`, **then** one invocation files a new draft intent under `intents/drafts/` — slug reused, body a by-id pointer to the issue, `promoted_from: iss-N` in its frontmatter — and stamps the issue's `promoted_to` with the minted `itd-N`, the issue keeping its folder; an issue already promoted is refused with the existing `itd-N`, and a post-mint stamp failure names the orphan draft and the `--intent` repair.
- **Given** a fresh `/abcd:ahoy` upgrade with an existing scratch buffer under `.abcd/.work.local/`, **when** `dev-sync` runs (a later phase, not yet built — § 3), **then** every entry in that scratch buffer is promoted to the structured ledger with provenance noting "migrated from `.abcd/.work.local/` scratch".
- **Given** a ledger containing 5 open issues, **when** the user runs `/abcd:capture list --open` (the flag is explicit — there is no implicit default), **then** the output lists all 5 with id, state, severity, and slug, in derived-priority order — unblocked issues first, then severity (`critical` → `nitpick`); rows blocked by an open dependency are demoted and annotated with their open blockers.
- **Given** a ledger with a mix of open, resolved, and wontfix issues, **when** the user runs `/abcd:capture list --all`, **then** every issue across all three states is listed; the equivalent unfiltered CLI form `abcd capture list` (no flag) instead exits 2 with a "choose a filter" message.
- **Given** an abcd-installed repo, **when** the user runs bare `/abcd:capture` (no args), **then** the output is a read-only status render — counts (`open N · resolved N · wontfix N`), up to 10 most recent open issues, and a three-way routing hint (capture vs intent, plus an ideate route) — and no `iss-*.md` file is created, moved, or field-mutated by the invocation itself.

## 5. Implementation status

- **Library primitives:** delivered by the predecessor's `spc-20-issue-ledger-primitives-iss-n-allocator` (predecessor store).
  The API the command surface consumes is the Go package
  `internal/core/capture` (allocator, find, read, build, mutate; capture,
  resolve, wontfix, list, status) — a port of the predecessor's `_issue_lib`
  and `issue_workflow` primitives.
- **Command flow:** delivered by the predecessor's `spc-21-abcdcapture-command-flow-text-ingest` (predecessor store).
- **Legacy `.abcd/.work.local/` scratch migration:** design target per the predecessor's `spc-22-workissuesmd-migration-promote-legacy` (predecessor store) — a later phase, not yet built (rides the `dev-sync` surface, § 3).
- **intent-auditor cross-check:** delivered by the predecessor's `spc-23-intent-auditor-extension` (predecessor store); the reviewer surface ships as `agents/intent-auditor.md`.
- **`promote <iss-N>` bridge:** a native engine sub-verb (spc-24, itd-119) —
  `internal/core/capture/promote.go` mints the draft under `intents/drafts/`
  and stamps the issue's `promoted_to` field itself, one invocation writing
  both edges (the draft's `promoted_from` is the reciprocal); the command file
  invokes the binary verb directly. It superseded the earlier
  command-orchestrated flow that leaned on `abcd intent "<text>"` and left the
  back-link to be written by hand (see the § 1 table row).
