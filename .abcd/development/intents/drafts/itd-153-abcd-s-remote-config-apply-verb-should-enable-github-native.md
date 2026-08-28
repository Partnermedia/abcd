---
id: itd-153
slug: abcd-s-remote-config-apply-verb-should-enable-github-native
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
promoted_from: iss-2608270636272755
---

# abcd's remote-config apply verb should enable GitHub native secret scanning AND secret-scanning push protection by default on every managed repo (belt-and-braces alongside the CI gitleaks full-history scan): push protection blocks a secret at push time, earlier than CI, and secret scanning covers the default branch continuously. Enabled by hand on [redacted-user]/abcd 2026-08-27 via PATCH /repos/{o}/{r} security_and_analysis (both were disabled). Notes: push protection requires secret scanning enabled first; both are free on public repos; secret_scanning_non_provider_patterns and secret_scanning_validity_checks remain optional further hardening. Homes: the read-only VERIFY side is itd-92 (the doctor now reports these toggles); the APPLY-by-default belongs to the separate adr-44-bound apply intent itd-92 defers to, adjacent to itd-106 (abcd sets up the CI a repo requires); the desired state should be mirrored in the repo-settings.json sibling (iss-2608270512210664).

## Press Release

> _Seeded by promotion from iss-2608270636272755. Expand into the full press-release narrative before planning._

## Why This Matters

Graduated from `iss-2608270636272755`: abcd's remote-config apply verb should enable GitHub native secret scanning AND secret-scanning push protection by default on every managed repo (belt-and-braces alongside the CI gitleaks full-history scan): push protection blocks a secret at push time, earlier than CI, and secret scanning covers the default branch continuously. Enabled by hand on [redacted-user]/abcd 2026-08-27 via PATCH /repos/{o}/{r} security_and_analysis (both were disabled). Notes: push protection requires secret scanning enabled first; both are free on public repos; secret_scanning_non_provider_patterns and secret_scanning_validity_checks remain optional further hardening. Homes: the read-only VERIFY side is itd-92 (the doctor now reports these toggles); the APPLY-by-default belongs to the separate adr-44-bound apply intent itd-92 defers to, adjacent to itd-106 (abcd sets up the CI a repo requires); the desired state should be mirrored in the repo-settings.json sibling (iss-2608270512210664).. Read that issue record for the source observation.

## Acceptance Criteria

> _Required (the itd-1 discipline): add at least one Given-When-Then bullet describing the verifiable bar for "shipped" before this draft can be planned._

## Open Questions

_None recorded yet._

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._
