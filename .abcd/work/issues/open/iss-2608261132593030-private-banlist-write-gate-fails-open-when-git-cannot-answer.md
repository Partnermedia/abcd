---
schema_version: 1
id: "iss-2608261132593030"
slug: "private-banlist-write-gate-fails-open-when-git-cannot-answer"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "bughunt-round-8"
found_at: "internal/core/banlist/private.go:610"
---

requireIgnoredStore and the ListPrivate warning skip whenever InRepo is false, so with git unanswerable in a repo-shaped tree banlist add --private writes the secret-pattern store unignored with exit 0