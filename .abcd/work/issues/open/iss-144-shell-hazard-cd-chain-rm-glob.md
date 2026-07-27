---
schema_version: 1
id: "iss-144"
slug: "shell-hazard-cd-chain-rm-glob"
severity: "major"
category: "observation"
source: "user-observation"
found_during: "2026-07-27 shell-safety design"
found_at: "agent shell usage in managed repos"
---

shell-hazard incident, first corpus entry: in another locally managed repo, an agent ran mkdir -p under a scratch dir, then cd, then rm -rf * in one && chain. The && made it accidentally safe (rm only after a successful cd); with separate statements or a ; separator, a failed cd would have left the harness's persistent cwd at the repo root and rm -rf * would have deleted the working tree. Failure class: destruction chained after cd + cwd-relative glob, one character between harmless and catastrophic. Safe successor: destructive commands name an absolute path and never rely on cwd or a glob. This incident and the agent's own post-mortem are the first acceptance-corpus entry for the shell-hazard registry (teach + enforce planes); guard fixture must prove the block fires on exactly this shape.