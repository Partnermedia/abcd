<!-- Adapted from mattpocock/skills (MIT). See README Acknowledgements. -->
---
term: spec
bounded_context: core
definition: A specced block of work tracked in .flow/ that implements one or more intents, broken into ordered tasks with acceptance criteria.
aliases: ["flow spec", "implementation spec"]
forbidden_synonyms: ["sprint", "milestone", "project", "feature", "epic"]
status: stable
introduced_in: phase-1
starts_when: null
ends_when: null
not_to_be_confused_with: core/intent
versions: null
---

# spec

A **spec** is the implementation unit in abcd's workflow. Where an intent describes *what* and
*why* in human/customer terms, a spec describes *how* in technical terms. Each spec is tracked
under `.flow/specs/` and decomposed into numbered tasks (`.flow/tasks/`).

## When to use

Use "spec" when referring to a specced technical work block with an `fn-N` identifier. Specs have
acceptance criteria, task lists, and are driven to completion by the Ralph loop or direct
implementation.

## When NOT to use

Do not call a spec a "feature" (too generic), "sprint" (carries Scrum cycle connotations), or
"milestone" (a milestone is the end condition of a [phase](phase.md), not an individual work
block) or "phase" (a phase bundles many specs).

## Examples

- "Spec `fn-3` implements the grill skill and glossary infrastructure."
- "Task `fn-3.1` is the first implementation task of that spec."

## Related terms

- [intent](intent.md) — the human-authored document that a spec realises
- [voyage](voyage.md) — a full lifecycle that contains many specs
