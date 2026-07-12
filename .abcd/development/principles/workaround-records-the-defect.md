# Workaround records the defect

**The rule.** When abcd itself breaks — an error, wrong or empty output, a missing
or retired path, a surprising behaviour — and a manual workaround is taken to keep
moving, the workaround is not complete until the underlying defect is recorded as a
ledger issue (`abcd capture`). Working around it to unblock is fine; a *silent*
workaround is not, because the workaround is the only trace the defect ever showed
itself, and taking it without recording erases that trace. Recording the issue is the
non-negotiable minimum — at least enough to investigate later.

**Why.** abcd improves only on recorded evidence; an unrecorded defect is invisible
and is never fixed. A manual workaround is exactly the moment the system revealed a
gap *and* the moment that signal is most likely to be lost — the immediate problem is
unblocked, attention moves on, and the defect survives to bite the next user. Capturing
it converts a private patch into a public, drainable work item, and (composing with
[fix-the-detector](fix-the-detector.md)) into a candidate detector plus its
acceptance corpus.

**Bounds.**

- Binds on defects in **abcd** — the binary, the plugin surface, the record's own
  tooling. Friction from the host harness, or an operator's own mistake (a mistyped
  command, a shell pipe that swallowed the output), is not an abcd defect: discern the
  source and capture only the abcd case, so the ledger stays signal, not noise.
- "Record" is the floor, not the ceiling. A one-line capture naming the failing
  behaviour and its site (`file:line` if known) is enough to move on; a genuine one-off
  gets a plain issue, a shared root cause gets the fix-the-detector treatment.
- Applies to any abcd use, not only formal reviews — and the `/abcd:run` loop captures
  the abcd defects it hits mid-run rather than only working around them.
- Composes with [reality-is-filable](reality-is-filable.md) (the ledger's taxonomy must
  be able to express the true broken state — an unfilable defect is itself a defect),
  and [enforcement-claims-are-facts](enforcement-claims-are-facts.md) (a broken gate is
  recorded, never quietly bypassed).

**Live instance.** The 2026-07-12 dogfooding session moved forward on three workarounds
and left a ledger trace for each: the always-latest binary wrapper worked around abcd
having no dev/track-latest install mode → iss-75; memory-lint writing to the retired
`.abcd/logbook/` → iss-73; a duplicate `iss-56` id → iss-74. The host-harness quirks
hit the same session (a guard-hook blocking a `gh` call, an empty `/plugin details`)
were deliberately *not* captured — not abcd defects.

**Promotion.** Principle now (entry rung). The enabling convention beneath it is the
existing `abcd capture` ledger (**exists**). The discipline/tool rung above it is
**absent**: a check that a session or run which took a workaround emitted a matching
capture is hard to make fully mechanical, so the realistic MVP toward it is a run-loop
self-report (the `/abcd:run` protocol asserts "defects hit → captured" in its handoff)
or a review-checklist rule, before any gate.
