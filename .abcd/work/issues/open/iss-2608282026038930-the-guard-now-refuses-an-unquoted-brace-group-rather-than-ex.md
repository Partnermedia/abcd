---
schema_version: 1
id: "iss-2608282026038930"
slug: "the-guard-now-refuses-an-unquoted-brace-group-rather-than-ex"
severity: "minor"
category: "ux"
source: "user-observation"
found_during: "itd-156 adversarial review follow-up"
found_at: "internal/core/guard/tokenize.go"
---

The guard now refuses an unquoted brace group rather than expanding it (itd-156/spc-49 scoped the expander out), so every ordinary shell brace an agent writes is blocked: mkdir -p foo/{a,b}, cp x{,.bak}, rm -rf dir{1..9}. That is the intended fail-closed posture — a word whose argv the guard cannot compute is one it cannot check — but it is a real usability cost on the PreToolUse path, paid on every command that uses a shell convenience nobody meant as a hazard. The scoped follow-up is the bounded expander the intent already names: enumerate the Cartesian product of a group's alternatives (nested groups and {a..z} ranges included) under a hard cap on the number of words produced, check each expansion against the registry, and refuse only when the cap is hit or an expansion matches a blocker. Detector: a corpus of everyday brace commands whose verdicts should be allow; acceptance: mkdir -p foo/{a,b} allows while git push {--force,} origin main still blocks.