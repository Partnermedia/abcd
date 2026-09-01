---
schema_version: 1
id: "iss-2609012037119454"
slug: "ghsa-865x-5m7q-qm79-cwe-59-ensureledgerdirs-alloc-go-provisi"
severity: "minor"
category: "security"
source: "agent-finding"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/capture/alloc.go"
---

GHSA-865x-5m7q-qm79 (CWE-59): ensureLedgerDirs (alloc.go) provisions the ledger with os.MkdirAll(filepath.Dir(issuesRoot)), which follows every ancestor, and only then Lstat-checks the leaves through safeMkdirLeaf; resolveRoots (roots.go) derives issuesRoot with filepath.Abs alone and checks no ancestor. A committed mode-120000 symlink at .abcd or .abcd/work therefore redirects the whole store — issues/, the .iss-alloc.lock flock, the three status directories and the record — to the symlink target outside the checkout, while the result still reports the repo-relative path; and because List and Status walk the same issuesRoot, every read serializes an out-of-tree ledger too. The fix must establish that every segment from repoRoot down to Dir(issuesRoot) is a real directory (a per-segment Lstat walk, created one level at a time with os.Mkdir, modelled on memory.memoryDir from the GHSA-72rp fix), placed where every verb passes first so the readers are covered, and keep the leaf guards; a custom IssuesRoot outside the repo is an operator-typed operand and keeps leaf-only behaviour. Test: .abcd/work symlinked out of tree, capture refuses with ErrPathUnsafe and nothing is written in the target.
