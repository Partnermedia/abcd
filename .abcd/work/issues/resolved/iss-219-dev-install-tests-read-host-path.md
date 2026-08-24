---
schema_version: 1
id: "iss-219"
slug: "dev-install-tests-read-host-path"
severity: "major"
category: "bug"
source: "manual-test"
found_during: "manual-capture"
found_at: "internal/core/ahoy/dev_install_test.go"
resolution: "ahoy detect tests run under setupHermetic (HOME pinned to temp, PATH scrubbed); the leaker test uses it and iss-249's hermetic work landed alongside"
impact: internal
---

ahoy dev-install tests read the host PATH: a stale installed abcd (e.g. the July dev shim in ~/.local/bin) makes the four install-mode tests fail with an unexpected '(shadowed on PATH)' suffix, so make preflight blocks pushes from any machine with abcd installed. The tests should run against an isolated PATH.

## Root cause confirmed (2026-08-16) — a test WRITES to the real host home

The "stale installed abcd" is not an external artefact the tests merely read — it is **leaked by the test suite itself**. `TestInstallAdoptsIdentityPinInteractively` (internal/core/ahoy) installs its symlink into the operator's **real** `~/.local/bin/abcd` rather than a sandboxed HOME, pointing at its own `t.TempDir()`; when that temp dir is cleaned up the symlink is left **dangling**, and it then shadows PATH for every later run. Observed specimen: `~/.local/bin/abcd -> /var/folders/.../TestInstallAdoptsIdentityPinInteractively.../002/abcd` (target gone), dated to the test run.

Causation proven: with the dangling symlink present the four install-mode tests fail (`install_mode = "dev (tip build) (shadowed on PATH)"`, `"pinned (shadowed on PATH)"`); after `rm`-ing it and running the package with the leaker **skipped**, `internal/core/ahoy` passes clean and the symlink does not reappear. The leaker's write to the real home is **condition-dependent, not every-run**: the specimen proves it happened, but a later full-suite preflight on the cleaned machine (no `abcd` on PATH) did **not** re-leak — most likely the adoption path only writes when there is an existing `abcd` on PATH to adopt, so a leaked symlink bootstraps its own next leak. Either way it is a latent host-pollution bug and the test must be sandboxed regardless.

So this is **two test-isolation defects, and severity is major** (a unit test corrupts the operator's real home and breaks the local push gate until manually cleaned):

1. **The leaker WRITE.** `TestInstallAdoptsIdentityPinInteractively` must install into a sandboxed HOME/bin (a `t.TempDir()` set as HOME, or an install-target override), never the real `~/.local/bin`. This is the source and the higher priority — it can pollute the host (condition-dependent, see above), and a leaked symlink then breaks every later run until cleaned.
2. **The victims' READ** (the original capture). The four install-mode tests must resolve install mode against an isolated PATH so no pre-existing/leaked `abcd` yields a `(shadowed on PATH)` false failure.

Fix both; the leaker first. Interim: the dangling symlink was removed manually 2026-08-16 (`/bin/rm -f ~/.local/bin/abcd`); the very next full-suite push did not re-leak it, so pushes from this machine no longer need the PATH workaround for now. That is a reprieve, not the fix — sandboxing the test is what closes the latent leak.