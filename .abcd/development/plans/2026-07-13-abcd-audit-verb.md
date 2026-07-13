# Plan — a read-only `abcd audit` convention verb (iss-86)

Status: **agreed 2026-07-13** — formalised as intent `itd-85` (draft); P1 not
yet started. Backing issue: `iss-86` (onboarding-audit-not-engine-backed).
Adjacent: `iss-87`, `iss-88`, `iss-90`, `iss-91` (all from the Manuscripts
dogfood).

## The gap

`abcd ahoy` reports abcd's own install-plumbing gaps (config, rules, marker
block, history store) and is silent on the *convention* gaps `prepare-this-repo`
exists to fix: a missing committed `.abcd/work/` tier, an absent `AGENTS.md`
router, decisions leaking into the gitignored layer, docs drift, privacy
hygiene. The onboarding audit is done by an agent reading prose, not by the
binary — so onboarding is glue, not engine. `abcd audit` is the read-only
surface that closes that: the binary checks the conventions, and
`prepare-this-repo` calls it instead of hand-auditing.

## SOTA declaration (per prefer-sota / sota-per-intent)

- **State of the art:** repolinter (declarative repo-convention rulesets), OPA
  Conftest (policy + severity/exit semantics), SARIF 2.1.0 (result interchange),
  and the `doctor`-command UX pattern. All mature.
- **Maturity vs fit:** every candidate engine is a separate runtime (Node,
  Python, Docker) — a new dependency, which is a hard stop. The one
  embeddable-in-Go option (OPA/Rego as a library) is rejected on weight and on
  forcing rule-authors to learn Rego. repolinter itself is archived.
- **Pick: bespoke-with-seam.** Build on the existing `internal/core/lint` rule
  engine (severity + fix-message + allow-context, already driven over docs).
  **Adapt** three external patterns behind seams, copying vocabulary not code:
  repolinter's rule-object *schema*, Conftest's severity→exit-code semantics,
  and SARIF as an *optional export* format. **Net new dependencies: zero.**
- Seam check: bespoke is only allowed with a seam — satisfied (rule-loader seam
  + output-serializer seam). No hard stop tripped.

## Design

1. **New verb `abcd audit`, distinct from `ahoy doctor`.** `doctor` answers
   "is the tool installed/configured for this repo" (environment/setup health);
   `audit` answers "does this repo conform to the working conventions." Mixing
   them muddies both. Read-only; runs against any repo given only a working
   directory.
2. **Engine: reuse `internal/core/lint`.** It already models severity +
   fix-message + allow-context escapes. Add the two primitives it lacks:
   - **path presence/absence** (`.abcd/development/`, committed `.abcd/work/`,
     gitignored `.abcd/.work.local/`, `AGENTS.md`);
   - **structural assertions** (a path is gitignored; a directory holds ≥1
     entry). Content banned-token checks (absolute paths, emails, private-repo
     names, change-narration tense) are the existing docs-lint primitive reused.
3. **Rule schema (adapt repolinter's vocabulary).** Rules are declarative data,
   separate from the evaluator: `{ id, severity (error|warn|off), where
   (conditional enablement, e.g. only check Diátaxis if docs/ exists), fix
   (remediation message), policyInfo (why this rule) }`. Bundled defaults ship
   in-binary; a repo may override/extend via config (later phase).
4. **Output & exit codes.** Native compact JSON on stdout (primary, stable
   `ruleId`s), a `doctor`-style human render (grouped by category, severity
   glyph, inline fix message, diagnostics to stderr so `--json` stays clean).
   Conftest-style tri-state exit: `0` clean · `1` warnings only · `2` any error.
   `--format sarif` is an optional serializer (deferred — see phasing).

## v1 rule set (the acceptance corpus)

Mirrors the `prepare-this-repo` Phase 2 audit:

| id | severity | checks |
|---|---|---|
| `three-tier-layout` | error | `.abcd/development/` present; committed `.abcd/work/` present; `.abcd/.work.local/` present **and** gitignored |
| `conventions-router` | error | `AGENTS.md` present (CLAUDE.md/GEMINI.md may be bridges) |
| `decision-durability` | warn | a committed `.abcd/work/DECISIONS.md` exists; decisions not living only in the gitignored layer |
| `docs-currency` | warn | reuse existing docs-lint: Diátaxis shape (where `docs/` exists), present-tense, no stray root markdown |
| `privacy-hygiene` | error | no absolute local paths, real emails, or private-repo names in committed files, honouring waiver escapes |

## Wiring (wired-or-it-isn't-done)

- `prepare-this-repo` Phase 2 calls `abcd audit --json` and renders it, instead
  of the agent hand-producing the gap report. This is the change that makes
  onboarding engine-backed — the point of `iss-86`.
- Standalone + CI: the tri-state exit code lets `abcd audit` gate a repo's CI.

## Draft acceptance criteria (BDD)

- **Given** a repo missing `.abcd/work/`, **when** `abcd audit --json`, **then**
  the `three-tier-layout` rule reports `severity: error` with a fix message
  naming the missing tier, and the process exits `2`.
- **Given** a repo whose decisions live only in gitignored `.work.local/`,
  **when** audit, **then** `decision-durability` reports `warn` and exit is `1`
  (no errors).
- **Given** a committed file containing an absolute local path, **when** audit,
  **then** `privacy-hygiene` reports `error` citing `file:line`, **unless** a
  waiver escape is present on that line.
- **Given** a conforming repo, **when** `abcd audit`, **then** exit `0` and a
  green human render; **and** `abcd audit --json` emits `{ "findings": [] }`.
- **Given** `docs/` is absent, **when** audit, **then** the `docs-currency`
  rule is skipped via its `where` condition, not failed.

## Phasing

- **P1 — engine + rules.** Rule schema + loader + the two new primitives on
  `internal/core/lint`; native JSON + human render; the five v1 rules; exit
  codes. One watched-fail→pass test per rule.
- **P2 — wire into onboarding.** `prepare-this-repo` Phase 2 consumes
  `abcd audit --json`. Plugin-surface + CLI parity (`commands/abcd/audit.md`),
  and a `04-surfaces` registry row (shipped) so `surface_coverage` stays honest.
- **P3 — optional SARIF export.** `--format sarif` serializer behind the seam;
  only if CI/IDE ingestion is wanted.

## To do at implementation (not now)

- `ACKNOWLEDGEMENTS.md` entries for the adapted patterns (repolinter rule
  schema, Conftest severity/exit semantics, SARIF) — added in the same change
  that lands them, never retroactively.
- A CHANGELOG entry (user-facing verb).
- Promote to intent `itd-85` (press-release-first + these AC) once the shape is
  agreed; `abcd intent plan` mints the spec and is the maintainer adoption gate.

## Decisions (agreed 2026-07-13)

1. **New `abcd audit` verb** — distinct from `ahoy doctor` (doctor = tool-setup
   health; audit = repo conformance).
2. **v1 rule set as drafted** — the five rules above; the `iss-88` managed-repo
   check stays an `ahoy`/detection concern, not folded in.
3. **SARIF deferred to P3** — native JSON first, `--format sarif` behind the seam.
4. **Formalised as intent `itd-85`** (draft) — press-release + these acceptance
   criteria + SOTA declaration. `abcd intent plan itd-85` is the adoption gate.
