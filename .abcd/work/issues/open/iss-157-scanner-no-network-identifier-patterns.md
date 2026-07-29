---
schema_version: 1
id: "iss-157"
slug: "scanner-no-network-identifier-patterns"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "managed-repo NEXT.md privacy-leak investigation 2026-07-29"
found_at: "internal/adapter/scanner/patterns.go"
---

shared scanner (launch dry-run, lifeboat pack, history capture) has token-shaped secret patterns and identity matchers (username, email, home path, github remote) but no network-identifier patterns — a lifeboat pack or launch bundle carrying a tailnet IP, device hostnames, or firewall posture ships clean. Same root gap as the audit rule; fix belongs in the shared primitive, not per-surface.