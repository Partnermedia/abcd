---
schema_version: 1
id: "iss-252"
slug: "release-test-tempdir-git-objects-flake"
severity: "minor"
category: "bug"
source: "agent-observation"
found_during: "v0.5.0 release verify (run 31960563726)"
found_at: "internal/core/release/emit_test.go"
resolution: "gittest.NewRepo now writes gc.auto=0, gc.autodetach=false, maintenance.auto=false and core.fsmonitor=false into the fixture's repo-local config at init, so no git command run in a fixture — by the fixture or by the code under test — can spawn a background maintenance process that outlives the test and races t.TempDir cleanup on .git/objects. Repo-level rather than per-command because the code under test runs its own git against the fixture. Pinned by TestNewRepoDisablesBackgroundMaintenance."
impact: internal
---

TestEmitRefuses/an_added_record_carries_no_impact flakes on CI teardown: t.TempDir RemoveAll fails with 'unlinkat .../001/.git/objects: directory not empty' — the subtest's assertions pass, but a background git process spawned in the scratch repo (auto-gc/maintenance or fsmonitor) is still writing objects when cleanup runs. First observed on the v0.5.0 release verify job (run 31960563726), where it fail-closed the whole release after the tag was pushed; a rerun recovered it. Fix direction: the gittest scratch-repo helper should disable background maintenance (gc.auto=0, maintenance.auto=false, core.fsmonitor=false) at init so no child process outlives the test. Detector: the release verify Test step; acceptance: the suite survives teardown on CI without rerun.