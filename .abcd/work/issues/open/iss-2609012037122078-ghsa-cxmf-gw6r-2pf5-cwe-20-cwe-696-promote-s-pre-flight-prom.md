---
schema_version: 1
id: "iss-2609012037122078"
slug: "ghsa-cxmf-gw6r-2pf5-cwe-20-cwe-696-promote-s-pre-flight-prom"
severity: "minor"
category: "security"
source: "agent-finding"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/capture/promote.go"
---

GHSA-cxmf-gw6r-2pf5 (CWE-20, CWE-696): Promote's pre-flight (promote.go) reads the source issue and checks only promoted_to before intent.CreateDraft; validateStrict and validateInvariants run only inside the stamp closure, under the lock, after the draft is minted. A record that is malformed by the package's own invariants — a frontmatter slug that is kebab-valid but disagrees with the filename slug, or an enum value outside the schema, both of which list and record-lint already reject — is therefore minted from anyway: a draft carrying the attacker-chosen slug and a promoted_from back-edge lands in the committed intents/drafts store, the stamp fails on the invariant, the orphan-draft remedy fails on the identical invariant, and every rerun mints one more orphan. Promote's own doc comment (written by ab79b54d for the grounds-append limb) already states that every refusal decidable from the pre-flight bytes is raised before the mint. The fix must run validateStrict and validateInvariants in the pre-flight, keeping the status findIssue returns, so an invalid record produces an error and nothing else; the stamp-step checks stay for what only the under-lock bytes can show. promoteReadingItem already validates before minting.
