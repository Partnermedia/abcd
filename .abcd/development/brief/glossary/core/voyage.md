<!-- Adapted from mattpocock/skills (MIT). See README Acknowledgements. -->
---
term: voyage
bounded_context: core
definition: A complete project lifecycle managed by abcd, running from initial brief through final delivery, encompassing all intents, specs, and oracles for a single project.
aliases: ["project lifecycle", "abcd voyage"]
forbidden_synonyms: ["project", "engagement", "sprint"]
status: stable
introduced_in: phase-1
starts_when: The brief is approved and committed to the repository.
ends_when: All scoped intents have been promoted, implemented, and the final release tag is cut.
not_to_be_confused_with: core/spec
versions: null
---

# voyage

A **voyage** is abcd's term for the full end-to-end lifecycle of a project. It begins when the
brief is approved and ends when the final release is tagged. The voyage encompasses all bounded
contexts, all intents, all specs, and all oracle reviews produced during that project. The
metaphor emphasises that the project has a defined departure (brief approval) and arrival
(release tag), with a navigable sequence of steps between them.

## When to use

Use "voyage" when referring to the full lifecycle of a project managed under abcd. The term
helps distinguish the overall arc from individual specs or milestones, which are stations along
the route.

## When NOT to use

Do not call a voyage a "project" (too generic), "sprint" (too narrow — implies a fixed-time box),
or "release" (a release is an event within a voyage, not the voyage itself). Do not confuse a
voyage with a spec.

## Lifecycle

| Phase | Condition |
|-------|-----------|
| Starts when | Brief is approved and committed to the repository |
| Ends when | All scoped intents implemented; final release tag cut |

## Examples

- "This voyage covers the first release of abcd."
- "At the end of the voyage, a retrospective captures what the team learned."

## Related terms

- [brief](brief.md) — the document that initiates the voyage
- [spec](spec.md) — a specced work block that is a station in the voyage
- [intent](intent.md) — a feature commitment that belongs to the voyage's scope
