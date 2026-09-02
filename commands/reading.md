---
name: reading
description: Assemble the input a cold reading is handed and validate the output it returns, by invoking the abcd binary. Bare invocation is a read-only status render; assemble produces the assembled input and its hashed manifest, and ingest validates one reading's output and writes its records.
argument-hint: "[] | assemble --position <widening|entailment|detection> --target <HEAD|sha> --scope <itd-N|spc-N|kind|preset> [--out <dir>] [--dry-run] | ingest --reading-json <path>"
---

# `/abcd:reading` — cold-reading input assembler

Blindness is a property of the input, not a promise the reader makes. A
positive include table names what may travel; fields are projected out of
records rather than files copied whole; and every run emits a manifest naming
what was passed, by path and field, hashed, so a reader can judge contamination
rather than accept a disclosure on trust. Bare invocation **performs zero
writes**.

Two things this surface does not do. It never runs a reading: it produces the
input a reading would be given, and dispatching that input to a reader is host
work. And it carries no free text at any position — the operator supplies a
position, a target state and a scope, each in a closed grammar, and the
reading's object and question come from its definition, so there is no channel
through which ledger content can travel in the framing of a request.

## Status (bare)

To render the assembler's state:

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" reading --json
```

Summarise the JSON for the user: `assembler_version`, `include_rows` and
`exclusion_rows` (what the table admits and what it refuses), `piles` (which
pile each position assembles from), `definitions`
(the reading definitions present under `agents/`), `staged_runs` (runs an
assembly has parked in the local tier), `orphaned_ingests`, and
`leftover_stages`. Zero writes.

**`piles` says which positions share one assembly and which have their own.**
Each row carries `position`, `pile` (`shared` or `own`), `rows`, `hash` and,
for an own pile, the `rule` saying why that position is handed its own object.
The positions share one assembly by default, so the ordinary answer is `shared`
at all four. Report a position on its own pile explicitly, with its rule: a
reading assembled from a narrower pile is not comparable with the others in the
way a shared assembly is.

**`orphaned_ingests` is not routine.** Each name is a run whose ingest reached
the ledger and never reached its commit marker, so its reading records are
sitting in the committed ledger for a run that never happened. Report it
whenever it is non-empty, and say that the next ingest that validates rolls
those records back.

**`leftover_stages` is the other thing a stage can be, and it is not an
orphan.** Each name is a run that DID commit — its `run.json` is down — and
whose stage merely failed to clear afterwards. Its records stay. Report it,
and say that the next ingest that validates clears the stage alone; never say
its records will be rolled back, because they will not be.

## Assemble one reading's input

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" reading assemble \
  --position widening --target HEAD --scope cold --json
```

`--position` takes one of four closed tokens — `widening`, `entailment`,
`comparative`, `detection`. An unknown token is refused by name. **The
comparative position does not assemble** and refuses, naming the channel it
lacks: its object is the widening reading's pre-admission output, which is not
repository material. `--target`
takes `HEAD` or a hexadecimal commit sha of 7 to 40 digits; a branch name or a
tag is refused, because it moves and the manifest's re-runnability rests on a
reference that cannot.

`--scope` names what the reading is **about**, and is required. It takes one
token in a closed grammar: a record id (`itd-N`, `spc-N`), a
material kind, or the name of a preset committed in
`.abcd/config/reading-presets.json`. **No repository path is accepted here** —
a path may be named only inside the committed preset file, where it is
reviewed, shape-validated and inside the dirty gate (adr-58). A scope
intersects what the position already admits and can only narrow it.

Naming a preset is running as reviewed. Naming a record or a kind directly is
an override, and the manifest stamps it as one, so drift between the committed
presets and what people actually run is countable.

**The comparative position does not assemble.** Its object is the widening
reading's pre-admission output, which is not repository material and has no
channel today. It refuses and names that, rather than returning the detection
position's corpus and reporting success. A per-position pile does not change
that: the admitted output it would name sits inside a record family the include
table may not reach, so the charter carries that entry as an example of the
shape and not as a declaration.

Assembly reads the working tree, so it refuses unless HEAD resolves to the
target **and** no included path is uncommitted. The preset configuration is in
that dirty set, for the same reason the record configuration is: an
uncommitted edit to it reshapes the assembly. Both refusals exit 2, as does
an unknown position, a missing operand, and any positional argument.

Report from the JSON: `run_id`, `position`, `target_commit`, `item_count`,
`manifest_hash`, and — where the run wrote — `out_dir` and `artefacts`.

Report `pile` too: `pile.source` is `shared` or `own`, and `pile.hash` is the
hash of the rows the run assembled from. Say plainly when a run drew from a
position's own pile, because a reading narrowed to its own object is not
comparable with the others in the way a shared assembly is. The written
manifest carries the same stamp under `pile`.

Report `scope` too: `scope.source`, the token the operator gave; `scope.selectors`,
what it resolved to; and `scope.overridden` — true when the run named a record or
a kind directly rather than a committed preset. (The written manifest spells that
last one `scope_overridden`; the verb's own JSON nests it under `scope`.) Say plainly when a run was overridden, because
a run nobody can tell departed from the reviewed presets is a run whose drift
is invisible.

Also report `size`, on every run including a dry run: the total `bytes` and
`tokens_est`, and each row of `by_kind` (`kind`, `items`, `bytes`,
`tokens_est`). Report `tokens_est` as an estimate and say so, quoting the
report's own `basis` — it is bytes over a measured constant, not a tokenizer's
count, and it mis-states each kind by a few per cent in directions spc-68
records. There is no budget and no threshold: the assembler cannot know what a
given reader accepts, so it reports the weight and the operator decides whether
to dispatch it.

### Where the artefacts land

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" reading assemble \
  --position entailment --target HEAD --scope cold \
  --out .abcd/.work.local/scratch/reading-runs/manual --json
```

With `--out`, the assembled input (`bundle.json`) and the manifest
(`manifest.json`) are written into that directory as two separate files.
Without it, they land in the local-tier run directory
`.abcd/.work.local/scratch/reading-runs/<run-id>/`. With `--dry-run` and no
`--out`, nothing is written anywhere and the result is rendered only.

An output directory the include table can reach is refused, and the refusal
names the item that would be admitted. Writing a run where the table reaches it
commits the next run's contamination: the artefacts land as ordinary files, a
later commit puts them in the tree, and the instrument reads its own output.
Write outside the repository, or under the local tier. Both artefacts are also
refused as INPUT wherever they are found, by their `_type` tag, so a run
committed before this was true cannot ride in either.

`--out` must name an empty or absent directory: one run's artefacts are one
run's evidence, and dropping them beside another run's leaves a directory whose
manifest describes half of what is in it. Both files are written through a
temporary name and renamed into place, so a reader never opens a half-written
bundle.

### The host obligation this binary cannot discharge

The assembled input carries no repository path: each item is an ordinal key, a
material class and its text, and only the manifest maps a key back to a path
and a field. That is the half of the isolation the binary enforces.

The other half is yours. When you dispatch an assembled input to a reader,
grant that reader **no repository access** — no file tools, no path, no working
directory. Hand it the bundle's items and nothing else. A reader that can open
the repository is instructed blindness with extra steps, and no manifest can
detect it after the fact.

## Ingest one reading's output

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" reading ingest \
  --reading-json ./reading-output.json --json
```

`--reading-json` names the JSON the reading returned. It is the only operand:
the output states its own run, position and regime, and there is no flag that
could set one. A missing operand, a positional argument, and every refusal
below exit 2.

### What the output carries

One JSON document per run. The envelope names the run the assembly parked
(`run_id`), the position it read at, the regime it claims, the content hash of
that run's manifest, and the instrument — the model, the definition's content
hash and the assembler version, all three required. Each item is a flat object
carrying `pattern`, the pattern the reading read under, and exactly the body
fields its position declares:

| Position | Body fields |
| --- | --- |
| `widening` | `configuration`, `what_admits_it` |
| `entailment` | `claim_surfaced`, `claim_type`, `what_implies_it` |
| `comparative` | `candidate_id`, `criterion`, `characterisation` |
| `detection` | `tension`, `constraint_in_play`, `why_a_tension` |

**An item carries no identifier.** Identifiers are minted by the verb, and a
payload supplying its own is refused as an unknown field. Unknown fields are
refused at every level, so every violation names a field rather than guessing
at one.

### What is refused, and how far

An **item-level** violation refuses that item and lands the rest: an empty or
absent `pattern` at any position, a field the position's body does not declare,
or a reserved name carried as one of the item's own fields. A registered
signature does not refuse — it flags. A **list-level**
violation refuses the whole run: a wrong `_type`, a run id that resolves to no
parked manifest, a manifest hash that disagrees, an instrument claiming a
definition hash or an assembler version the artefacts do not carry, a regime
disagreeing with the definition, or a payload in which no item survived.

**A refusal leaves a record once the run's identity is proven** — that is, once
the run id resolves to a parked manifest whose content hash matches. From that
point every list-level refusal writes `refusal.json` under the run's directory,
including a definition that does not resolve, whose record states no regime
because there is none to state,
carrying the run metadata and the named reason and no items, and the refusal
message and the JSON render both name it. A refusal reached BEFORE that point —
a wrong `_type`, a run id that resolves to nothing, a manifest hash that
disagrees — writes nothing durable anywhere, because there is no proven run to
record against.

**A rerun is a new run with a new run id, never an amendment.** Once a run id
has an outcome — a commit marker or a refusal record — ingesting it again is
refused. Assemble again, and ingest the run that assembly parked.

### The supply regime is the definition's

Each position's definition states the regime the reading reads under, and the
verb reads it from there. An output whose self-declared regime disagrees is
refused. No operand and no configuration key sets a regime, by design.

Per regime, the reserved names — an item carrying one as a field of its own is
refused with the licence stated. The table is read at the run's own regime, one
row per regime:

- `evaluative`: `order`, `rank`, `recommended`, `score`. Arrangement order is
  never refused; items arrive in document order by mandate.
- `registrative`: `fix`, `remedy`, `resolution`.
- `explicative`: `disposition`, `status`.
- `generative` has no reserved names. Its licence is the widest, and the
  constraint on it falls at admission, so it runs the WHOLE registry as review
  flags rather than only its own regime's.

Prose that ranks, settles or proposes without the field is watched too, by a
registry of named signatures (`RG-EVAL-ORDERING`, `RG-EVAL-RECOMMENDATION`,
`RG-REG-FIXPROPOSAL`, `RG-EXPL-DISPOSITION`) reading every text value the item
carries, `pattern` included. All four are **observed, not enforcing**: a hit raises a review flag on the run record naming the item and
the signature id, and the item lands. They cannot tell a reading that proposes
from one reporting that the document proposes, so an enforcing registry refused
a reading for quoting its own material. Report `review_flags` and read them; the
structural halves above are the ones that refuse.

### Where the records land

No OTHER run's durable state is written to or deleted from until the whole
payload validates; a refusal after the run is proven writes its refusal record
and nothing else, and before its identity is proven a run writes nothing durable
anywhere. Once the payload validates the reading records land in the
reading-record family as one batch,
the run's manifest is promoted beside its run metadata, and the run metadata is
written **last** as the commit marker — a run without one never happened.

An ingest interrupted before that marker leaves a stage in the local tier.
Every later invocation names that orphan; the next one whose payload validates
sweeps it: it **rolls that run's reading records out of the committed ledger** —
the run never happened, so it must leave none — and clears the stage. Until
then the bare verb reports it as `orphaned_ingests`. A stage left behind AFTER
the marker — the commit path could not clear it — is reported as
`leftover_stages` instead: that run is complete, and the sweep clears the stage
and leaves its records alone. A refused run destroys no other run's records and
reports the orphans it left in place; the one rollback a refusal does perform is
of its OWN earlier crashed attempt, because a refused run leaves no reading
records. The ids a sweep removed are reported however the invocation ends. One
ingest runs at a time in a checkout: a second waits, and reports contention
rather than sweeping the first one's records away.

Report from the JSON: `run_id`, `records`, `refused_items`, `review_flags`,
`cleared_stages`, `rolled_back_records`, `pending_stages`, and `run_record` —
or, on a refusal that recorded one, `refusal_record`. A refusal renders the
JSON whenever it has one of these to disclose, so read it on exit 2 as well.

**Binary resolution.** Run `"${CLAUDE_PLUGIN_ROOT}/abcd"` — a plugin install
provisions the binary into the plugin root, so this is the rung that fires for a
plugin user. If that path does not exist, try `abcd` on `PATH`; if that fails
too, you are in a source checkout of this repo, where — and only there —
`go run ./cmd/abcd` works, the published payload carrying no `cmd/`. To put a
binary on `PATH`, run `ahoy install` through whichever rung just resolved:
`"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install`, `abcd ahoy install`, or
`go run ./cmd/abcd ahoy install` in a source checkout.

**User input:** $ARGUMENTS
