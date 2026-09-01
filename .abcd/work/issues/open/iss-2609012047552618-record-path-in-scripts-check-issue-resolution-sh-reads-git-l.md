---
schema_version: 1
id: "iss-2609012047552618"
slug: "record-path-in-scripts-check-issue-resolution-sh-reads-git-l"
severity: "minor"
category: "bug"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "scripts/check-issue-resolution.sh"
---

record_path in scripts/check-issue-resolution.sh reads git ls-tree --name-only, which C-quotes a path holding a non-ASCII byte (a record such as iss-999-é.md lists as "…/iss-999-\303\251.md", quotes included), so status_of yields a quoted path whose first component is a quoted-string prefix rather than open, resolved or wontfix, and the RS001 diagnosis added for the stale-branch shape silently regresses to the generic 'does not enter' text for exactly the records whose slug carries an accent. Found by the ruthless review of the hygiene branch. The fix is to list with -c core.quotePath=false (or -z) and to add a case with a non-ASCII slug to check-issue-resolution-cases.sh.
