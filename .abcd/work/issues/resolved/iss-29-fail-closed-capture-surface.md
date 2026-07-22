---
schema_version: 1
id: "iss-29"
slug: "fail-closed-capture-surface"
severity: "major"
impact: fix
category: "bug"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "internal/surface/cli/cli.go"
resolution: "Fixed the three fail-closed-capture instances behind the unrecognized-input-never-writes detector: (1) typo'd subcommand refused with did-you-mean + no write (suspectedTypoedSubcommand); (2) --json errors JSON-shaped via cli.Run; (3) docs-lint missing/unreadable config surfaces a repo-relative, path-safe error. Systemic PathError-into-json leak split out as iss-76."
---

fail-closed capture surface: a misspelled capture subcommand (e.g. capture resovle) is swallowed as capture text and files a NEW issue instead of erroring (internal/surface/cli/cli.go:455) — a typo becomes a ledger mutation; --json errors emit raw Go text, not JSON (internal/surface/cli/cli.go:165), and abcd docs lint without a config surfaces a raw file error. Detector (per unrecognized-input-never-writes): a surface test convention where every mutating verb has a malformed-input case asserting an error, a did-you-mean for near-misses, and no write occurred, plus a --json error-shape contract test. Acceptance corpus: the three instances above — the detector is proven when it flags all three.