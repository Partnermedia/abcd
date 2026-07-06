---
id: rfc-1
slug: pirate-mode-yolo-for-power-users
status: open
discussion_opened: 2026-05-04
discussion_closes: TBD
spawned_from: null
spawned_intents: []
related_intents: [itd-26]
authors: [project]
---

# RFC-1: Pirate Mode — Should abcd Have a YOLO Surface for Power Users?

## The Question

Should abcd have a global "pirate mode" — a flag in `.abcd/config.json` that suppresses transparent confirmation prompts across all commands, signalling "I know what I'm doing, get out of my way"?

This RFC is *not* asking whether power users should have *some* way to skip confirms. They should — that's table stakes. The question is whether the right shape is a **global mode** (one switch that quiets the whole tool) versus **per-command flags** (each command has its own `--yes` / `--non-interactive` opt-out, scoped to the operation, logged with intent).

## Why We're Asking

The impulse for pirate mode is real:

- abcd's universal-pattern § 6.1 — "transparent prompts, no silent defaults" — is correct in spirit but generates real friction for routine ops once you've internalised the choices. The tenth time you confirm "yes, install the PATH symlink", the prompt feels like a tax, not a safety net.
- Power users in mature CLIs reach for global expert modes (`set -e`, `git config --global advice.detachedHead false`, IDE "expert mode" panels). The pattern exists because the demand is real.
- abcd's metaphor centre is maritime; "pirate mode" lands cohesively as the rebellious counterpart to dignified ahoy/disembark/embark/launch.
- Naming a thing `pirate mode` is *consent theatre* — the cuteness encodes "yes I'm wearing the eyepatch on purpose, don't blame the tool when I sail into rocks".

But it's also full of pitfalls:

- abcd's centre of gravity is "show your work, log every action". Adding a global silencer fights the architecture.
- Playful framing makes destructive things feel less destructive. `--dangerously-skip-permissions` works because the *name* increases friction proportionate to the risk; pirate mode does the opposite.
- A single mode flag flattens at least three distinct populations (researchers running automation, power users on routine ops, YOLO operators) into the loosest one.
- Cute names tend to get engaged casually after the third use. The user who'd benefit most from a YOLO mode is the user most likely to engage it without thinking.

This RFC asks the community to weigh both sides, sharpen the question, and tell us whether we're seeing the design clearly.

## What We've Already Decided

These constraints are **not up for discussion** — they are abcd values that pirate mode (in any form) must respect:

1. **Licence checks are non-circumventable.** Per itd-26's design lock, no flag, no mode, no `--override` lets `/abcd:loot` skip the licence-acceptable-list check. The only pathway past a refusal is to amend the policy explicitly via `loot policy add`. Pirate mode does not change this.
2. **All actions are logged.** Per `Autonomous/.claude/CLAUDE.md` and brief § 6, every command run writes to `.abcd/logbook/`. Pirate mode does not suppress logging.
3. **Sanitisation pre-flight on `/abcd:launch` is non-circumventable.** Per brief § 10, the launch payload is always scrubbed for PII/secrets before publication. Pirate mode does not bypass.
4. **Hash-chain / audit trail integrity** (per itd-16, when it ships) is non-circumventable.

These are the non-negotiables. The discussion is about everything *else* — the routine confirms, the "are you sure?" prompts, the per-step gates that exist for safety rather than correctness.

## Considered Alternatives

### Option A — Global "pirate mode" flag

`.abcd/config.json` → `mode: pirate`. Suppresses all confirms across all commands except those touching the non-circumventable list above. Loud banner in every output. Per-session opt-out via `--no-pirate`.

- **For:** simple, memorable, encodes consent via the playful name itself.
- **Against:** flattens three populations into one switch; "fun" name lowers perceived stakes; one-flag-changes-everything is the highest blast-radius design.

### Option B — Per-command `--yes` / `--non-interactive` flags

Each command grows its own opt-out (`/abcd:ahoy --yes`, `/abcd:disembark --non-interactive`). Logged with intent ("user passed --yes on YYYY-MM-DD to /abcd:command"). No global mode.

- **For:** explicit, scriptable, audit-friendly, per-operation scope means a careless user can't accidentally suppress confirms on operations they actually wanted to think about.
- **Against:** more typing; doesn't address the "I've already answered this confirm 50 times" fatigue case as elegantly as a global mode.

### Option C — Config-level "I've answered this" memory

`.abcd/config.json` records confirms the user has answered ("yes, install PATH symlink, don't ask again on this machine"). Subsequent runs skip the prompt. Brief § 6.1's "transparent re-confirm rule" already implies this; we'd just need to make it explicit.

- **For:** addresses the actual fatigue case (re-asking known-answered questions) without a global flag; aligns with abcd's existing "transparent prompts" pattern; per-confirm scope.
- **Against:** doesn't help with confirms the user *hasn't* answered yet; doesn't give power users a single "shut up and let me work" gesture.

### Option D — `/abcd:ahoy --autonomous` for unattended use

A specific mode for CI / cron / batch contexts where no human is present. Refuses to run any operation requiring a confirm not pre-answered in config; fails loudly rather than guessing. This is the *real* unattended-use case dressed correctly.

- **For:** matches a real population (CI/automation users) with a real need (no-human-present); fails loudly on novel ops, which is the right behaviour in unattended contexts; doesn't pretend to be expert mode.
- **Against:** doesn't address the interactive-power-user case at all; needs separate work for that.

### Option E — Do nothing

Keep the universal-pattern § 6.1 transparent-prompts rule. Power users can grumble. abcd's safety-by-default reputation matters more than the friction.

- **For:** maximally safe; preserves architectural coherence; community-coding-conservative.
- **Against:** loses real power-user empathy; the demand will leak out as users patch around the prompts in their own dotfiles, which is *worse* than a designed escape hatch.

### Hybrid: B + C + D, no A

The combination of per-command `--yes` flags (B), config-level "already-answered" memory (C), and an explicit `--autonomous` mode for unattended use (D) covers all three populations without needing a global pirate mode. This is the design favoured in the conversation that produced this RFC.

## What We're Hoping to Learn

The feature question ("should pirate mode ship?") is downstream of a values-question:

> **How does abcd balance "show your work" (transparency, logging, no surprises) against "respect expertise" (don't make me re-confirm what I've already learned)?**

We don't know. The instinct is "transparency wins by default, with carefully-placed escape hatches", which is what Hybrid B+C+D codifies. But we'd be naive to treat that as settled.

What we'd find most useful:

- **Real friction stories.** "Confirming X every time genuinely costs me Y" is more useful than "in principle, there should be a mode".
- **Counter-examples from other tools.** Where have global expert modes worked well? Where have per-command opt-outs failed? `git --force` patterns, `npm --force` regrets, `docker --privileged` lessons — all useful.
- **Disagreement with the non-circumventable list.** Are there other items that should be added? Removed? Are we drawing the line in the right place between "always check" and "pirate-mode-can-skip"?
- **Tone read on the "pirate" framing.** Does playfulness here help (memorable consent theatre) or hurt (lowers perceived stakes)? Different audiences will see this differently; we want both reads.

## How to Engage

This RFC is intentionally the *first* community-facing artefact for abcd. The way it gets discussed sets cultural priors for everything that follows. Useful contributions:

- **Comment on the GitHub issue/discussion** linked from this RFC at launch time. (Link TBD — this RFC currently lives in private dev; will be opened for public discussion when the public `abcd/` repo launches.)
- **Reference your own experience** — a story about when expert mode helped or hurt is worth more than abstract argumentation.
- **Suggest constraints we missed.** The "What We've Already Decided" section is where we want pushback most: did we mark something non-negotiable that should actually be open?
- **Weigh in on tone.** If pirate mode (the *name*) feels wrong even if the *design* is sound, say so — naming is design.

We're not running this as a vote. The maintainer position will weigh comments based on the reasoning behind them, not the count. We'll respond to substantive comments with substantive replies (this is a deliberate cultural-prior choice — we want the community to know discussions are taken seriously).

## Resolution

_Empty. Populated when status moves to resolved-yes / resolved-no / resolved-modified / withdrawn. The Resolution section will summarise the discussion's main threads, name the maintainer's reasoning for the chosen outcome, and link any spawned intents (frontmatter `spawned_intents` field updated reciprocally)._
