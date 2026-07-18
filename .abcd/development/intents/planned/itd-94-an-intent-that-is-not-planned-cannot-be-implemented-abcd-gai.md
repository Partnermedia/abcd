---
id: itd-94
slug: an-intent-that-is-not-planned-cannot-be-implemented-abcd-gai
spec_id: spc-9
kind: standalone
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
---

# An intent that is not planned cannot be implemented: abcd gains a machine-checkable implement-readiness gate that refuses with a remedy and offers the planning interview

## Press Release

You ask your agent to implement itd-93. Instead of improvising against a draft
whose acceptance criteria nobody confirmed, it runs `abcd intent ready itd-93`,
reads the NOT READY verdict, and tells you plainly: "itd-93 is not specced, so
it cannot be implemented yet — want to plan it together?" The interview that
follows is yours: you confirm the press release, resolve the open questions,
and accept, edit, or strike every acceptance criterion. Only then does the
agent run `abcd intent plan` — your sign-off act — and the minted spec becomes
the design record it builds against. Autonomous drain runs get the same gate
for free: exit 1 is a journaled SKIP, never a guess.

## Why This Matters

Today "is this intent ready to implement" is re-derived from prose by every
run (the drain-run skip filter) and by every human conversation. iss-83 already
records the failure mode: a missing plan precondition invites the agent to
synthesize one, re-introducing the guess-the-gate risk that fail-closed design
exists to prevent. Facilitator-seeded acceptance criteria look real but carry
no maintainer sign-off; nothing machine-checkable distinguishes them. A single
read-only verb with a strict exit-code contract closes both holes: humans get
a refusal with a remedy and an offered interview, runs get a deterministic
step-0 gate.

## Acceptance Criteria

- Given an intent in `drafts/` (regardless of how good its seeded acceptance
  criteria look), when `abcd intent ready <itd-N>` runs, then it reports NOT
  READY with a `bucket` failure whose remedy names the exact next step
  (`abcd intent plan <itd-N>` after maintainer confirmation, or the planning
  interview when acceptance criteria are missing) and the process exits 1.
- Given a planned intent whose linked spec body still carries the minted
  `_Draft:` stub, when `abcd intent ready` runs, then the `spec_body` check
  fails with a write-the-spec-body remedy and the process exits 1.
- Given a planned intent with a linked spec whose reciprocal `intent:` matches
  and whose body is written, when `abcd intent ready` runs, then all four
  checks (`bucket`, `acceptance_criteria`, `spec_link`, `spec_body`) pass and
  the process exits 0.
- Given a structural fault (malformed id, unknown intent, unreadable record),
  when `abcd intent ready` runs, then it exits 2 with a one-line diagnostic —
  distinguishable by exit code from a NOT READY verdict (exit 1).
- Given the plugin surface `/abcd:intent`, when a user asks the host to
  implement an intent for which `ready` exits 1, then the surface instructs
  the host to refuse, present each failing check's detail and remedy, and
  offer the planning interview — and forbids authoring acceptance criteria or
  running `abcd intent plan` without the human's explicit in-session sign-off.
- Given the run protocol, when an unattended run picks an intent-backed item,
  then its step 0 is `abcd intent ready <itd-N>` and a nonzero exit is a
  journaled SKIP, never an improvised plan.

## Open Questions

_None recorded yet — the scaffold surface questions that block itd-93 do not
apply here; this gate is self-contained (read-only reporter + docs)._

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
