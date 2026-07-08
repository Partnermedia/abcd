# References Registry

Canonical bibliography for `.abcd/development/` documents. When citing prior art or external resources, copy the relevant entry from here into the citing document's `## References` block. This registry is the source of truth for *what the canonical entry looks like*; markdown can't transclude, so the copy is accepted by convention.

## Conventions

- **ID format**: lowercase-kebab slug of the project / source name (e.g. `paul`, `carl`, `claude-skills-docs`).
- **Citation form** in body text: markdown reference-style links — `[PAUL][paul]`, `[CARL][carl]`. Reused IDs collapse to one bibliography entry per document.
- **Bibliography block** at the bottom of each citing doc:

  ```markdown
  ## References

  [paul]: https://github.com/ChristopherKahler/paul "PAUL — Plan-Apply-Unify Loop, Kahler"
  [carl]: https://github.com/ChristopherKahler/carl "CARL — Context Augmentation & Reinforcement Layer, Kahler"
  ```

- **Link title** (the quoted string) carries the one-line description; surfaces on hover, removes the need for a numbered footnote list.
- **No IEEE-style `[1][2]` numbering** — renumbering on edit is hostile to a living doc and markdown doesn't auto-link the digits.

## Canonical entries

### Prior-art frameworks

```
[paul]: https://github.com/ChristopherKahler/paul "PAUL — Plan-Apply-Unify Loop, project orchestration framework for Claude Code (Kahler)"
[carl]: https://github.com/ChristopherKahler/carl "CARL — Context Augmentation & Reinforcement Layer, just-in-time rule injection for Claude Code (Kahler)"
[claude-skills-rezvani]: https://github.com/alirezarezvani/claude-skills "claude-skills — large skills/agents collection for Claude Code and other harnesses (Rezvani)"
[everything-claude-code]: https://github.com/affaan-m/everything-claude-code "everything-claude-code — agent harness performance optimisation system (Mahmood)"
[wshobson-agents]: https://github.com/wshobson/agents "wshobson/agents — multi-agent orchestration for Claude Code"
[awesome-claude-code]: https://github.com/hesreallyhim/awesome-claude-code "awesome-claude-code — curated list of Claude Code resources"
```

### Anthropic / Claude Code official

```
[claude-skills-docs]: https://code.claude.com/docs/en/skills "Claude Code Skills (Anthropic docs)"
[agent-skills-overview]: https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview "Agent Skills overview (Anthropic platform docs)"
[claude-code-plugins]: https://github.com/anthropics/claude-code "anthropics/claude-code — Claude Code CLI and plugin marketplace"
```

### Methodology / patterns

```
[bdd-given-when-then]: https://martinfowler.com/bliki/GivenWhenThen.html "Given-When-Then (Fowler) — BDD acceptance-criteria pattern"
[amazon-working-backwards]: https://www.allthingsdistributed.com/2006/11/working_backwards.html "Working Backwards (Vogels) — Amazon press-release-first product design"
```

### Citation, provenance, and publishing

```
[csl-spec]: https://docs.citationstyles.org/en/stable/specification.html "Citation Style Language 1.0.2 specification"
[csl-schema]: https://github.com/citation-style-language/schema/blob/master/schemas/input/csl-data.json "CSL-JSON input schema (csl-data.json)"
[pandoc]: https://pandoc.org/MANUAL.html "Pandoc user's guide — citeproc and bibliography formats"
[better-bibtex]: https://retorque.re/zotero-better-bibtex/exporting/extra-fields/ "Better BibTeX for Zotero — extra fields and custom-field round-tripping"
[hayagriva]: https://github.com/typst/hayagriva "Hayagriva — Typst bibliography file format"
[prov-json]: https://www.w3.org/submissions/prov-json/ "PROV-JSON — W3C member submission"
[adr-org]: https://adr.github.io/ "Architectural Decision Records — homepage and conventions"
[agent-evidence]: https://arxiv.org/pdf/2606.04990 "Survey of evidence tracing and provenance in LLM agents (arXiv 2606.04990)"
[finos-mi13]: https://air-governance-framework.finos.org/mitigations/mi-13_providing-citations-and-source-traceability-for-ai-generated-information.html "FINOS AI governance framework — MI-13 citations and source traceability"
[grep-vs-embeddings]: https://jxnl.co/writing/2025/09/11/why-grep-beat-embeddings-in-our-swe-bench-agent-lessons-from-augment/ "Why grep beat embeddings in a SWE-bench agent (Liu / Augment)"
[no-indexing]: https://vadim.blog/claude-code-no-indexing/ "Analysis of a major coding agent's grep-over-index retrieval design"
[grep-nuance]: https://www.nuss-and-bolts.com/p/on-the-lost-nuance-of-grep-vs-semantic "On the lost nuance of grep vs semantic search"
[sqlite-vec]: https://alexgarcia.xyz/blog/2024/sqlite-vec-hybrid-search/index.html "Hybrid full-text + vector search in SQLite (Garcia)"
[gitleaks-precommit]: https://m3ssap0.github.io/2023/09/29/pre-commit-gitleaks.html "Custom gitleaks rules in a pre-commit hook"
[vale]: https://vale.sh/ "Vale — syntax-aware prose linter"
[quarto-profiles]: https://quarto.org/docs/projects/profiles.html "Quarto project profiles"
[quarto-conditional]: https://quarto.org/docs/authoring/conditional.html "Quarto conditional content"
[quarto-typst]: https://quarto.org/docs/output-formats/typst.html "Quarto Typst PDF output"
[manubot]: https://manubot.org/ "Manubot — git-based manuscripts with automated citation resolution"
```

## Adding a new reference

1. Pick a slug — lowercase kebab, ≤ 30 chars, project-name-shaped.
2. Append the `[slug]: URL "one-line description"` line to the appropriate section above.
3. Copy the line into the citing document's `## References` block.
4. Use `[Display name][slug]` (or `[slug]` for the bare slug) in body text.

If a reference appears in three or more documents, that's a signal the registry entry is well-established; no further action required — the duplication is intentional, not a bug.
