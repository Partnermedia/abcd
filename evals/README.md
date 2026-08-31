# evals — repository evals

Two lanes share one harness and one built binary: a self-discovering smoke test
for the `abcd` binary, and the cold-reading read-block eval.

## The smoke harness

A self-discovering smoke test for the `abcd` binary. It builds the real binary,
walks the Cobra command tree **in-process** (via `cli.NewRootCommand()`) to
discover every command and flag, and exercises each against the built binary — so
a command added tomorrow is covered here with no edit.

## What it checks (v1)

- **Every** command and subcommand: `abcd <cmd> --help` exits 0, produces output,
  and never panics. This catches the failure unit tests miss — a command that
  compiles but crashes when actually invoked.
- **Read-only, no-argument verbs** (`version`, the bare status board) run for real
  to a graceful exit.
- **Flag hygiene:** an unknown flag is a clean non-zero error, not a panic.

## The cold-reading evals

`coldreading_*_test.go` falsifies the cold-reading input assembler's read-block
rather than restating it. It plants a sentinel token for every warm location
class the record names into a fixture repository state under
`testdata/cold-reading/`, materialises and commits that state in a temporary
directory with `HOME` redirected to a planted transcript store, runs
`abcd reading assemble` over it at each reading position, and asserts three
absences over what the assembler wrote: no planted token in the raw
serialisation, no excluded frontmatter key or heading at any depth in a parsed
item, and no item from an excluded record family.

The oracle is independent of the assembler by construction. The eval invokes
the binary out of process, its exclusion table is transcribed by hand from the
record (itd-183's exclusion list, brief invariants 14 and 15, adr-55) with the
source on every row, and `TestOracleImportsNothingFromTheAssembler` parses every
Go file here to check that nothing imports the assembler's package. An eval that
read the assembler's own include table could only ever confirm the table.

`testdata/cold-reading/baseline/` holds every plant in its canonical home;
`holed/` is the negative control, holding the replacement content for the two
files a relocated plant lands in; `home/` is the fixture HOME carrying the
planted transcript store, keyed on the fixture's root-commit sha at
materialisation.

### What keeps it from passing vacuously

An absence eval's characteristic failure is asserting nothing while looking
green, so the corpus is adversarial per rule and three separate guards stand
under the assertions:

- `TestEverySentinelIsPlanted` — the corpus keeps its plants, at their declared
  count, in their declared homes, all tracked by git (the assembler walks the
  tracked set, so an untracked plant tests nothing).
- The carrier floor — every plant-bearing file the include list names arrives at
  each position, and its own **cold marker text** is in the bundle's bytes. A
  manifest names what an assembly says it passed; only the bundle says what it
  actually passed.
- The declared table sizes — each oracle table asserts its count rather than
  merely being non-empty, because a `> 0` floor on a table whose size is known
  lets it halve unnoticed.

The plants are chosen so that each rule of the assembler's contract has one that
dies when the rule is removed — including the **positive** half of the field
projection, which needs a section that is neither projected nor on the exclusion
floor, and each excluded heading, which needs a home on a record type that
travels whole (on a projected type the projection keeps the heading out whatever
the floor says, so its exclusion cannot be falsified there).

`coldreading_coverage_test.go` is the matrix: one row per rule, the mutation that
removes it, and the plants that die. A rule no mutation can falsify carries its
reason in `Gap` rather than being quietly omitted, and
`TestEveryAssemblerRuleHasAFalsifier` fails if a row names a plant that has gone,
if a plant is named by no row, or if the number of declared gaps changes.

## Running them

Both lanes sit behind the `smoke` build tag so they stay out of the fast
unit-test lane:

```bash
make smoke                       # both lanes
go test -tags smoke ./evals/...  # the same thing

make evals-cold-reading                # the cold-reading evals alone
go test -tags coldreading ./evals/...  # the same thing
```

CI runs the smoke harness as the dedicated `smoke` job, and the release workflow
smokes the binary built from the tagged commit before publishing. The
cold-reading evals get their own `cold-reading-evals` job, which carries no
`inert` condition: the diff classifier stands the `smoke` job down on a change
confined to `docs/`, `.abcd/development/`, `.abcd/work/` and the root prose
files, and those are exactly the paths the cold-reading evals read.

A file visible to both lanes carries `//go:build smoke || coldreading`, which is
also how a later cold-reading eval joins the lane — no Makefile or workflow edit.

## `data/` (reserved for v2)

Fixture-driven, per-command scenarios — user-specified and synthetic inputs the
harness auto-discovers to drive richer smokes (e.g. `memory ingest` over a sample
corpus, `capture` round-trips). Deferred; the generalisation into an abcd-managed
eval framework is captured as intent **itd-75**.
