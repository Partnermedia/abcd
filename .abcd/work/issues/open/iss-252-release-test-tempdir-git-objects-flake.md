---
schema_version: 1
id: "iss-252"
slug: "release-test-tempdir-git-objects-flake"
severity: "minor"
category: "bug"
source: "agent-observation"
found_during: "v0.5.0 release verify (run 31960563726)"
found_at: "internal/core/release/emit_test.go"
---

TestEmitRefuses/an_added_record_carries_no_impact flakes on CI teardown: t.TempDir RemoveAll fails with 'unlinkat .../001/.git/objects: directory not empty' — the subtest's assertions pass, but a background git process spawned in the scratch repo (auto-gc/maintenance or fsmonitor) is still writing objects when cleanup runs. First observed on the v0.5.0 release verify job (run 31960563726), where it fail-closed the whole release after the tag was pushed; a rerun recovered it. Fix direction: the gittest scratch-repo helper should disable background maintenance (gc.auto=0, maintenance.auto=false, core.fsmonitor=false) at init so no child process outlives the test. Detector: the release verify Test step; acceptance: the suite survives teardown on CI without rerun.