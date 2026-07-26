---
schema_version: 1
id: "iss-126"
slug: "intent-reconcile-cannot-stamp-impact"
severity: "major"
category: "inconsistency"
source: "agent-finding"
found_during: "iss-117 review (2026-07-24 run queue, burst 1)"
found_at: "internal/core/intent/lifecycle.go"
---

intent Reconcile ships planned intents into shipped/ without stamping impact, so a no-impact seed passes plan and reconcile then trips the intent_impact_valid blocker on a record produced entirely by the tool's own verbs; a plan-or-ship verb must be able to stamp impact (sibling of the iss-117 class, found by its review)