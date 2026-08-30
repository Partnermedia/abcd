---
name: reading
description: Assemble the input a cold reading is handed, by invoking the abcd binary. Bare invocation is a read-only status render; assemble produces the assembled input and its hashed manifest.
argument-hint: "[] | assemble --position <widening|entailment|comparative|detection> --target <HEAD|sha> [--out <dir>] [--dry-run]"
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

Assembly reads the working tree, so it refuses unless HEAD resolves to the
target **and** no included path is uncommitted. Both refusals exit 2, as does
an unknown position, a missing operand, and any positional argument.

Report from the JSON: `run_id`, `position`, `target_commit`, `item_count`,
`manifest_hash`, and — where the run wrote — `out_dir` and `artefacts`.

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

**Binary resolution.** Run `"${CLAUDE_PLUGIN_ROOT}/abcd"` — a plugin install
provisions the binary into the plugin root, so this is the rung that fires for a
plugin user. If that path does not exist, try `abcd` on `PATH`; if that fails
too, you are in a source checkout of this repo, where — and only there —
`go run ./cmd/abcd` works, the published payload carrying no `cmd/`. To put a
binary on `PATH`, run `ahoy install` through whichever rung just resolved:
`"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install`, `abcd ahoy install`, or
`go run ./cmd/abcd ahoy install` in a source checkout.

**User input:** $ARGUMENTS
