---
schema_version: 1
id: "iss-2608270540523859"
slug: "move-contributing-md-and-security-md-from-repo-root-to-githu"
severity: "minor"
category: "tech-debt"
source: "user-observation"
found_during: "config-placement-reorg-2026-08-27"
found_at: "internal/core/site/compose.go"
---

Move CONTRIBUTING.md and SECURITY.md from repo root to .github/ to de-clutter root. NOT a plain git mv: both are load-bearing inputs to abcd's own site build. Required coupled changes: (1) internal/core/site/compose.go footer builds 'blob/main/SECURITY.md' links from a hardcoded root-relative list — update it or the security link silently drops; (2) the contributors/attribution page is config-driven from CONTRIBUTING.md via record_pages.contributors.policy.file (.abcd/config/site or equivalent) — repoint to .github/CONTRIBUTING.md; (3) remove CONTRIBUTING/SECURITY stems from the stray_root_docs allowlist (internal/core/lint/config.go); (4) fix ~6 relative links (README.md, ACKNOWLEDGEMENTS.md, AGENTS.md, CONTRIBUTING.md->SECURITY.md, and site markdown/explorer tests pinning ../../CONTRIBUTING.md and CONTRIBUTING.md#attribution); (5) internal/core/lifeboat/probe.go known-files list names CONTRIBUTING.md; (6) confirm GitHub still surfaces both community-health files from .github/ (it recognises root, docs/, and .github/). Verify with site-render + docs-lint gates. Related to the root-layout / config-placement ADR discussion and the governance/ mirror rename.