---
id: adr-17
slug: rp-chat-send-override-supersedes-acj1-env-skip
status: accepted
date: 2026-06-11
supersedes: null
superseded_by: null
related_intents: []
related_rfcs: []
related_adrs: []
---

# ADR-17: `rp chat-send` becomes a declared abcd override — fixed budget pre-flight on every path, delegated send; scoped supersession of fn-33 AC-J1's configuration-only mechanism

> Supersession note: this ADR supersedes a **spec-level** decision (fn-33
> Cluster J / AC-J1, "Configuration, not interception"), not a prior ADR — so
> the frontmatter `supersedes` chain field stays `null` and the linkage is
> recorded here and in Related Documentation.

## Context

Direct CLI invocations of `flowctl rp chat-send` wedge with `[-32602] Invalid
params: path already exists: $TMPDIR/rp-render-*.txt`. Root cause: the
vendored `scripts/ralph/flowctl.py` chat-send budget pre-flight calls
`_rp_render_export`, which pre-creates the export target via
`NamedTemporaryFile(delete=False)`; RepoPrompt's `prompt export` refuses
existing paths (verified on RP 2.1.32 — no overwrite flag exists).

fn-33 Cluster J / AC-J1 mitigated this with **configuration, not
interception**: the dispatcher exports `FLOW_RP_CHAT_BUDGET_SKIP=1`
process-scoped to `--abcd-driven` invocations only. Contract #1 deliberately
left direct calls "vanilla upstream, bug intact", and the fn-33 tradeoff
paragraph accepted that driven sends get **no pre-flight at all** ("forgoes
the trim loop... fails at the model-input wall"). The fn-33 spec itself named
an abcd-owned budget verb invoked from the abcd-driven flow as a recorded
"future option".

The 2026-06-09 session showed the boundary was drawn through the wrong path:
the direct CLI is exactly the fallback agents reach for when RepoPrompt's MCP
connection dies mid-session (RP app restart), so the one path needed in that
failure mode was the one path that wedged (4/4 attempts, with and without
`--selected-paths`).

Constraints already locked when fn-44 was decided:

- `scripts/ralph/flowctl.py` is byte-sealed and upstream-unforked (wrap-only
  rule; the seal gate is green) — the fix may invoke it, never edit it.
- The fn-33 rejected-shim design (an `rp-cli` PATH shim) stays rejected, along
  with its risk classes: recursion, delete-predicate, TOCTOU, PATH
  manipulation.
- The dispatcher routes a declared registry verb (`verbs.json`,
  `override_upstream: true`) to an abcd handler — the same mechanism already
  carrying `rp setup-review`, `spec close`, and `epic close`. All overrides
  are declared, never silent shadowing.
- Upstream flow-next HEAD (1.10.2 era) has ALREADY removed the render
  pre-flight and the SKIP guard from `cmd_rp_chat_send` — a future re-vendor
  makes the bug disappear on its own.

## Decision

fn-44 intercepts `rp chat-send` as a declared abcd override routing to
`abcd_flowctl_ext:rp_chat_send`, which applies `--selected-paths` to the live
RP selection, runs abcd's already-fixed `mkdtemp`-based budget pre-flight, and
delegates the actual send to the vendored `flowctl.py rp chat-send` with
`FLOW_RP_CHAT_BUDGET_SKIP=1` on the child env. Four decisions are recorded:

1. **Scoped supersession of "Configuration, not interception".** The fn-33
   AC-J1 principle is superseded **for chat-send ONLY** — contract #1 (a
   direct invocation routes to vanilla upstream) is intact for every other
   verb. The objections that killed the fn-33 shim design do NOT apply to a
   declared registry override: no recursion (the handler invokes the vendored
   `flowctl.py` directly as a child, never the dispatcher); no
   delete-predicate (nothing deletes someone else's temp files — abcd's own
   pre-flight allocates via `mkdtemp` and never pre-creates the export
   target); no TOCTOU (no check-then-delete window exists); no PATH
   manipulation (routing is an explicit `verbs.json` entry consumed by the
   dispatcher's Python router, not an executable shadowed on `$PATH`). The
   override is declared and enumerated in the same registry as the existing
   `rp setup-review` / `spec close` / `epic close` overrides.
2. **Driven-path enforcement REVERSAL.** fn-33's accepted "no pre-flight on
   driven sends" tradeoff is deliberately reversed: an `--abcd-driven`
   chat-send now gets the abcd pre-flight — one extra RP render round-trip,
   a prompt set/measure/restore cycle internal to `rp_render_payload`,
   live-selection trimming, and a new budget-exceeded fail-fast (exit 2,
   stderr-only, before any child is spawned). The cost is accepted: budget
   enforcement on the driven path is reintroduced rather than forgone.
3. **`--selected-paths` semantic upgrade — durable even on failure.** Upstream
   on RP 2.1.32 treats the flag as payload-only, and the modern oracle payload
   drops it — silently ignored. abcd upgrades it to a **durable live-selection
   set**: applied to the live RP selection BEFORE the pre-flight (so the trim
   loop and the send see the same state) and NOT forwarded to the child. The
   mutation is durable on success AND on failure — after a successful send the
   live selection is the post-trim survivors (the operator's input is consumed
   destructively); on a failed pre-flight or send NO restore is attempted (a
   failure-path restore can itself fail or race; the operator sees exactly the
   selection the failed send would have used).
4. **The dispatcher's driven-branch SKIP export is retained-but-vestigial as
   the fn-44 ROLLBACK PATH.** After fn-44 no upstream-routed verb reads
   `FLOW_RP_CHAT_BUDGET_SKIP` (chat-send was the only consumer, and it can no
   longer reach the upstream branch while the override is live). The export is
   RETAINED deliberately: reverting fn-44 — drop the `verbs.json` entry plus
   the `_dispatch.py` guard-list entries — instantly restores the AC-J1
   protection for driven invocations with zero dispatcher changes. The
   dispatcher comment and the pinning test (asserted with a NON-chat-send
   upstream-routed verb) label it "retained as fn-44 rollback path — vestigial
   while the override is live".

**Enforcement owner after re-vendor.** Upstream HEAD has already removed the
pre-flight and the SKIP guard, so a future re-vendor makes the child-env SKIP
a harmless no-op and the `-32602` bug disappears on its own. The override does
not retire to nothing: the abcd handler remains the budget **enforcement
owner** — the lasting value of the override is budget enforcement on every
invocation path, not the bug workaround.

## Alternatives Considered

- **Keep AC-J1's env-skip only (status quo).** Rejected: the direct CLI stays
  100% wedged, and the 2026-06-09 incident showed it is exactly the fallback
  path used when RP's MCP connection dies mid-session. A mitigation scoped
  away from the failure mode it is needed in is not a mitigation.
- **`rp-cli` PATH shim (fn-33's original rejected design).** Still rejected,
  for the original reasons: recursion risk, delete-predicate, TOCTOU, PATH
  manipulation. fn-44's declared registry override carries none of these
  (decision 1).
- **Edit the vendored `flowctl.py`.** Rejected outright: the file is
  byte-sealed and upstream-unforked (wrap-only rule); the seal gate would go
  red and every re-vendor would re-open the wound.
- **Re-vendor upstream HEAD now (the pre-flight is already removed there).**
  Rejected for this spec: a re-vendor is a separate heavyweight operation with
  its own blast radius, and it would fix only the bug — abcd would still have
  no budget enforcement owner on the chat-send path. The override delivers now
  and self-retires into enforcement owner when the re-vendor eventually lands.

## Consequences

- Every chat-send invocation path — direct CLI, `--abcd-driven`, standalone
  `abcd-rp` — is safe from the `-32602` pre-create wedge; the MCP-death
  fallback path works.
- Budget enforcement is preserved on direct paths and REINTRODUCED on the
  driven path, at the cost of one extra RP render round-trip and a prompt
  set/measure/restore cycle per driven send (decision 2's accepted reversal).
- `--selected-paths` is consumed destructively: operators get a durable
  live-selection set, including after failures — a deliberate departure from
  upstream's silently-ignored payload flag (decision 3).
- The rollback path is two reverts away (verbs.json entry + `_dispatch.py`
  guard lists), with zero dispatcher changes, because the vestigial SKIP
  export stays in both committed dispatcher copies — a standing obligation to
  keep the export and its "rollback path" labeling in lockstep
  (byte-identity-tested) until the override itself is retired.
- After a future re-vendor the child-env SKIP becomes a no-op and the abcd
  handler remains the budget enforcement owner; nothing needs to be undone.
- "Configuration, not interception" survives as the default posture for every
  other verb; chat-send is the recorded, declared exception.

## Related Documentation

- [`../../../../.flow/specs/fn-44-rp-chat-send-override-fixed-budget-pre.md`](../../../../.flow/specs/fn-44-rp-chat-send-override-fixed-budget-pre.md) — the fn-44 spec (design decisions 1–8, boundaries, test surfaces)
- [`../../../../.flow/specs/fn-33-phase-3-to-4-cleanup-placeholder.md`](../../../../.flow/specs/fn-33-phase-3-to-4-cleanup-placeholder.md) — fn-33 Cluster J / AC-J1, the superseded spec-level decision (amendment pointer next to its locked-decisions block)
- [`../../../../scripts/abcd/overlay/sources/flowctl-dispatcher.sh`](../../../../scripts/abcd/overlay/sources/flowctl-dispatcher.sh) — the dispatcher carrying the retained rollback-path export
- [`../../../../scripts/abcd/README.md`](../../../../scripts/abcd/README.md) — module table (handler in `abcd_flowctl_ext.py`, standalone front door in `abcd_rp_cli.py`)
