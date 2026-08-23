---
id: adr-52
slug: the-semantic-gate-sits-on-the-wrong-side-of-the-tag
status: proposed
date: 2026-08-23
supersedes: null
superseded_by: null
related_intents: []
related_rfcs: []
related_adrs: [adr-37, adr-38]
---

# ADR-52: The semantic gate sits on the wrong side of the tag

## Context

Changelog-driven auto-release ([adr-37](0037-changelog-driven-releases.md))
runs `detect` → `tag` → `release`. The deterministic gates run in `release`'s
`verify` job. The semantic gate — `receipt_gate`, verifying the host-run LLM
passes CI cannot run — runs in the `release` job, after the environment
approval.

The tag is therefore created **before** the gate that can refuse the release,
and `detect` states the tag is never moved once it exists. That immutability is
load-bearing: it is what makes the heal path safe, because a re-release always
rebuilds from the same immutable commit rather than from a moved-on default
branch.

The two properties compose badly. When `receipt_gate` refuses, the result is not
a blocked release but a consumed version:

- the tag exists and will not be moved;
- the heal path re-releases **from the tagged commit**, so every retry checks out
  the same tree and derives the same content commit;
- that tree can never gain the missing receipts, because a receipt must live in a
  commit that is an ancestor of the released commit and name its predecessor.
  Landing receipts on the default branch afterwards does not change the tagged
  tree.

No sequence of pushes to the default branch heals it. Recovery requires deleting
a tag the machinery treats as immutable, or abandoning the version.

Observed 2026-08-23 (iss-2608231226347380): v0.6.2 tagged at `b06fa80`, the
receipt gate refused for want of the two semantic receipts, no Release was
published, and the version was recoverable only by deleting the tag. The
deterministic gates did not have this problem — they run in `verify`, before
`tag`. Only the semantic gate sits on the wrong side.

## Decision

**None yet.** This ADR is `proposed` and records the problem and the option
space so the next release does not rediscover it. The two surface-level defects
that made the failure likely are already fixed separately
(iss-2608231226274000, iss-2608231226342272); this is the structural one
underneath them, and it is deliberately not being patched during a release
recovery.

## Alternatives Considered

1. **Arm `receipt_gate` in `verify`.** `verify` already runs before `tag` on the
   deterministic side, so the semantic gate would join the gates that refuse
   before a version is consumed. The cost: `verify` currently runs on a commit
   whose relationship to the eventual content commit is the same derivation the
   `release` job makes, so the arming logic moves rather than simplifies, and
   `verify` is also invoked on the rehearsal path where no receipts exist.

2. **Gate tagging in `detect`.** `detect` already reads the tree and decides
   `need_tag`; it could refuse to tag when the required receipts are absent for
   the commit that would become the content commit. This keeps the tag
   immutability guarantee untouched and fails before anything is consumed. The
   cost: `detect` would need the required-gates list, which `release.yml` owns
   deliberately as the trust root, so either the list moves or `detect` performs
   a weaker presence-only check that can still let an invalid receipt through to
   the real gate.

3. **Accept the wedge and document the recovery.** Add the tag-deletion
   procedure to the runbook and treat a consumed version as the cost of a
   fail-closed semantic gate. The cost: it normalises deleting tags, which is
   precisely the guarantee the heal path depends on, and it leaves the most
   expensive failure mode as the documented one.

4. **Verify the gate locally before merging.** Not an alternative to the above
   but a mitigation available today: `record-lint --release-gate <content-sha>
   --require-gate …` reproduces the gate's verdict on a release branch before
   the merge that triggers tagging. It converts the wedge into a pre-merge
   refusal without touching the machinery, and belongs in the runbook whichever
   structural option is chosen.

## Consequences

While this stays `proposed`, every release carries the risk that a semantic
refusal consumes its version, and the mitigation is procedural: run the passes
and prove the gate locally before merging the release branch.

Whichever option is adopted, it touches the tag-immutability guarantee that
makes the heal path safe, so it wants a deliberate decision rather than a patch
applied mid-recovery — which is why this is recorded rather than fixed.
