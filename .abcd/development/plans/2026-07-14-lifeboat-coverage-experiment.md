# Plan: The lifeboat as a coverage experiment — prove the premise before building the packer

> This plan pulls the lifeboat round-trip out of Phase 6 and inverts its build order.
> The decisions it locks are ratified in adr-35 (and amend adr-4); this plan sequences
> the build and does not re-argue them — follow the ADR for the *why* and the
> alternatives rejected.

## Context and thesis

The lifeboat (`/abcd:disembark` packs it, `/abcd:embark` unpacks it) is what the press
release says abcd *is*. It is also the only major surface with **zero code**: no
`internal/core/lifeboat/` package, no verbs, no agent prompts. The roadmap parks it at
[Phase 6](../roadmap/phases/phase-6-lifeboat.md) — last — on the stated grounds that it
"depends on every prior substrate being native."

**That rationale is mostly false, and the error is load-bearing.** Checked against the
binary rather than the record (per [`CONTEXT.md`](../../work/CONTEXT.md): *where the record
and the binary disagree, verify against the binary*):

| Roadmap says it blocks | Reality in the binary |
|---|---|
| Phase 4 native spec engine | `spec.Load` / `Create` / `Close` / `Validate` all ship. **Not a blocker.** |
| Phase 5 run seam | Backgrounding is a host affordance; the design's own checkpoint is "forensic only — abcd ships no `resume` verb". **Not a blocker.** |
| Phase 3 reviews (itd-28) | Reviews are already committed markdown under `.abcd/work/reviews/`. **Not a blocker.** |
| itd-2 in-session subagent dispatch | The host-delegation seam is **already shipped twice**: `memory.Distiller` fed by `--pages-json`, and `intent review ingest --verdict-json` with a dead-letter path. **Not a blocker.** |
| Phase 2 history + memory | The *packages* are built. The **data is empty**. Real — and not a code problem. |

So none of Phases 3, 4, or 5 gate this work.

### The goal is not "ship the lifeboat"

The aim is to **learn what can actually be transferred out of a repo**, and specifically
**whether the brief's structure is sound** — whether it is a shape a product thinker can
fill when a lifeboat cannot derive it all. The candidate corpus is a handful of repos of
mixed record quality: at least one abcd-managed (this one), the rest with little or none.
**The messy repos are instruments, not deliverables.**

This inverts the build. The primary output of the first phase is not a lifeboat — it is a
**verdict on the brief structure**: which sections are derivable from a real repo in
practice, which are always blank, and what the record buys over raw git. The lifeboat is
the apparatus that produces the verdict.

Therefore: **`probe` → coverage report, across all repos, before `pack` exists at all.**
The question gets answered without writing a single lifeboat.

### The headline number

Run the same probe over this repo (rich record) and over a git-only repo. **The delta in
brief-section coverage is what the record is worth.** That number tells us whether abcd's
premise — that writing things down is the thing worth doing — survives contact with
reality. Nothing else here matters as much.

## What this changes versus the design as written

**1. Disembark becomes read-only and out-of-tree.** `abcd disembark <source-repo> to <dest>`
— point it at any repo, touch nothing, write elsewhere. [`02-disembark.md`](../brief/04-surfaces/02-disembark.md)
assumes cwd *is* the source and writes `.abcd/lifeboat/` plus `voyage/` back into it. That
is backwards for mining a dead or archived project, and it forces `ahoy install` into a repo
we only want to read.

Consequence: **`voyage/` moves to `~/.abcd/voyage/<source-root-sha>/`**, matching the
existing history-store convention (keyed on the root-commit SHA). It is therefore never
committed — which **dissolves the `privacy-hygiene` collision** (voyage records absolute
source paths; the `privacy-hygiene` audit rule flags `/Users/<name>/` in committed files,
so abcd would have failed its own audit) and most of the gitignore work. **adr-4 must be
amended**: it defines voyage as the operations namespace *inside the source repo*.

**2. Coverage becomes a first-class, aggregatable output** — not an exemption footnote.
Clean brief plus a separate `coverage.{json,md}`. The brief carries only what abcd could
ground, each claim citing its source. Coverage carries what is missing, what was searched,
and the confidence of each derived claim — in a schema that **aggregates across repos**,
because that aggregate *is* the experiment's readout.

**3. The graveyard is a new first-class section that exists nowhere in the design.**
"Extract what failed, so it isn't tried again." Three layers, and the order is the
anti-fiction discipline:

| Layer | Source | Works on |
|---|---|---|
| 1. **Archaeology** | reverted commits, branches abandoned unmerged, files and directories deleted after substantial history, dependencies added then removed, wholesale rewrites | **any git repo** |
| 2. **Recorded abandonment** | superseded ADRs and intents, `wontfix` issues, `## Alternatives Considered`, rejected options named in decision logs | repos with a record |
| 3. **Interpretation** | an agent reads layers 1 and 2 and says what was tried and why it was dropped | anything — **but every claim must cite the evidence beneath it** |

Layer 3 sits at the bottom and cannot float free. This is the mechanism itd-11 was reaching
for, made structural.

## The data constraint that no code fixes

`~/.abcd/` does not exist. **Zero transcripts.** `hooks/hooks.json` wires only
`UserPromptSubmit`, `SessionStart`, and `PreCompact`; `history.Capture` is built, redacts
on write, and **is called by nothing**. Pass B — mining chat for the rationale nobody wrote
down — has no corpus and cannot get one retroactively. **This is the only cost on the board
that is permanent and compounding**, which is why M1 is a hook and not a feature.

Related: [`02-disembark.md`](../brief/04-surfaces/02-disembark.md) sources the spine from
the native spec store, which holds **exactly one spec**. That is a fiction. The spine is the
intent corpus where one exists, and git where it does not.

Honesty consequence: Phase 6's acceptance bullet — *"a reader with no prior context can
reconstruct not just what the project is but why it was built the way it was"* — **cannot be
met by a transcript-less lifeboat as literally written.** It is re-authored to "the recorded
why, with every claim citing its source." Pass B ships as a declared exemption in
`_provenance.json`, never a silent gap.

## Blockers in the existing code

- **`surface_coverage` is a tripwire.** The rule asserts a `staged` registry row has **no**
  backing surface. Rows 2 and 3 of [`04-surfaces/README.md`](../brief/04-surfaces/README.md)
  are `staged`. Adding `commands/abcd/disembark.md` without flipping the row to `shipped`
  **and rewriting its prose** turns `make preflight` red. Must land in the same commit.
- **`fsutil.WriteFileAtomic` is unusable for embark's target writes.** It calls `os.MkdirAll`
  on the path's directory, so a manifest entry of `../../.ssh/authorized_keys` would create
  the directory and write the file. Embark needs `os.Root` containment — the pattern the
  privacy audit rule already adopted after a PoC showed a leaf-only `O_NOFOLLOW` is
  insufficient (a symlinked *intermediate* directory still escapes).
- **`launch.ResolveBundle` is the wrong thing to reuse for disembark.** Its `DenyNamespaces`
  structurally denies `.abcd` — exactly what disembark must read. **Mirror its safety
  disciplines** (symlink-escape, control-char, hardlink, duplicate rejection); **do not reuse
  the function.**
- **The Stop hook cannot call `history capture` directly.** From stdin the verb *requires*
  `--session <id>`, and a Stop hook receives its session id inside a JSON payload. It needs a
  new hidden verb (M1).

## Milestones

### M0 — Record work (lands before code; adr-33 and adr-34 require it)

1. **The mapping table** → `internal/core/lifeboat/mapping.go` as the single source of truth,
   rendered into [`00-meta.md`](../brief/00-meta.md). That file calls the table "the contract";
   the 2026-07-06 plan-consistency review found **it does not exist anywhere**. Reframed: it is
   the experiment's **hypothesis**, and M2 is expected to revise it.
2. **adr-35**, recording: the re-scope out of Phase 6; disembark is read-only and out-of-tree;
   voyage moves to the operator level (**amends adr-4**); the spine is re-sourced from the spec
   store to the intent corpus plus git; coverage is a first-class output; the graveyard is a new
   section; itd-2 is not a prerequisite.
3. **`DECISIONS.md` entry**, same day — ADRs graduate from it.
4. **Mint the owning intent.** Disembark and embark have **none**; itd-7, itd-8, itd-9, itd-11,
   itd-16, itd-23, and itd-35 are all satellites. `itd-88` in `planned/`, press-release-shaped,
   with acceptance framed as the experiment: *given a repo with no record, when I probe it, then
   I get a coverage report naming every brief section it cannot fill and what was searched.*
   Read [`itd-21`](../intents/drafts/itd-21-no-lifeboat-scaffolding.md) first — it may already
   own part of the greenfield path.
5. **Phase-index edit** ([`roadmap/phases/README.md`](../roadmap/phases/README.md) — per adr-33
   the *sole* ownership source), and rewrite phase-6's `## Dependency rationale` and
   `## Phase Acceptance` per the honesty consequence above.
6. **Register the verdict vocabulary.** `sufficient` is a member of no registered enum.
   **Reuse `{SHIP, NEEDS_WORK, MAJOR_RETHINK}`** — do not mint a third family.
7. `abcd intent plan itd-88` and `abcd spec create`. `intent_lifecycle` and `spec_lifecycle`
   are **blocker**-severity lint rules. Use our own tool.

### M1 — Start the transcript clock

The only irreversible item; ship it first so the corpus accrues during the rest of the build.

- New hidden verb `abcd hook session-end`, following the `hook prompt-router` pattern: hidden,
  fail-closed, **always exit 0**, never blocks the host. Reads the Stop-hook JSON payload from
  stdin, extracts `session_id` and `transcript_path`, calls `history.Capture` (which already
  redacts on write).
- `hooks/hooks.json`: add a `Stop` entry.
- Tests: payload parse; malformed or missing payload → exit 0 and no write; re-capture is
  idempotent.

*Gate:* finish a session; `abcd history list` shows a record; `~/.abcd/history/<root-sha>/` exists.

### M2 — `disembark probe <repo>` → the coverage report

**The experiment. No lifeboat is written.** Read-only, out-of-tree, no writes to the source.

**Tiered source adapters.** Every brief section is fed by adapters that degrade:

| Tier | Reads | Present in |
|---|---|---|
| **0 — Git** | commit history, authors, branches, reverts, file lifespans, tags, dependency churn | **every repo** |
| **1 — Conventions** | `README`, `docs/`, `CHANGELOG`, `LICENSE`, ADRs wherever they live, issue exports, `CONTRIBUTING` | most repos |
| **2 — abcd-native** | `.abcd/development/{decisions,intents,specs,brief,roadmap}`, `.abcd/work/{issues,reviews,DECISIONS.md}`, `.abcd/memory/` | this repo only |

```go
type Source interface {
    Section() Section
    Tier() Tier
    Probe(SourceContext) (Evidence, error)     // cheap; what material exists, and where
    Plan(SourceContext) ([]PlannedFile, error) // the full plan; ZERO writes (M3)
}
```

`Probe` runs every source's `Probe` concurrently. `Pack` is `Plan` plus a write. One code path,
so **`dry-run` can never lie** about what `to` will do — and a test asserts exactly that.

**The coverage schema** — stable, and it aggregates:

```json
{
  "schema_version": 1,
  "repo": {"name": "telemetry-cli", "root_sha": "...", "commits": 214},
  "tiers_present": ["git", "conventions"],
  "sections": [
    {"name": "product/context", "status": "grounded", "confidence": "high",
     "tier": "conventions", "evidence": ["README.md", "docs/why.md"]},
    {"name": "product/press-release", "status": "blank",
     "searched": ["README", "docs/", "git log", "CHANGELOG", "0 ADRs found"],
     "question": "What problem was this solving, and for whom? Nothing in the repo names a user."},
    {"name": "graveyard", "status": "grounded", "confidence": "medium", "tier": "git",
     "evidence": ["3 reverted commits", "2 unmerged branches", "src/engine-v1 deleted after 40 commits"]}
  ],
  "summary": {"grounded": 6, "partial": 2, "blank": 9}
}
```

Every grounded claim **cites its source file**. Every blank carries **what was searched** and
**the question a human must answer**. `status` is one of `grounded`, `partial`, `blank` — a
blank is a first-class result, not a failure.

Verbs: `abcd disembark probe <repo> [--json]`, and **`abcd disembark coverage <coverage.json>...`**
for the cross-repo aggregate — one table, section × repo. **That table is the artefact that
answers "is the brief structure sound."**

*Gate:* probe this repo and each messy repo; read the aggregate; **decide from data which brief
sections survive.** The M0 mapping table is a hypothesis, and this is where it meets evidence —
it is expected to change before M3 writes anything.

### M3 — `disembark <repo> to <dest>`

Only after M2's aggregate has settled the section list.

```
<dest>/
├── brief/                  # ONLY grounded sections, each citing its source
├── coverage.{json,md}      # gaps, what was searched, the questions  <- first-class
├── graveyard/              # M4
├── rescue/                 # spine: intents where they exist, git-derived where not
├── docs/adrs/              # verbatim, wherever they were found
├── activity/issues/
└── _provenance.json        # schema_version, source, tiers, manifest_sha256
```

Safety — the two cheap things that stop abcd destroying *our own* work (the hostile-lifeboat
battery is deferred; these lifeboats are ours):

- **Destination safety gate.** [`02-disembark.md`](../brief/04-surfaces/02-disembark.md) has
  none, and this is the rule that stops `disembark to ~/important-project` from backing up and
  overwriting a real directory. Refuse unless the destination is absent, an empty real directory,
  or an existing directory with a parseable `_provenance.json`. **Never overwrite a directory
  abcd did not produce.** Also refuse a symlinked destination, one inside `.git/`, or one that is
  an ancestor of the source.
- **Path validation on every write**: relative, cleaned, no `..`, no control characters.
- Write to a **staging directory**, then rename. A crash leaves staging, never a half-lifeboat.
- `_provenance.json` written **last** — it is the commit marker *and* the key to the gate above.
- **Secret-scan the planned bytes** before writing. Refuse on hard-fail; do **not** redact — a
  secret in a source file is a bug to fix at source, not to paper over in the artefact.
- **Never mutate the source repo.** A test hashes the source tree before and after.

**Pin the hash.** adr-4 asserts a chain but never defines it: `manifest_sha256` is SHA-256 over
the concatenation of `"<sha256>  <path>\n"` for every manifest entry, sorted lexicographically by
path, POSIX separators, LF only, with `_provenance.json` excluded (it cannot hash itself).

`voyage` appends to `~/.abcd/voyage/<source-root-sha>/history.jsonl` — genuinely append-only, not
a whole-file rewrite. Omit adr-4's `shared_with`: nothing produces it, and an empty field is a lie
in the schema.

**Strip the ahoy marker block on pack** (new export `ahoy.StripMarkerBlock`). Carrying a stale
`BEGIN ABCD` block into a new repo plants a stale rules-loader. The design never says this, and
it must.

Ship `commands/abcd/disembark.md` **and flip surface-registry row 2 to `shipped` in the same
commit** (see the tripwire above).

### M4 — The graveyard

Three layers, strictly in this order:

1. **Archaeology** (Tier 0, deterministic, any repo): reverted commits; branches unmerged into
   the default branch, ranked by divergence age; files and directories deleted after substantial
   history; dependencies added then removed; wholesale rewrites. Output: `graveyard/archaeology.json`
   — **evidence only, no interpretation.**
2. **Recorded abandonment** (Tier 1/2): superseded ADRs and intents, `wontfix` issues,
   `## Alternatives Considered`, rejected options named in decision logs. Output:
   `graveyard/abandoned.json` — **what the project explicitly declared dead.**
3. **Interpretation** (host-delegated LLM): reads layers 1 and 2, produces `graveyard/lessons.json`.
   **Every entry MUST carry an `evidence[]` array citing layer-1 or layer-2 ids. An entry with no
   evidence is dropped by the Go validator** — not by the model's good intentions. Per-entry
   confidence; low-confidence entries land in `graveyard/low-confidence/` rather than the main
   file (itd-11's mechanism).

The validator enforcing cite-or-be-dropped is the whole discipline. It is small, and it is the
difference between a graveyard and a séance.

### M5 — `embark` and the round-trip

- `abcd embark probe <path>` — read-only: schema version (**refuse anything greater than known**,
  with the upgrade message — itd-9's only irreversible half; migrators wait), manifest
  reconciliation, path validation, target-emptiness, conflict detection.
- `abcd embark from <path>` — the write path, over `os.Root` containment plus independent path
  validation. Two layers, so a bug in one is not a CVE.
- **Conflicts: one bulk prompt**, not a per-file barrage. Design correction:
  [`03-embark.md`](../brief/04-surfaces/03-embark.md) has a refusal *write* a conflict report;
  core must not write on a refusal path (transport-agnostic-core violation). **Core returns the
  conflicts; the surface renders them.**
- **The coverage report is what embark hands the product thinker.** Scaffold the grounded brief;
  surface the blanks and their questions as the first thing they see.
- **Never inject lifeboat text verbatim into `CLAUDE.md`.** Re-inject the *current* marker block.
  That line is the difference between a data leak and a persistent instruction implant — lifeboat
  content otherwise lands in a file the agent obeys.

**Round-trip property test** — assert through the real stores, not by diffing bytes: `intent.Load`,
`spec.Load`, and `capture.List` on the target equal the source; ADRs byte-identical; the target's
`CLAUDE.md` carries the *current* marker block; the source tree is byte-identical after the pack.
Then the **closure property**: re-packing an embarked repo reproduces the same verbatim manifest
hash. That is the gate, and it is a property, not an eyeball check.

### M6 — Synthesis over the written record

Each LLM output is an injected function type fed by a `--*-json <file|->` flag, mirroring the
shipped `memory.Distiller` seam, with `intent review ingest`'s dead-letter handling for malformed
model output. **No agents supplied = deterministic mode**; `_provenance.json` records which.

| Agent | Reads | Writes |
|---|---|---|
| `principle-distiller` | ADR consequences, decision logs, resolved issues, reviews | `principles.json` |
| `graveyard-interpreter` | `archaeology.json` + `abandoned.json` | `lessons.json` (**cite-or-dropped**) |
| `press-release-composer` | brief, spine, principles | `press-release.json` |
| `lifeboat-oracle` | packed lifeboat vs. source repo | `audit/oracle-<ts>.json`, existing verdict enum |

Orchestration lives in `commands/abcd/disembark.md` — the host agent dispatches the subagents,
writes their JSON, calls the binary. Exactly as `commands/abcd/memory.md` already says: *you are
the distiller*. **Four agents, not the design's fourteen** — fourteen will not fit one host context.

Each agent needs `prompt_version`, `reads_untrusted_input: true`, and a canary fixture per itd-5.
**None of that infrastructure exists**; this is where it ships.

## Deliberately deferred

The hostile-lifeboat battery; itd-35's Merkle chain (**integrity is not trust** — a hash proves the
sender's bytes are intact, not that the sender is benign, so the containment work is what matters);
`--with-code` (itd-8); schema migrators (itd-9 beyond the version stamp); itd-7 RepoPrompt
portability; backgrounded execution.

## Acceptance

**M2 is self-verifying and is the point.** If the cross-repo aggregate shows the brief structure
holds — most sections grounded on rich repos, the blanks consistent and answerable on poor ones —
the premise is real, and the structure earns its place for repos abcd manages from the start. If
half the brief is permanently blank on every repo, **the structure is wrong, and we have learned
that for the cost of one milestone, before writing a packer.**

`make preflight` clean throughout.
