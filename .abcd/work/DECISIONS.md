# DECISIONS

Append-only, one line per decision, newest last. Date-prefixed. Architecture-shaping
decisions graduate to an ADR under [`../development/decisions/adrs/`](../development/decisions/adrs/).
Graduate this file to per-file `decisions/<date>--<slug>.md` if size or
parallel-agent merge contention bites.

- 2026-07-06 — Rebuild abcd from scratch in Go, no external tools (specstory,
  RepoPrompt, flow-next, Ralph, codex); ship an MVP, extend via the companion harness then
  Claude Code.
- 2026-07-06 — Transport-agnostic Go core; CLI is the reliable default front door;
  MCP is an additive front door on the same core, added later.
- 2026-07-06 — Peer with the companion harness via conventions + MCP; no Go dependency either way.
- 2026-07-06 — LLM work host-delegated by default; native/CLI/API/MCP oracles are
  opt-in adapters.
- 2026-07-06 — Spec/task layer native-minimal; the companion harness `ccpm` the primary deeper
  backend; flow-next dropped. Autonomous run not a Ralph port (host orchestrators).
- 2026-07-06 — Single repo, curated release (no dev→public mirror); the repo is the
  marketplace. Private companion repo deferred (trigger: shared transcripts).
- 2026-07-06 — Three-tier `.abcd/` layout: development (durable) / work (shared) /
  .work.local (local). `docs/` user-facing only.
- 2026-07-06 — Module path `github.com/REPPL/abcd-cli`; Cobra approved as the CLI
  framework (matches ferry and the companion harness).
- 2026-07-08 — Confidential sources: global user-level corpus (CSL-JSON + grep
  corpus, local no-remote git), append-only JSONL influence ledger per repo,
  banlist patterns generated from confidential entries into the itd-74 private
  guard; convention + skill first, `abcd source` verbs deferred (itd-76). Quarto
  chosen for eventual paper reconstruction; RAG rejected at this scale.
- 2026-07-08 — Personas in any scenario are always Alice, Bob, Carol (in that
  order); the user is they/them. Recorded as a principle.
- 2026-07-08 — Consume-model interview: `spec_id` is SCALAR (never a list) —
  split-the-intent is doctrine (itd-67/72 precedent); task decomposition lives
  inside the spec. Principles/disciplines get a promotion path: enforced
  principle ⇒ discipline-kind intent (personas principle promotes when its
  registry lint ships). Coverage vocabulary (uncovered / covered-shallow /
  covered-deep / orphaned / unwanted) lands as itd-53's gate reporting
  language; "done" = covered-deep AND the intent's own criteria MET.
- 2026-07-08 — persona_registry lint shipped (record-lint blocker: quote
  attributions must name registry personas); the personas principle promoted
  to discipline itd-79 the same change, per the promotion path — first test
  case of enforced-principle ⇒ discipline. principles/ file retired.
- 2026-07-08 — Persona SSOTs reconciled: `personas.json` is the single registry
  (13, expandable, alphabetical sequence); selection is BY ROLE, the role's
  registered name is used, never a name picked directly; all personas and the
  real user are they/them. Principle file updated to point at the registry;
  registry-membership lint is the intended gate.
- 2026-07-08 — Intents gain a required `## Prior Art` section (positions the
  intent against corpus + outside work; ≥1 resolvable reference or an explicit
  "none found — searched X"). Coherence stays at promotion (itd-42), whose
  Tier 2 now also loads `principles/`; capture stays severity + edges.
- 2026-07-08 — Edges stay one-way (dependent-authored), reverse views derived
  only; itd-78 lint rejects hand-authored reverse fields; edges gain optional
  content fingerprints (`itd-N@hash`) so a target's change marks inbound edges
  suspect. Intent doneness = spec closed AND the intent's own itd-1 criteria
  MET (never inferred from spec close) — full consume-vocabulary decision
  deferred to a follow-up interview.
- 2026-07-08 — itd-76 grilled: leak guard promises literal strings only
  (paraphrase risk stated, handled behaviourally + review); citation is a
  two-level AND (source permission_status AND per-line cited_publicly); author
  bans default on with per-source ban_authors opt-out; standalone `source`
  domain (itd-16 a possible backend, not a dependency); pre-commit auto-
  refreshes the generated banlist; public render proven by structural filter
  AND post-render lint; team share of citation data via committed
  `.abcd/work/references.json` (share/ingest); durability = machine backup +
  git bundle, multi-machine deferred.
- 2026-07-08 — `~/.abcd/` blessed as abcd's user-level home (fourth tier,
  additive to repo `.abcd/`), path configurable; relocation wizard recorded as
  itd-77.
- 2026-07-08 — Author bans FLIPPED to opt-in (`ban_authors: true`), superseding
  today's default-on decision: the actual corpus population (own submitted
  work, purchased reports, private repos) makes author bans near-pure false
  positives — they would ban the user's own name — while title/alias patterns
  carry the real protection.
- 2026-07-08 — Corpus restructured to class-segregated per-source folders
  (confidential/<key>/, public/<key>/): confidentiality is declared at
  ingestion and LOCATION is its single source of truth (flag mirrors, tooling
  refuses on mismatch); derived artifacts inherit by location; declassification
  is a visible git mv.
- 2026-07-08 — Severity ≠ priority (records an earlier-session decision):
  intents declare `severity` (capture-ledger enum) and edges (`blocked_by`,
  `builds_on`); effective priority is DERIVED via priority inheritance (max of
  own severity and severity of everything transitively blocked) and never
  stored — a minor blocker of a major intent jumps the queue while staying
  minor. Phases keep sequencing authority (adr-9); lint makes contradictory
  schedules fail. Recorded as itd-78; piloted on itd-76/77.
- 2026-07-08 — Predecessor spc-N artefacts inside intents (do-not-implement
  banners, implementation-complete AC tables) are demoted to Prior Art design
  input per the delivery-state provenance doctrine — never implementation
  authority, never a delivery claim (iss-16 itd-66, iss-17 itd-50); their
  deltas become spec-time Open Questions.
- 2026-07-08 — itd-37's itd-36 edge downgraded blocked_by → builds_on: the
  capture + enforcement half ships independently (Phase 0 registration) and
  only extraction-to-memory waits on itd-36 (iss-18); the launch deepenings'
  unscheduled state is recorded in the phase index pointing at adr-33
  (iss-20); itd-6 stays planned/ — ADR-25 superseded its framing only, and
  scheduled implies planned per adr-34 (iss-22).
- 2026-07-08 — Post-review recording follows fix-the-detector: findings are
  captured as clustered issues (iss-29..49), each naming the detector (gate,
  lint rule, or test convention) that catches its class and carrying its
  instances as the detector's acceptance corpus; instances drain behind the
  armed detector, never hand-fixed ahead of it. Ten principles recorded from
  the 2026-07-08 multi-agent review; distillation in research/notes.
- 2026-07-10 — The practice/MVP/tool trichotomy lands as an amendment to the
  principles README promotion path (one canonical three-rung ladder:
  principle -> enabling convention/script/format -> discipline-kind intent or
  core absorption), never as a third doctrine file — the adversarial review
  found standalone adoption would duplicate and contradict existing doctrine.
  Intake rules kept verbatim: articulate the full ladder for every candidate;
  never fabricate an absent rung (research/notes/2026-07-09).
- 2026-07-10 — Doctrine grows on observed need: the 31 deferred medium
  proposals from the extraction stay parked in the 2026-07-09 research note
  until a live instance arises; calibrate-the-judge deliberately waits for
  the first live LLM gate (its measured-agreement requirement is already
  recorded in verifier-selects-gates-decide's promotion path).
- 2026-07-10 — Public sources whose titles collide with locally-banned
  private names are cited by author + arXiv/DOI identifier, never by title
  or corpus key, in committed artifacts; the corpus ledger carries the real
  key. First instance: Tan et al. (UCL, 2026; arXiv 2604.09581).
- 2026-07-10 — AI-generated-only ("tainted") proposals are recorded as
  hypotheses and never adopted until independently verified against a
  citable source — the manual form of tier-travels-with-the-source (iss-52).
- 2026-07-10 — CONTEXT.md goes status-free: it keeps orientation and the
  live sharp-edges list only; hand-written phase/status claims are banned
  (extending adr-5's no-status-in-design-docs rule to the work tier) and a
  record-lint rule on .abcd/work/CONTEXT.md is the detector, armed before
  the rewrite per fix-the-detector. The content rewrite rides with iss-35's
  brief-vs-surface reconciliation. Rejected: deleting the file (loses the
  only committed shared home for sharp edges); generating it (a committed
  generated file is its own drift problem).

- 2026-07-10: Repo preparation is a plugin skill (`/abcd:prepare-this-repo`),
  superseding the external scaffold-repo script's entry point. Grilled rulings:
  the committed AGENTS.md working-conventions section is full-inline and
  NAMELESS (a pre-public repo name never lands in target repos) between dated
  markers for later tooling; the skill hard-refuses not-owned repos (no audit,
  no local layer — we don't impose our principles on others' repos); legacy
  root `.work/` layouts migrate propose-then-sign-off, never leaving two
  working-state homes; no re-run/update machinery now — the CLI will own
  managed-repo migration (gaps seeded as iss-56/iss-57). Rejected: a standalone
  handover prompt file (drifts, unversioned); naming abcd in private-only
  target repos (two-class rule someone eventually gets wrong).
- 2026-07-11 — iss-35's brief↔surface cross-check is **bidirectional, but only
  the structural half is deterministically lintable**: Direction B (every
  `commands/`+`skills/` entry has a brief home) is a coverage lint like
  `directory_coverage`; Direction A (brief claims match *binary behaviour* —
  flags, exit codes, schema fields, counts) is irreducibly semantic and stays an
  LLM/agent job (encoding binary facts into the linter just moves the drift).
  So "graduate the detector to a record-lint rule" is a *reshaping* (extract the
  deterministic half; keep the semantic half as a periodic/agent check), not a
  port. The graduation is a design gate held for maintainer sign-off — options in
  `.abcd/development/plans/2026-07-11-iss35-record-lint-graduation.md` (recommend
  Option A, structural `surface_coverage` rule) — and it is **blocked** until the
  docs/history surface-taxonomy adjudication is decided (a coverage rule fires on
  the three chapterless shipped verbs the moment it is armed).
