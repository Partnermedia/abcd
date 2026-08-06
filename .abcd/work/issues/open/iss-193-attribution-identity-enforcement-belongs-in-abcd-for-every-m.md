---
schema_version: 1
id: "iss-193"
slug: "attribution-identity-enforcement-belongs-in-abcd-for-every-m"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
---

Attribution-identity enforcement belongs in abcd for every managed repo: today's incident (2026-08-06, this repo's own history rewrite) showed autonomous harness runs committing as an AI author identity and squash-merges accreting AI co-author trailers, falsifying the contributor graph the Assisted-by convention exists to protect. abcd should prevent the class, not each repo by hand: ahoy install pins the maintainer's git identity (author/committer) into the managed repo config, and a managed gate (pre-commit/pre-push and CI) rejects commits whose author/committer is outside the repo's identity allowlist or that carry an AI Co-authored-by trailer, with Assisted-by as the one sanctioned disclosure channel. Likely intent-scale (user-facing capability across managed repos); relates to iss-84 (managed pre-commit gates), iss-85 (managed attribution config) and iss-119 (Assisted-by declared but unenforced) — reconcile with all three rather than duplicating them.