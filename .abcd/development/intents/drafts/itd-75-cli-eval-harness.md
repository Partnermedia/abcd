---
id: itd-75
slug: cli-eval-harness
spec_id: null
kind: standalone
suggested_kind: null
reclassification_history: []
severity: minor
---

# abcd Proves Your CLI Actually Runs — From Fixtures You Drop In

## Press Release

> **Every command a repo exposes gets exercised, and you extend the coverage by dropping a folder — never by writing test plumbing.** abcd builds the binary, discovers the whole command surface by walking it (no hand-kept list), and smokes each command: help renders, nothing panics, read-only verbs run. Then it goes further — a fixtures folder of user-specified and synthetic inputs is auto-discovered and replayed against the matching commands, asserting the *shape* of what comes back. Adding a scenario is dropping an input and an expected-shape file into `evals/data/<command>/`; the harness picks it up on the next run, in local dev, in CI, and in the release gate against the very binary about to ship.
>
> "I used to find out a command crashed only when a user hit it," said Kira, a maintainer. "Now every command is smoked on every push, and when I want a real scenario I drop a sample corpus in a folder and abcd runs my tool against it. My confidence in a release stopped being a feeling."

## Why This Matters

Unit tests prove functions behave; they do not prove the *assembled binary* runs. A command can compile, pass its unit tests, and still panic the moment it is invoked — a broken wiring, a nil dependency, a flag that explodes on parse. The only honest check is to run the real binary through its real command surface. Doing that by hand rots: someone must remember to add each new command to the list. Discovering the surface from the binary itself removes that decay, and a fixtures-folder convention lets coverage grow by contribution rather than by editing a harness.

This is also the natural home for the "wired or it isn't done" rule: a command that is not reachable and runnable from the built binary is not done, and the eval harness is what makes that enforceable rather than aspirational.

## What It Looks Like

- **Self-discovering smoke (shipped as the v1 dogfood).** Walk the command tree, run every command's `--help` and the read-only verbs against the built binary, assert no panic and graceful exit. New commands are covered automatically.
- **Fixture-driven scenarios.** `evals/data/<command>/` holds inputs (user-specified real samples and synthetic generators) plus an expected-shape assertion. The harness matches folders to commands and replays them — `memory ingest` over a corpus, `capture` round-trips, `launch --dry-run` over a scratch tree — with no Go edit per scenario.
- **Any abcd-managed repo inherits it.** `abcd ahoy` scaffolds the harness and an empty `evals/data/` with instructions, and wires the smoke into the repo's CI and release gate. A repo becomes self-smoking by being abcd-managed.
- **Release-artefact smoke.** The gate runs the harness against the binary built from the tagged commit, so the thing that ships is the thing that was exercised.

## Dogfood (v1 running on this repo)

`evals/` in abcd-cli is the working prototype: a `smoke`-tagged Go harness that builds `abcd`, discovers the command tree via the exported root command, and smokes every command; wired into a `smoke` CI job, the release verify gate, and `make smoke`. This intent lifts that from a repo-local harness into an abcd capability every managed repo inherits, and adds the fixtures layer. Relates to the install surface (`ahoy`) and the "wired or it isn't done" principle.

## Open Questions

- Fixture format: one folder per command with an input corpus + an expected-shape file, or a single manifest describing scenarios?
- How synthetic data is generated and kept deterministic (seeded generators vs committed samples), and how a repo marks a fixture as containing only safe, non-sensitive data.
- Which commands are safe to run for real by default versus opt-in, and how a mutating command declares a scratch sandbox.
