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
