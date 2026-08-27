---
schema_version: 1
id: "iss-2608271711539855"
slug: "docs-lint-harness-rules-miss-host-tokens-in-paths-and-env-vars"
severity: "minor"
category: "observation"
source: "agent-finding"
found_during: "structural consistency review of .abcd/ and docs/ (2026-08-27)"
found_at: "docs/how-to/install.md"
---

the docs-lint harness rules miss bare host-name tokens in paths and env vars: docs/how-to/install.md names the host harness twice in hand-written prose — a .claude-plugin/ directory link and the $CLAUDE_PLUGIN_DATA cache variable — with no docs-lint finding and no sanctioned allow escape. Fix the detector first: widen the harness/claude-code pattern in .abcd/docs-lint.json so a product-name token inside a path or env var is caught, and watch it fire on both install.md sites before deciding the prose remedy. Whether install pages get a per-line allow marker is already an open decision on iss-216 — do not pre-empt it here.