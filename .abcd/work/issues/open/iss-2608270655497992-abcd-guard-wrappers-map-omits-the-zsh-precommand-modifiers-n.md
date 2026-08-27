---
schema_version: 1
id: "iss-2608270655497992"
slug: "abcd-guard-wrappers-map-omits-the-zsh-precommand-modifiers-n"
severity: "minor"
category: "security"
source: "agent-finding"
found_during: "security-cut-agent-flagged-siblings-2026-08-27"
found_at: "internal/core/guard/match.go"
---

abcd guard wrappers map omits the zsh precommand modifiers noglob and nocorrect, which behave like the command/exec/nohup wrappers already handled, so a blocker after noglob/nocorrect stays out of command position. Wrapper-family sibling of the closed coproc/#318 keyword gap. Flagged by the #318 fix agent.