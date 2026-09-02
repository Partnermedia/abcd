---
schema_version: 1
id: "iss-2609020127281995"
slug: "ideate-writes-caller-free-text-into-a-committed-tier-with-no"
severity: "major"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/ideate/record.go"
---

ideate writes caller free text into a committed tier with no scanner call — the fifth committed-tier writer the store-before-commit sweep missed. internal/core/ideate/record.go renders the verdict payload and writes it with fsutil.CreateExclusiveIn into .abcd/development/research/, and nothing in internal/core/ideate imports the scanner, so a secret or an identity pasted into the verdict text lands in the durable record in clear while capture, memory, history and intent all route their text through sc.ScanText plus scanner.Redact first. Evidence: internal/core/ideate/record.go (the CreateExclusiveIn call and render), and the absence of any scanner reference in the package. The fix is to route the rendered record through the same shared redactor before the exclusive create, with the stage-two residual check the other stores apply.
