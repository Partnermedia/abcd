# SOTA survey — interrogation and fresh-eyes review instruments

Dated 2026-08-19. Compiled from two host-run web research passes; every source
below was fetched during the runs. Evidence tiers per
[`prefer-sota`](../../principles/prefer-sota.md): each pick is named as the
presumptive SOTA and then challenged for fit against this repo's conventions.

**Why this note.** Three record surfaces under design need positioning against
current practice: the cold reading ([itd-86](../../intents/drafts/itd-86-cold-reading-surface.md)),
the grill (itd-27 / itd-42, planned), and any future audit that reads the
shipped repository against the brief's stated intent. The survey asked four
questions: who does Socratic interrogation of plans, who enforces (rather than
instructs) reviewer blindness, who detects divergence between shipped state and
stated intent, and who reports findings without resolving them.

## 1. Socratic interrogation of plans (the grill's neighbours)

1. **`grill-me` / `grilling` — Matt Pocock**
   ([mattpocock/skills](https://github.com/mattpocock/skills), MIT; mechanics in
   the [grilling SKILL.md](https://raw.githubusercontent.com/mattpocock/skills/main/skills/productivity/grilling/SKILL.md)).
   Maintained repo, primary source fetched. The mechanic: the agent maintains a
   **frontier of a design tree**; each round it asks every question whose
   prerequisites are settled, numbered, each carrying a **recommended answer**;
   the session ends when the frontier is empty — "every branch of the design
   tree visited, nothing left silently assumed." Fact-finding is the agent's
   job, never the user's (sub-agents dispatched for lookups). A sibling variant
   writes orientation docs and decision records.
   *Fit challenge:* maximal-context by design — an alignment instrument, and the
   per-question **recommended answer is a recommending act**. Any abcd use whose
   contract forbids recommending (assessment without selection) must strip that
   element; the frontier/empty-frontier mechanic itself is adoptable. If the
   mechanic is adopted, the `ACKNOWLEDGEMENTS.md` entry lands in the same
   change, per the standing rule.
2. **`brainstorming` — Superpowers, Jesse Vincent**
   ([obra/superpowers](https://blog.fsck.com/2025/10/09/superpowers/)).
   One question at a time; proposes 2–3 approaches; hard gate — no
   implementation until the user approves a presented design. Maintained repo +
   author write-up. *Fit:* full context; produces an approval gate (a verdict).
3. **GitHub Spec Kit `/speckit.clarify`**
   ([github/spec-kit](https://github.com/github/spec-kit), official org).
   Structured clarifying questions that insert `[NEEDS CLARIFICATION]` markers
   into the spec. *Fit:* mutates the artefact it interrogates.
4. Marketplace derivatives and aggregator listings repeat the grill-me
   mechanics; their install counts carry no methodology. Anecdote tier.

## 2. Enforced vs instructed reviewer blindness

The load-bearing distinction in current practice: only three mechanisms found
**enforce** blindness; everything else instructs it in a prompt, and documented
self-preference bias makes instructed-only isolation the weak form.

1. **Host subagent context isolation**
   ([Claude Code subagent docs](https://code.claude.com/docs/en/sub-agents),
   official). A subagent gets a fresh context window — no conversation history.
   The docs enumerate what it **still inherits**: the project-instructions file
   hierarchy, a git-status snapshot, preloaded skills, and the delegation
   prompt. So a stock subagent is conversation-blind but **not project-blind**;
   full blindness requires actively withholding the project-instructions
   inheritance and writing a context-free delegation prompt. Read-only tool
   allowlists mechanically enforce "no fixes."
   *Fit:* this is the mechanical half of itd-86's blindness contract, and the
   inheritance list is the exact checklist of what must be withheld.
2. **fresheyes — Dan Shapiro**
   ([danshapiro/fresheyes](https://github.com/danshapiro/fresheyes), MIT).
   Enforcement by separate process and different model vendor; reviews
   committed changes only; explicitly motivated by the self-preference-bias
   literature it cites. *Fit:* closest published enforced fresh-eyes tool, but
   verdict-driven (PASSED/FAILED, fix-resubmit loop).
3. **LLM-as-judge bias controls** — position-swapping with inconsistency-as-tie,
   verbosity/length controls, rationale-before-verdict
   ([survey, arXiv 2412.05579](https://arxiv.org/pdf/2412.05579)). Evidence
   tier. *Fit:* protocol-level enforcement built for producing verdicts.
4. **CriticGPT** ([OpenAI](https://openai.com/index/finding-gpt4s-mistakes-with-gpt-4/)).
   Trained critic; critiques preferred partly because it produces **fewer
   nitpicks**. *Fit:* not deployable here, but direct empirical support for a
   restrained, findings-only output contract.
5. **Adversarial/red-team review skills and judge panels** — instructed-only
   blindness, verdict-maximising; circulating quality multipliers are
   untraceable. Anecdote/marketing tier; not worth adopting.

## 3. Divergence between shipped state and stated intent

Nothing found audits prose intent against a shipped repository as a first-class
product; the field splits into artefact-vs-artefact checkers and code-anchor
drift detectors.

1. **Spec Kit `/speckit.analyze`** — "read-only cross-artifact consistency and
   quality analysis … a non-destructive report that surfaces discrepancies
   before implementation." Official org. *Fit:* nearest neighbour to a
   divergence audit — read-only, no fixes — but docs-vs-docs, pre-build, full
   context.
2. **Swimm** ([CI docs](https://docs.swimm.io/continuous-integration/)) —
   code-coupled docs verified in CI; anchor-based (symbols/paths), so blind to
   semantic drift in prose; team-SaaS economics. Anti-recommendation at this
   repo's scale.
3. **AST-anchored doc-drift tools** (e.g.
   [drift-vscode](https://github.com/pallaprolus/drift-vscode)) — deterministic
   staleness flags, same anchor limitation.
4. **The contrarian position** ([HackerNoon on drift detection](https://hackernoon.com/when-documentation-lies-detecting-drift-between-code-and-reality)):
   drift checking should be algorithmic precisely because language models
   cannot reliably judge importance; deterministic checks give traceability.
   Contested, and worth holding in the record: it argues for keeping the
   deterministic lint floor (docs-lint, record-lint) load-bearing underneath
   any semantic reading — the layered shape itd-60 already proposes.

## 4. Detection without resolution

No published review tooling makes "no fixes, no ranking, no verdict" its stated
contract — the combination appears to be unclaimed territory.

1. **Design-critique practice** — "problems, not solutions": prescribing a fix
   is cognitively cheaper than articulating the problem, yields shallow
   solutions, and moves the decision away from the author
   ([Nielsen](https://jakobnielsenphd.substack.com/p/design-crit),
   [NN/g](https://www.nngroup.com/articles/design-critiques/), the Pixar
   Braintrust model — diagnose, never prescribe). Established practitioner
   literature; the strongest articulated rationale for itd-86's prohibitions,
   from a human-process tradition nobody has transplanted verbatim into an
   agent instrument.
2. **Read-only review skills** honour no-fix but keep filtering and
   prioritising — ranking survives everywhere.
3. **CriticGPT's fewer-nitpicks result** (above) is the best published evidence
   that withholding judgment density improves critique usefulness.

## Positioning summary

The blindness contract itd-86 specifies — no project context, no cross-run
memory, no fixes, no ranking, no verdict — is not implemented by any single
found instrument. Its parts exist separately: enforced context isolation (host
subagents, with the project-instructions inheritance actively withheld),
enforced cross-vendor independence (fresheyes), read-only no-fix reporting
(`/speckit.analyze`), and the no-verdict rationale (design-critique literature
plus the CriticGPT nitpick finding). The combination is the differentiator, and
the shipped-state-vs-stated-intent reading is the least-served capability of
the four surveyed.

**Three most load-bearing sources:** the
[grilling SKILL.md](https://raw.githubusercontent.com/mattpocock/skills/main/skills/productivity/grilling/SKILL.md)
(interrogation mechanics), the
[host subagent docs](https://code.claude.com/docs/en/sub-agents) (what
isolation enforces and what it leaks), and
[github/spec-kit](https://github.com/github/spec-kit) (clarification and
read-only divergence reporting from an official org).
