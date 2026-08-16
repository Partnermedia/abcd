# Out of Phase Scope

This brief describes the work bundled into the seven planned phases (see [`roadmap/phases/README.md`](../../roadmap/phases/README.md)). **Later-phase items live as press-release intents**: the uncommitted bench in `.abcd/development/intents/drafts/` (enumerated below), and the committed-but-unscheduled intents in `planned/` — valid per [adr-34](../../decisions/adrs/0034-lifecycle-and-scheduling-orthogonal.md), listed in [intents/README.md](../../intents/README.md) § Planned, and scheduled when a phase doc's `## Scope` names them.

**In a later phase.** The set below is the live `drafts/` corpus — the
uncommitted bench. Per
[adr-34](../../decisions/adrs/0034-lifecycle-and-scheduling-orthogonal.md) no
phase-scoped intent lives in `drafts/` (scheduled ⇒ `planned/`), so there is
nothing to subtract: the filesystem is the list, and it is **not**
hand-counted —

```sh
# Live later-phase (uncommitted bench) IDs = the drafts/ corpus, no exclusions.
ls .abcd/development/intents/drafts/itd-*.md \
  | sed -E 's#.*/(itd-[0-9]+).*#\1#' | sort -V -u
```

Intents that have left `drafts/` (moved to `planned/`, `shipped/`, `superseded/`,
or `disciplines/`) are NOT in this list at all — the enumeration command cannot
emit them. (itd-31 and itd-32, both superseded and moved to `superseded/`, are
recorded only in the historical note at the end of this section, not here.)

The list is gated rather than trusted: the `index_drift` record-lint rule holds
the marked region to the ids in [`drafts/`](../../intents/drafts/), so a capture,
a promotion, or a supersession that does not edit this list fails the record
gate. That is what keeps "not hand-counted" true after the day it was written.

<!-- index: later-phase-intents -->
- `itd-8` — `--with-code` bundling (lifeboat carries source code)
- `itd-9` — Cross-version lifeboat schema migration
- `itd-10` — `/abcd:ahoy destroy` deeper uninstall
- `itd-11` — Pass B transcript-noise mitigation
- `itd-12` — `.abcd/work/notes/` distiller weighting
- `itd-13` — Scheduled `dev-sync` (cron / launchd)
- `itd-14` — Prompt registry + versioning (heavier successor to itd-5)
- `itd-15` — Self-dogfooded SOTA audit (recurring per-disembark sibling to itd-5)
- `itd-16` — `/abcd:audit` umbrella + chain substrate (default application: hash-chain over conversation/edit history; reframed as umbrella on 2026-05-08, lifeboat-integrity application extracted to itd-35)
- `itd-17` — Per-backend per-agent oracle effectiveness tracking
- `itd-18` — `.claude/settings.local.json` permission templates
- `itd-19` — ABCDevelopment stage-aware defaults
- `itd-21` — `/abcd:init-project` empty-repo scaffolding
- `itd-22` — Harness portability — the shared adaptor machinery for multi-harness support (host profile seam, adaptor ladder, parity suite) under the host-tier policy ([adr-39](../../decisions/adrs/0039-host-tier-policy.md), mechanism per [adr-23](../../decisions/adrs/0023-transport-agnostic-core.md)); per-host adoption files as its own intent at an explicit decision; awaiting the planning interview
- `itd-23` — Spec Kit interop
- `itd-25` — `/abcd:dredge` cross-corpus synthesist (split from itd-4 capture)
- `itd-26` — `/abcd:loot` OSS-vendor with provenance (pulled to an earlier phase on 2026-05-08)
- `itd-30` — Design fictions as an alternative intent capture format (`--format=fiction`)
- `itd-33` — Agent-communication infrastructure (multi-agent coordination via `.abcd/coordination/`)
- `itd-35` — `/abcd:audit lifeboat <path>` lifeboat-integrity verification (sibling sub-verb under itd-16's umbrella; captured 2026-05-08)
- `itd-39` — Scope-aware memory retrieval (extends itd-3's recall hook to the memory store)
- `itd-41` — Phase negotiator — Socratic phase-proposer (per [adr-10](../../decisions/adrs/0010-phase-negotiator-grounded-tradeoffs.md))
- `itd-44` — A fourth intent kind for infrastructure choices the product thinker wants to record
- `itd-51` — Harness-adoption-readiness rubric ("safe enough to adopt" before a new harness arrives)
- `itd-55` — abcd can tell whether its own reasoning rests on bedrock or an unexamined assumption
- `itd-57` — Manual-hold sentinel blocking a spec from autonomous pickup until a human lifts it
- `itd-59` — Autonomous-run passes leave the same durable, queryable transcript an interactive session does
- `itd-60` — Doc-fidelity anti-drift: a spec cannot close until the brief and public docs reflect what was built
- `itd-61` — Brief-change derivation: a human brief edit reconciles its implied intents and principles before commit
- `itd-62` — Pluggable fail-closed safety gate wrapping a trusted scanner
- `itd-64` — Benchmark-driven configuration optimisation from abcd's own runs
- `itd-70` — Launch release retention (newest-per-line prune of superseded releases)
- `itd-75` — CLI eval harness: fixture-driven proof the CLI actually runs
- `itd-76` — Source-provenance ledger: consult a source you cannot cite, and every debt is recorded
- `itd-77` — Relocatable user-level home
- `itd-78` — Intent-dependency graph: what to build first, even when it is something small
- `itd-82` — Ledger drain: one verb triages the open ledger into work that ships itself and work a human must think about
- `itd-83` — The review bar fires by itself, in every repo abcd manages
- `itd-85` — Read-only repo-conformance audit
- `itd-86` — Cold-reading surface: abcd reads its own design documents as a stranger would
- `itd-87` — Recurrence escalation in capture: a finding that returns after closure is kept, not discarded
- `itd-90` — Brief interview for the blanks: the product thinker is handed only the questions they alone can answer
- `itd-91` — AI-attribution preference declared once and followed by every commit, PR, and record
- `itd-92` — Branch-protection verification on managed repos, gating the launch
- `itd-97` — The facilitator is a mode, not a person
- `itd-98` — Solo vs duo is measured, not debated
- `itd-99` — A team of product thinkers decides as one
- `itd-106` — abcd sets up the CI a repo requires, and reports what it did
- `itd-107` — Autonomous routines assemble from one versioned template
- `itd-108` — The plugin installs from the curated release artifact, not the repo, and every cut release reaches users automatically
- `itd-109` — Acceptance criteria verify themselves; the manual rest renders for a human (`abcd verify`, sha-keyed receipts)
- `itd-110` — The grill interview renders with clear structure and colour
- `itd-112` — Generated ferry-style CLI banner from the identity block, for abcd and managed CLIs
- `itd-113` — The MCP front door opens — abcd's core verbs from any MCP-capable harness (the [adr-39](../../decisions/adrs/0039-host-tier-policy.md) universal floor)
<!-- /index -->

**Later-phase items with no intent id.** These four were written into the brief
before they were captured as intents, so no `itd-N` derives them and they sit
outside the gated list above. Each is either superseded by a decision or waiting
for a capture pass:

- `.abcd/work/issues/` ledger cleanup bundle (sweep the workshop before a later phase)
- abcd warns when you reach past it into a tool it was built to hide — **obsolete under no-hard-deps ([adr-22](../../decisions/adrs/0022-bundled-deps-as-pluggable-adapters.md))**: with native defaults there is no wrapped foreign surface to reach past; the abstraction boundary is retired
- abcd's largest source files become navigable packages without changing behaviour
- One command re-vendors upstream and restores the abcd overlay in a single guarded step — **obsolete under no-hard-deps ([adr-22](../../decisions/adrs/0022-bundled-deps-as-pluggable-adapters.md))**: no external tool re-vendors itself onto abcd's state, so there is no overlay to re-apply

**Phased-in additions captured post-brief (2026-05-07):** itd-27 (`/abcd:intent grill` sub-verb + glossary), itd-28 (spec-tied reviews in the native spec review store), and itd-34 (three intent kinds with three lifecycle paths) were captured after this brief was written and are scoped into the planned phases. They are listed in `intents/README.md` and the relevant phase docs; this section is canonical for **later-phase** items only and does not enumerate phased-in intents.

**Later-phase additions captured post-brief (2026-05-07):** itd-30, itd-31, itd-32, and itd-33 were captured in the same audit pass. itd-30 and itd-33 remain in the later-phase list above; itd-31 and itd-32 have since been superseded (itd-31 absorbed by itd-48; itd-32 superseded by itd-31) and moved to `intents/superseded/`, so they are no longer in the canonical later-phase set above — this note records their capture timing and supersession for the brief's history. (See `superseded/itd-31-cross-document-fidelity-reviewer.md` and `superseded/itd-32-audit-role-taxonomy.md`.)

Each intent captures the press-release-shaped scope and acceptance criteria. A later-phase intent enters work by being scoped into a phase, then promoted to `planned/` via `/abcd:intent plan <itd-N>` and to `shipped/` via `/abcd:intent ship <itd-N>` (or automatically when the linked spec closes).

The brief does not get re-versioned. What has shipped is defined by which phases are complete and which intents are in `shipped/`; this brief stays the canonical current-state design record.
