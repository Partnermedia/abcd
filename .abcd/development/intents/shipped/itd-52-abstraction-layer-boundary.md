---
id: itd-52
slug: abstraction-layer-boundary
spec_id: fn-42-abstraction-layer-boundary-warn-when-a
kind: standalone
suggested_kind: standalone
reclassification_history: []
related_adrs: [adr-15]
routed_from: ["fn-33:I-D1"]
created: 2026-06-03
updated: 2026-06-11
prd_path: null
---

# abcd Warns When You Reach Past It Into A Tool It Was Built To Hide

## Press Release

> **abcd is a configuration that bundles flow-next and other developer tooling and hides their complexity behind the `/abcd:*` surface — and now it notices, and gently warns, when someone reaches past that surface to drive a bundled dependency directly.** The whole proposition of an abcd configuration is that an operator works `/abcd:embark`, `/abcd:ahoy`, `/abcd:ralph-up` and never has to think about whether flow-next, Ralph, RepoPrompt, or codex is underneath. But nothing stops someone from invoking a dependency's own surface directly — `/flow-next:setup`, raw `flowctl`, a bare `/flow-next:*` skill — and when they do, they silently leave abcd's guarantees behind. One concrete case already bit: `/flow-next:setup` installs a thin `flowctl` that bypasses abcd's routing. This intent makes the boundary legible: when an operator drops below the abstraction into a bundled dependency, abcd surfaces a clear warning and points back at the `/abcd:*` path that wraps it — without ever blocking the legitimate flows where abcd itself drives a dependency under the hood.

> "I ran the dependency's own setup because the docs mentioned it, and quietly broke the routing abcd had set up for me," said Dave, a staff engineer adopting abcd. "I didn't want to go around abcd — I just didn't know I had. A one-line 'heads up, this reaches past abcd, the wrapped path is X' would have saved me an afternoon."

## Why This Matters

An abstraction layer that cannot tell when it is being bypassed is an abstraction in name only. abcd's contract with its operators is that the `/abcd:*` surface is where the guarantees live — routing, guards, safe orchestration. The moment someone drives a bundled dependency directly, those guarantees silently do not apply, and the failure is invisible until something downstream breaks (a bypassed dispatcher, a missing receipt, lost overlay state). The cost is always paid later, by someone who did not know they had stepped outside.

The hard part — and the reason this is an intent, not a one-line rule — is that abcd **legitimately drives those same dependencies under the hood.** `/abcd:ralph-up` runs after `/flow-next:ralph-init`; abcd's own flows call flow-next skills constantly. So the boundary is not "dependency surfaces are forbidden" — it is "a *person* reaching past abcd into a dependency is worth a heads-up; abcd driving that dependency itself is normal operation." A naive warning that fires on every dependency call would be noise and would punish the exact wrapped workflows abcd is built to provide. Getting that distinction right is the work.

## What's In Scope

- A boundary statement in the abcd contributor/agent documentation (the always-loaded instructions) declaring the principle: operators work the `/abcd:*` surface; bundled dependencies are implementation detail; reaching past abcd into a dependency surface is the thing to surface — distinguished from abcd driving a dependency under the hood, which is normal.
- An enumeration of the bundled-dependency surfaces (the flow-next skill set and the raw `flowctl` entry) classified as **sanctioned-when-abcd-drives** vs **warn-when-invoked-directly**.
- The concrete first guard: detect direct invocation of a dependency surface that has an `/abcd:*` equivalent or that mutates abcd-managed state (the `/flow-next:setup` case), and surface a warning that names the wrapped `/abcd:*` path.
- Reconciliation with existing documentation that already (correctly) instructs sanctioned dependency steps such as `/flow-next:ralph-init`, so the boundary statement does not contradict them.

## What's Out of Scope

- **Blocking dependency surfaces.** This is a warn-and-redirect boundary, not a lockout. An operator who knows what they are doing can proceed.
- **The specific `/flow-next:setup` bypass FIX.** That concrete repair (doctor probe + overlay coverage + corrected docs snippet) is fn-33 cluster I (I16); this intent is the general boundary principle, of which that fix is one instance.
- **A deterministic enforcement hook.** A hook that fires on direct dependency use is a possible later hardening; the first deliverable is the documented boundary + the single concrete warn case.

## Acceptance Criteria

> _Given-When-Then per the itd-1 discipline._

- **Given** the abcd instructions, **when** a contributor reads the surfaces section, **then** it states the abstraction-boundary principle and classifies bundled-dependency surfaces as sanctioned-when-abcd-drives vs warn-when-invoked-directly.
- **Given** an operator invokes a dependency surface directly that has an `/abcd:*` equivalent or mutates abcd-managed state, **when** abcd is in a position to notice, **then** a warning surfaces naming the wrapped `/abcd:*` path; **and given** abcd itself drives that same dependency under the hood, **then** no warning fires.
- **Given** the boundary statement, **when** it is checked against existing docs, **then** it does not contradict the already-sanctioned dependency steps (e.g. `/flow-next:ralph-init` before `/abcd:ralph-up`).
- **Given** the principle is stated, **when** a contributor looks for the concrete instance, **then** the `/flow-next:setup` bypass is cross-referenced to its fn-33 fix.

## Open Questions

> **Lifecycle note (fn-42):** this intent stays in `drafts/` — the rules.json
> heads-up (fn-42 R7) and the deterministic PreToolUse live-detection hook
> remain open; fn-42 only annotates the resolved Open Questions in place.

- Does the boundary live only in the always-loaded instructions, or also as a deterministic hook that fires on direct dependency invocation? (The instructions govern agent behavior; a hook governs every session including unattended ones.) *Partially resolved by fn-42: the boundary lives in the always-loaded instructions (`CLAUDE.md` § "The abcd abstraction boundary") + the static artifact probes; the deterministic PreToolUse hook stays deferred (adr-15).*
- ~~How is the bundled-dependency surface enumerated and kept current as flow-next adds or renames skills — a maintained list, or a pattern match on the dependency's command namespace?~~ **DECIDED (fn-42): a maintained map** — the classification table in `.abcd/development/brief/05-internals/04-universal-patterns.md` § 9, governing documentation + future heads-up keywords, not a static detector. A namespace-pattern auto-warn would need the deferred live hook to be non-vacuous (adr-15).
- ~~What exactly distinguishes "abcd is driving this dependency" from "a person reached past abcd" at detection time — an env marker abcd sets when it orchestrates, the call's parent context, or an explicit allowlist of abcd-internal call sites?~~ **RESOLVED by fn-37.3** (not fn-42): the process-scoped `--abcd-driven` argv sentinel, prepended by the session mirror's flowctl shim and consumed by the dispatcher in the single exec it decorates. The env-marker option was rejected as not inheritable-safe (adr-15 cites; fn-37.3 decided).
- Should the warning ever escalate (e.g. when the direct call is known to corrupt abcd-managed state), or always stay advisory?

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._

## References

- Concrete instance + fix: **fn-33** cluster I (I16, the `/flow-next:setup`
  dispatcher bypass).
- Source: design discussion generalizing the `/flow-next:setup` bypass into a
  governing principle (`.work/issues.md`, 2026-06-02/03).
- Principle kin: the abcd abstraction-layer boundary also underwrites **fn-37**
  (abcd enforces its guarantees at the `/abcd:*` surface; direct harness use is
  vanilla territory).
