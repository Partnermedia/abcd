# Principles vs state of the art — 2026-07-08

Each of the ten principles recorded from the 2026-07-08 review was mapped
against current (2026) state of the art by a dedicated web-grounded research
pass — prior art with verified attribution, tooling, refinements, and
explicit anti-adoption calls, evidence-tiered (evidence / consensus /
anecdote-marketing). This note is the distillation; per-principle sources are
inline.

## Overall result

None of the ten is behind SOTA in substance. Seven are aligned; three are
ahead of published practice in at least one specific respect. The recurring
gap is uniform and unsurprising: **operationalisation** — every principle
currently holds by discipline, and SOTA's consistent finding is that
conventions survive only as automated checks. That is the fix-the-detector
thesis restated, so the mapping is self-consistent: the promotion paths the
principle files already name are the actual deliverable.

Three elements appear to be genuine contributions (no published equivalent
found):

- **retire-the-name's same-change coupling** — ban-travels-with-rename exists
  structurally only in annotation-based code ecosystems (`@InlineMe`,
  `//go:fix inline`); in the docs/terminology world GitLab and Grafana run
  retired-name Vale rules but with no same-change mandate (verified by direct
  fetch). Stating it for declaration-less names (paths, flags, terms) is new.
- **unrecognized-input-never-writes' promotion clause** — no published CLI
  guideline mandates a per-mutating-verb test that malformed input produced
  no write. The guides cover unknown flags; the payload-absorption seam that
  bit this repo is under-covered everywhere.
- **one-canonical-primitive's boundary statement** — the literature implies
  but never crisply states the infrastructure/domain split; the derived
  litmus is worth adopting: *if two copies must change together to stay
  correct, it is one piece of knowledge — infrastructure; if they may
  legitimately diverge, it is domain.*

## Per-principle verdicts and refinements

### fix-the-detector — aligned, at the strict end

Prior art: Spolsky "fix everything two ways" (2007); poka-yoke (Shingo);
the Beyoncé Rule (*Software Engineering at Google*); Google Tricorder
(ICSE 2015 / CACM 2018); CodeQL variant analysis; Semgrep rule-from-incident;
ESLint bulk suppressions (2025). Refinements: (1) add a false-positive budget
and a kill criterion — Tricorder's central finding is that detectors above
~10% perceived-FP rate get dismissed wholesale; (2) the acceptance corpus
must include *negative* examples (Semgrep `ruleid:`/`ok:` convention) —
recall-only proof is half a proof; (3) run variant analysis before arming so
the baseline is the true instance count, not the review's sample; (4) ship
the autofix with the detector where the class is mechanical; (5) name AI
agents as the detector's runtime consumer — deterministic gates are how
agents self-correct. Attribution flag: "pave the cowpaths" is a W3C/UX
principle, not lint prior art; "never fix a bug twice" is folklore with no
coiner.
Sources: cacm.acm.org/research/lessons-from-building-static-analysis-tools-at-google;
semgrep.dev/docs/writing-rules/testing-rules; eslint.org/blog/2025/04/introducing-bulk-suppressions.

### enforcement-claims-are-facts — aligned in substance, behind in operationalisation

Prior art: Martraire's *Living Documentation* "reconciliation mechanism" (the
closest named ancestor of the gate cross-check); SOC 2 control-gap auditing
(a described-but-never-run control is an audit exception); SLSA/in-toto
("provenance without verification is security theater"); Diátaxis on
generated reference. The vigilance-suppression rationale has controlled
human-factors backing (automation-complacency: Parasuraman & Riley 1997,
Parasuraman & Manzey 2010) — upgrade the Why from internal anecdote to cited
evidence. Refinements: (1) state the enforcement hierarchy — *generated from
the mechanism* beats *verified against it* beats *discipline*; (2) name the
third state: a gate that runs but does not block is a half-phantom, so
claims must state blocking semantics; (3) "demonstrably runs" means ran on
recent real changes, not merely defined (SOC 2 Type I/II distinction);
(4) gate names become greppable identifiers with exactly one authoritative
definition, making the cross-check a string match. Tool fit: mdox-style
regenerate-and-diff, Go testable examples, cram-style session replay.
Sources: oreilly.com/library/view/living-documentation-continuous/9780134689418;
abseil.io/resources/swe-book/html/ch10.html; diataxis.fr/reference.

### retire-the-name — aligned on mechanism, ahead on the coupling

Prior art: *SWE at Google* ch. 15 (deprecation = discovery, migration,
**backsliding prevention**); `//go:fix inline` and Error Prone `@InlineMe`
(ban-marker-travels-with-retirement by construction); GitLab/Grafana Vale
substitution rules (retired terms → successor). Refinements: (1) every ban
entry names its successor (`old → new`) so the ban is a migration aid;
(2) sweep-to-zero then ban in the same change; where zero is unreachable,
use a counted baseline, not a broad allow-context; (3) allow-contexts stay
narrow and enumerable — an open "historical passages" category is the
reintroduction vector; (4) for Go identifiers prefer `// Deprecated:` plus
staticcheck (annotation-adjacent beats out-of-band); reserve banned_tokens
for the declaration-less cases it was built for; (5) date-stamp entries and
prune bans whose term no longer appears anywhere.
Sources: abseil.io/resources/swe-book/html/ch15.html; go.dev/blog/inliner;
docs.gitlab.com/development/documentation/styleguide/word_list.

### ratchet-not-big-bang — aligned, on the strong side

The pattern became first-class in mainstream tooling 2024–25 (ESLint bulk
suppressions v9.24, Notion's eslint-seatbelt, PHPStan baselines, Android lint
baselines); notably the default Go linter still lacks frozen-baseline support
(open request golangci-lint #3356), so a record-lint baseline fills a real
gap. Refinements: (1) prune-by-default — a fixed violation still in the
baseline fails the gate (ESLint's default semantics); (2) granularity is
exact entries or per-file-per-rule counts, never file-level excludes
(RuboCop's documented fossilisation vector), with ESLint's spillover rule;
(3) declare a non-baselinable tier: security/trust-boundary rules arm at
mandatory zero (Abramov, "Suppressions of Suppressions", 2025); (4) the
regeneration rule is mechanical — new baseline ⊆ old, checked by the gate
itself; (5) one entry per line, stable sort (Notion's merge-conflict fix);
(6) emit `baseline: N (was M)` per run. Explicitly rejected: per-entry expiry
dates (zero tool adoption; ritual renewal is the predictable failure),
diff-based gating (`--new-from-rev` resurrects on file moves, misses
aggregate lints), time-window new-code definitions (violations age out
silently).
Sources: eslint.org/docs/latest/use/suppressions; notion.com/blog/how-we-evolved-our-code-notions-ratcheting-system-using-custom-eslint-rules;
phpstan.org/user-guide/baseline; qntm.org/ratchet;
overreacted.io/suppressions-of-suppressions.

### one-canonical-primitive — aligned, ahead on the boundary statement

Prior art: DRY-as-knowledge (Hunt & Thomas); rule-of-three (Fowler) — which
governs abstraction *creation*, so it lends no licence to copy once the
canonical home exists; Google one-version rule; paved-road/golden-path
(discoverability failure is the causal driver of reinvention). Evidence:
ALICE (OSDI 2014) found 60 crash-consistency bugs in mature applications,
several exactly the missing-directory-fsync class; Juergens (ICSE 2009)
traced ~107 faults to inconsistently-updated clones; GitClear 2025 shows
AI-assisted coding measurably accelerating duplicate-block creation.
Refinements: (1) the canonical primitive states its crash-consistency
*contract* explicitly — even google/renameio scopes to atomicity, not
durability, and says so; (2) accretion guard: an option that changes the
crash-consistency contract is a second named primitive, not a flag (Metz's
wrong-abstraction failure mode); (3) state the discoverability mechanism in
the rule (canonical-primitives table where agent context sees it — the
copier in 2026 is often a code assistant); (4) disclaim the Go-proverb
objection: "a little copying" addresses cross-module dependency cost, not
intra-repo imports. Rejected: generic clone detectors as the gate — a
15-line copy that *omits* the fsync is below token thresholds, and the more
divergent (broken) it is, the less detectable; name-based bans invert this.
Promotion shape: a go/analysis vet check flagging `os.Rename`/`os.WriteFile`
outside the canonical package — no new dependency.
Sources: research.cs.wisc.edu/adsl/Publications/alice-osdi14.pdf;
github.com/google/renameio (issue 11); sandimetz.com/blog/2016/1/20/the-wrong-abstraction;
gitclear.com/ai_assistant_code_quality_2025_research.

### guards-prove-themselves — strongly aligned; independently reinvented five times

The same rule exists as Semgrep rule fixtures, SIEM detection unit tests
(Panther, Splunk contentctl), policy-as-code testing (OPA `opa test`,
Sentinel paired pass/fail mocks), security chaos engineering (Shortridge &
Rinehart), and OWASP SAMM abuse-case regression suites. Hard evidence for
"guards fail silent": CardinalOps' production SIEM datasets — 13–18% of
deployed detection rules can never fire, unnoticed, because a rule that
never fires looks identical to a rule with nothing to catch. Refinements:
(1) the paired test must *depend* on the guard — cheap mutation discipline:
disable the guard once, watch the test go red, restore (the targeted form of
Google's changed-lines mutation testing); (2) make the rule bidirectional —
every guard needs the `ok:` case too, because an over-eager guard gets
silently loosened by whoever it blocks, and the loosening is the regression;
(3) every bypass ever found becomes a permanent fixture colocated with the
guard; (4) continuous verification where the guard *runs*: a doctor-style
self-check feeding each live guard a known-bad input — CI proves logic, not
deployment; (5) never gate on coverage percentage — 100% line coverage of a
guard file proves nothing about refusal.
Sources: owaspsamm.org/model/verification/requirements-driven-testing;
semgrep.dev/docs/writing-rules/testing-rules; arxiv.org/abs/2102.11378;
cardinalops.com/blog/3rd-annual-state-of-siem-detection-risk-report-mitre-attck;
securitychaoseng.com.

### unrecognized-input-never-writes — aligned; promotion clause is original

Prior art: clig.dev ("check early and bail out"); RFC 9413 (the IETF's
formal revision of Postel — strictness for maintained systems); git's
twenty-year default of suggest-and-exit-nonzero, with its opt-in autocorrect
as the canonical cautionary tale (the 100ms cancel window shown faster than
human reaction time, reworked in Git 2.49); clap/Cobra did-you-mean
conventions. The repo's actual incident sits in a documented Cobra gap:
suggestions never fire when a parent accepts free-text positional args — the
typo falls through as payload. Refinements: (1) guard the payload seam, not
just the flag parser — parents take `cobra.NoArgs`; free-text verbs run
their first token through the edit-distance check against sibling verb names
before writing; (2) a stable JSON error envelope with a machine-dispatchable
`code`, `message`, `suggestion` — agent hosts dispatch without
string-matching (take the RFC 9457 shape, skip the URI machinery); (3) a
distinct exit code for usage errors vs operational failures (jq, gh,
sysexits EX_USAGE); (4) document which stream carries the JSON error body
and never mix; (5) if a confirmation prompt is ever added: non-TTY resolves
to refuse, never to assume-yes.
Sources: clig.dev; datatracker.ietf.org/doc/html/rfc9413;
blog.gitbutler.com/why-is-git-autocorrect-too-fast-for-formula-one-drivers;
cli.github.com/manual/gh_help_exit-codes.

### spec-moves-with-the-surface — aligned with the harder, rarer camp

2025–26 spec-driven development splits into spec-first (Kiro, spec-kit —
spec written, then decays; documented "illusion of work" failure) and
spec-anchored (the spec stays authoritative: buf breaking, oasdiff contract
gates — a decade of scale evidence). This principle is spec-anchored. Rust's
stabilisation-report practice independently validates keying the spec to the
*shipped* surface: the truthful spec is written at landing time, because
implementation always diverges from design. GitLab — the maximal docs-as-code
culture — could not hold same-change docs org-wide by norm alone; the
promotion is the deliverable, not a nice-to-have. Refinements: (1) make the
record-lint cross-check **bidirectional** — surface→row catches unspecified
live verbs, row→surface catches spec'd-but-unreachable features (the exact
sibling-project failure the principle cites); contract gates diff both
directions; (2) give the amendment an audit trail — the brief file must be
touched in any change that touches commands/ or skills/, so
criterion-amended-to-admit-surface is visible; (3) rows are fixed fields
(surface, criterion, wiring status) — deterministic gates outlast semantic
ones; (4) do not extend toward spec-as-source (Tessl's inversion inherits
model-driven development's inflexibility plus LLM non-determinism).
Sources: buf.build/docs/breaking; github.com/oasdiff/oasdiff;
martinfowler.com/articles/exploring-gen-ai/sdd-3-tools.html;
forge.rust-lang.org/libs/maintaining-std.html;
docs.gitlab.com/development/documentation/workflow.

### loud-staging — a legitimate synthesis; no single named SOTA practice exists

The principle unifies four separately-named practices: Fowler's keystone
interface / dark launching (surface-refusal leg), feature-flag lifecycle
governance with owner+ticket+expiry at creation (record-row leg), ticketed
TODO lints (doc-comment leg), and Rust's `#[expect(lint, reason)]` — the
only mainstream mechanism where the disclosure itself expires automatically
(`unfulfilled_lint_expectations` fires when the suppressed condition stops
holding). Ahead of common practice on three-site disclosure; behind on
enforcement. Refinements: (1) a fixed machine-readable grammar —
`// Staged: wired by itd-N.` — joins the three sites by script; (2) the
promotion: run `x/tools/cmd/deadcode` per main package in CI; every reported
symbol must match a Staged annotation referencing an open intent, else fail;
(3) lease expiry enforced by the same join — intent closes, annotation must
go in that change; (4) a calendar backstop so a stalled intent cannot extend
the lease indefinitely; (5) state the tracer-bullet default explicitly:
staging a layer is the exception, a thin wired end-to-end slice is the norm
(Hunt & Thomas; Cockburn's walking skeleton); (6) prefer dark-launch where
cheap — executed-with-results-discarded cannot rot silently. Evidence that
expiry automation works at scale: Uber's Piranha (ICSE-SEIP 2020) generated
cleanup diffs for 1,381 stale flags, 65% landed unmodified. Rejected:
blanket `//nolint:unused` or keep-alive idioms — they defeat the
reachability signal and are worse than no check.
Sources: martinfowler.com/bliki/KeystoneInterface.html; go.dev/blog/deadcode;
doc.rust-lang.org/reference/attributes/diagnostics.html;
manu.sridharan.net/files/ICSE20-SEIP-Piranha.pdf.

### reality-is-filable — strongly supported by five independent lineages

Bowker & Star's *Sorting Things Out* (1999) is the academic anchor:
residual categories are where classification systems silently absorb
unfilable reality — "classifications enact silences", the principle's claim
27 years earlier, with the design answer *distribute and control residual
categories rather than pretend they will not occur*. Evans (DDD ch. 9):
filing awkwardness is the pre-failure signal — "scrutinize awkwardness".
Kanban/lean (Benson & DeMaria Barry's *Personal Kanban*; DeGrandis's *Making
Work Visible* — note: not DeMarco); value-stream mapping; event sourcing;
Nygard's ADR gate (matched exactly by the principle's bounds). Refinements:
(1) capture the event even when the state lags — an always-writable
append-only layer means a transition-fact is never lost while the taxonomy
fix is pending; the v0.1.0 fact survived only because the CHANGELOG existed;
(2) add the symmetric bound — the better-documented failure in workflow
tooling is taxonomy *bloat*, not gaps (the ≤7-status consensus; states are
added on observed unfilable reality, never speculatively); (3) complete
expand-contract: retire states that go permanently empty, or the taxonomy
stops mapping reality from the other direction; (4) the fix is not always a
directory — a frontmatter field or ledger line may express the reality more
cheaply than growing the tree; (5) one sanctioned residual category with a
dwell-time lint beats pretending zero residual is achievable.
Sources: direct.mit.edu/books/monograph/4738; martinfowler.com/bliki/ParallelChange.html;
cognitect.com/blog/2011/11/15/documenting-architecture-decisions;
kurrent.io (Young, event sourcing).

### readme-and-currency footnote

The mapping surfaced one cross-principle theme worth recording: several
researchers independently flagged that the strongest 2026 argument for the
whole batch is agent consumption — stale or phantom claims are worse than
absent ones because agents consume them uncritically on every run, and
deterministic detectors are the mechanism agents self-correct against. The
record is not just read by people.

## Recurring refinement themes (candidates for a principles revision pass)

1. **Bidirectionality** — spec cross-checks gate both directions; guards
   need pass-cases as well as refuse-cases; ratchets need prune-by-default.
   One-directional checks are half-checks.
2. **Corpora need negative examples** — recall against found instances
   proves half; near-miss non-findings prove the other half.
3. **Detector lifecycle** — false-positive budgets, kill criteria, and
   successor-naming keep the enforcement layer itself from becoming debt.
4. **Event capture beside state taxonomy** — append-only ledgers are the
   layer that saves facts while structures lag.
5. **Contracts stated, not implied** — the canonical primitive names its
   durability contract; the gate claim names its blocking semantics; the
   ban names its successor.
