---
schema_version: 1
id: "iss-2608211849467013"
slug: "guard-redirect-tokenizer-bypass"
severity: "major"
category: "security"
source: "user-observation"
found_during: "bughunt-round-6"
found_at: "internal/core/guard/tokenize.go:187"
resolution: "guard tokenizer now recognises redirection operators; glued and leading forms block again"
impact: fix
resolved_by:
  commit: "021ba3f"
---

guard tokenizer does not recognise redirection operators, so gluing >file onto a token silently downgrades a Tier-1 block to allow