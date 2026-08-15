---
schema_version: 1
id: "iss-219"
slug: "dev-install-tests-read-host-path"
severity: "minor"
category: "bug"
source: "manual-test"
found_during: "manual-capture"
found_at: "internal/core/ahoy/dev_install_test.go"
---

ahoy dev-install tests read the host PATH: a stale installed abcd (e.g. the July dev shim in ~/.local/bin) makes the four install-mode tests fail with an unexpected '(shadowed on PATH)' suffix, so make preflight blocks pushes from any machine with abcd installed. The tests should run against an isolated PATH.