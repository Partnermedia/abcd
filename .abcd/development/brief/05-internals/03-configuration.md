# Configuration Model

This file holds the configuration schema (`meta.json`, `config.json`), the visibility-driven gitignore policy, and the `dev-sync` namespace that pumps volatile sources into curated artefacts.

## `.abcd/meta.json`

(unchanged from the original cut):

```json
{
  "schema_version": 1,
  "setup_version": "0.1.0",
  "setup_date": "2026-05-04",
  "project_name": "abcdDev"
}
```

## `.abcd/config.json`

(expanded — example values shown; ahoy populates from prompts on first run, no silent default for `visibility`):

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
  "oracle": {
    "backend": "auto"                   // "rp" | "codex" | "in-session" | "auto"
                                        // auto: cascades RP MCP → Codex CLI → in-session subagent
                                        // explicit: pins the choice
                                        // Cross-model perspective via RP routing (user configures models inside RepoPrompt)
  },
  "scan": {
    "deep": false                       // TruffleHog toggle, asked only if visibility=private
  },
  "disembark": {
    "maxAgentTokens": 100000            // per-agent context budget; over → stream + summarise
  },
  "embark": {
    "scan": true                        // default: include `embark scan` discovery in onboarding suggestions
  },
  "memory": {
    "backend": "auto"                   // "auto" | "claude" | "opencode" | "<custom-path>" — see 04-universal-patterns.md § 7 dispatcher
  },
  "reviews": {
    "backend": "auto"                   // "auto" | "repoprompt" | "codex" | "<custom-path>" — see 04-universal-patterns.md § 7 dispatcher
  },
  "review": {                           // review-QUEUE drain knobs (fn-43, itd-53) — singular "review", a DIFFERENT
                                        // block from "reviews" (backend dispatcher) above, and a DIFFERENT FILE from
                                        // the .flow/config.json "review" block (body_max_bytes/render_max_bytes,
                                        // read by flow-next's _review_lib). Read via config_io.read_config — the
                                        // safe-read pattern: absent file / absent block / absent key all mean OFF,
                                        // never a crash on the default path.
    "autodrain": false,                 // default OFF (opt-in). true → the Ralph POST-ITERATION edge drains owed
                                        // fidelity reviews via `flowctl review-queue autodrain`; absent/false →
                                        // drain() never fires. Trigger boundary + cost contract: see adr-16.
    "autodrain_max_reviews": 1          // per-firing cap on PROCESSED queue entries (not successes); default 1;
                                        // malformed / bool / < 1 falls back to 1
  },
  "dev_sync": {                         // per-source enable flags (asked during ahoy); semantic names per 04-universal-patterns.md § 7
    "reviews": { "enabled": true  },    // reviews backend (vendor-detected) → activity/reviews/ — sweeps ad-hoc chats not tied to a spec only;
                                        // spec-tied reviews land in .flow/reviews/ via fn-2 Stop hook at write-time (not controlled by this flag)
    "memory":  { "enabled": true  },    // memory backend (vendor-detected) → .abcd/memory/
    "work":    { "enabled": true  },    // .work/issues.md + notes/ → activity/{issues,notes}/
    "rp":      { "enabled": true  }     // RP workspace pull (per itd-7) → .abcd/rp/
  },
  "intent": {
    "auto_link": true,                  // /abcd:intent plan injects bidirectional link automatically
    "auto_ship": true                   // the flowctl `spec close` close-hook (fn-36) reconciles linked
                                        // intents planned/ → shipped/ on a successful close (fn-28
                                        // intent_lifecycle.reconcile); /abcd:intent "<text>" always lands in
                                        // drafts/ (no auto-trigger)
  },
  "capture": {
    "default_severity": "minor"         // default severity when /abcd:capture omits it (per itd-4)
  },
  "rules": {
    "force_refresh_every_n": 5          // prompt-router-hook re-injects every N prompts (per itd-3)
  },
  "plugins": {                          // detection cache, refreshed each ahoy
    "flow_next": { "detected": true },
    "repo_prompt": { "detected": true }
  },
  "scout": {
    "issue_scout": { "enabled": false } // opt-in (asked during ahoy)
  }
}
```

**`review.autodrain` (fn-43, itd-53):** the opt-in drainer for owed fidelity reviews. When `true`, the Ralph post-iteration edge — and ONLY that edge; the Claude Code SessionEnd/Stop hook is a recorded non-goal because oracle dispatch exceeds the 5s hook timeout — invokes the dedicated `flowctl review-queue autodrain` verb, which calls the existing `drain()` capped by `review.autodrain_max_reviews` (default 1, counting *processed* entries, not successes). Off-path cost contract: with autodrain off, the edge still runs the verb once per iteration as a single cheap config-read subprocess that early-exits 0 — "never fires" is a guarantee about `drain()`, not about the subprocess. The loop-body edit is overlay-managed (`patch:ralph-autodrain-trigger` — see [`scripts/abcd/overlay/README.md`](../../../../scripts/abcd/overlay/README.md)) so it survives `/flow-next:ralph-init` re-vendor. The full boundary / report-vs-block / cost-bound decision record is [`adr-16`](../../decisions/adrs/adr-16-fn43-autodrain-boundary-and-gate-defaults.md); the companion consistency gate's `RC*` codes are registered in [`06-lint.md § 1`](./06-lint.md#1-lint-code-namespace).

**Audit-loop mode + budget — per-intent frontmatter (itd-50, fn-52).** The audit-loop policy is NOT configured in `config.json`; it is elected **per intent** in the intent's own frontmatter, so the choice is portable with the intent (it survives the lifeboat) and one intent can loop while another stays record-only:

```yaml
# intents/<dir>/itd-N-*.md frontmatter
audit_mode: loop-to-acceptance   # "record-only" (default) | "loop-to-acceptance"
audit_budget: 3                  # iteration ceiling for loop-to-acceptance (default 3)
```

- **`audit_mode`** — `record-only` is the default (and the value when the key is **absent**): a `NOT_MET` is recorded to `## Audit Notes`, no re-work. `loop-to-acceptance` makes a `NOT_MET` re-open the linked work and iterate until `MET` or the budget bounds it. Set at plan time.
- **`audit_budget`** — the **declared** iteration ceiling for `loop-to-acceptance`. **Default `3`** when the key is absent but the mode is `loop-to-acceptance` (the intent-grain equivalent of the implementation-grain `MAX_REVIEW_ITERATIONS`). One iteration = one full re-open + re-review cycle that returned `NOT_MET`. A `record-only` intent **ignores** budget entirely (not an error).
- **Budget validation is fail-closed.** A malformed, zero, or negative `audit_budget` on a `loop-to-acceptance` intent is a **policy error**: the loop never starts (recorded fail-closed, issue logged) — it never silently coerces to the default or loops forever.
- **Enqueue-time snapshot.** The review-queue entry **snapshots** the effective `audit_mode` + declared `audit_budget` (plus the `audit_iterations_used` counter and the terminal `audit_outcome`) at enqueue time — the drainer reads the snapshot, never live frontmatter, so a mid-loop edit cannot change an in-flight loop. These four queue-entry fields are additive; a legacy entry without them parses as `record-only`.
- **The loop terminal + the gate live in the drainer/policy layer** (`audit_loop_policy.py`, `verification_receipt.py`), never in the pure on-close hook. For the full state-coverage table, the `UNACHIEVABLE` replan surface, and the gated manual-verification **receipt schema** (`{intent_id, machine_rollup, state, justification?, recorded_by_role, ts}` under `.abcd/logbook/audit/verify-<ts>/`, distinct from the machine verdict), see [`../04-surfaces/05-intent.md` § 7 Role 1 — the audit loop](../04-surfaces/05-intent.md).

Schema versioning + cross-version migration **comes in a later phase** (abcd stamps `schema_version: 1` everywhere; migrators added if/when a later phase changes the shape — see itd-9).

## The history store

The history store is a **user-scope** artefact, shared across every abcd-managed repo on the machine, living at `~/.abcd/history/`. ahoy bootstraps it once (first `install` on a fresh machine creates it transparently — see [`../04-surfaces/01-ahoy.md`](../04-surfaces/01-ahoy.md)). Layout:

```
~/.abcd/history/
  index.json                  registry — root-SHA → repo entry
  <root-sha>/
    meta.json                 identity + lineage for one repo
    specstory/                live transcripts (SpecStory output_dir points here)
    prompt-exports/           RepoPrompt ad-hoc oracle reviews
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
- **`corpus`** — paths (relative to the `<root-sha>/` dir) of the captured evidence: `{"specstory": "specstory/", "prompt_exports": "prompt-exports/"}`.

Re-founding (clean-history rebuild — the case in [`../research/notes/ahoy-history-store-manual-scaffolding.md`](../research/notes/ahoy-history-store-manual-scaffolding.md)) produces a *new* root SHA. ahoy registers the new entry with `supersedes` → old SHA, marks the old entry `superseded_by` → new SHA, and leaves the old corpus in place under its own `<root-sha>/` dir for lifeboat review.

> **Legacy:** `~/ABCDevelopment/.abcd/changelog.md` is a hand-maintained toolchain changelog that predates abcd's `.abcd/` namespace. It is **not** an ahoy-managed artefact and is not part of the `.abcd/` namespace defined below.

## The workspace registry — `~/.abcd/workspaces.json`

`~/.abcd/workspaces.json` (user scope) is the registry of **everything abcd manages and where** — the human-structure counterpart to the root-SHA-keyed history `index.json`. ahoy reads it to classify the current folder and to resolve user-scope state without a hardcoded path or directory walk (per itd-40).

```json
{
  "schema": 1,
  "entries": [
    {
      "kind": "workspace",             // "workspace" | "repo"
      "name": "abcd",
      "path": "~/ABCDevelopment/Apps/abcd",
      "repos": ["abcd", "abcdDev"]      // workspace only — names of member repo entries
    },
    {
      "kind": "repo",
      "name": "abcdDev",
      "path": "~/ABCDevelopment/Apps/abcd/abcdDev",
      "workspace": "abcd",             // repo only — parent workspace name, or null for standalone
      "root_commit": "<sha>"           // repo only — REFERENCE into history index.json; not duplicated identity
    }
  ]
}
```

**Two registries, one fact each — the cross-reference rule:**

| Registry | Scope | Holds | Keyed on |
|---|---|---|---|
| `~/.abcd/workspaces.json` | user | **human structure** — workspace↔repo groupings, mutable paths, `kind` tag | `name` (mutable label) |
| `~/.abcd/history/index.json` | user | **identity + lineage** — name, github, status, `supersedes` | `root_commit` (immutable) |

A repo entry in `workspaces.json` carries **only a `root_commit` reference** into `index.json`; it never copies `github`, `status`, or lineage. Identity lives in `index.json`; structure lives in `workspaces.json`. This keeps the two from drifting — `path` is mutable and refreshed in both by ahoy, but every other identity field has exactly one home.

If `~/.abcd/workspaces.json` does not exist, the first `/abcd:ahoy install` creates it (alongside the `~/.abcd/history/` store). See [`../04-surfaces/01-ahoy.md`](../04-surfaces/01-ahoy.md) for the classification mechanics that consume it.

## The two `.abcd/` scopes

`.abcd/` is **one namespace pattern instantiated at two scopes**, not a repo-only directory. `/abcd:ahoy` classifies the folder it runs in (see [`../04-surfaces/01-ahoy.md`](../04-surfaces/01-ahoy.md)) and acts on the scope that applies:

| Scope | Location | Holds |
|---|---|---|
| **user** | `~/.abcd/` | one per machine — the `workspaces.json` registry, the shared `history/` store (`index.json` + per-root-SHA corpus), `config.json`, and the user-scope `memory/` (personal, cross-project knowledge) |
| **workspace** | `<workspace>/.abcd/` | per workspace — `development/` (brief, roadmap, research, activity, voyage, personas), `memory/`, `lifeboat/`, `logbook/`, `rp/`. **The primary home for project work.** |

**There is no development-environment scope.** A folder like `~/ABCDevelopment/` is just where a user happens to keep repos — abcd does not privilege it or hardcode it. Anything that was previously "development-wide" is **user-scoped**: it lives under `~/.abcd/`, one per machine. The shared history store and `index.json` live there; `~/.abcd/` is the single user-scope root.

**There is no repo-scope `.abcd/` — with four named exceptions.** A repo (a single repository, possibly the dev repo inside a workspace) keeps only what *must* be physically in-repo: the CLAUDE.md/AGENTS.md marker block, the gitignored `.specstory/cli/config.toml` redirect shim, and the per-repo `<repo>/.abcd/rules.json` (modular rules loader override; see fn-14) plus `<repo>/.abcd/config.json` (per-repo lint and loader config; the modular loader reads `rules.force_refresh_every_n` and `docs.target` from it). All four are explicit carve-outs from the "no repo-scope `.abcd/`" rule — physically in-repo because the Claude Code hook and the marker-block installer must read them from the repo directory deterministically without resolving workspace inheritance. Everything else a repo would need is workspace-scoped and inherited. Earlier drafts of this brief located `development/`, `memory/`, `lifeboat/`, `logbook/`, `rp/` at the repo — they are **workspace-scoped**; wherever the rest of this file says "repo" for those artefacts, read "workspace".

**The `~/.claude/` boundary.** `~/.claude/` is the vendor harness directory. abcd keeps it minimal — **only the abcd plugin install lives there.** No abcd-specific material is written under `~/.claude/`; it routes to the scope-appropriate `.abcd/` instead. The one interaction abcd has with `~/.claude/` is *read-only*: `dev-sync memory` harvests `~/.claude/projects/<encoded-cwd>/memory/` as a source (see [`02-adapters.md`](./02-adapters.md)). abcd never writes there.

## 1. Visibility-driven gitignore policy

Set by ahoy:

| Directory | Public default | Private default |
|---|---|---|
| `.abcd/` | gitignored | **committed** (entire namespace: `development/` (brief, roadmap, research, activity, voyage, personas), `memory/`, `lifeboat/`, `logbook/`, `rp/` — visibility is the single switch, no per-subdirectory exceptions) |
| `.flow/` | gitignored | **committed** |
| `.specstory/` | gitignored | **committed** |
| `memory/` (legacy snapshot) | gitignored | **committed** if present¹ |
| `.work/` | gitignored | gitignored (local-only scratch, per global abcd CLAUDE.md) |

¹ New projects use `.abcd/memory/` (curated by `dev-sync memory`). `memory/` is the legacy `cp -r` snapshot pattern that some existing projects maintain manually — abcd respects it if present, but doesn't write to it.

**No exceptions to the visibility rule.** Earlier drafts of this brief carved out `.abcd/logbook/` as always-gitignored (sensitivity concern). Locked decision: visibility is **one switch**. If sensitivity is a concern, set visibility=public (which gitignores all of `.abcd/` including logbook). Per-subdirectory exceptions create maintenance burden and contradict the transparent-prompts principle ([`04-universal-patterns.md § 1`](./04-universal-patterns.md#1-transparent-prompts)).

**Sensitivity concern still valid for `/abcd:launch` payload**: regardless of visibility, the launch payload manifest ([`../04-surfaces/04-launch.md § 2`](../04-surfaces/04-launch.md#2-payload-manifest-default-deny)) excludes `.abcd/` entirely from what ships to the public sibling repo. So a private repo that commits its logbook locally still doesn't leak it on launch.

In private repos, the entire `.abcd/` namespace is reproducible from a fresh clone — embark from a freshly cloned repo works including logbook (useful for diagnosing past command runs).

**Memory locations to keep straight:**

abcd's curated memory exists at **two scopes** (per § The two `.abcd/` scopes), and there are two non-abcd memory locations alongside it:

1. **`<workspace>/.abcd/memory/`** — the **primary** abcd memory: curated semantic summaries written by `dev-sync memory`, tracked in private repos, the canonical input for `principle-distiller`. Most memory lives here.
2. **`~/.abcd/memory/`** — **user-scope** memory: personal preferences and cross-project principles that have no single workspace home.
3. **`.flow/memory/`** — flow-next's own memory namespace (e.g., `.flow/memory/pitfalls.md`). Owned by flow-next; abcd reads but doesn't write.
4. **`memory/`** — legacy `cp -r` snapshot at the repo root that some existing projects maintain (per the abcd v0 convention). abcd respects if present but doesn't write to it.

Which scope a curated page lands in is a routing decision — see [`07-memory.md`](./07-memory.md) § scope routing. Retrieval across the two scopes is **not** a flat union (that would overflow context); it is keyword-recall + budget-bracketed injection per itd-39. When the brief says "memory" without qualification, it means the workspace-scope `.abcd/memory/`.

**`.work/` is local-only everywhere.** Working notes, drafts, status trackers stay gitignored. abcd consumes them via `dev-sync` ([§ 2](#2-abcddevelopmentactivity-namespace-and-dev-sync)) which promotes useful content into tracked `.abcd/development/activity/` artefacts before disembark.

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

## 2. `.abcd/development/activity/` namespace and `dev-sync`

`.abcd/development/activity/` is the **curated-from-volatile-sources** namespace. Volatile inputs (gitignored or external) get analysed and promoted into tracked `.abcd/development/activity/` artefacts via `abcd dev-sync`. This solves three problems: noisy sources stay gitignored; curated lessons get tracked; abcd doesn't have to read volatile sources every time.

**Source → target table:**

| Volatile source (gitignored or external) | Curated abcd target (tracked in private repos) | Adapter |
|---|---|---|
| Agent memory (vendor-detected per [`04-universal-patterns.md § 7`](./04-universal-patterns.md#7-vendor-agnostic-adapters-with-environment-branching) — Claude Code: `~/.claude/projects/<encoded-cwd>/memory/`; OpenCode: TBD via itd-22) | `.abcd/memory/` | `memory.py` (dispatcher) |
| Ad-hoc oracle reviews not tied to a spec (vendor-detected per [`04-universal-patterns.md § 7`](./04-universal-patterns.md#7-vendor-agnostic-adapters-with-environment-branching) — RP: `~/Library/Application Support/RepoPrompt/Workspaces/Workspace-<project>-<UUID>/Chats/`; future: Codex CLI). **Spec-tied reviews are written directly to `.flow/reviews/<spec-id>/` at review time by the fn-2 Stop hook post-processor — `dev-sync reviews` does NOT sweep those.** | `.abcd/development/activity/reviews/` | `reviews.py` (dispatcher) |
| `.work/issues.md` | `.abcd/development/activity/issues/{open,resolved,wontfix}/iss-N-<slug>.md` (per itd-4) | `work_dir.py` (migration on first sync after install) |
| `.work/notes/`, `.work/<feature>/` | `.abcd/development/activity/notes/` | `work_dir.py` |
| RP workspace state (`~/Library/Application Support/RepoPrompt/Workspaces/<id>/workspace.json`) | `.abcd/rp/workspace.json` (per itd-7) | `rp_state_backend.py` |

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

- **Memory (volatile) → `.abcd/memory/` (curated):** Source detected per [`04-universal-patterns.md § 7`](./04-universal-patterns.md#7-vendor-agnostic-adapters-with-environment-branching) vendor-agnostic dispatcher — under Claude Code: `~/.claude/projects/<encoded-cwd>/memory/`; under OpenCode: TBD via itd-22. The repo-local legacy `memory/` snapshot (the `cp -r` pattern from abcd v0) is the workflow `dev-sync memory` replaces. Output is *not verbatim*: distilled summaries grouped by domain, written as actionable suggestions for future agents (e.g., "When implementing UI hit areas, always use `.contentShape(Rectangle())` — source: `feedback_hit_target_full_box`"). Why curated: raw memories grow unbounded and contain personal phrasing ("user got annoyed when X"). Inputs to `principle-distiller` (Pass C).

- **Reviews backend (volatile) → `.abcd/development/activity/reviews/` (curated):** Source detected per [`04-universal-patterns.md § 7`](./04-universal-patterns.md#7-vendor-agnostic-adapters-with-environment-branching) vendor-agnostic dispatcher. The primary backend is **RepoPrompt**. `dev-sync reviews` sweeps **ad-hoc oracle reviews not tied to a spec** from two RP sources: (a) the chat store at `~/Library/Application Support/RepoPrompt/Workspaces/Workspace-<project>-<UUID>/Chats/ChatSession-*.json` (plain JSON, well-structured); and (b) the history store's `prompt-exports/` — RP's `export_response: true` writes ad-hoc reviews to `<cwd>/prompt-exports/` with a **hardcoded, non-configurable path**, which ahoy redirects into `~/.abcd/history/<root-sha>/prompt-exports/` via the `<repo>/prompt-exports` symlink. **Spec-tied reviews are NOT swept here** — flow-next's fn-2 Stop hook writes them directly to `.flow/reviews/<spec-id>/` at review time. `dev-sync reviews` renders its sources → `.abcd/development/activity/reviews/oracle-{review,chat}-<timestamp>-<description>-<hash>.md` (the format `review-collator` consumes). Dedup by content hash; idempotent. One project may have multiple RP workspaces — match by content of `workspace.json`. **Stability risk:** vendor JSON schemas may change between releases. Mitigation: probe defensively, version-stamp in `_provenance.json`, on parse failure log a warning and fall back to existing `.abcd/development/activity/reviews/*.md` (don't lose what was already synced). **Privacy:** filter strictly by workspace → project-path match before reading chat content. Inputs to `review-collator` (Pass A).

- **`.work/` (volatile, local-only) → `.abcd/development/activity/issues/`, `.abcd/development/activity/notes/` (curated):** `.work/issues.md` (the abcd CLAUDE.md mandatory issue log) gets parsed entry-by-entry; each entry promoted to `.abcd/development/activity/issues/open/iss-N-<slug>.md` (per itd-4 ledger structure). `.work/notes/`, `.work/<feature>/` get distilled into `.abcd/development/activity/notes/`. Files in `.work/` are never moved or deleted — `dev-sync work` is read-and-curate, source stays put. Inputs to `principle-distiller` (Pass C) and `chat-distiller` (Pass B, as auxiliary context).

- **RP workspace state (volatile) → `.abcd/rp/workspace.json` (curated, per itd-7):** Read the RP workspace whose root path matches the current repo (walk `~/Library/Application Support/RepoPrompt/Workspaces/`, parse each workspace's root-path field, match against `git rev-parse --show-toplevel`; multi-match = most-recently-modified). Write to `.abcd/rp/workspace.json` with `~/`-relative path normalisation. Workspace.json only for now; presets, mcp-routing scoping, `--preset <name>` flag, and `abcd rp link` window helper come in a later phase.

**Reviews as a first-class pitfall source:**

Plan/implementation/completion reviews (the ones in `.abcd/development/activity/reviews/`) are *exceptionally* useful for spotting issues. The `review-collator` agent must extract every "P0 / P1 / watch out for X / found bug" finding as a candidate pitfall — **even when the original issue was fixed**, the lesson survives. Output:

- `reviews-consolidated.json` — full review summaries (existing)
- `candidate-pitfalls.json` — extracted findings ready for distiller dedup

`principle-distiller` (Pass C) has four pitfall sources to dedupe by topic-hash or canonical phrasing: source `memory/pitfalls.md` (or `.abcd/memory/pitfalls.md` after curation), `candidate-pitfalls.json`, Pass B chat-distiller deltas, and code-rescuer's `code-principles.json`.

**Distinct from `principles.json`:** `.abcd/memory/`, `.abcd/development/activity/reviews/`, `.abcd/development/activity/issues/`, `.abcd/development/activity/notes/` are **persistent rolling artefacts** in the source repo, refreshed by `dev-sync`, used as ongoing input to future agents. `principles.json` is **per-disembark synthesis** written into the lifeboat at `.abcd/lifeboat/principles.json`. The lifeboat consumes `.abcd/development/activity/`; `.abcd/development/activity/` is not the lifeboat.

## 3. Plugin shape — directory layout

**Plugin layout** (mirrors flow-next package layout for install/uninstall/marketplace mechanics):

```
abcdDev/
├── .claude-plugin/plugin.json
├── .claude-plugin/marketplace.json
├── README.md
├── commands/abcd/                      # user-facing command surfaces — canonical list in ../04-surfaces/README.md
│   ├── ahoy.md / disembark.md / embark.md / launch.md / intent.md / capture.md / memory.md
│   └── …                               # plus operator-internal commands (deps-check, ralph-up, session)
│   # NOTE: `uninstall` is a sub-verb of /abcd:ahoy (not a standalone command). The ahoy command
│   # markdown handles the install/uninstall/dry-run/destroy sub-verb dispatch internally.
├── skills/                             # see 08-skills.md — abcd ships ZERO user-facing skills.
│   #                                   # The entries below are Claude Code plugin-runtime workflow
│   #                                   # files that each command points at internally; they are NOT
│   #                                   # user-facing skills surfaced under /abcd:. Per the round-2
│   #                                   # command-structure refactor, the only would-be user-facing
│   #                                   # skill (/abcd:grill) was promoted to a /abcd:intent grill
│   #                                   # sub-verb. A later phase may introduce new user-facing skills here.
│   ├── abcd-ahoy/{SKILL.md, workflow.md}      # plugin-runtime workflow for /abcd:ahoy
│   ├── abcd-disembark/{SKILL.md, workflow.md} # plugin-runtime workflow for /abcd:disembark
│   ├── abcd-embark/{SKILL.md, workflow.md}    # plugin-runtime workflow for /abcd:embark
│   ├── abcd-launch/{SKILL.md, workflow.md}    # plugin-runtime workflow for /abcd:launch
│   ├── abcd-intent/{SKILL.md, workflow.md}    # plugin-runtime workflow for /abcd:intent
│   ├── abcd-capture/{SKILL.md, workflow.md}   # plugin-runtime workflow for /abcd:capture
│   ├── commit-attribution/SKILL.md     # harvested from legacy ~/ABCDevelopment/.claude/skills/
│   └── secrets-and-pii/SKILL.md        # consolidated from pii-protection + secret-scan
├── agents/                             # 15 agents — see 01-agents.md
│   ├── flow-essence.md
│   ├── decision-archaeologist.md
│   ├── review-collator.md
│   ├── chat-distiller.md
│   ├── principle-distiller.md
│   ├── artefact-curator.md
│   ├── brief-composer.md
│   ├── press-release-composer.md       # Pass C — project-level press release (embark contract)
│   ├── lifeboat-oracle.md
│   ├── code-rescuer.md                 # narrowed to principle extraction
│   ├── issue-scout.md                  # opt-in; prefers flow-next:github-scout, internal fallback
│   ├── embark-scaffolder.md            # extended: doc-architecture at scaffold time (per legacy harvest)
│   ├── launch-gatekeeper.md            # extended: doc updates + security-audit at promotion (per legacy harvest)
│   ├── intent-fidelity-reviewer.md      # 14th — compares shipped reality to intent press release
│   └── documentation-auditor.md        # 15th — subagent-only; invoked by disembark/embark/launch pre-flight
├── scripts/
│   ├── abcd-cli                        # bash wrapper → abcd_cli.py (named abcd-cli; scripts/abcd/ is the support package)
│   ├── abcd_cli.py                     # subcommand dispatch
│   └── abcd/                           # Python support package — canonical per-module inventory in scripts/abcd/README.md
│       ├── __init__.py
│       ├── …                           # see scripts/abcd/README.md § Module inventory for the shipped modules
│       ├── defaults/                   # plugin-bundled CLAUDE.md domain rules + marker block (per itd-3)
│       ├── schemas/                    # JSON Schema files for inter-agent contracts (see schemas/README.md)
│       └── tools/                      # external-tool monitor registry + flow-next verb sidecar
└── hooks/                              # Claude Code event hooks — manifest + handlers in hooks/README.md
    ├── hooks.json                      # UserPromptSubmit → prompt_router_hook.py; SessionStart / PreCompact → prompt_router_reset.py
    ├── prompt_router_hook.py           # CARL-style rule injector (per itd-3); reads .abcd/rules.json + plugin defaults
    └── prompt_router_reset.py          # per-session dedup-state reset (SessionStart / PreCompact)
```

The `scripts/abcd/` package is large and evolving — its authoritative
per-module inventory (with the private-`_name`/public-`name` convention) lives in
[`scripts/abcd/README.md`](../../../../scripts/abcd/README.md); this brief points
at it rather than duplicating the file list, which is how the earlier inline tree
went stale. The lifeboat-pipeline adapters (`flow_next.py`, `specstory.py`, the
`memory.py` / `reviews.py` dispatchers and their per-vendor backends, etc.) are a
phase-4 **design target**; [`02-adapters.md`](./02-adapters.md) owns the contract,
per-row status, and the target package location, so this brief does not restate
the layout here.

**Plugin-internal development namespace** (committed in private repos, gitignored in public):

```
.abcd/
├── meta.json
├── config.json
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
│   │   │   ├── planned/                # itd-N-<slug>.md (has linked flow-next spec, work pending or in flight)
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
│   ├── activity/                       # curated-from-volatile-sources (was .abcd/dev/)
│   │   ├── reviews/                    # synced from reviews backend (RP / Codex / future) per 04-universal-patterns.md § 7
│   │   ├── issues/{open,resolved,wontfix}/  # iss-N-<slug>.md ledger entries (per itd-4)
│   │   └── notes/                      # distilled from .work/notes/
│   └── voyage/                         # embark/disembark provenance and history (see ../04-surfaces/03-embark.md § 7)
│       ├── README.md
│       ├── embark/
│       │   ├── provenance.json         # source path + manifest hash + timestamp + files written
│       │   └── from/<timestamp>/       # opt-in via embark --archive: verbatim copy of input lifeboat
│       └── disembark/
│           └── history.jsonl           # append-only manifest log of every disembark run
├── memory/                             # curated memory artefact (memory backend → .abcd/memory/, vendor-detected per 04-universal-patterns.md § 7)
├── lifeboat/                           # disembark output snapshot only — regenerable, overwritten each run (per ../02-constraints/01-platform.md, ../04-surfaces/03-embark.md § 7)
├── logbook/                            # per-command / per-phase run logs (design target — no automatic session-log hook ships)
└── rp/                                 # RP workspace pull (per itd-7); workspace.json only for now
```

**User-facing docs** (ships in public sibling repo):

```
docs/
├── README.md
├── tutorials/                          # learning guides
├── guides/                             # task-oriented how-tos
├── reference/                          # command reference, config schemas
└── explanation/                        # conceptual: lifeboats, dev-sync, intents, capture, etc.
```

**Doc framework note**: flow-next handles specs/memory/config under `.flow/`. Plugin-internal design docs live under `.abcd/development/`. User-facing docs live under `docs/`. The old abcd v0 `docs/development/` mandate (planning/roadmap/process/decisions inside `docs/`) is informative — we use the *shape* (planning vs roadmap vs process), not the *location* (`.abcd/development/` instead of `docs/development/`).
