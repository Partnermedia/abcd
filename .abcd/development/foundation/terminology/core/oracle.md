<!-- Adapted from mattpocock/skills (MIT). See README Acknowledgements. -->
---
term: oracle
bounded_context: core
definition: An AI model invoked via a structured transport to review, reason over, or validate artefacts produced during a voyage.
aliases: ["reviewer", "AI reviewer"]
forbidden_synonyms: ["LLM", "model", "bot", "AI", "chatbot"]
status: stable
introduced_in: phase-1
starts_when: null
ends_when: null
not_to_be_confused_with: core/transport
versions: null
---

# oracle

An **oracle** is abcd's abstraction for an AI model used in a review or reasoning role. The oracle
receives context through a transport (e.g., RepoPrompt) and returns a structured verdict
(SHIP / NEEDS_WORK / MAJOR_RETHINK). Referring to it as "oracle" rather than "LLM" or "AI"
emphasises its role as a decider rather than a generator.

## When to use

Use "oracle" when referring to the AI model invoked for a review step in the abcd pipeline. The
oracle is always invoked through a named transport — never directly.

## When NOT to use

Do not call the oracle an "LLM", "model", "bot", or "AI" in workflow documentation — these terms
elide the transport and role distinction. Do not confuse the oracle with the transport that
delivers context to it.

## Examples

- "The impl-review skill sends the diff to the oracle via the RepoPrompt transport."
- "The oracle returned NEEDS_WORK on round 6."

## Related terms

- [transport](transport.md) — the mechanism that delivers context to the oracle
- [intent](intent.md) — one of the artefact types that oracles review
