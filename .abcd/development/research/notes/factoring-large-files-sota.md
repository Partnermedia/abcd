# Factoring Large Python Source Files — SOTA research + abcd plan

Research deliverable (review session 2026-06-02). Scope: **abcd-owned**
`scripts/abcd/*.py` only. `scripts/ralph/*` and `scripts/ralph/flowctl.py` are
vendored flow-next/Ralph, slated for replacement, and are **out of scope** here.

Companion: the issue ledger entry `[Maintainability/MEDIUM] God-modules` in
`.work/issues.md` (2026-06-02) points here for the remediation plan.

---

## 1. SOTA synthesis — when a large file is a problem, and how to split

**There is no authoritative universal LOC threshold.** The literature supports a
two-factor model, not a line-count rule: the real liability is a file that is
**large × frequently-changing × low-cohesion** (a "large-active file"), where
contributors must edit it "for seemingly unrelated reasons."

- *Large-**active** files* are the TD signal, not size alone — decompose along
  **co-change / responsibility** seams, each seam mapping to a distinct
  responsibility. https://arxiv.org/pdf/2302.09153
- **Change coupling predicts defects better than complexity metrics** — the
  justification for splitting along co-change seams rather than arbitrary size.
  https://link.springer.com/chapter/10.1007/978-3-319-17837-0_1
- LOC is the most-replicated defect predictor (correlational, confounded by
  activity). https://www.mdpi.com/2073-8994/11/2/212 · class size is a
  *confounder*, a proxy not a root cause. https://arxiv.org/pdf/2106.04687
- **Counter-position — Locality of Behaviour (LoB):** behaviour should be obvious
  from looking at one unit; splitting cohesive behaviour across files creates
  "spooky action at a distance" and raises comprehension cost.
  https://htmx.org/essays/locality-of-behaviour/
- Fowler's reconciling rule: "code that changes together [lives] in the same or
  nearby modules"; strict layering at scale produces change-coupled modules —
  prefer **feature/responsibility orientation** for larger units.
  https://martinfowler.com/articles/refactoring-dependencies.html

**Decision framework (when to split / which strategy):**
1. Split only when **size co-occurs with churn and low cohesion**, or when a
   single file holds **≥2 independent responsibilities the agent/human edits
   separately**. Pure LOC reduction is a vanity metric.
2. Choose the axis the code **already changes on**: role / lifecycle-phase /
   verb / document-type — usually marked by the file's own banner comments.
3. **Preserve the public API with a façade** so call sites don't churn.
4. **Do NOT introduce registries/plugin abstraction just to lower LOC** — that
   violates abcd's "three similar lines beat a premature abstraction" rule and
   raises an agent's discovery cost (scatters cohesive logic).
5. **Leave cohesive sub-2k-line modules alone** unless churn data flags them.

**Python mechanics:** module → package; minimal `__init__.py` explicit
re-export (`__all__`); `__main__.py` to preserve `python -m` entrypoints;
extract a **leaf** `_types.py`/`_base.py`/`_substrate.py` first to break cycles;
`if TYPE_CHECKING:` for annotation-only back-edges; keep `__init__.py`
side-effect-free. Tooling: **LibCST** codemods (lossless, preserves the
`spc-NN`/`itd-N` banner comments) → `ruff` (F401/`__all__`) + `mypy/pyright` to
prove no dangling refs → existing test suite as the characterization net (a pure
move is correct iff the suite passes unchanged).
https://libcst.readthedocs.io/en/latest/codemods.html ·
https://www.pythonmorsels.com/fixing-circular-imports/ ·
https://learn.scientific-python.org/development/patterns/exports/

**Agent-maintainability verdict:** target **a handful of ~500–1500-line
cohesive modules** per former monolith — better than both the 6k-line
wholesale-load file (token-expensive, truncation-risky per edit) and a swarm of
tiny files the agent must first discover. The change-locality seam serves humans
and agents alike. (2024–26 agent-context literature: context files / symbol
indices are how agents navigate; monolith loading is the central scaling
problem.) https://blog.jetbrains.com/research/2025/12/efficient-context-management/

---

## 1b. Granularity: one-file-per-function is an ANTI-PATTERN (the ahoy.py question)

A follow-up asked specifically: split `ahoy.py` into `ahoy/__init__.py` + **one
file per command/function**? Answer: **package YES, one-per-function NO.**

**"One class/function per file" is a Java/C# habit, not Python practice.** Java
enforces it at the *compiler* level (public class ↔ same-named file); Python's
unit of namespacing and cohesion is the **module**, meant to "group related
objects into a discrete namespace" (https://docs.python.org/3/tutorial/modules.html).
Canonical Python puts many symbols per module: stdlib `argparse.py` (~2.6k LOC,
all parser/action/formatter classes together), `requests` `models.py`
(`Request`+`PreparedRequest`+`Response`), Django `fields/__init__.py` (dozens of
`Field` subclasses in one file). The Hitchhiker's Guide endorses **minimal/empty
`__init__.py`** and warns aggressive splitting just multiplies parent-`__init__`
import cost (https://docs.python-guide.org/writing/structure/). Rule of thumb
across sources: **"one *idea* per file, not one *function* per file."**

**The boundary principle is Common Closure (CCP):** "classes that change together
belong together" — the component-level Single Responsibility Principle (Robert C.
Martin). Counterweight Common Reuse Principle (CRP) splits apart for reuse; early
in a system you weight CCP (maintainability) over CRP (reuse). So you split on a
**co-change responsibility**, never an individual function. ~20 `_detect_*` fns
that all return `list[Gap]` and all change when the gap model changes are, by
CCP, ONE module — not 20 files.
Sources: https://link.springer.com/chapter/10.1007/978-1-4842-4119-6_8 ·
https://martinfowler.com/articles/microservice-trade-offs.html ·
https://learn.scientific-python.org/development/patterns/exports/ ·
https://testdriven.io/tips/3660b476-7aaa-4f7b-af22-28aa00fc871e/

**Over-splitting (one-per-file) costs:** circular-import multiplication (every
cross-call becomes a cross-module import; importing any submodule re-runs parent
`__init__`), `__init__.py` bloat + import-time cost, "where does this live?"
discovery tax, lost locality-of-behaviour, larger review diffs. **Under-splitting
(monolith) costs:** merge surface + agent/LLM context cost (the only real
pressure on ahoy).

**A/B/C decision rule:** (A) one-symbol-per-file = reject (Java anti-pattern,
justified only for a genuinely independently-reusable symbol); (B) a few cohesive
responsibility-modules along CCP co-change seams = the target; (C) leave whole =
only if truly one responsibility. Tie-break for agent-maintainability toward
"small enough to load cheaply, large enough to stay cohesive" → **B, never A**.

### ahoy.py verdict: **Option B** (8 phase-modules, NOT per-function)

ahoy.py is the cleanest B candidate in the repo — 4,522 lines already
banner-delimited into co-change phases, near-zero cycle risk (it's a leaf: only
`scripts/abcd_cli.py` imports it, nothing imports it back), CLI `python3 -m
scripts.abcd.ahoy`, and a 34-test / ~32-symbol surface that reaches deep private
helpers (`ahoy._apply_visibility`, `ahoy._detect_*`, `ahoy._build_parser`, …) —
all preserved by an `__init__.py` re-export façade.

Proposed `scripts/abcd/ahoy/` (line ranges from the live file):
- `model.py` (L87–365: `Gap`/`GapCategory`/`DetectionResult`/`InstallConfig`/
  `InstallOptions`/`*Result` — change together)
- `detect.py` (L366–1677: ~20 `_detect_*` + `_run_detection_pass`; imports model)
- `resolve.py` (L1678–1747: `_resolve_install_target`)
- `apply.py` (L1748–3544: all `_apply_*` + the 15 sibling writer imports; ~1.8k —
  may sub-split along T3–T6 banners only if it stays cohesive)
- `approval.py` (L3545–3965: `_compute_approved_categories`,
  `_validate_install_options`)
- `report.py` (L3966–4057: `dry_run`, `_render_status`; imports model, detect)
- `runners.py` (`_run_install`/`_run_doctor`/`_run_uninstall`/`_interactive_prompt`)
- `cli.py` (`_build_parser`, `main`)
- `__init__.py` (façade: explicit `__all__` for public names + re-bind the private
  test helpers as attributes + re-bind `ahoy.markers`/`ahoy._history_audit`/
  `ahoy._repo_identity` aliases tests reach through ahoy)
- `__main__.py` (`from .cli import main; sys.exit(main())`)

Acyclic order leaf→root: `model → detect/resolve/approval → apply → report →
runners → cli → __init__/__main__`.

**Sequenced migration (each step independently committable, suite green):**
1. `ahoy.py` → `ahoy/_legacy.py`; add `__init__.py` (`from ._legacy import *` +
   private re-binds) + `__main__.py`. Suite + `tests/scripts/abcd/_no_write_lint.py`
   green ⇒ façade proven before any code moves. **NB: retarget `_no_write_lint`
   (R10 zero-mutation lint) to the package.**
2. Extract `model.py` (no inbound deps). 3. `detect.py`. 4. `resolve.py`+
   `approval.py`. 5. `apply.py` (largest). 6. `report.py` → `runners.py` →
   `cli.py`. 7. Delete `_legacy.py`; final suite + `python3 -m scripts.abcd.ahoy
   --help` smoke.

**Coordination note:** ahoy.py is NOT on the spc-34 surface, but this is a large
mechanical change — sequence it when no other agent is mid-edit on the abcd
package, and use LibCST to preserve the `spc-NN`/`itd-N` banner comments.

---

## 2. The dual public-API constraint (every abcd target)

Both must survive **byte-for-byte** any split:
1. **CLI:** every target runs as `python -m scripts.abcd.<module> …` → package
   needs `__main__.py` (or `main()` re-exported from `__init__.py`).
2. **Symbol imports:** tests import internals heavily (~57 sites:
   `intent_workflow` ~25, `intent_fidelity_reviewer` ~20, `intent_lint` ~12) →
   `__init__.py` re-export façade keeps the dotted import path stable.

`scripts/abcd/__init__.py` already exists and is minimal (good). Per-file
conversion: `module.py` → `module/` package + `module/__init__.py` (re-export) +
`module/__main__.py` (delegates to `main()`).

---

## 3. Per-file proposals (abcd-owned, ranked by seam-cleanliness × agent-edit cost)

Recommended global order: **B1 reviewer → B3 workflow → B2 lint → B4 ahoy.**
Each extraction = one commit = green suite (Mikado leaf-first).

### B1. `intent_fidelity_reviewer.py` (6,333) — SPLIT, highest priority
Clearest seams in the codebase (author banners confirm): three **roles** +
shared substrate + dispatch + parsers/writers + a self-contained spc-23
issue-drift subsystem. Proposed `intent_fidelity_reviewer/`:
`_types.py` (leaf: verdict/result dataclasses+enums) · `_common.py` (identity,
fenced-block extraction, atomic writes, `_oracle_dispatch*`, runtime-invariant
validators) · `itd1.py` · `mg004.py` · `role2.py` · `role3.py` ·
`issue_drift.py` (spc-23, already self-contained) · `__init__.py`+`__main__.py`.
Sequence: package-ize w/ façade → `_types` → `issue_drift` → `_common` →
roles/itd1/mg004 one per commit.

### B3. `intent_workflow.py` (3,978) — SPLIT, verb seam
Substrate + four lifecycle verbs (`shape`, `plan_single`, `reclassify`,
`plan_bundle`) + CLI. `intent_workflow/`: `_substrate.py` (paths, `ShapeReport`,
`_ShapeLock`/flock, declined store, findings.jsonl, atomic helpers,
`reserve_and_write_intent`, hash helpers — verbs depend *down* only) ·
`shape.py` · `plan_single.py` · `reclassify.py` · `plan_bundle.py` ·
`__init__.py`/`__main__.py` (`bare`/`grill` subcommands). Keep workflow→lint
import one-way.

### B2. `intent_lint.py` (4,418) — SPLIT, linter-per-class seam
`intent_lint/`: `_base.py` (`Finding`, config/severity, glossary loaders,
frontmatter/body/hash helpers) · `intent_linter.py` (`IntentLinter`, ~1.8k —
stays largest, acceptable: rule methods are cohesive & co-change) ·
`spec_linter.py` (`SpecLinter` MG/VR) · `bundle_linter.py` · `_cli.py`. **Do NOT
registry-ize the `_check_*` rules** — they share instance state and trivial
sequential dispatch; a registry is the premature abstraction CLAUDE.md forbids.
Note the dual run mode (`python -m …` AND `python scripts/abcd/intent_lint.py`)
+ `_bootstrap_path()` — preserve in `__main__.py`.

### B4. `ahoy.py` (4,522) — SPLIT, lower priority, detect/apply phase seam
`ahoy/`: `_model.py` (`Gap`, `InstallConfig/Result`, …) · `detect.py` (all
`_detect_*`) · `apply.py` (all `_apply_*` + `_run_install`) · `report.py`
(dry-run/status/doctor/uninstall) · `__init__.py`/`__main__.py`. Lower priority:
functions are already small/hermetic, so per-edit agent cost is lower than the
role/verb monoliths even at 4.5k lines.

---

## 4. Explicit leave-as-is

`capture.py` (2,077), `_issue_lib.py` (2,046), `mcp_bridge.py` (1,865),
`_prd_writer.py` (1,622) — cohesive, agent-loadable in one read, no clean
multi-responsibility seam. Splitting would be LOC-driven and scatter behaviour
(LoB cost) for no locality gain. Revisit only if churn flags one as large-active.

`scripts/ralph/*`, `flowctl.py` — out of scope (vendored, replacement = Strangler
Fig territory, not file-split territory).

---

## 5. Success metric (not LOC)

After each file: a *typical* edit ("fix a Role 2 rule", "tweak `plan_bundle`")
touches **one** new module, and **no test import path changed**. That — not the
line-count delta — is the win.

## Sources
See inline URLs above. Primary: arXiv 2302.09153 (large-active files), Springer
978-3-319-17837-0_1 (change coupling vs defects), htmx LoB essay, Fowler
refactoring-dependencies, LibCST codemods, Scientific-Python exports guide.
