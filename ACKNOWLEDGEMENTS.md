# Acknowledgements

abcd stands on ideas, tools, and writing from many sources. This file records them
in three parts: the **development** that built abcd, the **inspirations** that
shaped its design, and the **references** it draws on. Each entry is added in the
same change that lands what it records — the pull request that adopts a pattern,
cites a source in an ADR, or integrates a tool — so the list grows with the work
rather than being reconstructed later. Runtime dependencies are not listed here;
they live in `go.mod` and the licence notices they carry.

## Development

Development of abcd has been assisted by Claude Code (Anthropic). Per-commit
disclosure uses an `Assisted-by:` trailer; the human contributor is the author of
record and is responsible for all AI-assisted output — its correctness, licensing,
and fit for the project. See [`CONTRIBUTING.md`](CONTRIBUTING.md).

## Inspirations

Ideas and methodologies that shaped the design — not code abcd depends on.

- **Agentic Context Engineering (ACE)** — the append-only-delta model of a
  self-improving instruction record, and the two failure modes it names —
  *brevity bias* and *context collapse* — which itd-81 cites to strike itd-5's
  "shorter by >10%" prompt tiebreak.
- **Amazon "Working Backwards"** — the press-release format of abcd's intents.
- **Architecture Decision Records (MADR)** — the shape of the decision record.
- **ccpm (Claude Code PM, Automaze)** — the markdown spec/task conventions
  (PRD → epic → issue, directory-as-store) that abcd's native spec layer is
  convention-compatible with, and the designated deeper backend of the spec
  seam (ADR-24, ADR-26). <https://github.com/automazeio/ccpm>
- **Citation Style Language (CSL-JSON)** — the bibliography format of the
  confidential-sources design (itd-76), whose reserved `custom` field carries
  the confidentiality metadata.
- **Conftest (Open Policy Agent)** — the severity→exit-code convention (`0`
  clean / `1` warnings / `2` any error) the `abcd audit` verb adopts for its
  tri-state exit, taken as vocabulary without adopting the Rego engine (itd-85).
- **CriticGPT (OpenAI)** — the injected-bug construction behind itd-81's
  calibration corpus: natural defects are unlabelled, so ground truth is
  manufactured by reintroducing defects whose class is already known.
- **DITA subject scheme maps** — the controlled-vocabulary pattern behind the
  persona registry: a field's legal values live in a dedicated registry file
  and a processor flags unbound values (the `persona_registry` lint rule).
- **Diátaxis** — the four-type model behind the user documentation.
- **Domain-Driven Design (bounded contexts)** — the surface boundaries.
- **Doorstop** — the suspect-link fingerprint mechanism adopted for intent
  dependency edges (itd-78), and the store-one-direction/derive-the-reverse
  link model the edge schema follows (shared with OpenFastTrace and
  Sphinx-Needs).
- **GEPA (reflective prompt evolution)** — the score → reflect-on-failing-traces
  → minimal-delta → re-score loop that itd-81 adopts as a human-approved manual
  procedure rather than as a library dependency.
- **The Linux kernel's coding-assistants policy** — the `Assisted-by:` attribution
  model abcd adopts for AI-assisted commits.
- **Priority inheritance (real-time scheduling)** — the derived-priority rule
  of the intent dependency graph (itd-78): a minor blocker of a major intent
  computes to major.
- **repolinter** — the declarative rule-object schema (`id` / `severity` /
  `where` / `fix` / `policyInfo`) the `abcd audit` rule model adapts as data,
  separate from the evaluator (itd-85). The tool itself is archived and is not a
  dependency.
- **The Rust RFC process** — the required "Prior Art" section on intents.

## References & sources

Books, articles, and papers cited in the design record.

- Eric Evans, *Domain-Driven Design* (Addison-Wesley, 2003).
- A. Mavin, P. Wilkinson, A. Harwood, M. Novak, "Easy Approach to Requirements
  Syntax (EARS)" (RE, 2009).
- U.S. Copyright Office, *Copyright and Artificial Intelligence, Part 2:
  Copyrightability* (2025).
