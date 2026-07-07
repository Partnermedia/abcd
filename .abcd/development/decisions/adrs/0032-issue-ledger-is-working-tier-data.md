---
id: adr-32
slug: issue-ledger-is-working-tier-data
status: accepted
date: 2026-07-08
supersedes: null
superseded_by: null
related_intents: [itd-4, itd-36]
related_rfcs: []
related_adrs: [adr-3, adr-5, adr-26, adr-30]
---

# ADR-32: The issue ledger is working-tier data, not authored record

## Context

The capture ledger (itd-4) writes schema-checked `iss-N` files with folder-as-status
(`open/`, `resolved/`, `wontfix/`). It was sited at `.abcd/development/activity/issues/` —
inside the durable design-record tree that `record-lint` governs.

Dogfooding surfaced a shipped-vs-shipped contradiction: `capture`'s issue schema *requires*
a `created:` field, while `record-lint`'s `no_git_metadata` rule *bans* `created:` anywhere
under `.abcd/development` (git log/blame is canonical). Running `capture` therefore broke
`record-lint` — two shipped components mutually incompatible, the moment the tool is used on
its own repo. The finding is itself in the ledger as `iss-15`.

The root cause is a category error: the ledger is generated, schema-governed *data* that
opens and resolves, but it was living in the tree meant for authored, git-canonical *prose*.

## Decision

1. **Tier.** The issue ledger lives in the **work tier** — `.abcd/work/issues/` — not the
   design record. It is committed, shared working state (folder-as-status per adr-3), not
   durable design record. This takes it out of `record-lint`'s root and dissolves the conflict
   at its source.

2. **No git-inferable timestamps.** `created` and `updated` are dropped from the schema; git
   is the canonical source of time (consistent with `no_git_metadata` and the
   derive-don't-store principle). The reader **tolerates** legacy `created`/`updated` on read
   (accept, then drop) so an older ledger degrades gracefully rather than being rejected.

3. **Priority is derived, never stored.** Issues carry a one-direction `blocked_by: [iss-N]`
   dependency edge; the inverse and the ranking are computed. `capture list`/status order
   *unblocked-first, then by severity*, annotating blocked rows `[blocked-by iss-N]`. This
   reuses the spec engine's dependency-graph model (adr-26, the run-seam selector) rather than
   inventing a parallel priority label.

4. **README contracts by category.** Per-folder READMEs stay for the authored record — they
   are local membership contracts. A generated-data store gets **one** README at its root; its
   state leaves are self-evident. The tier boundary keeps `directory_coverage` aimed only at
   prose, so the lint needs no change.

## Alternatives Considered

- **Exempt `.abcd/development/activity/` in `record-lint`.** The first-pass fix, briefly
  recommended and then reversed. Rejected as symptom suppression: it carves the lint around a
  category error rather than fixing it, and leaves the ledger in the wrong tier. The reversal
  is the methodology-over-local-fixes principle in action — recorded so the reasoning is not
  re-litigated.
- **A stored `priority` field.** Rejected: priority is volatile and contextual; a stored label
  is the hand-maintained-status drift the roadmap dashboard and adr-5 already avoid.
- **Blanket-exempt all of `.abcd/` from READMEs.** Rejected: it discards the local contracts
  the authored record relies on, to fix a problem that only exists for generated data.
- **Keep the timestamps as portable data.** Rejected: the ledger is read in-repo with git
  present; git derives creation and the lifecycle moves, and the store is not bundled in the
  release artefact.

## Consequences

- The `capture` ⊥ `record-lint` conflict is gone at the root; running `capture` no longer
  breaks the gate.
- Design-record references to the old path (e.g. `build-sequence.md`) are now drift to
  reconcile, tracked as ledger issues.
- Future generated stores (logbook, lifeboat) inherit the rule: working/generated data lives
  outside the record tier and carries no git-inferable metadata.
- A self-hosting datapoint: abcd's first real contradiction was found by running abcd on abcd
  (`iss-15`) — the dogfooding thesis paying out in practice.
