# Assessment: `.abcd/development/brief/` structure vs state of the art

Specialist pass: documentation-architecture evaluation — is the brief structure state-of-the-art, and is it generically reusable? **Corpus reviewed:** all of `.abcd/development/brief/` (~5,500 lines across 60 files: 00-meta, README, 01-product ×5, 02-constraints ×4, 03-evidence ×4, 04-surfaces ×9+README, 05-internals ×10+README, 06-delivery ×3, glossary README + _template + 15 terms), plus `adr-5-brief-is-current-state.md` and `adr-7-grill-skill-and-glossary.md`.

## Scorecard

| Dimension | Verdict | Gap vs SOTA |
|---|---|---|
| Working-backwards product framing (PR) | **Strong** | Press release is genuine Amazon-style (present tense, customer quote, why-it-matters) and improves on vanilla PR by embedding Given-When-Then acceptance criteria. |
| PR/**FAQ** | **Half missing** | No FAQ artifact at all. The internal-FAQ content (hard questions, risks, "what will disappoint users") is scattered across 03-evidence/tradeoffs, Open Questions blocks, and ADRs. In Amazon practice the FAQ carries most of the value. |
| Classic PRD coverage | **Good minus one hole** | Problem ✓, users ~ (personas are quote-generators, not researched segments/JTBD), requirements ✓, non-goals ✓✓. **Success metrics: entirely absent** — no adoption/quality/fidelity targets anywhere; the "round-trip fidelity floor" open question admits it. |
| arc42 coverage | **~9 of 12 sections** | Goals ✓, constraints ✓ (rare, explicit chapter), context ✓, cross-cutting concepts ✓✓ (04-universal-patterns is a model arc42 §8), decisions ✓ (external ADRs), glossary ✓✓. Missing: consolidated **quality requirements/NFRs** (only 08-abcd has an NFR section), **risks & technical debt register** (the issue ledger is a runtime surface; the brief has no risk chapter — and the live issues log is *gitignored*), deployment/ops view. |
| Security / threat model | **Substantive but unconsolidated** | Injection canaries, fail-closed gates, PII/secret scans, path containment, replay defenses are all specified — scattered across 6+ files. No trust-boundary map, no single chapter a security reviewer (or agent) can load. |
| Diátaxis coherence | **Coherent externally, blurred internally** | User docs correctly Diátaxis. Inside the brief, reference (contracts, tables), explanation (rationale), and **status/changelog** are interleaved in the same files — the status material is the offender (see below). |
| Spec-driven-dev alignment | **Ahead in parts** | Constraints+invariants ≈ spec-kit constitution; surfaces ≈ specs; delivery ≈ plan. Verification matrix and derivable out-of-scope exceed spec-kit/OpenSpec/Kiro. Brief-skeleton itself is deliberately *not* machine-enforced (documented, defensible). |
| Agent consumability: chunking | **Mostly right, four outliers** | Most files 50–250 lines (1–4k tokens) — ideal. Outliers: `04-surfaces/05-intent.md` (458), `05-internals/03-configuration.md` (449), `04-surfaces/01-ahoy.md` (391), `05-internals/10-in-session-dispatch.md` (308). `02-constraints/04-naming.md` grows **monotonically by design** (VR001 forces every new term into it) — a scaling time bomb. |
| Agent consumability: ordering & anchors | **Good ordering, brittle anchors** | Two-digit numeric prefixes give deterministic sort. But numbered *section headings* + slug anchors have already drifted: `04-surfaces/README.md` and `05-internals/01-agents.md` both link `05-intent.md § 6 …` while the actual heading is now `## 7.` — a live broken deep-link. Slot reuse also observed (internals slot 7 retired for itd-32, recycled for memory). |
| Normative language (RFC 2119) | **Present but informal** | MUST/NEVER/ALWAYS used heavily and mostly correctly, but no conformance statement, mixed with rhetorical ALL-CAPS ("ZERO writes", "DISPLAY-ONLY"), and acceptance bullets are unnumbered (no AC-IDs to cite). Lint codes (`SD001`, `GL002`…) function as de-facto enforceable requirement IDs — genuinely good — but only for lintable rules. EARS is cited in ADR-7 yet not adopted for the brief's own criteria. |
| Self-consistency | **Visible drift despite lint** | Glossary README omits `distribution/` context and `core/disembark.md` from its layout/index, and its validation command points at a different path from where the files sit. `03-evidence/03-open-questions.md` references `research/adr/` while the canonical store is `decisions/adrs/`. Two "PLACEHOLDER"-bannered files (`invariants`, `dependencies`) actually contain canonical, load-bearing content. |
| Genericity / template reusability | **Skeleton yes, content no** | Chapter taxonomy is universal; significant abcd leakage in chapter *content* and one structural coupling (brief↔lifeboat contract in 00-meta). Detailed in Q2. |

---

## Q1 — Is the chapter taxonomy sound? What's missing / redundant?

**Sound.** The spine — *meta → why (product) → hard rails (constraints) → learnings (evidence) → what (surfaces) → how (internals) → when/proof (delivery) → vocabulary (glossary)* — is a defensible superset of PRD + arc42 + spec-kit. Putting constraints *before* the what/how is better than most PRDs (it matches spec-kit's constitution-first ordering), and a first-class delivery chapter with a verification matrix is rare.

**Missing vs SOTA:**

1. **Success metrics / quality goals** — the biggest hole. Nothing answers "how do we know abcd is working?" No fidelity floor, no adoption signal, no quality tree (arc42 §10, PRD success-metrics). `09-reflect.md` even codifies "no DORA/velocity telemetry" for retrospectives, so nothing backfills it.
2. **Risk register** — risks exist (Pass B signal density, vendor JSON schema instability, itd-36 "tight coupling logged as a principal risk") but are buried in prose across four chapters. arc42 §11 / Amazon internal FAQ both make this a named home. Worse, the live issues log is gitignored — discovered risks are structurally invisible to the committed brief.
3. **Security/threat model summary** — the material exists and is above-average; it needs a consolidating file (trust boundaries, untrusted-input inventory, gate map). The `reads_untrusted_input` agent frontmatter is already half the inventory.
4. **Operational/migration slot** — migration is explicitly deferred (itd-9), which is fine for abcd, but a *template* frozen for all future projects needs the slot (upgrade story, compat policy, support/telemetry), even if a project fills it with "n/a because X".
5. **FAQ** — see scorecard. The press release's "Open Questions" section is the embryo; the internal-FAQ discipline (forcing answers to the hard questions before build) is absent.

**Redundant / misplaced:**

- **03-evidence: 3 of 4 files are empty scaffolding** with meta-commentary about their own future population. A brief that ships empty chapters trains agents to expect emptiness (see Q3).
- **`01-product/05-personas.md` is 7 lines** — fold into press-release or context.
- **`02-constraints/04-naming.md` conflates three concerns**: the maritime metaphor convention (constraint), the bare-command-as-render surface discipline (constraint), and an unbounded vocabulary register (a registry, not a constraint). The register belongs in the glossary — which already exists as per-term files with a schema and a linter. Right now there are effectively **three vocabulary stores** (naming.md tables, `brief/glossary/`, and the `terminology/` path the linter/GL codes target), with `VR001` enforcing registration into naming.md and `GL00x` enforcing the term files. Two enforced registries with different scopes is drift-by-construction. *(Resolved by the review-time decision: `glossary/` is the SST — see `00-summary.md` § 5.)*
- **"PLACEHOLDER" banners on canonical content** (`02-constraints/02-dependencies.md`, `03-invariants.md`). Invariants lists nine numbered non-negotiables — that is not a placeholder. A canonical file must not disclaim its own authority; an agent deciding what to trust gets a contradictory signal.

## Q2 — Is it generic?

**The skeleton is generic; the flesh leaks abcd everywhere.** Test against a web app / library / data pipeline:

**Genuinely universal:** meta, product (PR/context/mental-model/scope), constraints (platform/deps/invariants), evidence, delivery (build-sequence/verification-matrix/out-of-scope), glossary-with-bounded-contexts. All map cleanly: a library's "surfaces" are its public API modules; a pipeline's are its jobs/contracts/CLIs; a web app's are pages/flows/endpoints.

**abcd-specific leakage:**

- **"Surfaces" as written assumes a command-per-surface CLI.** The chapter README, the bare-command-as-render discipline, sub-verb earning rules, and `SD001` are CLI-plugin concepts. The *concept* (one file per externally observable interaction surface, each with purpose/flow/acceptance) generalizes; the current text does not. The template needs an abstract definition plus per-project-type examples.
- **The brief↔lifeboat shape contract in `00-meta.md`** couples the template to an abcd product feature ("the same skeleton is a populated lifeboat"). Elegant for abcd, but a generic template cannot *require* consumers to have a disembark/embark pipeline. This is the one structural (not just textual) leakage.
- **03-evidence's population mechanism** ("populated post-build by the disembark process", sources like `.specstory/`, oracle reviews) is abcd machinery. The chapter concept is universal; the population contract isn't.
- **05-internals file names** (`01-agents`, `02-adapters`, `05-prompt-quality`, `10-in-session-dispatch`) are this product's subsystems. The template rule should be "one file per subsystem/cross-cutting concern, numbered, with a README index" — not these names. (`04-universal-patterns.md` as a *slot* — "cross-cutting patterns" — is universal and should be kept by name.)
- **Nautical vocabulary** is confined to content, not chapter names — good. But naming.md's metaphor tables, `personas.json` tooling, `itd-N`/`fn-N` ID grammar, and the phase-vs-plumbing-phase dual numbering in `06-delivery/01-build-sequence.md` (which the file itself admits is confusing enough to need a warning) are all project-specific.

**A project-agnostic template** would be: `00-meta` (conventions only, no lifeboat), `README`, `01-product/` (press-release+FAQ, context, mental-model, scope, users), `02-constraints/` (platform, dependencies, invariants, conventions), `03-metrics-and-risks/` (success metrics, risks, security posture) ← new, `04-surfaces/` (abstractly defined), `05-internals/`, `06-delivery/` (build-sequence, verification-matrix, out-of-scope), `07-evidence/` (advisory, clearly-marked population contract), `glossary/`. Everything abcd-specific becomes example content, not skeleton.

## Q3 — Is "evidence" inside a current-state brief coherent with adr-5?

**Conceptually yes — and it's one of the better ideas here.** adr-5 bans *history* (version labels, archives, changelogs) because git holds it. Evidence is not history: "what worked / what didn't / open questions / tradeoffs" are **present-tense claims about currently-operative knowledge** — exactly Naur's "theory" that the whole product exists to preserve. `04-universal-patterns.md § 8` even classifies the brief as *compounding-curated*, which licenses exactly this: entries carry provenance, get contradicted, get deprecated. The files' own framing ("evidence, not prescription"; "Reconsider when:") is present-state-compatible and better than most retro docs, which rot into archaeology.

**Three real problems, none fatal:**

1. **Three of four files are empty placeholders** waiting on a Phase-4 pipeline that hasn't shipped. Until disembark exists, the chapter is dead weight in the must-read path and normalizes empty canonical files. Either populate manually now (tradeoffs.md proves it can be done by hand — it's the best file in the chapter) or move the chapter out of the numbered spine until it has content.
2. **The evidence/ADR boundary is informal.** tradeoffs.md calls itself "the lightweight rolodex; ADRs are the deep dive" — workable, but the same decision can now live in tradeoffs, an ADR, *and* an intent's reclassification history. Only prose partitions them; Role-2 consistency review is the only guard and its mechanical half is deferred.
3. **Stale references already**: `03-open-questions.md` and `04-tradeoffs.md` point at `.abcd/development/research/adr/` while the canonical store is `decisions/adrs/` — the evidence chapter is itself drifting, which undercuts its claim to be curated.

Verdict: keep it, mark it explicitly *advisory* (it already says so — elevate that to frontmatter agents can read), define a staleness/pruning rule so compounding-curated is real, and fix the population story before freezing the template.

## Q4 — Granularity and numbering

**Granularity:** the modal file (~50–250 lines, 1–4k tokens) is well-sized for selective agent loading, and 00-meta's four-reason rationale for the split (concurrency, diff legibility, context budget, reusable shape) is exactly right and rarely articulated anywhere. Problems:

- **Four outliers** exceed ~400 lines (05-intent, 03-configuration, 01-ahoy) or ~300 (10-in-session-dispatch). 05-intent.md is really four documents (lifecycle, sub-verb reference, PRD/freeze protocol, reviewer roles+audit loop).
- **The verification matrix is one monolithic table** with paragraph-length cells (some rows are 100+ words). Agents can't load "just the launch rows" without the whole table; key it per-surface or split by chapter.
- **`04-naming.md`'s vocabulary register grows every spec by lint mandate** — the one file guaranteed to blow any per-file budget. It's already accreting rows that are effectively spec changelog entries (seven rows just for fn-14's rules.json domains).
- **Deep implementation detail is climbing into the brief.** `06-lint.md` carries per-code delivery status, task IDs, line-precision implementation notes, and adjudications of issue-ledger notes. That's a requirements registry fused with a delivery ledger — the fusion, not the size, is the problem (see recommendation 1).

**Numbered prefixes:** on **files**, keep them — two-digit padding, deterministic ordering, cheap for agents; insertion cost is real but low at this fan-out, and the observed failure mode (recycling retired slot 07 for a new topic) is a discipline problem, not a scheme problem (rule: retire numbers, never reuse — the lint-code registry already states exactly this rule for codes; apply it to file slots). On **section headings** (`## 7. Acceptance`), drop the numbers: they generate slug anchors that break on renumbering, and the corpus already has a live broken anchor (`§ 6` → `## 7.` for the intent-fidelity-reviewer section, referenced from at least two files). Numbered headings are also inconsistently applied across surface files. Use stable kebab anchors; add a link-check lint (the deferred `XD006` reference-rot code is precisely this — it should be the *first* XD code shipped, not deferred).

## Q5 — Where it exceeds common practice

This brief is ahead of SOTA in at least eight places:

1. **Lint-enforced, bounded-context glossary as files** — DDD ubiquitous language with schemas, forbidden synonyms, lifecycle fields, a template, and blocker-severity enforcement (`GL002`, cite-or-fail per ADR-7). No mainstream template (arc42, spec-kit, Kiro) does this; most projects have a stale wiki page at best.
2. **Acceptance criteria at every altitude, one format** — GWT on the press release, per surface, per phase, per discipline, and on the lint contract *itself* ("the lint contract is itself acceptance-checked"). Uniform-format/different-home is a genuinely good doctrine.
3. **Verification matrix as a first-class brief chapter** — a testable claims inventory ahead of implementation. Spec-kit has nothing equivalent.
4. **Canonical, derivable out-of-scope** — negative scope as press-release intents plus a filesystem-derived enumeration command to prevent hand-count drift. The embedded hand-maintained exclusion list partially reintroduces the drift it fights, but the instinct (derive, don't count) is ahead of practice.
5. **Artefact-lifecycle taxonomy** (regenerable / append-only / compounding-curated) with the full-crawl-vs-delta recomputation argument — no documentation framework names this, and it resolves real ambiguities (it's what licenses the evidence chapter).
6. **The brief self-describes and self-governs** — 00-meta + adr-5's "current-state, no archive, git is history" doctrine, with the ADR recording the failure mode it replaced (four archive dirs, a v5 label, changelog blobs). Most orgs never confront this.
7. **Lint codes as enforceable requirement IDs** — `SD001`/`VR001`/`GL00x` turn design disciplines into pre-commit blockers. This is the "machine-checkable structure" north star of spec-driven tooling, actually designed in.
8. **Epistemics**: the Naur recovery-humility framing, "investigation-gated" notes recording *verified absence* (08-abcd's "no stub ever existed, don't hunt for it"), and honest detectability tables ("rows without an artifact are `no`, not pretended-yes"). Unusually truthful documentation.

## Q6 — Ranked recommendations before freezing as the template

1. **Separate contract from delivery status (highest impact).** Strip the "Delivered in fn-N task .X", "Status: design target — Phase 4", shipped/deferred narration out of chapter bodies into either (a) a machine-readable status sidecar (`status.json` per chapter keyed by stable IDs) or (b) the spec/task layer, where delivery state already lives. This is the single largest source of file bloat, the main Diátaxis-mixing offense, and — by the brief's own doctrine — status *is* git/spec-inferable state that will go stale. `06-lint.md` and the surface-file status banners are the worst cases. Do this before templating, or every future project inherits the fused shape.
2. **Unify the vocabulary registries.** One store: the bounded-context glossary files. Fold `04-naming.md`'s "Reserved vocabulary" table into per-term files (they already have the richer schema); retarget `VR001` at the glossary; keep naming.md for *conventions only* (metaphor criterion, bare-as-render). Fix the glossary README drift now (missing `distribution/` context, missing `core/disembark.md`, wrong validation path) — it's the flagship chapter and its own index is stale. *(Decided at review time: glossary/ is the SST.)*
3. **De-abcd the skeleton.** Rewrite 00-meta without the lifeboat contract (make brief↔lifeboat an abcd-project note, not template semantics); define "surface" abstractly with per-project-type examples; move CLI disciplines (bare-as-render, SD001) from template text into abcd's constraint content; make 05-internals a "one file per subsystem" rule rather than a fixed agent/adapter roster.
4. **Add the missing chapters as template slots**: success metrics/quality goals (mandatory), risks & security posture (mandatory — and committed, not gitignored), operational/migration (may be "n/a because…"). Add an internal FAQ section to 01-product.
5. **Kill the PLACEHOLDER banners.** `03-invariants.md` is canonical — say so. Either consolidate the scattered invariants/dependency rules into these files (their stated future intent) or delete the banner and keep them as honest pointers. A file cannot be simultaneously canonical and provisional.
6. **Stabilize anchors and ship link-checking.** Unnumber section headings, adopt stable slugs, and promote the deferred `XD006` reference-rot lint to shipped — the corpus already has broken deep-links and two wrong ADR-path references, despite the strongest doc-lint regime around.
7. **Split the four oversized files** along contract/rationale seams (05-intent → lifecycle + sub-verb reference + PRD/freeze protocol + reviewer/audit-loop); set a soft ~300-line/file guidance in 00-meta. Re-key the verification matrix per surface.
8. **Settle the evidence chapter's operating rules**: advisory frontmatter flag, manual population permitted now (don't wait for disembark), staleness/pruning rule, and a crisp tradeoffs-vs-ADR partition test (Pocock's three clauses, already used in ADR-7, is the natural discriminator: passes → ADR; fails but useful → tradeoffs rolodex).
9. **Formalize normative language**: RFC 2119 conformance note in 00-meta; reserve caps for normative keywords; give acceptance bullets stable IDs (`AC-ahoy-03`) so reviews, the verification matrix, and receipts can cite them; consider EARS (already acknowledged in ADR-7) for new criteria.
10. **File-slot hygiene rule in 00-meta**: numbers are append-only and retired-never-reused (mirroring the lint-code rule), ending the slot-07 recycling pattern.

**Bottom line:** the taxonomy is sound and in several dimensions (glossary, verification matrix, out-of-scope, lifecycle taxonomy, lint-as-requirements) ahead of anything in spec-kit/arc42/PRD practice. The two systemic weaknesses are (a) delivery-status prose fused into the current-state contract — a slow-motion violation of adr-5's own doctrine — and (b) abcd-specific content baked into what is about to become a generic skeleton. Fix 1–3 before freezing; 4–10 are important but survivable post-freeze.
