# abcd Design Principles — SOTA grounding

Research deliverable. Scope: the six cross-cutting design principles distilled
from the brief, intents, ADRs, and code (extraction recorded separately; the
load-bearing five plus recovery-humility). For each principle this note records
the **canonical prior art** (who established the idea), the **current
best-practice consensus**, **how abcd aligns or diverges**, and at least one
**tension** worth carrying forward.

Method: one deep web-research pass per principle, each followed by an
adversarial fact-check pass that verified attributions, dates, and citation
soundness and trimmed or repointed weak sources. The fact-checker caught and
corrected: a fabricated verbatim Nygard quote (now a paraphrase), a
misattributed least-surprise citation (repointed to Raymond's primary text), a
marketing-blog git citation (repointed to `git-scm.com`), an over-broad
SoftwareMill title, a Goedecke date error, and an over-strong "coins" claim for
knowledge vaporization. Residual cautions are recorded inline.

Companion documents: [`related-work.md`](../related-work.md) (the prior-art
frameworks abcd shares DNA with — PAUL, CARL, Karpathy, Naur, Dell'Acqua);
[`../../brief/01-product/03-mental-model.md`](../../brief/01-product/03-mental-model.md)
(the four-layer model that Principle 1 grounds);
[`../../brief/05-internals/04-universal-patterns.md`](../../brief/05-internals/04-universal-patterns.md)
(the patterns Principles 3 and 4 implement).

> **What this note is not.** It is not a decision (those are ADRs) and not
> forward scope (those are intents). It is the evidence layer under the
> principles — why they are defensible, and where they knowingly diverge from
> consensus. Per [`README.md`](./README.md), a note may *inform* an ADR or
> intent; it is not one.

---

## 1. Right grain, right home

*Every concern lives in the artefact whose question and lifecycle match it.
Brief = what, Intent = why (press-release-shaped), Spec = how, Phase =
what-this-stretch-makes-true. Decision records split by lifecycle: ADR (settled,
retrospective) vs RFC (contested, prospective) vs intent (forward, user-facing).*

**Canonical prior art.** Four independently-canonical ideas converge here. The
**Single Responsibility Principle** — "a module should have only one reason to
change" ([Robert C. Martin][martin-srp]) — is the deepest root, applied to prose
rather than classes. **ADRs** ([Michael Nygard, 2011][nygard-adr]) establish the
lifecycle-keyed, single-decision, append-only record: per Nygard, a reversed
decision is kept and *marked superseded* rather than edited (paraphrased — the
"never reopen, supersede instead" lifecycle is faithful to his post, though not
a verbatim quote). [Martin Fowler][fowler-adr] independently confirms an ADR
"captures and explains a single decision" and credits Kruchten's decision
registers as antecedent. **Diátaxis** ([Daniele Procida][diataxis]) is the
documentation analogue: four modes split by user need, where "crossing or
blurring the boundaries… is at the heart of a vast number of problems in
documentation." **Working Backwards / PR-FAQ** ([Bryar & Carr, 2021][working-backwards];
the practice predates the book — commonly dated to the mid-2000s, though that
origin date is an approximation, not a cited fact) is the direct source for
"Intent = why, framed as a press release." [Specification by Example][adzic-sbe]
(Gojko Adzic, 2011; Jolt Award 2012) grounds "Spec = how" as living,
single-source-of-truth documentation. The RFC-vs-ADR split (prospective/contested
vs retrospective/settled) is [community convention][candost-adrs-rfcs], not a
single named originator.

**Consensus.** Design and documentation artefacts are separated by purpose and
lifecycle, not lumped together: a prospective feedback-seeking proposal (RFC)
versus a committed, immutable, single-decision record (ADR, superseded rather
than edited); product framing that starts from the customer outcome before
implementation; user docs partitioned by user need; specs that are living and
example-driven. The unifying meta-principle — increasingly stated explicitly —
is SRP applied to prose: each artefact has one question and one reason to change.

**abcd alignment.** Faithful and unusually explicit. Brief/Intent/Spec maps
cleanly onto Working Backwards, Specification-by-Example, and Diátaxis-style
separation by question. ADR(settled) vs RFC(contested) is textbook consensus.
"Plumbing has no user moment, so it lives in the brief not an intent" is a crisp
application of Amazon's rule (no customer narrative → no PR-FAQ) and of SRP.
**Divergence:** abcd elevates *intent* to a third decision-record kind alongside
ADR and RFC — the mainstream landscape recognises only the RFC/ADR pair, with
the forward user-facing "why" normally living *inside* a PR-FAQ or RFC. This is a
defensible local fusion of the PR-FAQ's customer-narrative role with a lifecycle
slot, but it is abcd's own taxonomy, not established prior art. "Phase =
what-this-stretch-makes-true" likewise has no direct canonical analogue.

**Tensions.**
- Boundaries blur in practice. Diátaxis's own community notes how-to guides
  bloat with explanation; "right grain, right home" needs continuous curation
  and can decay into a bureaucratic sorting exercise. The taxonomy is
  analytical, not a set of rigid bins — resist over-literal enforcement.
- Four artefact types × three decision-record kinds multiplies the places a
  reader must look and a writer must choose between — the very fragmentation
  single-source-of-truth aims to prevent. SRP's known critique (what counts as
  "one reason to change" is subjective) applies equally to "one question, one
  home."

---

## 2. No-surprise operations

*Defaults, prompts, mutations, and backend choices are explicit and reversible.
Bare invocation renders status/help with zero writes; mutation requires an
explicit verb. Visibility is one coarse switch.*

**Canonical prior art.** The root is the **Principle of Least Astonishment**
(POLA/POLS) — earliest known reference in the [PL/I Bulletin, 1967][wikipedia-pola];
codified in the Unix tradition by [Eric S. Raymond][raymond-taoup] as the "Rule
of Least Surprise: …always do the least surprising thing" (*The Art of Unix
Programming*, 2003). The "no silent defaults" arm is the **Zen of Python**
([Tim Peters][pep-20]: "Explicit is better than implicit"; "In the face of
ambiguity, refuse the temptation to guess"; "Errors should never pass
silently" — posted to python-list 1999, published as PEP 20 in 2004). The
operational template is the **Command Line Interface Guidelines**
([clig.dev][clig-dev]: "make the default the right thing for most users"; provide
`-n, --dry-run`; make destructive actions "hard to confirm by accident"; "if you
change state, tell the user"; help on no-args). **git** is the reference
implementation of safe-by-default for the remote with opt-in destructiveness
(`--force-with-lease` as the safe alternative to `--force` — see the
[official `git push` docs][git-push]). The reversibility axis is Bezos's
**one-way vs two-way doors** ([Amazon 2015][bezos-2015] and
[2016][bezos-2016] shareholder letters): irreversible decisions warrant
methodical friction, reversible ones stay light.

**Consensus.** Predictability and safety are design defaults, not afterthoughts:
separate read from mutate; make defaults safe rather than maximally convenient;
offer dry-run/preview and tier confirmation by severity; never fail or mutate
silently; and match friction to reversibility.

**abcd alignment.** A faithful synthesis. "Bare command renders status/help with
zero writes; mutation requires an explicit verb" is the read/mutate separation
git and clig.dev embody — a stronger, cleaner version of clig.dev's "display
help when run with no arguments." Transparent prompts (state + consequence +
how-to-change) operationalise least-astonishment and clig.dev's "if you change
state, tell the user," and the how-to-reverse element goes slightly beyond what
the sources spell out. **The one genuine divergence** is "visibility is one
coarse switch, no per-subdirectory exceptions": the canon (git nested
`.gitignore`/config) tolerates fine-grained overrides; abcd deliberately trades
expressiveness for predictability. Defensible under least-astonishment (fewer
hidden modes) but an opinionated choice, not inherited consensus.

**Tensions.**
- Least astonishment is relative to the audience's mental model — what is
  unsurprising to a Unix expert can astonish a novice. abcd's rules assume a
  developer-shaped expectation set and give no objective arbiter when populations
  diverge.
- The coarse single-switch rule can push users with a legitimate per-context need
  toward awkward workarounds — itself a source of surprise.
- Safe-by-default adds friction; both clig.dev and Bezos warn that uniform heavy
  process is a cost. If a mutation is genuinely a two-way door, demanding
  ceremony for it is its own astonishment.

---

## 3. Integrate without owning

*abcd is a configuration, not a fork. It never patches upstream
(flow-next/Ralph) freehand: deviations live in a declarative overlay manifest,
re-applied idempotently after upstream re-vendors, capped, and watched by a
dependency-watcher that retires entries when upstream lands the fix. Prefer an
external provider with a structural fallback; depend on a semantic role, not a
vendor; never pick the model.*

**Canonical prior art.** Three mature lineages converge. **Overlay-not-fork**
([kustomize, Regan & Wittrock / Kubernetes SIG-CLI, 2018][kustomize-intro]): treat
upstream config as data, express deviations as a small declarative overlay on an
untouched base — copying/forking "makes it difficult to benefit from ongoing
improvements." **Patch-queue discipline** ([Debian 3.0 (quilt)][debian-debsrc3];
[quilt push/pop][debian-quilt]): downstream changes live as a managed, ordered,
individually-described stack over pristine upstream. **Upstream-first carry-patch
retirement** ([Chromium OS][chromium-upstream]): minimise downstream-only code,
document every divergence, prefer landing fixes upstream so the carry-patch can
retire; rejected-but-kept patches are labelled with a link and reason. The
adapter half is **Ports and Adapters / hexagonal architecture**
([Alistair Cockburn][cockburn-hexagonal]) — depend on a port, not a concrete
vendor — and the wrap-don't-fork model is the **Strangler Fig**
([Martin Fowler, 2004][fowler-strangler]). Runtime robustness is **graceful
degradation** ([AWS Well-Architected REL05][aws-rel05]: "transform applicable
hard dependencies into soft dependencies"). The cost being avoided is **fork
debt** ([Nick Desaulniers][desaulniers-fork]: rebasing cost proportional to
divergence; eroded knowledge of intent; slower security response).

**Consensus.** Never edit upstream in place and never hard-fork unless you must;
keep upstream pristine and express every deviation as a discrete,
individually-justified, idempotently re-appliable unit over an unmodified base —
coupled with an upstream-first culture that actively retires carry-patches once
fixes land, because un-retired divergence accrues compounding fork debt. For
pluggable dependencies, depend on a port/semantic role; wrap rather than fork to
add behaviour; turn hard dependencies into soft ones with ordered fallback.

**abcd alignment.** Unusually faithful, and in two places *ahead of* common
practice. The "declarative overlay manifest re-applied idempotently after
upstream re-vendors over un-patched upstream" is a near-verbatim instance of the
kustomize base/overlay philosophy and the Debian quilt patch-queue. The
**dependency-watcher that retires overlay entries when upstream lands the fix**
operationalises Chromium's upstream-first culture as *automated machinery*, where
distros and most overlay tools leave retirement a manual practice. The **soft cap**
encodes the "minimise downstream-only code" norm numerically, which distros hold
culturally but rarely enforce. "Vendor-agnostic adapters by semantic role,"
"alternative entry points that call upstream under the hood," and "never pick the
model" map cleanly onto ports-and-adapters, strangler-fig, and
graceful-degradation/dependency-inversion. (This is the relationship the user's
own framing names: abcd never changes upstream dependencies; it offers an
alternative entry point that calls them under the hood, with fixes baked in.)

**Tensions.**
- Overlay/patch indirection is itself hard to debug — patches can mis-target or
  silently no-op when upstream restructures ([the kustomize JSON6902 silent
  "add-as-replace" failure][pauldally-json6902] is the canonical caution; the
  general "spooky action at a distance" point holds even though that specific
  article does not demonstrate the ordinal-index variant). The overlay needs its
  own fire-test that each entry *actually applied* — which abcd's fire-test gate
  and per-entry receipts are built to provide.
- The cap and watcher are themselves machinery that must be maintained
  (meta-fork-debt), and a cap can perversely incentivise squashing several
  deviations into one opaque entry to stay under it — defeating the auditability
  the patch-queue discipline exists for. The value depends on each entry carrying
  a documented reason and retirement trigger, not on the count.
- Adapters defer but do not eliminate fork risk: when upstream changes *semantics*
  (not structure), an adapter that still "applies" can be quietly wrong. The
  divergences upstream actively refuses are exactly the ones the watcher can
  never retire — so a residual, permanently-carried overlay floor is unavoidable
  and should be named as such, not treated as transient.

---

## 4. Canonical state is structured, auditable, lifecycle-declared

*Inter-agent data is schema-backed JSON; markdown is a deterministic render.
Every namespace declares a lifecycle class — regenerable (full-crawl-on-demand,
idempotent, never delta-application), append-only, or compounding-curated
(curator state + per-entry provenance). One authoritative home per lifecycle; an
immutable identity key survives rename/move.*

**Canonical prior art.** Four traditions. **Level-triggered reconciliation**
([Tim Hockin, Kubernetes][hockin-level]; Joe Beda): controllers compare full
desired vs actual state and converge, so a missed event self-heals on the next
reconcile — the direct ancestor of abcd's "full-crawl-on-demand, never
delta-application." **Declarative desired-state IaC** ([Terraform][terraform-iac]):
diff against a state-file source of truth, apply the minimal idempotent change
set. **CQRS / event sourcing** ([Greg Young][young-cqrs]; popularised via
[Fowler][fowler-cqrs]; [Azure pattern][azure-event-sourcing]): an append-only
immutable log is canonical, and read models are rebuildable projections — the
basis for treating markdown as a derived view over canonical JSON, and for the
append-only and regenerable classes. **Content-addressable identity**
([git, Torvalds][git-content]): objects named by the hash of their content give
identity intrinsic to content/lineage, not to a mutable path — the model for the
immutable root-commit SHA. Provenance is [W3C PROV][w3c-prov] (descriptive
Entity/Activity/Agent) with [SLSA/in-toto][slsa-in-toto] as the verifiable
upgrade path; the schema-first contract is [JSON Schema as single source of
truth][json-schema-contract].

**Consensus.** Keep one authoritative, machine-readable source of truth and treat
everything human-facing as a derived, regenerable projection. For state that must
converge, prefer declarative desired-state plus idempotent level-triggered
reconciliation over imperative deltas (self-healing against missed/out-of-order
events). Distinguish immutable append-only data from mutable derived views; give
artefacts content/lineage-based identity; record provenance as first-class
metadata. The mature caveat is **selectivity** — apply event-sourcing/CQRS/heavy
provenance only where the domain warrants it.

**abcd alignment.** Tight, and abcd names seams most sources leave implicit.
"Schema-backed JSON canonical, markdown a deterministic render" is textbook
CQRS/schema-first. "Regenerable = full-crawl-on-demand, idempotent, never
delta-application" is the Kubernetes/Terraform level-triggered rule transposed
into an agent-knowledge context — a faithful and somewhat novel move. The
immutable root-commit SHA is git's content-addressable insight. **Divergences:**
(1) abcd makes the lifecycle *class* an explicit, declared, per-namespace
property — more prescriptive than any single source. (2) The
**compounding-curated** class has no exact canonical analogue: it carries curator
state in a "view-ish" namespace, a hybrid the strict CQRS literature would
actually *warn against* (curator state in a derived store reintroduces a second
source of truth unless carefully bounded). (3) abcd's provenance is descriptive
(PROV-like) rather than verifiable (SLSA/in-toto) — a deliberate lighter-weight
choice.

**Tensions.**
- Selectivity (Fowler on CQRS): event-sourcing/CQRS "should only be used on
  specific portions of a system" and is "a significant force for getting a
  software system into serious difficulties" when over-applied. A blanket "all
  inter-agent state is structured + lifecycle-classed" risks imposing
  log/provenance ceremony on namespaces better served by plain regenerable files
  — mitigated only if the regenerable class stays genuinely cheap.
- The compounding-curated class is in direct tension with the CQRS rule that
  derived stores hold no authoritative state. Once curator-edited state lives
  there with per-entry provenance, that namespace *is* a source of truth and the
  "markdown is just a render" guarantee no longer holds for it. It needs explicit
  rules for how curator state is reconciled vs regenerated, or it quietly
  violates "no duplicated truth."
- Schema evolution (versioning immutable append-only entries) and full-crawl cost
  at scale are the known frictions: "full-crawl-on-demand" assumes crawls stay
  cheap. Record where that stops being affordable. (The brief's recomputation-
  discipline section already commits to re-evaluating cadence past ~50 entries.)

---

## 5. Cheap hard gates prevent compounding drift

*Disciplines (acceptance criteria, prompt-quality, modification grammar,
vocabulary registration) are a fixed per-spec tax; the failures they prevent
compound; gates are hard from day one, not soft-until-stable. Cheap static checks
(regex / set-membership lint) are kept separate from expensive dynamic checks
(LLM semantic judgement).*

**Canonical prior art.** Four traditions. The **cost-of-defect escalation**
argument ([Boehm & Basili, 2001][boehm-basili]; [Boehm 1981 curve][boehm-curve])
is the canonical "fix it early" basis (its evidentiary weakness is in the tensions).
**Technical debt** ([Ward Cunningham, OOPSLA 1992][cunningham-debt]): expedient
code accrues compounding interest — Cunningham later clarified the metaphor was
about code reflecting *evolving understanding*, not writing messy code. **Broken
windows** (Wilson & Kelling 1982 → software via Hunt & Thomas, *The Pragmatic
Programmer*, 1999), now with direct empirical support: [Amanatidis et
al.][broken-windows-td] found developers ~458% more likely to use
non-descriptive names when extending high-debt systems — debt begets debt.
**Fitness functions** ([Ford, Parsons, Kua & Sadalage, *Building Evolutionary
Architectures*][evolutionary-arch]) formalise gates as automated, continuous
governance, classified atomic-vs-holistic / triggered-vs-continual — the same
axis as abcd's cheap-static vs expensive-dynamic split. The strongest gate is a
design-time constraint: **make illegal states unrepresentable**
([Yaron Minsky; popularised by Scott Wlaschin][minsky-illegal]), with
**policy-as-code** ([OPA / Conftest][opa-conftest]) as the production
instantiation of mechanical set-membership/structural gates. The umbrella is
**shift-left** ([Larry Smith, 2001][shift-left], rooted in Deming's "build
quality in").

**Consensus.** Enforce quality as early and mechanically as possible: cheap
deterministic checks as hard gates on every commit; expensive judgement-heavy
checks later and less often; governance expressed as code; and — best of all — a
design-time constraint that makes the illegal state unrepresentable. Hard from
the start, because retrofitting discipline onto a drifted system is far costlier,
and tolerated small defects measurably breed more.

**abcd alignment.** Strong on three of four claims. "Gates hard from day one" is
the shift-left / broken-windows consensus, with the empirical broken-windows
result a good citation for the contagion mechanism. Separating cheap static lint
from expensive LLM judgement is a clean instance of the static-vs-dynamic
tradeoff and the atomic/holistic fitness-function axis — treating the LLM as the
"expensive dynamic" tier is a sensible extension. Vocabulary registration and
modification grammar are "make illegal states unrepresentable" / policy-as-code
applied to a project's own vocabulary and edit space. **Where abcd should soften:**
"the failures they prevent compound *exponentially*" inherits the Boehm
exponential framing, which [Laurent Bossavit][bossavit-leprechauns] showed rests
on weak secondary evidence. The qualitative claim (early prevention is cheap,
drift compounds) is well-supported; the word "exponentially" is not, absent
abcd's own measurements — better stated as "compound" or "super-linearly."

**Tensions.**
- The "exponential" quantification is the weakest link (see above) — treat as a
  sound qualitative heuristic, not a measured exponential.
- Hard gates from day one have a real cost the principle underweights: during
  genuine exploration, premature hard constraints can ossify a design before the
  team understands the domain — exactly Cunningham's point. Over-rigid vocabulary
  registration / modification grammar can become friction that discourages the
  exploration that finds the right abstractions.
- Shift-left has a documented failure mode at scale: too many blocking checks
  cause gate fatigue and rubber-stamping. A gate stays "cheap" only if it is fast
  *and* low-false-positive; a noisy cheap gate is functionally expensive.
- The LLM "expensive dynamic" tier is non-deterministic and can drift in its own
  judgement, so it lacks the reproducibility that makes traditional cheap gates
  trustworthy as hard gates. It needs its own reliability discipline (fixed
  prompts, sampled audits) — which abcd's `prompt_version` lock, golden-test
  fixtures, and same-chat re-review partially supply.

---

## 6. Recovery humility (Naur theory-building)

*Artefacts are the highest-fidelity recovery floor, not the theory itself — the
theory lives in the people who built it and the alternatives they rejected. The
lifeboat and the Modification Grammar discipline (closing Naur's Modification
axis) are implementations.*

**Canonical prior art.** The primary source is **Peter Naur, "Programming as
Theory Building" (1985)** ([gwern mirror][naur-1985]; *Microprocessing and
Microprogramming* 15(5):253-261): a program is a *theory* held by people, not its
code; it "dies" when the team holding the theory disperses and cannot be revived
from documentation alone. Naur (via Ryle) names three things the theory-holder
can do that code cannot encode — **map** program to world, **justify** why each
part is as it is, **modify** to new demands — the direct origin of abcd's
Mapping/Justification/Modification axis. The epistemological warrant is
**Polanyi's tacit knowledge** ("we know more than we can tell"); the management-
science counterpart is the **SECI model** (Nonaka & Takeuchi, 1995) — externalisation
is real but never complete, and re-internalisation requires socialisation, not
just reading docs. The engineering-practice lineage for capturing rationale is
**architecture knowledge vaporization** ([Jansen & Bosch, WICSA 2005][jansen-bosch],
who *popularise* the term — not coin it) and ADR rationale templates
([Tyree & Akerman, IEEE Software 2005][tyree-akerman]); crucially, ADRs target
*Justification*, not Naur's *Modification* axis — precisely the gap Modification
Grammar targets. The **LLM-era revival** is the live consensus:
[Baldur Bjarnason (2022)][bjarnason-theory] ("the death of a program happens when
the programmer team possessing its theory is dissolved") and
[Sean Goedecke (2026)][goedecke-ai] (the binding constraint for agents is theory
*retention*, not capability — they "build their theory from scratch every time").

**Consensus.** Code and docs are a lossy projection of a richer tacit theory, so
full recovery from artefacts alone is impossible — you recover behaviour and
structure, not live justification-and-modification judgement. The response is not
to pretend the theory serialises, but to externalise the highest-value,
hardest-to-recover slices (decisions, rejected alternatives, invariants,
extension rules) as first-class records to slow vaporization. In the agentic era
the operative constraint is *retention*: agents reconstruct understanding every
session, so persistent curated rationale plus access to the originating session
is the mitigation. The mature position is humility about what artefacts can do,
paired with discipline about capturing what is most likely to vaporize.

**abcd alignment.** Strong, with one honest extension. "Artefacts are the
highest-fidelity recovery floor, not the theory itself" is almost a direct
paraphrase of Naur's death-of-the-program claim and Polanyi's asymmetry; embark's
"hunt the originating session before trusting the lifeboat blindly" is textbook
SECI re-socialisation. **The extension:** abcd makes the contestable structural
claim that Naur's tacit theory decomposes into three *capturable* axes and that
Modification is "the genuinely new gap." Defensible (ADRs really do target
Justification, not extension grammar) but abcd's own synthesis, not Naur's — Naur
lists those three as things the theory-holder can *do*, and would likely resist
the implication that any cleanly serialises. **Notably, abcd's team caught this
reflexively:** the adversarial review that renamed the discipline from "Theory
transparency" to "Modification Grammar" because "theory" overpromised *is*
recovery humility applied to itself — the clearest signal abcd reads Naur
correctly rather than co-opting his authority.

**Tensions.**
- Capture paradox: Naur/Polanyi imply the most valuable theory is constitutively
  inarticulable, yet Modification Grammar asks authors to articulate extension
  rules. The risk is *false confidence* — a well-formed section can read as if
  the theory is preserved when only its articulable shell is. The `MG004`
  strip-the-name boilerplate test mitigates this but cannot detect a section that
  is concrete yet misses the load-bearing tacit invariant.
- Retention-vs-capability is genuinely unresolved, not settled. Goedecke leaves
  open whether agents build real theories; some argue LLMs cannot participate at
  all, others that persistent agent memory will close the retention gap. abcd
  leans on "agents build but cannot retain"; both stronger and weaker readings
  would change how much the recovery floor matters.
- Externalisation cost vs benefit, forever: itd-37 itself flags Modification
  Grammar as abcd's first *expensive* discipline (~15-30 min/spec) with
  boilerplate-rot as its principal failure mode — mirroring the long-standing
  empirical finding that rationale-capture systems are undermined by author
  effort and staleness (the same vaporization literature documents that capture
  systems are frequently abandoned). That the gate pays off is a bet, not yet a
  measured result at corpus scale.

---

## Cross-cutting note: names teach

A seventh thread runs through several principles: **metaphors must earn their
keep.** abcd's maritime naming is applied only when a cognate adds meaning
(`dredge` literally raises settled material; `loot` carries a licence-check
reflex), and neutral verbs (`intent`, `capture`, `grill`) are used otherwise.
This is least-astonishment (Principle 2) applied to vocabulary, and it pairs with
the HARD vocabulary-registration gate (Principle 5): a name is an interface, and
an unregistered or over-stretched name is drift. No separate SOTA pass was run
for this thread; it is recorded here as a corollary, not a standalone principle.

---

## How these six relate

The extraction identified five load-bearing principles (1–5) plus recovery
humility (6). The SOTA pass reinforces the ordering: **Principle 1 (right grain,
right home)** is the spine — SRP-for-prose — from which the artefact taxonomy,
the decision-record split, and "one authoritative home" (Principle 4) all
descend. **Principle 3 (integrate without owning)** is the most externally
validated and the place abcd is arguably *ahead* of common practice (automated
carry-patch retirement). **Principle 5 (cheap hard gates)** is the most
empirically contested — the qualitative core holds, the "exponential"
quantifier does not. **Principle 6** is the philosophical floor under the whole
project, and the one abcd most visibly applies to itself.

## References

[martin-srp]: https://en.wikipedia.org/wiki/Single-responsibility_principle "Single-responsibility principle (Robert C. Martin) — 'one reason to change'; the meta-principle abcd applies to artefacts rather than classes"
[nygard-adr]: https://www.cognitect.com/blog/2011/11/15/documenting-architecture-decisions "Michael Nygard (2011), 'Documenting Architecture Decisions' — canonical ADR essay; Status/Context/Decision/Consequences; supersede-don't-edit lifecycle"
[fowler-adr]: https://martinfowler.com/bliki/ArchitectureDecisionRecord.html "Martin Fowler, bliki 'Architecture Decision Record' — an ADR captures a single decision and is immutable/superseded; credits Nygard and Kruchten"
[diataxis]: https://diataxis.fr/ "Daniele Procida, Diátaxis — four-mode documentation framework split by user need; 'crossing or blurring the boundaries is at the heart of a vast number of problems in documentation'"
[working-backwards]: https://workingbackwards.com/concepts/working-backwards-pr-faq-process/ "'The Amazon Working Backwards PR/FAQ Process' (Bryar & Carr) — press-release-first method as a customer-focus forcing function; companion to the 2021 book 'Working Backwards'"
[adzic-sbe]: https://gojko.net/books/specification-by-example/ "Gojko Adzic, 'Specification by Example' (Manning, 2011; Jolt Award best book of 2012) — living documentation / single-source-of-truth executable specs"
[candost-adrs-rfcs]: https://candost.blog/adrs-rfcs-differences-when-which/ "Candost Dagdeviren, 'ADRs and RFCs: Their Differences and Templates' — community articulation of RFC=explore/prospective vs ADR=commit/retrospective (convention, not a single originator)"
[wikipedia-pola]: https://en.wikipedia.org/wiki/Principle_of_least_astonishment "Principle of least astonishment — origin PL/I Bulletin 1967 (W. N. Holmes); Unix-philosophy lineage; POLA/POLS definition"
[raymond-taoup]: http://www.catb.org/~esr/writings/taoup/html/ch01s06.html "The Art of Unix Programming, Eric S. Raymond (2003) — 'Rule of Least Surprise: In interface design, always do the least surprising thing' (primary source)"
[pep-20]: https://peps.python.org/pep-0020/ "PEP 20 – The Zen of Python by Tim Peters (created 2004; aphorisms posted to python-list 1999) — explicit-over-implicit, refuse-to-guess, errors-never-pass-silently"
[clig-dev]: https://clig.dev/ "Command Line Interface Guidelines, Aanand Prasad, Ben Firshman, Carl Tashian, Eva Parish — safe defaults, --dry-run, tiered destructive confirmation, 'if you change state tell the user', help on no-args"
[git-push]: https://git-scm.com/docs/git-push "git push — official documentation; --force-with-lease as the safe opt-in alternative to --force; safe-by-default remote design"
[bezos-2015]: https://s2.q4cdn.com/299287126/files/doc_financials/annual/2015-Letter-to-Shareholders.PDF "Amazon 2015 Letter to Shareholders, Jeff Bezos — Type 1/Type 2 decisions, one-way vs two-way doors, irreversible vs reversible"
[bezos-2016]: https://www.aboutamazon.com/news/company-news/2016-letter-to-shareholders "Amazon 2016 Letter to Shareholders, Jeff Bezos — 'Many decisions are reversible, two-way doors... use a light-weight process'"
[kustomize-intro]: https://kubernetes.io/blog/2018/05/29/introducing-kustomize-template-free-configuration-customization-for-kubernetes/ "Introducing kustomize — Jeff Regan & Phil Wittrock (Google/Kubernetes SIG-CLI), 2018; canonical base/overlay 'config as data, not a fork' statement"
[debian-debsrc3]: https://wiki.debian.org/Projects/DebSrc3.0 "Debian Wiki: Projects/DebSrc3.0 — the 3.0 (quilt) source format storing downstream changes as a documented patch queue over pristine upstream"
[debian-quilt]: https://wiki.debian.org/UsingQuilt "Debian Wiki: UsingQuilt — quilt push/pop stack model for managing ordered patch sets over an untouched base"
[chromium-upstream]: https://www.chromium.org/chromium-os/chromiumos-design-docs/upstream-first/ "Chromium OS 'Upstream First' policy (Google) — minimise downstream-only code, document every divergence, prefer landing fixes upstream; rejected-but-kept patches labelled with link and reason"
[cockburn-hexagonal]: https://en.wikipedia.org/wiki/Hexagonal_architecture_(software) "Hexagonal architecture (software) — Alistair Cockburn's Ports and Adapters; swappable technology-specific adapters behind ports (mid-1990s origin, named ~2005)"
[fowler-strangler]: https://martinfowler.com/bliki/StranglerFigApplication.html "Martin Fowler, bliki: Strangler Fig Application (2004) — incremental wrap-and-replace over a host system rather than big-bang rewrite/fork"
[aws-rel05]: https://docs.aws.amazon.com/wellarchitected/latest/reliability-pillar/rel_mitigate_interaction_failure_graceful_degradation.html "AWS Well-Architected Reliability Pillar REL05-BP01 — 'transform applicable hard dependencies into soft dependencies' (graceful degradation; page also cautions against generic fallback chains)"
[desaulniers-fork]: https://nickdesaulniers.github.io/blog/2023/02/01/forking-is-not-free-the-hidden-costs/ "Nick Desaulniers, 'Forking is not Free; the hidden costs' (2023) — fork debt; rebasing cost proportional to divergence; eroded intent; slower security response"
[pauldally-json6902]: https://pauldally.medium.com/the-most-common-problem-i-see-when-using-json6902-patching-with-kustomize-1d19a0f4a038 "Paul Dally (2022): common JSON6902 kustomize failure — 'add' silently behaving as 'replace'; the silent-mis-target debuggability critique of overlay patching"
[hockin-level]: https://speakerdeck.com/thockin/edge-vs-level-triggered-logic "Tim Hockin (Kubernetes/Google), 'Edge vs. Level triggered logic' — why Kubernetes controllers are level-triggered and reconcile full state"
[terraform-iac]: https://developer.hashicorp.com/terraform/tutorials/aws-get-started/infrastructure-as-code "HashiCorp Terraform docs, 'What is Infrastructure as Code with Terraform?' — declarative desired end-state, state file as source of truth, plan/apply reconciliation"
[young-cqrs]: http://codebetter.com/gregyoung/2010/02/13/cqrs-and-event-sourcing/ "Greg Young, 'CQRS and Event Sourcing' (2010) — primary-source articulation by the originator of CQRS / practiced event sourcing"
[fowler-cqrs]: https://martinfowler.com/bliki/CQRS.html "Martin Fowler, 'CQRS' bliki — attributes CQRS to Greg Young; warns it should be used only on specific portions of a system and is easy to misuse"
[azure-event-sourcing]: https://learn.microsoft.com/en-us/azure/architecture/patterns/event-sourcing "Microsoft Azure Architecture Center, 'Event Sourcing Pattern' — append-only immutable event store as system of record, materialized read views, 'Versioning events' guidance"
[git-content]: https://blogs.kenokivabe.com/article/git-object-hashing-and-content-addressability "Keno Kivabe, 'Git Object Hashing and Content Addressability' — git objects addressed by SHA of content; immutable, content/lineage-based identity (Torvalds design)"
[w3c-prov]: https://www.w3.org/TR/prov-dm/ "W3C, 'PROV-DM: The PROV Data Model' — Recommendation (30 April 2013); Entity/Activity/Agent core provenance model"
[slsa-in-toto]: https://slsa.dev/blog/2023/05/in-toto-and-slsa "SLSA/in-toto Community, 'in-toto and SLSA' (2023) — verifiable supply-chain provenance via Statement/Predicate/Subject attestations"
[json-schema-contract]: https://norbix.dev/posts/json-schema/ "norbix.dev, 'JSON Schema in Modern Microservices: Contract-First Validation Strategy' — schema/contract as single source of truth driving validation (practitioner blog)"
[boehm-basili]: https://www.cs.cmu.edu/afs/cs/academic/class/17654-f01/www/refs/BB.pdf "Boehm & Basili, 'Software Defect Reduction Top 10 List' (IEEE Computer, Jan 2001) — canonical source for the early-defect-is-cheaper claim"
[boehm-curve]: https://www.researchgate.net/figure/Historical-cost-to-fix-curve-Adapted-from-Boehm-1981-p-40_fig11_308264787 "Reproduction of Boehm's 1981 historical cost-to-fix curve (Software Engineering Economics, p.40)"
[cunningham-debt]: https://cmdev.com/papers/debt-metaphor/ "Transcript of Ward Cunningham's 2009 'debt metaphor' video — it is about evolving understanding, not bad code; origin in WyCash (OOPSLA 1992)"
[broken-windows-td]: https://arxiv.org/abs/2209.01549 "Amanatidis et al., 'The Broken Windows Theory Applies to Technical Debt' (arXiv; journal version in Empirical Software Engineering 2024) — ~458% higher use of non-descriptive names in high-debt code"
[evolutionary-arch]: https://www.oreilly.com/library/view/building-evolutionary-architectures/9781492097532/ch04.html "Ford, Parsons, Kua & Sadalage, Building Evolutionary Architectures 2nd ed. (2023), Ch.4 'Automating Architectural Governance' — fitness functions as automated governance gates"
[minsky-illegal]: https://fsharpforfunandprofit.com/posts/designing-with-types-making-illegal-states-unrepresentable/ "Scott Wlaschin, 'Making illegal states unrepresentable' (F# for Fun and Profit) — credits Yaron Minsky's design-time-constraint principle"
[opa-conftest]: https://github.com/open-policy-agent/conftest "Open Policy Agent / Conftest — policy-as-code engine (Rego) for testing structured config; part of the CNCF-graduated OPA project"
[shift-left]: https://en.wikipedia.org/wiki/Shift-left_testing "Shift-left testing — term coined by Larry Smith (2001, Dr. Dobb's Journal), rooted in Deming's 'build quality in'"
[bossavit-leprechauns]: https://leanpub.com/leprechauns/read "Laurent Bossavit, 'The Leprechauns of Software Engineering' — shows the Boehm exponential cost-of-defect curve rests on weak/secondary evidence"
[naur-1985]: https://gwern.net/doc/cs/algorithm/1985-naur.pdf "Peter Naur (1985), 'Programming as Theory Building', Microprocessing and Microprogramming 15(5):253-261 — canonical primary source (gwern scanned mirror)"
[jansen-bosch]: https://research.rug.nl/en/publications/software-architecture-as-a-set-of-architectural-design-decisions/ "Anton Jansen & Jan Bosch (WICSA 2005), 'Software Architecture as a Set of Architectural Design Decisions' — popularises 'knowledge vaporization' and first-class design decisions"
[tyree-akerman]: https://github.com/joelparkerhenderson/architecture-decision-record/blob/main/locales/en/templates/decision-record-template-by-jeff-tyree-and-art-akerman/index.md "Jeff Tyree & Art Akerman, 'Architecture Decisions: Demystifying Architecture' (IEEE Software 2005) ADR rationale template (third-party transcription) — the practice surface for the Justification axis"
[bjarnason-theory]: https://www.baldurbjarnason.com/2022/theory-building/ "Baldur Bjarnason (2022), 'Theory-building and why employee churn is lethal to software companies' — extract from 'Out of the Software Crisis'; the program dies when the theory-holding team dissolves"
[goedecke-ai]: https://www.seangoedecke.com/programming-with-ai-agents-as-theory-building/ "Sean Goedecke (2026), 'Programming (with AI agents) as theory building' — the binding constraint for agents is theory RETENTION, not capability"

## Related Documentation

- [`related-work.md`](../related-work.md) — prior-art frameworks abcd shares DNA with (PAUL, CARL, Karpathy, Naur, Dell'Acqua); this note extends its Naur and decision-record threads with full SOTA grounding.
- [`_references.md`](../_references.md) — canonical bibliography registry; the prior-art and methodology entries cited here are appended there.
- [`../../brief/01-product/03-mental-model.md`](../../brief/01-product/03-mental-model.md) — the four-layer model Principle 1 grounds, including the Naurian-gap section Principle 6 extends.
- [`../../brief/05-internals/04-universal-patterns.md`](../../brief/05-internals/04-universal-patterns.md) — the MCP-preferred/structural-fallback, plugin-preferred, vendor-agnostic-adapter, JSON-internal/MD-render, and artefact-lifecycle-taxonomy patterns that Principles 3 and 4 implement.
- [`../../decisions/adrs/`](../../decisions/adrs/) — the settled decisions these principles emerged from; a principle here may inform a future ADR.
