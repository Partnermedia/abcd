# Skills — Procedural Workflows That Aren't Commands

abcd's surface namespace (`/abcd:`) is **commands only**. The skill/command distinction is documented here for the boundary rules; abcd ships zero skills.

## What's a skill?

A **skill** is a procedural-workflow surface — a markdown-encoded interview, audit, or guided pass that runs against existing material in the repo. Skills are auto-registered by Claude Code's plugin system from the plugin's `skills/<skill-name>/` directory; abcd does not maintain a separate skill registry.

**abcd ships zero skills.** An earlier version of this brief proposed `/abcd:grill` as a skill; it has since been promoted to a sub-verb of `/abcd:intent` (per the round-2 command-structure review, see `04-surfaces/05-intent.md § 2`). Grill's mid-session glossary writes and per-session logbook output are command-shaped responsibilities — exactly the trigger for promotion this file describes below.

## What's a command?

A **command** is a stateful operation that creates, modifies, or moves artefacts. Commands have:

- A **brief surface file** (`04-surfaces/NN-<verb>.md`) describing acceptance criteria, interaction flow, and side effects (or a sub-verb row in an existing parent's surface file).
- A **logbook subdirectory** (`.abcd/logbook/<verb>/<timestamp>/`) for per-invocation reports.
- A **status + help** mode when called bare (the universal abcd convention).

abcd ships **six top-level commands**: `/abcd:ahoy`, `/abcd:disembark`, `/abcd:embark`, `/abcd:launch`, `/abcd:intent`, `/abcd:capture`. The `intent` parent has the largest sub-verb tree (`refine`, `grill`, `plan`, `ship`, `review`, `consistency`, `shape`, `reclassify`, `link`), plus the canonical bare quoted create `/abcd:intent "<text>"`. See [`04-surfaces/`](../04-surfaces) for per-command detail.

## Skill vs command — decision criteria

A surface is a **skill** when:

- The verb describes a *workflow that runs against existing content* — interview, audit, review, walkthrough, stress-test.
- Output is *findings or suggestions only*, never artefact creation/modification.
- The procedure is *re-runnable on the same input* without different effects (idempotent).
- The surface fits naturally as a workflow markdown file the agent reads and follows.

A surface is a **command** when ANY of:

- The verb describes a *state change* — install, pack, unpack, capture, plan, ship.
- Output includes *new or modified artefacts* (files, directory moves, frontmatter updates) — even alongside findings.
- Re-running has different effects (idempotent or not, but state-mutating).
- The surface needs an `acceptance` block, side-effect documentation, and (often) a checkpoint/resume protocol.
- The surface writes to a `.abcd/logbook/<verb>/` subdirectory.

When in doubt, ship as a command. The earlier "ship as a skill first; promote on mutation" guidance was overturned by the round-2 review: by the time you discover a skill is mutating state, downstream contracts have hardened around the skill shape and rework is expensive. Better to recognise command-shape up front.

## Skills are not in `04-surfaces/`

`04-surfaces/` documents commands. Skills do **not** get a surface file there. The plugin's `skills/<skill-name>/SKILL.md` is the executable form; the intent file (when one exists) is the canonical user-moment reference.

## Skill registration

**abcd ships zero user-facing skills under `/abcd:`.** The `skills/` directory in the plugin layout (per `03-configuration.md` § 3) holds **plugin-runtime workflow files** — these are how Claude Code's plugin architecture wires command markdown to its workflow files. They are NOT user-facing skills surfaced under `/abcd:` and are NOT what this document is talking about when it says "skills."

The plugin layout's `skills/` entries fall into two non-user-facing categories:

1. **Per-command workflow files** — `skills/abcd-ahoy/`, `skills/abcd-disembark/`, etc. These are the runtime mechanics each command's markdown points at. Not user-invokable in their own right; users invoke the command, the command runtime invokes the workflow.
2. **Internal hook helpers** — `skills/commit-attribution/`, `skills/secrets-and-pii/`. These are utility shims invoked by hooks (commit attribution, PII/secret scanning) — not surfaced to users via the `/abcd:` namespace.

```
skills/      (no user-facing skills under /abcd:; the skills/ directory holds
              plugin-runtime workflow files only — see 03-configuration.md § 3)
```

If a later phase introduces user-facing skills under `/abcd:` (i.e., a slash-invokable skill that does NOT have a parent command and is NOT a hook helper), this section gets a per-phase registration list.

## Future skills

itd-30 (design fictions, a later phase) is a **command extension**, not a new skill — it extends the canonical create `/abcd:intent "<text>"` with `--format=fiction`.

If a later phase introduces skills like `/abcd:plan-stress-test` (cross-intent adversarial review) or `/abcd:walkthrough` (read-aloud orientation pass), each gets:

- An intent file (capturing the user moment)
- A `skills/<skill-name>/{SKILL.md, workflow.md}` directory (the executable form)
- An entry in this file's "Skill registration" section
- **No** `04-surfaces/` file unless the skill is command-shaped (in which case it ships as a command, not a skill).

The skills-vs-commands boundary is enforced by reviewer judgment, not by lint. The strict rule from the round-2 review: **any logbook output, any artefact mutation, any state change → command, not skill.** Findings-only, idempotent, read-only-against-existing-content → skill.
