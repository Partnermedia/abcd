# fn-82.4 — `task reset` data-model discovery (GATE 1)

Blocking artifact for the `task reset --force` override. The model below is
**proven from the live tree with fixtures**, not assumed. GATE 2 (the `--force`
implementation) is unblocked because every element is fixture-confirmed.

## The two artifacts per task (NOT a `.json`+`.md` runtime pair)

A task's state is split across **three** files in two locations. A tracked
DEFINITION file, a tracked human-readable spec, and an UNTRACKED runtime-state
file under the git common-dir. `flowctl show` reads the merge of the definition
and the runtime file.

### FIXTURE A — tracked task DEFINITION: `.flow/tasks/<id>.json`

Definition-only. Carries NO runtime fields (no `status`, no `claimed_at`). For
`fn-82-enhancements-and-debt-lint-precision.4.json`:

```
keys: ['created_at', 'depends_on', 'id', 'priority', 'spec', 'spec_path', 'title']
has status? False
```

This is what a committed task file looks like. A definition that has leaked a
runtime field (e.g. `updated_at`) is the recorded KeyError-class bug
(`bug/runtime-errors/flowctl-task-set-spec-leaks-updated-at`): with no state
file, `load_task_with_state` skips the `status: todo` fallback and `flowctl show`
crashes `KeyError 'status'`. The `--force` release therefore writes the
definition through the canonicalizer (RUNTIME_FIELDS-stripped) — never a raw
dump.

### Tracked spec: `.flow/tasks/<id>.md`

Human-readable spec + an `## Evidence` section that upstream `clear_task_evidence`
scrubs on reset. The non-force / non-`in_progress` paths reach it via delegation
(upstream's own reset chain runs). The `in_progress` force-release does NOT
delegate, so it scrubs `## Evidence` itself via `_task_reset_clear_evidence` —
a local re-author of upstream's regex + empty template (kept in lockstep by the
delegation-parity tests) — so a force-released task never carries stale evidence
into the next worker.

### FIXTURE B — untracked RUNTIME STATE: `<git-common-dir>/flow-state/tasks/<id>.state.json`

The authoritative runtime state. For the same task (in_progress at capture):

```
keys: ['assignee', 'claimed_at', 'status', 'updated_at']
status: in_progress
```

Absent state file ⇒ `load_task_with_state` falls back to definition runtime
fields, then to `{"status": "todo"}`. A reset baseline is
`{"status": "todo", "updated_at": <now>}` (see `reset_task_runtime`,
flowctl.py:1009).

## FIXTURE C — flow-state path grammar

Resolved the SAME way `flowctl.get_state_dir()` (flowctl.py:855) and
`flowstate_heal.state_dir_for()` resolve it — NEVER hardcode `.git/flow-state`:

1. `FLOW_STATE_DIR` env override (if set), else
2. `git rev-parse --path-format=absolute --git-common-dir` + `/flow-state`
   (worktree-safe — shared across all worktrees), else
3. `.flow/state` fallback for non-git trees.

In this repo `git rev-parse --path-format=absolute --git-common-dir` ⇒
`<repo>/.git`, so:

- **state file:** `<git-common-dir>/flow-state/tasks/<id>.state.json`
- **lock file:** `<git-common-dir>/flow-state/locks/<id>.lock` — a **0-byte
  advisory fcntl lock** (`LocalFileStateStore.lock_task`, flowctl.py:940). It is
  held ONLY for the duration of a live flowctl state mutation via
  `fcntl.flock(f, LOCK_EX)`; it carries no pid payload. Liveness = "can a
  non-blocking `LOCK_EX` be acquired?" — if `flock` raises `BlockingIOError`, a
  live process is mid-mutation on this task.

## FIXTURE D — canonical runtime-field list (canonicalizer-sourced, authoritative)

Sourced from `scripts.abcd.tools.flowctl_runtime.RUNTIME_FIELDS` — the neutral
copy that `flowstate_heal._save_task_definition_bytes` filters against and that a
one-shot test reconciles against the live fork. NEVER hand-enumerated:

```
['assignee', 'blocked_reason', 'claim_note', 'claimed_at', 'evidence', 'status', 'updated_at']
```

The `--force` release strips exactly this set from the DEFINITION write (via the
same filter chain `flowstate_heal` uses), guaranteeing the definition stays
runtime-field-free and cannot re-trip the KeyError class.

## What `flowctl show` reads (proven)

`cmd_show` (flowctl.py:10854) reads every task through `load_task_with_state`
(flowctl.py:977) = `definition ∪ runtime` (runtime wins on RUNTIME_FIELDS). So
after a `--force` release the healed state file (`status: todo`) is what `show`
reports; a stale in_progress definition field would be masked by the state file
but is stripped anyway.

## Outcome

Model PROVEN by fixtures. GATE 2 is unblocked. The `--force` in_progress release
path (which does NOT delegate — upstream refuses in_progress outright):
1. liveness-checks the advisory lock (non-blocking `LOCK_EX`),
2. writes a definition-only artifact via the RUNTIME_FIELDS filter chain,
3. removes the `.state.json` so `flowctl show` reads the definition's `todo`
   fallback (equivalent end-state to upstream's `reset_task_runtime` baseline
   overwrite for the documented reader; upstream's baseline-write is NOT invoked
   on the force-release path, and no fresh `updated_at` state timestamp is
   written — removal is the safest heal for the KeyError class),
4. scrubs the `.md` `## Evidence` section via `_task_reset_clear_evidence`
   (matching upstream's `clear_task_evidence`, which the non-delegating release
   would otherwise skip),
5. distinguishes three exit outcomes (released / refused-live / nothing-to-do).

Any OTHER non-todo state (done/blocked) delegates to upstream's own reset
unchanged, so steps 2–4 there run via upstream (`reset_task_runtime` +
definition cleanup + `clear_task_evidence`).
