---
schema_version: 1
id: "iss-2608221126066631"
slug: "guard-process-substitution-redirection-family-allow"
severity: "minor"
category: "observation"
source: "agent-finding"
found_during: "bughunt round 7 merge-gate dual review"
found_at: "internal/core/guard/tokenize.go"
---

The guard tokenizer keeps process-substitution operands out of command position analysis, so a blocker-tier flag glued behind one escapes: 'git push >(cat) --force origin main' and 'git push >$(echo x) --force origin main' both return ALLOW while their plain-redirection spellings block (pre-existing on main; confirmed unchanged by the iss-2608220131352917 &> fix, same redirection family). Within the documented mistake-filter posture, but the &> precedent shows the family is worth sweeping: recognise >(...) / <(...) as redirection-shaped operands and keep the remaining argv in analysis.