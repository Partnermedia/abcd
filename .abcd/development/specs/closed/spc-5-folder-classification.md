---
id: spc-5
slug: folder-classification
intent: itd-40
---
# folder-classification

## Summary

spc-5 is record catch-up plus a small report cut: itd-40's classification
engine shipped long before this spec existed. Bare `abcd ahoy` was already a
strictly read-only classification pass (`ahoy.DryRun` → `Detect` →
`FolderKind`: `managed-repo` / `unmanaged-repo` / `unmanaged-folder`), install
already bootstrapped and registered `~/.abcd/history/index.json` keyed on the
root-commit SHA, and doctor already resolved the central/host location by
reading that index. What was missing was verified here AC-by-AC against the
built binary, and the two genuine gaps — both in the human report, not the
engine — were closed test-first: the unmanaged-repo report now names
`/abcd:ahoy install` as the adoption path, and the unmanaged-folder report now
states there is nothing to act on.

## Approach

Verification-first, per the run plan's definition of done (the binary is
checked before the record is trusted): each AC was exercised against
`bin/abcd-darwin-arm64` in hermetic scratch dirs (temp `HOME`, plugin root,
and PATH symlink target — the real `~/.abcd` untouched), and only the
behaviours the binary could not demonstrate became code. The one production
edit is confined to `newAhoyCommand`'s bare-ahoy render closure in
`internal/surface/cli/cli.go` — the classification core
(`internal/core/ahoy/detect.go`) needed no change, keeping bare invocation's
read-only contract untouched.

Two ACs were behaviourally met but untested; they received characterization
tests rather than code: AC4's registration half (the pre-existing idempotency
test used a bare `.git` mkdir with an empty root SHA, so registration by real
root-commit SHA was never exercised) and AC5's index-resolved location (no
prior test drove the `auditGaps`/doctor path).

### Deviation: none in behaviour, one in provenance

Implementation for this spec was delegated to a sub-agent worker (manual test
B of the run protocol) with the orchestrator re-running the gate; the
watched-fail evidence below is quoted from the worker's captured runs, and
the run findings log records the delegation experiment
(`.abcd/development/plans/2026-07-16-run-findings.md`, burst 2).

## Milestones as delivered

1. AC-by-AC verification of the shipped classification engine against the
   binary (no engine changes required).
2. Bare-ahoy report cut: adoption-path hint for `unmanaged-repo`,
   nothing-to-act-on line for `unmanaged-folder`
   (`internal/surface/cli/cli.go`), user-facing CHANGELOG entry.
3. Test corpus closing the coverage gaps: two watched-fail report tests plus
   two characterization tests for the already-met ACs
   (`internal/surface/cli/cli_test.go`).

## Acceptance-criteria satisfaction

AC as ordered in itd-40 → covering evidence (tests in
`internal/surface/cli/cli_test.go` unless noted):

1. **Marker block + `.abcd/` → `managed-repo`, mutates nothing** — met by the
   shipped engine; `internal/core/ahoy/detect_test.go`
   (`TestClassifyManagedRepoByAbcdDir`, `TestClassifyManagedRepoByMarker`);
   read-only by construction (`DryRun` delegates to `Detect`, which has no
   writer).
2. **Unmanaged git repo → `unmanaged-repo`, names `/abcd:ahoy install`,
   without adopting** — gap filled here, watched-fail:
   `TestAhoyBareUnmanagedRepoNamesAdoptPath` (asserts the hint and that
   `.abcd/`/`CLAUDE.md` do not appear afterwards); failing output captured
   before the change showed the report ending at the kind line.
3. **Plain folder → `unmanaged-folder`, nothing to act on, mutates nothing** —
   gap filled here, watched-fail:
   `TestAhoyBareUnmanagedFolderReportsNothingToActOn`; mutates-nothing also
   covered by `internal/core/ahoy/detect_test.go`
   (`TestDetectUnmanagedFolderShortCircuits`).
4. **First install bootstraps `~/.abcd/history/` + registers by root-commit
   SHA** — met by the shipped engine; characterization test added
   (`TestAhoyInstallBootstrapsAndRegistersByRootSHA`, real git repo so the
   root SHA is non-empty) over the pre-existing bootstrap coverage
   (`internal/core/ahoy/idempotency_test.go`,
   `TestInstallThenReinstallIsExactNoOp`).
5. **Central/host location resolved by reading `index.json`, never a
   hardcoded path or walk** — met by the shipped engine; characterization
   test added (`TestAhoyDoctorResolvesCentralLocationFromIndex`: doctor
   flags a hand-relocated `path` and quotes the value it read from the
   index, proving the lookup source).

Out-of-scope confirmations: bare `ahoy` never adopts (AC2's test asserts the
absence of installation artefacts); no grouping layer was added; iss-101's
index read-modify-write concurrency was deliberately left untouched (another
workstream owns it).
