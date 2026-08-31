---
schema_version: 1
id: "iss-2608311504353427"
slug: "the-intent-audit-ingest-does-not-neutralise-markdown-link-sy"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "ingesting the itd-187 fidelity verdict"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/intent/audit.go"
---

The intent-audit ingest does not neutralise markdown link syntax in untrusted verdict text, so a verdict that quotes code containing a bracket-then-parenthesis sequence writes a live markdown link into the committed record and record-lint then refuses the whole tree with a links_resolve blocker. Hit for real: an auditor quoting an assembled-input element path of the form items[0](itm-0001) produced a link whose target itm-0001 resolves to nothing, and the gate failed on a record the ingest had just written. The oneLine sanitiser already neutralises the two shapes that matter for spoofing -- newlines, so injected content cannot break out of its line, and HTML comment delimiters, so untrusted text cannot forge a review marker -- and link syntax belongs in the same set for a different reason: not spoofing but a self-inflicted gate failure that no amount of care in the auditor can prevent, because the offending text is a faithful quotation of the code under audit. The repair available today is a hand edit of the rendered line, which is the same accepted repair as the grounds body-lockout.
