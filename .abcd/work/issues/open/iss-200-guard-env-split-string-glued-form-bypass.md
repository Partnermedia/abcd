---
schema_version: 1
id: "iss-200"
slug: "guard-env-split-string-glued-form-bypass"
severity: "critical"
category: "bug"
source: "agent-finding"
found_during: "bug-hunt loop round 9 (state issue #197), fix attempt for iss-196 blocked at pre-PR review by two independent reviewers"
found_at: "internal/core/guard/match.go:129-150 (skipWrapperArgs), match.go:43 (wrapperValueFlags)"
---

guard's env wrapper table listed -S/--split-string as a value-owning flag for the SEPARATE-TOKEN spelling only ('env -S git push --force origin main'), but GNU env -S/--split-string also accepts the value glued to the flag ('env -Sgit push --force origin main', 'env --split-string=git push --force origin main'), which re-splits identically and executes identically; a fix attempted in bug-hunt loop round 9 (bugfix/env-split-string-guard-bypass) closed only the separate-token form and was BLOCKed at pre-PR review by both an independent correctness reviewer and an independent security reviewer, who each independently reproduced that the glued forms remain a complete, silent bypass of every guard blocker on the live PreToolUse path; suggested remediation (from both reviewers, converging independently): in commandOf/skipWrapperArgs, treat -S/--split-string as a value-owning env flag in all three spellings — the exact separate-token form (-S as its own token, value in the following token), the glued form (-S<value>, len>2), and the --split-string=<value> form (value non-empty) — and in every case read the first word of <value> as command position, rather than either treating the whole glued token as a consumed wrapper flag or skipping the following token outright in the separate-token case — no nested re-tokenizing required, contained to the same two functions