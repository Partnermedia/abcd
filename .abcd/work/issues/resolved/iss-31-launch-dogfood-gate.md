---
schema_version: 1
id: "iss-31"
slug: "launch-dogfood-gate"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "internal/core/launch"
resolution: "Addressed all three named acceptance-corpus instances of the launch dogfood gate. FIXED: (1) identity scanner /dev/null false positive — a machine username colliding with a system directory (HomeUser=dev) no longer hard-fails on /dev/null; suppressed only as a top-level absolute system-path segment, genuine leaks still caught (security-reviewed PASS, home_path_self backstop verified); (3) globRegexpCache data race — guarded by sync.RWMutex, -race clean. STALE: (2) 'payload omits skills/' — no skills/ dir exists; skills were reclassified to commands/abcd/ on 2026-07-11 and ship via the commands include. The aggregate 'launch --dry-run green on this repo' gate is separately blocked by version-location.json absence (filed iss-78) and a broader payload-completeness question re agents/ + hooks/ (filed iss-77). Live dry-run hard_fails: 2 -> 0."
---

launch dry-run cannot pass on its own repo: the identity scanner hard-fails on /dev/null as a local-username false positive (internal/adapter/scanner/identity.go:230) so the launch gates can never pass here; the launch payload omits skills/ and every file its own fallback instructions depend on (.abcd/config/launch-payload.json); globRegexpCache in the bundle resolver is an unsynchronized package-level map — a data race if the transport-agnostic core is ever driven concurrently (internal/core/launch/bundle.go:559). Detector: a dogfood gate — CI runs abcd launch --dry-run against this repo and expects a pass, plus -race coverage of concurrent resolver use. Acceptance corpus: the three instances above; the dogfood gate fails on all of them today.