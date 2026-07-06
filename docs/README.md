# Documentation

User-facing documentation, organised by [Diátaxis](https://diataxis.fr/) — each page
is exactly one type. Development records (brief, intents, ADRs, plans, research) live
under [`../.abcd/development/`](../.abcd/development/), not here.

| Directory | Diátaxis type | For |
|-----------|---------------|-----|
| [`tutorials/`](tutorials/) | Tutorial | Learning-oriented — a guided first run. |
| [`how-to/`](how-to/) | How-to | Task-oriented — accomplish a specific goal. |
| [`reference/`](reference/) | Reference | Information-oriented — config, schemas, and the generated [CLI reference](reference/cli/). |
| [`explanation/`](explanation/) | Explanation | Understanding-oriented — the mental model and the why. |

The CLI reference under `reference/cli/` is **generated** from the Cobra command tree
and must not be hand-edited — regenerate it. Everything else is hand-authored.
