---
schema_version: 1
id: "iss-2609012023256534"
slug: "rs001-in-scripts-check-issue-resolution-sh-cannot-tell-a-sta"
severity: "minor"
category: "ux"
source: "agent-observation"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "scripts/check-issue-resolution.sh"
---

RS001 in scripts/check-issue-resolution.sh cannot tell a stale branch from a forgotten resolution, and its message diagnoses the wrong one. On a branch 235 commits behind origin/main the pre-push preflight emitted 84 RS001 violations, every one instructing the reader to resolve an issue that already sat in a terminal folder at the base: the branch's own commits had been squash- or rebase-merged, so origin/main held the moved records, the two-dot diff of the two trees showed no record entering resolved/ or wontfix/, and the trailers went unsatisfied within the range. No line of the output said the branch was behind, and the one remedy that works, a rebase, appeared nowhere. The gate should distinguish the shapes it can prove: a record already terminal at the base ref (with how far behind the head is and which base-side commit placed it there), a record absent from the head tree while the base holds it, and an id with no record anywhere, and name a rebase as the likely remedy where the evidence points to one. Carried from the session handover into autonomous-run-2026-09-01.
