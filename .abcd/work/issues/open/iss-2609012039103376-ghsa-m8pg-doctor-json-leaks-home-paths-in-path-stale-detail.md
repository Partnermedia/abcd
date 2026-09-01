---
schema_version: 1
id: "iss-2609012039103376"
slug: "ghsa-m8pg-doctor-json-leaks-home-paths-in-path-stale-detail"
severity: "minor"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/ahoy/apply.go"
---

GHSA-m8pg-chhv-hxvq (CWE-200, advisory severity low): `ahoy doctor --json` leaks absolute home paths in the `history.path_stale` gap. `internal/core/ahoy/apply.go:auditGaps` composes the gap's Detail from two raw absolute paths — `entry.Path` from `~/.abcd/history/index.json` and `filepath.Abs(cwd)` — while every other path-bearing gap in the package (`detectPathSymlink`, `detectShadowedEntry`, `detectBinDirOnPath`) routes through `displayPath` (`fsutil.RedactHome`) precisely so a gap pasted into an issue carries no username. Text-mode doctor prints counts only; `--json` encodes the whole DoctorReport raw. Reproduced at v0.7.0 (8f68ffb3): register a repo, move it, `ahoy doctor --json` prints both paths in full. The fix must establish that both paths render through `displayPath` and that doctor's JSON output carries no home prefix; a foreign machine's home in `entry.Path` (a different username in a cross-machine registry) is not this advisory's scope.
