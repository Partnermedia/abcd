---
schema_version: 1
id: "iss-34"
slug: "untested-refusal-guards"
severity: "major"
category: "tech-debt"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "internal/core/launch/bundle.go"
---

refusal guards with zero coverage: the launch bundle symlink-dereference and scripts-deny guards (internal/core/launch/bundle.go:339), the memory quotation-budget and licence-detection compliance checks (internal/core/memory/lint.go:225), and the memory ask --file-back write path (internal/core/memory/ask.go:354) are all untested. A guard fails silent: when it regresses the system keeps working and simply stops refusing. Detector (per guards-prove-themselves): a convention that every refusal path ships a test presenting the forbidden input and asserting the rejection, its error shape, and the absence of side effects; a pairing lint between declared invariants and named tests is the promotion path. Acceptance corpus: the five guard paths above.