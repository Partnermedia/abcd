---
schema_version: 1
id: "iss-90"
slug: "docs-lint-needs-plugin-root-in-target"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "2026-07-13 B1 dogfood: prepare-this-repo audit of Manuscripts"
found_at: "internal/core/lint"
---

abcd docs lint is unusable inside a target repo without a plugin root: with ABCD_PLUGIN_ROOT/CLAUDE_PLUGIN_ROOT unset it reports plugin_root_status missing and cannot run, so the read-only doc-currency check the prepare-this-repo skill implies is unreachable when onboarding an external repo. Detector: docs lint runs against any repo docs/ given only a cwd (no plugin-root prerequisite), or fails loudly with a one-line remedy per loud-staging. Acceptance: abcd docs lint in a repo with docs/ but no plugin root.