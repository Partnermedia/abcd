# What Didn't

> **Status: LIVE.** Dead ends and abandoned approaches are recorded here as they are settled — from wontfix entries in the `.abcd/work/issues/` ledger, review pitfalls, and abandoned specs. The brief is the project's current state ([adr-5](../../decisions/adrs/0005-brief-is-current-state.md)): a dead end already closed with prejudice belongs here now. Lifeboat extraction *reads* this file and grounds what the record can prove; it is a reader, never the sole populator.

## Purpose

This file lists approaches that were tried during the abcd build (or a previous lifeboat) and failed, with the *why* attached. The next agent reading this file should be able to answer: "should I retry this with new tooling, or should I steer clear?"

## Format

For each entry:

```markdown
## <Approach name>

- **What was tried:** <short description>
- **Why it failed:** <the underlying property that made it fail>
- **When to retry:** <conditions under which the failure mode might no longer apply, e.g., new tooling, changed platform>
- **Source evidence:** <commit / review / memory key / issue pointing to the abandonment>
```

## Why this matters for next-agent design

Without this section, the next agent re-tries failed approaches because nothing tells them not to. With it, dead ends are documented with provenance — the agent gains time it would otherwise spend learning the same lessons.

## Related sources

Signals feeding this file:

- **wontfix entries in the [`.abcd/work/issues/`](../../../work/issues/) ledger** — the canonical home for explicit non-action decisions (`abcd capture wontfix`); an entry whose lesson generalises is promoted here
- **Review pitfalls** — the `.abcd/work/` area accumulates pitfall annotations that graduate here when they prove durable
