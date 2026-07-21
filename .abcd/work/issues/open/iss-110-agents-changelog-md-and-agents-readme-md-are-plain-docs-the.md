---
schema_version: 1
id: "iss-110"
slug: "agents-changelog-md-and-agents-readme-md-are-plain-docs-the"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
---

agents/CHANGELOG.md and agents/README.md are plain docs (the itd-5 prompt-version log; a readme) but the plugin agent-loader globs agents/*.md, so both are mis-registered as agents (abcd:CHANGELOG, abcd:README appear in the harness agent list). They have no agent frontmatter and are not invokable workers. Fix: either move these docs out of agents/ (e.g. to .abcd/development/ or a docs path) or make the loader skip non-agent files (require agent frontmatter). Surfaced by the derived-changelog plan adversarial review, which had assumed the abcd:CHANGELOG slot was free for a new composer agent.