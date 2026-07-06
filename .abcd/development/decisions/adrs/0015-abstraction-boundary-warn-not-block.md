---
id: adr-15
slug: abstraction-boundary-warn-not-block
status: superseded
date: 2026-06-11
supersedes: null
superseded_by: adr-22
related_intents: [itd-52]
related_rfcs: []
related_adrs: []
---

# ADR-15: The abstraction boundary warns, never blocks — argv-sentinel live discriminator, artifact-only static detection

> Superseded by [ADR-22](0022-bundled-deps-as-pluggable-adapters.md) — with
> native defaults there is no wrapped foreign surface for an operator to reach
> past, so the abstraction boundary is retired.

## Context

abcd bundles flow-next/Ralph, RepoPrompt, and codex behind the `/abcd:*`
surface. The proposition (itd-52) is that an operator never has to think about
what is underneath — but nothing stops a person from invoking a dependency's
own surface directly (`/flow-next:setup`, bare `flowctl`, a direct
`/flow-next:*` skill), and when they do, abcd's guarantees silently no longer
apply. One concrete case already bit: `/flow-next:setup` planted a thin
`.flow/bin/flowctl` that bypassed the abcd dispatcher (fixed in fn-33, I16).

The hard part is that abcd **legitimately drives those same dependencies under
the hood** — `/abcd:intent plan` calls `/flow-next:plan`; in-session Ralph
drives `flowctl.py` constantly. So the boundary is not "dependency surfaces are
forbidden"; it is "a *person* reaching past abcd is worth a heads-up; abcd
driving the dependency itself is normal." Any mechanism has to discriminate the
two without punishing the wrapped workflows abcd exists to provide.

Two structurally different boundaries fell out of the design work:

- **Live-call boundary** — "is THIS call abcd-driving or a person reaching
  past?" — decided at exec time.
- **Static-artifact boundary** — "did a reach-past leave a persistent bypassing
  artifact on disk?" — decidable after the fact by `ahoy doctor`.

## Decision

### 1. Warn-and-redirect, never block

A reach-past surfaces a warning that names the wrapped `/abcd:*` path; it is
never a lockout. An operator who knows what they are doing can proceed. The
boundary's job is legibility — "you have left the surface where the guarantees
live" — not enforcement.

### 2. The live discriminator is the process-scoped `--abcd-driven` argv sentinel — shipped by fn-37.3, cited here, not claimed

When abcd drives flowctl in-session, the fn-37.3 mirror shim
(`scripts/abcd/session/ralph_mirror.py`) prepends
`--abcd-driven --session-id <id>` to the argv of the real dispatcher
(`scripts/ralph/flowctl`), which consumes and strips it. The sentinel is
**process-scoped to that single exec** — never an ambient/inheritable env
marker, never persisted — so an interactive grandchild that later calls the
dispatcher directly carries no sentinel and is treated as a direct (vanilla)
call. This decision was made and shipped by **fn-37.3** (resolving itd-52
Open-Q#3); this ADR records it as the boundary's live discriminator, it does
not re-decide it.

### 3. Static detection is artifact-only — the doctor probes detect what persists, and only that

`ahoy doctor` probes run after the fact and read disk state. They therefore:

- **CAN detect** persistent bypass artifacts. Today: a `.flow/bin/flowctl`
  lacking the `ABCD_FLOWCTL_DISPATCHER` marker (`flow_bin_flowctl_probe` in
  `scripts/abcd/tools/doctor.py`, fn-33), and the `<!-- BEGIN FLOW-NEXT -->`
  marker block that `/flow-next:setup` writes into CLAUDE.md/AGENTS.md
  (`flow_next_marker_block_probe`, same module, fn-42.2 — the extend path
  held: the block was re-confirmed against the installed flow-next 1.13.0
  setup templates at implementation time; the predicate is
  fence-delimiter-anchored, never a bare `.flow/bin/flowctl` substring, so
  sanctioned dispatcher-warning prose does not trigger it).
- **CANNOT detect** the live process-scoped sentinel (it is gone when the probe
  runs) or any reach-past that leaves no artifact — a person running
  `/flow-next:plan` directly, bare `flowctl` outside Ralph, direct RP/codex
  use. Pretending otherwise would be a false-detection surface; the
  classification map (brief `05-internals/04-universal-patterns.md` § 9)
  records per-surface detectability honestly.

### 4. The deterministic live-detection hook is deferred

A PreToolUse hook that intercepts a direct dependency invocation at call time
is the only mechanism that could make live reach-past detection (and a
namespace-pattern auto-warn) non-vacuous. itd-52 named it as a possible later
hardening; this ADR keeps it deferred. Until it exists, live discrimination
exists only where the sentinel runs (the dispatch path), and everything else is
docs + the static artifact probes.

## Alternatives Considered

- **Block reach-past surfaces.** Rejected by itd-52's own scope: the boundary
  is warn-and-redirect, not a lockout. Blocking would punish legitimate
  expert use and the sanctioned composition points themselves.
- **An ambient env marker as the live discriminator** (abcd sets a variable
  when it orchestrates; callees check it). Rejected — not inheritable-safe: an
  exported marker leaks to every grandchild process, so an interactive shell
  spawned inside an abcd session would be misclassified as abcd-driven
  indefinitely. The argv sentinel is consumed by the single exec it decorates
  (fn-37.3's contract #1).
- **A namespace-pattern auto-warn over `/flow-next:*` slash-commands.**
  Rejected for now — a person running `/flow-next:plan` directly leaves no
  artifact distinguishable from `/abcd:intent plan` → `/flow-next:plan`, so a
  static warner would either false-positive on every sanctioned flow or detect
  nothing. It needs the deferred PreToolUse hook to be non-vacuous.
- **Chosen: warn-not-block + argv sentinel (live) + artifact-only probes
  (static) + maintained classification map (docs).** Honest about what each
  layer can see; zero false-detection surface.

## Consequences

- The classification of bundled-dep surfaces is a **documentation artifact**
  (the maintained table in `05-internals/04-universal-patterns.md` § 9), not a
  probe input — only its artifact-backed rows are machine-detectable.
- Each new persistent bypass artifact earns a sibling doctor probe (the fn-33
  probe is the template; fn-42.2 added the second,
  `flow_next_marker_block_probe`), keeping static detection congruent with
  the map.
- Non-artifact reach-past stays undetected until the PreToolUse hook lands —
  an accepted gap, recorded rather than papered over.
- The always-loaded instructions (`CLAUDE.md` § "The abcd abstraction
  boundary") state the principle and bless the five sanctioned composition
  points, so the boundary statement cannot contradict the sanctioned steps it
  coexists with.
- A future rules.json `BOUNDARY` heads-up (itd-52 / fn-42 R7) is deferred
  behind a rules-substrate harvest and a keyword-scoping design that avoids
  firing on sanctioned-flow prompts.

## Related

- [itd-52](../../roadmap/intents/drafts/itd-52-abstraction-layer-boundary.md) — the intent this realises (stays in `drafts/`; R7 + the live hook remain open)
- Sentinel (fn-37.3): `scripts/ralph/flowctl` (`--abcd-driven` handling), `scripts/abcd/session/ralph_mirror.py` (emission), `scripts/abcd/tools/_dispatch.py` (consumption), brief `05-internals/10-in-session-dispatch.md`
- Artifact probe (fn-33): `flow_bin_flowctl_probe` in `scripts/abcd/tools/doctor.py`, wired from `ahoy.py:_run_doctor`
- Classification map: `.abcd/development/brief/05-internals/04-universal-patterns.md` § 9
