---
schema_version: 1
id: "iss-2608301620343236"
slug: "the-len-rs-1-guard-in-scriptiocontinuaonly-is-unreachable-as"
severity: "nitpick"
category: "tech-debt"
source: "user-observation"
found_during: "itd-179-round-5-builder"
found_at: "internal/core/grounds/grounds.go"
---

the len(rs) != 1 guard in scriptioContinuaOnly is unreachable as called so no fixture can enter it

Reported by the round-5 builder and left in place with the reasoning stated in
`6330b8b2`'s body rather than claimed as tested, which was the right call at
the time.

`textUnits` emits a scriptio-continua letter as its own single-rune unit, so a
multi-rune unit never begins with one, and the sibling test
`!isScriptioContinuaLetter(rs[0])` already rejects every shape `len(rs) != 1`
would. Deleting it leaves the whole suite green, and the builder found no text
reaching `ValidateText` that distinguishes the two.

The branch has already ruled on this shape once, on its sibling: itd-189's
round 3 deleted two stand-downs under the finding that a branch no fixture can
enter is how a guard comes to look tested, and iss-2608301519254240 was raised
against the twin it left behind. Consistency says delete rather than test,
because a direct unit test on the helper would pin a shape the caller cannot
produce and would defend the guard against a future that has not been argued
for.
