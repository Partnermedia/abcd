# What Worked

> **Status: PLACEHOLDER.** Evidence-shaped content (patterns that earned their keep, with the *why*) accumulates as the project progresses — this is the slot lifeboat extraction populates from `.abcd/memory/`, oracle reviews, and the working record under `.abcd/work/`. During development, leave empty; populated post-build by the disembark process.

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
