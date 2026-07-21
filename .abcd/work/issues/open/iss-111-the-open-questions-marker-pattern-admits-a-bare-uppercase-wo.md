---
schema_version: 1
id: "iss-111"
slug: "the-open-questions-marker-pattern-admits-a-bare-uppercase-wo"
severity: "minor"
category: "tech-debt"
source: "impl-review"
found_during: "itd-95 P1 review"
found_at: "internal/core/lifeboat/sources_conventions.go"
---

The open-questions marker pattern admits a bare uppercase word followed by whitespace, so prose that merely mentions a marker matches. Measured against this repository the adapter reports 42 markers across 11 files at medium confidence, and every one is documentation about markers rather than a work marker. The precision cost lands on repos that document their own conventions. Options: require the trailing colon or parenthesis, or drop the bare-word alternative for the ambiguous markers only. spc-12 fixes the current pattern, so this is a design revisit.