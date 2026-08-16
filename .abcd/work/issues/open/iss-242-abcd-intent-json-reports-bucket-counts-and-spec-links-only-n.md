---
schema_version: 1
id: "iss-242"
slug: "abcd-intent-json-reports-bucket-counts-and-spec-links-only-n"
severity: "minor"
category: "ux"
source: "agent-finding"
found_during: "intent-planning-prep"
found_at: "internal/surface/cli"
---

abcd intent --json reports bucket counts and spec links only — no per-intent listing. A planning sweep over drafts/ (which intents are plannable vs seeded-placeholder AC, filed when) required shell-grepping 55 files and git history. The verb wants a list mode with per-intent id, title, bucket, ac_state (seeded|real), and filing date; the seeded placeholder is a stable string, so ac_state is cheap.