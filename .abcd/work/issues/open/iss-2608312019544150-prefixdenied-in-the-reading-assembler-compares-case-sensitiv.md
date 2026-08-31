---
schema_version: 1
id: "iss-2608312019544150"
slug: "prefixdenied-in-the-reading-assembler-compares-case-sensitiv"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
origin: researcher-authored
production_mode: hand-written
---

prefixDenied in the reading assembler compares case-sensitively while segmentDenied folds case, and deny.go's own header comment claims the deny binds every path component case-insensitively, so the comment is wrong about the boundary it documents
