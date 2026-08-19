---
id: itd-92
slug: abcd-verifies-branch-protection-on-managed-repos-and-gates-t
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
---

# abcd verifies the collaborator fence on managed repos — a capability ladder, not a checklist

## Press Release

Alice keeps her project in a plain local git repository — no forge, no
organisation, no CI. Bob pushes his to a personal GitHub account. Carol runs
hers in an organisation with outside collaborators. All three run the same
`abcd ahoy` doctor, and none of them is told to become someone else first.

The doctor probes what each environment can actually hold — remote or none,
personal or organisation, public or private — and renders one verdict per
fence piece in a closed vocabulary: **applied**, **available but off** (with
the enabling lever), **rule without executor** (specified but nothing
enforces it), **unverifiable** (the probe could not answer — never a green),
or **impossible here** (with the one-line reason and what the user forgoes,
in plain words: "anyone with push can publish; there is no rung between
stranger and publisher"). Alice's Tier 0 — hooks, the private banlist, the
guard, preflight — is a supported state with an honest protection statement,
not an error. Bob sees which levers his personal account offers and which
need an organisation. Carol sees drift the moment her live rulesets disagree
with the committed mirror.

Nothing is ever applied uninvited: verify is the default, apply is an
explicit, confirmed sub-verb that owns its own ordering (a required check
lands on the default branch *before* the ruleset requires it), and the launch
preflight gates on the verdicts where a gate is honest.

## Why This Matters

Captured 2026-07-17, the day abcd-cli's own `main` was protected by hand.
Extended 2026-08-19 from a full day of building the collaborator fence on
this repository manually — org migration, ruleset split, release environment,
merge queue, author-conditional required checks — recorded in
[`2026-08-19-collaborator-fence-field-research.md`](../../research/notes/2026-08-19-collaborator-fence-field-research.md).
The field evidence reframes the original three tiers: the fence is not one
protection with fallbacks but a **capability ladder** whose rungs exist or
not depending on forge, plan, and ownership — and the tool's job for the
amateur coder is to say truthfully which rung they are on and what it does
and does not hold, in the idiom the banlist's `reach:` lines already ship.

The original three-rung structure stands, generalised from branch protection
to the fence:

1. **Verify and report — the doctor.** Probe read-only, host-delegated
   (`gh` as the opt-in adapter), degrade loudly: no `gh`, unauthenticated,
   or an API error is "unverifiable", never a silent green
   (`principles/loud-staging.md`). Identity derives from caller-local facts
   (the noreply login), never from repository ownership — the org transfer
   falsified the ownership assumption in one day (iss-283).
2. **Apply on request — an explicit sub-verb, never bare.** Idempotent,
   tier-appropriate, ordering-aware (workflow before ruleset flip, per the
   field note's twice-bitten bootstrap), and it refreshes the committed
   mirror (`.abcd/work/rulesets/`) in the same change. Never scaffold a gate
   the repo has not configured — a gate that *looks* armed and is not is the
   false green (the `.Abcd` conditional lesson).
3. **Gate where it matters — the launch preflight.** A doctor warning is not
   enforcement; the launch gate can refuse a cut. The gate reads the verdict
   vocabulary: what it refuses on per tier is an open question below.

## Acceptance Criteria

- Given Alice's repository has no git remote, When the doctor runs, Then
  every forge-dependent fence piece reports **impossible here** with a
  one-line reason, the local-tier pieces (hooks, banlist, guard, preflight)
  report their own true state, and the run exits as a supported state — Tier
  0 is a report, never an error.
- Given Bob's repository is on a personal account, When the doctor runs, Then
  organisation-only pieces (the role ladder) report **impossible here**
  naming the organisation as the remedy, pieces his plan offers but has not
  enabled report **available but off** with the enabling lever, and nothing
  is applied.
- Given Carol's live ruleset differs from the committed mirror under
  `.abcd/work/rulesets/`, When the doctor runs, Then it reports the drift
  naming the rule that moved (closes the iss-277 gap).
- Given any probe cannot answer — `gh` absent, unauthenticated, API error —
  When the doctor runs, Then that piece reports **unverifiable** and the
  summary can never render green above it.
- Given a piece whose rule is specified but unenforced (retention with no
  executor, iss-282), When the doctor runs, Then it reports **rule without
  executor** — distinct from applied and from impossible.
- Given Carol confirms the apply sub-verb for a piece with a bootstrap
  ordering (a required check), When it applies, Then the workflow lands on
  the default branch before the ruleset requires the context, and the
  committed mirror is refreshed in the same change.

## Open Questions

- **What does the launch gate refuse on, per tier?** Refusing Alice's cut for
  pieces impossible on Tier 0 punishes her environment; refusing only
  available-but-off pieces makes the gate toothless on Tier 0. A per-tier
  gate policy (perhaps: refuse on unverifiable and available-but-off, report
  impossible-here) needs the planning interview.
- **The solo-admin caveat stands:** without `enforce_admins`, protection is
  advisory against the admin's own token — the honest claim is "verified,
  reported, and release-gated", never "guaranteed".
- **Which protection shape is "protected enough" per tier?** Force-push and
  deletion blocks plus required checks as the floor; required reviews
  deadlock solo maintainers; the two-reviewer external rule (iss-281) is
  org-tier only. Per-repo template in `.abcd/config`?
- **Forges beyond GitHub** (or air-gapped remotes): the whole forge column is
  "unverifiable — unsupported forge", loudly, until an adapter exists.

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
