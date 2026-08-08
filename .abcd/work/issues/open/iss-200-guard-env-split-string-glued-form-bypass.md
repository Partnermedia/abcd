---
schema_version: 1
id: "iss-200"
slug: "guard-env-split-string-glued-form-bypass"
severity: "critical"
category: "bug"
source: "user-observation"
found_during: "bug-hunt loop round 9 (state issue #197), fix attempt for iss-196 blocked at pre-PR review by two independent reviewers"
found_at: "internal/core/guard/match.go:129-150 (skipWrapperArgs), match.go:43 (wrapperValueFlags)"
---

guard's env wrapper table listed -S/--split-string as a value-owning flag for the SEPARATE-TOKEN spelling only ('env -S git push --force origin main'), but GNU env -S/--split-string also accepts the value glued to the flag ('env -Sgit push --force origin main', 'env --split-string=git push --force origin main'), which re-splits identically and executes identically; a fix attempted in bug-hunt loop round 9 (bugfix/env-split-string-guard-bypass) closed only the separate-token form and was BLOCKed at pre-PR review by both an independent correctness reviewer and an independent security reviewer, who each independently reproduced that the glued forms remain a complete, silent bypass of every guard blocker on the live PreToolUse path; suggested remediation (from both reviewers, converging independently): in commandOf/skipWrapperArgs, when wrapper==env and a token matches -S<value> (len>2) or --split-string=<value> (value non-empty), read the first word of <value> as command position rather than treating the whole token as a consumed wrapper flag — no nested re-tokenizing required, contained to the same two functions