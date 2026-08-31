---
id: itd-198
slug: an-assembly-reports-what-it-would-cost-before-a-reading-is
spec_id: null
kind: null
suggested_kind: standalone
reclassification_history: []
builds_on: [itd-183]
severity: critical
impact: fix
origin: researcher-authored
production_mode: hand-written
---

# An assembly reports what it would cost before a reading is commissioned

## Press Release

> **The assembler says how large a reading would be, per material kind, before anyone dispatches it.** Every assembly reports bytes and an estimated token count for each kind of material it passed, and the total, whether or not it writes an artefact. Test files are reported apart from other source, because on a repository of any size they are the largest single class and nobody had counted them. No budget is enforced and none is invented: the assembler cannot know what a given reader accepts, and a number it made up would be wrong for someone. It reports; the operator decides.

> "I assembled a reading and got nine point eight megabytes," said an AI/agent researcher who runs cold readings against their own design record. "Every test passed. Every criterion was met. The artefact could not be handed to anything, and there was no way to learn that except by trying."

## Why This Matters

The instrument shipped correct and undeliverable, and the gap that let it was the absence of a number. Measured over this repository at one commit: about 9.8 MB of artefact, roughly 2.3 million estimated tokens of item text, at every position
([iss-2608311501186646](../../../work/issues/open/iss-2608311501186646-the-assembled-input-for-a-real-reading-of-this-repository-is.md)).
Source is 82 per cent of it and test files are 53 per cent of the source; the records a reading is meant to reason about are 9 per cent. Thirteen intents, six adversarial delta reviews and five fidelity audits passed over that, because the evals run against a fixture repository of about thirty files, which is the right corpus for asserting a firewall and the wrong one for learning what an artefact weighs.

This intent adds the number, and the four mechanism changes it needs to be
checkable rather than asserted. It does not make the reading fit — the measured selections are 2.3 million tokens for everything, 1.3 million without tests, and 210 thousand for the records alone, so no selection this intent could offer both fits a reader and preserves any position's stated object. Making a reading fit is [itd-199](itd-199-a-reading-is-about-something-narrower-than-everything-its.md)'s work, and it is a redesign. The number comes first because it is cheap, it is what made the problem visible, and every later decision about what a reading should contain is a decision nobody can take without it.

## What's In Scope

- **A per-kind size report on every assembly**, whether or not an artefact is written, reachable through the existing dry-run path that already renders a result and writes nothing.
- **Bytes and an estimated token count** per material kind and in total, with the estimate labelled as a byte-derived estimate rather than a tokenizer's answer.
- **A `test` material kind**, split from `source`, which requires a suffix form in the include table's match grammar: the grammar today reads an entry beginning with a dot as an extension and anything else as an exact basename, and `_test.go` is neither. A third form needs its own disambiguation rule against the two that exist, and a case rule, since extensions match case-insensitively today and basenames do not.
- **An assembler version bump.** Adding a row moves the include table's rendering by itself, so the digest gate fires either way. The version is named here because the manifest's item shape also changes, and because the shape version the manifest carries covers both artefacts an assembly writes: the intent states which of the two versions moves rather than leaving a reader to infer it.
- **The kind column added to the include table's rendering**, which fixes a LATENT defect rather than one this split creates. The rendering emits positions, source, matches, fields and the admitting rule, and no kind, so today a kind reassignment on an existing row changes every bundle while the version the manifests carry stands still. That is true before this intent and is closed by it.
- **The kind recorded per manifest item**, so the report is checkable against the manifest rather than asserted beside it. Brief invariant 16 requires an attestation to state no more than its examination establishes, and a report the manifest cannot corroborate is exactly that shape.

## What's Out of Scope

- **Making a reading fit.** No selection, budget or refusal. itd-199 carries that.
- **Enforcing a budget.** The assembler cannot know a reader's capacity, so a threshold here would be a guess with a gate attached.
- **A tokenizer.** The estimate is bytes divided by a constant, and says so.
- **Reclassifying test-support packages.** Helper packages that are not `_test.go` files stay `source`; measured, they are three items and about 2,600 estimated tokens, which is not worth a second rule.

## Scope Conditions

- The byte-derived estimate is accurate enough to decide whether something is plausible, not to plan against. If it ever changes a decision it should not have, it is replaced by a real tokenizer rather than tuned.
- Reporting test files apart from source does not narrow any reading. The detection position's own definition names the tests as part of its object: "The shipped tree read against the claim record: the source, the tests, the delivered documentation and the build configuration on one side, and the shipped intents, the specs, the disciplines, the glossary and the brief's current text on the other." So this intent separates them for counting and never for admission, and ac-4 below binds that rather than leaving it to prose.
- The kind carried on a bundle item is a field a reading receives, so reassigning 407 of this repository's items from `source` to `test` DOES change the bundle's bytes. That is the whole of the bundle change, and ac-6 states it rather than claiming the bundle is untouched.
- This intent assumes Go source stays admitted. itd-194 narrows admission to what the exclusion floor can parse and states its parseable set as markdown whose frontmatter it resolves; every percentage here rests on source remaining in the bundle, and if itd-194 lands first that assumption is the one to re-measure.

## Acceptance Criteria

- **Given** an assembly, **when** its result is produced, **then** it carries bytes and an estimated token count for each material kind it passed, and a total.
- **Given** a dry run that writes no artefact, **when** its result is rendered on the CLI and reported through the plugin page, **then** the per-kind figures and the total appear on both surfaces.
- **Given** a report, **when** its token figures are read, **then** each is labelled as a byte-derived estimate rather than a tokenizer's count.
- **Given** each of the four positions, **when** an assembly runs before and after the kind split, **then** the set of admitted repository paths is identical at every position.
- **Given** a repository holding files whose names end in the test suffix, **when** an assembly admits them, **then** those items carry the `test` kind and every other admitted Go file carries `source`.
- **Given** the include table's rendering, **when** the assembler version is derived from it, **then** the rendering includes each row's kind, so a kind reassignment on an existing row moves the version.
- **Given** a manifest, **when** it is decoded strictly, **then** every item carries a kind and round-trips it.
- **Given** an assembled bundle, **when** it is decoded, **then** it carries no size figure and no field the report introduced: the only bundle change this intent makes is the kind label on items reassigned to `test`.

## Grounds

- pursued: This conjecture is pursued now because the absence of a number is what let an unusable artefact pass every gate the workstream built, and the number is cheap where the fix is a redesign. What would show it wrong is a reader that accepts the current artefact, which would make the count uninteresting, or a token estimate so far from a tokenizer's answer that it misleads the decision it exists to inform.
