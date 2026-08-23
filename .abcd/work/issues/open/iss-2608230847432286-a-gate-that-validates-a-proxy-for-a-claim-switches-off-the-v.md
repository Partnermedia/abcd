---
schema_version: 1
id: "iss-2608230847432286"
slug: "a-gate-that-validates-a-proxy-for-a-claim-switches-off-the-v"
severity: "major"
category: "process"
source: "user-observation"
found_during: "record-review"
found_at: ".abcd/development/principles/enforcement-claims-are-facts.md"
details: "enforcement-claims-are-facts covers the phantom gate: a check described but not running, whose harm is that readers stop compensating. Two instances found on 2026-08-22/23 show a second form the principle does not yet name, where the gate is real and runs but validates a proxy for the claim, so the compensating vigilance is switched off by a green tick rather than by an absent one. Proposed as a paragraph extending that principle with the two instances as worked examples, not as a new principle, per one-canonical-primitive."
suggested_fix: "Extend .abcd/development/principles/enforcement-claims-are-facts.md with a paragraph naming the proxy-gate form and its two worked examples. Do not add a new principle beside it: one-canonical-primitive forbids the third copy, and the Why paragraph of the existing principle already states the mechanism this shares. A maintainer decides adoption; two agents agreeing is not the gate."
related_issues: ["iss-2608221457227162", "iss-2608230752354926"]
---

a gate that validates a proxy for a claim switches off the vigilance an absent gate would have preserved

`enforcement-claims-are-facts` states the rule for a gate that does not exist:
a phantom gate is worse than no gate, because readers who believe a check exists
stop compensating with the vigilance they would otherwise apply.

Two findings from 2026-08-22/23 are the same failure with the opposite
mechanism, and the principle does not yet name it.

**Absent enforcement.** `research/notes/01-harness-interface.md` carried
`Status: Accepted (Phase 0 lock)` and the title `ADR-01` while filed in the
evidence folder. No gate ran on that status claim at all, because
`record_schema` is scoped to the configured record stores and a file outside
them is not a record it can see. Recorded as iss-2608221457227162, with the
detector gap as iss-2608230752354926.

**Enforcement measuring a proxy.** The `cli-verb-taxonomy-restructure` ideate
verdict marked a leg 1 claim `verified` that was wrong in its denominator. Its
own errata states the lesson exactly: "`ideate record` proves every cited id
resolves, and proves nothing about whether a claim is true. A `verified` mark is
a claim about the author's diligence, not an assertion the binary checked."
See [the verdict's errata](../../../development/research/notes/2026-08-22-ideate-cli-verb-taxonomy-restructure.md).

The second is the nastier form. An absent gate leaves a reader with nothing to
trust, so some vigilance survives. A gate that runs and reports green on the
wrong property actively withdraws the vigilance, and supplies false assurance in
its place. The existing principle's own reasoning — that the harm is the
withdrawal of compensating attention — applies with more force here, which is
why this belongs inside it rather than beside it.

Neither instance alone shows the pair. The absent-gate case reads as a filing
mistake; the proxy-gate case reads as one bad claim. Together they identify a
class: a status field is a claim, not evidence, and a gate that checks a claim's
form has not checked its truth.

Routing is left open deliberately. The paragraph is the cheapest rung, but
whether the class also warrants a detector — a rule that flags a status claim no
gate cross-checks — is a maintainer call. Two agents agreeing that a principle
should change is not the gate that changes it.
