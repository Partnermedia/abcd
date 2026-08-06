---
schema_version: 1
id: "iss-187"
slug: "rules-merge-nil-domains-panic"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "bug-hunt loop round 1"
found_at: "internal/core/rules/rules.go:193"
resolution: "Fixed in bugfix/iss-187-rules-merge-nil-domains-panic: rules.Merge now allocates out.Domains when nil before adding override keys, matching the guard.Merge idiom; regression test TestMergeNilBaseDomainsAddsNewKeys covers it."
impact: fix
---

rules.Merge panics with 'assignment to entry in nil map' (rules.go:193) when the base RuleSet's Domains field is nil and the overlay has at least one domain key. cloneRuleSet (rules.go:524) only allocates out.Domains when the source map is already non-nil, so a base like RuleSet{SchemaVersion: 1} (a Validate-accepted, valid zero-ish value) crashes on merge, contradicting Merge's own doc comment ('New domain keys are added') which promises no such precondition. The sibling loader guard.Merge (internal/core/guard/config.go:77-79) gets this right, explicitly allocating out.Entries when nil before the same kind of loop — Merge is simply missing that guard. Not reachable today: RuleSet.Load's only call site passes Merge(Defaults(), over), and Defaults() always has non-nil Domains, so this is a latent exported-API contract defect rather than a live crash — it becomes live the moment any caller merges onto a non-Defaults() base (e.g. a future multi-tier overlay starting from an empty RuleSet). Reproducing test (recover-guarded): internal/core/rules/, TestMergeNilBaseDomainsAddsNewKeys.