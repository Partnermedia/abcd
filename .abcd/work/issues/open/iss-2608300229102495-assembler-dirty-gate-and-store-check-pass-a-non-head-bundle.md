---
schema_version: 1
id: "iss-2608300229102495"
slug: "assembler-dirty-gate-and-store-check-pass-a-non-head-bundle"
severity: "major"
category: "bug"
source: "impl-review"
found_during: "itd-183 adversarial review, 2026-08-30"
found_at: "internal/core/reading/assemble.go (dirtyPaths, configured-store check)"
---

The dirty-tree gate discards a rename or copy's source path and mis-parses a worktree rename, so an included file renamed out of the include set is neither in the bundle nor refused and the manifest names HEAD for a bundle that does not equal HEAD; and the unconfigured-scan refusal checks only that a store key exists, so a store whose directory is absent (or an uncommitted retarget of .abcd/record-lint.json, which sits under the deny and is never in the dirty check) enumerates nothing and reports a clean assembly. Keep rename/copy sources in the dirty set for both status columns; stat every store directory the table depends on; add the lint config path to the dirty check.
