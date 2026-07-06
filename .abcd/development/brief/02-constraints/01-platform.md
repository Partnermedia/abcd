# Platform & Repo

This file holds locked decisions about *where* abcd runs and *how it ships*. These are non-negotiable starting conditions for any agent working on abcd — the agent designs architecture inside these rails, not around them.

## One repository

abcd lives in **one repository** and ships a **curated release artifact** cut from it — there is no dev→public mirror ([adr-28](../../decisions/adrs/0028-single-repo-curated-release.md)). The Go binary is the product ([adr-21](../../decisions/adrs/0021-rebuild-in-go.md)); its source, its user-facing documentation, and the design record all share this one tree. The design record (this brief, the roadmap, intents, ADRs, research) lives in-tree at `.abcd/development/`; `.work/` is a gitignored local-only surface. See [`../../../../CLAUDE.md`](../../../../CLAUDE.md) for the canonical "where things live" map.

**`.abcd/**` stays in-tree but is excluded from the release artifact by packaging** — exclusion is a build-time filter over the one tree, not a copy between two repos. `/abcd:launch` cuts a curated GitHub Release from this repo; **the repo is the marketplace**, so discovery, install, and the design record share one location.

## Front doors

The core is a transport-agnostic Go package ([adr-23](../../decisions/adrs/0023-transport-agnostic-core.md)) behind thin front doors. The **CLI (Cobra)** is the first front door and ships in the MVP. An **MCP server** and a **markdown plugin surface that shells to the Go binary** are later front doors — each is a new adapter over the unchanged core, not the substrate abcd is built on.

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
