---
schema_version: 1
id: "iss-160"
slug: "autonomous-cloud-runs-in-sibling-repos-leaked-harness-attrib"
severity: "minor"
category: "security"
source: "user-observation"
found_during: "manual-capture"
found_at: "autonomous cloud routines / public GitHub artifacts"
---

Autonomous cloud runs in sibling repos leaked harness attribution footers and a live session URL into public GitHub artifacts: the harness auto-appends a 'Generated with' footer plus a session link when a PR or issue is created, overriding the repos' Assisted-by-only attribution policy (commit messages stayed clean; the leak surface was PR bodies and issue comments, plus GitHub's public edit history retaining the pre-scrub revision). Leak shape only - no session ids reproduced here. Two remedies needed: (a) every autonomous routine prompt must ban session URLs and harness footers in public text AND mandate a post-create re-read-and-strip of every PR/issue/comment the loop creates, because the append happens outside the model's own text; (b) abcd should detect the class - session-URL and harness-footer patterns belong with the shared privacy pattern set (iss-154 family / itd-74 banlist territory) so audit and docs-lint flag them in any committed or posted text.