---
schema_version: 1
id: "iss-2608211144265046"
slug: "lint-md-line-zero-sentinel-never-emitted"
severity: "nitpick"
category: "documentation"
source: "agent-finding"
found_during: "bughunt-b-round-4"
found_at: "commands/lint.md"
resolution: "Corrected commands/lint.md and prepare-this-repo.md to the real omitempty contract (a line on content-scanning findings only; path-presence findings omit the key), and added the missing identity-positioning rule to the brief 16-lint.md line-carrying list."
impact: internal
---

commands/lint.md documented repolint JSON findings as carrying 'a file and line (line 0 means the finding is not tied to one line)', but repolint.go tags Line as json:line,omitempty so a zero line is omitted entirely, never emitted as 0. Line-less findings are the majority (only docs-currency/privacy-hygiene/identity-positioning set Line). An agent following the doc and testing line==0, or formatting file:line, produces file:undefined. The brief 16-lint.md carries the true contract but named only two line-carrying rules, omitting identity-positioning (rule_positioning.go also sets Line). commands/prepare-this-repo.md implied the same line sentinel.