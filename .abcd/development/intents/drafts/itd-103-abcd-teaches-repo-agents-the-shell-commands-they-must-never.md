---
id: itd-103
slug: abcd-teaches-repo-agents-the-shell-commands-they-must-never
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
impact: additive
---

# **abcd teaches repo agents the shell commands they must never run — and blocks them when they try.** An agent that runs `cd scratch && rm -rf *` is one failed `cd` away from deleting the working tree, and the facilitator watching the session usually has no way to know that. abcd ships a hazard registry: each entry a dangerous command pattern, its severity, the safe successor, and a plain-language why. One registry, two planes. The rules loader injects the matched safety rules before shell-heavy work, so the agent is taught the safe form up front; a deterministic guard — a core verb any harness hook can call — refuses a matching command at execution time and replies with the safe successor, so the block itself is the lesson. Hosts without hook support still get the teaching plane; hosts with hooks get both, wired in at install. The registry grows from reality: a facilitator who sees something scary needs to know exactly one move — capture it — and recurring captures promote patterns into the bundled defaults, each entry shipping with a fixture that proves its guard fires. "I couldn't have told you why that command was dangerous," said Nia, a facilitator. "I didn't have to. The guard refused it, told the agent what to run instead, and told me why in words I understood. The one time something new scared me, I captured it and moved on."

## Press Release

> _Seeded from a quoted-text intent capture. Expand into the full press-release narrative before planning._

## Why This Matters

**abcd teaches repo agents the shell commands they must never run — and blocks them when they try.** An agent that runs `cd scratch && rm -rf *` is one failed `cd` away from deleting the working tree, and the facilitator watching the session usually has no way to know that. abcd ships a hazard registry: each entry a dangerous command pattern, its severity, the safe successor, and a plain-language why. One registry, two planes. The rules loader injects the matched safety rules before shell-heavy work, so the agent is taught the safe form up front; a deterministic guard — a core verb any harness hook can call — refuses a matching command at execution time and replies with the safe successor, so the block itself is the lesson. Hosts without hook support still get the teaching plane; hosts with hooks get both, wired in at install. The registry grows from reality: a facilitator who sees something scary needs to know exactly one move — capture it — and recurring captures promote patterns into the bundled defaults, each entry shipping with a fixture that proves its guard fires. "I couldn't have told you why that command was dangerous," said Nia, a facilitator. "I didn't have to. The guard refused it, told the agent what to run instead, and told me why in words I understood. The one time something new scared me, I captured it and moved on."

## Acceptance Criteria

> _Required (the itd-1 discipline): add at least one Given-When-Then bullet describing the verifiable bar for "shipped" before this draft can be planned._

## Open Questions

_None recorded yet._

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
