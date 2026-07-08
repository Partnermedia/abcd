# Intent dependency sweep — method, findings, and lessons for the assessor

First full application of the [itd-78](../../intents/drafts/itd-78-intent-dependency-graph.md)
two-axis model (declared severity + edges, derived priority) to the whole live
corpus: 64 intents across `drafts/`, `planned/`, and `disciplines/`, assessed
by seven parallel readers proposing severity and evidence-quoted edges, with
human ratification before anything landed. This note records what the exercise
taught — it is the seed material for a future `abcd intent assess` flow.

## Method

- Each assessor read its batch completely and proposed: severity (one-line
  rationale), `blocked_by` and `builds_on` edges each carrying a verbatim
  evidence quote and a confidence mark (high = explicit in the intent's own
  prose; low = inferred), plus free-form anomaly notes.
- Auto-landing was confidence-gated: high-confidence edges landed directly;
  every severity and every low-confidence edge went to human review. Withheld
  regardless of confidence: edges into superseded targets and edges whose
  evidence read in the opposite direction.
- One extra reader mapped the phase docs' `## Scope` sections — the
  sequencing ground truth (adr-9) the graph is checked against, never a thing
  the graph overrides.

## Outcome

- Severity landed on all 64 live intents: 5 critical, 25 major, 32 minor,
  plus 2 nitpick on the two retired files. ~50 edges landed.
- Two priority inversions computed immediately: itd-74 (minor, effective
  major — blocks itd-76) and itd-20 (minor, effective major — blocks itd-33).
  The derived build-first order reproduced what two grilling sessions had
  concluded by hand.
- Two dead intents retired to `superseded/` (itd-47 per adr-22, itd-49 per
  adr-26); persona-principle violations corrected in itd-7, itd-43, itd-74.

## Lessons for the assessor (what to look out for)

1. **Most "blockers" are not intents.** The dominant hard-dependency targets
   are specs (`spc-N`), ADRs, brief internals, and core substrate (`ahoy`,
   `dev-sync`, the spec store). Genuine intent→intent hard edges are rare
   (~15 across 64). The schema stays intent-only; non-intent prerequisites
   live in prose. An assessor must resist "promoting" a spec dependency into
   the nearest related intent.
2. **Direction flips are the top error mode.** Four proposals carried evidence
   that literally stated the reverse edge ("the field itd-14 builds on" —
   offered as itd-5 *building on* itd-14). Rule: the evidence must be phrased
   from the dependent's perspective; when a file says "X sits on top of me",
   the edge belongs on X. Reciprocal confirmation (target file names the
   dependant) is the strongest signal.
3. **Superseded targets hide in live edges.** Three planned intents hard-cited
   a dead draft (itd-47). Lint rule: `blocked_by` must not point at
   `superseded/`; supersession should force re-pointing to whatever absorbed
   the capability.
4. **Edges surface lifecycle contradictions the lifecycle lint cannot see**,
   because they are content-level: a supersession banner in `drafts/`, an
   "implementation complete" table in `planned/`, a "do not implement from
   this draft" banner on a file that is another intent's blocker.
5. **Build the roster from one commit.** Two intents (itd-7, itd-43) fell
   through the sweep because directory listings ran on two branches straddling
   a pending lifecycle move. Derive the roster with `git ls-tree` at a single
   commit, never from working-directory listings taken at different times.
6. **Severity ambiguity clusters on mission framing.** The genuinely hard
   calls (itd-62, itd-51, itd-75) all hinge on *which product thesis* the
   weight is measured against. Rule: severity is weight under the brief's
   stated mission, not any single sub-thesis; an intent with no code surface
   can still be major.
7. **Cross-kind edges want a different relation.** Disciplines relate to the
   corpus by inheritance/conformance ("applies to every spec"), not
   sequencing; only the format base (itd-1) is a true blocker. Keep the edge
   schema two-kinded; record discipline application in prose until something
   machine-consumes it.
8. **Recording deliberate non-dependencies works.** itd-76's prose note that
   itd-16 is *not* a blocker prevented a false edge at assessment time —
   the cheapest lint there is.
9. **Evidence quotes are what make ratification cheap.** Reviewing 64
   assessments without re-reading 64 intents is only possible because every
   proposed edge carries the phrase it stands on. The high/low confidence
   split held: nothing auto-landed needed reversal at review.

## Follow-ups (capture as ledger issues after the triage branch merges — the
issue ledger lives on that unmerged branch, so numbering must not fork)

- itd-66: "NON-CANONICAL — do not implement" banner vs its role as itd-65's
  blocker; reconcile with the native spec that owns the contract.
- itd-50: "implementation complete" reconciliation table while in `planned/`;
  verify and move to `shipped/` through the audit-notes path if true.
- itd-37 is Phase-0-scoped with a Phase-2 blocker (itd-36); split the
  hard-on-half edge or reschedule.
- Phase-3 doc asserts itd-27 depends on itd-42; itd-42's prose asserts the
  reverse (and the landed edge follows the intent). Adjudicate — this is the
  first live catch for the itd-78 phase-consistency lint.
- The launch cluster (itd-65/66/67 critical, itd-72 major) is scheduled in no
  phase's `## Scope`.
- "itd-launch" cited by name, not number, in itd-8 and itd-16 — dangling
  reference class the graph lint should flag.
- itd-6 sits in `planned/` under an adr-25 supersession framing; itd-36 is
  partially superseded by adr-28 — both want a lifecycle look.
- Persona sweep of the remaining corpus (assessors only spot-checked).
