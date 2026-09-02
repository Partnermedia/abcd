---
schema_version: 1
id: "iss-2609020716560392"
slug: "ci-yml-s-concurrency-group-cancels-an-in-progress-run-for-ev"
severity: "major"
category: "bug"
source: "agent-observation"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: ".github/workflows/ci.yml"
---

ci.yml's concurrency group cancels an in-progress run for every ref except the default branch, and a merge-queue ref (gh-readonly-queue/main/pr-N-<sha>) is never the default branch, so a merge-group run is cancellable; GitHub's queue treats a cancelled required check as a failure, removes the entry and disarms auto-merge with no notice on the pull request. One PR was removed four times in one night (twice on the 15-minute job budget, twice on a cleanup flake), each time silently. Fix shape: exempt the queue refs from cancel-in-progress (attribution.yml already keys on run_id), and consider disabling the ruleset's 'only merge non-failing pull requests' so one flake waits for the group behind it instead of ejecting; a comment on the PR when the queue removes it would make the remaining drops loud.
