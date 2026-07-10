# Surfaces — User-Facing Commands

The brief's user-facing command surface is the set enumerated below (not all are shipped yet — see [`06-delivery/`](../06-delivery) for current delivery state). Each has its own file with the surface contract: purpose, flow, acceptance criteria. (Operator-internal commands are wiring, not part of this user-facing surface — e.g. `run`, the autonomous-run operator surface over the pluggable run seam (`status`/`pause`/`resume`/`preflight`): a design target (itd-29, `intents/planned/`); no `run` verb or `commands/abcd/run.md` is on any shipped surface.)

| # | Command | Purpose | File |
|---|---|---|---|
| 1 | `/abcd:ahoy` | Install / update abcd in any project | [`01-ahoy.md`](01-ahoy.md) |
| 2 | `/abcd:disembark` | Pack a lifeboat from the current project — design target, not shipped: no `disembark` binary verb or `commands/abcd/disembark.md` (Phase 6, adr-33) | [`02-disembark.md`](02-disembark.md) |
| 3 | `/abcd:embark` | Unpack a lifeboat into a (typically empty) target — design target, not shipped: no `embark` binary verb or `commands/abcd/embark.md` (Phase 6, adr-33) | [`03-embark.md`](03-embark.md) |
| 4 | `/abcd:launch` | Preview the curated release bundle and gates (read-only); cutting/publishing the artefact is a design target (itd-72) — see [`04-launch.md`](04-launch.md) | [`04-launch.md`](04-launch.md) |
| 5 | `/abcd:intent` | Capture / refine / grill / plan / ship / review / consistency / shape / reclassify / link intents (press-release format; three review roles, three verbs per [`05-intent.md § 7`](05-intent.md#7-the-intent-fidelity-reviewer-agent-three-roles-three-verbs)) — design target, not shipped: no `intent` binary verb or `commands/abcd/intent.md` (backing intents in `intents/planned/`) | [`05-intent.md`](05-intent.md) |
| 6 | `/abcd:capture` | Issue ledger (capture / list / resolve / wontfix; `promote` is a design target per spc-30/itd-46, skill-orchestrated — see [`06-capture.md`](06-capture.md)) | [`06-capture.md`](06-capture.md) |
| 7 | `/abcd:memory` | Multi-upstream curated knowledge substrate (per itd-36) — `ingest` external sources / `ask` queries / `lint` health-checks. Component spec: [`05-internals/07-memory.md`](../05-internals/07-memory.md). | [`07-memory.md`](07-memory.md) |
| 8 | `/abcd` | Top-level where-am-i status board (per itd-20, `intents/planned/`) — cross-command re-orientation. The shipped bare render is four read-only lines (directory, git repo, record present, `.abcd/` work tiers); the richer board (visibility, lifeboat, dev-sync, recent logbook, active intents, next actions) is a design target. `status` is a positional plugin alias for the bare render; `help` is not an alias (the binary has no `status` verb, and `abcd help` prints command usage). Read-only. | [`08-abcd.md`](08-abcd.md) |
| 9 | `/abcd:reflect` | Phase retrospective (per itd-24) — `/abcd:reflect <phase-id>` composes a five-section retrospective (went well / could improve / lessons / decisions / metrics) seeded by the spc-66 phase-audit receipt. Phase-only grain — design target, not shipped: no `reflect` binary verb or `commands/abcd/reflect.md` (itd-24) | [`09-reflect.md`](09-reflect.md) |

**Bare-command-as-help is a universal abcd convention** — every command shows read-only status when invoked without args (bare `abcd`, `abcd ahoy`, `abcd capture`, `abcd memory` all render status). The **suggested-next-actions** half of this convention is a design target: no shipped bare invocation yet emits next actions (bare `abcd launch` currently only hints to pass `--dry-run`). Provides discoverability without forcing the user to remember subcommand names.

**abcd ships three user-facing skills under `/abcd:`** — `/abcd:consult`, `/abcd:ingest`, and `/abcd:prepare-this-repo`. The plugin's `skills/` directory holds exactly these three, each a single `SKILL.md` addressed to the user (`ingest`'s frontmatter cross-references `/abcd:consult` as its companion); there are no plugin-runtime workflow files or hook helpers under `skills/`. The skill/command boundary is documented in [`05-internals/08-skills.md`](../05-internals/08-skills.md); the rest of the `/abcd:` surface is commands. (`/abcd:grill` was originally proposed as a user-facing skill but is now a sub-verb of `/abcd:intent` — its mid-session glossary writes and per-session logbook output are command-shaped.)

## Where to find related design

- **Plumbing internals** (agents, adapters, configuration, universal patterns, prompt quality): [`05-internals/`](../05-internals)
- **Build sequence** (which surfaces ship in what order): [`06-delivery/01-build-sequence.md`](../06-delivery/01-build-sequence.md)
- **Verification matrix** (test coverage across surfaces): [`06-delivery/02-verification-matrix.md`](../06-delivery/02-verification-matrix.md)
