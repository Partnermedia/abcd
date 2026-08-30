---
schema_version: 1
id: "iss-2608300320162381"
slug: "indented-atx-heading-escapes-the-floor"
severity: "major"
category: "security"
source: "impl-review"
found_during: "itd-183 third-round security review, 2026-08-30"
found_at: "internal/core/reading/project.go (redactExcluded, verifyRedaction), internal/core/site/sections.go (headingRe)"
---

An ATX heading preceded by one to three spaces is a heading to every CommonMark renderer but not to the section scan's heading pattern, so both the redactor and the verifier are blind to it and an excluded section travels while the manifest asserts it was refused — the same class as the closing-sequence and case-variant findings closed in the two prior rounds. Widen the section scan to the CommonMark indent, or add a raw-line refusal in the verifier for a non-fenced line matching the indented pattern whose normalised title matches an excluded heading; add both probe shapes to the spelling test.
