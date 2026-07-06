---
id: adr-18
slug: launch-payload-excludes-memory-gate-scoped-to-lifeboat
status: superseded
date: 2026-06-13
supersedes: null
superseded_by: adr-28
related_intents: [itd-36]
related_rfcs: []
related_adrs: [adr-4, adr-5]
---

# ADR-18: The public launch payload excludes `.abcd/memory/**` as policy; the restrictive-licence gate is scoped to the lifeboat, future/inert at launch

> Superseded by [ADR-28](0028-single-repo-curated-release.md) — one curated repo
> with a release-artifact view replaces the dev→public payload split; `.abcd/**`
> is excluded from the release artifact by packaging.

## Context

fn-38.7 built a launch-side restrictive-licence gate
(`scripts/abcd/launch_licence_gate.py`): a default-deny `sources/` allowlist
plus restrictive-licence detection layered on the fn-38.1 SPDX classifier. It
was wired into the launch dry-run renderer (`scripts/abcd/launch.py:402`,
`gate_dry_run()`).

That gate guards a payload it can never see. The launch payload manifest
(`.abcd/development/brief/04-surfaces/04-launch.md` §2, mirrored in
`launch.py:72-78` `_PAYLOAD_EXCLUDE`) excludes `.abcd/` — the entire namespace,
including `.abcd/memory/` — **wholesale** from the public payload. So no file the
gate's allowlist re-includes (`.abcd/memory/sources/...`) is ever in the publish
walk. The gate evaluates a PINNED fn-38 input set (on-disk `sources/` files plus
registry-linked memory pages) that the launch payload itself never publishes. The
`launch.py:402` call therefore makes launch *claim* to enforce a gate over files
it has already excluded.

`.work/issues.md:1906` (recorded during fn-38 planning) flagged this exact risk —
"do not silently ship a dead gate" — and it shipped anyway. Launch is still a
Phase-5 stub (`_STUB_FOOTER`), so nothing leaks today, but the contradiction is a
compliance landmine for the bootstrap launch (Phase 5 promotes abcdDev → public
abcd) and it shapes what the publish payload *is*. fn-38 §R8 durably claims the
no-op was "RESOLVED (.7) … provably not a dead gate" — that claim is what this
decision overturns.

Two distinct override mechanisms exist, and they are easy to conflate:

1. **`.abcd/launch.allow`** (`04-launch.md:31`; the file at `.abcd/launch.allow`) —
   a line-based allowlist read by `launch ship`. Each path is promoted to the
   public sibling *in addition to* the §2 include list. This is the mechanism
   that can actually put a path INTO the public payload. Its current comments
   carry no memory/provenance prohibition.
2. **`.abcd/launch-allowlist.json`** (`launch_licence_gate.py:76`, `_ALLOWLIST_REL`) —
   a JSON allowlist read only by the gate, naming `.abcd/memory/sources/` files
   that the gate evaluates (default-deny: an un-named `sources/` file is refused
   by the gate). It re-includes files into the GATE'S INPUT SET, never into the
   publish payload.

A name/semantics drift (`launch.allow` vs `launch-allowlist.json`) and the
override-reinclusion question are both policy hinges, not implementation details.

Constraints already locked: adr-5 (the brief is current state — live brief
sections must be edited to match a decision, not merely ADR-superseded); adr-4
(the lifeboat is regenerable output, and `/abcd:disembark` — a Phase-4 surface —
is where curated project memory/provenance is exported). The lifeboat's current
output shape (`02-disembark.md §5`) contains `research/pitfalls.{json,md}`
("source memory + Pass B deltas"), `assets/_manifest.json` (asset
provenance/classification), and a root `_provenance.json` — but **no verbatim
`.abcd/memory/` payload and no root memory `_manifest.json`**. So the lifeboat is
a *candidate* consumer of the gate, not a wired one.

The operator confirmed the plan-time lean (B) at the planning interview: the
public plugin repo carries plugin code, never per-project knowledge stores;
publishing curated memory/provenance is the lifeboat's job. This ADR records the
full evidence; the lean could be overturned only with cited contrary evidence,
and none was found (the brief's launch §2 has always excluded `.abcd/` — it never
promised memory publication).

## Decision

**We adopt (B): the wholesale exclusion of `.abcd/memory/**` from the public
launch payload is the policy, and the restrictive-licence gate is scoped to the
lifeboat/disembark surface — future/inert at launch against a named payload set,
never an unwired "the real consumer is lifeboat" assertion.**

Concretely:

1. **The launch payload excludes `.abcd/memory/**` (and the wider `.abcd/`
   namespace) as deliberate policy.** The public abcd repo is the plugin; it does
   not carry abcdDev's project memory. `_PAYLOAD_EXCLUDE` and §2 stand.

2. **The licence gate's real consumer is the lifeboat (`/abcd:disembark`),** which
   *is* the surface that publishes curated project memory/provenance (adr-4). At
   launch the gate is **future/inert against a named payload set** — it is not the
   launch payload's gate.

3. **The `launch.py:402` `gate_dry_run()` call is resolved (R2a, fn-50.2):** under
   (B) it is removed from the launch dry-run path or relabelled as non-launch
   diagnostics, so launch never claims to enforce a gate over files it excludes.

4. **The gate's consumer is named, not asserted (R2b, fn-50.2).** Because
   `02-disembark.md §5` does not today declare a verbatim `.abcd/memory/` payload,
   fn-50.2 chooses ONE of:
   - **(i)** update `02-disembark.md §5` to declare the exact memory/provenance
     payload paths the gate is scoped to, and add a consistency test asserting
     launch's payload sets and the gate's scope claims agree; OR
   - **(ii)** declare the gate **future/inert against a named payload set** until a
     later disembark spec wires it.

   **This ADR selects (ii): declare the gate future/inert against the lifeboat's
   already-existing provenance surface** (`02-disembark.md §5`'s
   `assets/_manifest.json` provenance/classification and root `_provenance.json`,
   plus `research/pitfalls.{json,md}` as the curated-memory surface), without
   inventing a new `.abcd/memory/`-verbatim lifeboat payload that fn-50 is not
   chartered to build (the spec's non-goals forbid building the lifeboat packer).
   fn-50.2 names this set in the gate docstring + brief and pins, with a test, that
   launch's payload sets and the gate's declared scope cannot contradict (the gate
   is NOT scoped to anything launch publishes). Option (i) — adding exact new
   payload paths to §5 — is deferred to the disembark spec that actually wires the
   packer, because asserting a payload §5 does not yet declare would reintroduce
   the same unwired-claim defect this ADR closes.

5. **No launch override may re-include `.abcd/memory/**` into the public payload
   (R2d).** The prohibition binds **`.abcd/launch.allow`** — the only mechanism
   that can promote a path into the publish payload. fn-50.2 adds the prohibition
   to `launch.allow`'s header contract and pins it with a test (a `.abcd/memory/**`
   line in `launch.allow`, or any `.abcd/` path, must be refused / never promoted).
   The gate's `.abcd/launch-allowlist.json` is NOT a payload-promotion mechanism —
   it only re-includes files into the gate's own evaluation input — so it does not
   reopen the contradiction and is not bound by this prohibition; but its purpose
   is documented as distinct from `launch.allow` (see 6) so the two are never
   conflated into a payload-promotion path.

6. **The `launch.allow` vs `launch-allowlist.json` name drift is reconciled as two
   documented-distinct purposes, not one name (R2d).** They are genuinely different
   mechanisms (payload promotion vs gate-input re-inclusion); collapsing them to one
   name would be wrong. fn-50.2 documents the distinction at both definition sites
   (`launch.allow` header, `_ALLOWLIST_REL` docstring) and in `04-launch.md`.

The decision is unambiguous on the policy hinge: **(B) forbids re-including
`.abcd/memory/**` via the payload-promotion override; (A) is not adopted.**

## Alternatives Considered

- **(A) Carve memory/provenance INTO the public payload under the licence gate**
  (exclude-set carve-in; the gate becomes the live launch payload's gate). The gate
  would fire over a real subset of `.abcd/memory/` in the publish walk, and a
  payload-preview test would prove it refuses a restrictive-licence fixture.
  *Rejected:* it inverts the long-standing §2 policy that the public plugin repo
  carries plugin code, not abcdDev's project knowledge store. The operator
  confirmed the public repo must never carry `.abcd/memory`. No brief evidence
  promises memory publication at launch — §2 has always excluded `.abcd/`
  wholesale. Carving memory in would also force a per-file licence decision into
  the launch path for content whose natural publication surface is the lifeboat.

- **(B) Wholesale exclusion is policy; gate scoped to the lifeboat** — **chosen.**
  Matches §2, adr-4 (lifeboat owns curated memory/provenance export), and the
  operator's confirmation. Closes the dead-gate contradiction by removing the
  false launch→gate claim and pinning the gate's scope to a surface launch does
  not publish.

- **(B-i) within B: add a new verbatim `.abcd/memory/` payload to `02-disembark.md
  §5` now.** *Rejected for fn-50:* fn-50's non-goals forbid building the lifeboat
  packer, and §5 does not declare such a payload today. Declaring payload paths
  that no surface yet produces would recreate the same unwired-claim defect. The
  exact memory payload, if any, is the disembark spec's call (deferred).

- **Leave the gate wired at launch as a documented no-op (status quo).**
  *Rejected:* this is precisely the dead gate `.work/issues.md:1906` warned
  against and fn-38 wrongly claimed resolved. A gate the launch path invokes over
  files launch excludes is a standing contradiction and a compliance landmine for
  the bootstrap promotion.

## Consequences

**Easier / gained:**

- The dead-gate contradiction is closed structurally: launch no longer invokes a
  gate over files it excludes (R2a), and a test pins that launch's payload sets and
  the gate's declared scope cannot contradict (R2b-ii).
- The override surface is safe: `.abcd/memory/**` cannot be promoted into the
  public payload by `launch.allow`, pinned by a test (R2d). The bootstrap launch
  cannot accidentally leak project memory.
- The two override mechanisms are documented-distinct, ending the
  `launch.allow` / `launch-allowlist.json` conflation risk.

**Harder / new obligations (fn-50.2):**

- **adr-5 coherence sweep (R2c):** every LIVE brief section asserting launch
  consumes the memory/provenance gate must be EDITED, not merely ADR-superseded.
  At minimum: `04-surfaces/04-launch.md §2`; `05-internals/07-memory.md` (the
  "launch-gate refuses to publish anything under `.abcd/memory/sources/`" lines at
  :38, :117, :166 — reframe as the lifeboat/disembark gate, not a launch payload
  gate); `05-internals/09-provenance-substrate.md` (§4 "Launch-gate licence
  gates", and the `/abcd:launch consumes the registry for licence-gate
  enforcement` line at :108 — relabel to the lifeboat consumer). The shipped-intent
  record `roadmap/intents/shipped/itd-36-memory-unification.md` is checked and
  reconciled by edit or explicit supersession reference. Supersession-reference-only
  is permitted ONLY for durable historical records (fn-38 prose, shipped-intent
  caveats), never to leave a live brief section contradicting this policy.
- **Licence-predicate semantics are untouched (R3):** `is_external_class`, SPDX
  classification, restrictive refusal, unknown-warn, override behaviour, and
  `source_token_count` stay green unchanged. Only the gate's scope/config/wiring
  and the launch-INTEGRATION assertions may change.
- **Ledger + durable-prose reconciliation (R4):** `.work/issues.md:1906` annotated
  `resolved-by: fn-50-launch-payload-policy-resolve-the-dead`; the fn-38.7 "documented
  no-op" caveat and fn-38 §R8's "provably not a dead gate" claim reconciled by an
  explicit supersession reference to this ADR.
- **Open obligation deferred to a later disembark spec:** if/when the lifeboat is to
  carry a verbatim `.abcd/memory/` payload, that spec declares the exact §5 paths and
  wires the packer to the gate. This ADR deliberately does not invent that payload.
