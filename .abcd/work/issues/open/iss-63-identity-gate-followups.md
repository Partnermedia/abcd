---
schema_version: 1
id: "iss-63"
slug: "identity-gate-followups"
severity: "minor"
category: "future-work-seed"
source: "agent-finding"
found_during: "iss-62-security-review"
---

iss-62 identity-gate follow-ups from security review (advisory, non-blocking): (1) the pre-commit shell guard mis-parses a pinned name containing an escaped double-quote (sed truncates it), blocking even a correctly-configured identity — fail-closed but a usability bug; WritePin also emits that escaped form. Consider delegating to 'abcd ahoy identity-check' when on PATH, or a more robust JSON parse. (2) A programmatic caller passing ApprovedCategories{ConfigChange:true} (without --yes) still auto-pins the current identity; deliberate per-category approval, flagged as a conscious choice.