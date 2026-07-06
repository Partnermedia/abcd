# Docs & design-record information architecture — SOTA

Research deliverable backing the record-IA choice. **Question:** how should abcd
organise its two documentation surfaces — the durable *design record* under
`.abcd/development/` and the *user-facing docs* under `docs/` — so that facts have
one canonical home, cross-references stay stable, and the surfaces do not rot as
the project grows?

The findings here ratify [ADR-30](../../decisions/adrs/0030-record-information-architecture.md)
(flat artefact-type folders; Diátaxis for user docs; generated CLI reference; a
numbering split between stable ADR handles and date-prefixed chronological
artefacts).

---

## TL;DR

The state of the art is settled and unglamorous: **Diátaxis** for user docs (as an
authoring lens, not a rigid four-folder sidebar), **MADR with sequential `NNNN`**
for decisions, **generated CLI reference** from the command tree, and a **hard
audience split** between contributor records and user docs — validated by how large
open-source CLIs (e.g. GitHub's `gh`) partition their documentation. The single
biggest differentiator against rot is a **CI gate on structural violations**, not
any particular folder shape.

---

## Findings

Ranked, each with an evidence tier: **[CONSENSUS]** = widely-agreed community
standard; **[EVIDENCE]** = observed in a concrete, inspectable project.

### 1. Diátaxis is the standard docs framework — as a lens, not a cage · [CONSENSUS]

The four-mode split — **tutorials** (learning), **how-to** (task),
**reference** (information), **explanation** (understanding) — is the de-facto
standard for structuring user documentation. The value is the **one-type-per-page**
discipline: it names the failure mode (boundary drift, where how-tos silently
accrete explanation and reference pages grow tutorials) and gives a gate to prevent
it. It is best applied as an *authoring lens* — task-shaped navigation is fine — not
as a mechanical four-folder sidebar that forces awkward nav and a filing backlog.

- https://diataxis.fr/ — the canonical framework.
- Adopted by Canonical/Ubuntu and widely cited as a documentation foundation.

### 2. Separate user docs from design records — the audience split · [EVIDENCE]

Contributor-facing records (decisions, plans, research) and user-facing docs serve
different readers and belong in different trees. Large CLIs demonstrate the split:
GitHub's `gh` reserves its in-repo `docs/` tree for a specific audience and ships
end-user documentation generated off the command tree. For abcd this validates the
`.abcd/` design-record vs `docs/` user-docs boundary **because of the publish
pipeline** — mixing records into `docs/` would force filtering on every `launch`.
This deliberately *overrides* the more common "ADRs live under `docs/decisions/`"
convention, which is correct for most projects but wrong here given that `docs/`
ships to users and `.abcd/**` does not.

- https://github.com/cli/cli/tree/trunk/docs — the audience-scoped in-repo tree.

### 3. Generate the CLI reference from the command tree · [EVIDENCE]

Hand-maintained per-command reference drifts from the actual commands. The SOTA is
to **generate** it: Cobra exposes `doc.GenMarkdownTree` / `GenManTree`, which walk
the command tree and emit one page per command. Segregate the output
(`docs/reference/cli/`), treat it as **build output** (never hand-edited), and add a
**CI freshness check** that regenerates and diffs. This is the same mechanism Hugo's
`gen doc` and kubectl's generated reference use.

- https://cobra.dev/ — `doc.GenMarkdownTree` / `GenManTree`.
- https://gohugo.io/commands/ — Hugo's generated command docs.
- kubectl's generated command reference — the same generate-and-check pattern.

### 4. ADRs: MADR, sequential `NNNN`, one file per decision · [CONSENSUS]

The community standard for architecture decision records is **MADR** (Markdown Any
Decision Records) with **sequential zero-padded `NNNN` filenames**, one file per
decision, a status field, and **superseded-by links on both files**. The sequential
number is the point: it is a **stable, order-free cross-reference handle** that other
documents link to and that never moves when the record is reorganised. This is
distinct from *chronological* artefacts (plans, research notes, an append-only
decision log), which should be **date-prefixed** because they are read newest-first
and are never cross-referenced by a stable handle.

- https://adr.github.io/ and https://github.com/adr/madr — the MADR standard.
- https://martinfowler.com/bliki/ArchitectureDecisionRecord.html — the original ADR
  framing.

### 5. Rot prevention — the CI gate is the differentiator · [CONSENSUS]

Ranked, the practices that keep a docs tree from rotting:

1. **One canonical home per fact** — link, never copy (the single most important
   rule; duplication is where drift starts).
2. **Shallow tree, stable categories** — deep or frequently-renamed folders break
   cross-references and raise the "where does this go?" cost.
3. **A CI gate on structural violations** — stray files, broken links, mixed
   Diátaxis type, present-tense-only, and generated-reference freshness. This is the
   biggest differentiator: without enforcement, every other rule decays.
4. **Segregate generated content** so it is never hand-edited.
5. **Scheduled rot removal + present-tense-only prose** (docs state what *is*, not
   what changed).

---

## Rejected (investigated)

- **Date-prefixed ADR filenames** — loses the stable cross-reference handle that is
  an ADR's whole value.
- **A rigid mechanical four-folder Diátaxis sidebar** — awkward navigation and a
  filing backlog; Diátaxis is a lens, not a mandatory folder layout.
- **Hand-maintained per-command CLI reference** — drifts from the command tree;
  generate it instead.
- **ADRs under `docs/`** — correct for most projects, wrong for abcd because `docs/`
  ships to users via `launch` and the design record must not.
- **An input/process (or durable/working) meta-grouping inside `development/`** —
  double-classifies against the `development/` ↔ `work/` ↔ `.work.local/` tiering
  that already carries the durability axis.
- **DITA / Information-Mapping** — heavyweight enterprise frameworks, disproportionate
  for a single-binary CLI.

---

## Recommendation

Adopt exactly what [ADR-30](../../decisions/adrs/0030-record-information-architecture.md)
records:

- **Design record** — flat artefact-type folders under `.abcd/development/`
  (`brief/`, `intents/`, `principles/`, `decisions/`, `roadmap/`, `plans/`,
  `research/`), one canonical home per concept, with a `README.md` map.
- **User docs** — Diátaxis under `docs/`, one type per page, with the CLI reference
  generated from the Cobra command tree into `docs/reference/cli/` and
  CI-freshness-checked.
- **Numbering split** — ADRs use stable sequential `NNNN`; plans, research notes, and
  the Tier-2 `DECISIONS.md` log use a date-prefix.
- **Enforcement** — a deterministic doc-lint gate (stray root markdown, broken links,
  mixed Diátaxis type, present-tense-only, generated-reference freshness) is the
  load-bearing anti-rot mechanism, complementing a periodic semantic fidelity read.

## Sources

- https://diataxis.fr/
- https://ubuntu.com/blog/diataxis-a-new-foundation-for-canonical-documentation
- https://github.com/cli/cli/tree/trunk/docs
- https://martinfowler.com/bliki/ArchitectureDecisionRecord.html
- https://adr.github.io/ · https://github.com/adr/madr · https://adr.github.io/madr/
- https://cobra.dev/ · https://gohugo.io/commands/
- https://docs.readthedocs.com/ — documentation-structure guidance
