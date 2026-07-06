# fn-34 flowctl divergence audit (live, v1.3.4)

Live re-audit of the vendored fork `scripts/ralph/flowctl.py` against the
**installed** flow-next plugin, resolved by path via
`_version_resolution.select_newest_version_dir` (TRACK LATEST — newest installed,
not a pinned version). Supersedes the stale 1.1.11 baseline and the obsolete
`+1989/−506` numbers from the original spec.

## Method

```bash
PLUGIN=$(python3 -c "import sys; sys.path.insert(0,'.'); \
  from scripts.abcd.tools.flowctl_loader import resolve_installed_flowctl; \
  print(resolve_installed_flowctl())")
diff -u "$PLUGIN" scripts/ralph/flowctl.py | diffstat
```

## Result (installed v1.3.4)

| | lines |
|---|---|
| installed plugin `flowctl.py` | 23 291 |
| repo fork `scripts/ralph/flowctl.py` | 24 143 |
| diffstat | **+2 171 / −1 319** (1 file, 3 490 changed) |

The fork carries ~850 more lines than upstream, dominated by the abcd-only `rp`
command surface (39 `cmd_rp_*` / `_rp_*` / `rp_*` functions in the fork; the
render-budget closure alone is 11 abcd-only functions).

## A/U/S classification (the adjustment categories)

- **A (abcd-added, re-authored in fn-34.7):** the `rp render-budget` closure —
  `cmd_rp_render_budget`, `rp_enforce_render_budget`, `rp_render_payload`,
  `_rp_render_export`, `_rp_embed_max_bytes`, `_rp_select_get_paths`,
  `rp_select_remove`, `_rp_native_remove`, `_rp_set_selection`,
  `_rp_set_selection_via_clear_add`, `_rp_normalize_selection_path`,
  `_rp_pattern_matches`, plus `RP_BUDGET_TRIM_LIST` /
  `FLOW_RP_EMBED_MAX_BYTES_DEFAULT`. These exist ONLY in the fork (confirmed
  absent from the installed plugin) and are re-authored into
  `scripts/abcd/abcd_flowctl_ext.py` — NOT imported from the fork (which stays
  upstream-unforked; its deletion is the deferred Phase B follow-up).
- **U (upstream, importlib-loaded — NOT re-authored):** the three RP helpers
  `run_rp_cli`, `run_rp_cli_unchecked`, `error_exit` are loaded by path from the
  installed plugin (fn-34.2 `flowctl_loader`). Note: `error_exit` is additionally
  re-authored locally as a trivial sys.exit wrapper for the extension's own
  diagnostics (documented in `abcd_flowctl_ext.py`); the two RP-cli helpers that
  need the plugin's exact behaviour are the genuinely loaded symbols.
- **S (shared/structural):** `RUNTIME_FIELDS` (7-element set) lived only in the
  fork; fn-34.7 COPIES it to the neutral `scripts/abcd/tools/flowctl_runtime.py`
  (reconciliation test asserts copy == fork at copy time). The remaining diff is
  structural drift (ordering, upstream refactors) not abcd-owned.

## The render-budget bug (fixed in re-authoring)

The fork's `_rp_render_export` used
`tempfile.NamedTemporaryFile(prefix=..., suffix=..., delete=False)` then read the
named file back and manually unlinked it. The re-authored body uses
`tempfile.mkstemp(...)` + immediate `os.close(fd)` + `os.unlink` in a `finally`
— the mkstemp+unlink pattern, no `NamedTemporaryFile`. Behaviour is otherwise
preserved (invoke `prompt export <tmp>`, read back UTF-8, return the string).

## Reconciliation

The diffstat above is reproduced by the live `diff … | diffstat`. When upstream
advances, this audit is re-run against the new newest plugin (track-latest); the
numbers are a snapshot, not a pin.

## Related

- [fn-34 spec](../../../.flow/specs/fn-34-detach-scriptsralphflowctlpy-from.md)
- Extension: `scripts/abcd/abcd_flowctl_ext.py`
- Loader: `scripts/abcd/tools/flowctl_loader.py` (fn-34.2)

**Status**: Live re-audit (fn-34.7). The fork stays upstream-unforked under the
corrected model; the live `scripts/ralph/flowctl` wrapper execs it for non-abcd
verbs. Deleting the fork is the deferred Phase B follow-up.
