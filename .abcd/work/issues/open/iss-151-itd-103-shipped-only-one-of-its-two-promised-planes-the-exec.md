---
schema_version: 1
id: "iss-151"
slug: "itd-103-shipped-only-one-of-its-two-promised-planes-the-exec"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
---

itd-103 shipped only ONE of its two promised planes: the execution-time guard plane is fully wired (abcd guard check/hook, ahoy health), but the TEACHING plane — the rules loader injecting matched shell-safety rules before shell-heavy work, per spc-16 'Two planes, one registry' — is not built; there is no guard/safety/hazard domain in internal/core/rules/. All four itd-103 ACs concern the guard plane only, so the fidelity review is 3 MET/1 MWC, but the intent headline ('two planes') is only half delivered. Wire the teaching plane as a rules-loader safety domain sourced from the same bundled hazard registry, or amend the intent's scope.