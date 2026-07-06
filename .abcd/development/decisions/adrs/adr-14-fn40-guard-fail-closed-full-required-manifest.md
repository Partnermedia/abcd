---
id: adr-14
slug: fn40-guard-fail-closed-full-required-manifest
status: accepted
date: 2026-06-10
supersedes: null
superseded_by: null
related_intents: []
related_rfcs: []
related_adrs: []
---

# ADR-14: Guard degraded fallback fails closed to the full required manifest

## Context

The 2026-06-02 security review left one HIGH finding open against the
abcd-owned protected-path PreToolUse guard
(`scripts/abcd/hooks/abcd_ralph_guard.py`): when its fragment layer is
**degraded** — the loader errors, a fragment is malformed or conflicting, or
the merged required tier comes out empty — the guard silently shrank from ~50
protected paths to a **4–5 path hardcoded minimum**. Every other abcd-owned
path (overlay scripts, `codex_invocation.py`, the control-plane modules, the
fragment files themselves) became writable under `YOLO=1`. A guard that reads
as "protected" while protecting almost nothing is the worst failure mode: the
operator believes the control plane is enforced, and nothing surfaces the
shrink.

The shrink was also **self-weakening**: the fragment files that define
protection were among the paths left writable, so a degraded state could be
made permanent by rewriting the fragments.

Constraints already locked when fn-40 was decided:

- The upstream half (`scripts/ralph/hooks/ralph-guard.py`) is external-dep —
  never forked, out of scope.
- Interception stays on the AGENT tool surface (Edit/Write/Bash-write) only;
  abcd's own python writers must not be self-blocked.
- The memory note `fn-370-spike-outcome-b` rules out any fallback that
  degrades security ("no fallback that degrades security").

## Decision

When the fragment merge is degraded/untrusted, the guard falls back to the
**full declared required manifest**, computed INDEPENDENTLY of the corrupt
merge: `protected_paths.required.json` (`include_globs` − `exclude_globs`) ∪
the baseline `REQUIRED_UPSTREAM`/`REQUIRED_ABCD` tiers ∪ hardcoded structural
self-protection (`protected_paths.d/**`, `_protected_paths.py`, the manifest
itself) ∪ the hardcoded floor. The floor is never removed: with every data
source unreadable, the fallback still blocks the flowctl wrappers, `ralph.sh`,
and the guard — never an empty set.

The fail-closed is **scoped — integrity, not coverage** (Saltzer–Schroeder
fail-safe defaults): a degraded merge means we don't *know* the true protected
set, so we deny over the full declared *required* set. We do not deny over the
whole tree (a self-inflicted DoS — the "blocker you can't clear" trap), and
the OPTIONAL tiers stay best-effort under degrade. Membership is decided by
string-pattern matching over the normalized candidate path, not filesystem
enumeration — so a write *creating* a required-but-absent file is blocked, and
enumerating the possibly-corrupt fragment dir to decide what protects that dir
(which would reopen the self-weakening loop) is avoided. The fallback runs
before the seal gate — under a loader outage the seal cannot be trusted
either.

## Alternatives Considered

1. **Fail open to empty (pre-fn-33.8 behaviour).** Loader failure → protect
   nothing. Rejected outright: it converts any fragment corruption — including
   deliberate corruption — into a full bypass of self-modification protection,
   and it is exactly the security-degrading fallback the memory note forbids.
2. **Keep the 4–5 path hardcoded floor as the whole fallback (status quo).**
   Simple, no new data dependency at degrade time. Rejected: it silently
   under-protects ~90% of the declared required set, including the fragment
   files themselves (self-weakening), and the operator gets no signal that
   enforcement shrank. This is the HIGH finding, not a fix for it.
3. **Fall back to the full required manifest, floor beneath (chosen).** The
   required set is recomputable WITHOUT trusting the corruptible fragments
   (reference-monitor tamperproof property): the manifest and baseline are
   read directly from their own files. Costs accepted: a degraded state may
   over-block (a path a healthy fragment `remove`s is still seen via the
   manifest/baseline), and a deliberately-partial dev fragment set now blocks
   rather than silently passing.

A fourth option — block *everything* under degrade — was dismissed during
planning rather than seriously considered: it false-blocks legitimate
operation (including the repair path's own collateral writes) and trades a
silent under-protect for a loud self-DoS.

**Three-clause test:**

- Hard to reverse? **Yes** — the fault-injection suite
  (`tests/abcd/test_sibling_fail_closed.py`) and the documented guarantee
  (`scripts/abcd/hooks/README.md`, overlay README) now pin degraded behaviour
  to the full manifest; weakening it means re-accepting the HIGH finding.
- Surprising without context? **Yes** — "on error, block MORE than the healthy
  merge would" (over-block of removed paths, owner-task blocks before repair)
  is counter-intuitive without the integrity-not-coverage frame.
- Real trade-off? **Yes** — full-manifest fail-closed trades availability under
  degrade (owner tasks blocked, removed paths over-blocked) for integrity of
  the control plane.

## Consequences

- **Stronger posture:** a corrupted/missing fragment layer can no longer be
  used to unprotect the control plane; the files that define protection are
  themselves unconditionally in the degraded set, closing the self-weakening
  loop.
- **A deliberately-partial dev fragment set now blocks rather than silently
  passing.** Developers exercising the guard with a stripped-down fragment dir
  hit the full required set; tests use explicit fixture dirs instead.
- **Over-block under degrade is accepted:** a path legitimately `remove`d by a
  healthy fragment is still blocked while degraded (the fallback reads
  baseline+manifest, not the merge). Safe direction; repaired by fixing the
  fragment layer.
- **Owner-in-progress tasks are blocked under degrade** (fallback runs before
  the seal gate). The repair path runs human-driven with `FLOW_RALPH` unset,
  so it stays reachable.
- **OPTIONAL tiers remain best-effort:** e.g. `scripts/abcd/dep_watcher.py`
  stays writable under degrade — the stated scope cut (integrity = restore the
  REQUIRED set), asserted as deliberate in the test suite.
- New obligation: the degraded warning and the docs
  (`scripts/abcd/hooks/README.md`, overlay README fail-closed subsection) must
  track the matcher's actual sources if they ever change.

## Related

- `scripts/abcd/hooks/README.md` — the fail-closed guarantee, operator-facing
- `scripts/abcd/overlay/README.md` — tier semantics + fail-closed subsection
- `tests/abcd/test_sibling_fail_closed.py` — fault-injection coverage pinning
  the behaviour
- `.flow/specs/fn-40-guard-fail-closed-on-integrity.md` — the implementing spec
