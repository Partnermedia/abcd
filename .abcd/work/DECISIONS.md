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
  managed-repo migration (gaps seeded as iss-84/iss-85, originally minted as
  duplicate iss-56/iss-57). Rejected: a standalone
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
- 2026-07-11 — iss-35 graduation SIGNED OFF (maintainer, 4 decisions):
  (1) **Graduation = Option C (hybrid)** — build the deterministic
  `surface_coverage` record-lint rule AND wire the LLM cross-check as a standing
  release gate for semantic (Direction-A) drift.
  (2) **docs/history/version = user-facing surfaces** — each gets a
  `04-surfaces/` chapter + README row (resolves adjudication item 5).
  (3) **consult/ingest/prepare-this-repo reclassify skills → commands via
  relabel** — they stay host-delegated markdown workflows (no Go verbs; the
  "host-delegated by default" boundary holds), but the brief calls them commands
  with command-shaped homes; the read-only skill boundary rule is kept as-is — the
  skill *classification* was what gave (resolves adjudication item 6). abcd ships
  zero skills again.
  (4) **Push/merge policy** — the run's blanket "never push" was an
  unattended-safety override, not the standing rule; normal repo policy resumes
  when the maintainer is driving (docs/chore direct-to-main OK; feat/fix via PR
  awaiting their merge). Main pushed to origin; `auto/context-status-lint` opened
  as PR #12 (awaits maintainer merge; no auto-merge on a feat).
- 2026-07-11 — itd-3 rules-loader hook is **Go**, not Python. abcd is Go-only, so
  the `UserPromptSubmit` router is a Go subcommand invoked by `hooks/hooks.json` —
  the intent's `hooks/prompt_router_hook.py` is a stale pre-Go-rebuild detail and
  is superseded. No Python is added for the loader.
- 2026-07-11 — itd-3 rules-loader **design signed off** (plan
  `2026-07-11-itd-3-rules-loader.md`, prefer-sota verdict). Surviving shape: a
  transport-agnostic `internal/core/rules` capability with two front doors
  (`abcd rules [domain]` verb + `abcd hook prompt-router`), **not** an adapter
  seam. Four intent deltas approved: **D1** event-driven refresh on
  `SessionStart(compact)` (fixed-N demoted to a ~15–20 backstop, not primary);
  **D2** keep the shipped `{schema_version,disabled,domains{}}` shape (legacy
  `extends`/`overrides` sketch superseded); **D3** zero model-facing tokens on
  no-match + out-of-band diagnostic log (supersedes the "<200-token header"
  acceptance criterion); **D4** `.abcdignore` rejected for v1. Build proceeds
  phased/TDD from Phase 1 (`internal/core/rules`).
- 2026-07-11 — itd-3 **shipped manually** ahead of the intent-lifecycle pipeline.
  Moved `planned/ → shipped/` by hand with `spec_id: spc-1` (reserved — the future
  native spec store adopts spc-1 for itd-3, never re-mints it) and a hand-authored
  `## Audit Notes` (the `intent-fidelity-reviewer` agent does not exist yet; judge
  = Claude Opus 4.8). Rollup 3 MET / 1 MET_WITH_CONCERNS / 1 INCONCLUSIVE / 1
  NOT_MET; every divergence is a signed-off D1–D4 delta, the one gap is the AC6
  legacy-harvest completeness. Inbound links repointed planned→shipped by hand —
  the link-drift-on-move the future reconcile pass automates.
- 2026-07-11 — Intent-lifecycle slice 1 (build sign-off given): the pipeline is
  **dogfooded** — itd-3 stays shipped as the reference fixture (option b), and a
  new tightly-scoped intent **itd-80-intent-lifecycle-automation** (ACs = the
  steel thread) is the pipeline's first real payload, driven drafts→planned→
  shipped through the machinery it specifies. Slice scope: minimal native spec
  store (`internal/core/spec`, directory-as-truth open/closed, `intent:` link),
  `abcd intent` (plan/link/review-ingest + bare render) and `abcd spec`
  (close + bare render) verbs, deterministic reconcile inside `spec close`
  (no vendor event), host-delegated `intent-fidelity-reviewer` markdown agent
  (Role 1 only) + async outbox/inbox verdict ingest to `## Audit Notes`.
- 2026-07-11 — `spc-N` minting rule for slice 1: `max(N over spec-store files ∪
  N over every intent's spec_id) + 1`, so the first mint is spc-2 (itd-3's
  reserved spc-1 is respected without a backing spec file). Reconciling the
  store's sequential minting with the brief's aspirational spc-numbering is
  deferred to the richer spec-store slice. Reviewer roles 2/3 (itd-48),
  loop-to-acceptance (itd-50), bundle/discipline lifecycles, and the spec
  dependency graph are all explicitly deferred.
- 2026-07-12 — clean-slate hardening run STEP 0 triage. A fresh adversarial
  sweep (15 ruthless + 9 security reviewers over current `main`, every finding
  independently verified) returned 34 real findings (19 CONFIRMED, 15 PLAUSIBLE,
  0 REJECTED; full corpus `.abcd/.work.local/logs/clean-slate-run/sweep-findings.json`).
  Key result: the sweep INDEPENDENTLY RE-CONFIRMED the 2026-07-08 review's
  code-defect backlog (iss-29/30/32/33/34) is real and still unfixed — prior runs
  deferred those code fixes for docs-reconciliation (iss-35/36) and itd-80 feature
  work. Draining them is this run. Two BLOCKs found: the scanner serialises a
  finding's snippet masking only its own token, leaking sibling secrets on the
  same line (iss-65). Triage disposition: newer-package findings (scanner, rules,
  intent, spec, frontmatter, lint receipt-gate, capture concurrency, core) minted
  as iss-65..72; older-package findings map to existing homes — memory ingest
  (C12/C13/P11)→iss-30, atomic-write/fsutil (P1/P6/seed2)→iss-32, ahoy install
  (C2/C3)→iss-33, launch glob panic (C11)→iss-34, identity fail-open (C8/P12)→iss-63,
  history redaction (C6/C7/P2)→iss-29. iss-70's C16 fix adds `policy.detector` to
  the receipt-JSON schema — a record-lint CONTRACT change flagged for maintainer
  sign-off before landing (a STOP-adjacent design surface). Ledger triage committed
  to `main` as a `chore:` record commit (matches prior record-to-main practice;
  keeps the fix branches clean); each code fix lands on its own `auto/*` branch + PR.
- 2026-07-12 — iss-66 rules-loader trust boundary. Fixed the two mechanical items:
  the Load Lstat→ReadFile TOCTOU on `.abcd/rules.json` (now open-once O_NOFOLLOW +
  fstat, C19) and the session-state dir moved off the world-writable shared /tmp to
  the per-user cache dir (P14). **P15 document-accepted, NOT changed:** a per-repo
  `.abcd/rules.json` can set a default domain dormant and flip the global kill
  switch (Merge is intentionally per-field + sticky-kill-switch). Rationale: rules
  are an *opt-in, opinionated-but-overridable* config layer; `.abcd/rules.json` is a
  committed file (editing it needs repo write access, like any committed guardrail),
  and the real enforcement of dangerous actions is harness-level (git-guardrails
  hooks, the iss-62 identity gate, pre-commit), not the injected advisory prose.
  Silencing a domain removes prose, not a hard gate. **Deferred design alternative
  (surfaced, not taken):** introduce a protected "guardrail" domain class that a
  per-repo override cannot set dormant and that the kill switch cannot silence —
  this adds a new protected-domain concept to the rules contract, a maintainer
  decision, not an autonomous change.
- 2026-07-12 — iss-30 (memory ingest boundary) partially resolved: the fetch/read
  subset — C12 (HTTP status), P11 (SSRF NAT64/6to4), C13 (local size cap), the
  ~user tilde mangle — landed in PR #38. iss-30 stays OPEN for its remaining
  instances (the larger "ingest test-suite" effort): the --keep-original
  partial-failure reporting, CRLF parser-parity (parseFrontmatter vs
  splitFileFrontmatter), and broader URL-ingest/content-type/PDF path coverage.
- 2026-07-12 — /abcd:auto-loop design recorded (plans/2026-07-12-abcd-auto-loop-skill.md,
  pending sign-off, not built). SOTA pass (sota-researcher, primary sources) backs the
  design: durable handoff + fresh-context resume over compaction/RAG (Anthropic
  long-running-harnesses — compaction "isn't sufficient"); delegate reads/reviews but
  keep implementation in ONE agent (Cognition "Don't Build Multi-Agents" + Anthropic
  multi-agent — converging read/write boundary); reviewers must be a SEPARATE fresh-
  context lens, not intrinsic self-review (Huang et al. 2310.01798; CriticGPT
  2407.00215); gate irreversible actions on action-class not self-confidence (RLHF
  miscalibration); attempt-journal lineage = Reflexion (NeurIPS 2023) + database WAL.
  Rejected: parallel multi-agent implementation, compaction-as-primary-continuity,
  RAG-over-ledger at single-milestone scale.
- 2026-07-12 — autonomous-run surface named /abcd:run, taking itd-29's reserved
  name as the host-delegated realization of its operator surface (not a parallel
  /abcd:loop). Discovery: itd-29 (autonomous-run-resilience, planned) already owns
  this surface over the ADR-27 run seam, already scopes out-of-band-merge/chain
  reconciliation (host-owns-git MVP → future read-only `abcd run reconcile --json`)
  and 429/quota (spc-35), and is deliberately deferred pending real evidence
  (revisit trigger #5: two end-to-end autonomous runs). Sequence C→A: run the loop
  as a plan+protocol under the harness loop now to dogfood + generate that evidence;
  formalize commands/abcd/run.md + brief row + surface_coverage, reconciled into
  itd-29, after 1-2 successful runs. Binary operator verbs (budget preflight, rewind,
  ship, run reconcile) stay deferred in itd-29.
- 2026-07-12 — Judge calibration captured as a DISCIPLINE (itd-81), not a standalone
  intent: verdict-rendering agents are plumbing (no user moment), and itd-5 is the
  precedent for a cross-cutting rule over agent prompts. Core rule: no judge ships
  unmeasured — a labelled corpus with known-good cases ≥40%, scored on true-negative
  rate as a first-class metric alongside recall, with a declared TNR floor gating the
  prompt lock. Evidence: LLM code judges systematically over-flag and ~1/3 of their
  errors are hallucinated code (2603.00539); judges over-rate LLM-written and
  under-rate human-written code (2507.16587); ground truth is manufactured by
  injecting defects (CriticGPT 2407.00215). CORRECTS itd-5: its pre-flight tiebreak
  ("passes goldens AND >10% shorter") selects for the brevity bias that ACE
  (2510.04618) identifies as destroying instruction quality — struck; the gate is the
  corpus score. CONSTRAINS itd-64: reviewer verdicts are not ground truth (the
  reviewer is the instrument under measurement) and its tuning loop must stay
  human-gated — unattended proxy-optimisation reward-hacks at 73.8% (OpenReview
  ikrQWGgxYg). Rejected: judge panels/juries (nine judges → 2.18 effective votes,
  correlated errors, no better than the single best judge — 2605.29800); 1-5 severity
  scores (middle-drift, position bias); reasoning inside a JSON schema (2408.02442).
- 2026-07-12 — itd-5 AMENDED (not superseded) per itd-81, two rules: (a) the v1.0.0
  pre-flight's "shorter by >10%" tiebreak is STRUCK — length selects for ACE's brevity
  bias, and it selected against goldens that never measured false positives; the gate
  is now the calibration-corpus score, ties to the candidate. (b) `1.0.0` now MEANS
  measured — an agent stays in the `0.x` band until it clears a corpus, because
  stamping 1.0.0 on an unmeasured prompt asserts a lock that never ran. All five
  shipped agents are `0.1.0`.
- 2026-07-12 — The four personal reviewer agents (ruthless, security, docs-currency,
  sota-researcher) MOVED from the machine-global `~/.claude/agents/` into abcd's
  plugin `agents/` and deleted at source; they now resolve as `abcd:<name>` in every
  repo with the plugin enabled, versioned in-repo and reviewable by PR. Frontmatter
  key is `prompt_version` (itd-5's name), not `version` — intent-fidelity-reviewer
  renamed. Colour encodes the DOMAIN EXAMINED, never rank or taste: red=trust
  boundary, orange=code correctness, blue=documentation truth, green=the record,
  purple=external evidence; cyan reserved for artefact-producing (non-verdict) agents.
  Accepted cost: the reviewers no longer resolve in repos without abcd installed.
- 2026-07-13 — Auto-merge is permitted ONLY to a non-protected trunk, gated on a SHIP
  review *verdict* (not merely green CI) + lint/smoke + an audit entry; never to `main`
  (explicit human `abcd spec ship` promotes). A bounded, opt-in reversal of the standing
  "a human merges" default — safe because the merge target is staging, not the protected
  branch, and the gate is a verdict, not a checkmark (green CI shipped a real leak during
  the 2026-07-12 drain; a security review then HELD it). Record homes: experience → itd-29
  (already scoped, deferred v2); enforceable form → a brief invariant + an ADR *when built*,
  not now (capture-now-build-later). SOTA is itd-29's (GitHub-native auto-merge, host-owns-
  git, no new dep); the ADR inherits it. Surfaced the `facilitator-default-thinker-optional`
  principle.
- 2026-07-13 — `abcd audit` (itd-85): a new read-only repo-conformance verb, distinct from
  `ahoy doctor` (doctor = tool-setup health, audit = does-the-repo-conform). Bespoke on
  `internal/core/lint` (adapt repolinter's rule-schema vocabulary + Conftest severity/exit
  codes + SARIF as an optional export), zero new deps → no dependency gate. v1 = five rules
  (three-tier-layout, conventions-router, decision-durability, docs-currency, privacy-hygiene);
  SARIF deferred to P3; wires into `prepare-this-repo` Phase 2, closing `iss-86`. SOTA-researched
  in plan `2026-07-13-abcd-audit-verb.md`.

- 2026-07-13 (itd-85 M1): kept `core.exists` (bool-only, swallows errors) and
  `ahoy.fileExists` (regular-file-only) as-is rather than folding all three
  `exists` copies into `fsutil.Exists`. Chose partial consolidation over the
  plan's full-consolidation because the other two hold different contracts;
  merging them would smuggle a behaviour change into a behaviour-preserving
  refactor. Only `lint.fileExists` (identical fail-closed contract) migrated.
- 2026-07-13 (itd-85, carry to M3): `gitutil.CheckIgnored` fails OPEN — git
  absent or not-a-repo returns "nothing ignored". The `three-tier-layout` rule
  MUST treat an empty result as "cannot tell", never as "compliant", or a repo
  with git unavailable silently passes the "is `.abcd/.work.local/` gitignored"
  assertion. Security review flagged this as the one consumer-side spec note.
- 2026-07-13 (itd-85 M2): audit engine uses severity vocabulary error|warn|off
  (repolinter/Conftest), NOT the record-lint engine's blocker|warn, because it
  maps directly onto the tri-state exit code (error->2, warn->1) and reads right
  in a human render. Reused docs-lint findings (blocker|warn) get mapped to
  error|warn at the docs-currency rule boundary in M3, not in the engine.
- 2026-07-13 (itd-85 M3): privacy-hygiene uses a deterministic, identity-INDEPENDENT
  absolute-path regex, NOT the identity-aware scanner (internal/adapter/scanner).
  Rejected the scanner because its home-path detection is identity-PARAMETERISED
  (kindHomeSelf=hardfail vs kindHomeOther=warn) — machine-dependent severity —
  whereas AC3's contract is "ANY absolute local path is an error", deterministic
  across machines. The scanner also scans the release BUNDLE (a curated allowlist
  excluding tests); audit scans all tracked files, a scope the scanner was not
  built for. Flagged for future consolidation: absolute-home-path detection now
  lives in two predicates (scanner identity matchers + audit regex); a later phase
  should extract a shared identity-independent path matcher.
- 2026-07-13 (itd-85 M3): docs-currency emits every finding at warn, downgrading
  docs-lint blockers, because audit is an advisory conformance surface and the
  authoritative docs gate is `abcd docs lint` (still exits 2 on a blocker).
  Re-raising a docs blocker as an audit error would double-gate the same check.
- 2026-07-13 (itd-85 M3): three-tier-layout does NOT require .abcd/.work.local/ to
  be present (diverges from the plan's literal "present and gitignored") — it is
  created on demand and a fresh clone has none; requiring presence would flag every
  clean checkout. The load-bearing assertion is "if present, gitignored". Mechanics
  revision, premise intact.
- 2026-07-13 (itd-85 M3): privacy-hygiene reads tracked files through os.OpenRoot
  (repo-root containment), not os.ReadFile. A leaf-only O_NOFOLLOW is insufficient
  — a symlinked INTERMEDIATE directory still escapes; security review PoC-confirmed
  an out-of-repo arbitrary read. os.Root refuses any escaping component. Plus
  O_NONBLOCK (FIFO/device non-blocking open) + IsRegular skip + 4 MiB size cap.
  Requires go 1.24+ (repo is 1.25); no new dependency.
- 2026-07-13 (itd-85 M7): acknowledged repolinter (rule schema) and Conftest
  (severity/exit vocabulary) in ACKNOWLEDGEMENTS now, since both are actually
  adapted in the shipped audit engine. DEFERRED the SARIF acknowledgement to P3:
  the serializer seam is shaped for SARIF but no SARIF is emitted yet, and the
  convention is to credit a pattern in the change that lands it, never ahead. Add
  the SARIF entry when the --format sarif serializer ships.

- 2026-07-13 (sensemaking method): recorded the ABCD method (cold reading / warm
  ledger / disposition) as a research note — the parent that itd-27, itd-42,
  itd-55, itd-86 and itd-87 had all been accumulating under without one. Minted
  exactly ONE principle (recurrence-is-signal) rather than one per method element:
  the cold/warm split is already stated by evaluator-outside-the-loop and
  verifier-selects-gates-decide, and one-canonical-primitive forbids a third
  near-copy. Recurrence was the only element with no counterpart in the record.
  REJECTED minting a `read-it-cold` principle for the same reason.
- 2026-07-13 (itd-86/87): recorded the two intents TOGETHER because they are
  coupled, not merely related — a blind cold reading re-raises old tensions by
  design, so pointing it at a ledger that dedupes them yields a detector fighting
  its own store. itd-87 is the precondition that makes itd-86's re-raising useful.
- 2026-07-13 (attribution): DEFERRED the ACKNOWLEDGEMENTS entry crediting the cold
  reading to abcd's co-author, pending confirmation of how they wish to be
  credited. Held loudly (stated in the method note), not silently; it must land
  before itd-86 ships. Do NOT guess the credit line.

- 2026-07-14 — The lifeboat is built as a COVERAGE EXPERIMENT, not a feature
  (adr-35, itd-88/spc-3). Probe before pack: `disembark probe <repo>` produces a
  cross-repo coverage aggregate BEFORE a packer exists, because the brief's
  structure is an untested assumption and building the packer first assumes the
  answer. The headline number is the delta in section coverage between a
  rich-record repo and a git-only one — that is what the record is worth, and if
  half the brief is permanently blank everywhere, the structure is wrong and we
  learned it for one milestone instead of a phase. Phase 6's "depends on every
  prior substrate being native" rationale was checked against the BINARY and found
  mostly false (spec engine ships; reviews are committed markdown; backgrounding is
  a host affordance; the itd-2 host-delegation seam already ships twice — memory's
  `--pages-json` and `intent review ingest --verdict-json`). The ONE real
  dependency is data, not code: `~/.abcd/` does not exist, `history.Capture` is
  called by nothing, and Pass B's corpus cannot be obtained retroactively — the
  only permanent, compounding cost on the board, which is why the transcript hook
  ships ahead of any lifeboat code. Rejected: building the packer first (assumes
  the answer); amending adr-4 in place (two of its three operative claims change —
  a replacement, not a clarification).
- 2026-07-14 — adr-4 SUPERSEDED by adr-35 and pruned per the ADR convention
  (superseded ADRs are pruned; git preserves the text; the successor carries the
  transition rationale). What survives is restated in adr-35: the lifeboat is
  regenerable output, and the `lifeboat`(noun)/`voyage`(verb) distinction is
  load-bearing. What changes: disembark is READ-ONLY and OUT-OF-TREE (a test hashes
  the source tree before and after), and `voyage/` moves to the OPERATOR level
  (`~/.abcd/voyage/<source-root-sha>/`, keyed like the history store). The voyage
  move is not cosmetic — voyage records absolute source paths, and the
  `privacy-hygiene` audit rule (itd-85) flags those in committed files, so abcd
  would have failed its OWN audit. adr-4's overwrite-with-`.bak` model is replaced
  by a destination safety gate (never overwrite a directory abcd did not produce);
  its `shared_with` field is dropped (nothing produces it, and an empty field is a
  lie in a schema); and its hash chain — asserted but never defined — is pinned.
  Nine inbound references repointed by hand (2 links, 7 prose/frontmatter).
- 2026-07-14 — The brief↔lifeboat mapping table now EXISTS. `00-meta.md` has always
  called it "the contract" while no such table existed anywhere (found by the
  2026-07-06 plan-consistency review). It lands as Go — `internal/core/lifeboat/
  mapping.go` is the single source of truth — and is rendered into `00-meta.md`
  between generated markers, with a test asserting the two agree so the document
  cannot drift from the code. It is framed as the experiment's HYPOTHESIS, stating
  the best status each brief section could reach at each source tier, in the SAME
  three-valued vocabulary the probe reports (`grounded`/`partial`/`blank`) so
  prediction and evidence are directly comparable. M2 is expected to revise it.
  A monotonicity test (a richer tier can never ground a section worse than a poorer
  one — tiers are CUMULATIVE) caught a real error in the first draft of the table.
- 2026-07-14 — Vocabulary registered in `02-constraints/04-naming.md`, fulfilling a
  claim adr-4 made and never kept (`voyage/`, `manifest_sha256`, `_provenance.json`,
  `history.jsonl` were absent from the registry). Added with them: `coverage.json`,
  `graveyard/`, and two new controlled enums — coverage `status ∈ {grounded, partial,
  blank}` and source `tier ∈ {git, conventions, abcd-native}`, both with the Go enum
  named as the machine-readable source of truth. The brief's `"sufficient"` oracle
  verdict — a member of NO registered enum — is retired in favour of the registered
  `{SHIP, NEEDS_WORK, MAJOR_RETHINK}`; no third verdict family is minted (four
  brief locations).
- 2026-07-14 — adr-35's blast radius across the record was FAR wider than the plan
  anticipated, and the line drawn is: **the brief, glossary and roadmap are
  reconciled; the intent corpus is NOT.** An adversarial review (four hostile lenses,
  every finding independently verified) found the first pass had rewritten the
  vocabulary registry to the new model while ~14 other files still asserted the old
  one as fact — including an INVARIANT (`03-invariants.md` #6), the product's own
  press release, the verification matrix (which encoded adr-4's `.bak` overwrite as a
  TEST GATE), and the lint-enforced glossary SSOT. A registry contradicting an accepted
  ADR is drift of exactly the kind iss-35 exists to prevent, so all of it was swept.
  The INTENTS (itd-2/8/9/10/13/15/19/22/24) were deliberately left alone and tracked as
  iss-94: an intent is a proposal with its own lifecycle, and silently rewriting nine of
  them inside an unrelated change is worse than recording the drift — each reconciles
  when it is next planned. Where adr-35 genuinely does not settle a question (where
  `embark scan` searches now that destinations are operator-chosen; what the `/abcd`
  status board reads now that there is no in-tree lifeboat to stat), the text carries an
  explicit `Open question (adr-35)` note rather than an invented answer.
- 2026-07-14 — iss-93: adr-35 promises disembark is READ-ONLY over the source (a test
  hashes the tree before and after), but two paths in the design still write into it —
  Pass-0 dev-sync (`.abcd/work/reviews/`, `.abcd/memory/`, `.abcd/work/issues/`) and the
  backgrounded-execution checkpoint (`.abcd/logbook/disembark/<ts>/_state.json`). Either
  they move out-of-tree (under `<dest>` or the operator-level voyage) or they leave the
  disembark path entirely. adr-35 does not settle it; the decision is owed before the
  packer ships, and the read-only test is what will force it.
- 2026-07-14 (M1, itd-89/spc-4) — Transcript capture is wired to `SessionEnd`, NOT the
  `Stop` the plan specified. The plan's letter was wrong on a matter of harness fact:
  `Stop` fires once per assistant TURN, and Claude Code's transcript file grows through a
  session, so a `Stop`-wired capture stores a fresh, larger superset every turn — proven
  by live test (one session, 4 turns → 4 records; a 100-turn session → 100 records and
  O(N²) bytes). `history.Capture`'s sha256 dedup only collapses byte-IDENTICAL
  re-captures, which never happens on a live transcript, so the plan's "re-capture is
  idempotent" acceptance is false under `Stop`. `SessionEnd` fires once at termination and
  by contract ignores exit code + stdout — a perfect fit for a fail-closed, non-blocking
  side-effect hook. Verified against the harness docs (code.claude.com/docs/en/hooks).
  Accepted cost, recorded not hidden: `SessionEnd` does not fire on a hard crash/SIGKILL,
  so an uncleanly-killed session is not captured; the `Stop`-with-session_id-dedup
  alternative that would recover that case needs a change to shipped core dedup semantics
  and is deferred. This is the M1 deviation the loop is required to surface.
- 2026-07-14 (M1) — iss-95: wiring the hook does NOT by itself start the clock. `history.
  Capture` requires `~/.abcd/history/<root-sha>/transcripts/` to already exist and
  deliberately never creates it (the `ownedDirsReal` symlink-safety discipline); `ahoy
  install` bootstraps it. On a machine where install has not run — INCLUDING THIS ONE,
  where `~/.abcd/` does not exist — `hook session-end` fails closed, logs to stderr, exits
  0, and captures nothing, silently. That is exactly itd-89's failure mode (a hook that
  looks wired while the corpus never accrues). Decision owed: hook self-bootstraps (changes
  Capture's precondition and has the hook create dirs, which the symlink discipline avoids)
  vs. `ahoy install` stays the sanctioned bootstrap and the not-installed case is made LOUD
  (ahoy doctor already flags `history.bootstrap_missing`). iss-96 records the adjacent point:
  automatic capture makes the scanner's secret-pattern coverage load-bearing — it catches
  anchored tokens (AKIA…, ghp_, sk-ant-) and home paths but not unanchored high-entropy
  values (a bare 40-char AWS secret, a prefixless token), so consider entropy detection or
  the gitleaks adapter for the transcript path.
- 2026-07-14 (M1, iss-95 — maintainer decision) — The store-not-bootstrapped case
  is made LOUD, not self-bootstrapped by the hook (rejects having `hook session-end`
  create `~/.abcd/history/`, which would put a dir-creating trust-boundary act inside
  a fail-closed hook and contradict the `ownedDirsReal` symlink discipline). Reality
  check: `ahoy install` ALREADY bootstraps the store (`bootstrapHistory`, plus the
  per-repo transcripts dir), and detection ALREADY emits `history.bootstrap_missing`
  as a required gap that bare `abcd`, `ahoy`, and `ahoy doctor` surface — so an
  installed user is never in the silent state. The only genuinely silent path is the
  `SessionEnd` hook itself, which by harness contract has NO output channel (its exit
  code and stdout are ignored), so it cannot speak at session end. "Loud" therefore
  lives where a channel exists: a SessionStart notice (SessionStart hook output is
  surfaced) that warns once when the store is absent, pointing at `/abcd:ahoy install`.
  Scoped as an M1 follow-up; keeps the hook fail-closed-silent and moves the loudness
  to the one event that can be heard.
- 2026-07-15 (M2 gate — maintainer-approved) — The lifeboat coverage experiment's
  cross-repo readout is in. Corpus (private repos anonymised): abcd-cli
  (git+conventions+abcd-native, 21/2/0 grounded/partial/blank), test repo 1 and
  test repo 2 (abcd-native scaffolding but no authored brief, 4/8/11 and 2/6/15),
  test repo 3 (git+conventions, no abcd, 3→4/8/11), and a git-only floor (0/2/21).
  Headline finding: **scaffolding is not a record** — test repos 1 and 2 carry
  `.abcd/` directories yet ground barely more than the record-less test repo 3,
  because their `.abcd/development/` has no `brief/`, no ADRs, no issue ledger; the
  native adapter is honest and grounds only authored prose. The
  brief structure holds (excluding the dogfood repo, 9 of 23 sections are blank across
  the messy corpus, not half). Decisions: (1) `product/personas` is demoted to a
  human-answered question in the lifeboat brief — the corpus confirms the M0 prediction
  that it is not derivable from a repository. (2) The other 8 always-blank sections stay
  in the brief but split: `product/mental-model`, `delivery/verification-matrix`,
  `delivery/out-of-scope` become human-answered questions; `evidence/what-didnt`,
  `evidence/open-questions`, `constraints/naming`, `glossary`, `internals` are blank more
  from thin adapters than genuine non-derivability and get adapter work before M3 decides
  (iss-98, iss-99, iss-100). (3) The dependency-manifest adapter under-detected Python/
  Ruby/PHP packaging (test repo 3's pyproject.toml+uv.lock read as blank) — fixed now, so
  test repo 3's `constraints/dependencies` grounds. M3 (the packer) builds to this list:
  grounded/partial sections extracted-and-cited, the human-question sections surfaced as
  the blanks-with-questions the coverage report already produces.
- 2026-07-15 (post-M2 gate design) — Graduated to adr-36 and itd-90. The coverage
  gate raised a lifecycle question the plan implied but never named: the coverage
  report knows what to ask, but nothing said who answers a blank, when, or where —
  and the person with the tool (facilitator) is rarely the person with the answer
  (product thinker). Decision (adr-36): a blank is a durable, fillable object, not
  a fill-now-or-lose snapshot; answering is decoupled from disembark/embark and runs
  as its own async, environment-agnostic step over the coverage JSON; blanks carry a
  `kind` (`extractable` = coverage debt abcd can fix, vs `human-owned` = personas/
  mental-model, never derivable and framed as a prompt not a failure — the durable
  form of the "personas is manual" gate call); and a filled blank is marked
  `authored-by` (a person + date), structurally distinct from a grounded section's
  `extracted-from` (a file), so an opinion never launders into a fact. Coverage
  schema grows to v2 (fillable object); mapping.go gains a per-section `Kind`. itd-90
  specifies the product-thinker-facing interview (draft). Boundary: distinct from
  itd-86 cold-reading (which reviews for contradictions, denied context; the interview
  answers questions, fed context — opposite direction).
- 2026-07-16 — M5 round-trip closure re-scoped: the plan's literal "re-pack of an
  embarked repo reproduces the same manifest hash" is structurally unmeetable
  (coverage/brief/archaeology are identity- and git-derived); ratified instead as
  (P1) record-derived sub-manifest closure — RecordManifestSHA256 over ADRs,
  issues, intents, specs, abandoned.json, recorded in provenance as
  record_manifest_sha256 — plus (P2) literal self-closure into a byte-copy of the
  source; the packer now carries specs (rescue/specs/) so the spec.Load
  round-trip assertion is real.
- 2026-07-16 — M6 synthesis, two plan amendments: the oracle audit is keyed by the
  lifeboat's manifest hash (audit/oracle-<manifest12>.json), not the plan's
  wall-clock <ts> (no timestamp ever enters a lifeboat artifact); and
  _provenance.json is never mutated post-pack — each synthesis artifact
  self-records its mode (deterministic|delegated) instead of the plan's
  "_provenance records which" (the commit marker stays immutable). Synthesis
  outputs (principles, press-release, audit/) join the lessons files outside
  manifest_sha256.
- 2026-07-17 — Burst 2 (run test B), two mechanics decisions: (1) an
  already-planned intent with spec_id null (itd-40, created before the
  lifecycle verbs existed) is routed through a transient git mv planned->drafts
  so `intent plan` can mint+link its spec fail-closed — there is no standalone
  spec-create verb, and hand-authoring a spec file would bypass the mint lock's
  id allocation; net record churn is the spec_id write. (2) Implementation was
  delegated to an Opus 4.8 worker (recorded protocol deviation, manual test B);
  the orchestrator re-ran the gate on the output, and commits carry the trailer
  of the model that authored them (worker code: claude-opus-4-8; orchestrator
  record work: claude-fable-5).
- 2026-07-17 — Burst 3 (M3, itd-4/spc-6), three record adjudications: (1) AC2's
  resolve note lives as the structured frontmatter scalar `resolution:` (the
  live setScalarField design), not body-appended prose as the AC letter says —
  recorded as intentional design evolution, not a gap. (2) AC3 (promote) is a
  genuine BLOCKED gap: skill-orchestrated by design but uncompletable with
  today's verbs (no intent-create until itd-46; no engine-backed related_intents
  back-link write; hand-editing frontmatter from markdown would violate the
  engine-backed convention) — spc-6 stays OPEN and itd-4 stays planned on it.
  (3) AC4 migration recorded satisfied-by-history (source absent, ledger
  populated iss-1..iss-103); no dead migration code built.
- 2026-07-17 — Burst 4 (M4, itd-46/spc-7): quoted-text create shipped; the
  seeded-draft shape spc-7 defines is canonical (the AC's "byte-identical to
  intent new" clause is historical — no such Go verb existed). Two intent
  scope bullets named old-system files with no native counterpart
  (commands/abcd/intent.md, docs/reference/commands.md) — adjudicated moot in
  the spec; the missing intent plugin-markdown surface is ledgered as iss-105
  rather than silently absorbed. Typo-guard asymmetry vs capture ledgered as
  iss-104, not fixed (ACs don't require it).
- 2026-07-17 — Burst 5 (M5, itd-43/spc-8): GL002 enforces a deliberate subset
  (['epic']) of the glossary's forbidden_synonyms — the others are common
  English words whose false-positive rate would sink the gate; each becomes
  opt-in via .abcd/record-lint.json as the corpus is readied. Two sweep hits
  inside internal self-quotes (itd-48 quoting a working-log line and spc-12's
  overview) were swept, not exempted: abcd's own records carry the canonical
  word even when quoting themselves. spc-8 stays OPEN and itd-43 planned on
  AC3 (spec-review token), blocked by itd-28's maintainer-gated dependency.
- 2026-07-17 — Release burst (maintainer-directed): adr-37 adopts
  changelog-driven releases (rolling Unreleased -> dated heading in a reviewed
  PR IS the release decision; auto-release.yml tags exactly that commit and
  calls release.yml as a reusable workflow; idempotent, GITHUB_TOKEN-only).
  Extends adr-31, does not replace it: number derivation stays itd-73's; the
  interim check is maintainer review of the roll PR. The detect step tolerates
  the historical [v0.1.0] heading style; new headings use the plain
  Keep-a-Changelog form. v0.2.0 rolls in the same PR as the port — the
  automation's first firing is its own acceptance test.
- 2026-07-18 — The iss-35 semantic release gate was self-referential
  (armed against the tagged commit, read receipts from that commit's own tree)
  and fail-closed the first public release; fixed in PR #99 by arming with the
  reviewed content commit (HEAD^2^ / HEAD^) from a full-history checkout, plus a
  check-reviews.sh RD001 exemption for sha-keyed receipt dirs. The gate is
  abcd-cli's OWN CI: it is NOT shipped or scaffolded to managed repos
  (launch-payload.json excludes .github/; ahoy/launch write no CI; lifeboat only
  reads .github/workflows as a grounding signal), so the flaw had no managed-repo
  reach — the iss-108 capture's "systemic" framing was corrected on resolve. Any
  future release-scaffolding intent should scaffold the fixed two-commit
  (roll -> receipts) pattern, not the original self-referential one.
- 2026-07-18 — Acceptance-criteria sign-off is implicit in a human running
  `abcd intent plan` (itd-94): no `ac_confirmed` frontmatter field, keeping
  directory-as-truth pure and adding no forgeable schema surface. The gate
  (`abcd intent ready`) distinguishes only drafts vs planned; agents are barred
  from unattended planning at the protocol layer (run-protocol step 0 +
  `/abcd:intent`'s interview script: `plan` is never run without the human's
  explicit in-session confirmation). Escalation path if violations appear: an
  `ac_confirmed_by:` field is lint-legal today and slots in as a fifth check.
- 2026-07-21 — itd-66 (`launch-payload-render-parity`) is DEFERRED to a follow-up
  after the derived-versioning + auto-changelog + distribution programme. itd-67's
  own AC frames its installability smoke as "light … later upgraded to call"
  itd-66's deep tier, so the split is clean: this programme builds the light tier
  and positions the surface-resolution seam so itd-66's deep render, `.abcd/**`
  leak-proof assertion, symlink resolution, parity diff, and isolated-subprocess
  deep smoke slot in as a drop-in upgrade rather than a rewrite. Ordering: this
  programme -> itd-66. Recorded in spc-11's "itd-66 is deferred" section.
- 2026-07-21 — itd-67's "a phase completed since the last launch -> minor" bump
  heuristic is SUPERSEDED by itd-73's `impact` derivation (spc-10). The bump falls
  out of the records' declared `impact`, not out of phase membership; spc-11
  consumes the derived number and never computes one. Recorded so the two intents
  do not both claim version selection.
- 2026-07-21 — `launch ship` does not tag. It writes the dated `## [X.Y.Z] - <date>`
  heading in the reviewed ship PR; the unchanged `auto-release.yml` greps it on
  merge and creates the tag. ADR-37 is preserved, not superseded: the reviewed ship
  IS the release decision, and the bot-on-main alternative stays rejected.
- 2026-07-21 — `impact` is a KNOWN property of the issue ledger schema, and
  `internal/core/capture` validates it against the shared enum in
  `internal/core/changelog` rather than a private copy. The back-fill added
  `impact:` to every resolved issue, which `validateStrict`'s
  additionalProperties:false allow-list rejected — `abcd capture` reported
  "resolved 0" and skipped all 57 records as malformed. Accepting the field
  without validating it was rejected: severity/category/source are all
  enum-checked on read, and a third definition of the impact enum is exactly what
  spc-10 exists to prevent. capture -> changelog is the import direction (no
  cycle: changelog imports launch/frontmatter/gitutil only).

- 2026-07-21 — The mapping table's per-tier status columns are a **ceiling**,
  not merely a prediction. Every conventions adapter already honours its row
  (`convGlossarySource` returns partial where the row says partial;
  `convPlatformSource` returns grounded where it says grounded), so itd-95's
  and itd-96's three new adapters cap at `StatusPartial` — the value all three
  rows predict — and carry signal strength in `Confidence` instead. Rejected:
  returning `StatusGrounded` for a dedicated `NAMING.md` or a prose-bearing
  `ARCHITECTURE.md`, which would have made the rendered brief table wrong and
  required editing `mapping.go`, the brief-to-lifeboat contract both intents
  put out of scope. Every acceptance bar asks only for "non-blank", so the
  ceiling satisfies them. Revisit only by amending the mapping row first.

- 2026-07-21 — A probe walk of a foreign tree must be bounded in **three**
  dimensions, not one. itd-95 shipped `WalkFiles` with a regular-file cap; an
  independent security review of the itd-96 branch showed both remaining
  dimensions were exploitable — a tree of directories holding no regular file
  never reaches a file cap, and `os.Root` re-resolves each directory from the
  containment root one component at a time, making a directory chain quadratic
  in its depth (depth-1500 did not finish in two minutes). Directories are now
  counted against the same cap and descent is capped at `maxWalkDepth`. The
  general rule: any new whole-tree traversal states which of {entries, depth,
  aggregate bytes} bounds it, because a per-item cap and a count cap multiply
  and their product is not a bound.

- 2026-07-22 — A release is a release regardless of path: the clean-cutover
  manual roll follows the SAME two-commit release-branch shape as a derived
  ship (the content commit, then a receipts commit carrying the sha-keyed
  PROMOTE receipts for docs-currency-reviewer and iss35-brief-surface-crosscheck).
  The v0.4.0 roll landed as one commit with no receipts; the receipt gate
  refused fail-closed — its first genuine firing, and correct. Recovery uses
  the workflow's own escape hatch (tag on the receipts commit; content = tag^).
  Rejected: weakening the required-gates list to unblock the tag — that edits
  the release contract the whole programme was built to preserve.

- 2026-07-22 — Detector findings are triaged adversarially BEFORE fixing, and
  the numbers justify it: the full-depth iss35 crosscheck returned 102
  discrepancies; independent refuters confirmed 95 and killed 7 (two
  cross-direction duplicates, two wrong-reality, three legitimately
  staged/exempt). 93% precision is high enough to trust the detector and low
  enough that unfiltered fixing would have written seven falsehoods into the
  record. The refutation criteria are the detector's own exemptions plus
  adr-5; the refuters re-probe the binary rather than trusting the finding.

- 2026-07-22 — When a brief is agreed upfront, it is a TARGET document; it
  becomes the state document claim-by-claim, at ship time, because shipping
  includes the brief row edit (spec-moves-with-the-surface). Unbuilt design
  lives in intents or under an explicit staged marking — never as unmarked
  present-tense brief prose. The 95 confirmed discrepancies are that ratchet
  skipped at scale; the crosscheck is the measuring instrument; promoting the
  principle to a mechanical discipline is the open follow-up (iss-121, iss-122).

- 2026-07-24 — iss-35 resolves: everything it stayed open for is verified
  shipped (surface_coverage armed as blocker, 16 surface chapters incl.
  docs/history/version, skills→commands relabel, machine-checked staged Status
  column, crosscheck wired as release gate). The only residue is iss-122.
  Decided in a maintainer grill interview this date.

- 2026-07-24 — iss-122 design: the crosscheck gate gets a committed input
  manifest (doc list, directions, checker count, prompt hash) with TIERED
  depth — full depth for feature/breaking releases, Direction-B-only shallow
  pass for patch. The receipt must echo the manifest hash AND the tier;
  receipt_gate refuses a receipt whose tier mismatches the release's declared
  impact. Automated refusal is PROCEDURAL ONLY (manifest/tier mismatch,
  undispositioned findings); confirmed findings go to the maintainer, whose
  PROMOTE with recorded dispositions is the gate (verifier-selects-gates-
  decide). Rejected: hard-blocking on confirmed majors (a stochastic LLM
  triage could block a release); never-worse ratchet (drift persists; tiers
  make counts incomparable).

- 2026-07-24 — itd-93 design settled in maintainer grill: (a) surface is a
  `launch` sub-verb (extends 04-launch; rejected: new top-level verb, ahoy
  install step, embark-time family); (b) SELF-SCAFFOLD PARITY — abcd-cli's own
  release workflows are regenerated from the shipped template and a test
  asserts parity, so the proven pattern and the template are one artifact
  (rejected: lockstep diff test between two artifacts; frozen verbatim copy);
  (c) built-in workflow_dispatch REHEARSAL mode — arms the full gate against a
  simulated roll, publishes nothing; a green rehearsal is the runbook
  precondition for the first real release; (d) the changelog dated-heading
  format IS the itd-73 seam — `launch ship` is one optional producer.
  Promotes standalone / severity minor, PRD synthesised from the interview
  (no grandfathering); the 5 seeded ACs stand, amended for rehearsal mode.

- 2026-07-24 — itd-28 gitleaks sign-off: Stage 2 runs through the scanner
  seam — native patterns are the default engine, gitleaks is the stronger
  opt-in adapter when the binary is present, and the hook reports which
  engine ran (loud staging). CI's gitleaks pass stays the authoritative
  backstop; iss-96 pattern parity folds into the spec. No new hard
  dependency — adr-22 holds. Rejected: gitleaks as hard local dependency
  (first adr-22 exception, per-machine install burden); CI-only Stage 2
  (secrets reach pushed remote history before CI sees them).

- 2026-07-24 — Next-run queue reshaped by the grill: the small consequences
  fold into Track 1 (resolve iss-35; implement iss-122 pinning; amend +
  promote itd-93 — promotion only, implementation later as its own focused
  run); the spine stays friction fixes → itd-94 → walk fixes → itd-88.
  itd-28 implements in the following run against the adapter decision.
  Queue: plans/2026-07-24-next-run-queue.md (supersedes the 2026-07-18
  queue file's pick-up role).

- 2026-07-25 — Record-id collisions across parallel branches (iss-115, iss-120)
  fixed as a class: one canonical allocator primitive (recordid.MaxAcrossRefs)
  mints max+1 over the union of the working tree AND every git ref, so a
  committed id on another branch is seen; git-unreadable-over-a-repo degrades
  loudly to tree-only, a non-repo mints quietly (no refs to collide). Detection
  completes the class — a new spec_id_unique record-lint rule via the shared
  validateIDUnique primitive (ADR ids left out of scope: this directive is
  iss-/itd-/spc- only). Rejected: id-range leases per programme, mint-at-merge,
  and non-sequential/hash ids — all trade away human-readable sequential ids
  without a maintainer mandate. The residual uncommitted-mint window (both
  branches mint before either commits) is accepted behind the armed detectors on
  the merged PR union.
- 2026-07-26 — itd-100 grill settled: crosswalk is a mapping not a registry (no canonical native definitions; gloss-only rows, docs/** links only, pending iss-40); alphabetical + thematic mini-index; footnote citations with per-line docs-lint allow for vendor names; British English. Positions: A2A and AP2/payments both WATCHING (AP2 via new capture — record silent, REJECTS never invented; x402 + OpenAI/Stripe ACP folded into the AP2 footnote); policy engine folded into policy-as-code row; agent skills is the sole REJECTS (brief 08-skills commands-only rule). Rejected-term admissions blessed (agentic workflow, ANP, AGNTCY, Verifiable Intent, harness engineering, ISO 22989-as-HITL-anchor). LinkedIn anecdote omitted from itd-100. Two OpenAI harness URLs await maintainer read before ship.

- 2026-07-14 (review): REJECTED a "super-reviewer" verb gated on a second model
  backend. Its motivating rationale does not survive: self-preference under genuine
  authorship is weak or absent (arXiv 2606.20093, gap -5.1pp, CI crosses zero), and
  the bias that DOES exist on code is a context-framing artefact, not model identity
  (arXiv 2603.04582). CHOSEN instead: a fresh-context, off-policy, same-model
  reviewer — re-present the diff in a new session as an artefact of unknown
  authorship. Captures the only measured debiasing effect, costs nothing, needs no
  second subscription, and removes the two-tier UX problem entirely. A second model,
  when present, is used for disagreement-as-triage per adr-25 asymmetric trust and
  itd-81 — NEVER a panel or a vote (nine judges = 2.18 effective votes; on code,
  consensus underperforms the naive baseline — the "popularity trap", arXiv 2510.21513).
- 2026-07-14 (review): REJECTED per-commit review as the unit. No serious tool does it
  (all per-PR); findings on WIP commits are false positives by construction; cost is
  $75-500 per PR-equivalent at published frontier rates. Unit of review is the BRANCH
  DIFF. Also rejected: spec-conformance as a BLOCKING gate — best agent scores 44.4%
  on the easier adjacent task (SpecBench) and no precision/recall for
  implementation-vs-spec is published anywhere. It ships advisory only, per
  verifier-selects-gates-decide. Full evidence:
  .abcd/development/research/2026-07-14-cross-model-review.md
- 2026-07-14 (research): REJECTED building an authoritative benchmark from abcd's own
  data. Three independently fatal defects: (1) the corpus is n=2 labelled triples, both
  graded by the model that wrote the code, with exactly one NOT_MET; (2) circularity —
  the system would generate the spec, write the code, judge conformance AND supply the
  labels, and self-preference appears perplexity-driven so it SURVIVES swapping model
  family; (3) N=1 has no methodological remedy (Epoch attacks SWE-bench Verified for
  concentration at 12 repos; we have one). Arithmetic also fails: ~500 oracle-backed
  tasks needed to resolve model differences, in-situ A/B needs 2,200-15,500 paired
  observations. CHOSEN instead: a task-quality instrument + methodology, validated
  against an EXTERNAL reference — OpenAI's 731-task SWE-bench Pro audit. Ground truth
  must originate outside the tool (executable test or retrospective event oracle, never
  an LLM judge). Template is Aider polyglot: the tool supplies the harness and
  distribution; ground truth comes from outside (225 external Exercism exercises).
- 2026-07-14 (research): ADOPTED pre-registration of acceptance criteria — hash +
  timestamp at spec time, BEFORE any code. The novelty claim is that abcd writes the
  specification before the code exists, which is exactly the defect that killed
  SWE-bench Verified (35.5% of audited tasks demanded implementation details never
  stated in the problem) and SWE-bench Pro (retracted 2026-07-08). Without a commitment
  artefact the claim is unprovable and CANNOT be retrofitted — every intent shipped
  without it is lost evidence. The primitive exists (receipt_id already hashes the AC
  section, excluding Audit Notes).
- 2026-07-14 (review): ADOPTED splitting the conformance reviewer into a terse binary
  verdict call and a SEPARATE explanation call. Asking for verdict+explanation+fix in
  one call drives GPT-4o's false-negative rate from 26.2% to 73.2% (HumanEval) and
  35.9% to 87.9% (MBPP) — it does not find more bugs, it REJECTS CORRECT CODE
  (arXiv 2603.00539, Springer Automated Software Engineering). All five abcd agents
  currently do the forbidden thing. Largest measured effect size in the research and
  free to apply. Also adopted: fix-guided verification filter (execute the proposed fix
  as a counterfactual; if no test outcome changes, the rejection was hallucinated —
  FNR 54.8% -> 16.3%). Full evidence:
  .abcd/development/research/2026-07-14-research-platform-benchmarks.md
- 2026-07-14 (research, UNVERIFIED): the SWE-bench Pro audit figures (27.4% automated /
  34.1% human) and the SWE-bench Verified figures (>=59.4% flawed tests / 35.5% hidden
  implementation details) come from SECONDARY sources — openai.com 403s to the research
  fetcher. These numbers are the reference standard the proposed paper is scored
  against. VERIFY AGAINST THE PRIMARY before writing anything. Do not cite the decimals
  until read first-hand.
- 2026-07-14 (research, CORRECTION — supersedes the two entries above): the SWE-bench
  figures are now VERIFIED against OpenAI's primary posts, and one claim is WITHDRAWN.
  The "35.5% of audited tasks demanded implementation details never stated" figure DOES
  NOT EXIST in the primary — it came from secondary reporting. OpenAI's real taxonomy
  for SWE-bench Pro (% of full dataset, agent-flagged / human-flagged, read per-bar from
  the chart's aria-labels): overly strict tests 14.4/17.8; low-coverage tests 4.1/9.4;
  misleading prompt 6.3/7.5; miscellaneous 1.9/1.2; UNDERSPECIFIED PROMPT 0.6/0.8.
  Underspecification is the SMALLEST category (~1%); "overly strict tests" is the largest
  by ~20x. CONSEQUENCE: the pitch "abcd's specs-before-code prevents the underspecification
  that killed these benchmarks" is aimed at a 1% defect and is DEAD. The surviving claim is
  stronger and is what the work now defends: SWE-bench's oracle (a merged PR's tests) is
  authored independently of, and after, the task statement, so it can demand what the spec
  never said — whereas in abcd THE ACCEPTANCE CRITERIA ARE THE ORACLE, so the oracle cannot
  exceed the spec. abcd therefore cannot OVER-reject; it can only UNDER-check (the
  low-coverage failure, 4-9%), which compounds with judges over-accepting AI-written code
  by up to 1.91x. Claim to defend: "a spec-derived oracle cannot over-reject; it can only
  under-check — and under-checking is measurable." Verified figures: SWE-bench Verified
  (2026-02-23) 27.6% subset audited, >=59.4% flawed tests, 138 problems x >=6 engineers;
  SWE-bench Pro (2026-07-08, retracted) 731-task split, 200 (27.4%) agent-flagged vs 249
  (34.1%) human-flagged — the two audits disagree by ~7 points, i.e. automated task-quality
  auditing UNDER-DETECTS relative to humans.
- 2026-07-14 (process): TWICE this session a research pass reported a paper as saying
  something it did not (a fabricated cross-family finding; the non-existent 35.5% figure;
  and SpecBench described as "deferring" conformance when it scopes code out permanently).
  RULE: for any number that will be load-bearing in a design doc or paper, open the PRIMARY
  before citing it. Secondary reporting of benchmark figures has been wrong every time it
  was checked in this session.
- 2026-07-14 (layout): ADOPTED "default to the local tier when in doubt". An artefact
  whose home is unclear (tool exports, oracle/review output, traces, intermediate
  analysis) goes to .abcd/.work.local/scratch/ or logs/ FIRST — never the repo root,
  never a tracked directory on a guess. Promotion to .abcd/work/ or
  .abcd/development/ is cheap and always available later; demotion is not, because a
  wrongly-committed artefact is already in the history. Guessing upward is
  irreversible; guessing downward costs nothing. Prompted by a RepoPrompt oracle export
  landing in a top-level prompt-exports/ during this session (moved to
  .abcd/.work.local/scratch/prompt-exports/). Recorded in AGENTS.md § Working-tree
  layout, the canonical home, rather than a new doc.
- 2026-07-27 (salvage): the four 2026-07-13/14 ideation-session entries above are appended out of chronological order — recovered from an unmerged ideation branch during branch cleanup, together with their research/plan docs; the /abcd:ideate seed from that session is re-recorded as itd-104 (its branch-local iss-93 capture id had collided with main's iss-93 and is retired unused).
- 2026-07-27 — DEFERRED the row-has-footnote structural docs-lint rule (spc-15 out-of-scope): no existing rule shape covers table-row-to-footnote structure, so it would need a bespoke check; the grill left it optional and the citation-gate intent (itd-101) is now its natural home — implement it there or not at all. Recorded per spc-15's deferral-is-recorded clause; surfaced by the itd-100 fidelity audit's gap check.
- 2026-07-27 — ADOPTED the canonical tagline: "A host-agnostic configuration layer for intent-driven development." Canonical identity home is the brief product chapter (01-product README, Identity section: title / tagline / pitch); README strapline, plugin manifest description, and AGENTS.md opening render from it. README's former strapline ("An opinionated, intent-driven development framework for product thinkers") is retired as a surface line; the product-thinker framing lives on in the README body. Rejected wordings: "for agent harnesses" (object shift, host/harness redundancy, plural over-promise), "over any agent harness" (drops the domain). Resolves iss-143; itd-102 generalises the drift check for managed repos.
- 2026-07-27 — itd-93 parity tension resolved: Branch A. The scaffold template is the single source for release.yml/auto-release.yml; abcd-cli's live workflows are regenerated from it; the only sanctioned live diff is the additive workflow_dispatch rehearsal (trigger + non-publishing dry-run job). Rejected: substitution-gating the rehearsal off for abcd (weakens the parity guarantee on the one novel component).
- 2026-07-27 — GRILLED itd-101/102/103/104 to settlement (six decisions, each now in its intent's Grill Settlements + acceptance criteria): itd-103 guard fails open loudly, blocker/warn tiers with committed-only overrides, shell-token-aware command-position matching with an itd-81-style TNR-floored corpus; itd-101 refresh is manual-with-nagging (scheduled CI later, own sign-off), staleness warns at 180d and blocks releases at 365d, human verifications age on the same clock; itd-102 identity lives in a parseable markdown block (config points at it), check is warn-tier and never rewrites; itd-104 ideate is an optional verb, never a pre-capture gate, with a fresh-context off-policy adversary. All four drafts are grill-settled and await intent plan.
- 2026-07-27: itd-101 spec minted as spc-16 collided with unmerged #156 (iss-80 class, both spc-N and iss-N allocators); renumbered to spc-17 by hand, duplicate capture folded into iss-80 — ids on this branch deliberately skip 16.
- 2026-07-27 — itd-101 part (a) settled four choices spc-17 left open. (1) The committed baseline lives at `.abcd/citations-baseline.json`, alongside docs-lint.json/record-lint.json/rules.json — it is config the gate reads every commit, not development record; rejected `.abcd/development/` (wrong tier) and `.abcd/work/` (not a session artefact). (2) A table is a CROSSWALK when its nearest preceding heading matches `(?i)crosswalk` (configurable) — the narrowest heuristic that selects docs/reference/terminology.md's table and leaves every ordinary directory-map/comparison table alone; rejected column-header and any-table-in-docs identification (both sweep up README.md and docs/README.md tables that carry no citations by design). (3) A page's CITATION CORPUS is its footnote definitions, including wrapped continuation lines — not its prose; a URL in a body paragraph stays links_resolve's business, which is what keeps the syntax and source-policy rules off ordinary text. (4) `refused_domains` ships EMPTY: the repo states its admission rule in prose ("no single-author coinages, no aggregators", docs/reference/terminology.md + ACKNOWLEDGEMENTS.md) but nowhere names a domain, and the gate must not invent an editorial blacklist the project never agreed. Also: staleness is measured from `last_checked` for automatic and manual entries alike (AC 3's "same clock"), and the 365-day threshold surfaces as a distinct rule id `citation_baseline_overdue` at warn — the commit gate never calendar-blocks, so promotion to blocker is the release gate's job in part (b).
- 2026-07-27 — itd-101 part (b) settled seven choices spc-17 left open, and consolidated one primitive. (1) The SSRF fetch guard moved from `internal/core/memory` to `internal/urlguard` rather than being copied for the second fetch path; its address predicate became a parameter so a fetch path can be exercised against an httptest server (which binds loopback) without the shipped policy ever being relaxed. (2) BLOCKED-FOR-AUTOMATION is classified by status code ONLY — 401, 403, 406, 429 — with no body sniffing for challenge-page markers: a heuristic over page text would be unreliable AND a reason to start reading content, and the challenge pages that matter answer with one of those codes anyway. Everything else non-2xx is broken. (3) One GET per URL, body never read and never retried: liveness is the status line, so a citation to a huge file costs a response header; a retry loop would make a run's duration a function of how many links are failing. (4) A blocked URL with NO prior entry writes NOTHING — the gate then reports it as unreceipted, which is exactly true; recording it broken would put a lie in a committed record the gate enforces. (5) A CURRENT manual receipt is preserved verbatim and not even re-requested; a STALE one is re-checked (AC 3's one clock) and, if the source still blocks, KEPT rather than deleted so the gate keeps warning honestly. Rejected: refetching every manual entry every run (downgrades a human's receipt to a robot's failure), and never refetching them (buys a permanent exemption from ageing). (6) `confirm` records ALIVE only — it is "I looked and it is there", never a channel for recording a link dead — and takes URLs positionally or a receipt file, both assembling ONE schema so the later generated checklist page is a different producer, not a second pathway. (7) `refresh` exits ZERO after recording broken links: it records, the gate decides; a verb that failed on a dead link could never write the record that reports it. Also: the release-gate promotion reuses `ArmReceiptGate`'s shape (`lint.ArmCitationOverdue`, armed by `abcd docs lint --release-gate`) with the FLAG as trust root, and the `abcd launch --dry-run` citation gate takes its measurement from the CLI as data because `core/lint` imports `core/launch` for its semver — the same shape the cobra-tree walk already uses.
- 2026-07-27 — itd-101 part (b) review round settled five more. (1) The committed `citation_baseline.baseline` config value is CONTAINED at one choke point (`lint.CitationPolicy`, which now returns an error): a `../../..` or absolute path is REFUSED, never silently normalised, because `SaveBaseline` MkdirAll's its parent and a contributor controls that file — this was a real arbitrary-file-write primitive found by the security review, reproduced against a built binary. Both writers and both readers resolve through the one check. (2) A refresh run in which NOTHING succeeded is REFUSED rather than committed, and only when a prior baseline existed: `Check` collapses DNS failure, refused connection and timeout into the same `broken` as a 404, so a run behind a captive portal would otherwise rewrite every entry as broken — stale human receipts included, since those are re-checked and so not covered by the blocked branch — and the operator would find the gate blocking every commit. A first run over genuinely dead citations has nothing to protect and still writes. (3) An unrecognised `Status` from a `Checker` is an ERROR, not a fall-through: the seam is exported for AC 5's adapter, and falling through would drop the URL from BOTH the baseline and the queue — the one outcome nothing downstream could detect. (4) The missing-entry lint message names BOTH verbs: the only state that produces a missing entry is a source refusing automated fetchers, so naming `refresh` alone sent the maintainer round a loop that cannot terminate — only `confirm` clears it. (5) A docs-lint config that arms the citation rule but fails to PARSE is reported to the release preflight as `Unreadable` and REFUSES, distinct from the nil "not armed" state — rendering a broken gate as an absent one would wave a release through on a false statement. Also consolidated: `daysBetween` is now the exported `lint.DaysBetween`, called by the verb and the gate alike, because two copies of the staleness boundary is two chances for the verb to call an entry current on the day the gate calls it overdue; and the CLI's ad-hoc `citeErrorDetail` was dropped for the canonical `scrubPaths`, which preserves the `*PathError` type the redactor needs (an absolute developer path was reaching machine output).
- 2026-07-28 — itd-101 part (b) second review round closed three holes the first round's fixes left open, all reproduced end-to-end by the reviewers. (1) The baseline-path containment check was purely LEXICAL, so a committed symlinked directory (`.abcd/evil -> /outside`) plus a lexically-innocent `"baseline": ".abcd/evil/x.json"` reopened the very arbitrary-file-write primitive the first fix claimed to close — `WriteFileAtomic` MkdirAll's *through* the link. Containment now also RESOLVES: `EvalSymlinks` on the deepest existing ancestor of the target (it usually does not exist yet on a first run), compared against an `EvalSymlinks`'d repo root, reusing the shape `internal/core/launch/bundle.go` already uses for symlinked bundle entries rather than adding a third copy. `CitationPolicy` now takes the repo root and returns a `CitationPolicySet` carrying both the repo-relative form (for messages, so no absolute path is ever rendered) and the absolute form (for I/O). (2) `cfg.Roots` had NO containment, and while reading outside the repo was inert for a reporting lint, this intent made the collector feed a live fetcher whose results are persisted into a baseline the workflow expects to be committed and pushed — `"roots": ["../private"]` therefore fetched and PUBLISHED every URL in a sibling directory. The collector (not the pre-existing lint walk) now refuses an escaping or symlinked root. (3) The wholesale-failure guard consulted `res.Preserved == 0`, but a preserved receipt was never fetched and so is no evidence the network worked: any repo that had ever run `confirm` — the designed steady state for robot-refusing sources — silently lost the protection entirely. The guard now reads a new `CheckOutcome.Answered` bit that the CHECKER sets when a host returned any status line at all. That is the bit `StatusBroken` was destroying (a DNS failure, a refused connection and a genuine 404 all landed there), and recovering it from a proxy was wrong in both directions — the old condition also made a sole genuine 404 permanently unrecordable, refusing forever with a false "check connectivity". RULE affirmed: when a guard needs a fact, carry the fact, never infer it downstream from something correlated.
- 2026-07-28 — itd-101 part (b) third review round closed the other half of the roots containment and recorded one accepted narrowing. (1) Containing the configured ROOT string is NOT enough: `WalkDir` yields a symlinked `.md` as an ordinary file and `os.ReadFile` follows it, so a committed `docs/leak.md -> ../private/notes.md` sits inside a perfectly contained root and still drags an outside file's citations into the live fetch and then into the committed, pushed baseline. Every collected FILE is now resolved-contained, not just its root; an in-repo symlink (the `CLAUDE.md -> AGENTS.md` bridge shape) still resolves, because containment is about where a path LANDS. RULE affirmed: a containment check on a directory says nothing about the files found under it. (2) ACCEPTED NARROWING in the refresh's wholesale-failure guard: keying it on `answered` (a host returned any status line) rather than on success means an intercepting proxy or DNS sinkhole that answers EVERY request with a non-2xx outside the blocked set — 503, 407, or a default-vhost 404 — no longer trips it, and a whole corpus would be rewritten `broken`. That is strictly narrower than the condition it replaced, but the replaced one was wrong in two commoner ways (any repo that had ever run `confirm` lost the protection entirely, and a sole genuine 404 could never be recorded). It ships because refresh is operator-initiated and its transcript prints the `broken:` count before anything is pushed; the guard covers failures that produce no status line at all, and nothing more. (3) A CHANGELOG entry must describe the tree AT ITS COMMIT, never the state it will reach once a human acts: rewording the citation bullet to the post-confirmation state asserted "every URL carries a receipt" while the repo's own gate printed the counter-evidence. Reverted to the true count.
- 2026-07-28 — itd-101 part (b) closed two robustness notes from the approving security review, and flagged two it left. (1) A page read is now BOUNDED and regular-file-checked (`fsutil.ReadGuarded`, 8 MiB), because containment answers "does this path land inside the repo" and says nothing about how big what it lands on is — and this collector's output is fetched over the network. It opens the RESOLVED path rather than the literal one: `O_NOFOLLOW` on the literal path refuses every symlink leaf, which would break the legitimate in-repo bridge (`CLAUDE.md -> AGENTS.md`) that containment has just approved. So `containedRealPath` returns the resolved path alongside its verdict, and the reader reads exactly the thing containment judged. (2) `StatusOK` and `StatusBlocked` now ENTAIL an answer in the wholesale-failure guard: a 2xx cannot exist without a host replying, and "blocked" is defined by a status code, so only `StatusBroken` is genuinely ambiguous. That is what those statuses MEAN, not inference from a proxy, and it removes a footgun where a third-party adapter omitting `Answered` would refuse a run in which every check succeeded. The `Checker` seam's doc comment now states both obligations an implementer owes. NOT DONE, deliberately: (a) `urlguard.BlockedIP` still omits CGNAT `100.64.0.0/10` and benchmark `198.18.0.0/15` — adding them would change `memory ingest` behaviour, and that extraction was landed as a behaviour-preserving port, so it belongs in its own change; (b) a narrow availability path remains where a repo whose entire baseline is current-manual plus one PR-added citation to an unroutable host refuses the whole refresh with a misleading "check connectivity" — the reviewer could not make it bite on a realistic corpus, and the guard's limits are already recorded above.
- 2026-07-27: itd-103 guard overrides live in dedicated committed .abcd/guard.json, not a rules.json domain — structured entry schema, and the rules kill switch must not silently disable a safety guard (spc-16).
- 2026-07-27: itd-103 guard wiring — `abcd guard hook` is a sub-verb of the USER-facing `guard` verb, not a fifth entrypoint under the hidden `hook` subtree. Reason: guard health has to be legible (AC 1), and a hidden adapter documents nothing; the other `hook` entrypoints are injection transport with no user-facing counterpart, this one is the same decision the user can also ask for by hand. Consequence: it appears in the generated CLI reference and must stay host-agnostic in its prose.
- 2026-07-27: itd-103 guard wiring — the two front doors disagree on ONE case by design. A guard that cannot be evaluated (unparsable command, registry that will not load) is a FAULT for `guard check` (exit 2: a script must never read silence as clearance) and fail-OPEN for `guard hook` (exit 0 + loud stderr: a broken guard must never stop a session). Rejected: one shared behaviour, which would either brick sessions or teach scripts that silence means safe.
- 2026-07-27: itd-103 guard wiring — the fail-open-loud shim is a shell wrapper in the plugin's `hooks/hooks.json` PreToolUse entry (pass through only the binary's own 0/2; everything else warns UNGUARDED and allows), not Go code. The failure it guards against is the Go binary not running at all, so it cannot live in Go. Tested by executing the committed manifest string under /bin/sh against a fake plugin root.
- 2026-07-28: itd-104 spec minted as spc-16 (third live iss-80 instance this run: 16 collided with #156, 17 with the itd-101 branch); renumbered to spc-18 by hand.
- 2026-07-28: itd-102 spec minted as spc-16 (fourth live iss-80 instance this run); renumbered to spc-19 by hand.
- 2026-07-28 — itd-102 implementation: repo positioning lives in a NEW `internal/core/positioning` package, not `internal/core/identity` — that package is the git commit-author gate (`.abcd/config/identity.json` pin, pre-commit hook) and shares nothing with repo self-description but the English word; folding them would put two unrelated concerns behind one name. The user-facing verb stays `abcd identity` (spc-19). For the same collision the registry is `.abcd/positioning.json` at the `.abcd/` top level (beside docs-lint.json / record-lint.json / rules.json), NOT `.abcd/config/identity.json`. Rejected: extending the commit-identity package; naming the verb `abcd positioning` (spc-19 fixes the verb).
- 2026-07-28 — itd-102: surface comparison is normalised CONTAINMENT of each required block field (markup, dashes, wrapping, case folded; trailing sentence punctuation trimmed from the needle), not line equality. Equality would flag abcd's own three conforming surfaces, each of which carries the tagline in a different rendering (inside `<p>`, concatenated with the pitch in the manifest, bolded mid-sentence and line-wrapped in AGENTS.md); containment catches all three iss-143 variants while accepting all three current ones. Consequence: `identity render` fires only on drift, and its template render is a proposal to adopt by hand, not a byte-exact reproduction of conforming prose.
- 2026-07-28 — itd-102: no scaffold entry point existed in ahoy/prepare for a repo record block, so the verb family gained a minimal `identity init` (write path: block + pointer, atomic, adopts an existing block, refuses to repoint an adopted registry). `abcd launch scaffold` is release-machinery-specific and `ahoy install` writes only abcd's own plumbing; extending either would have widened its remit.
- 2026-07-29 — privacy-leak follow-up: illustrative machine identifiers in anything committed or published always come from reserved documentation ranges (RFC 5737 IPv4, RFC 3849 IPv6, RFC 2606 domains, RFC 7042 MACs; hostnames derived from the persona registry, e.g. alice-laptop) — the itd-79 persona rule applied to infrastructure. Recorded as principle `examples-use-reserved-identifiers`; enforcement is the iss-154 allowlist-inversion lint (flag any identifier OUTSIDE the reserved ranges), whose shipping promotes the principle to a discipline. No new intent: real private identifiers are itd-74's banlist territory — the incident evidence extends itd-74's scope to machine identifiers (iss-158) rather than founding a second primitive. Incident issue text deliberately names no repo and no values (shape only), so the ledger itself cannot re-leak.
- 2026-07-29 — v0.5.0 scoped as "security & consistency" (plan: plans/2026-07-29-v0.5.0-security-and-consistency.md): the security half closes the NEXT.md leak class end-to-end (itd-74/spc-20; the atomic iss-154+157+125+153 detector item; iss-155/156; guard batch iss-159+144+148; boundary fixes iss-30/34), the consistency half retires the record-currency majors (iss-37..44, iss-80) — maintainer added the latter rather than deferring them. Version is derived by launch ship, not declared; 0.5.0 is the prediction given itd-74 additive. Deferred majors listed explicitly in the plan (largest: iss-124); the 2026-07-24 queue's pick-up role is superseded.
- 2026-07-29 — v0.5.0 cycle runs as a scheduled cloud loop (one item per 5-hour round, plan order): auto-merge authorised through the strict gate only (CI green + two adversarial MERGE verdicts; security lens mandatory on workstream A/B diffs); itd-74 is in autonomous scope, multi-round via PR resume, bounded by spc-20 and STOP condition 1. Rejected: human-merges-everything (stalls the ordered pipeline at merge cadence) and auto-merge-C/D-only (slows exactly the security half the release is for).
- 2026-07-29 — the A2 network-identifier detector lands as one atomic change (iss-154+157+125+153): an allowlist inversion built once in the scanner's canonical pattern set, folded into `DefaultPatterns` so Stage-1 redaction and the launch/lifeboat scan inherit it, and consulted directly by the audit privacy-hygiene rule so the two surfaces cannot disagree about what a leak is. Maintainer disposition on the first repo-wide run (28 findings, plan STOP 2): the exempt set is "values that name no individual host" rather than the reserved documentation ranges alone — loopback, unspecified, netmasks, masked CIDR prefixes, and the IANA special-use ranges (IPv4 link-local, multicast, benchmarking, protocol assignments; IPv6 link-local, multicast, NAT64 well-known prefix, benchmarking). What identifies private topology stays flagged, which is the incident class: RFC 1918, CGNAT/tailnet, IPv6 unique-local, and 6to4 (it embeds a routable address, so it names a host) — consistency of the rationale wins over convenience of the residue. Twelve deliberately illustrative lines take the sanctioned per-line waiver; the repo audits clean on privacy-hygiene. The plan's STOP threshold is clarified to count findings, not distinct identifiers. Also here: `/Users/Shared` and `/Users/Guest` stop reading as usernames (iss-153), in both the audit rule and the scanner's identity matcher.
- 2026-07-29 — iss-155: `three-tier-layout` gains the placement half of the tier convention — local-tier artefacts (`NEXT.md`, `scratch/`, `logs/`) found directly in a committed tier are now errors, each finding carrying a move-to-`.abcd/.work.local/` fix. The rule verified tier presence and the `.work.local` gitignore but never that local ephemera were ABSENT from `.abcd/work/` and `.abcd/development/`, which is exactly how a handover file carrying host infrastructure detail reached a public repo unflagged. Presence is checked on the filesystem, matching the tier checks themselves: an untracked NEXT.md in a committed tier is one `git add -A` from history. The existing rule extends rather than minting a sibling ID — one convention, one rule.
- 2026-07-30 — iss-156: the PII rules domain gains the network/infra recall vocabulary the leak incident needed (`ip`, `ips`, `ipv4`, `ipv6`, `vpn`, `tailscale`, `tailnet`, `wireguard`, `firewall`, `network`, `reachability`, `reachable`, `dns`, `ssh`, `subnet`, plus the `mac address` alias) and one new rule line forbidding committed hostnames, IP/MAC addresses, and other live network identifiers: redact or omit, and use a reserved documentation value (RFC 5737/3849/2606/7042, the same citation set as the scanner and the audit privacy-hygiene rule) only where an illustrative example is needed — the rule does not tell an agent to substitute a plausible fake identifier into a factual write-up. Data-only: recall matching is already word-bounded (the prompt is normalised to space-separated tokens and a single-token term must match a whole token), so the bare `ip` keyword is safe and no matcher change was needed — a substring matcher would have forced a phrase form like `ip address`. Placement is dictated by the matcher, not by word count (`aliases` also holds single words such as `pr` and `diataxis`): a term goes in `recall` when it should hit as a standalone token, and in `aliases` when only the multi-word phrase is safe — `mac` alone would recall on an Apple Mac, so `mac address` is a phrase alias and matches via the stemmed-phrase path. The stemmer's own limits force the explicit variants: a three-character floor keeps `ips` from stemming to `ip`, and `-ability` is not bridged to `-able`, so `reachable` is its own entry. Accepted trade: broad tokens like `network` over-fire on prompts with no privacy stake, which is benign here — the injection is four short rule lines, deduped once per session. The never-commit-identifiers rule previously existed only in a parent CLAUDE.md privacy section, so the loader could never inject it; the 2026-07-29 reserved-identifier principle now has a rule-text home the hook actually emits.
- 2026-07-30 — iss-169: the visibility-driven `.gitignore` block now carries the brief's §1 table verbatim — private ignores `.abcd/.work.local/` only, public ignores the anchored `/.abcd/` plus the legacy root-level `/memory/` (the leading slash pins each public entry to the repo root; an unanchored pattern matches at any depth and would also ignore nested paths such as an `internal/memory/` source package) — replacing a phantom root-level `.work/` that appeared under both visibilities. The brief was checked first and is current: the tier table directly above §1 commits `.abcd/development/` and `.abcd/work/` and gitignores the local tier, so the code alone had drifted, and the issue's hedge that the public set "needs rethinking" resolves to a path fix, not a policy question. Under public no separate `.abcd/.work.local/` entry is needed because `.abcd/` subsumes it — visibility stays one switch with no per-subdirectory exceptions. Upgrade needs no migration verb: `gitignoreBlockDrifts` compares set-wise so an old block reads as drift, and `applyVisibilityBlock` already strips every block before writing the canonical one; the test pins that shape for both visibilities. This closes an installer-versus-auditor contradiction — `three-tier-layout` asserts the local tier is gitignored, which the installer's own output guaranteed it was not.
- 2026-07-30 — itd-74 (round 6), increment 1: the private guard layer end to end plus the maintenance verbs on both layers (AC1–AC4, AC6). Private entry format is `KEY<whitespace>PATTERN` with the key charset `[A-Za-z0-9][A-Za-z0-9._/-]*` — deliberately free of every regex metacharacter, which is what makes legacy compatibility safe rather than best-effort: a line whose first field is really the head of a regex cannot pass for a key, so it falls back to the whole-line reading under the synthetic key `entry-<line-number>`, and even where a split does apply to an old two-word pattern the resulting match is a superset of the old one (over-blocking, never under-blocking). Refusal names the key alone; the matched text and the pattern never reach any output, and a malformed line is a refusal naming its line number, because a banlist that cannot be read must not look like a banlist that found nothing. Two hook-internal fixes fell out of writing the proof: the candidate text moves from a pipe into `grep` to a temp file (under `pipefail`, a matching `grep -q` exits early and the writer left holding a closed pipe reports 141, which the pipeline surfaces instead of grep's verdict — a staged diff larger than the pipe buffer could silently defeat the guard), and grep's own stderr is discarded so an engine error message can never echo a pattern. Public layer: entries are managed IN the existing `banned_tokens` family of `.abcd/docs-lint.json` under a `names/` id prefix, and the prefix is the ownership boundary — `list` renders the whole family (hand-curated harness/present_tense entries included, marked as such), `remove` refuses anything outside the namespace. Config edits are byte surgery on the array located through the standard decoder's input offsets, never a re-marshal: an add is one inserted line plus a separating comma, a remove is one deleted line, and add-then-remove is byte-identical to the original — asserted against this repo's own docs-lint.json rather than a synthetic fixture, because a surgical editor proven only on a two-entry toy is not proven. Redaction is structural, not conventional: the exported private entry type has no pattern field at all, so no future rendering can leak the value, and pattern-validation errors discard the engine's message because Go's regexp errors quote the expression. One format, two readers, one fixture: `testdata/parse-corpus.txt` and one shared probe table drive both the Go parser and the committed shell hook, so their agreement is checked rather than assumed. Residual: the two engines are RE2 and the platform's POSIX ERE, so a pattern valid in one and not the other is possible — the shared corpus is restricted to constructs both accept, and the malformed-line path fails safe on either side. Deferred to later increments: `ahoy` scaffolding of the guard artefacts and the seeded stub (AC5), the honest-reach line on the status/report surfaces (AC7; `abcd banlist` itself states it), and the intent/spec lifecycle moves.
- 2026-07-30 — itd-74 (round 6), fix pass after two adversarial reviews (correctness + security) both returned BLOCK. Five decisions, each replacing a mechanism rather than patching an instance. (1) THE STORE DECLARES ITS FORMAT. The private banlist's first line — `# abcd-banlist: keyed` — decides the whole file: keyed means every line must parse as `KEY<space-or-tab>PATTERN`, no declaration means every line is one whole-line pattern under `entry-<line-number>`, and no line is ever split. The previous per-line heuristic ("is the first field key-shaped?") was wrong in two independent ways at once: it printed part of a legacy line as a key, and on this layer a pattern IS the secret, so the guard leaked what it existed to withhold; and it narrowed an old whole-line pattern to the remainder after its first field, so protection did NOT "never weaken because the format grew a column" — the earlier record line and the brief both claimed otherwise and were wrong. A declaration costs the user one line and makes both classes unrepresentable; `add`/`remove` refuse a non-empty legacy store rather than migrate it, because writing a keyed line in would silently reinterpret every other line. Both readers now strip ASCII space and tab only: `strings.TrimSpace` and bash `[[:space:]]` are different sets, so a U+00A0-indented line was keyed by one reader and dead to the other, and a U+000B-separated line the reverse. (2) VALIDATE AGAINST THE ENGINE THAT ENFORCES, not a convenient third one. A private pattern is screened for the constructs POSIX ERE does not implement (a backslash before an alphanumeric, `(?`) and then handed to grep ITSELF on stdin to accept or refuse; RE2 is the fallback only when grep cannot be run, and the refusal says so. Checking under RE2 accepted `\d` and `(?i)` as healthy (grep reads them as something else — inert protection reported live) and accepted `[a-z-.]`, which grep refuses (its fail-safe branch then blocks every commit). The public layer's engine is Go's regexp because `abcd docs lint` enforces it, so a public add compiles the EXACT string it stores through the linter's own compile path, and stores it with the `(?i)` prefix all sixteen hand-curated entries carry — without it a verb-written entry was case-sensitive while the docs promised otherwise. `list --private` reports unusable and inert lines APART: the first stops every commit, the second stops nothing, and one message for both misdirects an incident. (3) THE GUARD READS STAGED BLOBS, not diff text, and fails closed. Four shapes staged a banned name that no diff-text reading could see: a content line beginning `++` becomes `+++` and is dropped with the headers, a NUL-bearing blob has no textual diff at all, a committed `.gitattributes` with `-diff` disables the reading repo-wide in one line, and a rename is status R which the `ACM` filter excluded (as it excluded T). `git show :<path>` over `--name-only -z --diff-filter=ACMRT` asks the question the guard is actually asking and has no shape to route around. Every git step is checked, replacing a `|| true` that turned any failure into a clean pass: a check that could not run must never be indistinguishable from one that passed. (4) CONTAINMENT IS THE WRITE PATH'S JOB. Reads and writes resolve through `os.Root`, so a symlinked `.abcd/.work.local` cannot land the private patterns outside the repo while the verb reports the in-repo path; the verbs resolve the repo root as the rules loader does rather than trusting cwd, which had created a second nested store that the root-anchored gitignore does not match and the guard does not read; `add --private` refuses when git does not ignore the store path, because the layer's whole safety is that the file is untracked and the guard cannot catch its own source; and both stores hold the shared flock across load-modify-write. (5) A PATTERN NEVER TRAVELS IN ARGV. The hook passes it to grep via `-f -` on stdin, the verb accepts `-` to read one line from stdin (the documented form), and a flag-parse failure withholds the offending token instead of quoting it. Accepted trade on that last point: `SetInterspersed(false)` was rejected because the documented `add --private KEY "PATTERN" --json` puts a persistent flag after the positionals; a flag-error surface closes the leak without breaking it, and a test pins the trailing flag. Residual, stated rather than fixed: the guard still trusts an out-of-repo `sync-banlist` executable it neither verifies nor sandboxes, and a garbling refresh is only partly caught (the zero-entry warning); and the empty-`$toplevel` branch is guarded but not test-covered, because git resolves the repo before invoking a hook.
- 2026-07-30 — itd-74 (round 6), SECOND fix pass after two fresh adversarial reviews returned BLOCK on the first fix pass's OWN code. Detector-first throughout, each fix reverted in place and watched to fail before it passed. The guard's staged-content read was the cluster: it now reads each staged blob stage-explicitly (`git show ":0:$path"`, closing the `0:README.md` rev-magic bypass), derives staged modes from `git diff --cached --raw -z` so a gitlink (mode 160000, no blob here) is SKIPPED rather than fail-closing every commit forever, scans the staged PATH strings alongside content so a banned name in a filename is refused by key, refuses any staged path under the local tier (the private store must never be committed, and the guard cannot catch its own source), and announces the format and entry count it actually read before the scan so a stripped `# abcd-banlist: keyed` line cannot silently downgrade every keyed entry to a non-matching whole-line pattern. The verbs: `add` now proves the composed `KEY<space>PATTERN` line round-trips (a whitespace-only pattern wedged the store as `key  `; leading/trailing whitespace was silently trimmed so the enforced pattern differed from the validated one) and rejects a NUL the two readers disagree on; the stdin path reads all of stdin and refuses trailing data rather than storing the first of a multi-line pattern silently; the verb resolves the git working-tree toplevel — the exact root the guard enforces at — so a repo nested under a parent holding a .abcd/ no longer writes its store into the parent where the guard never reads it; no cobra path echoes an unknown token (a would-be private value); the inert verdict is driven by the grep probe, not a static screen that falsely called `\b`/`\w`/`\s` "matches nothing" (GNU grep implements them); `remove --private` deletes ALL lines for a key (a duplicate no longer survives under a key the report calls gone) and `remove --public` no longer lets a bare hand-curated key shadow the managed `names/<key>` target it owns; the read path reports a store git does not ignore; and a mutation's entry count excludes unparseable lines so it agrees with `list`. Records corrected: the brief and this ledger had the RE2-vs-ERE divergence backwards — Go ACCEPTS `[a-z-.]` while grep -E refuses it (exit 2), the opposite of the earlier wording; engine.go was already right. Residual, recorded not fixed: the fsutil lock files are 0644 and never unlinked (a shared primitive; the banlist lock lives in the 0700 gitignored tier and the file is empty), a public-only add still creates the gitignored local-ephemeral tier to place its lock (nothing leaks; relocating hits the atomic-rename-inode problem), the `.abcd`-symlink asymmetry and a hand-written store's 0644 mode; and the out-of-repo `sync-banlist` trust and the empty-`$toplevel` branch carry over from the first pass. `RemovePrivate` surfaces the not-ignored condition only via the shared read path (list/bare), since `PrivateResult` carries no health field.
- 2026-07-30 — itd-74 (round 7), increment 2: the scaffolding and honest-reach halves (AC5, AC7), closing spc-20's mapping. `ahoy install` writes three artefacts, all create-if-absent: the committed guard hook at `.githooks/pre-commit`, a `.abcd/docs-lint.json` carrying an EMPTY public banned-names family, and the documented private stub in the gitignored local tier. Three decisions worth the record. (1) THE SCAFFOLDED HOOK IS A GENERALISATION, NOT A COPY. It drops this repo's iss-62 identity gate and the itd-76 dogfood refresh and keeps the name guard alone, so a byte-equality drift test between the two files would be wrong; the template is proved BEHAVIOURALLY instead — the scaffold test installs it into a real temp repo and drives three commits through it (loud warning and exit 0 with no entries, refusal naming the key alone with the pattern absent from all output, refusal to stage the store itself), which is the only evidence that distinguishes a guard that works from one that merely parses. Rejected: a symlink from `.githooks/pre-commit` to the embedded default (zero duplication, but breaks a Windows checkout and would have rewired increment 1's own test harness), and copying this repo's hook verbatim (it would scaffold repo-specific gates into every managed repo). (2) THE PUBLIC FAMILY IS SEEDED EMPTY. abcd cannot know which names a repo may not publish, and a ban nobody declared would fail a build over a word the maintainer never chose; the array's PRESENCE is what makes `banlist add --public` usable, which is all AC5 asks for. A docs-lint config that exists but carries no usable array is a NON-resolvable diagnostic gap: the config gates CI and a contributor owns it, so abcd reports the fault and never rewrites a file it cannot read. (3) THE STUB'S GAP IS RESOLVABLE ONLY ONCE THE FENCE COVERS IT. Writing the store into a repo git would track is the exact hazard the layer exists to prevent, so the stub write runs after the visibility step and is gated on the canonical `.gitignore` block being on disk; advertising it as resolvable regardless would leave a repo permanently "partial" on a fix apply refuses to make (the markerSymlink precedent). No new gitignore entry was needed — iss-169's fence already covers `.abcd/.work.local/` under private and `/.abcd/` under public — so AC5's gitignore clause is pinned by a new test against real `git check-ignore`, keyed to `banlist.PrivateRelPath` rather than to a spelling of the tier, and it passed on first run. AC7: the reach sentence lives once as `banlist.PrivateReachNote` and travels INSIDE the reported state (`BanlistHealth.Reach`), not as renderer prose — a machine consumer reading "hook installed" beside a present store would otherwise draw exactly the wrong conclusion, and the JSON envelope is a report surface too. The health pass reads no entry and spawns no subprocess: the store's content is the secret, and a status board is the surface that must not hold it. Stub examples are all commented out (a fresh scaffold parses to zero entries, so the guard warns loudly rather than looking like protection) and every illustrative value is a reserved documentation value or a persona-derived fixture host, judged by the repo's own network-identifier detector — with a control value asserted flagged first, so the assertion cannot pass vacuously. Residual, recorded not fixed: abcd does not set `core.hooksPath`, so a clone arms the hook by hand (the record scopes AC5 to the hook being committed, and silently repointing a repo's hooks path is a git-config change no acceptance criterion asked for); and the scaffolded template and this repo's prototype now share ~250 lines that no mechanism holds in sync. FIX PASS after two fresh adversarial reviews returned BLOCK on this increment's own code; detector-first throughout, each behaviour change watched failing first. (A) THE FENCE IS GIT'S DECISION, NOT A TEXT COMPARISON. Both the stub write and the gap's resolvability were derived from `gitignoreBlockDrifts` — a set comparison of the abcd-managed .gitignore block — which answers a different question from the one that matters. A repo can carry a byte-perfect block and still track the store (a negation after it, a tracked tier), a fully-configured repo never reaches the visibility step at all, and `stepConfigValues` returns nil whenever ANY unrelated config value is missing under a declined ConfigChange, which silently disabled the write while `localTierIgnored` reported the gap as resolvable — a permanent partial. Both now call `gitutil.IsIgnored`, the same primitive increment 1's `requireIgnoredStore` uses, evaluated on disk after the visibility step, and skipped outside a git repo for the same stated reason (a check cannot demand proof no one can supply). Health carries the verdict and a present store git would track is called out on every surface. (B) CONTAINMENT. The three writes used `os.MkdirAll` plus an atomic rename, and a repo that commits a symlink at `.githooks` — which a checkout materialises before abcd runs — took the 0755 hooks OUTSIDE the repo while every surface reported the in-repo path; watched, and confirmed by the reverted code writing `pre-commit` into the link target. They now go through `fsutil.CreateExclusiveIn` under an `os.Root` opened at the repo, which is both guarantees at once (no clobber, no check-then-write window) and refuses symlink traversal at every level. The local-tier case was already contained by increment 1's `os.OpenRoot` read path, which reports such a store as unreadable rather than absent. (C) BOM FAIL-OPEN, IN BOTH READERS. A leading UTF-8 BOM made the first line differ from the format declaration, so a KEYED store read as LEGACY: every keyed entry became a whole-line pattern matching nothing a commit contains, while `entries` stayed >=1 so the zero-entry warning never fired — a store that looked healthy and checked nothing, reproduced as a real commit going through. Fixed in the Go parser, the scaffolded template, and this repo's own prototype; and a first line that is not the declaration but WOULD be after stripping leading blanks is now a DAMAGED declaration that fails closed, because reading it as legacy is the same downgrade by another route. (D) MERGE-COMMIT BYPASS. git runs no pre-commit hook for a merge, so a banned name entered history the moment a branch carrying it was merged; a `pre-merge-commit` half now runs the same guard through one implementation (a shim, not a copy) in both the template and this repo. The reach sentence widened with it: `PrivateReachNote` now names what a hook cannot see — a rebase, `git am`, a cherry-pick, `--no-verify` — because "machines that have opted in" is necessary and not sufficient, and the shorter sentence let a reader believe an opted-in machine was covered. (E) TWO CLAIMS WITHDRAWN. A foreign pre-commit hook was reported as the abcd guard: presence is not identity, so each template carries `# abcd-name-guard: v1`, a hook without it is a non-resolvable diagnostic abcd never replaces, and "installed" became "committed" with the arming instruction beside it (git runs the hook the clone's hooks path selects, which abcd neither sets nor fully observes). And under `visibility: public` the fence ignores the anchored `/.abcd/`, so the "committed, CI-enforced" public family is untracked exactly where public exposure is the risk — detection now reports `banlist.public_family_ignored` and the board reads NOT ENFORCEABLE. The placement question is NOT resolved here (moving the file amends the iss-169 record; carving an exception gives up its one-switch property) and is captured as iss-176 for the maintainer, with three candidate reconciliations. Minors in the same pass: the private layer's shape comes from a new `banlist.SummarisePrivate` — the shared parser without the per-entry grep, so a status pass costs no subprocess per line — the public family's four faults report apart (an unreadable file is not an absent array), the local tier's 0700 is asserted rather than assumed from the creating call, the staged-path refusal is case-folded, the format is announced before the parse loop, and stale scratch files are swept. Residuals the reviews verified as accepted: the scaffolded template and this repo's prototype share ~250 lines that no mechanism holds in sync (byte-equality would be wrong — the template drops the identity gate and the dogfood refresh — and the template is proved behaviourally instead); `--no-verify`, rebase and server-side pushes remain uncoverable by construction; GNU/BSD grep divergence stands as increment 1 recorded it; commit messages, tag names and branch names are outside the guard's scope (it reads staged blobs and paths); a scaffolded repo gets no CI wiring for the public family; abcd still does not set `core.hooksPath`; and `a.note()` renders absolute paths across every apply step, pre-existing and repo-wide, captured as iss-177 rather than half-fixed here. FIX PASS 2 after a second pair of independent BLOCK reviews, both on the fix pass's own code. (F) THE MERGE SHIM NEVER STANDS IN. It was written whenever its own gap was present — including beside a FOREIGN pre-commit hook — so abcd's marker landed in the shim, the board read "pre-merge-commit hook committed", and merges stayed unchecked; worse, the maintainer's hook silently began running on merge commits, which git never did, so apply was taking over exactly the wiring the foreign-hook gap disclaims owning. The shim is now written only beside abcd's own guard (including one the same step just wrote — keyed on the artefact's STATE, not on a gap a fresh repo deliberately does not raise), and `banlist.merge_hook_inert` reports the alternative. The wrong assertion in the increment's own test — which asserted the buggy behaviour — was flipped and watched fail. (G) A COPY OF THE STORE IS A MISTAKE-NET, NOT AN ADVERSARY-NET. The staged-path refusal matched only the local tier, so `cp` of the banlist to notes.txt, or a `git mv` out of the tier, committed every pattern in clear while the guard announced a clean check — the entries cannot catch their own text, because they are escaped regular expressions and `carol-server\.example\.net` does not match itself. Three tests now: the tier path (not escapable, and applied to a rename's SOURCE path too — the loop had been overwriting it with the destination and discarding the one field that identified the file), a staged blob whose FIRST LINE is the format declaration, and a staged path whose basename is the store's filename. Their reach is stated rather than implied, here and in the brief and the spec: a copy with the declaration stripped, altered by a byte or displaced below a preamble passes the first-line test; a LEGACY store declares no format at all and passes it always, and passes the basename test the moment it is renamed; and the rename-source test only fires for a store git already tracks. What they catch is the accident that happens — a `cp`, a `.bak`, a stray duplicate — and they are NOT widened further on purpose: matching a store's key list or scanning deeper into every staged blob would route the secret through more code paths to catch fewer accidents. Round 3 also found the other direction: the first-line test blocked this repo's OWN committed corpus fixtures, and `--no-verify` is an off switch for the whole guard rather than a per-file escape, so the refusals now honour a published second-line `# abcd-banlist-example` marker — exempting a blob from the COPY tests and from nothing else, its content still scanned against every entry — and the two keyed corpora carry it. The tests also moved OUT from behind the absent-store early exit (whether this machine has a store says nothing about whether a commit is carrying one, and that exit was itself how the rename escaped), and the guard no longer creates its local tier in a repo that never opted in: with no store it scratches in a 0700 temp directory instead of leaving an unfenced abcd directory behind on every commit. (H) THE DECLARATION IS LINE 1 OR NOWHERE, in all three readers and for BOTH formats. A blank line or a comment above it, a duplicate below it, or any prefix bytes before it (a UTF-16 byte-order mark, an editor artefact — caught by a suffix test rather than a byte-class check no portable shell has), is a damaged declaration that fails closed. Round 3 found the header scan gated on the store not already being keyed, so a keyed store with the declaration repeated on line 2 read as healthy to Go — "present, keyed, 1 entry" on the status board — while the shell blocked every commit; a store one reader calls healthy and the other refuses is worse than either verdict alone, and a fourth shared corpus now drives both readers over exactly that file. (I) THE ENVIRONMENT IS PINNED FIRST, in all four hooks: xtrace off, IFS, LC_ALL and PATH as the FIRST statements, before anything is read or run — the pin had been landing after `rev-parse`/`cd`, and the merge shim had none at all while resolving its delegate through an external `dirname` (now `${0%/*}`, pure parameter expansion). The pin NARROWS the substitution class and does not close it: two of the pinned directories are user-writable on a typical developer machine, and `#!/usr/bin/env bash` resolves the interpreter through the inherited PATH before any of it runs — the earlier "root-owned" wording in this line was wrong and is corrected here. Every external tool the hooks call is probed after the pin so a missing one blocks loudly instead of surfacing as a mute exit 127, and a nonstandard prefix extends the pin through `git config --local abcd.guardPath` — read through the already-pinned git, and deliberately NOT an environment variable, since a repo-scoped direnv sets the environment and would reopen the exact hole. Minors in the same pass: scaffolded modes are chmod'd explicitly because `CreateExclusiveIn` and `MkdirAll` are umask-masked and a stripped exec bit turns the shim's fail-closed branch into a permanent merge block; the marker is matched by PREFIX so a v2 template cannot reclassify every v1 hook as foreign; `storePathIsSafe` distinguishes "not a repository" (skip, as before) from "repo-shaped and git will not answer" (fail closed) — `InRepo` is false for git-absent-from-PATH too, which the old comment's justification did not cover; `hooksPathArmed` resolves both sides before comparing, so an absolute `core.hooksPath` is no longer read as unarmed and told to downgrade itself; the absent-public-family gap is non-resolvable where git would ignore the path and abcd declines to write a config it would immediately call unenforceable; the detection envelope omits the banlist object for an unmanaged folder rather than serialising undeclared zero states; detection reads now resolve through the same containment root the writes use; the reach note is explicitly non-exhaustive and names `git revert` and a reapplied stash; and `abcd ahoy` scaffolds the `.githooks/* text eol=lf` attribute into managed repos (append-if-missing, never a rewrite) rather than leaving them the CRLF-shebang failure this branch fixed for abcd itself. CORRECTION to this line's earlier claim that the health pass "spawns no subprocess": it always did — it now spawns four (one `rev-parse`, one batched `check-ignore` for both paths, one `config --local`, and one `check-attr` for the hooks' EOL), down from six or seven, and `ahoy.Detect` is on the session-hook path, so that count is a cost paid per prompt. One earlier test was vacuous and is fixed rather than deleted: the local-tier half of the symlink containment test passed with containment reverted, because the escaping store is classified unreadable before any write is attempted, so it now asserts that protecting state by name. FIX PASS 3 (converge). MARKER IDENTITY: `classifyGuardHook` matched the marker as a substring of the whole blob, so a foreign hook that merely MENTIONED it — a comment, a grep for it — classified as abcd's own, the board claimed coverage, and the merge shim was written beside it; identity is now a whole LINE matching `^# abcd-name-guard: v<digits>$`, with the version a group so a v2 template cannot reclassify v1 hooks. The same line-wise fix went to the `.gitattributes` pin, where a commented-out attribute had read as pinned. REACH ACCURACY: a fast-forward `git pull` creates no commit and so runs no hook — the commonest uncovered path of all, now named — and the stash claim was probed FALSE (`git stash pop` then commit runs pre-commit normally) and removed: a bypass list naming a covered path teaches a reader to distrust the rest of it. Minors: the merge-inert gap no longer asserts a delegating shim that may not exist; `createContained` pinned modes on directories it did NOT create, widening a maintainer's deliberate 0700 to 0755, and now pins only what it made; `repoShaped` walked no ancestors, so in a repo SUBDIRECTORY with git unavailable the stub would have been written into a tracked tree; the stale-scratch sweep and the tier creation are conditional on the tier being the scratch home; and the foreign-hook gap's `Required: false` is documented as deliberate (a required gap nothing can resolve is a repo permanently reported incomplete for a state its maintainer chose). Also re-classified: `stepBanlist` now asks disk what occupies the hook paths rather than trusting the detection snapshot taken before the earlier steps ran. FIX PASS 4 (prose currency plus one-liners). The two CHANGELOG bullets, internal/README.md, and both brief surfaces described THREE scaffolded artefacts and a stale reach; there are FIVE (both guard hooks, the .gitattributes LF pin, the public family, the stub), and the reach now names a fast-forward `git pull` — the commonest uncovered path of all — and no longer names a reapplied stash, which probing showed IS covered and which reach_test.go now pins out. Security minors: an inherited shell FUNCTION shadowing `grep` defeated the PATH pin entirely (probed — the guard reported a clean check under `BASH_FUNC_grep%%=() { return 1; }`), so all four hooks now `unset -f` the commands they run as their first statement after the xtrace guard; `abcd.guardPath` refuses a value with an empty PATH element, which every consumer reads as the current directory; the scratch-directory trap is installed before the three inner mktemps that could exit past it; and `mktemp -d` takes an explicit portable template. The `.gitattributes` classification stops pattern-matching abcd's own line and asks `git check-attr eol` instead — git is already this package's authority for the ignore question, and a later `* text eol=crlf` overrides a line a regex still calls pinned; the line-anchored regex survives only to keep an append from duplicating. The marker regex tolerates a trailing CR, so a CRLF-mangled abcd hook classifies as OURS-but-unpinned and install can heal it rather than reporting a foreign hook for ever. Three residuals are widened rather than fixed, in the brief and the spec: the environment pin narrows common accidental and repo-scoped routes and closes nothing (anyone controlling the committer's environment wins); the example escape keeps content scanning so it cannot smuggle a plaintext banned name, but it protects no copy of a STORE — escaped patterns never match their own text — and a live store carrying the marker exempts every copy of itself; and the marker asserts identity, never integrity, so any file quoting it is treated as abcd's guard. Lifecycle: itd-74 sits in shipped/ with its fidelity review OWED (receipt rcp-3ceed52bdb99). That is the designed follow-up, not an omission — `abcd spec close` stamps the receipt and the review is a separate verb (`abcd intent review`), which record-lint and audit accept; no record file was hand-edited to produce or to clear it.
- 2026-08-01 — iss-96 (item A6), a VERIFICATION milestone rather than an implementation: the transcript scanner's tracked coverage gaps re-checked against the network-identifier set A2 folded into `DefaultPatterns`, and the residue re-scoped in place. Two findings, both empirical rather than inspected. (1) A2 DID reach this path — every scanner consumer inherits the network set, so a non-reserved address or a LAN/device hostname in a captured transcript is now a finding (all five kinds — `net:ipv4`, `net:ipv6`, `net:mac`, `net:lan_hostname`, `net:device_hostname` — exercised on the transcript path itself in the same corpus, not inferred from their presence in the pattern set, and a flagged private address asserted REDACTED in the stored record at the store boundary); that class therefore leaves iss-96's residue. (2) The classes iss-96 actually names — a 40-character secret-key value with no prefix, a bare password, a prefix-less API token, and now a genuinely high-entropy 40-character value — still produce ZERO findings at the transcript path's own entry point, with or without a `password:`, `api_key =` or `Authorization: Bearer` key name beside them, because the TOKEN patterns are prefix-anchored APART FROM `rp_session_key`, which keys on the literal JSON field name `"sessionKey"` rather than on a generic key-name class, and no pattern measures entropy. The set as a WHOLE is not uniformly prefix-anchored and must not be described as such: the network kinds are an allowlist inversion, the identity kinds key on the probed identity, `token:pem_private_key` is a literal header, and `Pattern.SkipAt` already lets a pattern accept or reject a match by what SURROUNDS it — none of which reaches an unlabelled value, which is why the residue stands, and two of which (`rp_session_key`, `SkipAt`) are the shipped precedent remediation option (b) would generalise rather than introduce. Disposition: iss-96 stays OPEN with a dated verification section appended in place (the iss-24 precedent; a `capture resolve` would have had to state an action nobody took, and the ledger's resolution field is exactly that claim), re-scoped from "coverage may have moved" to a single unresolved DECISION between an entropy/charset detector with a length floor, key-name context matching, and the opt-in external-scanner adapter over this path only — three options that differ in REACH, not merely in false-positive cost, and saying otherwise would misdescribe them: (a) reads the value, so it is the only one that reaches an UNLABELLED value; (b) covers labelled values only and is structurally incapable of this entry's own first-named case, a bare secret-key value with no key name, because there is no key to match; (c) has whatever reach its ruleset delivers, measured today as gitleaks 8.24.3's default rules finding ONE of the fixture's specimens (the keyword-adjacent passphrase) and passing the secret-key shape, the hex token, the base64 token and the high-entropy value. False-positive cost is the SECOND axis and lands hardest on (a) — a redaction false positive corrupts the record redaction exists to preserve — so reach and cost together are the open question and the bar is the maintainer's to set: grill-then-implement, not autonomous. The verification is PINNED in the tree rather than asserted in prose: `TestTranscriptPathMissesUnanchoredEntropy` at the pattern-set boundary and `TestCaptureStoresUnanchoredEntropyVerbatim` at the store boundary (each specimen present verbatim in the written record, the anchored token absent AND its `maskSecret` fingerprint present, since absence alone would also be satisfied by a store that dropped the line). THE PINS' REACH IS THE PATTERN SET, and the earlier claim that they "fail by design, never a silent pass" was too broad on two counts, both now fixed: the original specimens all REPEAT and measure 3.12, 2.16 and 2.75 bits per character, below the ~3.5-bit floor an entropy detector conventionally uses, so an entropy detector could have landed with every pin green — a 5.32-bit specimen assembled at run time from a fixed-seed shuffle, with its entropy asserted in-test at or above 4.5 bits per character, now closes that hole; and option (c) never consults `DefaultPatterns` at all, so NO ScanText-level pin can alarm for it by construction, which is stated in the entry so a stale close cannot lean on the pins alone. Within the pattern set they do fail by design — a charset/length floor, key-name matching or any new `DefaultPatterns` entry trips them — and that failure is the signal to re-point them and close the entry. Never close on assumption: the anchored controls in the same corpus are caught, and one of them sits INSIDE the negative test, which is what proves the specimens reached a working detector rather than a mis-wired or emptied one; the negative assertion is scoped to the token/secret kind family, so an unrelated widened pattern fails it under a DIFFERENT message rather than being credited as entropy coverage. Every credential-shaped specimen is assembled at run time — the discipline network_test.go already states for flagged network identifiers — so no literal of that shape enters this repo's history for a full-history secret scan to fire on.
