---
schema_version: 1
id: "iss-33"
slug: "ahoy-verb-hygiene"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "internal/core/ahoy/apply.go"
---

ahoy subtree hygiene: install persists unvalidated interactive answers for docs_target and oracle_backend and silently coerces any non-true scan_deep answer to false (internal/core/ahoy/apply.go:184-192); the history-index registration (registerRepo, re-founding lineage) is untested and swallows its own errors (apply.go:280); the wired read-only verbs doctor and dry-run have zero behavioral tests (apply.go:441); ahoy.Status is silent dead scaffolding — zero callers outside tests (apply.go:478). Detector: a wired-verb behavioral test convention (every registered sub-verb has at least one behavioural test against its real store) plus a coverage-plus-caller audit distinguishing loud staging from silent scaffolding per loud-staging. Acceptance corpus: the four instances above.