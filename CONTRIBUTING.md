# Contributing

abcd is a private incubation repo for now; it flips public at maturity. See
[`AGENTS.md`](AGENTS.md) for build/test/checks and working conventions, and
[`.abcd/development/`](.abcd/development/) for the design record.

- **Branch + PR** for substantive changes; CI gates the merge. Its `check` job
  builds, vets and tests (plain and race-enabled) on macOS + Linux, and on the
  Linux leg alone adds the `gofmt -l .` format gate and the record-lint and
  docs-lint steps; separate jobs run the reviews-charter check, `gitleaks`,
  `zizmor` and the smoke harness. `make preflight` runs the same build, vet, test
  and race steps locally via the pre-push hook, together with the lint-reviews,
  record-lint and docs-lint gates — but not gofmt, so run `gofmt -l .` before
  pushing.
- **Conventional-commit prefixes** (`feat`/`fix`/`docs`/`chore`/`refactor`/`test`/`ci`),
  no scopes; short title, body explains why.
- A **CHANGELOG** entry accompanies any user-facing change.
- **Docs** are Diátaxis (one type per page, present tense); the design record lives
  under `.abcd/`, never in `docs/`.

## AI assistance and authorship

Development of abcd is assisted by an AI coding assistant. Two rules keep that
honest:

- **Human author of record.** The human contributor is the author of every change
  they submit and is responsible for all AI-assisted output — its correctness, its
  licensing, and its fit for the project. AI assistance never transfers that
  responsibility.
- **Disclosure by trailer, not co-authorship.** AI-assisted commits carry an
  `Assisted-by: Claude:<model-version>` trailer (the Linux kernel format) —
  disclosure only. abcd never uses `Co-Authored-By:` for AI: it asserts an
  authorship the tool does not hold and inflates the contributor graph. A
  human-only `Signed-off-by:` (DCO) is deferred until the repo is public or takes
  its first outside contribution.
- **Wrote it yourself? Say so.** A change no AI assisted carries
  `Assisted-by: None` on its commit and in its pull-request body. Disclosure runs
  both ways: the gate refuses a silent omission because it cannot tell "no AI was
  involved" from "I forgot the trailer", so the human-only case states itself
  rather than being inferred. Never write a vendor trailer for work a model did
  not touch — a false disclosure is worse than none, and the reviewer reading the
  diff is the check on it.

## Acknowledgements

[`ACKNOWLEDGEMENTS.md`](ACKNOWLEDGEMENTS.md) credits the ideas, tools, and writing
behind abcd in three parts — development, inspirations, and references. Add an entry
**in the same change that lands it**: the PR that adopts an external pattern, cites
a source in an ADR, or integrates a tool. Adding it at the moment it lands is what
keeps the file from going stale — it is never reconstructed after the fact.
