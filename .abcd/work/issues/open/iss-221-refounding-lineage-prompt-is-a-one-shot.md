---
schema_version: 1
id: "iss-221"
slug: "refounding-lineage-prompt-is-a-one-shot"
severity: "minor"
category: "ux"
source: "manual-test"
found_during: "manual-capture"
found_at: "internal/core/ahoy"
---

ahoy install's re-founding lineage prompt is a one-shot: under --yes/non-TTY it prints and silently takes the [y/N] default No, registers the root as new, and a second run reports already_up_to_date without re-asking — no verb links lineage after the fact, leaving two active index entries for the same project (old epoch 85573f…, new 488a0aa… after the attribution rewrite). Non-TTY identity prompts should refuse rather than default, and lineage should be linkable post-registration