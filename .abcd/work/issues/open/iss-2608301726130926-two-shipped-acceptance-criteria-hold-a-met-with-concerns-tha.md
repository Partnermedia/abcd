---
schema_version: 1
id: "iss-2608301726130926"
slug: "two-shipped-acceptance-criteria-hold-a-met-with-concerns-tha"
severity: "major"
category: "observation"
source: "user-observation"
found_during: "phase-4-conditional"
origin: researcher-authored
production_mode: hand-written
found_at: ".abcd/development/intents/shipped"
---

two shipped acceptance criteria hold a MET_WITH_CONCERNS that becomes NOT_MET if itd-185 does not land in this phase

The facilitator's verdict rule, now filed as itd-192, says a criterion whose
producer does not exist is `MET_WITH_CONCERNS` if THIS PHASE wires it, and
`NOT_MET` if not. Two verdicts were issued under the first branch:

- itd-178 `ac-2`
- itd-180 `ac-1`

Both are `MET_WITH_CONCERNS` **because spc-63 / itd-185 is in scope**. itd-185
is the ingest door, and as of 2026-08-30 it sits in `intents/planned/` and has
not been started. So the condition under which those two verdicts were issued
is not yet satisfied, and it is the phase, not the criterion, that decides it.

**If itd-185 does not land in this phase, both flip to `NOT_MET` and both
verdicts need re-issuing.** That is not a judgement call at the time; it is the
rule itd-192 states, applied to a fact the phase will have settled.

This record exists because the obligation had no marker. It lived in one line of
a handover file and in the verdict's own prose, both of which are read by
whoever happens to be reading them. A verdict that silently keeps a rating whose
stated precondition failed is exactly the shape the fidelity audit exists to
catch, so leaving the check to memory would be the audit failing in its own
terms.

**Trigger: Phase 4, before the pull request to `main`.** Resolve it one of two
ways and record which:

- itd-185 landed: both verdicts stand as issued, and this closes on that fact.
- itd-185 did not land: re-issue both verdicts as `NOT_MET`, with the reason a
  promised intent went unwired, and close this citing the re-issue.

Do not close it by deciding the criteria are fine. The rule does not leave that
open.

