---
id: itd-187
slug: amnesia-is-a-repository-property-proven-by-an-eval-the-same
spec_id: null
kind: null
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

## Acceptance Criteria

- **Given** an unchanged repository state, **when** one definition is
  assembled twice, **then** the two assembled inputs are byte-identical.
- **Given** a nondeterminism introduced into the assembler (walk order,
  timestamps), **when** the eval runs, **then** it fails.

