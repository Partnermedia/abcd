---
schema_version: 1
id: "iss-2608301656193936"
slug: "the-bucket-blocker-claims-the-item-goes-on-being-reported-un"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "itd-189-round-5-ruthless"
found_at: "internal/core/lint/schema.go"
---

the bucket blocker claims the item goes on being reported unanswered for exactly the target shape the padding leg beside it stands down on

Found by the round-5 ruthless review. Branch-introduced, and the standing class
again.

The padding leg was deliberately taught to stand down on a target whose
filename is not itself a bare handle, because the family's reader never opens
that file. Its comment says so. The bucket leg immediately below has no such
gate and speaks about that shape anyway, ending:

```
...and the reading item it names goes on being reported as unanswered with no
sign that an answer was written
```

Probed with `readings/rdg-9/rdi-7-widen-the-frame.md` plus an admission naming
`rdi-7`: the gate emits that blocker while the report emits nothing at all
about `rdi-7`, because `readingItemFileRe` never matches that filename. The
operator is sent to find a report line that does not exist, and the remedy the
message implies -- refile the admission under `rdg-9` -- would not make it count
either, because no run reads that file.

So the diff reports on one shape two ways: one leg stands down on it by design,
and the leg beside it makes a false claim about it.

Remedy: the leading clause is true unconditionally, so append the "goes on being
reported as unanswered" tail only when `spellsHandleOf(join.sameBucketAs,
stem)` -- the same test the padding leg already computes one block above.
