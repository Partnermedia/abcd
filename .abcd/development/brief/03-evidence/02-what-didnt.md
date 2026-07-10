# What Didn't

> **Status: PLACEHOLDER.** Evidence of dead ends and abandoned approaches accumulates as the project progresses; lifeboat extraction populates this slot from wontfix entries in the `.abcd/work/issues/` ledger, oracle review pitfalls, and abandoned specs. During development, leave empty; populated post-build by the disembark process.

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

## Related sources during build

While this file is empty during development, related signals already exist:

- **the `.abcd/work/issues/` ledger** — running log of issues discovered during the build (see the abcd-CLAUDE.md "Mandatory Issue Recording" rule)
- **wontfix entries in the `.abcd/work/issues/` ledger** — once `abcd capture` ships (itd-4), this becomes the canonical home for explicit non-action decisions
- **Oracle review pitfalls** — the `.abcd/work/` area accumulates pitfall annotations that lifeboat extraction promotes here
