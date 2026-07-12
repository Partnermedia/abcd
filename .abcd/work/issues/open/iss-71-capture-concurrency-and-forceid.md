---
schema_version: 1
id: "iss-71"
slug: "capture-concurrency-and-forceid"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "clean-slate-sweep"
found_at: "internal/core/capture/workflow.go"
---

capture ledger concurrency + unvalidated ForceID: transition() (resolve/wontfix) takes no flock (only reservePath does) so two concurrent conflicting transitions on the same iss split-brain it across two status dirs — permanent corruption, findIssue then always ErrDuplicateIssueID (workflow.go:129, C4); reservePath builds a path from and O_EXCL-creates a placeholder for ForceID BEFORE validateStrict checks it against the iss-id regex, a pre-validation filesystem touch/traversal (alloc.go:91, P13). Detector: hold the ledger flock across findIssue+commitTransition; reject a non-matching ForceID before any filesystem op. Corpus: C4, P13.