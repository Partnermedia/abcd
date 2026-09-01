---
schema_version: 1
id: "iss-2608300924205748"
slug: "audit-ingest-writes-agent-prose-unredacted"
severity: "minor"
category: "security"
source: "impl-review"
found_during: "itd-181 adversarial security review, 2026-08-30"
found_at: "internal/core/intent/audit.go (ingestedBlock, deadLetterBlock)"
---

The intent-audit ingest path applies no privacy scanner to agent-produced prose: a rationale, narrowing or evidence reference carrying an absolute home path, a hostname or a person's name is written verbatim into the committed shipped intent record, with only the committed-file privacy lint downstream to catch it. Pre-existing for the per-criterion fields; the scope-condition dispositions inherit the convention. intent/redact.go already wraps scanner.Redact and is the canonical primitive to route the verdict's free-text fields through before the write.
