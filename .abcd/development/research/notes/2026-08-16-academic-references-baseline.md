# Academic references baseline — the literature the design record draws on

The design record cites its academic grounding piecemeal — an entry lands with
the change that uses it — but the References & sources section of
`ACKNOWLEDGEMENTS.md` had accumulated only three entries while the ideas the
record leans on (end-user programming, judgement-as-constraint, the security
evidence on AI-generated code) went unrecorded. This note establishes the
baseline: a curated set of primary sources, each verified against its publisher
or preprint page, admitted in one change together with the machine-readable
store that carries their canonical metadata.

## Where the entries live

- **Canonical metadata**: `../references.csl.json` — CSL-JSON, one entry per
  source, keyed `<authorfamily><year><distinctiveword>` (e.g. `naur1985theory`).
  CSL-JSON rather than BibTeX per the bibliography-format ruling in
  `2026-07-08-confidential-sources-provenance-sota.md` §1: one canonical
  format, `.bib` derivable on demand.
- **Human-readable entries**: the References & sources section of
  `ACKNOWLEDGEMENTS.md`, a hand-written numbered list in ACM reference style,
  ordered alphabetically by first author. The two are kept
  in sync by protocol (see `_references.md` § Adding a new reference); a lint
  cross-check is a possible later rung, captured as iss-248.

## Admission criteria

- **Primary sources only** — the paper, book, report, or original post; never
  an aggregator, explainer, or secondary write-up. (The same rule the
  terminology crosswalk applies.)
- **Verified citations only** — author list, venue, year, and DOI/arXiv id
  checked against the publisher or preprint page. Author names are taken from
  the publisher's *current* author record, never from an older citation
  string: publishers retroactively update names, and a person's current name
  is the correct one regardless of what an early printing carried.
  Unverifiable details are omitted rather than guessed. Every entry carries a
  resolver link: a DOI
  (verified registered via the doi.org handle API) or, failing that, an
  arXiv/repository URL; books link their Open Library work page.
- **Curated, not exhaustive** — a source earns its entry by anchoring something
  the record actually says. Tool repositories the design drew on are
  Inspirations entries, not references.

## What each group anchors

- **End-user programming** (`nardi1993smallmatter`, `ko2011euse`,
  `chilana2016conversational`, `scaffidi2005estimating`,
  `weinberg1971psychology`) — the population abcd serves: people who program
  as a means to a domain end, produce most of the world's software, and lack
  the professional's verification repertoire. Nardi's
  intention-versus-implementation distinction is the oldest statement of the
  gap intent-driven development closes.
- **AI-assisted development in practice** (`karpathy2023english`,
  `karpathy2025vibecoding`, `karpathy2026agentic`, `sarkar2022whatisitlike`,
  `barke2023grounded`, `sapkota2025vibeagentic`, `fawzy2025vibecoding`,
  `becker2025metr`) — the vocabulary (vibe coding, agentic engineering, the
  vibe-versus-agentic continuum) and the evidence that oversight, not
  production, is where practice strains: the METR trial's
  slower-while-feeling-faster result is the sharpest single datum for why the
  record insists on gates rather than trust.
- **Security of AI-generated code** (`pearce2022asleep`, `perry2023insecure`,
  `shukla2025degradation`, `zhao2025vibesafe`, `huang2025securevibe`) — the
  case for automated validation gates: generated code ships vulnerabilities at
  scale, assistance inflates confidence while degrading security, and
  iteration compounds rather than repairs the problem. This is the evidence
  base behind routing what a non-expert cannot see to machine-checked gates.
- **Human judgement and method** (`naur1985theory`, `bainbridge1983ironies`,
  `schon1983reflective`, `hevner2004dsr`) — why the human stays the limiting
  constraint by design: Naur's program-as-theory (the artefact is not the
  knowledge), Bainbridge's ironies (automation leaves humans the hardest,
  least-practised part), Schön's reflection-in-action, and design science as
  the method frame for building-as-enquiry.
- **Records and specification** (`evans2003ddd`, `mavin2009ears`,
  `nygard2011adr`, `kopp2018madr`, `bryar2021workingbackwards`,
  `uscopyright2025part2`) — the record formats abcd adopts: bounded contexts
  behind the surface boundaries, EARS behind constrained requirement syntax,
  ADR/MADR behind the decision record, working-backwards press releases behind
  intents, and the copyrightability report behind the attribution stance.

## Known caveats

- `weinberg1971psychology` — admitted for its distinction between programming
  as a profession and programming as a means past one's own problem; the exact
  passage wants verifying against a physical copy before any record document
  quotes it.
- `fawzy2025vibecoding` — a grey-literature review (101 practitioner sources),
  not a controlled study; cite it as practitioner-evidence synthesis only.
- `karpathy2025vibecoding` / `karpathy2026agentic` — primary posts with no
  methodology; citable as terminology origins, nothing more. Store titles are
  descriptive placeholders in brackets, not the posts' wording.
