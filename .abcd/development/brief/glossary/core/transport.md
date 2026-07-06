<!-- Adapted from mattpocock/skills (MIT). See README Acknowledgements. -->
---
term: transport
bounded_context: core
definition: The mechanism by which curated context and artefacts are packaged and delivered to an oracle for review or reasoning.
aliases: ["context transport", "review transport"]
forbidden_synonyms: ["API call", "request", "prompt", "chat"]
status: stable
introduced_in: phase-1
starts_when: null
ends_when: null
not_to_be_confused_with: core/oracle
versions: null
---

# transport

The **transport** is the layer between the abcd pipeline and the oracle. It handles packaging
(selecting which files, diffs, and context snippets to include), delivery (sending to the oracle's
API or interface), and receipt (capturing the response for parsing). RepoPrompt is abcd's primary
transport for code review; other transports may be used for intent review or linting tasks.

## When to use

Use "transport" when referring to the mechanism used to send context to an oracle. The transport
is distinct from the oracle itself — different transports can deliver to the same oracle model,
and the same transport can route to different oracle models.

## When NOT to use

Do not call the transport an "API call" (too implementation-level), "prompt" (conflates content
with delivery), or "chat" (elides the structured review framing).

## Examples

- "The impl-review skill uses the RepoPrompt transport to send the scoped diff to the oracle."
- "The transport packages the frontmatter, diff, and schema into a single context window."

## Related terms

- [oracle](oracle.md) — the AI model that receives context through the transport
- [intent](intent.md) — one of the artefact types transported for review
