---
schema_version: 1
id: "iss-2609020347585681"
slug: "a-here-document-body-whose-redirection-line-ends-in-a-list-o"
severity: "major"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/guard/tokenize.go"
---

A here-document body whose redirection line ends in a list operator reaches command position instead of being read as data, so an apostrophe in an ordinary document becomes ErrUnparsableCommand and the pre-tool-use hook fails OPEN and runs the whole line unguarded (CWE-636, the GHSA-5wx3-2c86-fjpx fail-open class, pre-existing at v0.7.0). Evidence symbol: tokenize's newline branch in internal/core/guard/tokenize.go, which withheld skipHeredocBodies while lastList was set. bash collects here-document bodies at the end of the PHYSICAL line, not at the end of the command list: probed on bash 3.2, `cat <<EOF &&` / `don't` / `EOF` / `echo ok` runs with the body as data and prints ok. The fix must start the body on the line after the redirection regardless of a trailing `&&`, `||` or `|`, and must leave ErrUnparsableCommand for an unterminated quote in COMMAND text only, which no shell runs either.
