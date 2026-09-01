---
schema_version: 1
id: "iss-2609012036271396"
slug: "ghsa-fh9j-8xmg-m33f-cwe-59-cwe-400-the-issue-ledger-readers"
severity: "major"
category: "security"
source: "agent-finding"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/capture/workflow.go"
---

GHSA-fh9j-8xmg-m33f (CWE-59, CWE-400): the issue-ledger readers scanLedger (workflow.go) and readWithChecksum (roots.go) load every well-formed iss-N-*.md through a bare os.ReadFile, with no O_NOFOLLOW, no O_NONBLOCK, no regular-file check and no byte cap, and List/Status never pass mutationPreamble so the write path's leaf guards never run on a read. In a hostile clone a committed FIFO at a record name hangs capture list and the bare capture status render, an oversize record is read and serialized unbounded, and a committed symlink to an out-of-tree schema-shaped file has its body (any secret span included) serialized into list --json, with no redactor on the read path. The package already owns the hardened reader (reading.go readRecordGuarded, fsutil.ReadGuarded under issueschema.RecordReadLimit) and the reading family uses it everywhere; the issue family was never swept. The fix must route both readers through it so a FIFO, device, symlink or oversize leaf becomes a SkipRecord the surfaces already render, and Promote, transition and commitCapture inherit the guard through readWithChecksum. Sibling unguarded reads outside this package are captured as their own records.
