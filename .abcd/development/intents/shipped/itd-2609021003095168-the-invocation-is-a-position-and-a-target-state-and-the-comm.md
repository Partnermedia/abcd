---
id: itd-2609021003095168
slug: the-invocation-is-a-position-and-a-target-state-and-the-comm
spec_id: spc-2609021004075744
kind: standalone
suggested_kind: null
reclassification_history: []
builds_on: [itd-183, itd-184, itd-199]
severity: major
impact: fix
origin: researcher-authored
production_mode: dictated-and-formatted
---

# The invocation is a position and a target state, and the committed preset for the position supplies what the reading is handed

Typed links: `builds_on` [itd-183](../shipped/itd-183-the-cold-reading-sees-exactly-what-the-assembler-passes-posi.md) (the assembler), [itd-184](../shipped/itd-184-four-cold-reading-definitions-one-blindness-core-each-positi.md) (the operand pin), [itd-199](../shipped/itd-199-a-reading-is-about-something-narrower-than-everything-its.md) (the presets); `refines` [itd-199](../shipped/itd-199-a-reading-is-about-something-narrower-than-everything-its.md) (its scope operand and override stamp are withdrawn; its presets stay).

## Press Release

> **A reading is commissioned with two things, as the design says: where it stands and what it reads against.** `abcd reading assemble --position <position> --target <commit>` is the whole invocation. What the reading is handed comes from the committed preset for that position, applied by the assembler with no flag; changing it is a commit to the preset file, reviewed and inside the dirty gate. The manifest records the preset entry applied and its hash, so a run is reproducible from the commit it names. There is no scope flag, no override, and no override stamp.

> "I want to run the readings exactly as the design specifies them, and the design says position and target state," said an AI/agent researcher who runs the four positions against their own record. "If I want the detection reading to see more, I change the preset and commit it. Nothing about a run should depend on what I typed at the prompt."

## Why This Matters

The design fixes the invocation at a position and a target state; the reading's object and question come from the definition, and the operator's remaining residue is when to run and what state to point at. itd-199 added a scope operand and an override stamp because the instrument could not be pointed at anything, and adr-58 recorded that departure. The maintainer has ruled that Iteration 2 runs exactly as the design specifies, and adr-2609021016286571 supersedes adr-58 accordingly.

What the scope operand did is not lost. The committed presets already map each position to what it reads, and a preset is a record fact the definition and the repository supply rather than something typed. Applying the position's preset with no operand keeps the calibration lever and removes the channel the design closes.

## What's In Scope

- **The assembler's invocation is `--position` and `--target`**, and a third operand of any kind is refused by name; itd-184's operand pin is updated to the two and fails closed on any addition, as it was designed to.
- **The committed preset for the position is applied by the assembler**, with no operand naming it. The preset file keeps its shape from itd-199 (kinds, records and paths per position, windows per the preset-windows intent); the `cold` and `warm` names retire as invocation tokens, and one entry per position stands. A repository that wants a wider reading commits a wider entry.
- **The manifest records the preset entry applied and its hash**; the override stamp leaves the manifest and the bundle, because nothing can be overridden.
- **No repository path is accepted at the invocation**, as before; the preset file remains the only place a path may be named.
- **The plugin surface page, the brief's reading chapter, the definitions' Object sections and the generated command reference** say two operands and no scope.

## What's Out of Scope

- The presets' content and windows, which the preset-windows intent carries.
- The comparative position's candidate run, which adr-2609021016272867 derives from the record; no operand there either.
- Any change to the blindness core beyond the fourth-condition sentence below, or to the supply regimes.

## Mechanism

We expect applying the committed preset per position with no operand to keep the instrument pointable without the channel the design closes, because the preset is the record's statement of what a position reads and the manifest names which commit's statement a run used. It fails if a run needs material no committed preset names, in which case the remedy is a commit to the preset, and the failure is visible as a dry run whose size report shows the gap.

## Scope Conditions

- The preset file is committed and inside the dirty gate; a run against an uncommitted preset edit refuses, as today. <!-- cond: cond-2609021004076748 -->
- One preset entry per position. A second named preset is not a concept this intent keeps; a repository with two calibrations commits one and records the other in its history. <!-- cond: cond-2609021004074586 -->
- The impact is `fix`: the verb's third operand shipped in v0.7.0 with no user able to hand its output to a reader, and the design the verb exists to serve never admitted the operand. <!-- cond: cond-2609021004073841 -->

## Acceptance Criteria

- **Given** an invocation carrying a `--scope` operand or any operand beyond position and target, **when** it runs, **then** the verb refuses and names the two operands the design admits.
- **Given** an invocation with a position and a target, **when** the assembly runs, **then** the bundle carries what the committed preset for that position names and nothing else, and the manifest records the preset entry applied and its hash.
- **Given** an uncommitted edit to the preset file, **when** an assembly runs, **then** the verb refuses, as it does today.
- **Given** the manifest of any run, **when** it is read, **then** it carries no override stamp and no scope source, and a reader can reproduce the run from the commit and the preset entry it names.
- **Given** itd-184's operand pin, **when** a third operand is added to the assemble verb, **then** the pin fails closed and names it.
- **Given** the plugin page, the brief's reading chapter, the four definitions and the generated command reference, **when** they are read, **then** each states two operands and none mentions a scope operand.
- **Given** the four definitions, **when** their shared blindness core is read, **then** its fourth condition says items are returned in the order they arise in the object, byte-identical across the four, each definition's `prompt_version` has moved PATCH and the agents changelog names the change (correction (7) of the 2026-09-02 ruling; iss-2609021153261145, resolved by this intent).

## Prior Art

- The design framework v4 section 8.2 and ruling M8; the companion v4 section 4.1; adr-2609021016286571, which supersedes adr-58; adr-2609021016272867.
- [itd-199](../shipped/itd-199-a-reading-is-about-something-narrower-than-everything-its.md) (the presets, kept; the operand, withdrawn), [itd-184](../shipped/itd-184-four-cold-reading-definitions-one-blindness-core-each-positi.md) (the operand pin).

## Open Questions

None.

## Audit Notes

<!-- abcd-review: OWED receipt=rcp-2d81906de2e8 -->
Fidelity review OWED (receipt rcp-2d81906de2e8).

## Grounds

- pursued: we expect a two-operand invocation with the committed preset applied per position to run Iteration 2 exactly as the design specifies while keeping the instrument pointable; a run that needs material no committed preset can name, with no way to commit it, would show it wrong
