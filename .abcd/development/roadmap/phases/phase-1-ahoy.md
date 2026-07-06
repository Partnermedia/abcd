# Phase 1 — ahoy

## Expectation

By the end of this phase, a user can run `/abcd:ahoy install` in any folder and
have abcd correctly installed: the folder is classified (managed repo / managed
workspace / unmanaged), registered in `~/.abcd/workspaces.json`, given its
CLAUDE.md marker block, and wired with the modular rules loader. Re-running
`install` is idempotent. The oracle cascade is whole — RP MCP, else Codex, else
the always-available in-session fallback. This is the **first phase that
delivers a command a user actually invokes**, and the moment abcd becomes
something a contributor installs rather than something that merely exists.

## Milestone

- `/abcd:ahoy install`, `uninstall`, `dry-run`, and `doctor` all run on a fresh
  repo and on a re-run, per the acceptance in `04-surfaces/01-ahoy.md`.
- The folder-classification pass correctly distinguishes the five folder
  kinds (per fn-15's amended matrix in `04-surfaces/01-ahoy.md` — the
  historical four-row matrix split `unmanaged` along the `.git/` axis into
  `unmanaged-repo` and `unmanaged-folder`, with `unmanaged-workspace` as the
  no-marker / sibling-repos shape) and writes/updates
  `~/.abcd/workspaces.json`.
- The rules loader is live: keyword-triggered rule injection works; an
  unrelated prompt injects zero abcd rules.
- The oracle cascade is whole: RP MCP (Phase 0's fn-5) → Codex → in-session
  fallback.
- The probe-only stubs for the other surfaces exist (`/abcd:disembark probe`,
  `/abcd:embark probe`, `/abcd:launch dry-run`, bare `/abcd:intent`,
  bare `/abcd:capture`) — they render or report, they do not yet act.
  Bare `/abcd:capture` (not `list`) per SD001: a `list` sub-verb would
  duplicate what the bare invocation already renders.

## Phase Acceptance

> _Roll-up acceptance per [adr-9 amendment](../../decisions/adrs/adr-9-phase-as-product-layer.md). Each bullet asserts an emergent, cross-intent truth or a phase-spanning user journey — never a copy of an intent's own `## Acceptance Criteria`._

- **Given** a fresh repo with abcd never installed, **when** a user runs
  `/abcd:ahoy install`, **then** in one command the folder is classified,
  registered in `~/.abcd/workspaces.json`, given its CLAUDE.md marker block,
  and wired with the rules loader — a journey spanning itd-40, itd-3, and the
  ahoy command that no single intent delivers alone.
- **Given** abcd installed in a repo, **when** any later command issues an
  oracle call, **then** the call resolves through the whole cascade — RP MCP,
  else Codex, else in-session subagent — and never hard-fails for lack of a
  backend. (Emergent: itd-2 and Phase 0's fn-5 each own one leg; the
  *whole-cascade* guarantee is owned by no single intent.)
- **Given** an installed repo, **when** a user's prompt contains a domain
  keyword, **then** the rules loader injects exactly the matching domain rules
  and nothing else — the just-in-time discipline-loading property that itd-3
  delivers and ahoy's marker block makes live.

## Scope

**Intents:** itd-3 (modular rules loader), itd-40 (folder classification +
`workspaces.json` — and the history-store scaffolding ahoy provisions per
managed folder, see below), itd-2 (in-session subagent oracle — the cascade's
always-available bottom).

**History-store scaffolding folds into itd-40.** ahoy currently has no spec for
provisioning the per-repo history store (`~/.abcd/history/` keyed on root-commit
SHA, `index.json`, the SpecStory redirect shim — done by hand so far, see
`.abcd/development/research/notes/ahoy-history-store-manual-scaffolding.md`).
itd-40 already owns the managed-folder model — folder classification and the
`~/.abcd/workspaces.json` registry — so "what ahoy provisions for each managed
folder" is the same intent. The history store is provisioning under that model,
not a separate concern; itd-40's scope and acceptance criteria absorb it when
itd-40 is planned.

**`/abcd:capture` moved out.** The capture surface (itd-4) was previously
scoped here. It is now Phase 2's sole intent — capture is a distinct
user-capability moment (a fast issue ledger), and bundling it with ahoy
produced a phase whose milestone mixed "abcd installs" with "the user can file
issues". Each is demoable on its own; they are now separate phases.

**Brief plumbing-phases:** the brief's "Phase 1 — `/abcd:ahoy` end-to-end" (the
`ahoy install` command flow, Steps 0–12 of `04-surfaces/01-ahoy.md`, plus the
probe-only stubs for the other surfaces).

## Maps against

- **Brief:** `04-surfaces/01-ahoy.md` (the command being built);
  `06-delivery/01-build-sequence.md` brief-Phase 1; `05-internals/03-configuration.md`
  (rules-loader config, `workspaces.json`, visibility-driven gitignore).
- **Intents deliver the expectation:** itd-3 delivers the marker block ahoy
  installs; itd-40 delivers the classification ahoy runs first; itd-2 completes
  the oracle cascade.
- **ADRs realised:** adr-3 (directory-as-truth — the lifecycle model the later
  capture and intent phases follow).

## Dependency rationale

- **itd-3 and itd-40 before `ahoy install`** — ahoy *installs* itd-3's marker
  block and *reads* itd-40's `workspaces.json` classification. ahoy must ship
  with both already in hand, so they precede the command flow within this
  phase.
- **itd-2 is independent of ahoy** — it can run in parallel with the rest of
  this phase. It is grouped here (rather than Phase 0) because the oracle
  cascade should be whole before the lifeboat pipeline (Phase 4) starts
  dispatching audits — but it has no hard dependency on the ahoy command flow.
- **This phase runs after Phase 0** — every spec here inherits the Phase 0
  disciplines, and the rules loader's `prompt_router_hook` builds on the
  Phase 0 hooks scaffold. itd-3's marker block also depends on the spec-terminology
  rename (itd-43) being complete so the block ships in settled vocabulary.

## Open questions

- None open. (The history-store-scaffolding question — whether it folds into
  itd-40, itd-3, or its own intent — was resolved into itd-40; see `## Scope`.)
