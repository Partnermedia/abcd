# Out of Phase Scope

This brief describes the work bundled into the six planned phases (see [`roadmap/phases/README.md`](../../roadmap/phases/README.md)). **All later-phase items live as press-release intents** in `.abcd/development/intents/drafts/` — see [intents/README.md](../../intents/README.md) for the full index.

**In a later phase.** The set below is the live `drafts/` corpus minus the
intents already scoped into a planned phase (a phase doc's `## Scope` section is
the single source of truth for which intents a phase bundles — see
[adr-9](../../decisions/adrs/0009-phase-as-product-layer.md)). It is **not**
hand-counted: derive set membership from the filesystem and subtract the
phased-in IDs, rather than maintaining a total that re-drifts —

```sh
# Live later-phase IDs = drafts minus phased-in.
# The command globs ONLY drafts/, so intents that have already left drafts/ for
# planned/ | shipped/ | disciplines/ (e.g. itd-6 planned, itd-27/28 shipped,
# itd-37 a discipline, and itd-20/24/63/69 planned under spc-83) are excluded by
# the filesystem itself — they can never appear in the output. The exclusion list
# below is therefore ONLY the phased-in IDs that are STILL physically in drafts/
# (lifecycle move pending): itd-2,3,4,7
# (phase-0/1/2 scoped) and itd-34,36,40,42 (later phased-in, captured post-brief).
ls .abcd/development/intents/drafts/itd-*.md \
  | sed -E 's#.*/(itd-[0-9]+).*#\1#' | sort -V -u \
  | grep -vxE 'itd-(2|3|4|7|34|36|40|42)'
```

The IDs struck through below (itd-46/48/49/50/53) are **shipped but still physically in
`drafts/`** — the lifecycle `drafts/` → `shipped/` move is pending, so the
enumeration command still emits them; the strikethrough is the authoritative
*status*, the enumeration command is the authoritative *set membership*. Intents
that have fully left `drafts/` (moved to `planned/`, `shipped/`, `superseded/`,
or `disciplines/`) are NOT in this list at all — the command cannot emit them,
and the bullet list is kept in lockstep with the command output. (itd-31 and
itd-32, both superseded and moved to `superseded/`, are recorded only in the
historical note at the end of this section, not here.)

- itd-8 — `--with-code` bundling (lifeboat carries source code)
- itd-9 — Cross-version lifeboat schema migration
- itd-10 — `/abcd:ahoy destroy` deeper uninstall
- itd-11 — Pass B transcript-noise mitigation
- itd-12 — `.abcd/development/activity/notes/` distiller weighting
- itd-13 — Scheduled `dev-sync` (cron / launchd)
- itd-14 — Prompt registry + versioning (heavier successor to itd-5)
- itd-15 — Self-dogfooded SOTA audit (recurring per-disembark sibling to itd-5)
- itd-16 — `/abcd:audit` umbrella + chain substrate (default application: hash-chain over conversation/edit history; reframed as umbrella on 2026-05-08, lifeboat-integrity application extracted to itd-35)
- itd-17 — Per-backend per-agent oracle effectiveness tracking
- itd-18 — `.claude/settings.local.json` permission templates
- itd-19 — ABCDevelopment stage-aware defaults
- itd-21 — `/abcd:init-project` empty-repo scaffolding
- itd-22 — OpenCode harness implementation — **obsolete under no-hard-deps ([adr-22](../../decisions/adrs/0022-bundled-deps-as-pluggable-adapters.md))**: with a transport-agnostic Go core ([adr-23](../../decisions/adrs/0023-transport-agnostic-core.md)) behind thin front doors, a second harness is just another host over the same core, not a special port
- itd-23 — Spec Kit interop
- itd-25 — `/abcd:dredge` cross-corpus synthesist (split from itd-4 capture)
- itd-26 — `/abcd:loot` OSS-vendor with provenance (pulled to an earlier phase on 2026-05-08)
- itd-29 — Autonomous-run resilience (rate-limit recovery, spec rewind, branch lifecycle)
- itd-30 — Design fictions as an alternative intent capture format (`--format=fiction`)
- itd-33 — Agent-communication infrastructure (multi-agent coordination via `.abcd/coordination/`)
- itd-35 — `/abcd:audit lifeboat <path>` lifeboat-integrity verification (sibling sub-verb under itd-16's umbrella; captured 2026-05-08)
- itd-39 — Scope-aware memory retrieval (extends itd-3's recall hook to the memory store)
- itd-41 — Phase negotiator — Socratic phase-proposer (per [adr-10](../../decisions/adrs/0010-phase-negotiator-grounded-tradeoffs.md))
- itd-43 — Spec-terminology rename (one canonical word for a specced block of work) — **in flight on spc-65; the remaining sweep.** spc-7 shipped only the atomic `epic_id`→`spec_id` field rename ([adr-11](../../decisions/adrs/0011-spec-terminology-rename.md)); the broader surface/prose/glossary sweep was deliberately parked in this intent and is now delivered by **spc-65** (lint enforcement + prose sweep). The intent's `spec_id` points at spc-65. Stays in `drafts/` until spc-65 closes.
- itd-44 — A fourth intent kind for infrastructure choices the product thinker wants to record
- `.work/issues.md` cleanup bundle (sweep the workshop before a later phase)
- ~~itd-46~~ — `/abcd:intent "<text>"` ↔ `/abcd:capture "<text>"` symmetric create paths — **shipped in spc-30 (LIVE).** Draft retained pending the `drafts/` → `shipped/` lifecycle move.
- itd-47 — spc-12's oracle-backed gates pass honestly without a human in the loop (not yet shipped)
- ~~itd-48~~ — `intent-fidelity-reviewer` gains its cross-doc + kind-classification roles (absorbed itd-31) — **shipped in spc-29 (Roles 2 + 3).** Draft retained pending the `drafts/` → `shipped/` lifecycle move.
- ~~itd-49~~ — Flow-state drift becomes visible before it compounds — **shipped in spc-41 (LIVE; the `FS` flow-state-drift family, `FS001`).** Draft retained pending the `drafts/` → `shipped/` lifecycle move.
- ~~itd-50~~ — The audit loop drives an intent to acceptance — or calls for a replan — **shipped in spc-52 (LIVE; the record-only ↔ loop-to-acceptance audit policy).** Draft retained pending the `drafts/` → `shipped/` lifecycle move.
- itd-51 — Harness-adoption-readiness rubric ("safe enough to adopt" before a new harness arrives)
- abcd warns when you reach past it into a tool it was built to hide — **obsolete under no-hard-deps ([adr-22](../../decisions/adrs/0022-bundled-deps-as-pluggable-adapters.md))**: with native defaults there is no wrapped foreign surface to reach past; the abstraction boundary is retired
- ~~itd-53~~ — A shipped intent no longer drifts out of audit just because nobody ran the review — **shipped in spc-43 (LIVE; the `RC` review-completeness family).** Draft retained pending the `drafts/` → `shipped/` lifecycle move.
- abcd's largest source files become navigable packages without changing behavior
- itd-55 — abcd can tell whether its own reasoning rests on bedrock or an unexamined assumption
- One command re-vendors upstream and restores the abcd overlay in a single guarded step — **obsolete under no-hard-deps ([adr-22](../../decisions/adrs/0022-bundled-deps-as-pluggable-adapters.md))**: no external tool re-vendors itself onto abcd's state, so there is no overlay to re-apply

**Phased-in additions captured post-brief (2026-05-07):** itd-27 (`/abcd:intent grill` sub-verb + glossary), itd-28 (spec-tied reviews in the native spec review store), and itd-34 (three intent kinds with three lifecycle paths) were captured after this brief was written and are scoped into the planned phases. They are listed in `intents/README.md` and the relevant phase docs; this section is canonical for **later-phase** items only and does not enumerate phased-in intents.

**Later-phase additions captured post-brief (2026-05-07):** itd-30, itd-31, itd-32, and itd-33 were captured in the same audit pass. itd-30 and itd-33 remain in the later-phase list above; itd-31 and itd-32 have since been superseded (itd-31 absorbed by itd-48 and shipped in spc-29; itd-32 superseded by itd-31) and moved to `intents/superseded/`, so they are no longer in the canonical later-phase set above — this note records their capture timing and supersession for the brief's history. (See `superseded/itd-31-cross-document-fidelity-reviewer.md` and `superseded/itd-32-audit-role-taxonomy.md`.)

Each intent captures the press-release-shaped scope and acceptance criteria. A later-phase intent enters work by being scoped into a phase, then promoted to `planned/` via `/abcd:intent plan <itd-N>` and to `shipped/` via `/abcd:intent ship <itd-N>` (or automatically when the linked spec closes).

The brief does not get re-versioned. What has shipped is defined by which phases are complete and which intents are in `shipped/`; this brief stays the canonical current-state design record.
