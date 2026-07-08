---
schema_version: 1
id: "iss-32"
slug: "atomic-write-consolidation"
severity: "major"
category: "tech-debt"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "internal/fsutil/fsutil.go"
---

atomic-write consolidation: four independent temp-file+rename implementations with divergent durability — internal/fsutil.WriteFileAtomic (file fsync + parent-dir fsync), ahoy marker.go writeFileAtomic (no parent fsync, mode-preserving), capture roots.go writeFileAtomic (no parent fsync despite a doc comment claiming durability), memory writer.go durableWrite (parent fsync, O_EXCL naming); isRealDir is likewise duplicated (ahoy store.go vs fsutil). internal/fsutil itself has zero tests. Detector (per one-canonical-primitive): a lint flagging private redefinitions of the canonical names (writeFileAtomic, isRealDir) outside internal/fsutil, plus a fsutil crash-safety test suite; consolidation then drains behind the armed lint. Acceptance corpus: the three non-canonical copies and the duplicate isRealDir.