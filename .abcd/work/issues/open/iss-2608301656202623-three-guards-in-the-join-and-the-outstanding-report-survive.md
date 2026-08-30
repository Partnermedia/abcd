---
schema_version: 1
id: "iss-2608301656202623"
slug: "three-guards-in-the-join-and-the-outstanding-report-survive"
severity: "nitpick"
category: "tech-debt"
source: "user-observation"
found_during: "itd-189-round-5-ruthless"
found_at: "internal/core/lint/schema.go"
---

three guards in the join and the outstanding report survive their own deletion and one equality check is inert by construction

Found by the round-5 ruthless review, reported as notes rather than findings.
Recorded so they are not rediscovered as new.

Three guards survive their own deletion against the full lint and capture
suite:

- `spellsHandleOf`'s `rest == ""`: deleting it turns the `proposal: rdi-`
  finding into silence.
- `checkRecordBucketField`'s `isAbsentValue(got)`: deleting it adds a second,
  weaker finding on `run: ""`.
- the Unadmitted leg's `answer.standing.wellFormed`: deleting it reports one
  item as both `Unreadable` and `Unadmitted`.

Separately, `admittedProposals`' `proposal == ""` check is INERT rather than
defensive: `readingItemFileRe` guarantees the queried item is never empty, so
the map lookup it guards cannot be reached. That is the same class round 4 swept
in `31107a30`, with one instance missed.

The reviewer also confirmed three sets of negative assertions are pinned to
strings that occur nowhere in production code and so pass unconditionally. It
did NOT report them as defects, and neither does this record: each sits beside a
positive assertion on the current wording and each commit states the
anti-regression intent. They are deliberate reintroduction pins. Noted only so a
later reader does not mistake them for live coverage.
