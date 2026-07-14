<!-- Adapted from mattpocock/skills (MIT). See README Acknowledgements. -->
---
term: brief
bounded_context: core
definition: The root document that defines a project's purpose, constraints, and success criteria before any implementation begins.
aliases: ["project brief", "brief doc"]
forbidden_synonyms: []
status: stable
introduced_in: phase-1
starts_when: null
ends_when: null
not_to_be_confused_with: core/intent
versions: null
---

# brief

The **brief** is the authoritative root document for a project. It establishes what the project
exists to do, the constraints it operates under, and the success criteria by which completion is
judged. It is written by a human stakeholder and treated as immutable once approved.

## When to use

Use "brief" when referring to the top-level project specification document that lives at the root
of the `.abcd/` hierarchy. A project has exactly one brief.

## When NOT to use

Do not use "brief" to describe an intent (which is feature-scoped) or a spec (which is
implementation-scoped). The brief is project-wide; intents and specs are narrower.

## Examples

- "The brief says the target persona is Carol, not Alice."
- "This intent is out of scope for the brief's Phase 1 boundary."

## Related terms

- [intent](intent.md) — a press-release-shaped feature description within the project scope
- [voyage](voyage.md) — the operations namespace recording what abcd did to produce a lifeboat; a
  [disembark](disembark.md) run grounds the brief's structure section by section
