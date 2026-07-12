---
schema_version: 1
id: "iss-70"
slug: "lint-receipt-gate-detector-binding"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "clean-slate-sweep"
found_at: "internal/core/lint/lint.go"
---

record-lint receipt-gate hardening (VSA gate): checkReceiptGate does not bind a receipt to the gate it attests — one genuine PROMOTE receipt satisfies EVERY required gate by copying it to each path (lint.go:601, C16); ArmReceiptGate keeps committer-editable required_gates when the caller supplies an empty list, fail-open arming (config.go:136, P9); workflowStepNames captures nested name keys (with: name:) as gate step names (lint.go:857, P4). NOTE C16 adds policy.detector to the receipt JSON schema — a record-lint CONTRACT change, surface for maintainer sign-off before landing. Detector: mismatched policy.detector BLOCKED per gate; empty arming fails closed; only step-indent name captured. Corpus: C16, P9, P4.