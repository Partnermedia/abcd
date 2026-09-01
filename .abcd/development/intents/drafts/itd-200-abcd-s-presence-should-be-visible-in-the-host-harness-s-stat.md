---
id: itd-200
slug: abcd-s-presence-should-be-visible-in-the-host-harness-s-stat
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: [itd-20]
severity: minor
impact: additive
promoted_from: iss-168
origin: extracted-from-record
production_mode: dictated-and-formatted
---

# A managed repository shows it is managed, and whose answer the loop is waiting on, in the host's status line

## Press Release

> **A repository under abcd management now looks managed, and the status
> line says whose answer the agent loop is waiting on.** The line leads with
> one badge in three states: abcd is here and nobody is waiting; the loop is
> parked on the facilitator; the loop is parked on the product thinker. The
> agent sets the state when it stops for a verdict, and the human sets it by
> hand to say which hat they are wearing. Where the host offers no status
> surface, one line at the stop names whose answer is owed, and the bare
> `/abcd` board answers the same question on demand. An unmanaged repository
> shows nothing. The install step offers the line, says why it is worth
> having, and lets the user configure or decline it.

> "I could see the loop was parked on someone who was not me before I read a
> single line of output," said a facilitator who runs the agents for a small
> product team. "The badge was there before the directory and the branch, so
> a narrow window never hid it." A product thinker on the same team never
> sees the line at all: "I was told an answer was owed to me. That is the
> whole point, and the terminal was never going to be where I heard it."

## Why This Matters

The status line is the visible face of abcd. Today a managed repository is
indistinguishable from any other until somebody runs a command, and a stop
that the agent loop parks on the product thinker is invisible to the person
sitting in front of the repository. Under the roles decisions the agents run
unattended and stop only to obtain a verdict, and the product thinker answers
on a surface of their own, so the terminal is exactly where that parked stop
goes unnoticed. One leading badge closes both gaps with one slot.

The record already holds the design, settled in [iss-168](../../../work/issues/open/iss-168-abcd-s-presence-should-be-visible-in-the-host-harness-s-stat.md)
by a live demonstration on 2026-08-29: the marker leads the line because the
harness hands the status command no terminal width and the right edge is what
truncation cuts first; the badge is inverse video with its word inside so it
reads without colour; there is no animation because the line re-renders on
events, not on a clock; the presence badge is a setting and the role badges
are fixed; the default is gold on dark grey at a measured 6.94 to 1.

## Decisions (grilled 2026-09-01)

The maintainer resolved these at the planning interview; they are commitments,
not options.

- **Three states, one slot, two writers.** The badge is a mode abcd stores:
  managed and nobody waiting, waiting on the facilitator, waiting on the
  product thinker. The agent writes it when it needs an answer, according to
  whom it is addressing; the human writes it to say which hat they wear. Both
  the status line and the agent read the same state.
- **Managed repositories get abcd's line; the rest keep the user's own.**
  The defaults ship with abcd and move into the user-level home under
  `~/.abcd/` at install time. In an abcd-managed repository they override the
  harness's own status setting; in any other repository the harness's setting
  stands untouched. The user-level setting can switch abcd's line off
  entirely.
- **The install step asks.** `ahoy install` offers the status line, says why
  it is worth having, and takes the basic configuration of its elements as
  part of onboarding.
- **Where the host has no status surface**, one line at the stop names whose
  answer is owed, once, and the bare `/abcd` board shows the same state on
  demand. The board is the fallback, never a separate status command
  ([itd-20](../planned/itd-20-top-level-abcd-dispatcher.md) owns it).
- **First harness only.** This cut wires the harness abcd already ships
  lifecycle hooks for. Others follow one at a time in the roster order the
  record sets.
- **Word and inverse video carry the meaning.** Colour reinforces; it is never
  the only signal.
- **abcd owns the whole row in a managed repository.** Once installed, the
  line is abcd's own elements, not a prefix on the user's previous status
  command, in this order: the badge, the repository name, the branch, the
  model, context used as a percentage, the five-hour and seven-day usage
  percentages, and the record's counts of intents and issues. The badge is
  fixed and leads; every element after it is one the user can switch off at
  install time or later in the same user-level setting. Elements that come
  from the harness's status payload (model, context, usage) render only when
  the payload carries them, and an absent field drops the element rather than
  showing a placeholder. The branch comes from the repository; the counts
  come from the record, and what they count exactly (open issues; intents not
  yet shipped) is settled in the spec.
- **Out of scope, recorded for a later iteration:** a communication path that
  adapts to the role, where the product thinker answers through a web surface
  and the facilitator stays at the terminal. This intent only makes the parked
  stop visible; it does not carry the answer back.

## What's In Scope

- One canonical render of the line in core, consumed by the status surface
  and by the `/abcd` board, so no front door invents its own words.
- The stored mode with its two writers: a verb the agent calls when it stops
  for a verdict, and the same verb or a setting the human uses by hand.
- Install-time detection of the first harness, the wiring `ahoy install`
  writes, the offer and its explanation, and the basic element configuration.
- The user-level defaults under `~/.abcd/`, the managed-only override, and
  the off switch.
- The one-line notice at the stop where no status surface exists.
- Contrast checking of a configured presence badge against its own
  background, refusing a pair below the bar with the measured ratio.

## What's Out of Scope

- Guard health, peer-session counts, uncommitted ledger counts and any other
  candidate for the line. Each is its own decision once the badge has lived
  in real sessions.
- Update-available signalling: knowing the latest release needs the network,
  which implicit operations never touch.
- Any second harness.
- The role-adapted communication path named in the Decisions.

## Mechanism

We expect a visible mode to make both the agent and the human address the
right person, because the current hat is always on screen.

## Scope Conditions

- Holds only on a harness that renders a status line by invoking a command
  and displaying its stdout, which is the first harness in the roster.
- Holds only in an abcd-managed repository; elsewhere abcd writes nothing to
  the line and changes no setting.
- Holds where the badge's word and inverse video are legible; colour is
  reinforcement and the design does not depend on it.
- Holds under the width the host chooses: the line is designed for
  truncation, not layout, and only the leading badge is guaranteed to survive.
- Holds for the payload fields the first harness supplies today (model name,
  context percentage, five-hour and seven-day usage); a harness that stops
  supplying one drops that element rather than breaking the line. The
  maintainer's own 2026-08-29 demonstration wiring is the reference rendering
  for every element except the two record counts, which are new.

## Acceptance Criteria

- **Given** an abcd-managed repository and no parked stop, **when** the host
  renders its status line, **then** the line begins with the presence badge
  in its configured or default colours, and nothing else abcd owns precedes
  it.
- **Given** the agent loop stops to obtain a verdict from the product
  thinker, **when** it records the stop, **then** the next render shows the
  product thinker badge, and the `/abcd` board reports the same state.
- **Given** the agent loop stops to obtain a verdict from the facilitator,
  **when** it records the stop, **then** the next render shows the
  facilitator badge.
- **Given** a facilitator sets the mode by hand, **when** the line next
  renders, **then** it shows the state they set, and the agent reads the same
  state.
- **Given** a repository that is not abcd-managed, **when** the host renders
  its status line, **then** abcd contributes nothing and the harness's own
  setting is untouched.
- **Given** the user has switched abcd's line off in their user-level
  settings, **when** a managed repository renders its status line, **then**
  abcd contributes nothing.
- **Given** a harness with no status surface, **when** the loop parks a stop,
  **then** exactly one line names whose answer is owed, and the `/abcd`
  board shows the state afterwards.
- **Given** `ahoy install` on the first harness, **when** it reaches the
  status line step, **then** it offers the line, states why it is worth
  having, lets the user switch each element after the badge on or off, and
  writes no wiring if the user declines.
- **Given** the line is installed with every element on, **when** the host
  renders it with a payload carrying model, context and usage figures,
  **then** the row reads, in order, the badge, the repository name, the
  branch, the model, the context percentage, the five-hour and seven-day
  usage percentages, and the intent and issue counts, each separated the same
  way.
- **Given** the payload lacks a field an element needs, **when** the line
  renders, **then** that element is absent and no placeholder takes its
  place.
- **Given** a configured presence badge whose foreground and background fall
  below the contrast bar, **when** the setting is read, **then** it is
  refused with the measured ratio and the default renders instead.
- **Given** a terminal narrower than the line, **when** the host truncates
  it, **then** the badge is what survives.

## Open Questions

1. The name and shape of the verb the agent calls to record a stop, and
   whether the human's hand-set mode uses the same verb or a setting. Spec
   detail.
2. Where the stored mode lives so that both the status command and the agent
   read it without a harness variable: repository-local state or the
   user-level home. Spec detail, constrained by the managed-only rule.
3. The roles decisions this intent refines live on the design branch under
   ids that collide with main's; they take their final ids at merge, and this
   record's links are re-pointed then.

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._

## Grounds

- pursued: the status bar is the visible face of abcd, so a managed repository should look managed; we expect facilitators to keep the line on after living with it, and if they switch it off in their own settings that shows this was the wrong call
