---
schema_version: 1
id: "iss-2609020231145566"
slug: "the-armed-gitleaks-path-detects-and-verifies-at-different-sc"
severity: "major"
category: "bug"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/history/history.go"
---

The armed gitleaks path detects and verifies at different scopes, so a benign recurrence of one fragment refuses every capture for good. toFindings locates the WHOLE reported value in the transcript and emits one finding per line that value crosses, but history.unsealedAugmented verifies each FRAGMENT by presence anywhere in the redacted text: a transcript that also quotes a standalone PEM end-of-key marker line, outside any key, keeps that line unsealed (detection never emitted a finding for it) while verification finds it, so Capture returns RedactionResidualError on every drain, the session is never stored, and its unredacted .raw copy stays staged indefinitely. Evidence: internal/adapter/gitleaks/gitleaks.go toFindings (the whole-needle strings.Index loop and its per-fragment split) and internal/core/history/history.go unsealedAugmented (the strings.Contains-anywhere check). The fix must give detection and verification one scope, so that any byte sequence verification looks for is a sequence detection located and sealed, while keeping the fail-closed contract TestAugmentUnlocatableReportFailsClosed pins.
