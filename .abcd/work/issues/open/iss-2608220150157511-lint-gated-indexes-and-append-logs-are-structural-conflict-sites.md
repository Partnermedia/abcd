---
schema_version: 1
id: "iss-2608220150157511"
slug: "lint-gated-indexes-and-append-logs-are-structural-conflict-sites"
severity: "minor"
category: "process"
source: "user-observation"
found_during: "abcdev-site session close-out 2026-08-22"
found_at: ".abcd/development/brief/06-delivery/03-out-of-scope.md"
---

Merge contention is structural in the shared registers, not the prose: append logs (.abcd/work/DECISIONS.md, 154 changes; the decomposition-calibration corpus, conflicted 2026-08-22) and the lint-gated indexes that index_drift forces every branch to co-edit (brief/06-delivery/03-out-of-scope.md drafts index, conflicted the same day; 04-surfaces/README.md; decisions/adrs/README.md; intents/README.md; release/surface.json). The fix shape for the indexes is already written down: the enumeration command in out-of-scope.md derives the list from the filesystem, so the committed copy could become generated-at-build and the lint compare against the derivation instead of a hand-edited region. Append logs tolerate union merges; the indexes do not