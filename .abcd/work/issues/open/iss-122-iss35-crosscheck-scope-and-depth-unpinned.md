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