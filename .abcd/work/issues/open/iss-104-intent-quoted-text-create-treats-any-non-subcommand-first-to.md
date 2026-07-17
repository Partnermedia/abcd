---
schema_version: 1
id: "iss-104"
slug: "intent-quoted-text-create-treats-any-non-subcommand-first-to"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
---

intent quoted-text create treats any non-subcommand first token as create text (no suspectedTypoedSubcommand guard like capture): 'abcd intent paln "x"' files a draft instead of erroring; consider porting capture's typo heuristic for symmetry