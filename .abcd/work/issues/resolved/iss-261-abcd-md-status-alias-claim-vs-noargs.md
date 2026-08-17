---
schema_version: 1
id: "iss-261"
slug: "abcd-md-status-alias-claim-vs-noargs"
severity: "minor"
category: "drift"
source: "impl-review"
found_during: "spc-26 build"
found_at: "commands/abcd.md"
resolution: "commands/abcd.md's stale status-alias claim replaced by the record-id dispatch section in the spc-26 sweep; the CLI gained the id positional, not the alias"
impact: fix
resolved_by:
  intent: "itd-121"
  spec: "spc-26"
---

commands/abcd.md claims 'status is a positional alias for the same bare render' and its argument-hint reads [status], but the binary's root command is Args: cobra.NoArgs — abcd status errors 'unknown command' with exit 2. Either wire the alias or drop the claim; found while building the abcd <id> root positional (spc-26), which gates its positional on the record-id shape and leaves other positionals on the unknown-command path.