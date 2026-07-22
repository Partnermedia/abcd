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
