---
schema_version: 1
id: "iss-131"
slug: "receipt-gate-hardening-nitpicks"
severity: "nitpick"
category: "tech-debt"
source: "impl-review"
found_during: "iss-122 reviews (2026-07-24 run queue, burst 6)"
found_at: "internal/core/lint/lint.go"
---

gate-hardening nitpicks from the iss-122 reviews, none blocking: (1) the manifest read in receipt_gate uses unbounded os.ReadFile where the sibling convention is fsutil.ReadGuarded (trusted committed file, self-DoS only — but it is the unhardened sibling the bughunt spine says to sweep); (2) receipt.example.json's manifestHash is not test-pinned against manifest.json so a manifest edit silently stales the example; (3) receipt parsing tolerates duplicate JSON keys (Go last-wins) — an audit-legibility gap under the attestation trust model, not an escalation