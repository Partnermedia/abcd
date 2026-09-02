---
schema_version: 1
id: "iss-2609020239008775"
slug: "the-memory-store-s-write-time-leaf-walk-judges-values-only-s"
severity: "minor"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/memory/redact.go"
---

the memory store's write-time leaf walk judges values only, so a secret can enter as a map KEY. storeRedactor.redactLeaves sanitises every string leaf a write introduces but never looks at the keys it walks, and its comment asserts 'Keys are schema-fixed and never walked' — false for the page source: block, where validateSourceBlock rejects no unknown key and checkBlockKey admits any identifier-shaped key, a token like ghp_ followed by alphanumerics included. A host distiller's --pages-json frontmatter can therefore carry a credential as a YAML key straight into a committed page, and only the MR001 residue lint catches it, on read, after the write has already happened. The fix must judge an introduced key with the same scanner its values go through (refuse, never rename) and correct the comment.
