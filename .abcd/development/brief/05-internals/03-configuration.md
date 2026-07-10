# Configuration Model

This file holds the configuration schema (`config.json`, including its `meta` setup block), the visibility-driven gitignore policy, and the `dev-sync` namespace that pumps volatile sources into curated artefacts.

## Setup metadata — `config.json["meta"]`

Setup metadata is a `meta` block inside `.abcd/config.json`; there is no separate `.abcd/meta.json` at repo scope (spc-16). ahoy stamps and reads it via `config.json["meta"]` (example values shown):

```json
{
  "meta": {
    "schema_version": 1,
    "setup_version": "0.1.0",
    "setup_date": "2026-05-04",
    "project_name": "abcd-cli"
  }
}
```

## `.abcd/config.json`

(example values shown; ahoy populates from prompts on first run, no silent default for `visibility`):

```json
{
  "repo": {
    "visibility": "private"            // "private" | "public" — set by ahoy each run, no silent default
  },
  "ai_transparency": {
    "level": "metadata"                 // "full" | "metadata" | "none" — separate axis from visibility
                                        // full: conversations + plans + tasks + metadata
                                        // metadata: plans + tasks + metadata (no conversation transcripts)
                                        // none: nothing
                                        // Drives what dev-sync captures and what disembark includes.
  },
  "docs": {
    "target": "both"                    // "claude_md" | "agents_md" | "both" | "skip"
  },
  "oracle": {                           // oracle seam — internal/adapter/oracle
    "backend": "host-delegated"         // "host-delegated" | "native" | "cli" | "api" | "mcp"
                                        // host-delegated (default): abcd emits a prompt, the host's subagent
                                        //   dispatch runs it — no API keys, no model config (adr-25)
                                        // native | cli | api | mcp: opt-in oracle adapters, selected when an
                                        //   operator wants abcd to reach a model directly; unreachable → host-delegated
  },
  "spec": {                             // spec seam — internal/adapter/spec
    "backend": "native"                 // "native" (directory-as-truth + dependency graph, adr-26) | "ccpm"
                                        //   (the companion harness over conventions, adr-24)
  },
  "run": {                              // run seam — internal/adapter/run
    "backend": "native"                 // "native" (thin Go loop, adr-27) | "workflows" | "companion"
                                        //   each iteration gates on a receipt and enforces the safety guard
  },
  "history": {                          // history seam — internal/adapter/history
    "backend": "native"                 // "native" (local redacted transcript store, root-SHA-keyed, adr-29)
                                        //   | "specstory" (opt-in capture source over the same store)
  },
  "scan": {                             // scanner seam — internal/adapter/scanner
    "deep": false                       // native secret/PII scan is the default; deep adds an opt-in TruffleHog
                                        //   backend (also gitleaks when wired), asked only if visibility=private
  },
  "disembark": {
    "maxAgentTokens": 100000            // per-agent context budget; over → stream + summarise
  },
  "embark": {
    "scan": true                        // default: include `embark scan` discovery in onboarding suggestions
  },
  "memory": {                           // memory harvest — read-only source over the native .abcd/memory/ substrate
    "harvest": "native"                 // "native" | "claude" | "<custom-path>" — vendor memory harvest is an
                                        //   opt-in read-only source; see 04-universal-patterns.md § 7
  },
  "reviews": {                          // review-artefact capture — written by whichever oracle adapter runs
    "capture": "oracle"                 // "oracle" (capture over the native review store, adr-25) | "none"
  },
  "dev_sync": {                         // per-source enable flags (asked during ahoy); semantic names per 04-universal-patterns.md § 7
    "reviews": { "enabled": true  },    // oracle-adapter capture → activity/reviews/ — sweeps ad-hoc reviews not tied to a spec only;
                                        // spec-tied reviews land in the native spec review store at write-time (not controlled by this flag)
    "memory":  { "enabled": true  },    // memory harvest → .abcd/memory/
    "work":    { "enabled": true  },    // .abcd/.work.local/ issues + notes/ → .abcd/work/{issues,notes}/
    "rp":      { "enabled": true  }     // RP workspace pull (per itd-7; opt-in RP adapter) → .abcd/rp/
  },
  "intent": {
    "auto_link": true,                  // /abcd:intent plan injects bidirectional link automatically
    "auto_ship": true                   // the native spec-close hook (spc-36) reconciles linked
                                        // intents planned/ → shipped/ on a successful close (spc-28
                                        // intent_lifecycle.reconcile); /abcd:intent "<text>" always lands in
                                        // drafts/ (no auto-trigger)
  },
  "capture": {
    "default_severity": "minor"         // default severity when /abcd:capture omits it (per itd-4)
  },
  "rules": {
    "force_refresh_every_n": 5          // prompt-router-hook re-injects every N prompts (per itd-3)
  },
  "adapters": {                         // wired-adapter registry (internal/registry), refreshed each ahoy —
                                        // records which optional external backends are present per seam
    "repoprompt": { "detected": true }, // an oracle (mcp) backend
    "ccpm":       { "detected": false } // the spec-seam deeper backend (the companion harness over conventions)
  },
  "scout": {
    "issue_scout": { "enabled": false } // opt-in (asked during ahoy)
  }
}
```

**Owed-review draining (spc-43, itd-53) is receipt gating in the `run` seam (adr-27).** Owed fidelity reviews are drained at the `run` seam's iteration boundary: each iteration gates on a **receipt** and applies the safety guard, a report-not-block step whichever adapter provides the loop (native Go loop, Claude Workflows, the companion harness). There is no autodrain config knob and no post-iteration edge to hang it on — receipt gating is part of the seam contract, inherited by every adapter loop rather than re-implemented per loop. The full report-vs-block / cost-bound decision record is [`adr-27`](../../decisions/adrs/0027-autonomous-run-pluggable-seam.md); the companion consistency gate's `RC*` codes are registered in [`06-lint.md § 1`](06-lint.md#1-lint-code-namespace).

**Audit-loop mode + budget — per-intent frontmatter (itd-50, spc-52).** The audit-loop policy is NOT configured in `config.json`; it is elected **per intent** in the intent's own frontmatter, so the choice is portable with the intent (it survives the lifeboat) and one intent can loop while another stays record-only:

```yaml
# intents/<dir>/itd-N-*.md frontmatter
audit_mode: loop-to-acceptance   # "record-only" (default) | "loop-to-acceptance"
audit_budget: 3                  # iteration ceiling for loop-to-acceptance (default 3)
```

- **`audit_mode`** — `record-only` is the default (and the value when the key is **absent**): a `NOT_MET` is recorded to `## Audit Notes`, no re-work. `loop-to-acceptance` makes a `NOT_MET` re-open the linked work and iterate until `MET` or the budget bounds it. Set at plan time.
- **`audit_budget`** — the **declared** iteration ceiling for `loop-to-acceptance`. **Default `3`** when the key is absent but the mode is `loop-to-acceptance` (the intent-grain equivalent of the implementation-grain `MAX_REVIEW_ITERATIONS`). One iteration = one full re-open + re-review cycle that returned `NOT_MET`. A `record-only` intent **ignores** budget entirely (not an error).
- **Budget validation is fail-closed.** A malformed, zero, or negative `audit_budget` on a `loop-to-acceptance` intent is a **policy error**: the loop never starts (recorded fail-closed, issue logged) — it never silently coerces to the default or loops forever.
- **Enqueue-time snapshot.** The review-queue entry **snapshots** the effective `audit_mode` + declared `audit_budget` (plus the `audit_iterations_used` counter and the terminal `audit_outcome`) at enqueue time — the drainer reads the snapshot, never live frontmatter, so a mid-loop edit cannot change an in-flight loop. These four queue-entry fields are additive; a legacy entry without them parses as `record-only`.
- **The loop terminal + the gate live in the drainer/policy layer** (`internal/core/audit`), never in the pure on-close hook. For the full state-coverage table, the `UNACHIEVABLE` replan surface, and the gated manual-verification **receipt schema** (`{intent_id, machine_rollup, state, justification?, recorded_by_role, ts}` under `.abcd/logbook/audit/verify-<ts>/`, distinct from the machine verdict), see [`../04-surfaces/05-intent.md` § 7 Role 1 — the audit loop](../04-surfaces/05-intent.md).

Schema versioning + cross-version migration **comes in a later phase** (abcd stamps `schema_version: 1` everywhere; migrators added if/when a later phase changes the shape — see itd-9).

## The history store

The history store is a **user-scope** artefact, shared across every abcd-managed repo on the machine, living at `~/.abcd/history/`. ahoy bootstraps it once (first `install` on a fresh machine creates it transparently — see [`../04-surfaces/01-ahoy.md`](../04-surfaces/01-ahoy.md)). Layout:

```
~/.abcd/history/
  index.json                  registry — root-SHA → repo entry
  <root-sha>/
    meta.json                 identity + lineage for one repo
    transcripts/              native local redacted transcript store (root-SHA-keyed, adr-29)
    prompt-exports/           oracle-adapter ad-hoc review exports
```

### `index.json`

```json
{
  "schema": 1,
  "description": "abcd history/lifeboat registry. Keyed on each repo's root-commit SHA (immutable under rename, GitHub-handle change, or remote move). Names, GitHub URLs, and paths are mutable labels held in each repo's entry and refreshed by ahoy.",
  "repos": [
    {
      "root_commit": "<sha>",          // immutable key — git rev-list --max-parents=0 HEAD
      "name": "<repo-name>",           // mutable label
      "github": "<git-remote-url>",    // mutable label
      "path": "~/...",                 // mutable label — ahoy REFRESHES this every run if the repo moved
      "status": "active",              // "active" | "superseded"
      "supersedes": "<old-sha>",       // present when this repo re-founded an older one
      "superseded_by": "<new-sha>"     // present on the old entry after re-founding
    }
  ]
}
```

`root_commit` is the only immutable field — it survives rename, remote move, and GitHub-handle change. `name`, `github`, and `path` are mutable labels: ahoy refreshes them on every run rather than treating them as write-once.

### `meta.json` (per `<root-sha>/`)

Identity and lineage for one repo. Beyond the obvious identity fields:

- **`aliases`** — array of prior names the repo has had (e.g. a repo renamed on GitHub).
- **`note`** — free-text provenance: why the repo was re-founded, where a backup of the full pre-refounding tree lives, what the old git history carries.
- **`corpus`** — paths (relative to the `<root-sha>/` dir) of the captured evidence: `{"transcripts": "transcripts/", "prompt_exports": "prompt-exports/"}`.

Re-founding (clean-history rebuild — the case in [`../research/notes/ahoy-history-store-manual-scaffolding.md`](../../research/notes/ahoy-history-store-manual-scaffolding.md)) produces a *new* root SHA. ahoy registers the new entry with `supersedes` → old SHA, marks the old entry `superseded_by` → new SHA, and leaves the old corpus in place under its own `<root-sha>/` dir for lifeboat review.

> **Legacy:** `~/ABCDevelopment/.abcd/changelog.md` is a hand-maintained toolchain changelog that predates abcd's `.abcd/` namespace. It is **not** an ahoy-managed artefact and is not part of the `.abcd/` namespace defined below.

The history-store `index.json` is the sole user-scope registry — it records each managed repo's identity and lineage keyed on the immutable `root_commit`; `name`, `github`, and `path` are mutable labels ahoy refreshes on every run. There is no separate `workspaces.json`: abcd manages one repository per tree (see § The two `.abcd/` scopes below), so there is no workspace↔repo grouping to register.

## The two `.abcd/` scopes

`.abcd/` is **one namespace pattern instantiated at two scopes**. abcd lives in **one repository** ([adr-28](../../decisions/adrs/0028-single-repo-curated-release.md)): its design record is **repo-scoped and in-tree**, and the user scope holds only state that is genuinely machine-wide. `/abcd:ahoy` classifies the folder it runs in (see [`../04-surfaces/01-ahoy.md`](../04-surfaces/01-ahoy.md)) and acts on the scope that applies:

| Scope | Location | Holds |
|---|---|---|
| **user** | `~/.abcd/` | one per machine — **machine-local shared state only**: the root-SHA-keyed `history/` store (`index.json` + per-root-SHA transcript corpus, [adr-29](../../decisions/adrs/0029-native-transcript-corpus.md)), machine `config.json` defaults, and the user-scope `memory/` (personal, cross-project knowledge). **Never the design record.** |
| **repo** | in-tree `.abcd/` | this repository's record and working files — the three-tier layout below, plus `config.json` (with its `meta` setup block), `rules.json`, and the `memory/`, native spec store, `lifeboat/`, `logbook/`, `rp/` namespaces. **The home for project work.** |

**The repo-scope three-tier working layout** (matching [`../02-constraints/01-platform.md`](../02-constraints/01-platform.md) and [`../01-product/02-context.md`](../01-product/02-context.md)):

| Tier | Path | Committed? | Holds |
|---|---|---|---|
| **record** | `.abcd/development/` | committed — excluded from the release artefact by packaging | the durable design record: brief, roadmap, intents, ADRs, research |
| **shared work** | `.abcd/work/` | committed | shared working files — `CONTEXT.md` + `DECISIONS.md` |
| **local ephemeral** | `.abcd/.work.local/` | gitignored | machine-local scratch — `NEXT.md`, `scratch/`, `logs/` |

**The record is repo-scoped and in-tree — no workspace layer holds it, and there is no `workspaces.json`.** abcd is one repository, so the record lives in that repository's tree; there is no dev→public mirror and no workspace registry. The user scope survives only for state that cannot live in any one repo's tree because it is shared across every abcd-managed repo on the machine: the `history/` store keyed on each repo's root-commit SHA, and machine `config.json` defaults. The repo keeps its `config.json` (carrying the `meta` setup block) and `rules.json` in-tree too — the Claude Code hook and the marker-block installer read them from the repo directory deterministically.

**The `~/.claude/` boundary.** `~/.claude/` is the vendor harness directory. abcd keeps it minimal — **only the abcd plugin install lives there.** No abcd-specific material is written under `~/.claude/`; it routes to the scope-appropriate `.abcd/` instead. The one interaction abcd has with `~/.claude/` is *read-only*: `dev-sync memory` harvests `~/.claude/projects/<encoded-cwd>/memory/` as a source (see [`02-adapters.md`](02-adapters.md)). abcd never writes there.

## 1. Visibility-driven gitignore policy

Set by ahoy:

| Directory | Public default | Private default |
|---|---|---|
| `.abcd/` | gitignored | **committed** (entire namespace: `development/` (brief, roadmap, research, activity, voyage, personas), the native spec store, `memory/`, `lifeboat/`, `logbook/`, `rp/` — visibility is the single switch, no per-subdirectory exceptions) |
| `memory/` (legacy snapshot) | gitignored | **committed** if present¹ |
| `.abcd/.work.local/` | gitignored | gitignored (local-only scratch, per global abcd CLAUDE.md) |

The native local transcript store is **always** gitignored (user-scope `~/.abcd/history/`, local working data — adr-29), so it is not a repo directory in this table.

¹ New projects use `.abcd/memory/` (curated by `dev-sync memory`). `memory/` is the legacy `cp -r` snapshot pattern that some existing projects maintain manually — abcd respects it if present, but doesn't write to it.

**No exceptions to the visibility rule.** Earlier drafts of this brief carved out `.abcd/logbook/` as always-gitignored (sensitivity concern). Locked decision: visibility is **one switch**. If sensitivity is a concern, set visibility=public (which gitignores all of `.abcd/` including logbook). Per-subdirectory exceptions create maintenance burden and contradict the transparent-prompts principle ([`04-universal-patterns.md § 1`](04-universal-patterns.md#1-transparent-prompts)).

**Sensitivity concern still valid for `/abcd:launch` payload**: regardless of visibility, the launch payload manifest ([`../04-surfaces/04-launch.md § 2`](../04-surfaces/04-launch.md#2-payload-manifest-default-deny)) excludes `.abcd/` entirely from what ships in the curated release artifact. So a private repo that commits its logbook locally still doesn't leak it on launch.

In private repos, the entire `.abcd/` namespace is reproducible from a fresh clone — embark from a freshly cloned repo works including logbook (useful for diagnosing past command runs).

**Memory locations to keep straight:**

abcd's curated memory exists at **two scopes** (per § The two `.abcd/` scopes), and there is one non-abcd memory location alongside it:

1. **`.abcd/memory/`** (repo scope) — the **primary** abcd memory: curated semantic summaries written by `dev-sync memory`, tracked in private repos, the canonical input for `principle-distiller`. Most memory lives here.
2. **`~/.abcd/memory/`** — **user-scope** memory: personal preferences and cross-project principles that have no single repo home.
3. **`memory/`** — legacy `cp -r` snapshot at the repo root that some existing projects maintain. abcd respects if present but doesn't write to it.

Which scope a curated page lands in is a routing decision — see [`07-memory.md`](07-memory.md) § scope routing. Retrieval across the two scopes is **not** a flat union (that would overflow context); it is keyword-recall + budget-bracketed injection per itd-39. When the brief says "memory" without qualification, it means the repo-scope `.abcd/memory/`.

**`.abcd/.work.local/` is local-only everywhere.** Working notes, drafts, status trackers stay gitignored. abcd consumes them via `dev-sync` ([§ 2](#2-abcdwork-namespace-and-dev-sync)) which promotes useful content into tracked `.abcd/work/` artefacts before disembark.

**AI transparency level** (separate axis from visibility — set by ahoy via `ai_transparency.level`):

| Level | Conversations (transcripts) | Plans | Tasks | Metadata (sessions, actions) |
|---|---|---|---|---|
| `full` | yes | yes | yes | yes |
| `metadata` | no | yes | yes | yes |
| `none` | no | no | no | no |

**Why separate from visibility:** a private repo might want `metadata`-only transparency to keep storage tight; a public OSS project might want `full` for credibility. Visibility decides what's *committed*; transparency decides what's *captured at all*.

Drives behaviour:
- **`dev-sync`** — `none` skips capture entirely; `metadata` skips conversation transcripts; `full` captures everything per source enable flags
- **`disembark`** — chat-distiller (Pass B) is no-op on `none`; runs on metadata/full
- **`launch` payload** — `ai_transparency` value carried into the public `marketplace.json` so consumers know what to expect

Sanitised export pattern (lifted from `~/.abcd/`'s `/export-transparency`): launch's pre-flight scrub strips absolute paths and PII regardless of transparency level. The transparency level determines what *exists*; pre-flight ensures what *ships* is sanitised.

**Visibility × transparency interaction (added post-audit 2026-05-07):** the two axes are independent. Any combination is valid: `private × none` (paranoid, no captures, no commit), `private × full` (everything captured locally, nothing public), `public × none` (public repo, no AI captures committed), `public × full` (everything captured AND committed — useful for OSS-credibility OSS projects). The launch payload's exclusion rules (per [`04-surfaces/04-launch.md § 2`](../04-surfaces/04-launch.md)) are unconditional regardless of `ai_transparency.level` — launch always ships the visibility-determined payload, and `ai_transparency` only governs what was captured in the first place.

## 2. `.abcd/work/` namespace and `dev-sync`

`.abcd/work/` is the **curated-from-volatile-sources** namespace. Volatile inputs (gitignored or external) get analysed and promoted into tracked `.abcd/work/` artefacts via `abcd dev-sync`. This solves three problems: noisy sources stay gitignored; curated lessons get tracked; abcd doesn't have to read volatile sources every time.

**Source → target table:**

| Volatile source (gitignored or external) | Curated abcd target (tracked in private repos) | Adapter |
|---|---|---|
| Agent memory (opt-in harvest per [`04-universal-patterns.md § 7`](04-universal-patterns.md#7-vendor-agnostic-adapters-with-environment-branching) — Claude Code: `~/.claude/projects/<encoded-cwd>/memory/`) | `.abcd/memory/` | memory harvest (`internal/core/memory`) |
| Ad-hoc reviews not tied to a spec (captured by whichever oracle adapter runs per [`04-universal-patterns.md § 7`](04-universal-patterns.md#7-vendor-agnostic-adapters-with-environment-branching) — e.g. RepoPrompt's local chat store; vendor paths in [`02-adapters.md`](02-adapters.md)). **Spec-tied reviews are written directly to the native spec review store at review time — `dev-sync reviews` does NOT sweep those.** | `.abcd/work/reviews/` | oracle-adapter capture (`internal/core/reviews`) |
| `.abcd/.work.local/issues.md` | `.abcd/work/issues/{open,resolved,wontfix}/iss-N-<slug>.md` (per itd-4) | workdir capture (`internal/core/workdir`; migration on first sync after install) |
| `.abcd/.work.local/notes/`, `.abcd/.work.local/<feature>/` | `.abcd/work/notes/` | workdir capture (`internal/core/workdir`) |
| RepoPrompt workspace state (opt-in adapter; vendor paths in [`02-adapters.md`](02-adapters.md)) | `.abcd/rp/workspace.json` (per itd-7) | RP workspace adapter (`internal/adapter/oracle`, opt-in) |

**`abcd dev-sync` triggers:**

- **Implicit:** `/abcd:disembark` Phase 0 runs `dev-sync` automatically (always-fresh-at-disembark)
- **Manual:** `abcd dev-sync` CLI for ad-hoc refresh

Scheduled/cron sync **comes in a later phase** of the plugin (itd-13).

**Per-source on/off:**

- `.abcd/config.json` extends with `dev_sync.{reviews,memory,work,rp}.enabled = true|false`
- ahoy asks per-source enable (transparent prompts)
- defaults: all sources on for private, all sources off for public
- disembark Phase 0 honours per-source flags

**Per-source provenance and curation rules:**

- **Memory (volatile) → `.abcd/memory/` (curated):** Source is an opt-in memory harvest per [`04-universal-patterns.md § 7`](04-universal-patterns.md#7-vendor-agnostic-adapters-with-environment-branching) — under Claude Code: `~/.claude/projects/<encoded-cwd>/memory/`. The repo-local legacy `memory/` snapshot (the `cp -r` pattern) is the workflow `dev-sync memory` replaces. Output is *not verbatim*: distilled summaries grouped by domain, written as actionable suggestions for future agents (e.g., "When implementing UI hit areas, always use `.contentShape(Rectangle())` — source: `feedback_hit_target_full_box`"). Why curated: raw memories grow unbounded and contain personal phrasing ("user got annoyed when X"). Inputs to `principle-distiller` (Pass C).

- **Reviews (volatile) → `.abcd/work/reviews/` (curated):** Reviews are captured by whichever oracle adapter runs per [`04-universal-patterns.md § 7`](04-universal-patterns.md#7-vendor-agnostic-adapters-with-environment-branching) — host-delegated by default ([adr-25](../../decisions/adrs/0025-host-delegated-llm-default.md)), with **RepoPrompt** as one opt-in adapter. When the RepoPrompt adapter is wired, `dev-sync reviews` harvests **ad-hoc oracle reviews not tied to a spec** from RepoPrompt's local chat store; **spec-tied reviews are NOT swept here** — the native spec review store captures them directly at review time. See [`02-adapters.md`](02-adapters.md) for the adapter's harvesting detail (vendor paths, the prompt-exports redirect, workspace matching, and the stability/privacy safeguards). `dev-sync reviews` renders its sources → `.abcd/work/reviews/oracle-{review,chat}-<timestamp>-<description>-<hash>.md` (the format `review-collator` consumes). Dedup by content hash; idempotent. Inputs to `review-collator` (Pass A).

- **`.abcd/.work.local/` (volatile, local-only) → `.abcd/work/issues/`, `.abcd/work/notes/` (curated):** `.abcd/.work.local/issues.md` (the abcd CLAUDE.md mandatory issue log) gets parsed entry-by-entry; each entry promoted to `.abcd/work/issues/open/iss-N-<slug>.md` (per itd-4 ledger structure). `.abcd/.work.local/notes/`, `.abcd/.work.local/<feature>/` get distilled into `.abcd/work/notes/`. Files in `.abcd/.work.local/` are never moved or deleted — `dev-sync work` is read-and-curate, source stays put. Inputs to `principle-distiller` (Pass C) and `chat-distiller` (Pass B, as auxiliary context).

- **RP workspace state (volatile) → `.abcd/rp/workspace.json` (curated, per itd-7):** The opt-in RepoPrompt adapter pulls RepoPrompt's own workspace state (the workspace whose root path matches the current repo) into `.abcd/rp/workspace.json`. The vendor filesystem layout and the match/normalisation mechanics live with the adapter — see [`02-adapters.md`](02-adapters.md). Workspace.json only for now; presets, mcp-routing scoping, `--preset <name>` flag, and `abcd rp link` window helper come in a later phase.

**Reviews as a first-class pitfall source:**

Plan/implementation/completion reviews (the ones in `.abcd/work/reviews/`) are *exceptionally* useful for spotting issues. The `review-collator` agent must extract every "P0 / P1 / watch out for X / found bug" finding as a candidate pitfall — **even when the original issue was fixed**, the lesson survives. Output:

- `reviews-consolidated.json` — full review summaries (existing)
- `candidate-pitfalls.json` — extracted findings ready for distiller dedup

`principle-distiller` (Pass C) has four pitfall sources to dedupe by topic-hash or canonical phrasing: source `memory/pitfalls.md` (or `.abcd/memory/pitfalls.md` after curation), `candidate-pitfalls.json`, Pass B chat-distiller deltas, and code-rescuer's `code-principles.json`.

**Distinct from `principles.json`:** `.abcd/memory/`, `.abcd/work/reviews/`, `.abcd/work/issues/`, `.abcd/work/notes/` are **persistent rolling artefacts** in the source repo, refreshed by `dev-sync`, used as ongoing input to future agents. `principles.json` is **per-disembark synthesis** written into the lifeboat at `.abcd/lifeboat/principles.json`. The lifeboat consumes `.abcd/work/`; `.abcd/work/` is not the lifeboat.

## 3. Plugin shape — directory layout

**Repository layout** (a Go binary plus the markdown plugin surface that shells to it):

```
abcd/
├── .claude-plugin/plugin.json
├── .claude-plugin/marketplace.json
├── README.md
├── go.mod / go.sum
├── cmd/
│   └── abcd/main.go                    # entrypoint — wires the CLI front door to the core
├── internal/
│   ├── core/                           # transport-agnostic core — one package per capability (adr-23)
│   │   └── …                           #   intent, capture, memory, lint, reflect, docfidelity,
│   │                                   #   render, schema, provenance, … — each returns structured results
│   ├── adapter/                        # the five seams — each: interface + native default + optional plug-in (adr-22)
│   │   ├── oracle/                     # host-delegated default; native | cli | api | mcp backends (adr-25)
│   │   ├── history/                    # native transcript store; specstory import (adr-29)
│   │   ├── spec/                       # native minimal store; the companion harness ccpm over conventions (adr-26)
│   │   ├── run/                        # thin native loop; Claude Workflows / the companion harness's loop (adr-27)
│   │   └── scanner/                    # native secret/PII scan; gitleaks / TruffleHog
│   ├── registry/                       # wired-adapter registry — resolves <seam>.backend to an implementation
│   └── surface/
│       ├── cli/                        # Cobra front door (ships in the MVP)
│       └── mcp/                        # MCP front door (later)
├── commands/abcd/                      # markdown command surfaces that shell to the binary — canonical list in ../04-surfaces/README.md
│   ├── ahoy.md / disembark.md / embark.md / launch.md / intent.md / capture.md / memory.md
│   └── …                               # plus operator-internal commands
│   # NOTE: `uninstall` is a sub-verb of /abcd:ahoy (not a standalone command). The ahoy command
│   # markdown handles the install/uninstall/dry-run/destroy sub-verb dispatch internally.
├── skills/                             # see 08-skills.md — abcd ships ZERO user-facing skills.
│   #                                   # The entries below are plugin-runtime workflow files that each
│   #                                   # command points at internally; they are NOT user-facing skills
│   #                                   # surfaced under /abcd:. The only would-be user-facing skill
│   #                                   # (/abcd:grill) is a /abcd:intent grill sub-verb. A later phase
│   #                                   # may introduce new user-facing skills here.
│   ├── abcd-ahoy/{SKILL.md, workflow.md}
│   ├── abcd-disembark/{SKILL.md, workflow.md}
│   ├── abcd-embark/{SKILL.md, workflow.md}
│   ├── abcd-launch/{SKILL.md, workflow.md}
│   ├── abcd-intent/{SKILL.md, workflow.md}
│   ├── abcd-capture/{SKILL.md, workflow.md}
│   ├── commit-attribution/SKILL.md
│   └── secrets-and-pii/SKILL.md        # consolidated pii-protection + secret-scan
├── agents/                             # 16 agents — see 01-agents.md (markdown, host-delegated)
│   ├── flow-essence.md / decision-archaeologist.md / review-collator.md / chat-distiller.md
│   ├── principle-distiller.md / artefact-curator.md / brief-composer.md / press-release-composer.md
│   ├── lifeboat-oracle.md / code-rescuer.md / issue-scout.md / embark-scaffolder.md
│   ├── launch-gatekeeper.md / intent-fidelity-reviewer.md / documentation-auditor.md
│   └── reflection-composer.md          # /abcd:reflect retrospective composer (itd-24)
└── hooks/                              # Claude Code event hooks — thin shims that shell to the binary
    ├── hooks.json                      # UserPromptSubmit → prompt-router; SessionStart / PreCompact → reset
    ├── prompt_router_hook              # CARL-style rule injector (per itd-3); reads .abcd/rules.json + plugin defaults
    └── prompt_router_reset             # per-session dedup-state reset (SessionStart / PreCompact)
```

The core is organised one package per capability under `internal/core/`, and the
five adapter seams under `internal/adapter/`. [`02-adapters.md`](02-adapters.md)
owns the seam catalogue — each seam's interface, native default, and optional
external plug-in — so this brief does not restate it here.

**Plugin-internal development namespace** (committed in private repos, gitignored in public):

```
.abcd/
├── config.json                         # config + the `meta` setup block (schema_version, setup_version, ...)
├── corpus.json
├── rules.json                          # per-repo override of plugin-bundled rule defaults (per itd-3)
├── development/
│   ├── personas.json                   # alphabetical placeholder personas (Alice, Bob, Carol, ...)
│   ├── brief/
│   │   └── README.md                   # canonical, current-state (no archive — per adr-5)
│   ├── roadmap/
│   │   ├── README.md                   # status dashboard
│   │   ├── intents/
│   │   │   ├── README.md               # intent format + lifecycle + index
│   │   │   ├── drafts/                 # itd-N-<slug>.md (captured intents, no plan yet)
│   │   │   ├── planned/                # itd-N-<slug>.md (has linked native spec, work pending or in flight)
│   │   │   └── shipped/                # populated as linked specs close + intent-fidelity-reviewer runs
│   │   ├── phases/
│   │   │   ├── README.md               # phase index
│   │   │   └── phase-N-<slug>.md       # ordered build plan; each ends in a milestone (per adr-9)
│   │   └── rfcs/
│   │       ├── README.md
│   │       └── rfc-N-<slug>.md         # community discussion artefacts (open / resolved-yes / resolved-no / ...)
│   ├── research/
│   │   ├── prompting/
│   │   │   ├── 01-general-best-practices.md   # SOTA baseline (early spec)
│   │   │   └── agents/<name>.md (×15)         # per-agent SOTA research (task #1 of each agent spec)
│   │   ├── phase/                             # per-phase study artefacts (design inputs that future phases consume)
│   │   │   ├── 0/                             # Phase 0: predecessor-notes, transcript-sampling, idelphi-rescue-study
│   │   │   ├── 1/                             # Phase 1+: as needed
│   │   │   └── ...
│   │   └── adr/                               # architecture decision records (e.g., 01-harness-interface.md)
│   ├── activity/                       # curated-from-volatile-sources
│   │   ├── reviews/                    # captured by the oracle adapter (RepoPrompt / codex / future) per 04-universal-patterns.md § 7
│   │   ├── issues/{open,resolved,wontfix}/  # iss-N-<slug>.md ledger entries (per itd-4)
│   │   └── notes/                      # distilled from .abcd/.work.local/notes/
│   └── voyage/                         # embark/disembark provenance and history (see ../04-surfaces/03-embark.md § 7)
│       ├── README.md
│       ├── embark/
│       │   ├── provenance.json         # source path + manifest hash + timestamp + files written
│       │   └── from/<timestamp>/       # opt-in via embark --archive: verbatim copy of input lifeboat
│       └── disembark/
│           └── history.jsonl           # append-only manifest log of every disembark run
├── specs/                              # native minimal spec store — directory-as-truth (adr-26): <state>/ dirs + dependency graph
├── memory/                             # curated memory artefact (memory harvest → .abcd/memory/, per 04-universal-patterns.md § 7)
├── lifeboat/                           # disembark output snapshot only — regenerable, overwritten each run (per ../02-constraints/01-platform.md, ../04-surfaces/03-embark.md § 7)
├── logbook/                            # per-command / per-phase run logs (design target — no automatic session-log hook ships)
└── rp/                                 # RP workspace pull (per itd-7; opt-in RP adapter); workspace.json only for now
```

**User-facing docs** (packaged into the curated release artifact):

```
docs/
├── README.md
├── tutorials/                          # learning guides
├── guides/                             # task-oriented how-tos
├── reference/                          # command reference, config schemas
└── explanation/                        # conceptual: lifeboats, dev-sync, intents, capture, etc.
```

**Doc framework note**: the native spec store, memory, and config live under `.abcd/`. Plugin-internal design docs live under `.abcd/development/`. User-facing docs live under `docs/`. We use the *shape* of a planning-vs-roadmap-vs-process split, not any one tool's *location* convention.
