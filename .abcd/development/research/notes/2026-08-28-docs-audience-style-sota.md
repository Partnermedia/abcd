# Docs audience and accessible-language rules: fit-challenge and SOTA verdict

**Date:** 2026-08-28. **Question:** should `docs/reference/writing-style.md`
gain (a) an "always write in accessible language" rule and (b) a two-register
audience model — product prose written for humans, technical prose assumed to
be consumed mostly by AI agents? **Method:** the prefer-sota sequence — an
independent adversarial fit-challenge over the repo's own record, and an
independent SOTA research pass, synthesised by a third party. Informs
[adr-53](../../decisions/adrs/0053-audience-by-placement.md).

## Fit-challenge (independent pass over the repo record)

Enumerated collisions, abridged to the load-bearing ones:

- A blanket "accessible language" row cannot carry an honest enforcement
  label (`enforcement-claims-are-facts`): there is no criterion even a
  reviewer can check. Generic plain-language doctrine (short sentences, avoid
  semicolons and dashes) contradicts the guide's ratified punctuation rules.
- Accessibility is already assigned to structure: Diátaxis one-type-per-page
  is the register gradient; a page too dense for its reader is mis-typed or
  mis-placed, not in need of re-registering.
- The audience boundary already has a canonical primitive: placement. The
  three-tier layout separates the agent-heavy development record from
  user-facing `docs/`, and the "JSON internal, MD render" invariant assigns
  the machine audience to machine surfaces (`--json`, the generated CLI
  reference, the rules loader). A register axis inside `docs/` would be a
  second copy of that boundary (`one-canonical-primitive`) and would open a
  "this page is agent-register" escape hatch in every path-scoped lint rule.
- One genuine inertia finding: audience-by-placement was enacted everywhere
  but ratified nowhere. Per prefer-sota's bounds, that calls for a deliberate
  ADR, not a silent keep. That ADR is adr-53.

## SOTA pass (independent research, sources fetched 2026-08-28)

- **No stylistic human/agent split.** The evidence converges on one prose
  register: documentation that works for AI is well-structured human
  documentation (kapa.ai's writing-for-AI guidance, which analogises
  AI-optimisation to screen-reader accessibility). Vendor guidance for agent
  instruction files keeps even those "short and human-readable".
- **The split that won adoption is by artifact class, not register:** plain
  Diátaxis-organised docs for everyone, plus dedicated agent instruction
  files (the AGENTS.md convention, 60k+ public repositories, Linux Foundation
  stewardship). The strongest counter-experiment, llms.txt (a dedicated
  machine-readable docs surface), failed empirically: roughly 10% adoption,
  ~97% of files receiving zero traffic, no citation correlation, no major
  provider committed (Ahrefs and SE Ranking data, read via aggregators).
- **"Accessible language" is adoptable only decomposed** into checkable rules
  (second person, active voice, define-on-first-use, self-contained sections,
  a ~25-word sentence checkpoint judged by a human). Readability-score gates
  (Flesch-Kincaid class) are rejected: gamed by sentence-splitting and
  penalise necessary domain terms.
- **Gate discipline (GitLab documentation-testing policy):** deterministic
  checks may block; heuristics stay advisory; and a rule is promoted to
  blocking only after every existing occurrence in the corpus is fixed. This
  matches the blocker/warn + allow-escape architecture the repo already
  ships, and is written into the guide's preamble as the promotion rule.
- **Attribution caution:** the circulating quote "good for humans is not good
  for agents; tokens are expensive" attributed to a vendor announcement does
  not appear in the announcement it is attributed to. Not citable.

Key sources: kapa.ai "Writing documentation for AI: best practices"
(<https://docs.kapa.ai/improving/writing-best-practices>); AGENTS.md
(<https://agents.md/>); llms.txt adoption evidence
(<https://www.digitalapplied.com/blog/llms-txt-in-practice-adoption-evidence-2026>);
GOV.UK clear-language guidance
(<https://guidance.publishing.service.gov.uk/writing-to-gov-uk-standards/writing-guidelines/clear-language/>);
GitLab documentation testing (Vale severity policy); ISO 24495-1:2023 as the
plain-language umbrella (principles, not checkable rules).

## Verdict (adopted by the maintainer, 2026-08-28)

The two-register proposal is declined; audience-by-placement is ratified as
adr-53. The accessible-language value enters the guide only as decomposed,
honestly-labelled rules: the Audience section (one register, density via
Diátaxis, self-contained sections, terms linked to the crosswalk) and the
warn-then-promote discipline in the preamble.
