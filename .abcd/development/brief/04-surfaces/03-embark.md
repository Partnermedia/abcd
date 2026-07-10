# `/abcd:embark` — Unpack a Lifeboat

> **Phase ownership** ([adr-33](../../decisions/adrs/0033-launch-phase-ownership-tiered.md)): the lifeboat round-trip — disembark packing and embark unpacking — ships in [Phase 6](../../roadmap/phases/phase-6-lifeboat.md).

> **Recovery humility.** Embark unpacks the lifeboat into a working repo. The lifeboat is the highest-fidelity floor the originating session could leave behind; it is not the activity that produced it. **When something here doesn't make sense, hunt the originating session before trusting the lifeboat blindly** — ask the prior author, surface the chat where the decision happened, look at the rejected alternatives. The lifeboat is a starting point, not an oracle. See [`01-product/03-mental-model.md § The Naurian gap`](../01-product/03-mental-model.md#the-naurian-gap) for the framing.

## Sub-verbs

Bare `/abcd:embark` shows status + help only — never mutates state. Current sub-verbs:

- **`/abcd:embark from <path>`** — unpack the lifeboat at `<path>` into the current repo. Path is required; use `home` as shorthand for the current repo's `.abcd/lifeboat/` (round-trip / self-test case). Flag-shaped modifiers: `--force` (override emptiness-rule refusal), `--archive` (copy input lifeboat verbatim to `voyage/embark/from/<timestamp>/` before unpacking), `--refresh-audit` (re-run oracle product audit instead of trusting cached).
- **`/abcd:embark scan`** — walk siblings (one level above current repo only), list discovered `.abcd/lifeboat/` directories ranked by mtime, present candidates via transparent prompt. **No unpacking.** Useful before `embark from <path>` when the user isn't sure where lifeboats live. Flag-shaped modifier: `--deep` (walks the full tree under `../` up to 3 levels for power users).
- **`/abcd:embark probe <path>`** — inspect a lifeboat at `<path>` without unpacking: show what would land where, run schema/audit checks, write nothing.
- **Later phase: `/abcd:embark from-spec-kit <path>`** — ingest a GitHub Spec Kit project directory as starter draft intents (per itd-23).

## 1. Source lookup

Per [`02-constraints/01-platform.md § Lifeboat path`](../02-constraints/01-platform.md#lifeboat-path), lifeboats are *output* — `.abcd/lifeboat/` in any repo is that repo's latest disembark snapshot, not a registry of inbound lifeboats. Embark therefore reads from external sources by default; the only time embark reads `.abcd/lifeboat/` in cwd is the round-trip / self-test case (via `from home`).

Path resolution under `from <path>`:

1. `<path>` is `home` → expand to current repo's `.abcd/lifeboat/`. **Round-trip case** — embark from this repo's own disembark output (e.g., for testing the disembark→embark round-trip on the same repo).
2. Any other `<path>` → validate, use. The normal case for inbound embark from another repo.
3. To *find* candidate lifeboats before running `from <path>`, use `embark scan` (or `embark scan --deep`) — that's a separate sub-verb, not a flag on `from`.

**Provenance and `--archive`**: `embark from <path>` records `source_path` and `source_manifest_sha256` in `.abcd/development/voyage/embark/provenance.json` (see [§ 7](#7-voyage-layout-embarkdisembark-provenance-and-history)). Opt-in `embark from <path> --archive` additionally copies the input lifeboat verbatim into `voyage/embark/from/<timestamp>/` for the case where the source repo will disappear. Off by default; `source_path` + hash is sufficient when the source repo persists.

No global `~/.abcd/archive/`.

## 2. Target emptiness

Hard stop unless only `.git/`, `.gitignore`, `LICENSE*`, `README.md`, `.github/` present. `embark from <path> --force` writes `embark-conflict-report.md` and proceeds with conflict resolution ([§ 4](#4-conflict-ux)).

## 3. Scaffold steps

0. **Read lifeboat:** `press-release.md`, `README.json`, `principles.json`, `spec-essence.json`, `decisions-timeline.json`, `code-principles.json`, plus the disembark-time `audit/press-release-oracle-*.md` and `audit/documentation-audit-*.md`.
1. **Press release interview (FIRST INTERACTION).** Show the user the press release + the disembark-time product audit findings. Ask: "Confirm / amend / reframe before scaffolding." Amended press release becomes the new repo's **initial brief**: written to `.abcd/development/brief/README.md` in the embarked target. Subsequent ahoy/work iterates on it like any other abcd-managed project's brief. **`embark from <path> --refresh-audit`** flag re-runs the oracle product audit before showing the user (uses current oracle config; flags drift vs the disembark-time audit in the report).
2. Show scaffold summary (referencing the amended press release); transparent confirm to proceed.
3. Create dirs; copy ADRs, terminology, docs verbatim from lifeboat to canonical target locations.
4. Create specs from `rescue/spec-plan.md` in the native spec store (via `/abcd:intent plan` / `ship`, per [adr-26](../../decisions/adrs/0026-native-spec-layer-ccpm-backend.md)), or create a minimal native spec structure.
5. Write curated memory files to `.abcd/memory/` — one per principle, grouped by domain, filename `<type>_<domain>_<slug>.md` (e.g., `feedback_ui_full_box_hit_target.md`) with frontmatter conforming to the shipped memory-page schema — a typed `source:` provenance block (class/classes, licence, source_hash, weighting_note) that memory-lint's ML001 blocker requires; the shipped schema carries no `name`/`description`/`type` keys (`internal/core/memory/schema.go`, `lint.go`). Only the `<type>_<domain>_<slug>.md` filename grammar matches the shipped store (`ParsePageFilename`). The volatile memory backend's source store is left untouched in the new repo's environment — the harness (Claude Code, OpenCode, etc.) will populate it as the user works.
6. Inject principles into CLAUDE.md/AGENTS.md between BEGIN/END markers (idempotent). The marker block content is the modular-rules-loader block the embark scaffolder (`internal/core/...`) emits (per itd-3) — *not* a verbose copy of every principle. Principles are exposed via the rules-loader's domain rules, surfaced on demand by prompt-keyword recall.
7. Apply asset curation ([§ 5](#5-asset-curation-per-_manifestjson-classifications)).
8. **Documentation-auditor** (subagent) runs post-scaffold, before final report, to verify the target's user-facing docs (tutorials, guides, reference, explanation) are well-formed.
9. **If `.abcd/rp/workspace.json` exists in the lifeboat** (per itd-7): ask the user "Register this RP workspace with RepoPrompt now?" — if yes, write to the embarker's `~/Library/Application Support/RepoPrompt/Workspaces/`. If RP isn't installed, warn gracefully and continue.
10. **Write voyage provenance** (per [§ 7](#7-voyage-layout-embarkdisembark-provenance-and-history)): create `.abcd/development/voyage/embark/provenance.json` with `source_path`, `source_manifest_sha256`, `timestamp`, `files_written`, `press_release_amended_diff`, and (if `--refresh-audit`) `audit_drift`. If `embark --archive` was passed, also copy the input lifeboat verbatim to `.abcd/development/voyage/embark/from/<timestamp>/`.
11. Write `embark-report.{json,md}` (includes the amended press release diff, audit drift if `--refresh-audit`, RP workspace registration status, voyage provenance path).

`embark-scaffolder` agent emits a JSON scaffold plan; deterministic Python applies it. Agent does judgement (which principles where, how to phrase injection); Python does file creation. **Press release is treated as a hard input** to the scaffolder — if a principle or spec-essence entry contradicts the amended press release, the scaffolder flags the conflict in the embark report.

## 4. Conflict UX

When the target already has files that the lifeboat would write, **a single bulk prompt** covers all conflicts (no per-file barrage):

```
embark detected N conflicts across:
  • 3 native spec store files
  • 1 CLAUDE.md (will inject markers if 'merge')
  • 2 .abcd/memory/ files
  • 1 .abcd/development/decisions/adrs/
  • 1 .abcd/development/brief/README.md (existing brief vs incoming press release)

How to resolve all conflicts?
  → keep target (skip everything in lifeboat that conflicts)
  → replace target (lifeboat wins everywhere)
  → merge where possible, prompt otherwise (CLAUDE.md gets marker injection;
    ADRs/specs/memory get per-file prompt)
  → abort and write embark-conflict-report.md
```

Single decision, transparent (shows scope before asking).

## 5. Asset curation (per `_manifest.json` classifications)

- **`keep`** → copy verbatim
- **`adapt`** → transparent prompt with curator's suggested adaptation; user accepts/edits/skips
- **`drop`** → skip silently

## 6. Acceptance

- **Given** any abcd-aware terminal, **when** the user runs bare `/abcd:embark`, **then** the dispatcher shows whether a lifeboat is detectable in the current location, the available sub-verbs (`from <path>`, `scan`, `probe`; later phase: `from-spec-kit`), and suggested next actions — bare invocation never mutates state.
- **Given** a lifeboat at `<path>` and an empty target repo (only `.git/`, `LICENSE`, `README.md`), **when** `/abcd:embark from <path>` runs, **then** the target receives the press-release interview as the first interaction, the amended press release becomes `.abcd/development/brief/README.md`, and all sections in [§ 3](#3-scaffold-steps) land at canonical locations.
- **Given** `<path>` is `home`, **when** any embark sub-verb resolves a `<path>` argument equal to `home`, **then** the path expands to the current repo's `.abcd/lifeboat/` (round-trip case) and embark proceeds as normal. (Generalises to future sub-verbs that take a `<path>` argument; matches disembark's parallel rule for sub-verb-agnostic resolution.)
- **Given** a non-empty target repo, **when** `/abcd:embark from <path>` runs without `--force`, **then** the command refuses and writes `embark-conflict-report.md` listing the conflicts.
- **Given** a non-empty target with conflicts and `--force`, **when** `embark from <path> --force` runs, **then** the bulk conflict prompt fires once with a summary of all conflicts, and the chosen resolution is applied uniformly.
- **Given** the user runs `/abcd:embark scan`, **when** the command completes, **then** `.abcd/lifeboat/` directories in sibling locations (one level above) are listed ranked by mtime with their detected source repo, no unpacking occurs, and the user is shown candidates ready to pass to `embark from <path>`.
- **Given** the user runs `/abcd:embark scan --deep`, **when** the command completes, **then** the search recursively walks up to 3 levels under `../`.
- **Given** the user runs `/abcd:embark probe <path>`, **when** the command completes, **then** the lifeboat at `<path>` is inspected (file tree, schema validation, would-be writes), no target mutation occurs, and the user sees a report ready to inform the decision to run `embark from <path>`.
- **Given** a lifeboat containing `.abcd/rp/workspace.json` and RP installed on the embarker, **when** `embark from <path>` runs, **then** the user is asked whether to register the workspace with RP and the choice is applied.
- **Given** a lifeboat containing `.abcd/rp/workspace.json` and RP *not* installed, **when** `embark from <path>` runs, **then** the command warns gracefully and continues without failing.
- **Given** the user passes `--refresh-audit`, **when** `embark from <path> --refresh-audit` runs, **then** the oracle product audit re-runs against the current lifeboat content and the drift vs the disembark-time audit is reported.
- **Given** an asset manifest entry classified as `adapt`, **when** `embark from <path>` applies asset curation, **then** the user is shown the curator's suggested adaptation and asked transparently to accept / edit / skip.
- **Given** an `embark from <path>` run completes, **then** `.abcd/development/voyage/embark/provenance.json` exists with `source_path`, `source_manifest_sha256`, `timestamp`, and `files_written` populated (per [§ 7](#7-voyage-layout-embarkdisembark-provenance-and-history)); `embark/from/<timestamp>/` is absent unless `--archive` was passed.
- **Given** an `embark from <path> --archive` run completes, **then** the input lifeboat is copied verbatim to `.abcd/development/voyage/embark/from/<timestamp>/` and the path is referenced from `provenance.json`.

## 7. Voyage layout — embark/disembark provenance and history

Lifeboat *operations* (embark, disembark) write provenance and history to `.abcd/development/voyage/`. The `.abcd/lifeboat/` directory itself holds only the latest output snapshot ([`02-disembark.md § 5`](02-disembark.md#5-output-shape)); it does not accumulate.

```
.abcd/
├── lifeboat/                                ← output snapshot only (02-disembark.md § 5); regenerable
└── development/
    └── voyage/
        ├── README.md
        ├── embark/
        │   ├── provenance.json              ← source path, manifest hash, timestamp, files written
        │   └── from/<timestamp>/            ← --archive: verbatim copy of input lifeboat (opt-in)
        └── disembark/
            └── history.jsonl                ← append-only manifest log of every disembark
```

**`embark/provenance.json`** records, for the embark that bootstrapped this repo:

- `source_path` (the `<path>` argument passed to `embark from <path>`)
- `source_manifest_sha256` (hash of input lifeboat's `_provenance.json` + file tree)
- `timestamp`
- `files_written` (target paths created in `.abcd/development/` — including ADRs at `.abcd/development/decisions/adrs/` — the native spec store, `.abcd/memory/`, etc.)
- `press_release_amended_diff` (diff between input lifeboat's `press-release.md` and the brief that landed at `.abcd/development/brief/README.md` after the [§ 3 step 1](#3-scaffold-steps) interview)
- `audit_drift` (only if `--refresh-audit`: drift vs disembark-time `audit/press-release-oracle-*`)

**`embark/from/<timestamp>/`** — opt-in via `embark --archive`. Verbatim copy of the input lifeboat at the moment of embark, for the case where the source repo will disappear. Off by default; the `source_path` + hash in `provenance.json` is sufficient when the source repo persists.

**`disembark/history.jsonl`** — append-only, one JSON object per disembark run:

```json
{
  "timestamp": "2026-05-04T14:30:00Z",
  "manifest_sha256": "abc123...",
  "files": ["README.md", "press-release.md", "rescue/specs/spc-1-foo.md", ...],
  "label": "post-itd-7-ship",
  "shared_with": ["idelphiDev"],
  "oracle_backend": "host-delegated",
  "oracle_verdict": "sufficient"
}
```

Manifests are small (file list + hashes, not contents); the log answers "what did this repo's lifeboat look like at point T?" and "did I share that lifeboat with anyone?" without keeping stale snapshots around. Acceptance criteria for voyage writes live in [`02-disembark.md § 7`](02-disembark.md#7-acceptance) (disembark-side history.jsonl) and [§ 6](#6-acceptance) above (embark-side provenance.json + optional --archive copy).
