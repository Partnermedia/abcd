---
schema_version: 1
id: "iss-2608301519254418"
slug: "two-message-legs-tell-the-author-the-reader-refuses-and-skip"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "itd-189-round-3-ruthless"
found_at: "internal/core/lint/schema.go"
---

two message legs tell the author the reader refuses and skips their record when the admission reader counts it and the surprise store has no reader at all

Found by the round-3 ruthless review. FIX-FIRST. **The fourth instance on this
branch of one class: a message asserting a mechanism that is not the case.**

`checkRecordUnknownFields` (schema.go:786-788) emits, as a BLOCKER, that a key
outside the schema "makes this a record the reader refuses and skips —
invisible to every admission surface while it still sits in the store". Probed:
`ReadReadingOutstanding` returns `Unadmitted=[]` — the record is not skipped and
not invisible, it is ACTIVELY COUNTED. For `srp` the message is worse: no
surprise reader exists anywhere in the tree, so "invisible to every surprise
surface" names an empty surface set and a refusal nobody performs. The author is
sent to look for a record that is being read.

`checkRecordDuplicateKeys` (schema.go:1201-1203) carries the same claim for
EVERY store: "the record reader refuses a duplicated key, so the file is skipped
by every admission surface". True for `iss` (capture/parse.go:122); false for
`adm`, `srp`, and — pre-existing — `rdi`, `rdg`, `dsp`.

The sting: `6226f7d2` fixed exactly this defect in `checkRecordRequiredFields`
one function away, and did not check the two legs beside it. The flag it added,
`readerFailsClosed`, is correct for all four stores (verified independently by
both reviewers) — it was simply applied to one leg.

Remedy, one condition, both sites: gate the reader clause on
`r.store.readerFailsClosed` exactly as `checkRecordRequiredFields:745` now does,
keeping the schema-closed half unconditional. `adr` is unaffected (nil
knownFields, early return); `iss` keeps the claim, which is true of it. Then
widen the flag's godoc: it covers required properties, the closed schema AND the
duplicate key — one declaration, three legs, rather than one leg declaring and
two assuming.
