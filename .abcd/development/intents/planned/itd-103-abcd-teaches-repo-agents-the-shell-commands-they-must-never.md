---
id: itd-103
slug: abcd-teaches-repo-agents-the-shell-commands-they-must-never
spec_id: spc-16
kind: standalone
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

- Given the guard hook cannot execute the abcd binary, when any command runs, then the guard fails OPEN with an unmissable in-session warning, and guard health is reported by ahoy status — never a silent no-op, never a bricked session.
- Given a command matches a blocker-tier registry entry, then it is refused with the safe successor as the block message; warn-tier matches pass with the warning injected; no in-session override exists — the only escape is a committed per-repo config override.
- Given a hazard pattern appears inside a quoted string argument (for example a ledger capture whose text mentions a dangerous command), when the guard evaluates, then it does not fire: matching is shell-token-aware and applies in command position only, including cd-chain structure across compound separators.
- Given a registry entry is proposed for the bundled defaults, then it ships with known-bad and known-good fixtures (known-good at least 40% of its corpus) and clears a declared true-negative-rate floor before admission; string payloads inside eval or shell -c are a documented v1 gap, not a silent one.

## Open Questions

- The registry config's file home (extend rules.json with a guard domain vs a dedicated committed file) — a spec-time decision.
- Which non-shell tool calls (if any) the guard family later covers.

## Grill Settlements (2026-07-27)

- Fail-open-loud on guard breakage; blocker/warn tiers mirror the docs-lint family; overrides are committed and reviewable only.
- Shell-token-aware matching in command position was chosen precisely because a raw-regex guard would have blocked this repo's own incident-capture command — the known-good corpus is not optional.

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
