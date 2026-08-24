# Documentation

User-facing documentation, organised by [Diátaxis](https://diataxis.fr/) — each page
is exactly one type. Development records (brief, intents, ADRs, plans, research) live
under [`../.abcd/development/`](https://github.com/intentdriven/abcd/tree/main/.abcd/development/), not here.

| Directory | Diátaxis type | For |
|-----------|---------------|-----|
| [`tutorials/`](tutorials/README.md) | Tutorial | Learning-oriented — a guided first run. |
| [`how-to/`](how-to/README.md) | How-to | Task-oriented — accomplish a specific goal. |
| [`reference/`](reference/README.md) | Reference | Information-oriented — config, schemas, and the [CLI reference](reference/cli/README.md) (generated from the Cobra command tree and gated against drift by a build test). |
| [`explanation/`](explanation/README.md) | Explanation | Understanding-oriented — the mental model and the why. |

The CLI reference under `reference/cli/` is generated from the Cobra command tree
and gated against drift by a build test. Everything else is hand-authored.
