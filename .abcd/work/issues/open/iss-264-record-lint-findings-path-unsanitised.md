---
schema_version: 1
id: "iss-264"
slug: "record-lint-findings-path-unsanitised"
severity: "minor"
category: "security"
source: "impl-review"
found_during: "spc-27 build, security-reviewer note"
found_at: "cmd/record-lint/main.go"
---

cmd/record-lint prints Finding.File and Finding.Message raw — no termsafe.Sanitize and no path scrub on the findings path (its scrubPaths covers only the two error exits), so any rule echoing repo content lets a hostile clone inject terminal escapes into CI logs, and any finding message embedding an os error can leak absolute paths. The abcd CLI surface sanitizes at its findings renderer; the standalone binary does not. Found by the spc-27 security review, which also caught (and the change fixed) the one in-diff instance.