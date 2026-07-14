---
id: adr-35
slug: lifeboat-as-coverage-experiment
status: accepted
date: 2026-07-14
supersedes: adr-4
superseded_by: null
related_intents: [itd-88, itd-11, itd-2, itd-21]
related_rfcs: []
related_adrs: [adr-1, adr-5, adr-25, adr-28, adr-29, adr-33]
---

# ADR-35: The lifeboat is a coverage experiment — read-only, out-of-tree, and proven before it is packed (supersedes adr-4)

## Context

The lifeboat is what abcd's press release says abcd *is*: the artefact that makes a project survivable. It is also the only major surface with **zero code** — no `internal/core/lifeboat/`, no verbs, no agent prompts. The roadmap parks it at Phase 6, last, on the stated grounds that it "depends on every prior substrate being native."

Checked against the binary rather than the record — per `CONTEXT.md`'s rule that *where the record and the binary disagree, verify against the binary* — **that rationale is mostly false, and the error is load-bearing**:

| The roadmap says it blocks | What the binary shows |
|---|---|
| Phase 4 native spec engine | `spec.Load` / `Create` / `Close` / `Validate` all ship. Not a blocker. |
| Phase 5 run seam | Backgrounding is a host affordance; the design's own checkpoint is "forensic only — abcd ships no `resume` verb". Not a blocker. |
| Phase 3 reviews (itd-28) | Reviews are already committed markdown under `.abcd/work/reviews/`. Not a blocker. |
| itd-2 in-session subagent dispatch | The host-delegation seam already ships **twice**: `memory.Distiller` fed by `--pages-json`, and `intent review ingest --verdict-json` with a dead-letter path. Not a blocker. |
| Phase 2 history + memory | The *packages* are built. The **data is empty**. Real — and not a code problem. |

So the sequencing argument does not hold. Three further facts forced this decision:

1. **The premise is untested.** The brief's structure is an assumption: that these are the sections a project's theory decomposes into. Building a packer first assumes the answer. Nobody has ever run the extraction against a repository that *has no record* to see which sections come back empty.
2. **adr-4 put the operations namespace inside the source repo** (`.abcd/development/voyage/`). Voyage records absolute source paths; the `privacy-hygiene` audit rule (itd-85) flags `/Users/<name>/` in committed files. abcd would have failed its own audit.
3. **`~/.abcd/` does not exist. There are zero transcripts.** `history.Capture` is built, redacts on write, and **is called by nothing** — no `Stop` hook is wired. Pass B (mining chat for the rationale nobody wrote down) has no corpus and **cannot get one retroactively**. This is the only cost on the board that is permanent and compounding.

## Decision

**The lifeboat is built as an experiment whose first output is a verdict on the brief's own structure, not a lifeboat.**

1. **Probe before pack.** `abcd disembark probe <repo>` produces a **coverage report** — which brief sections a real repository can ground, which are blank, and what was searched — across a corpus of repos of mixed record quality, **before a packer exists at all**. Running the same probe over a rich-record repo and a git-only repo yields a delta in section coverage, and **that delta is what the record is worth**. If half the brief is permanently blank on every repository, the structure is wrong and we have learned it for the cost of one milestone.

2. **Disembark is read-only and out-of-tree.** `abcd disembark <source-repo> to <dest>` — point it at any repository, touch nothing, write elsewhere. A test hashes the source tree before and after. This replaces adr-4's model of a lifeboat written back into the source at `.abcd/lifeboat/`: mining a dead or archived project must not require `ahoy install` into a repo we only want to read.

3. **Voyage moves to the operator level** — `~/.abcd/voyage/<source-root-sha>/`, keyed on the root-commit SHA, matching the existing history-store convention. **This is the clause that supersedes adr-4**, which defined voyage as the operations namespace *inside* the source repo. It is therefore never committed, which dissolves the `privacy-hygiene` collision above.

4. **The lifeboat remains regenerable output** (adr-4's surviving core, restated here because supersession prunes the original): it is the latest snapshot, not an accumulating archive; there is never a `lifeboat-v1/` / `lifeboat-v2/`; history lives in the append-only voyage log, not in stale snapshots. The **naming distinction stays load-bearing** — `lifeboat/` is the artefact (noun; what gets carried), `voyage/` is operations (verb; what we did to produce it). What changes is *where each lands*: the lifeboat at an operator-chosen destination, voyage at the operator level. adr-4's overwrite-in-place-with-`.bak` model is replaced by a **destination safety gate** — refuse unless the destination is absent, an empty directory, or one carrying a parseable `_provenance.json`. **Never overwrite a directory abcd did not produce.**

5. **The hash chain is pinned.** adr-4 asserted a chain but never defined it. `manifest_sha256` is SHA-256 over the concatenation of `"<sha256>  <path>\n"` for every manifest entry, sorted lexicographically by path, POSIX separators, LF only, with `_provenance.json` excluded (it cannot hash itself). adr-4's `shared_with` field is dropped: nothing produces it, and an empty field is a lie in a schema.

6. **Coverage is a first-class output**, not an exemption footnote. The brief carries only what abcd could ground, each claim citing its source; a separate `coverage.{json,md}` carries what is missing, what was searched, and the question a human must answer. Its schema **aggregates across repositories**, because that aggregate *is* the experiment's readout. `status` is one of `grounded` / `partial` / `blank`, and **a blank is a first-class result, not a failure**.

7. **The graveyard is a new first-class section** — "extract what failed, so it isn't tried again" — in three layers, and the order is the anti-fiction discipline: **archaeology** (Tier 0, any repo: reverted commits, branches abandoned unmerged, files deleted after substantial history, dependencies added then removed), then **recorded abandonment** (superseded ADRs and intents, `wontfix` issues, alternatives-considered sections), then **interpretation** (host-delegated LLM). Interpretation sits at the bottom and cannot float free: **every interpreted entry must carry an `evidence[]` array citing layer-1 or layer-2 ids, and an entry with no evidence is dropped by the Go validator** — not by the model's good intentions. This is the mechanism itd-11 was reaching for, made structural.

8. **The spine is re-sourced.** The design sources it from the native spec store, which holds **exactly one spec** — a fiction. The spine is the **intent corpus** where one exists, and **git** where it does not.

9. **The lifeboat-oracle reuses the registered review-verdict enum** `{SHIP, NEEDS_WORK, MAJOR_RETHINK}`. The brief's "sufficient" verdict was a member of no registered enum; no third verdict family is minted.

10. **itd-2 is not a prerequisite**, per the table above.

11. **Phase 6's fidelity claim is re-authored.** *"A reader with no prior context can reconstruct not just what the project is but why it was built the way it was"* **cannot be met by a transcript-less lifeboat as literally written.** It becomes "**the recorded why, with every claim citing its source**", and Pass B ships as a **declared exemption in `_provenance.json`, never a silent gap**.

## Alternatives Considered

1. **Build the packer first, as the design is written** (Phase 6, adapters → Pass A → Pass B → Pass C → embark). Rejected: it assumes the brief's structure is right, which is the very thing in question, and it spends the most expensive milestone on that assumption. The probe answers it for a fraction of the cost, and the packer is strictly better for having the answer.

2. **Amend adr-4 in place** (append an `## Amendment` section, as adr-9 does). Rejected: adr-9's amendment refined *how* one section was expressed. Here, two of adr-4's three operative claims change — where the lifeboat lands, and where voyage lives — and its overwrite model is replaced by a safety gate. That is a replacement, not a clarification, and the record's convention is that a replacement gets a new id and prunes the original.

3. **Keep voyage in the source repo and gitignore it.** Rejected: a gitignored in-tree operations log still lives in a tree we have promised not to touch (decision 2), and it duplicates a convention abcd already has — the history store keyed on root-commit SHA under `~/.abcd/`. One operator-level home, one keying rule.

4. **Keep disembark cwd-based** (`disembark` packs *this* repo). Rejected: it forces `ahoy install` into a repository we only want to read, and it is backwards for the primary use case — mining a project that is already dead.

5. **Mint a new verdict enum for the lifeboat oracle** (e.g. `{sufficient, insufficient}`). Rejected: `{SHIP, NEEDS_WORK, MAJOR_RETHINK}` is registered and means exactly this. A third family is vocabulary drift.

## Consequences

**Gains:**

- The premise gets tested before it gets built. The cross-repo coverage aggregate is a number a human can read and act on, and it can only say "the brief structure is wrong" *before* a packer has been written to that structure.
- Disembark can be pointed at any repository — including dead ones, archived ones, and ones abcd has never touched — without modifying them.
- The `privacy-hygiene` collision is dissolved rather than exempted; abcd stops being at risk of failing its own audit rule.
- A blank section becomes a *question a human must answer*, carried in the artefact, instead of a silent omission.

**Costs / obligations:**

- **The transcript clock must start now.** Pass B's corpus is unrecoverable for every session that runs before a `Stop` hook is wired. This ADR does not fix that; it names it as the reason the hook ships first, ahead of any lifeboat code.
- **Vocabulary must be registered** in `02-constraints/04-naming.md`: `voyage/`, `manifest_sha256`, `_provenance.json`, `history.jsonl`, `coverage.json`, `graveyard/`, plus the two new enums (`status ∈ {grounded, partial, blank}` and `tier ∈ {git, conventions, abcd-native}`). adr-4 *claimed* this registration and never made it — the terms were absent from the registry. This ADR fulfils the claim.
- **The mapping table is a hypothesis and is expected to lose.** It is generated from `internal/core/lifeboat/mapping.go` and rendered into the brief's `00-meta.md`, which has called it "the contract" while no such table existed anywhere. The probe is what will revise it.
- **`surface_coverage` is a tripwire.** Rows 2 and 3 of the surface registry are `staged`, and that rule asserts a `staged` row has *no* backing surface. Adding `commands/abcd/disembark.md` without flipping the row to `shipped` in the same commit turns `make preflight` red.
- **`fsutil.WriteFileAtomic` is unusable for embark's target writes** — it calls `os.MkdirAll` on the path's directory, so a manifest entry of `../../.ssh/authorized_keys` would create the directory and write the file. Embark needs `os.Root` containment, the pattern the `privacy-hygiene` rule already adopted after a proof-of-concept showed a leaf-only `O_NOFOLLOW` is insufficient (a symlinked *intermediate* directory still escapes).
- **`launch.ResolveBundle` must not be reused for disembark** — its `DenyNamespaces` structurally denies `.abcd`, which is exactly what disembark must read. Mirror its safety disciplines (symlink-escape, control-character, hardlink, duplicate rejection); do not reuse the function.

**Downstream decisions enabled:**

- The section list that survives the coverage aggregate is what the packer is built to — so M3 onward has an evidence-backed target schema instead of an assumed one.
- The graveyard's cite-or-be-dropped validator generalises: it is the shape any host-delegated synthesis over an untrusted corpus should take.
