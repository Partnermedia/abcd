# Phase 2 — capture

## Expectation

By the end of this phase, a user working in an abcd-installed repo can run
`/abcd:capture "<text>"` and have the finding land immediately in a structured
issue ledger — no context switch, no form, no triage decision at capture time.
Issues carry stable `iss-N` IDs and a folder-as-status lifecycle
(`open` / `resolved` / `wontfix`); `/abcd:capture list` is a filtered query
over that ledger. This is the phase that makes "I noticed something" a
one-line, low-friction act rather than a deferred intention.

## Milestone

- `/abcd:capture "<text>"` writes a structured issue-ledger entry under
  `.abcd/development/activity/issues/open/` with a stable `iss-N` ID, per the
  acceptance in `04-surfaces/06-capture.md`.
- The folder-as-status lifecycle works: `capture resolve <iss-N>` and
  `capture wontfix <iss-N>` move the entry between status folders; no `status:`
  field is stored (directory is truth, per adr-3).
- `/abcd:capture list` returns a filtered view of the ledger (the earned
  sub-verb — bare `/abcd:capture` with no argument renders help per the
  bare-command-as-render discipline).
- The ledger directory is created on first capture if absent — the user never
  scaffolds it by hand.

## Phase Acceptance

> _Roll-up acceptance per [adr-9 amendment](../../decisions/adrs/0009-phase-as-product-layer.md). Each bullet asserts an emergent, cross-intent truth or a phase-spanning user journey — never a copy of an intent's own `## Acceptance Criteria`._

- **Given** an abcd-installed repo with no prior captures, **when** a user runs
  `/abcd:capture` three times and then `capture resolve` on one entry, **then**
  the ledger holds two entries under `open/` and one under `resolved/`, with
  stable IDs unchanged across the move — the directory-as-truth lifecycle
  working end to end.
- **Given** the rules loader is live (Phase 1), **when** a user's prompt
  carries an ISSUES-domain keyword and they then capture a finding, **then**
  the loader injects the issue-ledger discipline rules and the capture lands
  under that discipline — itd-3 and itd-4 composing into one coherent
  capture experience.

## Scope

**Intents:** itd-4 (`/abcd:capture` issue ledger — capture-only; cross-corpus
synthesis via `/abcd:dredge` is a separate, later-phase intent and is not in
scope here).

**Specs:** itd-4 is delivered across four flow-next specs rather than as a
monolithic build. `fn-20-issue-ledger-primitives-iss-n-allocator` ships the
library layer (`_issue_lib` + `issue_workflow` with zero command surface);
`fn-21-abcdcapture-command-flow-text-ingest`, `fn-22-workissuesmd-migration-promote-legacy`,
and `fn-23-intent-fidelity-reviewer-extension` build on top to add the
command flow, legacy migration, and reviewer cross-check respectively.

This is a deliberately small phase — one intent, one command. It is a phase of
its own rather than a rider on Phase 1 because capture is a distinct
user-capability moment with its own demoable milestone, and because the intent
authoring surface (Phase 3) is heavier and higher-stakes — the two "the user
records something" surfaces are kept apart so each phase milestone stays sharp.

**Brief plumbing-phases:** none of its own — `/abcd:capture`'s command flow is
covered by the brief's surface spec `04-surfaces/06-capture.md`. The probe-only
`capture list` stub from brief-Phase 1 becomes the real filtered query here.

## Maps against

- **Brief:** `04-surfaces/06-capture.md` (the command being built);
  `05-internals/03-configuration.md` (the `.abcd/development/activity/`
  namespace and visibility-driven gitignore).
- **Intents deliver the expectation:** itd-4 delivers the whole phase — the
  capture command, the ledger, and the folder-as-status lifecycle.
- **ADRs realised:** adr-3 (directory-as-truth — capture's `open`/`resolved`/
  `wontfix` folders ARE the status, no field stored).

## Dependency rationale

- **Runs after Phase 1** — capture writes into the rules loader's
  ARTEFACTS/ISSUES domain; itd-3's per-domain rule injection makes capture's
  discipline visible, and the ledger lives under the `.abcd/` namespace ahoy
  provisions. Capture has nothing to install of its own, so it cannot precede
  ahoy.
- **Independent of Phase 3** — capture and intent are sequenced (capture
  first, as the smaller surface) but have no hard dependency on each other.
  The one soft link: a captured issue can later be promoted toward an intent
  draft — but `capture promote` is itself part of itd-4's scope and does not
  require the intent phase to have shipped.

## Open questions

- `capture promote <iss-N>` produces an intent *draft* — confirm whether that
  draft must already satisfy the Phase 0 acceptance-gate discipline (itd-1) at
  promotion time, or whether a promoted draft is allowed to be incomplete until
  it reaches `/abcd:intent` in Phase 3. (Lean: incomplete is fine in `drafts/`;
  the gate bites at `/abcd:intent plan`, not at capture-promote.)
