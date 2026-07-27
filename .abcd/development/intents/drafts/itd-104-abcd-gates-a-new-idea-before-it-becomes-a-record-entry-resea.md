---
id: itd-104
slug: abcd-gates-a-new-idea-before-it-becomes-a-record-entry-resea
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
impact: additive
---

# **abcd gates a new idea before it becomes a record entry: research it against primary sources, grill it against the existing record, and let an independent adversary try to kill it.** A product thinker's exciting idea is cheapest to kill in its first hour — and most expensive to kill after it has minted intents, specs, and branches. `/abcd:ideate` runs the admission interview: SOTA research in which every load-bearing claim is checked against its primary source; a record grill — does the brief, an intent, an ADR, or a principle already cover, contradict, or supersede this?; and an independent adversarial review, evaluator outside the loop. The output is a verdict plus a recorded decision with rejected alternatives; only a surviving idea graduates to a draft intent. The protocol is already proven by hand: three ideas entered it on 2026-07-14 — two were killed (one by a measured null result its premise could not survive, one by three independently fatal methodological defects) and one was reframed by adversarial review before a line was built. The manual run also named the two load-bearing steps automation must keep: check the record first (one idea was already written and superseded), and open the primary (three secondary-source claims were falsified, one of them fabricated by tool extraction). "The ideas I'm proudest of are the ones abcd killed in an afternoon," said Iris, a product thinker. "Each one left a recorded reason behind, so I never talk myself back into them six months later."

## Press Release

> _Seeded from a quoted-text intent capture. Expand into the full press-release narrative before planning._

## Why This Matters

**abcd gates a new idea before it becomes a record entry: research it against primary sources, grill it against the existing record, and let an independent adversary try to kill it.** A product thinker's exciting idea is cheapest to kill in its first hour — and most expensive to kill after it has minted intents, specs, and branches. `/abcd:ideate` runs the admission interview: SOTA research in which every load-bearing claim is checked against its primary source; a record grill — does the brief, an intent, an ADR, or a principle already cover, contradict, or supersede this?; and an independent adversarial review, evaluator outside the loop. The output is a verdict plus a recorded decision with rejected alternatives; only a surviving idea graduates to a draft intent. The protocol is already proven by hand: three ideas entered it on 2026-07-14 — two were killed (one by a measured null result its premise could not survive, one by three independently fatal methodological defects) and one was reframed by adversarial review before a line was built. The manual run also named the two load-bearing steps automation must keep: check the record first (one idea was already written and superseded), and open the primary (three secondary-source claims were falsified, one of them fabricated by tool extraction). "The ideas I'm proudest of are the ones abcd killed in an afternoon," said Iris, a product thinker. "Each one left a recorded reason behind, so I never talk myself back into them six months later."

## Acceptance Criteria

- Given a one-line idea, when the user runs the intent or capture verb, then no ideate step is required or suggested as blocking — ideate is an optional verb, never a pre-capture gate; the routing help names it for big, unproven ideas.
- Given /abcd:ideate runs, then it executes three legs in order: primary-source research (every load-bearing claim checked against its primary), a record grill (does the brief, an intent, an ADR, or a principle already cover, contradict, or supersede this?), and an adversarial review that is fresh-context and off-policy — conducted by a session that did not do the research, receiving the idea as an artefact of unknown authorship.
- Given a verdict, then the decision is recorded with rejected alternatives whether the idea survives or dies; only a surviving idea graduates to a draft intent.
- Given the record grill leg runs, then a hit on an existing record entry (covered, contradicted, or superseded) is cited by id in the verdict — the check-the-record-first step that killed one of the three ideas in the proven manual run.

## Open Questions

- The verdict's schema and where it is stored (ledger entry vs a dedicated ideate record family) — a spec-time decision.
- Whether ideate can consume a lifeboat or external document as the idea source.

## Grill Settlements (2026-07-27)

- Optional verb, never a gate: capture friction stays at one line, and the routing help carries the nudge.
- The adversarial leg codifies the fresh-context, off-policy, unknown-authorship pattern — the only measured debiasing effect, per the salvaged 2026-07-14 research record.

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
