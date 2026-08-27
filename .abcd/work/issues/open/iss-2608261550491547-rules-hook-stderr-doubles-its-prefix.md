---
schema_version: 1
id: "iss-2608261550491547"
slug: "rules-hook-stderr-doubles-its-prefix"
severity: "nitpick"
category: "bug"
source: "impl-review"
found_during: "second-harness adaptor lab review (2026-08-24/26)"
found_at: "internal/surface/cli/cli.go"
---

The rules hook stderr doubles its prefix on load errors: the hook prints 'abcd rules: ...; injecting nothing' around an error that already carries the rules package's own 'rules:' prefix, so the user reads 'abcd rules: rules: ...'. Cosmetic; unwrap or drop one prefix.