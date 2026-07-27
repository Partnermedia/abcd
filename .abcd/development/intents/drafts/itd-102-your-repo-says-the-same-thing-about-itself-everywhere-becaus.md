---
id: itd-102
slug: your-repo-says-the-same-thing-about-itself-everywhere-becaus
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
impact: additive
---

# **Your repo says the same thing about itself everywhere — because abcd asked you once and holds every surface to it.** A project's positioning fragments silently: the README strapline, the package or plugin manifest description, and the agent conventions file are each edited at different moments, and soon three surfaces tell three stories about what the project is. When abcd prepares a repo it interviews the user for the project's identity — title, tagline, and a short elevator pitch — and records the answers in one canonical home in the development record. Every surface rendering derives from that home, and a deterministic positioning check compares the surfaces against it on every audit: a mismatch is highlighted with the exact drifted line, and the fix is either re-render from the record or a deliberate, recorded identity change — never a silent extra variant, and never a silent rewrite by abcd. "I changed my tagline once, in one place, and abcd chased it everywhere else," said Alice, a solo founder. "When a stray edit crept into the README, the next audit pointed at the exact line instead of letting the drift settle in."

## Press Release

> _Seeded from a quoted-text intent capture. Expand into the full press-release narrative before planning._

## Why This Matters

**Your repo says the same thing about itself everywhere — because abcd asked you once and holds every surface to it.** A project's positioning fragments silently: the README strapline, the package or plugin manifest description, and the agent conventions file are each edited at different moments, and soon three surfaces tell three stories about what the project is. When abcd prepares a repo it interviews the user for the project's identity — title, tagline, and a short elevator pitch — and records the answers in one canonical home in the development record. Every surface rendering derives from that home, and a deterministic positioning check compares the surfaces against it on every audit: a mismatch is highlighted with the exact drifted line, and the fix is either re-render from the record or a deliberate, recorded identity change — never a silent extra variant, and never a silent rewrite by abcd. "I changed my tagline once, in one place, and abcd chased it everywhere else," said Alice, a solo founder. "When a stray edit crept into the README, the next audit pointed at the exact line instead of letting the drift settle in."

## Acceptance Criteria

- Given repo onboarding runs (the install or prepare flow), when the identity interview completes, then title, tagline, and pitch land in a parseable markdown identity block (fixed bullet shape) whose location the repo's abcd config records — markdown stays the single source of truth.
- Given a rendered surface (README strapline, manifest description, conventions-file opening) diverges from the identity block, when the positioning check runs, then a warn-tier finding names the exact drifted line; per-repo config may upgrade the finding to blocker.
- Given drift is found, then abcd never rewrites a surface autonomously — re-rendering from the block is always a proposed diff a human adopts.
- Given abcd-cli itself, then the check points at the brief product chapter's existing Identity section unchanged, and iss-143 (the three-variant drift) is the acceptance corpus the check must catch.

## Open Questions

- Interview wording and which surfaces are registered for checking by default beyond the canonical three.
- Whether the pitch is required or optional at onboarding (title and tagline are required).

## Grill Settlements (2026-07-27)

- Identity home is a parseable markdown block, not a structured JSON file — consistent with markdown-as-single-source-of-truth; the config records only where the block lives.
- The check is warn-tier by default (highlight, never gate) and per-repo upgradeable; autonomous rewriting is permanently out.

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
