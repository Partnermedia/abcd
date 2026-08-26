---
schema_version: 1
id: "iss-2608261132597732"
slug: "inrepo-conflation-privacy-scan-silent-zero-coverage"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "bughunt-round-8"
found_at: "internal/gitutil/repo.go:155"
---

gitutil.InRepo collapses repo-shaped-but-git-unanswerable into not-a-repo, so TrackedFiles returns nil and the privacy-hygiene lint reports a clean pass after scanning zero files