# fn-25.1 (T1) — flowctl-validate integration feasibility

## Probe outcome

T1 investigated whether `flowctl` exposes a plug-in / extension surface for
additional validators (e.g. a `.flow/validators/*.py` discovery path, a
`flowctl plugin register` verb, or a `meta.json` hook).

**No such surface was found** in the current `flow-next/1.1.11` plugin
distribution. `scripts/ralph/flowctl.py` and `flowctl-py` carry the
verb dispatch table inline; there is no documented or undocumented
extension seam an out-of-tree validator could attach to without
modifying upstream files.

The spec's "forbidden third path" (in-place edit of
`scripts/ralph/flowctl.py` to insert overlay validation) is **NOT
taken**. That escalation is documented as a follow-up spec, not implicit
T1 work.

T1 therefore implements the **direct-invocation fallback**:

1. **CI gate** — `.github/workflows/overlay-validate.yml` runs
   `python3 scripts/abcd/overlay/validate_manifest.py` over the
   committed manifest+deps pair on every PR touching the overlay tree
   or its supporting helpers. The job installs the project's editable
   surface first (`pip install -e .`) so the validator can import
   `packaging`. Non-zero exit fails the PR check.

2. **Pre-commit hook** — `.pre-commit-config.yaml` carries a new
   `abcd-validate-overlay-manifest` hook that fires on commits touching
   `scripts/abcd/overlay/manifest.json` OR `scripts/abcd/overlay/deps.json`.
   The probe found `.pre-commit-config.yaml` present at repo root; the
   conditional-on-existence branch (per spec body) fired and the hook
   was appended.

3. **/abcd:ralph-up --check** (T5 deliverable) will additionally invoke
   the same validator as part of its pre-apply validation. T1 does NOT
   wire this leg — T5 owns the command.

## Dependency-surface probe outcome

T1 probed `Path("pyproject.toml").exists()` and confirmed:
- `pyproject.toml` exists at repo root.
- The `[project] dependencies` array exists and is the canonical
  declaration surface.

**Branch 1 (preferred) fired**: `packaging>=21.0` was appended to
`pyproject.toml`'s `[project] dependencies` array. The resolved install
command is `pip install -e .`, recorded in
`scripts/abcd/_install_command.json` for the runtime preflight helper
to read.

Verification: `python3 -c "import packaging.version; assert packaging.version.parse('1.0.0') < packaging.version.parse('1.1.0')"` succeeds after `pip install -e .`.

## Pinned upstream snapshot pristine-validation

T1 attempted the spec's **preferred sourcing path** (GitHub release-tag
extraction from `gmickel/claude-marketplace`) and found the API + tarball
both return 404 — the repo is **not publicly reachable** from this
operator's environment. Documented separately in
[`t1-snapshot-blocked.md`](t1-snapshot-blocked.md).

T1's **fallback** is the local-cache copy with mechanical contamination
removal (`tests/abcd/_build_pristine_fixture.py`). The script copies a
minimal subset of the cache tree (only files current overlay patches
target — total ~270KB), then:
- strips any line containing `patched-by abcd:` or `0.14.0-abcd1`,
- reverses the only patch transformation (`${TMPDIR:-/tmp}/` → `/tmp/`),
- re-scans the cleaned tree and FAILS if any contamination marker
  survives.

The resulting pristine fixture lives at
`tests/abcd/_fixtures/pristine-flow-next-snapshot/` and was used as
the source for the canonical pinned-snapshot at
`tests/abcd/_fixtures/_pinned-upstream-snapshot/1.1.11/` (per spec body's
"under the CI / `ABCD_PRISTINE_FIXTURE_PATH` path T1's `init_deps.py`
copies the fixture tree into the canonical pinned-snapshot location").

Pristine assertion: `tests/abcd/test_pinned_snapshot_pristine.py`
verifies `tests/abcd/_fixtures/_pinned-upstream-snapshot/<current_pin>/`
exists, contains no `patched-by abcd:` marker, and the `<current_pin>`
segment equals `deps.json.dependencies[0].current_pin`. All three
assertions pass.

## Files T1 landed

- `scripts/abcd/_install_command.json` — install-command config for
  the runtime preflight helper (branch 1 outcome).
- `scripts/abcd/_packaging_preflight.py` — `run_or_exit()` helper
  consumed by T5/T8's CLI passthrough branches (lazy-import contract).
- `scripts/abcd/_version_resolution.py` — shared
  `is_canonical_version` / `parse_version` /
  `select_newest_version_dir` helpers used by validator + scaffold +
  T4/T5/T6.
- `scripts/abcd/overlay/manifest.json` — empty `entries: []` array
  (T2/T3/T4 populate).
- `scripts/abcd/overlay/deps.json` — `flow-next` dep entry with
  `current_pin: "1.1.11"`, `overlay_items_watching: []` (T2-T4 append).
- `scripts/abcd/overlay/validate_manifest.py` — static-schema validator.
- `scripts/abcd/overlay/init_deps.py` — scaffold helper.
- `tests/abcd/_build_pristine_fixture.py` — operator-side fixture
  builder.
- `tests/abcd/_fixtures/pristine-flow-next-snapshot/` — committed
  pristine fixture (≈270KB, larger than the spec's aspirational 200KB
  cap because the patch-target set across `skills/` + `codex/skills/`
  + `scripts/hooks/ralph-guard.py` legitimately exceeds it).
- `tests/abcd/_fixtures/_pinned-upstream-snapshot/1.1.11/` — committed
  canonical pinned snapshot.
- `.github/workflows/overlay-validate.yml` — CI gate (path-filtered).
- `.pre-commit-config.yaml` — new `abcd-validate-overlay-manifest`
  hook entry (probe found the config present).
- Test suites: `test_version_resolution.py`,
  `test_validate_manifest.py`, `test_t1_preflight.py`,
  `test_init_deps.py`, `test_pristine_fixture_present.py`,
  `test_pinned_snapshot_pristine.py`, `test_packaging_preflight.py`.
- `pyproject.toml` — appended `packaging>=21.0` to `[project] dependencies`.

## What T1 did NOT do (downstream tasks own these)

- Populate `manifest.json` with entries (T2 hook_source,
  T3 hook_registration, T4 patches).
- Implement `apply.py` (T5).
- Implement `dep_watcher.py` (T6).
- Implement `/abcd:ralph-up` / `/abcd:deps-check` CLI verbs
  (T5/T8); the preflight helper is wired into those branches by their
  respective owners.
- Pre-populate `OPTIONAL_PROTECTED_PATHS_ABCD_CONTROL_PLANE`
  (T2 owns the protected-paths frozenset; T2's commit lists T1's
  helper paths as pre-populated entries owned by T1).
- Apply-time existence checks (`--apply-time` flag is reserved; T5
  wires the existence-check legs).
