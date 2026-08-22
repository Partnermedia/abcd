# What Worked

> **Status: LIVE.** Evidence-shaped content (patterns that earned their keep, with the *why*) is recorded here as it accrues — from `.abcd/memory/`, reviews, and the working record under `.abcd/work/`. The brief is the project's current state ([adr-5](../../decisions/adrs/0005-brief-is-current-state.md)): a lesson already learned belongs here now, not after a disembark run. Lifeboat extraction *reads* this file and grounds what the record can prove; it is a reader, never the sole populator.

## Purpose

This file lists patterns from the abcd build (or a previous lifeboat) that earned their keep, with the *why* attached. Reading it as the next agent should answer: "should I retain this pattern, or is there a better way available now?" — not "I must reuse this exactly."

## Format

For each entry:

```markdown
## <Pattern name>

- **What it did:** <short description>
- **Why it worked:** <the underlying property that made it succeed>
- **Caveat / when to revisit:** <conditions under which this might not hold>
- **Source evidence:** <commit / review / memory key / fixture pointing to the proof>
```

## Why this is *evidence*, not *prescription*

Patterns listed here are advisory. A future agent reading this file should treat them as informed defaults that earned their keep — but is free to propose alternatives if the platform, dependencies, or constraints have shifted (see [`02-constraints/01-platform.md`](../02-constraints/01-platform.md)). Architectural prescription belongs in [`02-constraints/`](../02-constraints), not here.
