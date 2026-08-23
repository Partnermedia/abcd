---
schema_version: 1
id: "iss-2608231226347380"
slug: "tagging-precedes-the-receipt-gate-so-a-failure-wedges-the-version"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "v0.6.2 release failure post-mortem 2026-08-23"
found_at: ".github/workflows/auto-release.yml"
---

`auto-release.yml` runs `detect` -> `tag` -> `release`, and the semantic
`receipt_gate` runs inside the `release` job, after the environment approval.
The tag is therefore created BEFORE the gate that can refuse the release.

When the gate refuses, the result is not a blocked release but a wedged one:

- The tag exists and is immutable by design — `detect` states it is NEVER moved.
- The heal path re-releases FROM the tagged commit, resolved to its immutable
  sha, so every retry checks out the same tree and derives the same content
  commit.
- That tree can never gain the missing receipts, because the receipts must live
  in a commit that is an ancestor of the released commit and name its
  predecessor. Landing them on the default branch afterwards does not change the
  tagged tree.

So the version is consumed. Recovery requires deleting a tag the workflow treats
as immutable, or abandoning the version and moving to the next one. Neither is a
path the release machinery offers.

Field hit 2026-08-23: v0.6.2 tagged at b06fa80, receipt gate refused, no Release
published, and no sequence of pushes to the default branch can heal it.

The deterministic gates do not have this problem — they run in `verify`, before
`tag`. Only the semantic gate sits on the wrong side of the tag. Whether the fix
is to arm `receipt_gate` in `verify`, to gate tagging on receipt presence in
`detect`, or to accept the wedge and document the recovery, is an ADR-shaped
question rather than a patch: it touches the immutability guarantee that makes
the heal path safe. Related: iss-2608231226342272 (the preview is silent about
this gate) and iss-2608231226274000 (the surface never documents the step).