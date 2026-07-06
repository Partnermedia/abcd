# Per-Agent Prompting SOTA Research

> Per the brief (§ "Research-driven prompts"): per-agent SOTA research is **task #1 of each agent's epic**. The output lives here as `<agent-name>.md` and references the baseline at [`../01-general-best-practices.md`](../01-general-best-practices.md).
>
> **Role.** Same as the baseline: research is the **gate** (audit reference for `lifeboat-oracle` and the prompt linter), not the **source** (prompt template). The author writes the agent's prompt informed by the research; the oracle audits alignment.

## Purpose of these files

For each of the 14 abcd agents, `<agent-name>.md` answers four questions specific to that agent's job:

1. **What is the closest prior art?** Existing prompts in public repos (Piebald, VoltAgent, EliFuzz, Cursor / Devin extracts), academic papers on the agent's task type, and any internal predecessors (e.g. the manual lifeboat for `flow-essence`).
2. **What are the agent-specific failure modes?** General failures from `01-general-best-practices.md` plus things that bite *this* agent specifically (e.g. context rot for `chat-distiller`, verbosity bias for `press-release-composer`, injection for `embark-scaffolder`).
3. **What techniques from the SOTA taxonomy fit this agent?** Named techniques from The Prompt Report (CoT, self-consistency, ReAct, plan-and-solve, etc.) with a *one-line* justification per technique.
4. **What golden-test fixtures should this agent have?** 2–5 input/output pairs that span the behaviour envelope, plus injection canaries where the threat model demands.

## File shape

Use `_template.md` as the starting point. Every per-agent research file MUST include:

- Frontmatter: `name`, `description` (one-line scope), `agent: <agent-name>`, `baseline: 01-general-best-practices.md`
- The four sections above (closest prior art / failure modes / SOTA techniques / fixture sketches)
- A short "Open questions" tail capturing things the author could not resolve from desk research and which need spike work or human judgement during the agent's epic

## When to update

- **Initial creation**: task #1 of each agent's flow-next epic.
- **Material drift**: when the baseline doc (`01-…`) supersedes to `02-…`, the periodic SOTA-audit oracle (B+C+D infrastructure, D component) flags any per-agent file whose pillars (failure modes, techniques, fixtures) are now inconsistent with the new baseline.
- **Post-launch incidents**: if a production failure traces back to a missing pitfall here, append a "Lessons learned" section.

Per-agent files are **immutable once their epic ships**. Subsequent additions go in a new section dated at the bottom of the file or, if structural, supersede with a new file (`<agent-name>-02.md`). Mirrors how the brief itself archives.

## Inventory (target: 14 files)

| Agent | Pass | Highest agent-specific risk | File |
|---|---|---|---|
| `flow-essence` | A | Spec staleness in newest-first ordering | TBD |
| `decision-archaeologist` | A | ADR / git-log / CLAUDE.md cross-source synthesis | TBD |
| `review-collator` | A | Format drift across model-emitted reviews | TBD |
| `chat-distiller` | B | **Context rot** (highest in the suite) | [present](chat-distiller.md) |
| `principle-distiller` | C | Domain-grouping bias; missed-rationale gaps | TBD |
| `artefact-curator` | C | Keep/adapt/drop classification accuracy | TBD |
| `brief-composer` | C | Coherent-narrative-from-fragments | TBD |
| `press-release-composer` | C | **Verbosity bias** (highest in the suite) | TBD |
| `lifeboat-oracle` | C | LLM-judge biases; verdict-tag drift | TBD |
| `code-rescuer` | opt-in | Principle-extraction without code-level recommendation | TBD |
| `issue-scout` | opt-in | GitHub-search precision/recall | TBD |
| `embark-scaffolder` | embark | **Prompt injection** (highest in the suite); idempotent placement | [present](embark-scaffolder.md) |
| `launch-gatekeeper` | launch | PII regex completeness; false-negatives | TBD |
| `intent-fidelity-reviewer` | post-ship | LLM-judge bias against own intent author | TBD |

Replace "TBD" with `[present](<agent-name>.md)` as each agent's epic kicks off.
