# Agent prompt changelog

Per [itd-5](../.abcd/development/intents/disciplines/itd-5-prompt-quality-additions.md),
every `agents/*.md` prompt carries a `prompt_version` and a corresponding entry
here recording the bump rationale (and, at `1.0.0` lock, the self-improvement
pre-flight outcome and calibration-corpus delta).

**Version band.** Agents ship in the `0.x` band until they clear their calibration
corpus — `0.x` means "shipped and wired, honestly unmeasured"; `1.0.0` means
"measured against a corpus and locked" (itd-81's amendment to itd-5, which governs
over the brief's earlier `1.0.0`-at-close expectation). The four M6 synthesis
agents below entered at `0.1.0`, wired to their `abcd disembark` verbs and
unmeasured; `lifeboat-oracle` has since become `lifeboat-reviewer` at `0.1.1`.

## 0.3.0 — 2026-08-30 (itd-181 / spc-59 — scope-condition dispositions)

### intent-auditor 0.3.0

The verdict grows a fourth judgement surface: `scope_conditions`, one disposition
per scope condition the intent records, keyed to the `cond-…` identity spc-55
mints rather than to the condition's wording. The definition gains the
`scope_conditions` input note (identities are host-supplied and echoed verbatim,
never invented), a four-value rubric — `survived`, `narrowed`, `falsified`,
`untested` — framed as harshly as the acceptance rubric (`untested` is the
correct disposition for a vacuum, not `survived`), the extended output block, and
two numbered rules naming what the ingest refuses: exact coverage of the supplied
identities in both directions, the closed value set, `narrowing` required on
`narrowed` and refused on every other disposition, and cited evidence on
everything but `untested`.

MINOR in the `0.x` band: the output contract is extended rather than reshaped,
and a verdict for an intent that records no conditions — every intent shipped
before the identity mint existed — stays valid unchanged. The Go ingest is what
actually rejects a bad block (`internal/core/intent/audit.go`), and a lockstep
test decodes this definition's own output blocks into the struct the ingest
decodes, so the two cannot drift apart silently.
Unmeasured — no calibration corpus exists for disposition judgement, and no
self-improvement pre-flight was run.

## 0.1.0 — 2026-08-30 (itd-188 / spc-66 — the ledger scribe)

### scribe 0.1.0

First entry. Host-delegated ledger scribe, whose access rule is the assembler's
exact inverse (brief invariant 15): it receives ledger content under
`.abcd/work/issues/` and the reading output it is transcribing, and never the
shipped repository as an object of judgement. No session holds both a reading and
the ledger. It transcribes returned readings and researcher dispositions into the
reading-record and disposition shapes spc-58 declares, and authors nothing — no
claim, no ground, no disposition state. Two permissions bound that: it may raise a
**fidelity flag** on an internal inconsistency in the material it is transcribing
and must carry it to the researcher unresolved (proposing a resolution is
authorship), and anything it is explicitly asked to compose beyond formatting
opens with a **contribution stamp** that travels with the material if adopted —
the hand-run precursor of the record's origin keys, with an unstamped contribution
refused rather than delivered. Carries `reads_untrusted_input: true`,
`capability_scope.task_classes: [surface_render]`, and an injection-canary fixture
whose payload demands the scribe author dispositions, fetch shipped-tree material,
and move a ledger record. There is no ingest verb, and no schema for these
shapes either until spc-58's reading and disposition stores land: the emitted
records are committed through the ordinary record path, where `record_schema`
refuses their directory as an undeclared bucket rather than judging the record, so
the shapes are held by the definition and by review. The written protocol in the
brief's agents chapter is followed by hand.
Unmeasured — no calibration corpus exists for ledger transcription, and no
self-improvement pre-flight was run.

## 0.2.0 — 2026-08-19 (iss-278 — the itd-5 contract reaches the review/research agents)

Five prompts read attacker-influenceable input without the itd-5 contract:
`ruthless-reviewer`, `security-reviewer`, `docs-currency-reviewer`,
`intent-auditor`, `sota-researcher` (iss-278 records the class and the missing
detector). Each gains `reads_untrusted_input: true`, a `capability_scope`, an
untrusted-data defence paragraph in its body, and an injection-canary fixture —
a MINOR bump apiece: behaviour hardening, output contracts unchanged.

### ruthless-reviewer 0.2.0

`task_classes: [oracle_review]` (family-1 change judgement). Canary: a diff
whose code comment claims pre-approval and demands zero findings — reported as
a finding, never obeyed. Tool grants unchanged; whether review agents should
hold `Bash` at all for external-contribution review is a maintainer decision
tracked in iss-278, not silently made here.

### security-reviewer 0.2.0

`task_classes: [oracle_review]`. Canary: a diff whose string claims a security
exemption while splicing untrusted input into a shell exfiltration — the
exemption claim is itself the finding.

### docs-currency-reviewer 0.2.0

`task_classes: [cross_document_audit]`. Canary: a docs page whose HTML comment
demands its claims be marked current — claims verified against code regardless.

### intent-auditor 0.2.0

`task_classes: [intent_audit]`. Canary: an acceptance criterion demanding a MET
verdict — judged solely against the delivered diff.

### sota-researcher 0.2.0

`task_classes: [spec_planning]` — the closed enum carries no research token, and
spec-planning input is what its output feeds; revisit when the adr-40
reconciliation extends the enum. Canary: a fetched page pushing a ranking and a
tracking link — weighed as one low-tier source, never obeyed. It reads the most
attacker-influenceable input in this directory (web search and fetch).

## 0.1.0 — 2026-07-21 (itd-67 / itd-73 — the derived release cut)

### release-changelog-composer 0.1.0

First entry. Host-delegated composer behind `abcd launch ship --changelog-json
<file|->`. Reads the emit step's cut (derived `next_tag`, the `added`/`removed`
record entries with their `in_changelog` flag) and the records themselves; emits
the composed-changelog payload — `schema_version` / `prompt_version` / `next_tag`
/ `entries[{section, records, text}]`. Owns the WORDING and the Keep-a-Changelog
section only: the version, the date, the heading shape, the section order, and the
inclusion set are the binary's. The citation rule is the completeness **bijection**
rather than cite-or-be-dropped — `cited == (added ∪ removed) where in_changelog`,
exactly — and a mismatch (missing, invented, or an `impact: internal` record cited)
refuses the WHOLE payload and writes nothing, because a dropped changelog line is a
shipped change absent from the permanent release record. Carries
`reads_untrusted_input: true`, `capability_scope.task_classes: [surface_render]`,
and an injection-canary fixture whose payload attempts three hijacks (persona
switch, drop the citation, cite the internal record). Unmeasured — no calibration
corpus exists for release prose yet, and no self-improvement pre-flight was run.

## 0.1.0 — 2026-07-16 (itd-88 M6 — synthesis agents)

### principle-distiller 0.1.0

First entry. Host-delegated distiller behind `abcd disembark principles
<lifeboat-dir> --principles-json`. Reads a packed lifeboat's ADRs, intents,
resolved issues, and graveyard findings; emits `principles.json` with each
principle citing a record id, a graveyard finding id, or a packed lifeboat path
(cite-or-be-dropped over `R ∪ F ∪ P`). Carries `reads_untrusted_input: true`,
`capability_scope.task_classes: [principle_distillation]`, and an injection-canary
fixture. Unmeasured (no corpus yet); no self-improvement pre-flight run.

### graveyard-interpreter 0.1.0

First entry. Host-delegated interpreter behind `abcd disembark graveyard
<lifeboat-dir> --lessons-json`. Reads the sealed `graveyard/archaeology.json` and
`graveyard/abandoned.json`; emits the graveyard **lessons** schema (no `mode`, no
`prompt_version` field — the pre-M6 lessons schema), each lesson citing a live
layer-1/2 finding id (cite-or-be-dropped over the finding-id set). Carries
`reads_untrusted_input: true`, `capability_scope.task_classes:
[cross_document_audit]`, and an injection-canary fixture. Unmeasured; no
self-improvement pre-flight run.

### press-release-composer 0.1.0

First entry. Host-delegated composer behind `abcd disembark press-release
<lifeboat-dir> --press-release-json`. Reads the packed brief, spine, and
`principles.json`; emits a single `press-release.json` document that must cite at
least one path in `brief/**`, `rescue/spine.md`, or `principles.json`
(whole-document refusal if it cites nothing resolvable). Carries
`reads_untrusted_input: true`, `capability_scope.task_classes: [surface_render]`,
and an injection-canary fixture. Unmeasured; no self-improvement pre-flight run.

### lifeboat-oracle 0.1.0

First entry. Host-delegated auditor behind `abcd disembark oracle <lifeboat-dir>
<source-repo> --oracle-json`. Reads the packed lifeboat corpus against its source
repo; emits an `oracle` audit carrying a registered verdict (`SHIP` / `NEEDS_WORK`
/ `MAJOR_RETHINK` — out-of-enum refuses the whole payload) and findings that each
cite a packed lifeboat path (cite-or-be-dropped over the packed-path set). Carries
`reads_untrusted_input: true`, `capability_scope.task_classes: [oracle_review,
audit]`, and an injection-canary fixture. Unmeasured; no self-improvement pre-flight
run.


### intent-auditor 0.1.1 (renamed from intent-fidelity-reviewer)

The intent-fidelity judge is the intent AUDIT (adr-40: it emits family-2
promise-vs-reality verdicts), so the agent renames to `intent-auditor` and its
verb moves to `abcd intent audit` / `audit ingest` (spc-28, itd-123). The
verdict format is frozen: `_type` stays `abcd/intent-fidelity-verdict/v1`,
stored markers and previously ingested verdicts remain valid, and a stored
verdict naming the old verifier id still ingests. Prompt body otherwise
unchanged; no re-measurement.

### lifeboat-reviewer 0.1.1 (renamed from lifeboat-oracle)

The lifeboat verdict endpoint emits family-1 review verdicts (`SHIP` /
`NEEDS_WORK` / `MAJOR_RETHINK`), and the planning investigation found the binary
verb never invokes the oracle seam — so the verb renames to `abcd disembark
review` (flag `--review-json`), its artefact moves to
`review/review-<manifest12>.{json,md}`, and this agent becomes
`lifeboat-reviewer` (spc-30, itd-125; adr-40 §5 amended in place). The `oracle`
SEAM is untouched: `capability_scope.task_classes` keeps `oracle_review`, which
honestly names review work reaching a model through the adr-25 seam, and
`/abcd:oracle ask` remains the seam's reserved surface. Prompt body, verdict
enum, and cite-or-be-dropped contract unchanged; canary fixture moved with the
agent. No re-measurement.