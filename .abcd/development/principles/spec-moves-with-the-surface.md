# Spec moves with the surface

**The rule.** A verb, skill, or surface lands with its brief row in the same
change. The brief never trails the shipped surface: if a landing surface
violates a brief criterion, that same change either amends the criterion or
the surface does not land. There are no unspecified live verbs.

**Why.** The brief is what agents build from; every claim it makes is a
premise inherited by all future surface decisions. The 2026-07-08 review's
single Critical finding was this drift compounded: the brief asserts a skills
boundary that shipped skills already violate, undercounts its own command
surface, and gives implemented, user-reachable verbs no spec home at all — so
the specification actively misleads about the system it specifies. A sibling
project's audit offers the end state of unchecked spec-vs-shipped divergence:
ten built features discovered unreachable only by a later dedicated sweep.

**Bounds.**

- The row can be small — surface name, criterion it satisfies, wiring status.
  What is non-negotiable is same-change, not exhaustiveness.
- Script-first MVPs fronted by a plugin surface (per
  [script-first-mvp](script-first-mvp.md)) still get their row: the brief
  records the surface and its staging, which is what keeps one surface
  truthful across the later swap.

**Promotion.** A record-lint cross-check — every entry under `commands/` and
`skills/` resolves to a brief surface row — would promote this to a
discipline.
