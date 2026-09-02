---
schema_version: 1
id: "iss-2608300924205748"
slug: "audit-ingest-writes-agent-prose-unredacted"
severity: "minor"
category: "security"
source: "impl-review"
found_during: "itd-181 adversarial security review, 2026-08-30"
found_at: "internal/core/intent/audit.go (ingestedBlock, deadLetterBlock)"
resolution: "The audit ingest routes every free-text field of a verdict — per-criterion rationales, gap-audit claims, scope-condition rationales and narrowings, evidence refs and quotes, and the dead-letter reason — through intent/redact.go before the write, then through the existing termsafe neutraliser. The scanner is built once per ingest and fails closed before any block is composed; identifiers, enum values and hashes are validated shapes and keep the neutraliser alone."
impact: fix
---

The intent-audit ingest path applies no privacy scanner to agent-produced prose: a rationale, narrowing or evidence reference carrying an absolute home path, a hostname or a person's name is written verbatim into the committed shipped intent record, with only the committed-file privacy lint downstream to catch it. Pre-existing for the per-criterion fields; the scope-condition dispositions inherit the convention. intent/redact.go already wraps scanner.Redact and is the canonical primitive to route the verdict's free-text fields through before the write.

## Grounds

- pursued: we expect agent-produced prose to be redacted at the moment it is written into a committed record rather than left to the downstream committed-file lint, so a home path, hostname or person name in a verdict never reaches the shipped intent; a free-text verdict field that still lands verbatim, or a structural field corrupted by the redactor, would show it wrong.
