# Surfaces — User-Facing Commands

The brief's user-facing command surface is the set enumerated below (not all are shipped yet — see [`06-delivery/`](../06-delivery/) for current delivery state). Each has its own file with the surface contract: purpose, flow, acceptance criteria. (Operator-internal commands under `commands/abcd/` — e.g. `deps-check`, `ralph-up`, `session`, and `run` (the itd-29 autonomous-run operator surface: `status`/`pause`/`resume`/`preflight`) — are wiring, not part of this user-facing surface.)

| # | Command | Purpose | File |
|---|---|---|---|
| 1 | `/abcd:ahoy` | Install / update abcd in any project | [`01-ahoy.md`](./01-ahoy.md) |
| 2 | `/abcd:disembark` | Pack a lifeboat from the current project | [`02-disembark.md`](./02-disembark.md) |
| 3 | `/abcd:embark` | Unpack a lifeboat into a (typically empty) target | [`03-embark.md`](./03-embark.md) |
| 4 | `/abcd:launch` | Promote `*Dev` → public sibling repo | [`04-launch.md`](./04-launch.md) |
| 5 | `/abcd:intent` | Capture / refine / grill / plan / ship / review / consistency / shape / reclassify / link intents (press-release format; three review roles, three verbs per [`05-intent.md § 6`](./05-intent.md#6-the-intent-fidelity-reviewer-agent-three-roles-three-verbs)) | [`05-intent.md`](./05-intent.md) |
| 6 | `/abcd:capture` | Issue ledger (capture / list / promote / resolve / wontfix) | [`06-capture.md`](./06-capture.md) |
| 7 | `/abcd:memory` | Multi-upstream curated knowledge substrate (per itd-36) — `ingest` external sources / `ask` queries / `lint` health-checks. Component spec: [`05-internals/07-memory.md`](../05-internals/07-memory.md). | [`07-memory.md`](./07-memory.md) |
| 8 | `/abcd` | Top-level where-am-i status board (per itd-20) — cross-command re-orientation: project + visibility, lifeboat, dev-sync, recent logbook, active intents, next actions. `status` / `help` are byte-identical aliases. Read-only. | [`08-abcd.md`](./08-abcd.md) |
| 9 | `/abcd:reflect` | Phase retrospective (per itd-24) — `/abcd:reflect <phase-id>` composes a five-section retrospective (went well / could improve / lessons / decisions / metrics) seeded by the fn-66 phase-audit receipt. Phase-only grain. | [`09-reflect.md`](./09-reflect.md) |

**Bare-command-as-help is a universal abcd convention** — every command shows status + suggested next actions when invoked without args. Provides discoverability without forcing the user to remember subcommand names.

**abcd ships zero user-facing skills under `/abcd:`.** The skill/command boundary is documented in [`05-internals/08-skills.md`](../05-internals/08-skills.md) for later additions; the `/abcd:` surface namespace is commands only. The plugin's `skills/` directory does contain plugin-runtime workflow files and internal hook helpers (per `05-internals/03-configuration.md § 3`), but those are not user-facing skills under `/abcd:` — they are wiring. (`/abcd:grill` was originally proposed as a user-facing skill but is now a sub-verb of `/abcd:intent` — its mid-session glossary writes and per-session logbook output are command-shaped.)

## Where to find related design

- **Plumbing internals** (agents, adapters, configuration, universal patterns, prompt quality): [`05-internals/`](../05-internals/)
- **Build sequence** (which surfaces ship in what order): [`06-delivery/01-build-sequence.md`](../06-delivery/01-build-sequence.md)
- **Verification matrix** (test coverage across surfaces): [`06-delivery/02-verification-matrix.md`](../06-delivery/02-verification-matrix.md)
