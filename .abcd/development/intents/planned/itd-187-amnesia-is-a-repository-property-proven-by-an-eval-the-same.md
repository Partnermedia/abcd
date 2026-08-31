---
id: itd-187
slug: amnesia-is-a-repository-property-proven-by-an-eval-the-same
spec_id: spc-65
kind: standalone
suggested_kind: standalone
reclassification_history: []
builds_on: []
severity: minor
impact: additive
---

# Amnesia is a repository property, proven by an eval — the same state assembled twice is byte-identical, so no case run is spent evidencing it

## Press Release

> **Amnesia is a property of what the assembler passes, not an instruction
> to an agent.** The eval assembles one definition twice over an unchanged
> repository state and asserts the two assembled inputs are byte-identical
> — the manifest sits outside the comparison, carries content hashes and
> no timestamps, and the assembler walks paths in lexicographic order.
> Making this a repository eval means any reader can check it, and the
> closing run of Iteration 2 carries only the properties a case run can
> carry (purpose durability and convergence — never amnesia).

## What's In Scope

- The double-assembly comparison in CI, with the identity relation stated:
  byte-equality of the assembled input, manifest excluded.
- The determinism preconditions it enforces on the assembler: hash-only
  manifests, lexicographic walk order.

## Scope Conditions

None stated.

## Acceptance Criteria

- **Given** an unchanged repository state at one commit, **when** one definition
  is assembled twice from two distinct filesystem paths, **then** the two
  assembled inputs are byte-identical, the manifest excluded from the
  comparison.
- **Given** the order-adversarial fixture, **when** the eval runs, **then** the
  item paths in the assembled input agree with the eval's own lexicographic
  sort, so a consistent-but-not-lexicographic order fails.
- **Given** a manifest carrying a timestamp-shaped key or a timestamp-shaped
  scalar value, **when** the eval runs, **then** it fails.
- **Given** two artefacts differing only in item order, and two differing only in
  one scalar value, **when** the comparator runs over each pair, **then** it
  reports a difference naming the differing item.

**Disclosed residue (ac-2 to ac-4).** A nondeterminism introduced into the
shipped assembler is a precondition no eval may establish for itself, because an
eval must not patch the code under test. The three criteria above catch each
named nondeterminism through an oracle the assembler does not supply, and the
comparator's own capacity to fail is proved by ac-4. What remains is discharged
by hand: the walk sort removed by a one-line local patch, the test watched red,
the patch reverted before the branch is pushed, and the run recorded in the
pull-request body. That is a recorded hand-run, not a standing gate.


## Grounds

- pursued: Amnesia is a property of what the assembler passes, not an instruction an agent can be trusted to follow, and a case run could only ever exhibit it rather than prove it. It is pursued now because a case run is the scarcest thing in the cycle: making amnesia a repository eval leaves the closing run of Iteration 2 carrying only the properties a case run can carry, purpose durability and convergence, and lets any reader check the rest for themselves.
