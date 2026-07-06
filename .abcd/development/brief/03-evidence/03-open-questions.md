# Open Questions

> **Status: PLACEHOLDER.** Open architectural questions accumulate during build; lifeboat extraction populates this slot from unresolved review threads, RFCs, and issue ledger items tagged "open-question". During development, leave empty; populated post-build by the disembark process.

## Purpose

This file lists architectural questions that the team did NOT resolve — left open deliberately or simply unfinished. The next agent reading this file should treat these as legitimate degrees of freedom: things the original team punted on, where a fresh design pass is permitted (or expected).

## Format

For each entry:

```markdown
## <Question>

- **Status:** open (not resolved)
- **What's at stake:** <consequences of resolving one way vs another>
- **Current best guess:** <if any — clearly labelled as guess, not decision>
- **Source:** <RFC / review thread / issue ledger entry>
```

## Why this is separate from `02-what-didnt.md`

`what-didnt.md` records *settled* dead ends: approaches that were tried, failed, and abandoned with prejudice. `open-questions.md` records *unsettled* design questions: things the original team didn't try at all, or tried inconclusively, or deferred. The next agent reads them differently — closed-with-prejudice vs open-for-fresh-thought.

## Related sources during build

- **`.abcd/development/research/adr/`** — Architecture Decision Records. Open questions that get resolved promote into ADRs.
- **`.abcd/development/roadmap/rfcs/`** — Request for Comments. Multi-stakeholder discussion artefacts (open / resolved-yes / resolved-no).
- **`.work/issues.md`** entries flagged as open questions or future-work seeds.
