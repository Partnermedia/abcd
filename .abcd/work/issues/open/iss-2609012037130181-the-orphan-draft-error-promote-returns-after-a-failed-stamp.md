---
schema_version: 1
id: "iss-2609012037130181"
slug: "the-orphan-draft-error-promote-returns-after-a-failed-stamp"
severity: "minor"
category: "ux"
source: "agent-finding"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/capture/promote.go"
---

The orphan-draft error Promote returns after a failed stamp (internal/core/capture/promote.go) names the repair as abcd capture promote <iss-N> --intent <itd-N>, without --grounds. The issue route requires --grounds — the CLI refuses with exit 2 (requires --grounds ... nothing written) before it reaches the stamp, and core's requireGrounds refuses the same way — so the documented repair fails on its own text for every orphan, not only the one the advisory produced. TestPromoteStampFailureReportsOrphanAndLinkRepairs passes Grounds to the repair call the message never mentions, which is why it never noticed. The remedy must print a command that runs as printed: the same --intent plus the promotion's own grounds, quoted for a shell, and a test that extracts the remedy from the error and runs it. The reading-item route is unaffected: it refuses grounds, so its remedy is correct without them.
