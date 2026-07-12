---
id: itd-82
slug: review-bar-fires-itself
spec_id: null
kind: null
suggested_kind: standalone
reclassification_history: []
related_adrs: []
prd_path: null
severity: major
builds_on: [itd-3, itd-81]
---

# The Review Bar Fires By Itself, In Every Repo abcd Manages

## Press Release

> **abcd's reviewer agents now invoke themselves at the moment they are needed.** A
> `REVIEW` domain in the bundled rules loader recalls on the language of shipping —
> *review*, *diff*, *PR*, *present*, *ship*, *release* — and injects the standing
> rule just-in-time: a non-trivial diff gets `abcd:ruthless-reviewer` before it is
> presented; a change touching a trust boundary gets `abcd:security-reviewer` before
> it lands; a release gets `abcd:docs-currency-reviewer`. Nothing to remember,
> nothing to configure, no text copied into your repo. Install abcd and the review
> bar is simply on.
>
> "I kept forgetting to run the security reviewer on exactly the changes that needed
> it most — the ones I was in a hurry to ship," said Carol, a solo maintainer. "Now
> it just happens. The one time it blocked me, it was right."

## Why This Matters

[itd-81](../disciplines/itd-81-judge-calibration.md) shipped four reviewer agents
into the plugin, and they are *reachable* — `abcd:ruthless-reviewer` and friends
resolve in every repo with abcd enabled. But nothing **calls** them. They fire only
when the host model happens to choose to delegate, or when a human remembers to ask.
Under abcd's own **wired-or-it-isn't-done** rule, that is not done: an agent nothing
invokes is scaffolding with a good description.

The rule that makes them fire currently lives in the maintainer's **personal
machine-level instructions** — a file no abcd user has, in a directory no abcd user
has. Every other repo abcd manages gets the agents and none of the discipline. The
capability is shipped; the reason to use it is not.

abcd already owns the right mechanism. The modular rules loader ([itd-3](../shipped/itd-3-modular-rules-loader.md))
ships eight bundled domains with keyword recall, injects matched rules just-in-time
on `UserPromptSubmit` via a wired hook, and is per-repo overridable through
`.abcd/rules.json`. `COMMITTING` already carries standing rules of exactly this shape
("Substantive work goes on a branch + PR", "never force-push"). A review bar is the
same kind of rule, and it belongs in the same place — not copied into every repo's
`AGENTS.md`, which would be a fourth copy of text that has a canonical home.

## What's In Scope

- A **`REVIEW` domain** in the bundled `rules.json`, `state: active`, recalling on
  the vocabulary of shipping (`review`, `diff`, `pull request`, `present`, `ship`,
  `release`, `security`, `merge`) with the standing rules that name the agents:
  ruthless before presenting a non-trivial diff; security on any trust boundary
  (auth, secrets, input parsing, network, subprocess); docs-currency before a
  release tag. Whether a security BLOCK is a hard stop or advisory is an open
  question below — it turns on whether the agent has been measured.
- Rules **point at** the agents and the canonical discipline
  ([itd-81](../disciplines/itd-81-judge-calibration.md)); they never restate its
  content. This follows the `OPINIONS` precedent, whose rules point at
  `.abcd/development/principles/` rather than copying them.
- A **host-agnostic fallback**: a one-line pointer in the marked conventions section
  of `AGENTS.md` that `prepare-this-repo` already owns, so a harness that does not
  run the hook still learns the agents exist. A **pointer, not a restatement** — if
  the line grows into a copy of the rules, the design has failed.
- Per-repo override: a repo that wants a different bar (or none) turns the domain off
  or replaces its rules through `.abcd/rules.json`, using the loader's existing
  per-field override and sticky kill switch.

## What's Out of Scope

- **Enforcement.** The domain *injects a rule*; it does not gate a commit. A hook
  that hard-blocks a commit until a reviewer has run is a separate, heavier
  capability, and it would need [itd-81](../disciplines/itd-81-judge-calibration.md)'s
  corpus scores first — enforcing an unmeasured judge is exactly the failure that
  discipline exists to prevent.
- **Auto-fixing what a reviewer finds.** Findings go to the human.
- **Calibrating the reviewers.** That is itd-81's corpus work, and it is a hard
  prerequisite for trusting anything this intent causes to fire.
- **Changing the agents' prompts.** They ship as-is at `0.1.0`.

## Acceptance Criteria

> _BDD format, per the [itd-1 discipline](../disciplines/itd-1-acceptance-gates.md)._

- **Given** a repo with the abcd plugin enabled, **when** `abcd rules REVIEW` runs,
  **then** it renders the review domain's rules, each naming the agent it invokes.
- **Given** a session in an abcd-managed repo, **when** the operator's prompt matches
  the review domain's recall vocabulary, **then** the prompt-router hook injects the
  review rules exactly once per session signature (the loader's existing dedup).
- **Given** a repo whose `.abcd/rules.json` disables or overrides the `REVIEW`
  domain, **when** the loader runs, **then** the repo's choice wins and the bundled
  default does not reappear.
- **Given** an abcd-managed repo prepared by `prepare-this-repo`, **when** its
  `AGENTS.md` conventions section is read, **then** it points at the reviewer agents
  in one line and does **not** restate the rules the loader injects.
- **Given** a harness that does not run the prompt-router hook, **when** an agent
  reads `AGENTS.md`, **then** it can still discover that the reviewer agents exist
  and what each is for.
- **Given** this intent ships, **when** the wired-or-it-isn't-done check runs, **then**
  each of `abcd:ruthless-reviewer`, `abcd:security-reviewer`, and
  `abcd:docs-currency-reviewer` is demonstrably invoked from a production path, not
  merely present in `agents/`.

## Open Questions

- **One `REVIEW` domain, or rules folded into the existing `COMMITTING` domain?**
  `COMMITTING` already recalls on `pr`/`branch`/`merge` and already carries
  branch-and-PR rules, so a reviewer rule is arguably at home there — but review also
  fires on `release` and `security`, which `COMMITTING` does not recall. Leaning
  toward a separate domain; confirm at spec time by looking at real recall overlap.
- **Does the security rule's BLOCK belong in a rule at all,** given the agent is
  unmeasured (`0.1.0`)? A rule that says "a BLOCK stops the change" grants authority
  to a judge whose false-positive rate is unknown. Options: ship the rule as advisory
  until itd-81 produces a TNR figure, or ship it as a stop and accept the noise.
  **Recommend advisory until measured** — but this is a maintainer call.
- **Host-agnosticism.** The loader's injection path is a Claude Code hook. Naming the
  fallback "`AGENTS.md`" assumes other harnesses read that file. Is that assumption
  good enough, or does the fallback need its own surface?

## Dependencies

- **Hard prerequisite:** [itd-3](../shipped/itd-3-modular-rules-loader.md) — the rules
  loader, its domains, and the prompt-router hook. Phases 1 + 3 have shipped; this
  intent adds a domain to what exists and builds no new machinery.
- **Hard prerequisite:** [itd-81](../disciplines/itd-81-judge-calibration.md) — the
  agents this intent wires, and the discipline that says they must be measured before
  they are trusted. Wiring an unmeasured judge into a hard gate is out of scope for
  exactly that reason.
- **Coordinated with:** `prepare-this-repo` — it owns the marked `AGENTS.md`
  conventions section the fallback pointer lands in.

## Audit Notes

_Empty. Populated by `intent-fidelity-reviewer` when this intent ships._

## References

- [`itd-81-judge-calibration.md`](../disciplines/itd-81-judge-calibration.md) — the
  discipline that shipped the agents and governs when they may be trusted.
- [`itd-3-modular-rules-loader.md`](../shipped/itd-3-modular-rules-loader.md) — the
  mechanism this intent extends.
- [`itd-5-prompt-quality-additions.md`](../disciplines/itd-5-prompt-quality-additions.md)
  — the prompt-quality floor every agent named here inherits.
