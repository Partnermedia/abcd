# Contributing

abcd is a private incubation repo for now; it flips public at maturity. See
[`AGENTS.md`](AGENTS.md) for build/test/checks and working conventions, and
[`.abcd/development/`](.abcd/development/) for the design record.

- **Branch + PR** for substantive changes; CI (build/vet/test on macOS + Linux,
  gitleaks, zizmor) gates the merge. `make preflight` runs the same checks locally
  via the pre-push hook.
- **Conventional-commit prefixes** (`feat`/`fix`/`docs`/`chore`/`refactor`/`test`/`ci`),
  no scopes; short title, body explains why.
- A **CHANGELOG** entry accompanies any user-facing change.
- **Docs** are Diátaxis (one type per page, present tense); the design record lives
  under `.abcd/`, never in `docs/`.
