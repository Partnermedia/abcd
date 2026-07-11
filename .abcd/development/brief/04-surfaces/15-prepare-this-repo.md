# `/abcd:prepare-this-repo` — Repo Onboarding Bridge

`/abcd:prepare-this-repo` brings the current repository up to abcd's working
conventions: it reads the abcd record, audits the repo against it, then adopts
the three-tier `.abcd/` layout, a nameless working-conventions section in
`AGENTS.md`, and the commit gates. It is an **interim bridge** — abcd cannot yet
manage repositories directly, so the command does by hand what the CLI will
later take over, in a shape the CLI can adopt without unpicking.

It is a **host-delegated command**: no Go verb backs it, there is no bare-status
render and no binary sub-verbs. The whole workflow runs in the host agent from
the markdown in [`commands/abcd/prepare-this-repo.md`](../../../../commands/abcd/prepare-this-repo.md);
the abcd binary is never invoked. It takes no argument — it always operates on
the current repository.

## What it does

- **Refuses on repos the user does not own.** Phase 0 checks the origin remote
  and stops entirely — no audit, no writes — unless the repository is the user's
  or an org they control. Imposing these conventions on a third-party repo would
  interfere with its own development principles.
- **Audits before it touches anything.** It produces a gap report (existing
  structure, documentation shape, decision and working-state hygiene,
  principles followed or violated, privacy) and presents it before adopting
  anything.
- **Adopts the conventions.** The three-tier layout, a merged (never
  overwritten) `AGENTS.md` with verified repo facts plus the marked
  working-conventions block, and — where absent — a secrets + absolute-path
  pre-commit gate. AI-attribution hooks are opt-in only.

## Flow

Five phases, each gated on the one before:

1. **Refuse unless owned** — origin-remote ownership check; stop if it fails.
2. **Orient** — read the abcd record from `$ABCD` (the abcd checkout, three
   levels up from the command file): the three-tier README, the brief,
   principles, ADRs, intents, `docs/` Diátaxis rules, and the lint configs as
   patterns.
3. **Audit** — write the gap report to the target's
   `.abcd/.work.local/scratch/` and present it before any change.
4. **Adopt** — create the three tiers with a repo-specific `CONTEXT.md`, migrate
   any historical `.work/` layout to the new tiers (propose then wait for sign-off;
   never leave a repo with both the old and new working-state homes), merge into
   `AGENTS.md`, offer the commit gates.
5. **Attribution (opt-in)** — install the AI-disclosure hook only if the user
   says the repo requires it.

When abcd's own record has conflicting sources, the command trusts a fixed
authority order: `AGENTS.md`, then `work/CONTEXT.md`'s live-constraints section,
then ratified ADRs, then everything else read for understanding only.

## Boundaries

- **Nameless, self-contained output.** The working-conventions block written
  into `AGENTS.md` never mentions abcd, this command, or any private repository
  — the conventions read as the repo's own, between dated markers so later
  tooling can find and replace them.
- **Never commit downstream assets.** Anything tooling will later provide
  (`personas.json`, lint-config JSON, content copied from the abcd record) is
  applied, not copied. Only content about the target repository is committed.
- **Privacy.** `private-names.txt`, if present, is read-only context for the
  audit and never reproduced in any committed or published artefact.

## Acceptance

- **Given** a repo the user does not own, **when** they run
  `/abcd:prepare-this-repo`, **then** it stops at Phase 0 with no audit and no
  writes.
- **Given** an owned repo, **when** the command runs, **then** a gap report
  exists under `.abcd/.work.local/scratch/` and was presented before anything
  was adopted.
- **Given** sign-off, **when** it adopts, **then** the three-tier layout exists
  with a repo-specific `CONTEXT.md`, `AGENTS.md` carries verified repo facts and
  the marked nameless working-conventions section (the done-test — a fresh agent
  can build and test from `AGENTS.md` alone — passes), and any historical `.work/`
  layout is fully migrated or fully left alone.
- **Given** the adoption completes, **then** nothing from `private-names.txt`
  and no abcd-internal content appears in any committed artefact.

## Composition

`/abcd:prepare-this-repo` is one of the three host-delegated user-facing
commands under `/abcd:` that carry no Go verb; `/abcd:consult` and `/abcd:ingest`
(over the `~/.abcd/sources` corpus) are the others. It supersedes the older
`scaffold-repo` layout (historical `.work/` at the repo root) and migrates it on sight. Its
end state is the `.abcd/` layout the shipped abcd surfaces
(`/abcd:capture`, `/abcd:docs`, `/abcd:ahoy`) then operate over.

## References

- Plugin command: [`commands/abcd/prepare-this-repo.md`](../../../../commands/abcd/prepare-this-repo.md)
- The three-tier layout it adopts: [`../01-product`](../01-product) and the abcd `.abcd/README.md`
- The invariants the working-conventions block encodes: [`../02-constraints`](../02-constraints)
