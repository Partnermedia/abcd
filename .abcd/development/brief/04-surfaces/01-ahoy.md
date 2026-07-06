# `/abcd:ahoy` — Install / Update

Mirrors flow-next-setup pattern. Built on a **detect → contract → render/apply**
architecture: a single detection pass produces a versioned, machine-readable
state contract; every sub-verb is a thin consumer of it. Detection logic lives
in exactly one place so bare invocation, `dry-run`, `doctor`, and `install`
cannot drift apart.

## Sub-verbs

Bare `/abcd:ahoy` shows status + help only — never mutates state. Current sub-verbs:

- **`/abcd:ahoy install`** — install or update the plugin in this repo
  (idempotent; covers first-install and upgrade). Runs the detection pass, then
  the apply pass over the resulting gaps.
- **`/abcd:ahoy uninstall`** — reversible marker-only removal: removes the
  marker block and the `/usr/local/bin/abcd` symlink (if owned by this plugin).
  NEVER mutates `hooks/hooks.json` (plugin-static per fn-14 T7 + fn-16 T1).
  Leaves `.abcd/` intact. Re-running `install` re-installs cleanly.
- **`/abcd:ahoy dry-run`** — run the detection pass and render the
  `DetectionResult` envelope as JSON (per fn-16 T1 — JSON ONLY; no
  unified-diff renderer). Never mutates state.
- **`/abcd:ahoy doctor`** — run the **full** detection pass and report every
  gap read-only, including user-scope state (history store, `index.json`,
  `workspaces.json`, symlink, hook) that `install` touches but never otherwise
  re-validates.
  Distinct from bare invocation, which only shows summary status + help.
- **Later phase: `/abcd:ahoy destroy`** — nuclear uninstall (per itd-10): removes
  `.abcd/` namespace too. Distinct from `uninstall`'s reversible behaviour.

## What abcd manages — repos, workspaces, and `~/.abcd/`

abcd manages exactly two kinds of folder, and keeps one user-scope directory:

```
~/.abcd/                       USER SCOPE — one per machine
  workspaces.json                registry of every managed repo + workspace
  history/                       shared history store + index.json
  memory/  config.json           user-scope memory + config

<anywhere>/<workspace>/        WORKSPACE — a collection of related repos
  .abcd/                         workspace namespace — development/, memory/,
                                 lifeboat/, logbook/, rp/
  CLAUDE.md                      role declaration (public: X / dev: Y)
  <repoA>/  <repoB>/           REPO — a single repository
    .specstory/cli/config.toml  gitignored redirect shim
    CLAUDE.md                   marker block

<anywhere>/<repo>/             REPO — a standalone single repository
  .specstory/cli/config.toml    gitignored redirect shim
  CLAUDE.md                     marker block
```

`/abcd:ahoy`'s detection pass **classifies the folder** into one of five
kinds (`managed-repo` / `managed-workspace` / `unmanaged-workspace` /
`unmanaged-repo` / `unmanaged-folder` — see detection step 0 and itd-40),
then `install` acts on the matching kind:

| Folder kind | What `install` does | `.abcd/` written |
|---|---|---|
| **repo** | the full repo install flow below; register in `workspaces.json` | **four-file carve-out** per fn-16 T1 + `05-internals/03-configuration.md:171-184` — repo install writes `<repo>/.abcd/config.json` and `<repo>/.abcd/rules.json` (no `meta.json` / `corpus.json` / workspace subdirs at repo scope; setup metadata lives at `config.json["meta"]`). Plus the CLAUDE.md/AGENTS.md marker block and gitignored `.specstory/` shim. |
| **workspace** | declare public/dev roles in the workspace `CLAUDE.md`; create the gitignored workspace `.work/` + the workspace `.abcd/` namespace; register the workspace and its repos in `workspaces.json` | `<workspace>/.abcd/` |

There is **no host or development-environment layer**. A folder like
`~/ABCDevelopment/` is just where a user keeps repos — abcd does not privilege
it. Everything that is genuinely machine-wide (the `workspaces.json` registry,
the shared `history/` store, user memory and config) lives under **`~/.abcd/`**,
which abcd bootstraps once. The `.abcd/` namespace is therefore scoped at two
levels — **user** (`~/.abcd/`) and **workspace** (`<workspace>/.abcd/`); a repo
is an install target but not an `.abcd/` scope. See
[`05-internals/03-configuration.md` § The two `.abcd/` scopes](../05-internals/03-configuration.md#the-two-abcd-scopes).

The history store and `workspaces.json` are shared across every repo on the
machine; they are bootstrapped once. An `install` that finds no `~/.abcd/`
**bootstraps it transparently** before registering — the user is never blocked
by missing user-scope state.

When a repo sits inside a workspace, CLAUDE.md files form an **inheritance
chain**: repo → workspace. The repo's marker block references its parent
(`../CLAUDE.md`) so agents pick up the workspace-level instructions too. A
standalone repo has no parent and its CLAUDE.md stands alone.

## Architecture: detection pass + apply pass

`install`, `dry-run`, `doctor`, and bare `/abcd:ahoy` all run the **same
detection pass** and differ only in what they do with its output:

| Sub-verb | Detection | Then |
|---|---|---|
| bare `/abcd:ahoy` | full | render summary + help, no gap detail |
| `doctor` | full | render every gap read-only |
| `dry-run` | full | render the canonical `DetectionResult` JSON envelope (per fn-16 T1 — no unified-diff) |
| `install` | full | run the apply pass over the gaps |

This means **idempotency is a property of detection, not of a version stamp.**
A check compares *actual state* (`git check-ignore -v`, `readlink`, file
presence, hook-config grep, `index.json` lookup) — never `setup_version` alone.
If a user hand-deletes the marker block while `setup_version` is current,
detection still reports it `missing` and `install` repairs it.

### The detection pass

Probes the folder, the repo, and `~/.abcd/`, produces an in-memory state contract (the
`ahoy-state.json` shape — see [`05-internals/03-configuration.md`](../05-internals/03-configuration.md)).
Steps, run in parallel where independent:

0. **Folder classification** (per itd-40) — classify `cwd` into one of five
   kinds. Classification keys on a **signal hierarchy** — abcd-owned markers
   decide managed-vs-unmanaged; weaker signals only disambiguate shape:
   - **Strong (managed):** an entry for `cwd` in `~/.abcd/workspaces.json`, an
     `.abcd/` directory, or a CLAUDE.md/AGENTS.md abcd marker block.
   - **Weak:** a `.git/` directory means `cwd` is *a* repo — **not** that it is
     *managed*, and not which workspace it belongs to. Necessary, not sufficient.
   - **Heuristic:** sibling repo-shaped subdirs suggest a workspace shape — used
     only to tell repo from workspace, never to decide managed-vs-unmanaged.

   The five kinds (fn-15 amends the historical four-row matrix by splitting
   `unmanaged` along the `.git/` axis — bare `/abcd:ahoy` offers `install` on
   an `unmanaged-repo` to adopt it, but reports-and-stops on an
   `unmanaged-folder`, so the two need distinct kind tokens):

   | Kind | Strong marker? | `.git/`? | Sibling repos? | `install` acts as |
   |---|---|---|---|---|
   | `managed-repo` | yes | yes | — | `repo` |
   | `managed-workspace` | yes | no | — | `workspace` |
   | `unmanaged-workspace` | no | no | yes | `workspace` (after `install` adopts it) |
   | `unmanaged-repo` | no | yes | — | `repo` (after `install` adopts it) |
   | `unmanaged-folder` | no | no | no | none — nothing to act on |

   Bare `/abcd:ahoy` **reports the kind and stops** — it never adopts an
   `unmanaged-repo` or `unmanaged-workspace`; it names `/abcd:ahoy install` as
   the way to do that. Host discovery is a `~/.abcd/workspaces.json` lookup —
   no hardcoded path, no parent-directory walk. Subsequent detection checks
   are gated on the folder kind; a `repo` runs all of them, a `workspace` runs
   its subset.
1. Resolve `${ABCD_PLUGIN_ROOT:-${CLAUDE_PLUGIN_ROOT}}`.
2. **Dependency probe** — `gitleaks` and `presidio` on PATH; `trufflehog` if
   `scan.deep` is (or would be) enabled.
3. **`.abcd/` skeleton** — presence of `meta.json`, `config.json`, `rules.json`.
4. **Identity** — `git rev-list --max-parents=0 HEAD` → root SHA. Cross-check
   against `~/.abcd/history/index.json`: is this SHA registered?
   Is there a sibling entry with a matching name/github but a *different* root
   SHA (the re-founding signal — see below)?
5. **History-store wiring** — does the `~/.abcd/history/` store exist at all
   (bootstrap gap if not)? Does `<root-sha>/{specstory,prompt-exports}/`
   exist? Does `.specstory/cli/config.toml`'s `[local_sync] output_dir`
   actually resolve to `index.json`'s recorded path? Is the registered entry's
   `path` still accurate (mutable label — refresh if the repo moved)?
6. **Visibility state** — compare current `.gitignore` allowlist entries
   against the visibility table in
   [`05-internals/03-configuration.md § 1`](../05-internals/03-configuration.md#1-visibility-driven-gitignore-policy).
7. **Marker drift** — diff the existing CLAUDE.md/AGENTS.md marker block
   against the current template; classify `current` / `outdated` / `missing`.
   Also check the marker block's parent-`CLAUDE.md` reference is present and
   resolves (inheritance chain — see § The three layers).
8. **PATH symlink** — does `/usr/local/bin/abcd` exist, and does it point at
   this plugin, a different binary, or nothing?
9. **`prompt-exports` symlink** — is `<repo>/prompt-exports` a symlink
   resolving to `~/.abcd/history/<root-sha>/prompt-exports/`? RepoPrompt's
   `export_response: true` writes ad-hoc oracle reviews to `<cwd>/prompt-exports/`
   with a hardcoded, non-configurable path — the symlink is the only mechanism
   that redirects those into the history store for `dev-sync reviews` to curate.
10. **Hook manifest verification** (verify-only per fn-16 T1) — VERIFY that
    `hooks/hooks.json` is present in the plugin install AND contains the three
    required event entries (`UserPromptSubmit`, `SessionStart`, `PreCompact`)
    each referencing the expected `prompt_router_hook.py` / `prompt_router_reset.py`
    commands. A missing or malformed manifest surfaces as a non-resolvable
    `plugin-owned` diagnostic gap. Neither install nor uninstall ever mutates
    `hooks.json` — the manifest is plugin-static per fn-14 T7.
11. **Version** — read `meta.json.setup_version`; classify `first-time` /
    `upgrade` / `current`. One input among many — never the sole gate.

Each detected discrepancy becomes a **gap** with a stable `id`, a `category`,
a `scope`, a `title`, `detail`, and a `fix_hint`. Gaps are grouped by category:

| `category` | Examples | Apply behaviour |
|---|---|---|
| `safe-autocreate` | `.abcd/` skeleton, `.abcd/bin/` scripts, history-store dirs | apply once category approved, no per-item prompt |
| `config-change` | visibility, oracle backend, PATH symlink, specstory redirect | transparent confirm; skip-if-set with "current value" notice |
| `plugin-owned` | CLAUDE.md/AGENTS.md marker block (per itd-3); `hooks/hooks.json` manifest verification (verify-only per fn-16 T1) | silent overwrite on marker drift; non-resolvable diagnostic for malformed/missing manifest |
| `dependency` | `gitleaks`, `presidio`, `trufflehog` install | offer brew/pip with ONE category-level approval covering all three tools (per fn-16 T1 — per-CATEGORY, not per-tool) |
| `user-state` | `index.json` entry, re-founding, stale/duplicate entries | guided; never auto-edit user-scope app-state, report extras read-only |

### Templates as files

Every artefact `install` writes — the marker block, the `.abcd/rules.json`
skeleton, the `.specstory/cli/config.toml` redirect shim, `.abcd/usage.md` —
comes from a canonical file under `scripts/abcd/defaults/`, never from inline
prose in this surface or in the apply logic. If a template is stale, edit the
template file. Drift detection (step 7) is only meaningful because the marker
block has one canonical source.

## Install flow (`/abcd:ahoy install`)

`install` = detection pass, then apply pass. The apply pass walks the gaps
grouped by category, asks **one** category-level approval question per
category present, and applies the approved categories' gaps.

1. Run the detection pass (above).
2. **Dependency gaps** (`dependency`) — surface brew/pip commands for missing
   tools with ONE category-level approval (per fn-16 T1 — per-category, not
   per-tool). fn-16 NEVER auto-executes package managers; the user runs the
   install commands manually.
3. **Skeleton gaps** (`safe-autocreate`) — `abcd init --json` creates `.abcd/`,
   writes `meta.json` + `config.json` (idempotent merge). Create `.abcd/bin/`
   only if `--with-scripts`; copy scripts with `chmod +x`; write
   `.abcd/usage.md`. Create history-store dirs (see step 6).
4. **Config gaps** (`config-change`) — transparent prompts; each skips with a
   "current value" notice if the key is already set:
   - **Visibility** (private/public) — always re-confirms, shows current state
     and consequences (which directories will be tracked/untracked).
   - **If private:** `scan.deep` — probe `trufflehog`, offer install if missing.
   - **Docs target** (`CLAUDE.md` / `AGENTS.md` / Both / Skip).
   - **Oracle backend** (RP / Codex / In-session / Auto — defaults to Auto).
5. **Apply visibility** — add/remove `.gitignore` allowlist entries to match
   the visibility table in
   [`05-internals/03-configuration.md § 1`](../05-internals/03-configuration.md#1-visibility-driven-gitignore-policy).
   Show the resulting tracked-vs-ignored list before confirming.
6. **Registries** (`user-state` + `config-change`) — if `~/.abcd/workspaces.json`
   does not exist, create it (first-ever run). Register `cwd` in it: a `repo`
   or `workspace` entry per the classification in step 0, with its `path` and
   — for repos — a `root_commit` reference and parent `workspace` name (the
   single repo-vs-workspace question is asked here if the structure is
   ambiguous). If the `~/.abcd/history/` store does not exist, bootstrap it:
   create the directory and write an initial `index.json` with its `schema` +
   `description` header (see
   [`05-internals/03-configuration.md`](../05-internals/03-configuration.md)
   for both schemas and the two-registry cross-reference rule). Then create
   `~/.abcd/history/<root-sha>/{specstory,prompt-exports}/`,
   register the repo in `index.json` (refreshing the entry's `path` if the
   repo moved), and install the gitignored `.specstory/cli/config.toml`
   redirect from `scripts/abcd/defaults/specstory-cli-config.toml` — note the
   template uses a `[local_sync]` section with `output_dir` (this is what
   SpecStory's CLI actually reads). Finally create the `<repo>/prompt-exports`
   symlink → `~/.abcd/history/<root-sha>/prompt-exports/` so ad-hoc oracle
   reviews land in the store (per the
   [history-store note](../research/notes/ahoy-history-store-manual-scaffolding.md)).
7. **Marker block** (`plugin-owned`) — inject/refresh the block between
   `<!-- BEGIN ABCD -->` / `<!-- END ABCD -->` in the target docs. **Silent
   overwrite on drift, per itd-3 — no prompt for hand-edits inside the block;
   users edit outside the markers.** Content comes from
   `scripts/abcd/defaults/claude-md-marker-block.md`. Write the minimal
   `.abcd/rules.json` skeleton if missing.
8. **PATH symlink** (`config-change`) — transparent prompt: "Install `abcd`
   symlink to `/usr/local/bin/abcd`? Default: yes for private repos, no for
   public." If accepted AND the target is absent or already points at this
   plugin → write it. If a different `abcd` binary exists → refuse, show what
   it points to, suggest manual resolution.
9. **Hook registration** (`plugin-owned`, VERIFY-ONLY per fn-16 T1) — install
   verifies that `hooks/hooks.json` is present (the manifest is plugin-static
   per fn-14 T7). Install NEVER writes `hooks.json`; uninstall NEVER mutates
   `hooks.json`. A missing manifest surfaces as a `plugin-owned` non-resolvable
   diagnostic gap.
10. Stamp `setup_version` + `setup_date` in `meta.json`.
11. Print summary: installed files, config table, symlink status, hook status,
    notes (re-run hint, uninstall hint).

**Same-version re-install:** if the detection pass reports zero gaps, `install`
prints "already up to date" and exits without writing. This falls out of
detection naturally — it is not a version-stamp short-circuit.

**Mid-run changes:** accept `abcd config set <key> <value>` inline; re-run only
the affected detection check and its apply step.

## Re-founding (the `supersedes` flow)

When a repo is re-created with clean history — typically to strip in-repo
transcripts before sharing — it is genuinely a new repo with a new root SHA.
The detection pass (step 4) flags this as a `user-state` gap when the current
root SHA is absent from `index.json` **and** a sibling entry has a matching
name/github under a different SHA.

ahoy never auto-decides this. It surfaces the candidate predecessor and asks
the user to confirm. On confirmation, the apply pass:

1. Registers the new root SHA in `index.json`.
2. Sets the new entry's `supersedes` → old SHA.
3. Marks the old entry's `superseded_by` → new SHA.
4. Leaves the old repo's corpus in place under its own `<root-sha>/` dir for
   lifeboat review — nothing is moved or deleted.

If the user declines, ahoy registers the new SHA with no lineage link and
notes the orphaned-predecessor possibility in the summary.

## Sub-verb semantics

**Uninstall (`/abcd:ahoy uninstall`):** removes the BEGIN/END marker block from
CLAUDE.md/AGENTS.md and the `/usr/local/bin/abcd` symlink **if it points at this
plugin** (otherwise leave it alone). `hooks/hooks.json` is plugin-static per
fn-14 T7 — uninstall NEVER mutates it (per fn-16 T1 brief amendment).
**Leaves the entire `.abcd/` namespace intact** (meta, config, corpus, rules,
development/, memory/, lifeboat/, logbook/, rp/) and the history store. Deeper
removal (`/abcd:ahoy destroy`) comes in a later phase as itd-10.

**Uninstall ↔ install is a tested round-trip invariant:** after `uninstall`
then `install`, the detection pass must report zero gaps, and the resulting
state must be byte-identical to a fresh install modulo `setup_date`. This is an
acceptance criterion, not just prose — see § Acceptance.

**Dry-run (`/abcd:ahoy dry-run`):** runs the detection pass, then renders the
canonical `DetectionResult` envelope as JSON (per fn-16 T1 — JSON ONLY; no
unified-diff renderer in this command surface). Exits without writing. The
JSON envelope shape is `{folder_kind, adopted, root_sha, parent_workspace,
plugin_root_status, repo_identity, signals, gaps}` so consumers can drive
the Claude Code skill's two-pass approval protocol off `folder_kind` +
`gaps`.

**Doctor (`/abcd:ahoy doctor`):** runs the detection pass and reports every gap
read-only, grouped by category. Unlike bare invocation it shows full gap detail
including user-state checks — the history store exists and is writable, the
`index.json` entry matches this root SHA, the `output_dir` redirect resolves,
the symlink and hook are intact, and any stale/duplicate `index.json` entries.
The check users reach for after a repo rename, a machine migration, or "why
aren't my transcripts showing up." Never mutates state, never auto-fixes
user-scope app-state.

## Acceptance

- **Given** any abcd-aware terminal, **when** the user runs bare `/abcd:ahoy`,
  **then** the dispatcher runs the detection pass and shows summary install
  status (installed?, plugin version, marker drift state, visibility) plus the
  available sub-verbs with suggested next actions — and never mutates state.
- **Given** a fresh repo with no `.abcd/` directory, **when** `/abcd:ahoy
  install` runs to completion, **then** the four-file repo carve-out
  (`.abcd/config.json` + `.abcd/rules.json` per fn-16 T1) is written, the
  visibility-driven `.gitignore` allowlist entries are present, the
  history-store dirs + `index.json` entry + `.specstory/cli/config.toml`
  redirect are present, the CLAUDE.md/AGENTS.md marker block from
  `scripts/abcd/defaults/claude-md-marker-block.md` is installed, and
  `hooks/hooks.json` is verified present (verify-only — never mutated).
- **Given** a repo with `install` already run, **when** `/abcd:ahoy install`
  runs again with no state changes, **then** the detection pass reports zero
  gaps, the message reads "already up to date", and nothing is written.
- **Given** a repo where the marker block was hand-deleted but `setup_version`
  is current, **when** `/abcd:ahoy install` runs, **then** detection reports the
  marker `missing` and the apply pass restores it — idempotency keys off state,
  not the version stamp.
- **Given** a repo with `install` run at an older `setup_version`, **when**
  `/abcd:ahoy install` runs, **then** `setup_version` is updated, the marker
  block is refreshed, and existing config keys are preserved (skip-if-set).
- **Given** install is running and `gitleaks` or `presidio` is not on PATH,
  **when** the dependency category is approved, **then** the user is shown the
  brew/pip install commands with ONE category-level approval (per fn-16 T1 —
  per-category, not per-tool); fn-16 NEVER auto-executes package managers.
- **Given** RP MCP is unreachable, **when** the detection pass probes oracle
  backends, **then** `oracle.backend` is set to `"codex"` (if Codex CLI
  present) or `"in-session"` (otherwise), and a one-time hint is surfaced.
- **Given** a repo whose root SHA is absent from `index.json` while a sibling
  entry matches its name/github, **when** `/abcd:ahoy install` runs, **then**
  detection flags a re-founding candidate, ahoy asks before linking, and on
  confirmation sets `supersedes` / `superseded_by` and leaves both corpora in
  place.
- **Given** the user runs `/abcd:ahoy uninstall` then `/abcd:ahoy install`,
  **when** both complete, **then** the detection pass reports zero gaps and the
  resulting state is byte-identical to a fresh install modulo `setup_date`.
- **Given** the user runs `/abcd:ahoy dry-run`, **when** the command completes,
  **then** the detection pass runs, the canonical `DetectionResult` JSON
  envelope (`{folder_kind, adopted, root_sha, parent_workspace,
  plugin_root_status, repo_identity, signals, gaps}` per fn-16 T1) is printed
  to stdout, and no files are modified.
- **Given** the user runs `/abcd:ahoy doctor` on an installed repo whose
  `.specstory/cli/config.toml` `output_dir` no longer matches `index.json`,
  **then** the user-state gap is reported read-only with both paths cited, and
  no files are modified.
- **Given** a fresh machine with no `~/.abcd/`, **when** `/abcd:ahoy install`
  runs in a repo, **then** `~/.abcd/` is bootstrapped — the `history/` store
  (directory + `index.json` with `schema`/`description` header) and
  `workspaces.json` — before the repo is registered, so the user is not blocked
  by missing user-scope state.
- **Given** `/abcd:ahoy install` runs in a repo, **when** the apply pass
  completes, **then** `<repo>/prompt-exports` is a symlink resolving to
  `~/.abcd/history/<root-sha>/prompt-exports/`, so RepoPrompt's
  `export_response` writes land in the history store.
- **Given** a registered repo that has been moved on disk, **when**
  `/abcd:ahoy install` or `doctor` runs, **then** detection notices the stale
  `index.json` `path` and `install` refreshes it (the root SHA is unchanged, so
  the entry is updated, not duplicated).
- **Given** a git repository with no abcd markers and no `~/.abcd/workspaces.json`
  entry, **when** the user runs bare `/abcd:ahoy`, **then** it reports
  `unmanaged-repo`, names `/abcd:ahoy install` as the way to adopt it, and
  mutates nothing — bare invocation never adopts.
- **Given** a folder with sibling repo-shaped subdirectories but no abcd
  markers and no `~/.abcd/workspaces.json` entry, **when** the user runs bare
  `/abcd:ahoy`, **then** it reports `unmanaged-workspace`, names
  `/abcd:ahoy install` as the way to adopt it, and mutates nothing.
- **Given** no `~/.abcd/workspaces.json` exists, **when** the first
  `/abcd:ahoy install` runs, **then** the registry is created and the managed
  folder is recorded with the correct `repo`/`workspace` `kind`.
- **Given** a folder with sibling repo-shaped subdirectories, **when**
  `/abcd:ahoy install` runs, **then** the user is asked the single
  repo-vs-workspace question and the answer is recorded in `workspaces.json`.
