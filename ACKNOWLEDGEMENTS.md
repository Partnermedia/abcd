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

- **"20 Must-Know Agentic AI Terms" (Andreas Horn, LinkedIn, 2026)** — the
  practitioner term list whose assessment seeded the terminology crosswalk
  (itd-100, `docs/reference/terminology.md`). Credited as the prompt, not a
  source: every crosswalk citation is to a primary anchor, and the crosswalk's
  admission rule (no single-author coinages, no aggregators) was formulated in
  reaction to it.
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
- **Claude Code's permission and sandboxing model** — the posture adr-42 adopts
  for `abcd guard`: an argument-constraining command pattern is documented as
  fragile agent steering rather than a boundary, an unsound pattern is refused
  outright instead of shipped, and the enforcing control sits at the execution
  layer.
  <https://code.claude.com/docs/en/permissions>
- **Conftest (Open Policy Agent)** — the severity→exit-code convention (`0`
  clean / `1` warnings / `2` any error) the `abcd lint` verb adopts for its
  tri-state exit, taken as vocabulary without adopting the Rego engine (itd-85).
- **CriticGPT (OpenAI)** — the injected-bug construction behind itd-81's
  calibration corpus: natural defects are unlabelled, so ground truth is
  manufactured by reintroducing defects whose class is already known.
- **Cursor's terminal command controls** — the decisive negative precedent
  behind adr-42: a shipped command denylist bypassed via `bash -c`, subshells
  and base64, replaced by an allowlist, which was then bypassed by poisoning an
  allowed command's environment, published as GHSA-82wg-qcm4-fp2w /
  CVE-2026-22708 (identifiers as recorded during the iss-272 investigation; the
  advisory host was unreachable when this entry was written). abcd takes the
  lesson, not the mechanism.
  <https://cursor.com/security>
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
- **Go's embedded build info (`runtime/debug.BuildInfo` VCS stamping)** — the
  source of the running binary's vintage that itd-111's staleness detection
  reads (build revision and the `vcs.modified` dirty flag), with its documented
  stamping holes driving the first-class *unknown* outcome.
- **GTFOBins** — the `shell` / `command` function taxonomy, which names the
  class `abcd guard` is exposed to (a binary that spawns a shell or runs a
  command) and, by its privilege-escalation inclusion criterion, demonstrates
  that no curated list covers abcd's threat: `nice`, `setsid`, `stdbuf` grant no
  privilege and are perfect bypasses (adr-42). <https://gtfobins.github.io>
- **Homebrew's auto-update-on-use and the `update-notifier` pattern (npm)** — the
  UX grammar itd-111 keeps (cached comparison, a gentle nudge, a one-command
  fix) while rejecting their implicit background network check: abcd implements
  the same grammar over disk-only sources, and the network answers only an
  explicit `--check` (adr-38).
- **git's "behind upstream" notice** — the disk-only precedent itd-111 follows:
  a comparison against locally cached refs, refreshed only by an explicit fetch,
  never a background poll.
- **The Linux kernel's coding-assistants policy** — the `Assisted-by:` attribution
  model abcd adopts for AI-assisted commits.
- **mattpocock/skills (Matt Pocock, MIT)** — two adaptations: the
  glossary-file format (each term a frontmattered Markdown file with aliases
  and forbidden synonyms) behind the brief's terminology glossary, and the
  "grill me" frontier-questioning pattern — an interview that advances by
  asking only what the answers so far cannot settle — adapted as the frontier
  rounds of the brief-creation interview (itd-142).
  <https://github.com/mattpocock/skills>
- **OpenAI Codex's sandbox/approval split** — the vocabulary adr-42 borrows for
  naming what a parse layer is: the OS-enforced sandbox is the boundary, the
  approval policy is "a workflow choice layered on top of" it, and the pattern
  engine carries no threat model.
  <https://github.com/openai/codex>
- **Priority inheritance (real-time scheduling)** — the derived-priority rule
  of the intent dependency graph (itd-78): a minor blocker of a major intent
  computes to major.
- **repolinter** — the declarative rule-object schema (`id` / `severity` /
  `where` / `fix` / `policyInfo`) the `abcd lint` rule model adapts as data,
  separate from the evaluator (itd-85). The tool itself is archived and is not a
  dependency.
- **The Rust RFC process** — the required "Prior Art" section on intents.
- **sudo's `NOEXEC` tag and the sudoers(5) shell-escape statement** — thirty
  years of the same job, and the normative form of adr-42's conclusion:
  restricting users to programs that offer no shell escape "is often
  unworkable", so the answer is an execution-layer control that revokes the
  capability, not a list of the programs that hold it.
  <https://manpages.ubuntu.com/manpages/noble/en/man5/sudoers.5.html>

## References & sources

CSL-JSON: [`.abcd/development/research/references.csl.json`](.abcd/development/research/references.csl.json)

1. Lisanne Bainbridge. 1983. Ironies of automation. *Automatica* 19, 6 (1983),
   775–779. [doi:10.1016/0005-1098(83)90046-8](https://doi.org/10.1016/0005-1098%2883%2990046-8)
2. Shraddha Barke, Michael B. James, and Nadia Polikarpova. 2023. Grounded
   Copilot: How programmers interact with code-generating models. *Proceedings
   of the ACM on Programming Languages* 7, OOPSLA1 (2023), 85–111.
   [doi:10.1145/3586030](https://doi.org/10.1145/3586030)
3. Joel Becker, Nate Rush, Elizabeth Barnes, and David Rein. 2025. Measuring
   the impact of early-2025 AI on experienced open-source developer
   productivity. METR. [arXiv:2507.09089](https://arxiv.org/abs/2507.09089)
4. Colin Bryar and Bill Carr. 2021. *Working Backwards: Insights, Stories, and
   Secrets from Inside Amazon*. St. Martin's Press, New York.
   <https://openlibrary.org/works/OL20924654W>
5. Parmit K. Chilana, Rishabh Singh, and Philip J. Guo. 2016. Understanding
   conversational programmers: A perspective from the software industry. In
   *Proceedings of the 2016 CHI Conference on Human Factors in Computing
   Systems (CHI '16)*, 1462–1472. [doi:10.1145/2858036.2858323](https://doi.org/10.1145/2858036.2858323)
6. Eric Evans. 2003. *Domain-Driven Design: Tackling Complexity in the Heart
   of Software*. Addison-Wesley, Boston.
   <https://openlibrary.org/works/OL4464385W>
7. Ahmed Fawzy, Amjed Tahir, and Kelly Blincoe. 2026. Vibe coding in practice:
   Motivations, challenges, and a future outlook — a grey literature review.
   In *Proceedings of the 48th International Conference on Software
   Engineering: Software Engineering in Practice (ICSE-SEIP 2026)*, 212–223.
   [doi:10.1145/3786583.3786866](https://doi.org/10.1145/3786583.3786866)
8. Alan R. Hevner, Salvatore T. March, Jinsoo Park, and Sudha Ram. 2004.
   Design science in information systems research. *MIS Quarterly* 28, 1
   (2004), 75–105. [doi:10.2307/25148625](https://doi.org/10.2307/25148625)
9. Ken Huang. 2025. *Secure Vibe Coding Guide*. Cloud Security Alliance.
   <https://cloudsecurityalliance.org/blog/2025/04/09/secure-vibe-coding-guide>
10. Andrej Karpathy. 2023. The hottest new programming language is English.
    Post on X (24 January 2023).
    <https://x.com/karpathy/status/1617979122625712128>
11. Andrej Karpathy. 2025. Post coining the term "vibe coding". X (February
    2025). <https://x.com/karpathy/status/1886192184808149383>
12. Andrej Karpathy. 2026. Post proposing the term "agentic engineering". X
    (February 2026). <https://x.com/karpathy/status/2019137879310836075>
13. Amy J. Ko et al. 2011. The state of the art in end-user software
    engineering. *ACM Computing Surveys* 43, 3 (2011), 21:1–21:44.
    [doi:10.1145/1922649.1922658](https://doi.org/10.1145/1922649.1922658)
14. Oliver Kopp, Anita Armbruster, and Olaf Zimmermann. 2018. Markdown
    architectural decision records: Format and tool support. In *Proceedings
    of the 10th ZEUS Workshop* (CEUR-WS Vol. 2072).
    <https://ceur-ws.org/Vol-2072/paper9.pdf>
15. Alistair Mavin, Philip Wilkinson, Adrian Harwood, and Mark Novak. 2009.
    Easy approach to requirements syntax (EARS). In *Proceedings of the 17th
    IEEE International Requirements Engineering Conference (RE '09)*, 317–322.
    [doi:10.1109/RE.2009.9](https://doi.org/10.1109/RE.2009.9)
16. Bonnie A. Nardi. 1993. *A Small Matter of Programming: Perspectives on End
    User Computing*. MIT Press, Cambridge, MA.
    <https://openlibrary.org/works/OL1923390W>
17. Peter Naur. 1985. Programming as theory building. *Microprocessing and
    Microprogramming* 15, 5 (1985), 253–261.
    [doi:10.1016/0165-6074(85)90032-8](https://doi.org/10.1016/0165-6074%2885%2990032-8)
18. Michael Nygard. 2011. Documenting architecture decisions. Cognitect blog
    (15 November 2011).
    <https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions>
19. Hammond Pearce, Baleegh Ahmad, Benjamin Tan, Brendan Dolan-Gavitt, and
    Ramesh Karri. 2022. Asleep at the keyboard? Assessing the security of
    GitHub Copilot's code contributions. In *Proceedings of the 43rd IEEE
    Symposium on Security and Privacy (S&P 2022)*, 754–768.
    [arXiv:2108.09293](https://arxiv.org/abs/2108.09293)
20. Neil Perry, Megha Srivastava, Deepak Kumar, and Dan Boneh. 2023. Do users
    write more insecure code with AI assistants? In *Proceedings of the 2023
    ACM SIGSAC Conference on Computer and Communications Security (CCS '23)*,
    2785–2799. [doi:10.1145/3576915.3623157](https://doi.org/10.1145/3576915.3623157)
21. Ranjan Sapkota, Konstantinos I. Roumeliotis, and Manoj Karkee. 2025. Vibe
    coding vs. agentic coding: Fundamentals and practical implications of
    agentic AI. [arXiv:2505.19443](https://arxiv.org/abs/2505.19443)
22. Advait Sarkar, Andrew D. Gordon, Carina Negreanu, Christian Poelitz, Sruti
    Srinivasa Ragavan, and Ben Zorn. 2022. What is it like to program with
    artificial intelligence? In *Proceedings of the 33rd Annual Conference of
    the Psychology of Programming Interest Group (PPIG 2022)*. [arXiv:2208.06213](https://arxiv.org/abs/2208.06213)
23. Christopher Scaffidi, Mary Shaw, and Brad A. Myers. 2005. Estimating the
    numbers of end users and end user programmers. In *Proceedings of the 2005
    IEEE Symposium on Visual Languages and Human-Centric Computing (VL/HCC
    '05)*, 207–214. [doi:10.1109/VLHCC.2005.34](https://doi.org/10.1109/VLHCC.2005.34)
24. Donald A. Schön. 1983. *The Reflective Practitioner: How Professionals
    Think in Action*. Basic Books, New York.
    <https://openlibrary.org/works/OL3466056W>
25. Shivani Shukla, Himanshu Joshi, and Romilla Syed. 2025. Security
    degradation in iterative
    AI code generation — a systematic analysis of the paradox. In *Proceedings
    of the 2025 IEEE International Symposium on Technology and Society (ISTAS
    2025)*. [arXiv:2506.11022](https://arxiv.org/abs/2506.11022)
26. U.S. Copyright Office. 2025. *Copyright and Artificial Intelligence,
    Part 2: Copyrightability*. <https://www.copyright.gov/ai/>
27. Gerald M. Weinberg. 1971. *The Psychology of Computer Programming*. Van
    Nostrand Reinhold, New York.
    <https://openlibrary.org/works/OL1958820W>
28. Songwen Zhao, Danqing Wang, Kexun Zhang, Jiaxuan Luo, Zhuo Li, and Lei Li.
    2025. Is vibe coding safe? Benchmarking vulnerability of agent-generated
    code in real-world tasks. [arXiv:2512.03262](https://arxiv.org/abs/2512.03262)
