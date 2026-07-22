# `/abcd:ahoy` — Install / Update

Built on a **detect → contract → render/apply**
architecture: a single detection pass produces a versioned, machine-readable
state contract; every sub-verb is a thin consumer of it. Detection logic lives
in exactly one place so bare invocation, `dry-run`, `doctor`, and `install`
cannot drift apart.

## Sub-verbs

Bare `/abcd:ahoy` shows status + help only — never mutates state. The
`/abcd:ahoy` slash command dispatches only this bare read-only pass and names
the CLI for anything that writes; the sub-verbs below ship on the CLI as
`abcd ahoy <sub-verb>`. Current sub-verbs:

- **`/abcd:ahoy install`** — install or update the plugin in this repo
  (idempotent; covers first-install and upgrade). Runs the detection pass, then
  the apply pass over the resulting gaps.
- **`/abcd:ahoy uninstall`** — reversible marker-only removal: removes the
  marker block and the `/usr/local/bin/abcd` symlink (if owned by this plugin).
  NEVER mutates `hooks/hooks.json` (plugin-static per spc-14 T7 + spc-16 T1).
  Leaves `.abcd/` intact. Re-running `install` re-installs cleanly.
- **`/abcd:ahoy dry-run`** — run the detection pass and render the
  `DetectionResult` envelope as JSON (per spc-16 T1 — JSON ONLY; no
  unified-diff renderer). Never mutates state.
- **`/abcd:ahoy doctor`** — run the **full** detection pass and report every
  gap read-only, including user-scope state (history store, `index.json`,
  symlink, hook) that `install` touches but never otherwise
  re-validates.
  Distinct from bare invocation, which only shows summary status + help.
- **`abcd ahoy identity-check`** — exit non-zero if the git commit identity
  does not match `.abcd/config/identity.json` (the commit-identity gate).
  Read-only; a CLI-only operator/CI check with no slash-command surface.
- **Later phase: `/abcd:ahoy destroy`** — nuclear uninstall (per itd-10): removes
  `.abcd/` namespace too. Distinct from `uninstall`'s reversible behaviour.

## What abcd manages — repos and `~/.abcd/`

abcd manages exactly one kind of folder — a **repository** — and keeps one
user-scope directory for machine-local state:

```
~/.abcd/                       USER SCOPE — one per machine (machine-local state only)
  history/                       shared history store + index.json (identity/lineage,
                                 keyed on root-commit SHA — adr-29)
  config.json                    machine config.json defaults (a later phase)
  memory/                        user-scope memory (personal, cross-project — a later
                                 phase; the shipped store is repo-scope .abcd/memory/)

<anywhere>/<repo>/             REPO — a single repository (the only install target)
  .abcd/                         repo-scope record + config.json + rules.json
  CLAUDE.md                      marker block (stands alone)
```

`/abcd:ahoy`'s detection pass **classifies `cwd`** into one of three kinds
(`managed-repo` / `unmanaged-repo` / `unmanaged-folder` — see detection step 0
and itd-40), then `install` acts on the matching kind:

| Folder kind | What `install` does | `.abcd/` written |
|---|---|---|
| **`managed-repo`** | a git repo abcd already manages — run the repo install flow below (idempotent update) | repo-scope `.abcd/config.json` (with the `meta` setup block) + `.abcd/rules.json` per spc-16 T1 (no `meta.json` at repo scope; setup metadata lives at `config.json["meta"]`). Plus `.abcd/config/identity.json` when the git-identity pin is adopted (per iss-62 — a `config-change` write) and the CLAUDE.md/AGENTS.md marker block. See [`05-internals/03-configuration.md` § The two `.abcd/` scopes](../05-internals/03-configuration.md#the-two-abcd-scopes). |
| **`unmanaged-repo`** | a git repo with no abcd yet — bare `/abcd:ahoy` offers `install` to adopt it; `install` runs the same repo flow | same as `managed-repo` |
| **`unmanaged-folder`** | not a git repo — nothing to act on; reports and stops | none |

There is **no workspace, host, or development-environment layer**. A folder
like `~/ABCDevelopment/` is just where a user keeps repos — abcd does not
privilege it, and it groups nothing. abcd lives in **one repository**
([adr-28](../../decisions/adrs/0028-single-repo-curated-release.md)): the design
record is repo-scoped and in-tree. Everything genuinely machine-wide — the
`history/` store, plus machine config and user memory in a later phase — lives
under **`~/.abcd/`**, which abcd bootstraps once. The `.abcd/` namespace is scoped at two levels —
**user** (`~/.abcd/`) and **repo** (in-tree `.abcd/`). See
[`05-internals/03-configuration.md` § The two `.abcd/` scopes](../05-internals/03-configuration.md#the-two-abcd-scopes).

The history store is shared across every repo on the machine; its `index.json`
is the **sole user-scope registry** (identity and lineage keyed on each repo's
immutable root-commit SHA). An `install` that finds no `~/.abcd/` **bootstraps
it transparently** before registering — the user is never blocked by missing
user-scope state.

Each repo's CLAUDE.md/AGENTS.md marker block **stands alone** — there is no
repo→workspace inheritance chain.

## Architecture: detection pass + apply pass

`install`, `dry-run`, `doctor`, and bare `/abcd:ahoy` all run the **same
detection pass** and differ only in what they do with its output:

| Sub-verb | Detection | Then |
|---|---|---|
| bare `/abcd:ahoy` | full | render summary + help, no gap detail |
| `doctor` | full | render every gap read-only |
| `dry-run` | full | render the canonical `DetectionResult` JSON envelope (per spc-16 T1 — no unified-diff) |
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

0. **Folder classification** (per itd-40) — classify `cwd` into one of three
   kinds. Classification keys on a **signal hierarchy** — abcd-owned markers
   decide managed-vs-unmanaged; the `.git/` signal only disambiguates shape:
   - **Strong (managed):** an entry for `cwd`'s root-commit SHA in
     `~/.abcd/history/index.json`, an in-tree `.abcd/` directory, or a
     CLAUDE.md/AGENTS.md abcd marker block.
   - **Weak:** a `.git/` directory means `cwd` is *a* repo — **not** that it is
     *managed*. Necessary, not sufficient.

   The three kinds (bare `/abcd:ahoy` offers `install` on an `unmanaged-repo`
   to adopt it, but reports-and-stops on an `unmanaged-folder`, so the two need
   distinct kind tokens):

   | Kind | Strong marker? | `.git/`? | `install` acts as |
   |---|---|---|---|
   | `managed-repo` | yes | yes | `repo` (idempotent update) |
   | `unmanaged-repo` | no | yes | `repo` (after `install` adopts it) |
   | `unmanaged-folder` | no | no | none — nothing to act on |

   Bare `/abcd:ahoy` **reports the kind and stops** — it never adopts an
   `unmanaged-repo`; it names `/abcd:ahoy install` as the way to do that.
   Identity resolution is a `~/.abcd/history/index.json` lookup keyed on the
   root-commit SHA — no hardcoded path, no parent-directory walk. The detection
   checks run on a repo kind; an `unmanaged-folder` short-circuits.
1. Resolve `${ABCD_PLUGIN_ROOT:-${CLAUDE_PLUGIN_ROOT}}`.
2. **Adapter probe** — the native secret + PII scan needs no external tool;
   this step only reports which **opt-in** scanners are on PATH: `gitleaks`
   (deeper secret scan) and `trufflehog` if `scan.deep` is (or would be)
   enabled.
3. **`.abcd/` skeleton** — presence of `config.json` (which carries the `meta` setup block) and `rules.json`.
4. **Identity** — `git rev-list --max-parents=0 HEAD` → root SHA. Cross-check
   against `~/.abcd/history/index.json`: is this SHA registered?
   Is there a sibling entry with a matching name/github but a *different* root
   SHA (the re-founding signal — see below)? Separately, compare the git author
   identity a commit would use against the committed `.abcd/config/identity.json`
   pin (iss-62), emitting a `git_identity.unpinned` / `.mismatch` / `.unset` /
   `.uncheckable` `config-change` gap.
5. **History-store wiring** — does the `~/.abcd/history/` store (abcd's native
   local redacted transcript corpus, per
   [adr-29](../../decisions/adrs/0029-native-transcript-corpus.md)) exist at all
   (bootstrap gap if not)? Does the `<root-sha>/transcripts/` directory exist? Is
   the registered entry's `path` still accurate (mutable label — refresh if the
   repo moved)?
6. **Visibility state** — compare current `.gitignore` allowlist entries
   against the visibility table in
   [`05-internals/03-configuration.md § 1`](../05-internals/03-configuration.md#1-visibility-driven-gitignore-policy).
7. **Marker drift** — diff the existing CLAUDE.md/AGENTS.md marker block
   against the current template; classify `current` / `outdated` / `missing`.
   The marker block stands alone — there is no parent-`CLAUDE.md` reference to
   verify.
8. **PATH symlink** — does `/usr/local/bin/abcd` exist, and does it point at
   this plugin, a different binary, or nothing?
9. **Hook manifest verification** (verify-only per spc-16 T1) — VERIFY that
   `hooks/hooks.json` is present in the plugin install AND contains the three
   required event entries (`UserPromptSubmit`, `SessionStart`, `PreCompact`)
   each referencing the expected prompt-router hook commands. The shipped
   manifest also wires `abcd hook session-start` (a second `SessionStart`
   command) and `abcd hook session-end` (a fourth `SessionEnd` event);
   verification covers only the three prompt-router commands above. A missing or
   malformed manifest surfaces as a non-resolvable `plugin-owned` diagnostic
   gap. Neither install nor uninstall ever mutates `hooks.json` — the manifest
   is plugin-static per spc-14 T7.
10. **Version** — read `config.json["meta"].setup_version`; classify `first-time` /
    `upgrade` / `current`. One input among many — never the sole gate.

Each detected discrepancy becomes a **gap** with a stable `id`, a `category`,
a `scope`, a `title`, `detail`, and a `fix_hint`. Gaps are grouped by category:

| `category` | Examples | Apply behaviour |
|---|---|---|
| `safe-autocreate` | `.abcd/` skeleton, history-store dirs (`.abcd/bin/` scripts in a later phase) | apply once category approved, no per-item prompt |
| `config-change` | visibility, oracle adapter, PATH symlink, git-identity pin | transparent confirm; skip-if-set with "current value" notice |
| `plugin-owned` | CLAUDE.md/AGENTS.md marker block (per itd-3); `hooks/hooks.json` manifest verification (verify-only per spc-16 T1) | silent overwrite on marker drift; non-resolvable diagnostic for malformed/missing manifest |
| `dependency` | opt-in scanners: `gitleaks`, `trufflehog` install | offer brew with ONE category-level approval covering the opt-in scanners (per spc-16 T1 — per-CATEGORY, not per-tool) |
| `user-state` | `index.json` entry, re-founding, stale/duplicate entries | guided; never auto-edit user-scope app-state, report extras read-only |

### Templates as files

The marker block `install` writes comes from a canonical file under
`internal/core/ahoy/defaults/` (`claude-md-marker-block.md`), never from inline
prose in this surface. If the template is stale, edit the template file. Drift
detection (step 7) is only meaningful because the marker block has one
canonical source. The `.abcd/rules.json` skeleton is written inline by the
apply pass (`stepRules`); a later phase moves it — and `.abcd/usage.md`, once
that artefact ships — to canonical files under `defaults/` too.

## Install flow (`/abcd:ahoy install`)

`install` = detection pass, then apply pass. The apply pass walks the gaps
grouped by category, asks **one** category-level approval question per
category present, and applies the approved categories' gaps.

Non-interactive override flags pre-answer the prompts: `--yes` approves every
resolvable category without prompting, `--adopt` / `--refuse-adopt` decide the
unmanaged-repo adoption question, `--docs-target` (`claude_md` | `agents_md` |
`both` | `skip`) sets the marker target, `--oracle-backend`
(`host-delegated` | `native` | `cli` | `api` | `mcp`) sets the oracle,
`--scan-deep` (`true` | `false`) toggles the deep scan, and `--visibility`
(`private` | `public`) sets repo visibility. `--yes` does not adopt an
unmanaged repo or pin an unset git identity — those still need `--adopt` and an
interactive confirmation.

1. Run the detection pass (above).
2. **Dependency gaps** (`dependency`) — surface brew/pip commands for missing
   tools with ONE category-level approval (per spc-16 T1 — per-category, not
   per-tool). spc-16 NEVER auto-executes package managers; the user runs the
   install commands manually.
3. **Skeleton gaps** (`safe-autocreate`) — the apply pass creates `.abcd/` and
   writes `config.json` (idempotent merge; the `meta` setup block is a key
   within it). Create history-store dirs (see step 6). Later phase: create
   `.abcd/bin/` (only with `--with-scripts`; scripts copied with `chmod +x`)
   and write `.abcd/usage.md`.
4. **Config gaps** (`config-change`) — transparent prompts; each skips with a
   "current value" notice if the key is already set:
   - **Visibility** (private/public) — always re-confirms, shows current state
     and consequences (which directories will be tracked/untracked).
   - **If private:** `scan.deep` — probe `trufflehog`, offer install if missing.
   - **Docs target** (`CLAUDE.md` / `AGENTS.md` / Both / Skip).
   - **Oracle** — host-delegated by default (abcd hands prompts to the host's
     subagent dispatch); an opt-in oracle adapter (native / CLI / API / MCP)
     can be wired instead (per
     [adr-25](../../decisions/adrs/0025-host-delegated-llm-default.md)).
5. **Apply visibility** — add/remove `.gitignore` allowlist entries to match
   the visibility table in
   [`05-internals/03-configuration.md § 1`](../05-internals/03-configuration.md#1-visibility-driven-gitignore-policy).
   Show the resulting tracked-vs-ignored list before confirming.
6. **History-store registry** (`user-state`) — if the `~/.abcd/history/` store
   does not exist, bootstrap it: create the directory and write an initial
   `index.json` with its `schema` + `description` header (see
   [`05-internals/03-configuration.md`](../05-internals/03-configuration.md)
   for the schema). Then create the `~/.abcd/history/<root-sha>/transcripts/`
   transcript directory (abcd's native local redacted transcript corpus, per
   [adr-29](../../decisions/adrs/0029-native-transcript-corpus.md)), write the
   per-repo `<root-sha>/meta.json` (`root_commit`, `name`, `github`, and a
   `corpus` block pointing at `transcripts/`), and register the repo in
   `index.json` by its immutable `root_commit`, refreshing the entry's mutable
   `path` if the repo moved. The history `index.json` is the
   **sole user-scope registry** — there is no `workspaces.json`. Transcript
   capture is native — no external tool and no per-repo redirect shim.
7. **Marker block** (`plugin-owned`) — inject/refresh the block between
   `<!-- BEGIN ABCD -->` / `<!-- END ABCD -->` in the target docs. **Silent
   overwrite on drift, per itd-3 — no prompt for hand-edits inside the block;
   users edit outside the markers.** Content comes from
   `internal/core/ahoy/defaults/claude-md-marker-block.md`. Write the minimal
   `.abcd/rules.json` skeleton if missing.
8. **PATH symlink** (`config-change`) — transparent prompt: "Install `abcd`
   symlink to `/usr/local/bin/abcd`? Default: yes for private repos, no for
   public." If accepted AND the target is absent or already points at this
   plugin → write it. If a different `abcd` binary exists → refuse, show what
   it points to, suggest manual resolution.
9. **Hook registration** (`plugin-owned`, VERIFY-ONLY per spc-16 T1) — install
   verifies that `hooks/hooks.json` is present (the manifest is plugin-static
   per spc-14 T7). Install NEVER writes `hooks.json`; uninstall NEVER mutates
   `hooks.json`. A missing manifest surfaces as a `plugin-owned` non-resolvable
   diagnostic gap.
10. Stamp `setup_version` + `setup_date` in `config.json["meta"]`.
11. Print summary: installed files, config table, symlink status, hook status,
    notes (re-run hint, uninstall hint).

**Same-version re-install:** if the detection pass reports zero actionable
gaps (gaps both required and resolvable), `install` prints "already up to
date" and exits without writing. This falls out of
detection naturally — it is not a version-stamp short-circuit.

**Mid-run changes (a later phase):** accept `abcd config set <key> <value>`
inline; re-run only the affected detection check and its apply step.

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
spc-14 T7 — uninstall NEVER mutates it (per spc-16 T1 brief amendment).
**Leaves the entire `.abcd/` namespace intact** (meta, config, corpus, rules,
development/, memory/, lifeboat/, logbook/, rp/) and the history store. Deeper
removal (`/abcd:ahoy destroy`) comes in a later phase as itd-10.

**Uninstall ↔ install is a tested round-trip invariant:** after `uninstall`
then `install`, the detection pass must report zero actionable gaps, and the resulting
state must be byte-identical to a fresh install modulo `setup_date`. This is an
acceptance criterion, not just prose — see § Acceptance.

**Dry-run (`/abcd:ahoy dry-run`):** runs the detection pass, then renders the
canonical `DetectionResult` envelope as JSON (per spc-16 T1 — JSON ONLY; no
unified-diff renderer in this command surface). Exits without writing. The
JSON envelope shape is `{folder_kind, adopted, root_sha,
plugin_root_status, repo_identity, signals, gaps}` so the plugin command
(`commands/abcd/ahoy.md`) summarises state off `folder_kind` + `gaps` and
names `abcd ahoy install` for anything actionable.

**Doctor (`/abcd:ahoy doctor`):** runs the detection pass and reports every gap
read-only, grouped by category. Unlike bare invocation it shows full gap detail
including user-state checks — the history store exists and is writable, the
`index.json` entry matches this root SHA, the PATH symlink and hook manifest
are intact, and any stale/duplicate `index.json` entries.
The check users reach for after a repo rename, a machine migration, or "why
aren't my transcripts showing up." Never mutates state, never auto-fixes
user-scope app-state.

## Acceptance

- **Given** any abcd-aware terminal, **when** the user runs bare `/abcd:ahoy`,
  **then** the dispatcher runs the detection pass and shows summary install
  status (installed?, plugin version, marker drift state, visibility) plus the
  available sub-verbs with suggested next actions — and never mutates state.
- **Given** a fresh repo with no `.abcd/` directory, **when** `/abcd:ahoy
  install` runs to completion, **then** the two-file repo carve-out
  (`.abcd/config.json` + `.abcd/rules.json` per spc-16 T1) is written,
  `.abcd/config/identity.json` is pinned when the git-identity gate is adopted
  (per iss-62), the
  visibility-driven `.gitignore` allowlist entries are present, the
  history-store dirs + `index.json` entry are present, the
  CLAUDE.md/AGENTS.md marker block from
  `internal/core/ahoy/defaults/claude-md-marker-block.md` is installed, and the
  `hooks/hooks.json` check runs verify-only (never mutated) — the manifest is
  plugin-static per spc-14 T7, and a missing or malformed manifest surfaces as
  a non-resolvable `plugin-owned` diagnostic gap.
- **Given** a repo with `install` already run, **when** `/abcd:ahoy install`
  runs again with no state changes, **then** the detection pass reports zero
  actionable gaps, the message reads "already up to date", and nothing is written.
- **Given** a repo where the marker block was hand-deleted but `setup_version`
  is current, **when** `/abcd:ahoy install` runs, **then** detection reports the
  marker `missing` and the apply pass restores it — idempotency keys off state,
  not the version stamp.
- **Given** a repo with `install` run at an older `setup_version`, **when**
  `/abcd:ahoy install` runs, **then** `setup_version` is updated, the marker
  block is refreshed, and existing config keys are preserved (skip-if-set).
- **Given** install is running and an opt-in scanner (`gitleaks` / `trufflehog`)
  is not on PATH, **when** the dependency category is approved, **then** the user
  is shown the brew install commands with ONE category-level approval (per spc-16
  T1 — per-category, not per-tool); spc-16 NEVER auto-executes package managers.
- **Given** no oracle adapter is wired, **when** the detection pass resolves the
  oracle, **then** `oracle` stays **host-delegated** (abcd needs no API keys or
  model config — it emits prompts the host runs, per adr-25), and an opt-in
  adapter (native / CLI / API / MCP) can be configured later.
- **Given** a repo whose root SHA is absent from `index.json` while a sibling
  entry matches its name/github, **when** `/abcd:ahoy install` runs, **then**
  detection flags a re-founding candidate, ahoy asks before linking, and on
  confirmation sets `supersedes` / `superseded_by` and leaves both corpora in
  place.
- **Given** the user runs `/abcd:ahoy uninstall` then `/abcd:ahoy install`,
  **when** both complete, **then** the detection pass reports zero actionable
  gaps and the resulting state is byte-identical to a fresh install modulo
  `setup_date`.
- **Given** the user runs `/abcd:ahoy dry-run`, **when** the command completes,
  **then** the detection pass runs, the canonical `DetectionResult` JSON
  envelope (`{folder_kind, adopted, root_sha,
  plugin_root_status, repo_identity, signals, gaps}` per spc-16 T1) is printed
  to stdout, and no files are modified.
- **Given** the user runs `/abcd:ahoy doctor` on an installed repo whose
  registered history-store `path` no longer matches `index.json`, **then** the
  user-state gap is reported read-only with both paths cited, and no files are
  modified.
- **Given** a fresh machine with no `~/.abcd/`, **when** `/abcd:ahoy install`
  runs in a repo, **then** `~/.abcd/` is bootstrapped — the `history/` store
  (directory + `index.json` with `schema`/`description` header) — before the
  repo is registered, so the user is not blocked by missing user-scope state.
- **Given** a registered repo that has been moved on disk, **when**
  `/abcd:ahoy install` or `doctor` runs, **then** detection notices the stale
  `index.json` `path` and `install` refreshes it (the root SHA is unchanged, so
  the entry is updated, not duplicated).
- **Given** a git repository with no abcd markers and no `~/.abcd/history/index.json`
  entry for its root-commit SHA, **when** the user runs bare `/abcd:ahoy`,
  **then** it reports `unmanaged-repo`, names `/abcd:ahoy install` as the way to
  adopt it, and mutates nothing — bare invocation never adopts.
- **Given** a folder that is not a git repository and has no abcd markers,
  **when** the user runs bare `/abcd:ahoy`, **then** it reports
  `unmanaged-folder`, reports there is nothing to act on, and mutates nothing.
