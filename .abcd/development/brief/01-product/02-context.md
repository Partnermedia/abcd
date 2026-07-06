# Context

You did a manual lifeboat from iDelphiZero → iDelphi (`idelphiDev/.work/lifeboat/rescue/extraction.md` is the gold standard). Before rebuilding iDelphi from scratch, you want to automate: (a) **packing** lessons from an in-progress repo into a portable lifeboat artefact, and (b) **unpacking** that artefact to bootstrap a clean rebuild.

The plan ships these user-facing commands as a Claude Code plugin (`abcd`):

- **`/abcd:ahoy install`** — install / update abcd in any project (bootstraps configuration, gitignore, marker blocks, PATH symlink). Bare `/abcd:ahoy` shows status+help.
- **`/abcd:disembark to <path>`** — *pack* a lifeboat from the current project to `<path>` (use `home` for current repo's `.abcd/lifeboat/`). Bare `/abcd:disembark` shows status+help; `probe` and `dry-run` sub-verbs preview without writing.
- **`/abcd:embark from <path>`** — *unpack* a lifeboat at `<path>` into a (typically empty) target project (use `home` for current repo's `.abcd/lifeboat/`). Bare `/abcd:embark` shows status+help; `scan` and `probe <path>` sub-verbs discover/inspect without unpacking.
- **`/abcd:launch ship`** — promote `*Dev` (private) repo to its public sibling, scrub for PII/secrets, bump versions, update marketplace.json. Bare `/abcd:launch` shows status+help; `dry-run` sub-verb runs the full pre-flight gate suite without writing the public payload.
- **`/abcd:intent`** — bare quoted `/abcd:intent "<text>"` is the canonical create (fn-30/itd-46), plus `refine` / `grill` / `plan` / `ship` / `review` / `consistency` / `shape` / `reclassify` / `link` sub-verbs (the `consistency` sub-verb shipped in fn-29 per itd-48, which superseded itd-31). There is no plain `list` sub-verb — it is folded into the bare render per SD001. Manages **intents** (press-release-format intent docs at `.abcd/development/roadmap/intents/{drafts,planned,shipped,disciplines,superseded}/`). `plan` calls `/flow-next:plan` + `/flow-next:plan-review`; `ship` calls `/flow-next:work` (or full pipeline if from drafts/); the flowctl `spec close` close-hook (fn-36) reconciles standalone/bundle intents planned → shipped automatically on a successful close (fn-28 `intent_lifecycle.reconcile`); disciplines move from drafts/ to disciplines/ on plan and stay there. Bare `/abcd:intent` shows status+help.
- **`/abcd:capture`** — capture / list / promote / resolve / wontfix issues (the structured ledger version of `.work/issues.md`). Issues live at `.abcd/development/activity/issues/{open,resolved,wontfix}/iss-N-<slug>.md`. See itd-4. The cross-corpus synthesist (`/abcd:dredge`) comes in a later phase as itd-25.
- **`/abcd:memory`** — curate a queryable knowledge substrate (`ingest` external sources / `ask` queries / `lint` health-checks) from specs, ADRs, reviews, and memory. See itd-36 and [`05-internals/07-memory.md`](../05-internals/07-memory.md).

**Plus the `intent grill` sub-verb** (sibling of `refine` under `/abcd:intent`):

- **`/abcd:intent grill <itd-N>`** — Socratic-questioning sub-verb that stress-tests an intent (or brief section, via `--brief-section <id>`) before planning, surfacing unstated assumptions and emerging glossary terms. Sibling of `/abcd:intent refine` (gentle / user-driven). See itd-27.

**The pack/unpack model:**

```
[source repo: .flow/, .abcd/memory/ (or legacy memory/), ADRs, docs, code, transcripts]
        │
        │ /abcd:disembark to <path>   (PACK; `home` = current repo's .abcd/lifeboat/)
        ▼
[lifeboat artefact: a portable directory]
   ├── README.md, press-release.md, principles.md   ← synthesised
   ├── rescue/specs/, docs/adrs/                    ← verbatim copies
   ├── research/, audit/                            ← passes B/C outputs + audits
   └── _provenance.json                             ← (full shape in [04-surfaces/02-disembark.md § 5](../04-surfaces/02-disembark.md#5-output-shape))
        │
        │ /abcd:embark from <path>   (UNPACK; `home` = current repo's .abcd/lifeboat/)
        ▼
[target repo: files placed at canonical locations]
```

Lifeboats are portable directories; share by copy/tar/git, not by global archive. **The lifeboat is always *output*** — `.abcd/lifeboat/` in any repo is the latest disembark snapshot, regenerable from current state. Embark/disembark provenance and history live separately at `.abcd/development/voyage/` (see [`02-constraints/01-platform.md`](../02-constraints/01-platform.md) and [`04-surfaces/03-embark.md § 7`](../04-surfaces/03-embark.md#7-voyage-layout-embarkdisembark-provenance-and-history)); the lifeboat itself never accumulates past versions.

**Repos:**

- `abcdSubZero/` — existing Python CLI (first-version lifeboat, **reference only**, not canonical). Phase 0 includes a skim for patterns to learn from; do not plan to port.
- `abcdDev/` — **this directory**. The private dev repo where the plugin is built. Currently bootstrapped (initial commit + this brief); Phase 0 onwards adds the plugin scaffolding here.
- `abcd/` — NEW public release target. Receives clean payload via `/abcd:launch ship`.

**Validation corpus:** user-maintained list in `.abcd/corpus.json`, seeded with:

- `idelphiDev/` (primary) — 41 flow-next specs, 401+ specstory transcripts, manual lifeboat present
- `abcdSubZero/` — Python CLI, no flow-next, lighter docs (exercises adapter graceful-degradation)
- `idelphiSubZero/` — likely sparse on .specstory/.flow (exercises Pass A with thin inputs, code-rescuer leaning on source)

Per-phase acceptance runs against the corpus with documented exemptions where a feature genuinely doesn't apply (e.g., flow-essence on a no-`.flow/` repo emits empty `spec-essence.json` — oracle should note and pass).
