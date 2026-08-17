---
schema_version: 1
id: "iss-263"
slug: "dispatch-unparseable-issue-reads-as-not-found"
severity: "nitpick"
category: "ux"
source: "impl-review"
found_during: "spc-26 build, ruthless-reviewer note"
found_at: "internal/core/record/record.go"
---

describeIssue discards ListResult.Skipped, so an issue file that exists but is unparseable (broken frontmatter) makes abcd iss-N report 'not found in the issue ledger' — a diagnostic that misleads about a record physically present in a status dir. Surface the skip roster in the fault: 'iss-N present but unreadable at <path>: <parse error>'.