# Invariants

> **Status: PLACEHOLDER.** Invariant-shaped rules are currently scattered through `05-internals/04-universal-patterns.md`, the visibility rule in `04-surfaces/`, and various command-acceptance sections. Future iterations may consolidate the load-bearing ones here. For now, treat this file as a navigational pointer.

## Properties the system must preserve regardless of how it's built

The following are non-negotiable invariants — any architectural choice that violates them is wrong even if it works.

1. **Transparent prompts** — every `AskUserQuestion` shows current state, consequences of each option, and how to change later. No silent defaults. See [`05-internals/04-universal-patterns.md § 1`](../05-internals/04-universal-patterns.md#1-transparent-prompts).

2. **MCP-preferred, structural-fallback** — every external-tool call has a configured backend AND a structural fallback. Never blocks. See [`05-internals/04-universal-patterns.md § 2`](../05-internals/04-universal-patterns.md#2-mcp-preferred-structural-fallback).

3. **Plugin-preferred, internal-fallback** — when another plugin already provides a capability, prefer it; reimplement only when the preferred provider isn't installed. See [`05-internals/04-universal-patterns.md § 3`](../05-internals/04-universal-patterns.md#3-plugin-preferred-internal-fallback).

4. **JSON internal, MD render** — all inter-agent data is JSON; markdown is a render step. See [`05-internals/04-universal-patterns.md § 4`](../05-internals/04-universal-patterns.md#4-json-internal-md-render).

5. **Visibility is one switch** — `repo.visibility` (private | public) is the single switch governing what gets committed. No per-subdirectory exceptions. If sensitivity is a concern, set visibility=public (which gitignores the entire `.abcd/` namespace). See [`05-internals/03-configuration.md § 1`](../05-internals/03-configuration.md#1-visibility-driven-gitignore-policy).

6. **Lifeboat is always *output*** — `.abcd/lifeboat/` is regenerable, overwritten by each disembark. History lives in `.abcd/development/voyage/`, not as accumulated snapshots. See [`02-constraints/01-platform.md § Lifeboat path`](./01-platform.md#lifeboat-path) and [`04-surfaces/03-embark.md § 7`](../04-surfaces/03-embark.md#7-voyage-layout-embarkdisembark-provenance-and-history).

7. **abcd never picks the model** — when the oracle backend is RP MCP, RepoPrompt routes models internally based on the user's UI configuration. abcd issues an MCP call and consumes the result. Cross-model perspective lives in RP, not in abcd. See [`05-internals/01-agents.md` § Oracle backend resolution](../05-internals/01-agents.md#oracle-backend-resolution).

8. **Same-chat re-review for RP** — when abcd re-runs an oracle/review/audit after applying fixes, the re-call MUST stay in the same RP chat. Never `--new-chat`, never fresh `rp builder`. The harness threads `chat_id` through the audit-fix loop. "Same chat" is narrowed (post-fn-5, per ADR-02 § 3) to **same-`MCPBridge`-instance / same-MCP-session**: a `chat_id` is meaningful only within the `MCPBridge` instance that produced it, and cross-instance reuse is undefined behaviour.

9. **Acceptance discipline applies uniformly** — every intent's press release is followed by a `## Acceptance Criteria` block in Given-When-Then format (per itd-1). Every brief phase has an `## Acceptance` block in the same format. The format is uniform across the boundary; the *home* differs to match the nature of the work. See [`01-product/03-mental-model.md`](../01-product/03-mental-model.md).
