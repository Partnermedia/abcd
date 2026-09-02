---
schema_version: 1
id: "iss-2609020716570699"
slug: "nothing-tells-an-agent-that-a-record-it-is-about-to-fix-has"
severity: "major"
category: "process"
source: "agent-observation"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: ".abcd/work/issues"
---

Nothing tells an agent that a record it is about to fix has been claimed or resolved by another session until the resolution gate refuses the push. In one night a peer session re-fixed two issues a paused branch also fixed, and two of its open PRs duplicate merged work. The claim signal that worked in every published multi-agent run is the repository itself: a claim written into the open record (claimed_by: account and harness, branch) and pushed alone through the queue before any fix starts, so a losing race is a push rejection; plus a duplicate-guard required check that fails a PR whose Resolves trailer names a record already resolved on origin/main; plus capture resolve refusing a record that is already terminal on the fetched origin/main. Folder membership is already the status signal, so the claim extends the one canonical primitive rather than adding a lock file that rots. Refines iss-2608220750029993.
