---
schema_version: 1
id: "iss-2608220150152332"
slug: "ci-docs-omit-inert-path-classifier"
severity: "minor"
category: "documentation"
source: "agent-finding"
found_during: "bughunt-b-round-7"
found_at: "AGENTS.md"
resolution: "AGENTS.md and CONTRIBUTING.md record the inert-path CI classifier standdown"
impact: internal
resolved_by:
  commit: "554f97f"
---

AGENTS.md and CONTRIBUTING.md describe the CI check job as unconditional but since the inert-path classifier landed a docs/.abcd-only pull request stands the macOS leg, race lane, zizmor, govulncheck and smoke down; the merge-queue run still runs the full set so nothing ships under-verified