---
schema_version: 1
id: "iss-2608220150157501"
slug: "pre-policy-tool-author-and-dependabot-commits-need-a-rule"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "abcdev-site-plan investigation 2026-08-21"
found_at: "git history"
---

One early commit carries the tool name in the git author field (pre-dating the Assisted-by policy) and two commits are dependabot; any contributors rendering from git history needs an explicit rule labelling bot and tool author rows separately from the humans who are authors of record, or the pre-policy commit silently appears as a human contributor