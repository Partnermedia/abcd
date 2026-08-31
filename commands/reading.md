---
name: reading
description: Assemble the input a cold reading is handed and validate the output it returns, by invoking the abcd binary. Bare invocation is a read-only status render; assemble produces the assembled input and its hashed manifest, and ingest validates one reading's output and writes its records.
argument-hint: "[] | assemble --position <widening|entailment|detection> --target <HEAD|sha> --scope <record-id|kind|preset> [--out <dir>] [--dry-run] | ingest --reading-json <path>"
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
position and a target state, and the reading's object and question come from
its definition, so there is no channel through which ledger content can travel
in the framing of a request.

## Status (bare)

To render the assembler's state:

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" reading --json
```

Summarise the JSON for the user: `assembler_version`, `include_rows` and
`exclusion_rows` (what the table admits and what it refuses), `definitions`
(the reading definitions present under `agents/`), and `staged_runs` (runs
sitting in the local tier). Zero writes.

## Assemble one reading's input

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" reading assemble \
  --position widening --target HEAD --json
```

`--position` takes one of four closed tokens — `widening`, `entailment`,
`comparative`, `detection`. An unknown token is refused by name. `--target`
takes `HEAD` or a hexadecimal commit sha of 7 to 40 digits; a branch name or a
tag is refused, because it moves and the manifest's re-runnability rests on a
reference that cannot.

`--scope` names what the reading is **about**, and is required. It takes one
token in a closed grammar: a record id (`itd-N`, `spc-N`, `adr-N`, `iss-N`), a
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
position's corpus and reporting success.

Assembly reads the working tree, so it refuses unless HEAD resolves to the
target **and** no included path is uncommitted. The preset configuration is in
that dirty set, for the same reason the record configuration is: an
uncommitted edit to it reshapes the assembly. Both refusals exit 2, as does
an unknown position, a missing operand, and any positional argument.

Report from the JSON: `run_id`, `position`, `target_commit`, `item_count`,
`manifest_hash`, and — where the run wrote — `out_dir` and `artefacts`.

Report `scope` too: the token the operator gave, the selectors it resolved to,
and whether the run was an override.

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
  --position entailment --target HEAD \
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
a reserved name, or a body matching a registered signature. A **list-level**
violation refuses the whole run: a wrong `_type`, a run id that resolves to no
parked manifest, a manifest hash that disagrees, an instrument claiming a
definition hash or an assembler version the artefacts do not carry, a regime
disagreeing with the definition, or a payload in which no item survived.

**A refusal leaves a record once the run's identity is proven** — that is, once
the run id resolves to a parked manifest whose content hash matches. From that
point a list-level refusal writes `refusal.json` under the run's directory,
carrying the run metadata and the named reason and no items, and the refusal
message and the JSON render both name it. A refusal reached BEFORE that point —
a wrong `_type`, a run id that resolves to nothing, a manifest hash that
disagrees — writes nothing anywhere, because there is no proven run to record
against.

**A rerun is a new run with a new run id, never an amendment.** Once a run id
has an outcome — a commit marker or a refusal record — ingesting it again is
refused. Assemble again, and ingest the run that assembly parked.

### The supply regime is the definition's

Each position's definition states the regime the reading reads under, and the
verb reads it from there. An output whose self-declared regime disagrees is
refused. No operand and no configuration key sets a regime, by design.

Per regime, the reserved names — a field naming one is refused with the licence
stated:

- `evaluative`: `order`, `rank`, `recommended`, `score`. Arrangement order is
  never refused; items arrive in document order by mandate.
- `registrative`: `fix`, `remedy`, `resolution`.
- `explicative`: `disposition`, `status`.
- `generative` has no reserved names. Its licence is the widest, and the
  constraint on it falls at admission, so a signature hit there raises a review
  flag on the run record instead of refusing the item.

Prose that ranks, settles or proposes without the field is caught too, by a
registry of named signatures (`RG-EVAL-ORDERING`, `RG-EVAL-RECOMMENDATION`,
`RG-REG-FIXPROPOSAL`, `RG-EXPL-DISPOSITION`). That half is bounded by the
registry: a fix proposal or a disposition phrased outside it is not caught, and
the structural halves above carry no such bound.

### Where the records land

Nothing durable is written until the whole payload validates. The reading
records land in the reading-record family, the run's manifest is promoted
beside its run metadata, and the run metadata is written **last** as the commit
marker — a run without one never happened. An ingest interrupted before that
marker leaves a stage in the local tier, and the next invocation names the
orphan, rolls the run back and clears it. One ingest runs at a time in a
checkout: a second waits, and reports contention rather than sweeping the
first one's records away.

Report from the JSON: `run_id`, `records`, `refused_items`, `review_flags`,
`cleared_stages`, and `run_record` — or, on a refusal that recorded one,
`refusal_record`.

**Binary resolution.** Run `"${CLAUDE_PLUGIN_ROOT}/abcd"` — a plugin install
provisions the binary into the plugin root, so this is the rung that fires for a
plugin user. If that path does not exist, try `abcd` on `PATH`; if that fails
too, you are in a source checkout of this repo, where — and only there —
`go run ./cmd/abcd` works, the published payload carrying no `cmd/`. To put a
binary on `PATH`, run `ahoy install` through whichever rung just resolved:
`"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install`, `abcd ahoy install`, or
`go run ./cmd/abcd ahoy install` in a source checkout.

**User input:** $ARGUMENTS
