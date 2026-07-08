# Loud staging

**The rule.** Unwired core code is permissible only while it announces itself
at every point of contact: a doc comment naming the wiring intent, a surface
error stating the capability is not wired at this stage, and a brief or
roadmap row scheduling it. Silent unwired code is dead scaffolding regardless
of its quality.

**Why.** This is the boundary "wired or it isn't done" actually needs.
Staging is sometimes correct — a contract worth landing a phase before its
publish path — and the difference between staging and scaffolding is whether
the code can be *trusted to explain its own status*. The 2026-07-08 review
drew the line empirically: the staged launch-ship core was downgraded by
adversarial verifiers precisely because it discloses loudly (doc comment,
CLI refusal message, scheduled intents), while a silently unreached status
function in the same review stayed a Major finding.

**Bounds.**

- Disclosure has three mandatory sites: the code (doc comment naming the
  intent), the surface (an explicit not-wired refusal, never a stub that
  half-works), and the record (the intent that will wire it).
- Loudness expires: when the named intent ships or is superseded, the staged
  code is wired or removed in that change. Disclosure is a lease, not a
  licence.

**Promotion.** No mechanical check distinguishes announced staging from
silent scaffolding yet; a coverage-plus-caller audit wired into the record
would promote this.
