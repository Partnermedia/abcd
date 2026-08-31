# The cold-reading fixture corpus

Inert on disk; materialised into a temporary directory per run by
`evals/coldreading_fixture_test.go`, which commits it under an invented identity
in a reserved example domain so the manifest's commit reference resolves.

- `baseline/` — a whole miniature repository state with every plant in its
  canonical home. One sentinel token per warm location class, of the form
  `ABCD-EVAL-SENTINEL-<CLASS>`, so a leak names the class that leaked.
- `holed/` — the negative control. It holds the replacement content for the two
  included files a relocated plant lands in; the relocations themselves are
  declared in the eval's `holes` table, which also names the canonical home each
  plant is removed from. A baseline plus declared holes rather than two full
  trees, so the variants cannot drift apart in everything the hole does not
  touch.
- `home/` — the fixture `HOME`, carrying a planted session-transcript store. Its
  `ROOT_COMMIT_SHA` directory is renamed to the fixture repository's own
  root-commit sha at materialisation, so the transcript class of brief invariant
  15 is genuinely reachable if the assembler ever walked it.

Nothing here is a live record: the corpus is never scanned by this repository's
own record, docs or site gates, whose roots are the repository's own `.abcd/`
and `docs/`.

## Why the plants sit where they do

Each plant is placed so that some rule of the assembler's contract is
falsifiable by it — the corpus is adversarial per rule, not merely
representative. Two placements are load-bearing and easy to get wrong:

- **An excluded heading needs a home on a record type that travels WHOLE.** On a
  projected record type the projection keeps the heading out whatever the
  exclusion floor says, so deleting that heading's exclusion row leaks nothing
  and the rule cannot be falsified there. Every excluded heading therefore has a
  home in a brief chapter, a discipline or a spec; the copies on the shipped
  intent exercise the projected shape as well.
- **The projection's positive half needs a section that is neither projected nor
  excluded.** Without one, a corpus cannot tell "the projection is positive at
  field granularity" from "the projection is gone and the redaction cleaned up
  after it". That is the `## Residue` section on the shipped, draft and planned
  intents.

`evals/coldreading_coverage_test.go` holds the matrix that records this, rule by
rule, including the rules this corpus cannot falsify and why.
