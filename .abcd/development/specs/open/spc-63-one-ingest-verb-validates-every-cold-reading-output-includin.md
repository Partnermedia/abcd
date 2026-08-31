---
id: spc-63
slug: one-ingest-verb-validates-every-cold-reading-output-includin
intent: itd-185
---

# The cold-reading output contract: a strict per-position schema, a supply-regime gate, and a staged ingest that leaves evidence or nothing

## Bundle

itd-185,
[itd-183](../../intents/shipped/itd-183-the-cold-reading-sees-exactly-what-the-assembler-passes-posi.md)
and
[itd-184](../../intents/planned/itd-184-four-cold-reading-definitions-one-blindness-core-each-positi.md)
are one design under one bundle kind, and the ceremony cannot give them one
spec: a spec's `intent:` is a single id, captured as iss-2608300108376943.

| Spec | Component it owns |
| --- | --- |
| [spc-61](../closed/spc-61-the-cold-reading-sees-exactly-what-the-assembler-passes-posi.md) | The input assembler, the include table, the pathless bundle, the manifest, and the bundle's shared decisions |
| [spc-62](spc-62-four-cold-reading-definitions-one-blindness-core-each-positi.md) | The four reading definitions under `agents/` and the blindness-core byte-identity test |
| spc-63 (this record) | The output contract, the supply-regime gate, and the ingest sub-verb |

The package name, the verb tree, the run-identifier form and the artefact
layout are the bundle's shared decisions, stated once in
[spc-61 § The package, the verb tree and the artefact layout](../closed/spc-61-the-cold-reading-sees-exactly-what-the-assembler-passes-posi.md#the-package-the-verb-tree-and-the-artefact-layout-shared-bundle-decisions).
This spec uses them and does not restate them.

## Summary

spc-63 delivers itd-185's contract: a reading that quietly exceeds its licence
is refused at ingest, and the refusal names the offending field or item. The
verb is `abcd reading ingest --output-json <path>` in the same
`internal/core/reading` package as the assembler, built on the
output-contract idiom the repository already carries (agent emits JSON, a
deterministic verb validates it, the verb writes the record):
`intent audit ingest --verdict-json`
([`internal/core/intent/audit.go`](../../../../internal/core/intent/audit.go),
whose `validateVerdict` already runs `DisallowUnknownFields`) and
`launch ship --changelog-json`.

Three properties distinguish this contract from a structural schema check.
Ids are **minted by the verb**, never self-supplied. The **supply regime** is
resolved from the definition through the run's position, so no operator input
can set it. And **provenance is enforced for every regime**: an item with an
empty pattern field is refused, which the definitions instruct and nothing
else checks.

The policy is settled by rulings (4), (5), (8), (12) and (18) of 2026-08-28
and is not this spec's to reopen. This spec settles the payload schema, the
reserved-name tables, the signature registry, the staging protocol and the
refusal-record shape.

## Scope

**In.** The output payload schema and its per-position bodies; the regime gate
and its signature registry; id minting; the staged write protocol and its
orphan sweep; the manifest-reference check; refusal records; the
`abcd reading ingest` sub-verb on both planes.

**Out.** The definitions that state each regime (spc-62); the assembler and
the manifest (spc-61); the reading-record and disposition record *schemas*,
which belong to
[itd-180](../../intents/shipped/itd-180-a-cold-reading-s-findings-land-as-reading-records-and-the-re.md)
and spc-58 (this verb validates against them and writes them, it does not
define them); admission, which is
[itd-189](../../intents/shipped/itd-189-what-the-widening-reading-proposes-is-admitted-or-declined-o.md)'s.

## Approach

### The payload

One JSON document per run, read behind `fsutil.ReadGuarded` with a byte cap,
decoded with `DisallowUnknownFields` at every level, so every violation names
a field rather than guessing at one:

```
_type          = "abcd.reading.output/1"
run_id         = rdg-<...>          (must match a parked assemble run)
position       = widening | entailment | comparative | detection
regime         = generative | explicative | evaluative | registrative
manifest_sha256
instrument     = {model, definition_sha256, assembler_version}
items[]        = {pattern_named, body}
```

**Ids are the verb's to mint.** The payload carries no item identifier; a
reading holds no mint. On acceptance the verb mints `rdi-<yymmddHHMMSS><rrrr>`
per item through `recordid.Minter.Mint("rdi")`
([adr-45](../../decisions/adrs/0045-record-ids-are-timestamp-numeric-and-capture-stable.md)),
run-scoped and never content-derived, so a re-raise of the same finding in a
later run is distinguishable from its first appearance.

**Provenance is one envelope field.** itd-180 calls it the pattern named and
itd-185 calls it the pattern-basis field; they are one requirement, and the
field is `pattern_named` on the item envelope, never in a body (ruling (18)).
Empty or absent refuses the item, at every regime without exception.

### The bodies, per position

Item bodies are position-typed, and the schema for the wrong position's body
is not merely unusual but undecodable. JSON keys are US English, as code-side
text is throughout the repository:

| Regime | Body fields |
| --- | --- |
| `generative` | `configuration`, `what_admits_it` |
| `explicative` | `claim`, `claim_type` (`criterion` / `causal` / `context`), `what_implies_it` |
| `evaluative` | `candidate`, `criterion`, `characterization` |
| `registrative` | `tension`, `constraint_in_play`, `why` |

### The regime gate

**The regime's source of truth is the definition.** The verb resolves the run's
position to `agents/cold-reading-<position>.md`, reads its `regime:`
frontmatter key (spc-62), and compares. A payload whose self-declared regime
disagrees is refused at list level. There is no `--regime` flag and no
configuration key: the value cannot be reached from operator input.

Two enforcement layers sit behind that:

- **Reserved names.** Strict decoding already refuses an unknown field, but a
  generic "unknown field" message is a poor account of a licence breach. Each
  position therefore declares a reserved-name table (`evaluative`: `rank`,
  `score`, `recommended`, `order`; `registrative`: `resolution`, `fix`,
  `remedy`; `explicative`: `disposition`, `status`), and a payload naming one
  is refused with the field named and the licence stated. Arrangement order is
  never refused: items arrive in document order by mandate.
- **Semantic signatures.** Prose that ranks, settles or proposes without the
  field is checked too, through a registry of named detectors over body text
  (`RG-EVAL-ORDERING`, `RG-EVAL-RECOMMENDATION`, `RG-REG-FIXPROPOSAL`,
  `RG-EXPL-DISPOSITION`). `generative` carries no regime-specific refusal; the
  widening prohibitions raise review flags on the run record instead, because
  the generative licence is the widest and the constraint falls at admission.

**Every signature ships enforced.** Each registry entry carries a literal mode
(`enforce` or `flag`) in Go, with no configuration seam: degrading a signature
on observed noise is a code change plus a decision-log entry, which is what
makes it a recorded weakening of the claimed property from enforced to
observed rather than a quiet runtime toggle. Whether the signatures lint
cleanly in practice is the recorded open question, and the degradation path
exists precisely because of it.

### Manifest reference

`manifest_sha256` is the content hash of the assembler's manifest, the one
unforgeable reference. Ingest resolves the parked run at
`.abcd/.work.local/scratch/reading-runs/<run_id>/manifest.json`, hashes it,
and refuses the run when nothing resolves or the hashes disagree. Only after
acceptance is the manifest promoted to
`.abcd/development/readings/<run-id>/manifest.json`.

**Instrument identity** (ruling (12)) is `model` plus `definition_sha256`
plus `assembler_version`, all three required and all three carried into run
metadata, so two runs claiming the same instrument are provably the same. The
definition hash is recomputed here from the definition file and compared with
the payload's claim; the assembler version is compared with the manifest's.

### Staged writes, run metadata last

Nothing durable is written until the whole payload validates.

1. Validate: schema, regime, provenance, manifest reference, instrument.
2. Stage every reading record into
   `.abcd/.work.local/scratch/reading-ingest/<run-id>/`.
3. Move the staged records into the reading-record family (spc-58's
   `.abcd/work/issues/readings/<run-id>/`).
4. Write `.abcd/development/readings/<run-id>/run.json` **last**: the run
   metadata is the commit marker, so a run without one never happened.

An orphaned stage found on a later invocation is reported by name and cleared.
A crash mid-ingest therefore leaves evidence, never half a run.

### Refusal granularity and refusal records

An item-level violation refuses that item and lands the rest, naming the
refused item's ordinal and the rule. A list-level violation (bad `_type`,
regime mismatch, unresolvable manifest, missing instrument field) refuses the
whole run. **A refused run still writes a durable refusal record**:
`.abcd/development/readings/<run-id>/refusal.json`, carrying the run metadata
and the named reason and no items. The event is durable, and a rerun is a new
run with a new run id, never an amendment.

## Acceptance criteria mapping

The criteria were split on 2026-08-31, before this spec was built, so that no
criterion conjoins a structural half a gate holds with a semantic half bounded
by a registry. The numbering below is the positional authority ac-1..ac-13.

| itd-185 criterion | How spc-63 satisfies it | Test |
| --- | --- | --- |
| ac-1 — malformed output refused, nothing durable anywhere | Validation is step 1 of four; staging is local-tier; the durable move happens only after the whole payload validates | `TestMalformedPayloadWritesNothing`, `TestNoDurableWriteBeforeValidation`, `TestUnknownFieldRefusedAtEveryLevel` |
| ac-2 — a fault between staging and the commit marker leaves no half-run, and the orphan is named and cleared | Run metadata is written last as the commit marker; an orphaned stage found on a later invocation is reported by name and cleared | `TestRunMetadataLandsLast`, `TestOrphanedStageIsReportedAndCleared` |
| ac-3 — the manifest reference resolves, and a mismatch refuses the run | `manifest_sha256` is checked against the parked manifest's own content hash before promotion | `TestManifestReferenceMustResolve`, `TestManifestHashMismatchRefusesRun` |
| ac-4 — a registrative reserved name refuses, naming ordinal, field and licence | The `registrative` reserved-name table: `resolution`, `fix`, `remedy` | `TestRegistrativeResolutionFieldRefused` |
| ac-5 — a registered fix-proposal signature refuses, naming item and signature | `RG-REG-FIXPROPOSAL`, shipped in `enforce` mode with no configuration seam | `TestRegistrativeProseFixProposalRefused`, `TestEverySignatureShipsEnforced` |
| ac-6 — an evaluative reserved name refuses, naming the field | The `evaluative` reserved-name table: `rank`, `score`, `recommended`, `order` | `TestEvaluativeRankScoreRecommendedRefused` |
| ac-7 — arrangement order alone is accepted | Arrangement order is never inspected: items arrive in document order by mandate | `TestEvaluativeDocumentOrderIsNeverRefused` |
| ac-8 — an explicative disposition-bearing field refuses, naming the field | `disposition` and `status` are reserved on the explicative body, and strict decoding refuses any field outside that body schema, so the violation is impossible to express rather than merely caught | `TestExplicativeDispositionRefused`, `TestWrongPositionBodyIsUndecodable` |
| ac-9 — a registered disposition signature refuses, naming item and signature | `RG-EXPL-DISPOSITION`, shipped in `enforce` mode | `TestExplicativeProseDispositionRefused`, `TestEverySignatureShipsEnforced` |
| ac-10 — a list-level refusal writes a refusal record and no items | The refusal path writes `refusal.json` carrying run metadata and the named reason; nothing is ever moved out of the stage | `TestListLevelRefusalWritesRefusalRecordOnly` |
| ac-11 — an empty or absent `pattern_named` refuses the item at every regime | Provenance is one envelope field, checked before the body, at all four regimes without exception | `TestEmptyPatternNamedRefusesItemAtEveryRegime` |
| ac-12 — a self-declared regime disagreeing with the definition refuses the run | The regime's source of truth is the definition, resolved through the run's position; the payload's claim is compared, never trusted | `TestRegimeComesFromTheDefinitionNotThePayload`, `TestSelfDeclaredRegimeMismatchRefusesRun` |
| ac-13 — item ids are minted by the verb, and a supplied id is an unknown field | `recordid.Minter.Mint("rdi")` on acceptance; the payload schema carries no item identifier at all | `TestItemIDsAreMintedByTheVerb` |

ac-5 and ac-9 are the two criteria bounded by the signature registry rather
than by the schema, and itd-185 discloses that residue. Their structural
halves, ac-4, ac-6 and ac-8, are unbounded in the other direction: the field is
present or it is not.

Two behaviours this spec delivers carry no criterion of their own, and are
recorded here rather than left to be discovered at the audit: item-level
refusal granularity, which lands the surviving items when one item is refused
(`TestItemLevelViolationLandsTheRest`), and the generative position's
review-flag path, which raises a flag on the run record instead of refusing
(`TestGenerativeHasNoRegimeRefusalButFlagsRecommendation`).


## Tests

Each case is written to fail before the change and pass after, in
`internal/core/reading/` unless named otherwise.

- `ingest_schema_test.go`: `TestMalformedPayloadWritesNothing`,
  `TestUnknownFieldRefusedAtEveryLevel`,
  `TestWrongPositionBodyIsUndecodable`,
  `TestNoDurableWriteBeforeValidation`.
- `ingest_regime_test.go`: `TestRegimeComesFromTheDefinitionNotThePayload`,
  `TestSelfDeclaredRegimeMismatchRefusesRun`,
  `TestEvaluativeRankScoreRecommendedRefused`,
  `TestEvaluativeDocumentOrderIsNeverRefused`,
  `TestRegistrativeResolutionFieldRefused`,
  `TestRegistrativeProseFixProposalRefused`,
  `TestExplicativeDispositionRefused`,
  `TestExplicativeProseDispositionRefused`,
  `TestGenerativeHasNoRegimeRefusalButFlagsRecommendation`,
  `TestEverySignatureShipsEnforced` (a property over the registry).
- `ingest_provenance_test.go`: `TestEmptyPatternNamedRefusesItemAtEveryRegime`.
- `ingest_stage_test.go`: `TestRunMetadataLandsLast`,
  `TestOrphanedStageIsReportedAndCleared`,
  `TestItemLevelViolationLandsTheRest`,
  `TestListLevelRefusalWritesRefusalRecordOnly`.
- `ingest_identity_test.go`: `TestManifestReferenceMustResolve`,
  `TestManifestHashMismatchRefusesRun`,
  `TestInstrumentIdentityRequiresAllThreeParts`,
  `TestItemIDsAreMintedByTheVerb` (injected clock and entropy; a
  payload-supplied id is an unknown field).
- `internal/surface/cli/reading_surface_test.go`:
  `TestIngestRequiresOutputJSON`, `TestIngestReachesBothPlanes`. The
  no-operator-surface guard on the regime is spc-62's
  `TestNoOperatorSurfaceSetsARegime`, in its own file.

## Out of scope

- **Running a reading.** The instrument ships unrun for the whole cycle: this
  verb is exercised against fixture payloads only, and no reading is
  commissioned by this delivery.
- Whether the semantic signatures lint cleanly in practice. Untested, recorded
  as the open question, and the reason the degradation path is reserved.
- Teaching the issue-resolution gates about the new reading-record folders,
  and the reading-record schema itself: spc-58's.
- The standing tension with the repository's widen-options promotion clause
  ("calibrated before it gates") is recorded, not resolved here; the ruled
  design governs the instrument meanwhile.
