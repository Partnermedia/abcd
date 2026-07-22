---
schema_version: 1
id: "iss-32"
slug: "atomic-write-consolidation"
severity: "major"
impact: fix
category: "tech-debt"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "internal/fsutil/fsutil.go"
resolution: "Consolidated the four named atomic-write/real-dir copies onto internal/fsutil (one-canonical-primitive), armed behind TestNoNonCanonicalAtomicWritePrimitives + a new fsutil crash-safety suite (fsutil had zero tests). Deleted ahoy writeFileAtomic+isRealDir, capture writeFileAtomic; memory durableWrite -> thin writeStringAtomic adapter over fsutil.WriteFileAtomic. Added fsutil.WriteFileAtomicPreserveMode. ruthless PROMOTE, security PASS (symlink-refusal preserved, O_EXCL->CreateTemp stronger, no mode downgrade). A 5th inline copy (memory/ingest.go storeOriginal), outside this issue's named corpus, is filed as iss-79."
---

atomic-write consolidation: four independent temp-file+rename implementations with divergent durability — internal/fsutil.WriteFileAtomic (file fsync + parent-dir fsync), ahoy marker.go writeFileAtomic (no parent fsync, mode-preserving), capture roots.go writeFileAtomic (no parent fsync despite a doc comment claiming durability), memory writer.go durableWrite (parent fsync, O_EXCL naming); isRealDir is likewise duplicated (ahoy store.go vs fsutil). internal/fsutil itself has zero tests. Detector (per one-canonical-primitive): a lint flagging private redefinitions of the canonical names (writeFileAtomic, isRealDir) outside internal/fsutil, plus a fsutil crash-safety test suite; consolidation then drains behind the armed lint. Acceptance corpus: the three non-canonical copies and the duplicate isRealDir.