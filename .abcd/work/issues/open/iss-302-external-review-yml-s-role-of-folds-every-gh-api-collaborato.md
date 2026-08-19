---
schema_version: 1
id: "iss-302"
slug: "external-review-yml-s-role-of-folds-every-gh-api-collaborato"
severity: "minor"
category: "security"
source: "agent-finding"
found_during: "bughunt-round-1"
found_at: ".github/workflows/external-review.yml"
---

external-review.yml's role_of() folds every 'gh api collaborators/<u>/permission' non-zero exit into the string 'none' via '2>/dev/null || echo none', indistinguishable from a legitimate no-role answer, so a transient/permission failure silently misclassifies authors and drops approvals; the job grants only contents:read + pull-requests:read while that endpoint needs push access (or an App members permission), so it likely 403s for every lookup and wedges the gate red for all human PRs with no diagnostic. Not fixed autonomously: required-check workflow, and the permission behaviour needs a real Actions run to confirm
## Evidence

- `.github/workflows/external-review.yml` — `role_of() { gh api
  "repos/$REPO/collaborators/$1/permission" --jq .role_name 2>/dev/null || echo none; }`
  folds every non-zero exit (403 / 404 / rate-limit / network / jq error) into `none`,
  indistinguishable from a legitimate no-role answer, so a transient/permission failure
  misclassifies the author as external or drops a real approval.
- The job grants `contents: read` + `pull-requests: read` only; the
  `collaborators/{u}/permission` endpoint needs push access (or, for an App token, the org
  `members` permission), so it likely 403s for every lookup — silently wedging the gate red
  for all human-authored PRs with no diagnostic. Landed today (iss-281).

## Adversarial review

CONFIRMED (substantive, with a verification caveat) by an independent refuter: the masking is
certain from the code; the "breaks every PR" severity rests on the token genuinely lacking
endpoint access, which needs a real Actions run to confirm. Proposed fix (NOT applied this
round): fail loud on a genuine API error, and either add the endpoint's permission or derive
role from `author_association` (which the existing token can read). Left open: required-check
workflow, unverifiable here.
