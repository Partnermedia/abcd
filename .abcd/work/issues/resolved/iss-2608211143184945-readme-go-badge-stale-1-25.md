---
schema_version: 1
id: "iss-2608211143184945"
slug: "readme-go-badge-stale-1-25"
severity: "nitpick"
category: "documentation"
source: "agent-finding"
found_during: "bughunt-b-round-4"
found_at: "README.md"
resolution: "Bumped the README Go badge and alt text from 1.25 to 1.26 (major.minor, matching CHANGELOG's 'Released binaries build on Go 1.26' and avoiding a fourth patch pin no gate covers)."
impact: internal
---

README.md's Go badge advertises 'Go 1.25' while go.mod declares go 1.26.7 and every CI/release pin is 1.26.7; CHANGELOG records the deliberate move to 1.26 for support-window reasons. The badge is the front page's only toolchain statement, and 1.25 is a version the project just retired for security. GOTOOLCHAIN=auto self-corrects a from-source build, so harm is bounded, but the front page contradicts a deliberate security decision. The go-version lockstep test walks .github/workflows only, so no gate covers the badge.