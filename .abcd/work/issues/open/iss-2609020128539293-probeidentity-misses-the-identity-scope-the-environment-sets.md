---
schema_version: 1
id: "iss-2609020128539293"
slug: "probeidentity-misses-the-identity-scope-the-environment-sets"
severity: "major"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/adapter/scanner/identity.go"
---

ProbeIdentity misses the identity scope the environment sets, so a CI or direnv persona is committed in clear. The probe unions every scope git config --get-all reports, but GIT_AUTHOR_NAME, GIT_AUTHOR_EMAIL, GIT_COMMITTER_NAME and GIT_COMMITTER_EMAIL are an identity source git config never lists: they outrank every config file when a commit is written, and CI runners, direnv profiles and rebase wrappers set them routinely. An identity that authors the caller's commits but is absent from the matcher set is stored verbatim by every write-time redactor (capture, memory, history, intent), which is the same class the three GHSA identity advisories reported for a repo-local persona. Evidence: scanner.ProbeIdentity in internal/adapter/scanner/identity.go reads only git config. The fix is to fold those four environment values into the Other* sets, read once in ProbeIdentity and subject to the same guards the config values are, proved by a test that sets them in-process and asserts the address is redacted.
