---
schema_version: 1
id: "iss-2609020319494139"
slug: "tests-that-build-a-throwaway-git-repository-through-gittest"
severity: "major"
category: "bug"
source: "agent-observation"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/gitutil/repo.go"
---

Tests that build a throwaway git repository through gittest.Env and their own git init still let git spawn detached background maintenance (gc --auto, maintenance.auto), which races the test's cleanup and its own tree walks: the merge queue dropped PR 590 twice on this class within one hour — TestTagAndExplicitKeyShapesAreRecognised (internal/core/reading/assemble_test.go:1241) failed on 'TempDir RemoveAll cleanup: .git/objects: directory not empty', and TestAssembleDryRunWritesNothing (internal/surface/cli/reading_surface_test.go:218) failed on 'lstat .git/objects/maintenance.lock: no such file or directory' on the macOS leg. iss-252 closed this for gittest.NewRepo by writing gc.auto=0, gc.autodetach=false and maintenance.auto=false into the fixture repo's local config, but a test that inits its own repository under gittest.Env gets none of that, and gitutil's isolated environment deliberately scrubs the parent's GIT_CONFIG_COUNT injection without supplying its own. The lifeboat package works around it with a TestMain and an allowlist exemption from the hermetic-git rule. The class remedy is to carry the three async-disable keys in the isolated environment itself, so every isolated git call, abcd's own subprocesses included, runs without detached maintenance and every fixture inherits it.
