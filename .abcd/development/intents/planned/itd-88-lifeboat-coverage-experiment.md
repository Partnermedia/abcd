---
id: itd-88
slug: lifeboat-coverage-experiment
spec_id: spc-3
kind: standalone
suggested_kind: null
reclassification_history: []
builds_on: []
severity: major
related_adrs: [adr-35]
---

# Know What a Repo Can and Cannot Tell You — Before You Trust the Lifeboat

## Press Release

> **abcd ships `/abcd:disembark probe <repo>` — point it at any repository and get back an honest account of what can be rescued from it.** The probe reads a repo without touching it and reports, section by section, what a lifeboat could actually ground: which parts of the project's theory are written down somewhere, which are only half there, and which are simply absent. Every grounded claim cites the file it came from. Every blank carries what abcd searched and **the question a human has to answer instead**. Run it across several repos and `/abcd:disembark coverage` puts them in one table — so you can see, in a number, what keeping a record is worth.
>
> "I inherited a repository and three sentences of context," said Iris, product lead. "The probe gave me a list of exactly what the repo could not tell me — and a question for each one. That list *is* my first week. I stopped guessing what I didn't know."
>
> "I mine dead projects for a living, and the danger is always the confident summary that turns out to be invented," said Jack, consultant. "This one refuses to fill a section it can't cite. When it says blank, it means blank — and it shows me where it looked."

## Why This Matters

The lifeboat is what abcd's press release says abcd *is* — the surface that makes a project survivable. It is also the only major surface with **zero code**, and it has **no owning intent**: itd-7, itd-8, itd-9, itd-11, itd-16, itd-23 and itd-35 are all satellites orbiting an intent that was never written. This is it.

It matters *now*, and as an experiment rather than a feature, because **the brief's structure is an assumption nobody has tested**. The brief asserts that a project's theory decomposes into these sections. Build the packer first and you have assumed the answer — you will get a lifeboat shaped like the brief whether or not the brief is a shape reality fits.

So invert the build. **Probe before pack.** Run the same probe over a repo with a rich record and a repo with nothing but git, and the delta in section coverage is **what the record is worth**. That number is the point. If it is large, abcd's premise — that writing things down is the thing worth doing — survives contact with reality. If half the brief comes back permanently blank on every repository, **the structure is wrong, and we have learned that for the cost of one milestone instead of a whole phase.**

The honesty discipline is the other half. A rescue tool that invents a plausible section is worse than one that leaves it empty, because the reader cannot tell the difference. A blank is therefore a **first-class result**, not a failure: it names what is missing, what was searched for it, and the question a human must answer.

## What's In Scope

- **`abcd disembark probe <repo>`** — read-only, out-of-tree. Reports per-section coverage: `grounded` / `partial` / `blank`, with confidence, the tier it was grounded from, and the evidence cited.
- **Tiered source adapters that degrade**: Tier 0 git (every repo), Tier 1 conventions (`README`, `docs/`, `CHANGELOG`, `LICENSE`, `CONTRIBUTING`, ADRs wherever they live), Tier 2 abcd-native (`.abcd/`).
- **`abcd disembark coverage <coverage.json>...`** — the cross-repo aggregate: one table, section × repo. **This table is the artefact that answers whether the brief structure is sound.**
- **A stable, aggregatable coverage schema** — `schema_version`, per-section status, evidence, what was searched, and the question a blank raises.
- **The graveyard as a first-class section** — archaeology (Tier 0), recorded abandonment (Tier 1/2), then interpretation that **must cite the evidence beneath it or be dropped by the validator**.
- The packer (`disembark <repo> to <dest>`), embark, and the round-trip — built *after* the aggregate settles the section list.

## What's Out of Scope

- **Greenfield scaffolding** — starting a new project with abcd conventions and no lifeboat to embark from is **itd-21**'s (`/abcd:init-project scaffold`). This intent never scaffolds; it only reads existing repositories.
- **Mining chat transcripts for unrecorded rationale (Pass B)** — there is no corpus. It ships as a **declared exemption** in `_provenance.json`, never a silent gap, until the transcript store has data.
- The hostile-lifeboat threat battery, the Merkle chain (itd-16), `--with-code` (itd-8), schema migrators beyond the version stamp (itd-9), RepoPrompt portability (itd-7), and backgrounded execution — all deliberately deferred.
- Writing anything into the source repository, ever.

## Acceptance Criteria

> _BDD format, per `itd-1-acceptance-gates`. These gates are checked by `intent-fidelity-reviewer` when this intent moves to `shipped/`._

- **Given** a repository with no abcd record at all (git and nothing else), **when** the user runs `abcd disembark probe <repo>`, **then** a coverage report is produced naming **every** brief section it cannot fill, each blank carrying **what was searched** and **the question a human must answer** — and the run never fails merely because the repo is poor.
- **Given** any repository, **when** `probe` runs against it, **then** the source tree is **byte-identical afterwards** — a test hashes it before and after — and abcd writes nothing inside it.
- **Given** a section reported `grounded`, **when** the report is read, **then** every claim in it **cites the source file it came from**; a claim with no citation is a defect, not a stylistic lapse.
- **Given** coverage reports from several repositories of mixed record quality, **when** the user runs `abcd disembark coverage <coverage.json>...`, **then** one table renders section × repo, and the **delta between a rich-record repo and a git-only repo is legible as a number**.
- **Given** the same repository probed twice with no changes between runs, **when** the two reports are compared, **then** they are identical — the probe is deterministic, so a delta in the aggregate means a delta in the repos, never in the tool.
- **Given** a repository whose richest tier is Tier 0, **when** it is probed, **then** the `graveyard` section is still `grounded` — what a project abandoned is recoverable from git history alone (reverts, branches abandoned unmerged, files deleted after substantial history, dependencies added then removed).
- **Given** an interpreted graveyard entry that cites no layer-1 or layer-2 evidence id, **when** the Go validator runs, **then** the entry is **dropped** — cite-or-be-dropped is enforced by code, not by the model's good intentions.
- **Given** the packer exists (a later milestone), **when** `dry-run` and `to <dest>` are compared, **then** they report the same planned writes — one code path, so **`dry-run` cannot lie about what `to` will do**, and a test asserts exactly that.

## Prior Art

- **[adr-35](../../decisions/adrs/0035-lifeboat-as-coverage-experiment.md)** ratifies this intent's shape: probe before pack, read-only and out-of-tree, coverage as a first-class output, the graveyard as a new section, `voyage/` at the operator level (superseding adr-4).
- **Satellites of this intent, all pre-existing**: itd-11 (Pass B pitfall mitigation — its low-confidence-quarantine mechanism becomes the graveyard's cite-or-be-dropped validator), itd-35 (lifeboat integrity audit), itd-16 (hash-chain/Merkle audit), itd-8 (`--with-code`), itd-9 (schema migrators), itd-7 (RepoPrompt workspace portability), itd-23 (Spec Kit interop). Each assumed an owning intent that did not exist.
- **itd-21 (`no-lifeboat-scaffolding`)** was read first to check for overlap: it owns the **greenfield** path (scaffold a *new* project with no lifeboat to embark from). It does not overlap — this intent only ever reads repositories that already exist. The boundary is recorded in What's Out of Scope.
- **Naur (1985), *Programming as Theory Building*** — the recovery-humility frame the lifeboat already cites: the theory lives in the people, and an artefact is a proxy for it. This intent takes that seriously enough to *measure* the proxy's coverage rather than assert it.
- **`repolinter` and Conftest** are the closest external analogues already adopted in this repo (by itd-85's audit engine): a rule set producing a per-rule verdict over a repository. Coverage differs in that a failing rule there is a defect, whereas a blank here is **information** — the question a human must answer.

## Open Questions

- Which brief sections survive the cross-repo aggregate? `product/personas` is predicted blank below abcd-native and only partial there; if that holds across the corpus, the section is not derivable from a repository at all and should not be in a lifeboat's brief. **This is the question the experiment exists to answer, and it is deliberately open.**
- How many repositories does the aggregate need before the delta is trustworthy rather than anecdotal?
- Does `partial` earn its place as a status, or does it become a bucket where every hard call goes to die?

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
