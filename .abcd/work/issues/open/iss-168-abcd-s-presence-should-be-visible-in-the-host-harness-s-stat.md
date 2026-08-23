---
schema_version: 1
id: "iss-168"
slug: "abcd-s-presence-should-be-visible-in-the-host-harness-s-stat"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "itd-105 grill session"
found_at: "commands/abcd"
related_intents: [itd-20]
related_issues: [iss-164, iss-165]
---

abcd's presence should be visible in the host harness's status line / command line: a managed repo shows an 'abcd-managed' indicator (and possibly guard health) so the user can tell at a glance whether the current session is under abcd management without running a command. Needs a per-host adapter (e.g. a statusline hook) with the usual basics-built-in stance.

Design direction (maintainer, 2026-08-10), stated in harness-neutral terms: two layers over one fact. A host-agnostic default render lives in core — a single canonical line that both layers format, so no front door improvises its own words — and it surfaces ON DEMAND ONLY, at the `/abcd` board (the bare render and its `status` positional alias, `commands/abcd.md`). It is never streamed to stdout on every command and never injected into ordinary session output. Some harnesses show a persistent status bar and some offer no such surface at all; where one exists, ambient continuous display is exclusively its job, and where none exists the default layer stands alone rather than degrading to nothing. The default layer's contract is that a user who asks gets the answer, not that the answer is permanently on screen. The enhancement is an automatic option — self-detected, never configured. Adapters land one harness at a time, in a maintainer-set roster order (first the harness abcd already ships lifecycle hooks for, then opencode, then others much later). `/abcd:status` is not a separate command and must not become one: `status` is already a positional alias for the same bare render, and itd-20 owns that board's expansion, which is why this issue's `found_at` is `commands/abcd`.

Detection is INSTALL-time, not render-time. Where a harness has a status surface, it works by invoking a command and displaying its stdout, so by render time the harness has identified itself by calling — the enhancement therefore lives in the wiring, not in the print path, and the render can take the host as a flag or read whatever the harness supplies on stdin. What `ahoy install` must self-detect is which harness configurations are present, and therefore which wirings to write. Detection fails closed: no recognised harness means the default layer alone, and `unknown` is a legitimate terminal answer rather than a guess. Absence of a harness-supplied environment variable proves nothing — the plugin-root ladder is `ABCD_PLUGIN_ROOT` -> a harness-specific plugin-root variable -> executable-ancestor fallback (`internal/core/ahoy/store.go:85`), so a dev shim or an explicit `ABCD_PLUGIN_ROOT` yields no harness variable while running inside that very harness. Detection rests on positive evidence, never on a missing variable, and core has no host abstraction today to hang it on.

Reusable already: `internal/core/ahoy/guard_health.go` computes the guard-health half (`PluginRootResolved`, `HookInstalled`, `BinaryReachable`, with reasons) and `detect.go` classifies `managed-repo`, so the fact is largely assembled — what is missing is the one-line canonical render and the per-host wiring. Constraint to settle once: adapter ids in code and config are not user-facing prose and are unaffected, but any `docs/` page naming a harness trips the harness-naming blocker (`harness/claude-code`, `.abcd/docs-lint.json:14`) and needs generic phrasing or the `docs-lint: allow` escape.