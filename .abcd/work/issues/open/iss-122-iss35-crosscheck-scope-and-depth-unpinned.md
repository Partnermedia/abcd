---
schema_version: 1
id: "iss-122"
slug: "iss35-crosscheck-scope-and-depth-unpinned"
severity: "minor"
category: "tech-debt"
source: "agent-observation"
found_during: "v0.4.0 release gate"
found_at: ".abcd/development/release-gate/brief-surface-crosscheck.js"
---

The iss35 crosscheck's scope and depth are unpinned, so the gate is not reproducible: v0.3.0's receipt records zero findings four days ago while a full-depth run (17 brief docs, both directions, 22 checkers) returns 102 discrepancies, and the receipt's own promptHash field is the literal 'no-pinned-prompt' admission. The maintainer choosing briefDocs per run means two honest runs of the same gate can disagree by two orders of magnitude; the gate needs a pinned input manifest and depth so a PROMOTE means the same thing every release.
---

**Design decided (2026-07-24, maintainer grill; see DECISIONS.md):** a
committed input manifest under `.abcd/development/release-gate/` pins the doc
list, directions, checker count, and prompt hash, with **tiered depth** — full
depth for feature/breaking releases, a Direction-B-only shallow pass for patch
releases. The receipt must echo the manifest hash **and** the tier, and
receipt_gate refuses a receipt whose tier mismatches the release's declared
impact — so a PROMOTE is unambiguous within its tier. Automated refusal is
**procedural only** (manifest/tier mismatch, undispositioned findings);
confirmed findings route to the maintainer, whose PROMOTE with recorded
dispositions is the gate (verifier-selects-gates-decide). Rejected: hard-block
on confirmed majors; never-worse ratchet. Queued: Track 1 of
`2026-07-24-next-run-queue.md`.