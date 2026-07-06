# fn-82.6 — overlay/deps hygiene outcome declaration

Machine-readable outcome for the two independent halves of fn-82.6 (R6). The
two declaration lines below are the load-bearing contract — the branch each half
took is explicit regardless of which sub-checklist is filled.

```
pin: void
patch:ralph-guard-memory-hint: stays
patch:ralph-guard-git-push: stays
```

## Part (a) — dependency pin advance: VOID

`scripts/abcd/overlay/deps.json` `current_pin` stays `1.1.11`. The overlay pinned
snapshot at `tests/abcd/_fixtures/_pinned-upstream-snapshot/1.1.11/` is untouched.

A valid range receipt for a `1.1.11 → <target>` advance **cannot be produced in
this context**, so the pin-advance half exits VOID with `deps.json` and the
pinned snapshot provably unchanged (the spec's designed VOID path):

1. **The advance surface refuses under FLOW_RALPH=1 (fail-closed by design).**
   Pin advancement is a manual-operator action; `advance_pin()`
   (`scripts/abcd/deps_check.py`) refuses with exit 8 (`ralph_context_refused`)
   BEFORE the ledger gate whenever `FLOW_RALPH=1`. This task ran inside a Ralph
   session, so the sanctioned advance command is structurally unavailable here:

   ```
   $ FLOW_RALPH=1 python -m scripts.abcd.deps_check --advance-pin flow-next --to 2.5.1
   abcd deps:check --advance-pin: pin advancement is a manual-operator action;
   refusing to run inside a Ralph session (FLOW_RALPH=1 detected). ...
   # EXIT=8
   ```

2. **No range receipt exists in the operator ledger.** There is no
   `.work/abcd/dep_watcher_ledger.json` carrying a `range_receipts[]` entry that
   covers `dep=flow-next from_pin=1.1.11 to_version >= <target>` with
   `range_complete: true`. Building one requires a live `compare_api` fetch via
   `dep_watcher.check_deps()`, which fn-67 hermeticity forbids in tests and which
   a Ralph session must not perform. Absent that receipt the gate would refuse
   (`refused_range_missing`).

3. **`--force` is FORBIDDEN for this spec's acceptance (R6).** The command exposes
   `--force --confirm-dep-id`, but using it would fail R6. So the operator-bypass
   path is not an option to synthesize an advance.

Stepwise-vs-direct is therefore not decided — no advance occurred. When an
operator later advances the pin from a normal interactive session, the
range-receipt evidence they produce (via `deps-check --check`) will decide
stepwise (through the landed `1.3.4` fixture) vs direct; that decision is
recorded at advance time, not here.

**No-receipt finding:** filed to `.work/issues.md` (2026-07-03) — the pin lags
newest installed flow-next; advancing it needs a manual-operator interactive
session to produce a `range_complete` receipt. A lagging pin degrades to warn,
never fail (track-latest, fn-34), so this is debt, not breakage.

## Part (b) — manifest-patch migration probe: both patches STAY

`patch:ralph-guard-memory-hint` (`manifest.json:154`) and
`patch:ralph-guard-git-push` (:186) were probed against the `normalize_rule`
shape (`scripts/abcd/normalize_rules.json` + `scripts/abcd/tools/normalize.py` +
`scripts/abcd/schemas/normalize_rule.schema.json`). Neither fits. No migration;
no `normalize_rule.schema.json` change (fn-77..80 remain open — a needed
extension would be a recorded finding, not an edit).

### Why the `normalize_rule` model cannot express these patches

The `normalize_rule` engine is a **stateless, span-aware regex substitution** over
an installed dependency's files. Each rule is `(match regex → rewrite template)`
where the template's `<token>` is interpolated from the matched text's basename
minus a **hardcoded** leading `/tmp/` (`normalize._render_rewrite`:
`matched[len("/tmp/"):]`). It has:

- no marker / ownership sentinel (idempotency is by construction — the rewritten
  form never re-matches `match`);
- no `--unapply` / reverse direction, and therefore no predecessor capture;
- no structural anchor-walking — it is line/span regex only;
- a read-only `check()` vs write `apply()` split driven by a version-staleness
  verdict, not an apply/check/unapply/list-targets patch contract.

Both probed patches require capabilities **outside** that model:

| Requirement of the patch | `normalize_rule` support |
|---|---|
| Marker-gated ownership (`# patched-by abcd:…`) so `--unapply` only reverses OUR bytes | none — no marker concept |
| Byte-exact `--apply`→`--unapply` round-trip via a `.abcd-overlay-predecessor` sidecar | none — no reverse op, no sidecar |
| Structural enclosing-block walk (`_find_enclosing_git_push_block`, `_find_handle_pre_tool_use_bounds`) and insert-vs-replace change-kinds | none — flat regex span match |
| A **fixed** multi-line replacement block (`MARKED_BLOCK` / `NEW_BLOCK`), NOT a `<token>`-interpolated `/tmp/` basename | `rewrite` requires `<token>`; `_render_rewrite` hardcodes `/tmp/` |
| `target_class: both` (repo file + plugin-cache copy) with per-class receipts | `target_globs` glob one installed-tool root; no repo-file class |

The schema itself blocks a mechanical migration: `rewrite` is documented as a
`/tmp/`-basename template and `additionalProperties: false` leaves no field to
carry a fixed block, a marker, or a reverse operation. Expressing either patch
would require a schema extension — forbidden while fn-77..80 are open, and it
would be a recorded finding, not an edit made here. No such extension is even
warranted: these are structural, reversible, marker-owned block patches, a
fundamentally different transform class from `normalize_rule`'s idempotent
one-way path scoping.

### Upstream self-retirement check against the pin

Each patch's `upstream_resolution_criterion` (manifest.json) was evaluated
against the current pin (`1.1.11`) and the pinned snapshot
(`tests/abcd/_fixtures/_pinned-upstream-snapshot/1.1.11/scripts/hooks/ralph-guard.py`):

- **`patch:ralph-guard-memory-hint`** — retires when upstream's memory-add hint
  emits the itd-39 `--track`/`--category` grammar. The pinned snapshot's guard
  still carries the deprecated `--type <type>` hint (the patch's `OLD_BLOCK`
  fingerprint), so upstream has **not** self-retired it. STAYS.
- **`patch:ralph-guard-git-push`** — retires when upstream blocks `git push`
  natively in `handle_pre_tool_use`. The abcd-local block is still the maintained
  policy (the patch inserts/migrates it), so upstream has **not** self-retired it.
  STAYS.

Neither patch is upstream-self-retired at the current pin; both remain `required`
overlay entries. When the pin later advances, `deps-check --check`'s
`retirement_match` evaluation re-runs against the new upstream and may flip either
to self-retired at that time.
