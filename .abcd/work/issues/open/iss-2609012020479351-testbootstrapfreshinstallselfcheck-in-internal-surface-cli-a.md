---
schema_version: 1
id: "iss-2609012020479351"
slug: "testbootstrapfreshinstallselfcheck-in-internal-surface-cli-a"
severity: "minor"
category: "bug"
source: "agent-observation"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/surface/cli/bootstrap_freshinstall_test.go"
---

TestBootstrapFreshInstallSelfCheck in internal/surface/cli asserts wall-clock time: it wraps one exec of the provisioned binary answering version in a timer and fails when elapsed exceeds five seconds. It failed on the macOS CI leg after v0.7.0 merged (run 33496438371: 5.77s, 'want about a second') on a tree the merge-queue leg had just passed; locally the same test passes in about one second. The timer measures a single process spawn of an ~18MB freshly written executable, which on macOS carries first-run code-signature validation against a cold page cache, so the number is a property of the runner, not of the code. The test's own comment names the budget as loosened 'so a loaded CI box cannot fail a check about provisioning on a timing wobble', and it then did exactly that. Fourth site of the class recorded by iss-2608292246210181, iss-2608301301041887 and iss-2608290810037763; carried from the session handover into autonomous-run-2026-09-01. The fix must assert the behaviour the gate is about -- the provisioned binary answers version correctly, as the version report, with no Go on PATH -- and carry no duration at all.
