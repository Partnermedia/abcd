# Invariants

> **Status: PLACEHOLDER.** Invariant-shaped rules are currently scattered through `05-internals/04-universal-patterns.md`, the visibility rule in `04-surfaces/`, and various command-acceptance sections. Future iterations may consolidate the load-bearing ones here. For now, treat this file as a navigational pointer.

## Properties the system must preserve regardless of how it's built

The following are non-negotiable invariants — any architectural choice that violates them is wrong even if it works.

1. **Transparent prompts** — every `AskUserQuestion` shows current state, consequences of each option, and how to change later. No silent defaults. See [`05-internals/04-universal-patterns.md § 1`](../05-internals/04-universal-patterns.md#1-transparent-prompts).

2. **Adapter over native default** — every capability abcd once delegated to a bundled tool has a native default; the external tool is an opt-in adapter over the same seam ([adr-22](../../decisions/adrs/0022-bundled-deps-as-pluggable-adapters.md)). A missing or misbehaving adapter degrades to the native path; abcd never blocks on an absent tool. See [`05-internals/04-universal-patterns.md`](../05-internals/04-universal-patterns.md).

3. **Native default front door** — abcd's logic lives in a transport-agnostic Go core; the CLI is the reliable default front door that exercises it, and every capability is reachable with no plugin host present ([adr-23](../../decisions/adrs/0023-transport-agnostic-core.md)). Additional front doors (MCP server, markdown plugin surface) are thin adapters over the same core. See [`05-internals/04-universal-patterns.md`](../05-internals/04-universal-patterns.md).

4. **JSON internal, MD render** — all inter-agent data is JSON; markdown is a render step. See [`05-internals/04-universal-patterns.md § 4`](../05-internals/04-universal-patterns.md#4-json-internal-md-render).

5. **Visibility is one switch** — `repo.visibility` (private | public) is the single switch governing what gets committed. No per-subdirectory exceptions. If sensitivity is a concern, set visibility=public (which gitignores the entire `.abcd/` namespace). See [`05-internals/03-configuration.md § 1`](../05-internals/03-configuration.md#1-visibility-driven-gitignore-policy).

6. **Lifeboat is always *output*, and disembark never writes to the source** — the lifeboat is regenerable, and it is written **out-of-tree** to an operator-chosen `<dest>`, never back into the repo being read (`disembark <source-repo> to <dest>`). Operations history lives at the operator level, `~/.abcd/voyage/<source-root-sha>/`, keyed on the root-commit SHA like the history store — never committed, and never accumulated as stale snapshots. Per [adr-35](../../decisions/adrs/0035-lifeboat-as-coverage-experiment.md), superseding adr-4's in-tree `.abcd/lifeboat/` and `.abcd/development/voyage/`. See [`02-constraints/01-platform.md § Lifeboat path`](01-platform.md#lifeboat-path) and [`04-surfaces/03-embark.md § 7`](../04-surfaces/03-embark.md#7-voyage-layout-embarkdisembark-provenance-and-history).

7. **Acceptance discipline applies uniformly** — every intent's press release is followed by a `## Acceptance Criteria` block in Given-When-Then format (per itd-1). Every brief phase has an `## Acceptance` block in the same format. The format is uniform across the boundary; the *home* differs to match the nature of the work. See [`01-product/03-mental-model.md`](../01-product/03-mental-model.md).
