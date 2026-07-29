# "explainability" — follow-up plan and run queue (2026-07-29)

**Status:** queued behind the v0.5.0 cycle
([`2026-07-29-v0.5.0-security-and-consistency.md`](2026-07-29-v0.5.0-security-and-consistency.md));
consumed by the same generic protocol
([`2026-07-12-abcd-run-protocol.md`](2026-07-12-abcd-run-protocol.md)) once
that queue drains. The version is **derived, never declared**; with these
items additive on top of a shipped v0.5.0, derivation predicts v0.6.0 — a
prediction, not an override.

**Release framing (maintainer, 2026-07-29):** the tool explains itself.
Every choice abcd asks, every result it reports, and its very presence in a
session must be understandable to the persona roster's non-implementers —
the product thinker and the facilitator are the bar, not the staff
engineer. The 2026-07-29 manual marketplace-install test showed the current
distance: prompt options nobody can choose between (an unexplained
`host-delegated | native | cli | api | mcp`), completion summaries in
implementer jargon ("marker blocks", "identity gate", raw env-var names),
a `--yes` flag whose behaviour contradicts its help text, and no way to
tell from inside a session that a repo is abcd-managed at all. The root
cause is structural, not cosmetic: core emits bare enum values and result
data, so every front door improvises its own words — and improvised words
are wrong words. The fix puts canonical plain language in core, rendered
identically by every surface; the status line is a feature, but one whose
job is precisely to alert the user that the repo is abcd-managed.

## Run contract

Same as the v0.5.0 cycle's (gate, budget, trailer, strike limit, one PR
per item, ledger item moved open → resolved in the fixing PR), with two
adjustments:

- `reviewers:` correctness `abcd:ruthless-reviewer` always; on every diff
  touching user-facing strings, a **persona-lens review**: the reviewer
  reads each prompt choice, summary, and notice against the persona
  registry's roles (`.abcd/development/personas.json` — product thinker,
  facilitator) and must name which persona would fail to understand any
  string it blocks on. A BLOCK leaves the PR open for the human.
- **Merge policy is not inherited**: the v0.5.0 auto-merge authorisation
  was scoped to that cycle. The maintainer re-authorises (or withholds)
  auto-merge when this cycle starts.

## Workstream A — the explanation layer (core)

1. **iss-163** — canonical per-choice help text lives in core. Every
   prompted choice carries plain-language data — what the option is, what
   it requires (credentials, network, cost), and its consequences — beside
   the enum value, surfaced identically by every front door (interactive
   CLI, plugin surface, `--help`). The foundation item: A2 and B4 build on
   the same structure. Includes the "what is an oracle" definition the
   backend prompt currently assumes.
2. **iss-164** (blocked by iss-163) — persona-readable result summaries:
   every reported install/upgrade item carries a "what this is / why it
   matters / what, if anything, you should do" framing in core, so the
   completion summary reads for the product thinker and the facilitator,
   not only for abcd's implementers. No raw env-var names or internal
   tier vocabulary without an in-place explanation.
3. **itd-63 `setup-wizard-explains-installs`** — the intent frame A1/A2
   deliver into ("told what is being installed and why, not just asked to
   run a command"). **Not implementation-ready**: planned but `spec_id`
   null (verified 2026-07-29). The first milestone of this item is the
   lifecycle, not code — grill the open questions, `abcd intent plan`
   is already done, so mint/write the spec and pass `abcd intent ready
   itd-63`; only then implement.

## Workstream B — prompts an agent (or a human) can actually answer

4. **iss-167** — every prompt gains a non-TTY answer channel (per-category
   flags, stdin when not a TTY, or a structured answers input). Found
   because a host agent could not drive `ahoy install` at all: piped and
   pseudo-TTY answers arrive as declines. Same prompt seam as A1 — land
   with or after it, never in parallel with it.
5. **iss-166** — `--yes` semantics match its words: either it covers the
   optional identity-pin category too, or its help text and the completion
   summary state plainly that optional categories are excluded and name
   the command that applies them. Silent divergence is the bug; either
   honest behaviour is acceptable.
6. **iss-171** — the PATH-install detector and installer stop assuming one
   blessed layout. The issue body records the full no-sudo redesign
   (PATH-scan + `EvalSymlinks` + classify dev-shim/owned/foreign;
   `~/.local/bin` default; system dirs only behind an explicit `--bin-dir`
   failing loudly when unwritable; refuse dangling symlinks; "not on PATH"
   as its own loud gap with the printed one-line fix). It sits in this
   cycle because its user-visible deliverable is honest, actionable
   install reporting — the current behaviour reports a false
   `symlink.missing` while running as `abcd` from PATH, and offers a fix
   that would dangle. This is the one item in the cycle with sanctioned
   behaviour change beyond wording; its filesystem-write diff gets the
   security-style review in addition to the persona lens. Builds on
   v0.5.0's iss-170 (`EvalSymlinks` in the resolver — same seam, land
   after it).

## Workstream C — presence and orientation

7. **iss-168** — the harness status line shows the repo is abcd-managed
   (and, if cheap, guard health). A host-adapter feature under the
   basics-built-in / SOTA-delegated stance recorded in iss-165: the
   adapter is host-specific by nature, while committed user-facing prose
   stays host-agnostic per the docs-lint rules — the adapter is named only
   where the attribution conventions allow.
8. **itd-20 `top-level-abcd-dispatcher`** — "`/abcd` tells you where you
   are": the command-surface complement to C7. **Not
   implementation-ready**: planned but `spec_id` null (verified
   2026-07-29). Same lifecycle-first milestone as A3. Minor severity;
   last in the queue, dropped first if the cycle needs trimming.

## Ordering and collisions

- A1 (iss-163) lands before A2 (iss-164) and B4 (iss-167): all three
  touch the prompt/explanation seam, and A1 defines the structure the
  other two consume.
- A3 (itd-63) and C8 (itd-20) each begin with the intent lifecycle
  (spec written, `intent ready` green) — implementing an unready intent
  is a STOP, not a shortcut.
- B5 (iss-166) is independent; safe in parallel with anything.
- C7's rendering consumes whatever A2 standardises for guard-health
  wording — prefer landing after A2, not required.
- The v0.5.0 cycle's D19 (`plugin-surface-parity`, iss-161's namespace
  flattening) renames every registered command; if this cycle starts
  before that item lands, any user-facing string naming `/abcd:*`
  commands must be written against the flattened names, not the current
  double-namespaced ones.

## STOP conditions (this cycle)

1. **Explanations live in core as data.** A front door hard-coding its
   own user-facing copy for a prompt choice or result item is a STOP —
   that is the root cause this cycle exists to remove
   (transport-agnostic-core boundary).
2. **Wording changes that alter behaviour.** Any diff that changes what a
   flag or prompt *does* beyond iss-166's two sanctioned resolutions and
   iss-171's recorded redesign is a STOP.
3. **Persona-lens BLOCK** stops the change (standing rule, mirrors the
   security-review BLOCK).
4. **Scope creep into the grill family.** iss-165 (grill delegates to an
   external SOTA interviewing skill) belongs with itd-27/itd-42 in a
   grill-focused cycle; touching it here is a STOP.
5. **Missing or ambiguous record** — an issue body or intent that does
   not support the work as scoped: fail closed, never synthesise.

## Explicitly in (maintainer, 2026-07-29)

iss-163, iss-164, iss-166, iss-167, iss-168, iss-171, plus the two
planned intents itd-63 and itd-20 as their frames. The maintainer named
iss-163/164/166/168 explicitly; iss-167 rides the same prompt seam as
iss-163/166 (one plumbing change, three issues); iss-171 (the bin-target
redesign, added 2026-07-29 while this plan was in review) delivers the
honest install reporting the cycle is about and is the one sanctioned
behaviour change; itd-63/itd-20 are the standing planned intents this
cycle realises.

## Explicitly out

iss-165 (grill delegation — grill cycle, with itd-27/itd-42); iss-90
(docs-lint's plugin-root dependency — functional fix, partially mooted by
v0.5.0's iss-170 resolver repair; re-triage after that cycle ships);
everything else in the ledger not named above.
