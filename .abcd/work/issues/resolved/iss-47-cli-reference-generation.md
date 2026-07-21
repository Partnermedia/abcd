---
schema_version: 1
id: "iss-47"
slug: "cli-reference-generation"
severity: "minor"
category: "future-work-seed"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "docs/reference/cli"
resolution: "Generated CLI reference (docs/reference/cli/commands.md) walked from the Cobra tree by cli.GenerateReference, with a go-test drift gate; hand-rolled, no new dependency."
---

generated CLI reference with a drift test: generate the verb reference pages from the cobra command tree and wire a go test that fails when the committed pages drift from the tree — replacing the phantom freshness gate that docs/reference/cli/README.md currently claims (see phantom-enforcement-claims). A sibling project runs exactly this shape: generated CLI docs gated by a drift check inside the normal test run. Acceptance corpus: the reference pages that do not exist today; the drift test is proven by regenerating after any verb change.