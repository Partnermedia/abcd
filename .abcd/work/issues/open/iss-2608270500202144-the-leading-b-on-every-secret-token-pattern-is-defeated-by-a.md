---
schema_version: 1
id: "iss-2608270500202144"
slug: "the-leading-b-on-every-secret-token-pattern-is-defeated-by-a"
severity: "critical"
category: "security"
source: "agent-finding"
found_during: "github-ledger-dedup-2026-08-27"
found_at: "internal/adapter/scanner/patterns.go"
---

the leading \b on every secret token pattern is defeated by a word-char prefix from a percent-encoded delimiter (%3Dghp_...), so a live PAT/JWT/AWS key embedded in a URL-encoded URL survives history.Capture redaction and is written raw into the committed transcript. GitHub mirror: #370