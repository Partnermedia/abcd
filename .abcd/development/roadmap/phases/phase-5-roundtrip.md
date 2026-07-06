# Phase 5 — Round-trip and ship

## Expectation

By the end of this phase, abcd closes its own loop. A user can take a lifeboat
produced by `/abcd:disembark` and run `/abcd:embark` to unpack it into a fresh
target — and the new repo's brief, memory, and ADRs are a faithful subset of
the source. They can run `/abcd:launch` to promote a `*Dev` repo to its public
sibling, with a secret-scan gate and a payload manifest. And abcd itself ships:
the first public release of the plugin is cut by running abcd's own launch
flow on `abcdDev`.

This is the phase where abcd stops being a project that is built *with* tools
and becomes a project that is built *with itself*.

## Milestone

- `/abcd:embark from <path>` unpacks a lifeboat into an empty target per the
  acceptance in `04-surfaces/03-embark.md`; the round-trip test passes
  (disembark a corpus repo → embark into an empty target → faithful subset).
- `/abcd:launch ship` runs the full scan stack and promotes `abcdDev` → public
  `abcd` per `04-surfaces/04-launch.md`.
- The bootstrap launch is done: abcd's first public release is pushed (a manual
  `git push`, documented in `/abcd:launch` as the first-launch exception).

## Phase Acceptance

> _Roll-up acceptance per [adr-9 amendment](../../decisions/adrs/adr-9-phase-as-product-layer.md). Each bullet asserts an emergent, cross-intent truth or a phase-spanning user journey — never a copy of an intent's own `## Acceptance Criteria`._

- **Given** a project disembarked in Phase 4, **when** a user runs
  `/abcd:embark` into an empty target and then inspects it, **then** the new
  repo's brief, memory, and ADRs are a faithful subset of the source — the
  round-trip property that only exists because disembark (Phase 4) and embark
  (this phase) agree on one lifeboat schema.
- **Given** a `*Dev` repo, **when** a user runs `/abcd:launch ship`, **then**
  the public sibling is promoted with the secret-scan gate passed and the
  payload manifest honoured — the full disembark→embark→launch arc is walkable
  end to end.
- **Given** the launch surface works, **when** abcd's own `abcdDev` is launched,
  **then** abcd has shipped itself — the emergent proof that the framework is
  usable for the kind of project it was built to serve.

## Scope

**Intents:** itd-7 (RP workspace pull — carries `.abcd/rp/workspace.json` in the
lifeboat for user-account migration; presets and routing are deferred to a
later phase).

**Brief plumbing-phases:** Phase 6 — `/abcd:embark from <path>`; Phase 7 —
`/abcd:launch ship`.

**Bootstrap work** (no intent): the first launch of abcd itself is a manual
`git push`, per `06-delivery/01-build-sequence.md` and the v1 milestone's
bootstrap exception.

## Maps against

- **Brief:** `04-surfaces/03-embark.md` (embark); `04-surfaces/04-launch.md`
  (launch); `06-delivery/01-build-sequence.md` Phases 6–7.
- **Intents deliver the expectation:** itd-7 delivers the RP workspace
  portability that makes a lifeboat carry the user's RP setup across machines.
- **ADRs realised:** adr-4 (lifeboat as regenerable output — embark is the
  unpack half of the round-trip adr-4 defines).

## Dependency rationale

- **embark runs after Phase 4** — embark unpacks what disembark packs; the
  lifeboat schema must be settled by the Phase 4 pipeline before embark can
  consume it.
- **launch runs after embark** — launch promotes a repo; the round-trip
  validation (embark proving the lifeboat is faithful) should pass before the
  ship surface is exercised on the real repo.
- **itd-7 depends on the dev-sync foundation** (Phase 1's ahoy work) and on the
  RP plumbing (Phase 0's fn-5). It is conceptually independent of the
  embark/launch chain and can run in parallel with Phase 4 if capacity allows.
- **The bootstrap launch is last** — it is abcd shipping itself; everything
  else must work first.

## Open questions

- The bootstrap launch is a manual `git push` by definition (no public repo
  exists to launch *into* yet). Confirm the public `abcd` repo's initial state
  and the exact manual steps before Phase 5 closes.
- itd-7 ships `workspace.json` only; presets and routing are explicitly
  deferred — confirm which later phase picks them up when this phase plans.
