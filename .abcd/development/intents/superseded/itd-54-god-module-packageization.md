---
id: itd-54
slug: god-module-packageization
spec_id: fn-53-god-module-packageization-behavior
kind: standalone
suggested_kind: standalone
reclassification_history: []
related_adrs: []
routed_from: ["fn-33:I-D4"]
created: 2026-06-03
updated: 2026-06-03
prd_path: null
---

> **⚠️ Superseded by [ADR-21](../../decisions/adrs/0021-rebuild-in-go.md)** (the Python god-modules this intent factored no longer exist; the from-scratch Go rebuild replaces them). Preserved as historical record per the supersession lifecycle.

# abcd's Largest Source Files Become Navigable Packages Without Changing A Line Of Behavior

## Press Release

> **abcd splits its handful of multi-thousand-line source modules into cohesive, responsibility-scoped packages — preserving every public entry point and every test import byte-for-byte, so the suite stays green at every step and behavior never changes.** A few abcd-owned files have grown past the point where a contributor (or an oracle working within a context budget) can comfortably load and reason about them: the intent-fidelity reviewer, the intent lint engine, the intent workflow, and `ahoy`. They are cohesive, not tangled — but their sheer size taxes navigation, inflates merge surface, and burns context. This intent factors them into responsibility/phase packages following the Common Closure principle (things that change together stay together), targeting cohesive modules rather than one-file-per-function. The dual public surface survives untouched: the `python -m scripts.abcd.<module>` entry points and the deep test-import sites keep working through a re-export façade, so a pure structural move is provably correct exactly when the existing suite stays green.

> "I open the fidelity reviewer and it's six thousand lines — I can't hold it in my head and neither can the oracle I ask to review my change," said Henry, a contributor making a small edit. "I don't want it rewritten. I want it in pieces that match how it actually thinks — collect, judge, render — so I can find the part I need and trust that moving it didn't change anything."

## Why This Matters

abcd values code that an agent can load and reason about within a context budget — that is part of how the framework keeps itself maintainable and reviewable. A 6,000-line module defeats that: every change pays a navigation and merge-surface tax, and every oracle review of that file consumes a large slice of its input budget on lines unrelated to the change. The cost is paid continuously, on every edit, by every contributor and every review.

The discipline that makes this safe rather than risky is that it is a **pure structural move, not a rewrite.** SOTA research on factoring large files (captured in the framework's research notes) concludes: split into cohesive responsibility/phase packages (the Common Closure principle), targeting roughly 500–1500-line modules — and explicitly *not* one-file-per-function, which scatters co-changing code and multiplies imports. The two public contracts must survive byte-for-byte: the `python -m` command entry point (preserved via `__main__.py`) and the deep test-import sites (preserved via an `__init__.py` re-export façade with explicit `__all__`). A move is correct precisely when the suite stays green at every commit — giving a mechanical, verifiable acceptance test for each step.

## What's In Scope

- Factor the largest abcd-owned modules into cohesive responsibility/phase packages, one file per intent-sized unit, starting with `ahoy` (the cleanest, already banner-delimited into detect/apply phases, near-zero cycle risk).
- For each target: preserve the `python -m scripts.abcd.<module>` entry point and every deep test-import site via a re-export façade; keep load-bearing fn-NN/itd-N banner comments through the move.
- Per-target verification: the existing suite is green at every commit; the zero-mutation lint and type/format checks pass; no behavior change.
- A confirming structural read of each non-`ahoy` target's actual seams before that target is split (the `ahoy` layout is validated against the live file; the others rest on structural maps and must be verified first).

## What's Out of Scope

- **Behavior change of any kind.** This is a move, not a refactor of logic; a pure move is correct iff the suite stays green.
- **Files that should stay whole.** Cohesive, agent-loadable single-concern modules with no clean multi-responsibility seam are explicitly left as-is (per the research note's leave-as-is list).
- **flow-next / Ralph code.** The vendored `flowctl.py` and the `scripts/ralph/` tree are external-dependency surface, not abcd-owned, and out of scope.
- **The `codex_invocation.py` duplication.** That is a single-source-of-truth dedup (fn-33 cluster I), not a split.

## Acceptance Criteria

> _Given-When-Then per the itd-1 discipline._

- **Given** a target module is factored into a package, **when** the existing suite runs, **then** it is green — at every individual commit of the split, not just at the end.
- **Given** the factored package, **when** a contributor invokes `python -m scripts.abcd.<module>`, **then** the command entry point behaves exactly as before the split.
- **Given** the factored package, **when** the deep test-import sites resolve, **then** every prior import path still resolves through the re-export façade with no change to call sites.
- **Given** a non-`ahoy` target, **when** it is scheduled for splitting, **then** a confirming structural read of its live seams is recorded first (the structural map is treated as a proposal to validate, not a settled design).
- **Given** the split is complete for a target, **when** the type/format/zero-mutation checks run, **then** they pass and no behavior has changed.

## Open Questions

- One spec per target file, or a single phase-boundary package-ization sweep spec covering all targets in sequence?
- File order: `ahoy` is the validated first candidate; the research note proposes `intent_fidelity_reviewer → intent_workflow → intent_lint → ahoy` by value — reconcile "safest first" (`ahoy`) against "highest-value first" (the reviewer).
- Coordination: this is a large mechanical change to the `scripts/abcd/` package — how is it scheduled so it does not collide with another agent mid-edit (e.g. fn-33's cluster-I work touches several of the same files)?
- Does the re-export façade keep a deprecation path for the old flat module name, or is the package the only public name once split?

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._

## References

- Source plan + SOTA sources: the factoring-large-files research note under
  `.abcd/development/research/notes/`; logged in `.work/issues.md` 2026-06-02 as
  the god-module finding + the package-ization spec-candidate.
- Coordination caveat: fn-33 cluster I touches some of the same files (guard,
  overlay, doctor) — sequence to avoid mid-edit collision.
- Not a split: `codex_invocation.py` dedup is fn-33 (single-source-of-truth),
  not this intent.
