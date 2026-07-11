# Skills — Procedural Workflows That Aren't Commands

abcd's surface namespace (`/abcd:`) carries both **commands** and a small set of **skills**. The skill/command distinction is documented here for the boundary rules; abcd ships **three** user-facing skills under `/abcd:` — `consult`, `ingest`, and `prepare-this-repo` (each a `skills/<name>/SKILL.md`).

## What's a skill?

A **skill** is a procedural-workflow surface — a markdown-encoded interview, audit, or guided pass that runs against existing material in the repo. Skills are auto-registered by Claude Code's plugin system from the plugin's `skills/<skill-name>/` directory; abcd does not maintain a separate skill registry.

**abcd ships three skills** (`consult`, `ingest`, `prepare-this-repo`). An earlier version of this brief proposed `/abcd:grill` as a skill; its **design target** is promotion to a sub-verb of `/abcd:intent` (per the round-2 command-structure review, see `04-surfaces/05-intent.md § 2`) — but neither `grill` nor the `intent` parent is on any shipped surface (no binary verb, no `commands/abcd/` file). Grill's mid-session glossary writes and per-session logbook output are command-shaped responsibilities — exactly the trigger for promotion this file describes below.

## What's a command?

A **command** is a stateful operation that creates, modifies, or moves artefacts. Commands have:

- A **brief surface file** (`04-surfaces/NN-<verb>.md`) describing acceptance criteria, interaction flow, and side effects (or a sub-verb row in an existing parent's surface file). Every shipped verb has one — `docs`, `history`, and `version` are chapters `10`–`12`.
- A **logbook subdirectory** (`.abcd/logbook/<verb>/<timestamp>/`) for per-invocation reports.
- A **status + help** mode when called bare — the bare-status-board convention holds for `ahoy`, `capture`, and `memory` (and bare `abcd`); it is not universal (`version` prints only the version string, and the `docs`/`history` cobra parents print help/usage with no status board).

abcd ships **seven top-level commands**: `/abcd:ahoy`, `/abcd:capture`, `/abcd:docs`, `/abcd:history`, `/abcd:launch`, `/abcd:memory`, `/abcd:version`. **Design targets — not on any shipped surface (no binary verb, no `commands/abcd/` file):** `/abcd:disembark`, `/abcd:embark`, and `/abcd:intent` (design record in `04-surfaces/02-disembark.md`, `03-embark.md`, `05-intent.md`). The `intent` parent's design gives it the largest sub-verb tree (`refine`, `grill`, `plan`, `ship`, `review`, `consistency`, `shape`, `reclassify`, `link`), plus the canonical bare quoted create `/abcd:intent "<text>"` — all design target, none shipped. See [`04-surfaces/`](../04-surfaces) for per-command detail.

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

**abcd ships three user-facing skills under `/abcd:`** — `consult`, `ingest`, and `prepare-this-repo`, each a `skills/<name>/SKILL.md` that the plugin system auto-registers as a slash-invokable skill. (`skills/` ships in the tree but is currently absent from the release payload `.abcd/config/launch-payload.json` includes — tracked as drift in iss-61.)

The shipped `skills/` entries are three user-facing workflow skills:

1. **`consult`** — consult the local sources corpus (default `~/.abcd/sources`) and record source→decision provenance in its append-only ledger.
2. **`ingest`** — register a URL or document into the local sources corpus with extracted reference metadata.
3. **`prepare-this-repo`** — an interim bridge that brings the current repository up to abcd's conventions (three-tier `.abcd/` layout, an AGENTS.md section, commit gates).

There are no per-command workflow directories (`skills/abcd-<verb>/`) and no hook-helper directories in the shipped tree.

```
skills/      consult/            (consult the sources corpus; record provenance)
             ingest/             (register a source into the corpus)
             prepare-this-repo/  (interim repo-onboarding bridge)
```

The three skills above ARE the current registration list. If a later phase introduces further user-facing skills under `/abcd:` (a slash-invokable skill that does NOT have a parent command), it is added here.

## Future skills

itd-30 (design fictions, a later phase) is a **command extension**, not a new skill — it extends the canonical create `/abcd:intent "<text>"` with `--format=fiction`.

If a later phase introduces skills like `/abcd:plan-stress-test` (cross-intent adversarial review) or `/abcd:walkthrough` (read-aloud orientation pass), each gets:

- An intent file (capturing the user moment)
- A `skills/<skill-name>/{SKILL.md, workflow.md}` directory (the executable form)
- An entry in this file's "Skill registration" section
- **No** `04-surfaces/` file unless the skill is command-shaped (in which case it ships as a command, not a skill).

The skills-vs-commands boundary is enforced by reviewer judgment, not by lint. The strict rule from the round-2 review: **any logbook output, any artefact mutation, any state change → command, not skill.** Findings-only, idempotent, read-only-against-existing-content → skill.
