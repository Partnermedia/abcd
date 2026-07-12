---
schema_version: 1
id: "iss-78"
slug: "launch-dryrun-needs-version-location"
severity: "minor"
category: "observation"
source: "agent-finding"
found_during: "2026-07-12 /abcd:run iss-31"
---

launch dry-run cannot fully pass on this repo even after iss-31's three instances are fixed: .abcd/config/version-location.json is absent, so CheckLockstep reports the contract unreadable and retention refuses on an empty (non-SemVer) version. This is the un-self-installed source-repo state, not a code defect, but it blocks the 'abcd launch --dry-run is green on its own repo' dogfood gate. Resolve by running the version-location decision setup (or documenting the repo as pre-launch). Needed before the aggregate launch dogfood gate (a CI/test that dry-run is green) can be armed.