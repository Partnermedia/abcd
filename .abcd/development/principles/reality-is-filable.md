# Reality is filable

**The rule.** The record's taxonomy must always be able to express the
current true state of the work. When something real cannot be filed — a
state with no directory, a transition with no representation — the taxonomy
is the defect, and it is fixed before more reality accrues.

**Why.** A record that cannot express what happened goes quiet exactly when
it matters most, and a quiet record reads as "nothing happened". The
2026-07-08 review found the live instance: v0.1.0 shipped working capability
from a drafts-stage intent while `shipped/` sat empty — the lifecycle had no
way to represent delivered work mid-stream, so the record's most important
fact (things shipped) was invisible to anyone reading the lifecycle
directories. Unfilable reality also breeds workarounds: facts get parked in
commit messages, changelogs, or heads, where the record's own tooling cannot
see them.

**Bounds.**

- The fix is a deliberate taxonomy change (and an ADR when it shapes the
  record's architecture), not an ad-hoc folder invented at filing time.
- "Fix the taxonomy" and "file the backlog" are separable: the state must
  become representable immediately; migrating stragglers can trail as an
  issue.

**Promotion.** A lifecycle lint — e.g. no CHANGELOG delivery entry whose
intent still sits in `drafts/` — would promote this to a discipline.
