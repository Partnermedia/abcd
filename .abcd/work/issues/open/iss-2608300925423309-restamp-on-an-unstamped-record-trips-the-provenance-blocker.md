---
schema_version: 1
id: "iss-2608300925423309"
slug: "restamp-on-an-unstamped-record-trips-the-provenance-blocker"
severity: "major"
category: "bug"
source: "impl-review"
found_during: "itd-178 adversarial security review, 2026-08-30"
found_at: "internal/core/capture/workflow.go (productionModeExtras, transition), internal/core/lint/provenance.go"
---

capture resolve --production-mode and capture wontfix --production-mode on a record that predates disclosure (every existing issue) append a lone production_mode with no origin, and the branch's own record_provenance blocker then reports the record the command just wrote as a state no command produced; the restamp test covers only an already-stamped fixture. Refuse a restamp when the record carries no origin (this record predates disclosure; nothing to restamp), state it in spc-56, and test on an unstamped fixture. Low, noted: resolveProductionMode reads the pin on paths that mint nothing, so a malformed pin refuses unrelated operations — fail-closed by design.
