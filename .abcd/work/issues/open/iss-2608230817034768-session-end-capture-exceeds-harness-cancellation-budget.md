---
schema_version: 1
id: "iss-2608230817034768"
slug: "session-end-capture-exceeds-harness-cancellation-budget"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "session-end 'Hook cancelled' post-mortem 2026-08-23"
found_at: "internal/surface/cli/cli.go"
---

`hook session-end` redacts the whole transcript in-line before it writes, at
roughly 0.7 s per MB, and the host cancels a SessionEnd hook during shutdown
rather than wait for it. Every session whose transcript is big enough to push
that work past the cancellation budget is therefore dropped — silently, and
permanently, since the capture is the only chance the transcript ever gets.
Field hit 2026-08-23: session `7d884491` exited with `SessionEnd hook [...]
failed: Hook cancelled` and never reached the store.

This is **not** iss-2608210934566223 recurring. That was the blocking bootstrap
download at exit, fixed in `baf0551`; on the field-hit machine the plugin-root
binary was present, so the bootstrap branch was never taken. The hook reached
`exec "$r/abcd" hook session-end` and was cancelled inside abcd's own work. The
two failures share a symptom and a cancellation mechanism, and nothing else —
`baf0551` removed one source of slowness at exit but left the budget itself
unguarded.

Measured over one repo's harness transcripts, replaying each through
`hook session-end` and checking it against `abcd history list`:

| transcript | size | redaction | captured |
| --- | --- | --- | --- |
| `8db3dbd6` | 0.01 MB | 0.22 s | yes |
| `6802420c` | 0.69 MB | 0.68 s | yes |
| `38ab27e8` | 1.25 MB | 1.12 s | yes |
| `f1a5692a` | 1.98 MB | 1.53 s | yes |
| `134e647f` | 3.33 MB | 2.75 s | no |
| `51fc2a94` | 3.96 MB | 2.93 s | no |
| `7d884491` | 4.11 MB | 3.02 s | no |
| `cc5a634b` | 5.10 MB | 3.77 s | no |
| `093cd456` | 6.35 MB | 4.78 s | no |
| `4fbbb0a3` | 8.10 MB | 5.65 s | no |
| `9c89b576` | 11.80 MB | — | no |

The split is clean and monotone: everything at or under 1.53 s was stored,
everything at or over 2.75 s was dropped, and the cliff sits somewhere in
between — call the budget two seconds until it is measured directly. Cost
tracks content as well as bytes (secret and home-path hit density), so size is
a proxy for the real variable, not the variable itself.

The consequence is that capture works precisely where it matters least. A short
session is cheap to redact and gets stored; a long, dense, expensive session —
the one actually worth keeping — is the one guaranteed to be dropped. Eleven of
this repo's transcripts are absent from its store, and the store's own listing
cannot distinguish "never ended" from "ended and lost", which is why the loss
went unnoticed for a week.

Directions, none of them taken here, and none to be taken before a detector is
armed: move the write ahead of the redaction and redact in place afterwards, so
a cancellation costs cleanliness rather than the record; or capture the raw
transcript at exit and redact on the next SessionStart; or hand the work to a
detached process the harness's shutdown does not own. Each trades a different
thing away, and choosing between them wants an ADR, not a patch.
`iss-2608210934566224` (the recovery sweep) is the detector this needs and is
now the blocking dependency, not a nice-to-have: without it there is no way to
watch a fix fail.
