---
schema_version: 1
id: "iss-2608221457227161"
slug: "the-guard-tokenizer-does-not-perform-brace-expansion-so-a-fl"
severity: "nitpick"
category: "bug"
source: "agent-finding"
found_during: "all-dimensions bug-hunt round 7"
found_at: "internal/core/guard/tokenize.go"
promoted_to: itd-156
---

The guard tokenizer does not perform brace expansion, so a flag wrapped in a single-element brace group with an empty alternative (git push {--force,} origin main) expands in bash to byte-identical argv --force yet the guard reads the literal token {--force,} and allows it — a Tier-1 blocker miss of the same mutate-the-flag-token shape as the round-6 redirection fix. Distinct from the $'...' quoting gap (this is expansion, not quoting, and breaks no written invariant); recorded for a scoped follow-up because a correct bounded brace-expander is larger than this round's scope.