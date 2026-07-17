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

# abcd verifies branch protection on managed repos and gates the launch on it

## Press Release

> _Seeded from a quoted-text intent capture. Expand into the full press-release narrative before planning._

## Why This Matters

abcd verifies branch protection on managed repos and gates the launch on it.
Captured 2026-07-17, the day abcd-cli's own `main` was protected by hand —
work the tool should carry for every repo it manages. The design discussion
that produced this capture sketched three tiers:

1. **Verify and report — `ahoy doctor`.** Doctor already inspects state
   beyond the working tree read-only and the engine already resolves the
   origin remote. A `github.branch_protection` gap: query the default
   branch's protection through the `gh` CLI as a host-delegated, opt-in
   adapter, and report a gap when protection is absent or missing the repo's
   required contexts. Must degrade loudly — no `gh`, unauthenticated, or an
   API error reports "unverifiable", never a silent green.
2. **Apply on request — an explicit sub-verb, never bare.** Writing
   protection is an outward-facing remote mutation: a transparent-confirm
   install step or dedicated verb applying an idempotent protection template
   (required contexts derivable from the repo's own CI check names).
3. **Gate where it matters — the launch preflight.** Per
   enforcement-claims-are-facts, a doctor warning is not enforcement; the
   launch gate suite (itd-65) can refuse to cut a release while the default
   branch is unprotected.

## Acceptance Criteria

> _Required (the itd-1 discipline): add at least one Given-When-Then bullet describing the verifiable bar for "shipped" before this draft can be planned._

## Open Questions

- **The solo-admin caveat:** without `enforce_admins`, protection is
  advisory against the admin's own token — the honest claim is "verified,
  reported, and release-gated", never "guaranteed". Should the doctor gap
  distinguish protected-but-admin-exempt from unprotected?
- **Which protection shape is "protected enough"?** Force-push/deletion
  blocks and required status checks seem like the floor; required reviews
  deadlock solo maintainers. Per-repo template in `.abcd/config`?
- **Adapter dependency:** `gh` as the query/write path keeps abcd
  host-delegated, but forges beyond GitHub (or air-gapped remotes) need the
  check to no-op loudly, not fail.

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
