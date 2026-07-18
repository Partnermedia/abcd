# Documentation

User-facing documentation, organised by [Diátaxis](https://diataxis.fr/) — each page
is exactly one type. Development records (brief, intents, ADRs, plans, research) live
under [`../.abcd/development/`](../.abcd/development/), not here.

| Directory | Diátaxis type | For |
|-----------|---------------|-----|
| [`tutorials/`](tutorials/) | Tutorial | Learning-oriented — a guided first run. |
| [`how-to/`](how-to/) | How-to | Task-oriented — accomplish a specific goal. |
| [`reference/`](reference/) | Reference | Information-oriented — config, schemas, and the [CLI reference](reference/cli/) (a planned generated reference; today `abcd <verb> --help` is authoritative). |
| [`explanation/`](explanation/) | Explanation | Understanding-oriented — the mental model and the why. |

The CLI reference under `reference/cli/` is a placeholder for a planned generated
reference (from the Cobra command tree); that generation is not yet wired, so the
live CLI reference today is `abcd <verb> --help`. Everything else is hand-authored.
