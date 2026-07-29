---
id: itd-105
slug: autonomous-routines-assemble-from-one-versioned-template
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
impact: additive
---

# Autonomous Routines Assemble From One Versioned Template

## Press Release

> **abcd now assembles the prompts that drive autonomous maintenance loops.** A
> repository's recurring bug-hunt or plan-drain routine is no longer a
> hand-written wall of text: `abcd routine render` detects the repo's toolchain,
> gate commands, and working-tier layout, fills a versioned archetype template,
> and emits the finished routine prompt plus a host-ready schedule manifest. The
> safety invariants every loop needs — state comments trusted only from the
> repo's own maintainers, resumption of a pushed branch that never got its pull
> request, a bounded wait on CI, one round per scheduled run — are baked into
> the template, not remembered per repo. When the template improves, the whole
> fleet re-renders in one pass.
>
> "I had six repos hunting bugs on a schedule, and every prompt was a slightly
> different vintage," said Alice, a maintainer. "One of them was told to commit
> into a directory the repo ignores. Now they all render from the same template,
> and when I fix the template I have fixed the fleet."

## Why This Matters

A fleet audit of eight live routines found three prompt generations with
divergent safety properties: different stop sentinels, merge gates with and
without a bounded CI wait, single- and multi-round pacing, and one prompt
instructing every round to commit a decision log into a path its repository
gitignores. None of these differences was a decision — each was drift from
hand-assembly. Re-hardening the fleet meant editing every routine individually,
a full session of expert attention for changes that were mechanically identical.

The assembly is deterministic work sitting outside the binary. Detecting a
toolchain from its manifest, finding where the tracked decision log lives,
choosing gate commands, slotting a hunt focus into a template — all of it is
exactly the kind of behaviour the transport-agnostic core exists to own, and
today it lives in an unversioned, untestable prompt-assembly skill on one
person's machine. Under **wired or it isn't done**, the discipline exists but
nothing in abcd carries it.

The split follows the house boundary: the binary owns detection, the template,
and rendering; the plugin page owns the interview (focus, cadence, merge
authorisation) and hands the rendered artefact to whatever the host uses to
schedule it. The binary never contacts a scheduling service.

## What's In Scope

- A versioned routine template with archetypes — bug hunt (security /
  correctness / docs / all focuses) and plan drain — carrying the safety
  invariants: maintainer-only trust for loop-state comments, pushed-branch-
  without-PR resumption, bounded CI wait, one-round-per-run default with an
  explicit capped multi-round variant, refuter count, stop conditions, and
  public-artifact hygiene (iss-160): session URLs and harness attribution
  footers are banned from all public text, and because the harness can append
  them outside the model's own words, every created pull request, issue, and
  comment is re-read as stored and stripped by edit before the round continues.
- `abcd routine render`: repo detection (toolchain and gate commands from the
  manifest; tracked-versus-ignored working tiers, so a rendered prompt can
  never instruct a commit into an ignored path; platform caveats for
  cross-compiled targets) plus deterministic output — same repo, same template
  version, byte-identical render.
- A neutral schedule manifest alongside the prompt (name, cadence, repo,
  model routing), for the host front door to translate.
- A plugin page front door that interviews, renders through the binary, and
  hands off to the host's scheduler.
- An optional **release on completion** template switch, default off, offered
  explicitly by the interview: when the loop reaches its drained stop (the
  configured run of nitpick-only rounds) and substantive fixes have merged
  since the last release, the closing round initiates the repository's release
  flow instead of only posting the stop sentinel. The routine never bypasses a
  release gate — in an abcd-managed repo that means the changelog-driven
  release gate runs exactly as it would for a human-initiated cut.

## Acceptance Criteria

- Given a Go repository with a `Makefile` preflight target, When Alice runs
  `abcd routine render` with the all-dimensions hunt archetype, Then the
  emitted prompt contains that repo's actual gate commands and every template
  safety invariant, and rendering twice produces byte-identical output.
- Given a repository whose candidate decision-log path is gitignored, When a
  routine is rendered for it, Then the prompt directs round records to the
  loop-state issue only and contains no instruction to commit into an ignored
  path.
- Given a fleet rendered at template version N, When Bob re-renders it at
  version N+1, Then each routine's prompt differs from its predecessor only by
  the template change.
- Given the plugin page interview, When Carol completes it, Then the rendered
  prompt and schedule manifest are produced by the binary alone, with no
  network call made by the binary.
- Given the interview's defaults accepted, When a routine is rendered, Then it
  contains no release step; Given release on completion switched on, When the
  loop's final stop follows rounds that merged substantive fixes, Then the
  rendered prompt directs the closing round to initiate the repository's
  release flow, subject to every release gate the repo declares.
- Given any archetype, When a routine is rendered, Then the prompt bans
  session URLs and harness attribution footers from every public artifact and
  directs a post-create re-read-and-strip of each pull request, issue, and
  comment the loop creates.

## Open Questions

- Verb name and family: `abcd routine` as its own verb, or a sub-verb of an
  existing family?
- Does the binary emit only the neutral manifest, or also host-specific
  schedule configuration behind an adapter seam?
- Where fleet membership is recorded (which repos carry which routines at
  which template version) — per-repo record, user home, or nowhere?
- Whether the plan-drain archetype's work-queue contract (plan file wins,
  strike rules) is part of this intent or its own.
- What "initiate the release flow" means per repo — invoke the release
  machinery directly, or open the release cut as a final gated pull request —
  and how the closing round interacts with receipt-gated releases that need
  host-run semantic reviews.

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
