# Decompose Before Filing

**The rule.** One proposal is rarely one record. Before an intent is filed,
the proposal is decomposed into its parts and each part is routed to its
record home — user-facing capability → an intent; trust-boundary rule → an
ADR (plus a brief invariant); standing stance → a principle; plumbing → the
brief. The records the proposal touches are surfaced with **typed** links —
`supersedes` / `reverses` / `duplicates` / `refines` — never a vague
"related". The analysis is **advisory**: it renders one of three outcomes
(FILE-AS-IS / SPLIT / HOLD) as a proposal, and a human adopts the routing.
A reversal ("this reverses invariant X") is only ever *flagged* for the human
to confirm — never auto-classified, never auto-filed.

**Why.** The record information architecture
([adr-30](../decisions/adrs/0030-record-information-architecture.md)) supplies
typed homes; the failure mode is filing a monolith into one of them. The
founding case: the 2026-07-13 auto-merge review found one "feature" was four
record types — an experience, a trust rule, a stance, and plumbing — and only
one was an intent. A monolith denies the trust rule its ADR scrutiny, buries
the reusable stance, and bloats the intent's scope; naming decomposition as a
capture-time rule makes the routing deliberate instead of accidental. The
reversal flag stays advisory because contradiction detection is unreliable
even on frontier models and over-flags by default — the most valuable check
is the one that must never hold a gate shut on its own.

**The ladder.** This file is the principle rung. The enabling MVP beneath it
is the hand-run four-piece protocol documented in the `/abcd:intent` surface
page, plus the deterministic Go pre-pass (lexical candidate-finder +
atomicity smell) when it lands. The enforced discipline above it is
[itd-84](../intents/disciplines/itd-84-intent-decomposition.md), whose
automated capture-time validator is deferred until the hand-run protocol is
calibrated (~50 graded captures, under the
[itd-81](../intents/disciplines/itd-81-judge-calibration.md) judge-calibration
discipline). Until that rung ships, the documented protocol is the gate,
announced as not-yet-automated ([loud-staging](loud-staging.md)).
