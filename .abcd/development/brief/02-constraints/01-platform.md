# Platform & Repos

This file holds locked decisions about *where* abcd runs and *what repos exist*. These are non-negotiable starting conditions for any agent working on abcd — the agent designs architecture inside these rails, not around them.

## This directory IS abcdDev

The cwd is the private dev repo where abcd is built. The plugin lives at the repo root (`commands/`, `skills/`, `hooks/`, `scripts/`, `agents/`, `tests/`, `docs/`); spec/task tracking lives at `.flow/`; the design record (this brief, the roadmap, intents, ADRs, research) lives at `.abcd/development/`; `.work/` and `.specstory/` are gitignored local-only surfaces. See [`../../../../CLAUDE.md`](../../../../CLAUDE.md) for the canonical "where things live" map. Curated snapshots ship to the public `abcd/` repo via `/abcd:launch`.

## Repos

See [`01-product/02-context.md`](../01-product/02-context.md#repos) for the canonical Repos list (SSOT). Summary: `abcdSubZero/` (reference), `abcdDev/` (this directory, private dev), `abcd/` (public release target via `/abcd:launch`).

## Lifeboat path

`.abcd/lifeboat/` (per-repo, gitignored unless private).

**Lifeboat is always *output*.** `.abcd/lifeboat/` holds the latest snapshot this repo would ship — produced by `disembark`, regenerable from current state. Disembark overwrites it (with a `.bak` safety net per [`04-surfaces/02-disembark.md § 7`](../04-surfaces/02-disembark.md#7-acceptance)); there is no `lifeboat-v1/` / `lifeboat-v2/` proliferation. Past disembarks are recorded as manifests (hash + file list + label), not as preserved snapshots — see [`04-surfaces/03-embark.md § 7`](../04-surfaces/03-embark.md#7-voyage-layout-embarkdisembark-provenance-and-history).

## Embark sources

**Input lifeboats are external by default.** Embark reads from `embark from <path>` (use `home` for the current repo's own lifeboat — the round-trip / self-test case); the source repo's `.abcd/lifeboat/` is the canonical copy. Embark records source path + manifest hash in `.abcd/development/voyage/embark/provenance.json`. Opt-in `embark from <path> --archive` copies the input lifeboat verbatim into `.abcd/development/voyage/embark/from/<timestamp>/` for the rare case where the source repo will disappear.

**Embark sources, in order (post bare-as-help refactor — see [`04-surfaces/03-embark.md`](../04-surfaces/03-embark.md)):**

1. `embark from home` → expands to `.abcd/lifeboat/` in cwd (round-trip case: embark from this repo's own disembark output, e.g., for testing)
2. `embark from <path>` (any explicit path)
3. `embark scan` (or `embark scan --deep`) → discovery sub-verb that walks sibling directories (`../`), lists `.abcd/lifeboat/` candidates ranked by mtime — does not unpack; pass the chosen path to `embark from <path>`
4. Free-text path input via the embark interview if `<path>` is omitted on `from`

**No global archive (`~/.abcd/archive/`).** Lifeboats live in their producing repo; share externally by copy.

## Validation corpus

See [`01-product/02-context.md`](../01-product/02-context.md) for the canonical validation-corpus list (SSOT). Summary: `idelphiDev/` (primary), `abcdSubZero/`, `idelphiSubZero/`. Per-phase acceptance runs against the corpus with documented exemptions where a feature genuinely doesn't apply.
